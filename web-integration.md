# Web Integration Guide (QR Connect Flow)

Ringkas untuk mengintegrasikan API ini ke website/front-end.

## Alur Singkat
1) `POST /api/v1/sessions/create` dengan payload `agentId`, `agentName`, `apiKey` (Langchain) → respons berisi `qrCode` (raw) dan `qrCodeBase64`.
2) Tampilkan QR (pakai `qrCodeBase64` → `data:image/png;base64,<value>`).
3) Polling status/detail setiap 3–5 detik memakai `GET /sessions/detail?agentId=...` atau `GET /sessions/status?agentId=...`.
4) Berhenti polling saat `status` menjadi `connected` atau `disconnected`.

## Trigger Langchain (setelah sesi aktif)
- Endpoint: `POST /api/v1/langchain/execute`
- Body minimal:
```json
{
  "agentId": "agent_01",
  "message": "User Message",
  "sender": "6281234567890"
}
```
- Backend akan meneruskan ke `{langchainUrl}/agents/{agentId}/execute` dengan header `Authorization: Bearer <apiKey>` dari sesi, lalu menyimpan respon ke tabel `langchain_executions`. Gunakan field `langchainUrl` dan `apiKey` saat `create` session agar panggilan ini berhasil.

## Endpoint & Payload
- Buat sesi:
```bash
curl -X POST http://localhost:8080/api/v1/sessions/create \
  -H "Content-Type: application/json" \
  -d '{"agentId":"agent_01","agentName":"Bot Test","apiKey":"langchain-api-key"}'
```
Respons awal idealnya sudah memuat QR. Jika tidak, gunakan polling detail untuk mengambil QR terbaru.

- Polling status/detail (tanpa auth):
```
GET /api/v1/sessions/status?agentId=agent_01
GET /api/v1/sessions/detail?agentId=agent_01
```

## Status yang Perlu Ditangani
- `waiting_scan`: QR aktif; tampilkan `qrCodeBase64`.
- `qr_timeout`: QR kadaluarsa; backend otomatis reconnect dan menerbitkan QR baru. Front-end cukup terus polling detail/status untuk mendapatkan QR terbaru; update tampilan saat `qrCodeBase64` berubah.
- `connected`: sukses dipindai; simpan `phoneNumber` (jika ada) dan hentikan polling.
- `disconnected`: sesi terputus; beri opsi untuk membuat sesi baru.
- `initializing`: startup awal; tunggu transisi ke `waiting_scan` atau `qr_timeout`.

## Pola Front-End Disarankan
- Polling interval 3–5 detik; gunakan etag-like check (bandingkan `qrCodeBase64` atau `status`) untuk menghindari rerender berlebih.
- Saat `qr_timeout`, tampilkan pesan “Membuat QR baru…” sambil tetap polling. QR akan terisi kembali otomatis.
- Gunakan `qrCodeBase64` untuk <img> atau canvas; jangan cache terlalu lama.
- Tampilkan fallback tombol “Refresh QR” hanya jika ingin memberi kontrol manual; tidak wajib karena backend auto-regenerate.

## Keamanan & Konfigurasi
- Jangan tampilkan `apiKey` Langchain ke user akhir; kirim dari server-side atau melalui proxy aman.
- Gunakan HTTPS di lingkungan publik; set rate limit via infra (Nginx/Cloud) bila diperlukan.

## Observabilitas
- Log event penting di front-end (QR muncul, timeout, connected) untuk debug.
- Jika perlu tracing server-side, tambahkan header korelasi di request (mis. `X-Request-ID`) lalu sambungkan ke log aplikasi.
