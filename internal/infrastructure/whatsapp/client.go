package whatsapp

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type ClientManager struct {
	Container *sqlstore.Container
	Log       waLog.Logger
}

func NewClientManager(db *sqlx.DB) (*ClientManager, error) {
	storeLogger := waLog.Stdout("Database", "DEBUG", true)

	container := sqlstore.NewWithDB(db.DB, "postgres", storeLogger)

	if err := container.Upgrade(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to upgrade device store: %w", err)
	}

	return &ClientManager{
		Container: container,
		Log:       waLog.Stdout("Client", "INFO", true),
	}, nil
}

func (m *ClientManager) NewClient() (*whatsmeow.Client, error) {
	device := m.Container.NewDevice()
	return whatsmeow.NewClient(device, waLog.Stdout("Client", "INFO", true)), nil
}

func (m *ClientManager) GetClientByJID(jid types.JID) (*whatsmeow.Client, error) {
	device, err := m.Container.GetDevice(context.Background(), jid)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, fmt.Errorf("device not found")
	}
	return whatsmeow.NewClient(device, waLog.Stdout("Client", "INFO", true)), nil
}

func (m *ClientManager) GetClientByPhoneNumber(phone string) (*whatsmeow.Client, error) {
	devices, err := m.Container.GetAllDevices(context.Background())
	if err != nil {
		return nil, err
	}

	// Find all matching devices
	var matched []*store.Device
	for _, d := range devices {
		if d.ID != nil && d.ID.User == phone {
			matched = append(matched, d)
		}
	}

	if len(matched) == 0 {
		return nil, fmt.Errorf("no device found for phone %s", phone)
	}

	// Sort by DeviceID descending (assuming higher ID = newer)
	// This is a heuristic; ideally we'd use LastSeen but it's not easily available in this struct view
	// without checking the store details.
	// Simple bubble sort or just finding max is enough for small N
	best := matched[0]
	for _, d := range matched {
		if d.ID.Device > best.ID.Device {
			best = d
		}
	}

	return whatsmeow.NewClient(best, waLog.Stdout("Client", "INFO", true)), nil
}
