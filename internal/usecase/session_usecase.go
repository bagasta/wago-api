package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"whatsapp-api/internal/domain/entity"
	"whatsapp-api/internal/domain/repository"
	"whatsapp-api/internal/infrastructure/whatsapp"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type SessionUseCase struct {
	sessionRepo         repository.SessionRepository
	messageRepo         repository.MessageRepository
	waManager           *whatsapp.ClientManager
	clients             map[string]*whatsmeow.Client
	mu                  sync.RWMutex
	defaultUser         string
	defaultLangchainURL string
	langchainUC         *LangchainUseCase
}

func NewSessionUseCase(
	sessionRepo repository.SessionRepository,
	messageRepo repository.MessageRepository,
	waManager *whatsapp.ClientManager,
	defaultUser string,
	defaultLangchainURL string,
	langchainUC *LangchainUseCase,
) *SessionUseCase {
	return &SessionUseCase{
		sessionRepo:         sessionRepo,
		messageRepo:         messageRepo,
		waManager:           waManager,
		clients:             make(map[string]*whatsmeow.Client),
		defaultUser:         defaultUser,
		defaultLangchainURL: defaultLangchainURL,
		langchainUC:         langchainUC,
	}
}

func (uc *SessionUseCase) CreateSession(ctx context.Context, agentID, agentName, langchainAPIKey, langchainURL string) (*entity.Session, error) {
	// Check if session exists
	existing, _ := uc.sessionRepo.GetByUserIDAndAgentID(ctx, uc.defaultUser, agentID)
	if existing != nil {
		return nil, fmt.Errorf("session already exists for agent %s", agentID)
	}

	// Create new WhatsApp client
	client, err := uc.waManager.NewClient()
	if err != nil {
		return nil, err
	}

	// Save session to DB
	session := &entity.Session{
		UserID:    uc.defaultUser,
		AgentID:   agentID,
		AgentName: sql.NullString{String: agentName, Valid: true},
		Status:    "initializing",
		LangchainURL: sql.NullString{
			String: fallbackString(langchainURL, uc.defaultLangchainURL),
			Valid:  fallbackString(langchainURL, uc.defaultLangchainURL) != "",
		},
		LangchainAPIKey: sql.NullString{
			String: langchainAPIKey,
			Valid:  langchainAPIKey != "",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	uc.mu.Lock()
	uc.clients[agentID] = client
	uc.mu.Unlock()

	// Handle events
	client.AddEventHandler(func(evt interface{}) {
		uc.handleEvent(agentID, evt)
	})

	// Get QR Channel
	qrChan, _ := client.GetQRChannel(context.Background())
	firstQR := make(chan struct{}, 1)
	go uc.listenForQR(session, client, qrChan, firstQR)

	if err := client.Connect(); err != nil {
		return nil, err
	}

	// Wait briefly for initial QR so caller can render it
	select {
	case <-firstQR:
	case <-time.After(15 * time.Second):
	}

	return session, nil
}

func (uc *SessionUseCase) listenForQR(session *entity.Session, client *whatsmeow.Client, qrChan <-chan whatsmeow.QRChannelItem, firstQR chan<- struct{}) {
	for evt := range qrChan {
		switch evt.Event {
		case "code":
			qrBase64, _ := whatsapp.GenerateQRCode(evt.Code)

			session.QRCode = sql.NullString{String: evt.Code, Valid: true}
			session.QRCodeBase64 = sql.NullString{String: qrBase64, Valid: true}
			session.Status = "waiting_scan"
			session.LastQRGeneratedAt = sql.NullTime{Time: time.Now(), Valid: true}

			uc.sessionRepo.Update(context.Background(), session)

			select {
			case firstQR <- struct{}{}:
			default:
			}
		case "timeout":
			session.Status = "qr_timeout"
			session.QRCode = sql.NullString{Valid: false}
			session.QRCodeBase64 = sql.NullString{Valid: false}
			uc.sessionRepo.Update(context.Background(), session)

			// Reconnect to get a fresh QR and continue emitting codes
			go client.Connect()
		}
	}
}

// refreshQRIfStale triggers a reconnect to emit a fresh QR if the last code is old/expired.
func (uc *SessionUseCase) refreshQRIfStale(agentID string, session *entity.Session) {
	if session == nil {
		return
	}
	if session.Status != "waiting_scan" && session.Status != "qr_timeout" {
		return
	}
	if session.LastQRGeneratedAt.Valid && time.Since(session.LastQRGeneratedAt.Time) < 50*time.Second {
		return
	}

	uc.mu.RLock()
	client := uc.clients[agentID]
	uc.mu.RUnlock()
	if client != nil {
		go client.Connect()
	}
}

func (uc *SessionUseCase) handleEvent(agentID string, evt interface{}) {
	switch e := evt.(type) {
	case *events.Connected:
		uc.updateSessionStatus(agentID, "connected")
	case *events.LoggedOut:
		uc.updateSessionStatus(agentID, "disconnected")
		uc.mu.Lock()
		delete(uc.clients, agentID)
		uc.mu.Unlock()
	case *events.PairSuccess:
		// Get JID
		uc.mu.RLock()
		client := uc.clients[agentID]
		uc.mu.RUnlock()
		if client != nil {
			jid := client.Store.ID
			uc.updateSessionPhone(agentID, jid.User)
			uc.updateSessionJID(agentID, jid)
		}
	case *events.Message:
		go uc.handleIncomingMessage(agentID, e)
	}
}

func (uc *SessionUseCase) handleIncomingMessage(agentID string, msgEvt *events.Message) {
	if msgEvt == nil || msgEvt.Info.IsFromMe {
		return
	}

	text := extractText(msgEvt)
	if text == "" {
		return
	}

	session, err := uc.sessionRepo.GetByAgentID(context.Background(), agentID)
	if err != nil || session == nil {
		log.Printf("incoming msg: session not found for agent %s: %v", agentID, err)
		return
	}

	from := msgEvt.Info.Sender.User
	to := session.PhoneNumber.String

	if uc.messageRepo != nil {
		msg := &entity.Message{
			SessionID:   session.ID,
			AgentID:     agentID,
			MessageID:   sql.NullString{String: msgEvt.Info.ID, Valid: msgEvt.Info.ID != ""},
			FromNumber:  sql.NullString{String: from, Valid: from != ""},
			ToNumber:    sql.NullString{String: to, Valid: to != ""},
			MessageText: sql.NullString{String: text, Valid: text != ""},
			MessageType: sql.NullString{String: "text", Valid: true},
			Direction:   sql.NullString{String: "incoming", Valid: true},
			Status:      sql.NullString{String: "received", Valid: true},
			CreatedAt:   time.Now(),
		}
		if err := uc.messageRepo.Create(context.Background(), msg); err != nil {
			log.Printf("failed to store incoming message: %v", err)
		}
	}

	if uc.langchainUC != nil {
		// Logic to check if we should respond
		shouldRespond := true
		if msgEvt.Info.IsGroup {
			shouldRespond = false
			log.Printf("[Group Debug] Message from group %s. Sender: %s", msgEvt.Info.Chat, from)

			// Check if mentioned
			// We need the bot's JID to check mentions
			uc.mu.RLock()
			client := uc.clients[agentID]
			uc.mu.RUnlock()

			if client != nil && client.Store != nil && client.Store.ID != nil {
				me := client.Store.ID.User
				botLID := client.Store.LID
				if botLID.User != "" {
					log.Printf("[Group Debug] Bot JID User: %s | Bot LID: %s", me, botLID.String())
				} else {
					log.Printf("[Group Debug] Bot JID User: %s | Bot LID: <none>", me)
				}

				// Check in mentioned JIDs
				ext := msgEvt.Message.GetExtendedTextMessage()
				if ext != nil && ext.ContextInfo != nil {
					log.Printf("[Group Debug] Mentioned JIDs: %v", ext.ContextInfo.MentionedJID)
					for _, mentioned := range ext.ContextInfo.MentionedJID {
						// MentionedJID is usually the full JID (User@Server)
						// Check if it matches me (User) or me@s.whatsapp.net
						// Also check if it matches the bot's LID if available

						// Parse the mentioned JID to get the user part
						parsedMention, _ := types.ParseJID(mentioned)
						mentionedUser := parsedMention.User

						isMatch := false
						if mentionedUser == me {
							isMatch = true
						}

						// Try to resolve LID -> phone number and compare with our JID
						if !isMatch && (parsedMention.Server == types.HiddenUserServer || parsedMention.Server == types.HostedLIDServer) {
							if botLID.User != "" && parsedMention.User == botLID.User {
								isMatch = true
							}
							if altJID, err := client.Store.GetAltJID(context.Background(), parsedMention); err == nil && altJID.User == me {
								isMatch = true
							}
						}

						// Check against LID if available in store
						if !isMatch && client.Store.ID.Device > 0 {
							// We can try to check if the mentioned JID corresponds to our LID
							// But client.Store.ID usually has the phone number JID.
							// However, WhatsApp sometimes mentions the LID (User@lid) instead of Phone (User@s.whatsapp.net)
							// Since we don't easily know our own LID from just client.Store.ID (unless we query server),
							// we can check if the mentioned JID is an LID and if we are the only one mentioned?
							// Or better: The fallback text check usually handles this if the user typed @<Name>

							// Let's try to be smarter:
							// If the mentioned JID is an LID, we might need to map it.
							// For now, let's assume if the text contains the bot's number, it's valid.
						}

						if mentioned == me || mentioned == me+"@s.whatsapp.net" || mentionedUser == me {
							isMatch = true
						}

						if isMatch {
							log.Printf("[Group Debug] Bot mentioned! Will respond.")
							shouldRespond = true
							break
						}
					}
				} else {
					log.Printf("[Group Debug] No ExtendedTextMessage or ContextInfo found")
				}

				// Fallback 1: Check text for @<bot_number>
				if !shouldRespond {
					if strings.Contains(text, "@"+me) {
						log.Printf("[Group Debug] Bot mentioned in text (number)! Will respond.")
						shouldRespond = true
					}
				}

				// Fallback 2: Check text for @<PushName>
				if !shouldRespond && client.Store.PushName != "" {
					if strings.Contains(strings.ToLower(text), strings.ToLower("@"+client.Store.PushName)) {
						log.Printf("[Group Debug] Bot mentioned in text (pushname)! Will respond.")
						shouldRespond = true
					}
				}

				// Fallback 3: Check if mentioned JID matches our LID (if we can find it)
				// Note: Mapping LID to Phone is complex without extra queries.
				// We rely on PushName fallback for now.
				if !shouldRespond && ext != nil && ext.ContextInfo != nil {
					// Just log for debugging
					log.Printf("[Group Debug] Mentioned JIDs (LID check skipped): %v", ext.ContextInfo.MentionedJID)
				}
			} else {
				log.Printf("[Group Debug] Client or Store ID not available")
			}
		}

		if shouldRespond {
			uc.sendTyping(agentID, msgEvt.Info.Chat)
			log.Printf("[Langchain] Executing for agent %s...", agentID)
			exec, err := uc.langchainUC.Execute(context.Background(), agentID, text, from, nil)
			if err != nil {
				log.Printf("langchain execute failed for agent %s: %v", agentID, err)
				uc.stopTyping(agentID, msgEvt.Info.Chat)
			} else {
				reply := extractLangchainReply(exec)
				if reply != "" {
					// Reply to the chat (group or user)
					target := msgEvt.Info.Chat
					if msgEvt.Info.IsGroup {
						log.Printf("[Group Debug] Sending reply to group: %s", target)
					} else {
						log.Printf("[Direct Debug] Sending reply to user: %s", target)
					}

					if err := uc.sendTextMessage(agentID, target, reply); err != nil {
						log.Printf("failed to send langchain reply to %s: %v", target, err)
					} else {
						log.Printf("Successfully sent reply to %s", target)
					}
				} else {
					log.Printf("Langchain returned empty reply")
					uc.stopTyping(agentID, msgEvt.Info.Chat)
				}
			}
		} else {
			log.Printf("[Group Debug] Ignoring message (not mentioned)")
		}
	}
}

func (uc *SessionUseCase) updateSessionStatus(agentID, status string) {
	session, err := uc.sessionRepo.GetByAgentID(context.Background(), agentID)
	if err == nil {
		session.Status = status
		switch status {
		case "connected":
			session.ConnectedAt = sql.NullTime{Time: time.Now(), Valid: true}
			session.QRCode = sql.NullString{Valid: false}
			session.QRCodeBase64 = sql.NullString{Valid: false}
		case "disconnected":
			session.DisconnectedAt = sql.NullTime{Time: time.Now(), Valid: true}
		}
		uc.sessionRepo.Update(context.Background(), session)
	}
}

func (uc *SessionUseCase) updateSessionPhone(agentID, phone string) {
	session, err := uc.sessionRepo.GetByAgentID(context.Background(), agentID)
	if err == nil {
		session.PhoneNumber = sql.NullString{String: phone, Valid: true}
		uc.sessionRepo.Update(context.Background(), session)
	}
}

func (uc *SessionUseCase) updateSessionJID(agentID string, jid *types.JID) {
	session, err := uc.sessionRepo.GetByAgentID(context.Background(), agentID)
	if err == nil && jid != nil {
		meta := map[string]string{
			"jid": jid.String(),
		}
		data, _ := json.Marshal(meta)
		session.SessionData = data
		if err := uc.sessionRepo.Update(context.Background(), session); err != nil {
			log.Printf("Failed to update session JID for %s: %v", agentID, err)
		} else {
			log.Printf("Updated session JID for %s to %s", agentID, jid.String())
		}
	} else if err != nil {
		log.Printf("Failed to get session for JID update %s: %v", agentID, err)
	}
}

func (uc *SessionUseCase) GetSession(ctx context.Context, agentID string) (*entity.Session, error) {
	session, err := uc.sessionRepo.GetByAgentID(ctx, agentID)
	if err != nil || session == nil {
		return session, err
	}

	uc.refreshQRIfStale(agentID, session)
	return session, nil
}

func fallbackString(primary, secondary string) string {
	if primary != "" {
		return primary
	}
	return secondary
}

func extractText(msg *events.Message) string {
	if msg == nil || msg.Message == nil {
		return ""
	}
	if conv := msg.Message.GetConversation(); conv != "" {
		return conv
	}
	if ext := msg.Message.GetExtendedTextMessage(); ext != nil && ext.GetText() != "" {
		return ext.GetText()
	}
	return ""
}

func (uc *SessionUseCase) sendTextMessage(agentID string, to types.JID, text string) error {
	uc.mu.RLock()
	client := uc.clients[agentID]
	uc.mu.RUnlock()
	if client == nil {
		return fmt.Errorf("client not found for agent %s", agentID)
	}

	msg := &waProto.Message{
		Conversation: &text,
	}
	_, err := client.SendMessage(context.Background(), to, msg)
	return err
}

func (uc *SessionUseCase) sendTyping(agentID string, to types.JID) {
	uc.mu.RLock()
	client := uc.clients[agentID]
	uc.mu.RUnlock()
	if client != nil {
		if err := client.SendChatPresence(context.Background(), to, types.ChatPresenceComposing, types.ChatPresenceMediaText); err != nil {
			log.Printf("Failed to send typing presence to %s: %v", to, err)
		}
	}
}

func (uc *SessionUseCase) stopTyping(agentID string, to types.JID) {
	uc.mu.RLock()
	client := uc.clients[agentID]
	uc.mu.RUnlock()
	if client != nil {
		if err := client.SendChatPresence(context.Background(), to, types.ChatPresencePaused, types.ChatPresenceMediaText); err != nil {
			log.Printf("Failed to send paused presence to %s: %v", to, err)
		}
	}
}

func extractLangchainReply(exec *entity.LangchainExecution) string {
	if exec == nil || len(exec.LangchainResponse) == 0 {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal(exec.LangchainResponse, &m); err != nil {
		return ""
	}
	if resp, ok := m["response"].(string); ok && resp != "" {
		return resp
	}
	if msg, ok := m["message"].(string); ok && msg != "" {
		return msg
	}
	return ""
}

type MessageStats struct {
	Incoming  int `json:"incoming"`
	Responded int `json:"responded"`
}

func (uc *SessionUseCase) GetMessageStats(ctx context.Context, agentID string) MessageStats {
	if uc.messageRepo == nil {
		return MessageStats{}
	}
	stats := MessageStats{}
	// Best-effort queries; ignore errors for lightweight stats.
	if incoming, err := uc.messageRepo.CountByAgentAndDirection(ctx, agentID, "incoming"); err == nil {
		stats.Incoming = incoming
	}
	if responded, err := uc.messageRepo.CountByAgentAndDirection(ctx, agentID, "outgoing"); err == nil {
		stats.Responded = responded
	}
	return stats
}

func (uc *SessionUseCase) DeleteSession(ctx context.Context, agentID string) error {
	uc.mu.Lock()
	client, ok := uc.clients[agentID]
	delete(uc.clients, agentID)
	uc.mu.Unlock()

	if ok {
		client.Disconnect()
		// Delete device from store to prevent stale sessions
		if client.Store != nil {
			// Only attempt to delete if we have a valid JID (device is known)
			if client.Store.ID != nil && !client.Store.ID.IsEmpty() {
				log.Printf("Attempting to delete device %s from store for agent %s", client.Store.ID.String(), agentID)
				if err := client.Store.Delete(context.Background()); err != nil {
					log.Printf("Failed to delete device from store for agent %s: %v", agentID, err)
				} else {
					log.Printf("Deleted device from store for agent %s", agentID)
				}
			} else {
				jidStatus := "nil"
				if client.Store.ID != nil {
					jidStatus = "empty"
				}
				log.Printf("Skipping device store deletion for agent %s: device JID is %s (not paired)", agentID, jidStatus)
			}
		}
	}

	return uc.sessionRepo.Delete(ctx, agentID)
}

// InitializeSessions loads connected sessions from DB on startup
func (uc *SessionUseCase) InitializeSessions(ctx context.Context) error {
	sessions, err := uc.sessionRepo.GetAllSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all sessions: %w", err)
	}

	for _, session := range sessions {
		// Only reconnect if it was previously connected or in a state that expects connection
		// You might want to adjust this logic based on your requirements
		if session.Status == "connected" || session.Status == "initializing" || session.Status == "waiting_scan" {
			log.Printf("Restoring session for agent %s (Status: %s)", session.AgentID, session.Status)
			go func(agentID string) {
				if _, err := uc.ReconnectSession(context.Background(), agentID); err != nil {
					log.Printf("Failed to restore session %s: %v", agentID, err)
				} else {
					log.Printf("Successfully restored session %s", agentID)
				}
			}(session.AgentID)
		}
	}
	return nil
}

func (uc *SessionUseCase) ReconnectSession(ctx context.Context, agentID string) (*entity.Session, error) {
	// 1. Get Session from DB
	session, err := uc.sessionRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// 2. Resolve Client
	uc.mu.Lock()
	client, exists := uc.clients[agentID]
	uc.mu.Unlock()

	if !exists {
		// 1. Try to restore specific device using JID from metadata
		if len(session.SessionData) > 0 {
			var meta map[string]string
			if err := json.Unmarshal(session.SessionData, &meta); err == nil {
				if jidStr, ok := meta["jid"]; ok && jidStr != "" {
					log.Printf("Attempting to restore session using stored JID: %s", jidStr)
					parsedJID, _ := types.ParseJID(jidStr)
					if !parsedJID.IsEmpty() {
						client, err = uc.waManager.GetClientByJID(parsedJID)
						if err == nil {
							log.Printf("Restored session using stored JID %s", jidStr)
						} else {
							log.Printf("Failed to restore by JID %s: %v", jidStr, err)
						}
					}
				}
			}
		}

		// 2. Fallback to phone number lookup if JID failed
		if client == nil && session.PhoneNumber.Valid {
			log.Printf("Attempting to restore session using phone number: %s", session.PhoneNumber.String)
			// Try to find by phone number first (searches all devices)
			client, err = uc.waManager.GetClientByPhoneNumber(session.PhoneNumber.String)
			if err != nil {
				log.Printf("Failed to restore by phone number: %v", err)
				// If not found by phone, try JID construction as backup (though GetClientByPhoneNumber should cover it)
				jid := types.NewJID(session.PhoneNumber.String, types.DefaultUserServer)
				client, err = uc.waManager.GetClientByJID(jid)
			}

			if err != nil {
				// Fallback: Create new client if old one is corrupted/missing
				log.Printf("Could not find existing device for %s, creating new one: %v", session.PhoneNumber.String, err)
				client, err = uc.waManager.NewClient()
			} else {
				log.Printf("Restored session using phone number lookup")
			}
		} else if client == nil {
			log.Printf("No phone number or JID found, creating new client")
			client, err = uc.waManager.NewClient()
		}
	}
	if err != nil {
		return nil, err
	}

	// 3. Update Map
	uc.mu.Lock()
	uc.clients[agentID] = client
	uc.mu.Unlock()

	// 4. Setup Listeners
	// Ensure we are listening for QR codes
	qrChan, _ := client.GetQRChannel(context.Background())
	firstQR := make(chan struct{}, 1)

	// We pass the session pointer. listenForQR updates it and the DB.
	go uc.listenForQR(session, client, qrChan, firstQR)

	// Events
	client.AddEventHandler(func(evt interface{}) {
		uc.handleEvent(agentID, evt)
	})

	// 5. Connect
	if !client.IsConnected() {
		// Retry logic for connection
		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			if err := client.Connect(); err != nil {
				log.Printf("Connection attempt %d/%d failed for agent %s: %v", i+1, maxRetries, agentID, err)
				if i < maxRetries-1 {
					time.Sleep(time.Duration(i+1) * 2 * time.Second) // Exponential backoff: 2s, 4s, 6s...
					continue
				}
				return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
			}
			log.Printf("Connected successfully for agent %s", agentID)
			break
		}
	}

	// 6. Wait for QR or short timeout
	select {
	case <-firstQR:
	case <-time.After(2 * time.Second):
	}

	// 7. Refresh session from DB to get latest status (Connected, WaitingScan, etc)
	refreshedSession, err := uc.sessionRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		return session, nil // Return the one we have if DB fails
	}

	// If we got a QR code in the memory struct but DB read was too fast/slow,
	// we might want to ensure we return the QR.
	// listenForQR updates the passed 'session' struct AND the DB.
	// So 'session' should have QR if it happened.
	// 'refreshedSession' should also have it.

	return refreshedSession, nil
}
