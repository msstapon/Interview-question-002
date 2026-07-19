# interview-question-002

ระบบ **สมัครสมาชิก / ลงชื่อเข้าใช้งาน** ตามโจทย์ IT 02 — 3 หน้าจอ (IT 02-1 ล็อกอิน, IT 02-2 สมัครสมาชิก, IT 02-3 หน้าต้อนรับ)

- **Frontend**: Vue 3 + Vite + TypeScript + Pinia + Vue Router
- **Backend**: Go + Echo v4 + GORM (Clean Architecture) — Argon2id password hashing, RS256 JWT
- **Database**: PostgreSQL (schema จัดการด้วย migration แบบ pure-Go)

โครงสร้าง backend อ้างอิงสถาปัตยกรรมเดียวกับโปรเจกต์ `keyvc-backend` (domain → usecase → repository → delivery + infrastructure + pkg) รายละเอียดการออกแบบดูที่ [`DESIGN.md`](./DESIGN.md)

```
interview-question-002/
├── DESIGN.md          # เอกสารออกแบบ
├── backend/           # Go API (Echo + GORM)
└── frontend/          # Vue 3 SPA
```

## Prerequisites
- Go 1.24+
- Node.js 18+ (แนะนำ) — โปรเจกต์ pin Vite 4 ให้ทำงานกับ Node 16 ได้เช่นกัน
- PostgreSQL (โจทย์นี้ต่อผ่าน SSH tunnel)

---

## 1. Database via SSH tunnel

Postgres เข้าถึงได้ผ่าน bastion เท่านั้น เปิด tunnel ค้างไว้ 1 เทอร์มินัล:

```bash
ssh -L 5433:127.0.0.1:5432 <user>@<bastion-host>
```

จากนั้น backend จะต่อ `localhost:5433` เหมือน DB อยู่เครื่องเดียวกัน (ตั้งค่าใน `backend/.env`)

---

## 2. Backend

```bash
cd backend
cp .env.example .env          # แก้ค่าการเชื่อมต่อ / origin ตามจริง
go run ./cmd/keys             # สร้าง RS256 keypair -> ./secrets/  (ครั้งเดียว)
go run ./cmd/migrate create-db # สร้าง database (DB_NAME) ถ้ายังไม่มี
go run ./cmd/migrate up        # สร้างตาราง users
go run ./cmd/api               # รัน API ที่ :8080
```

(ถ้ามี `make`: `make keys`, `make db-create`, `make migrate-up`, `make run`)

### API (envelope `{ success, data, error }`, prefix `/api/v1`)
| Method | Path | Auth | Body / ผลลัพธ์ |
|--------|------|------|----------------|
| POST | `/api/v1/auth/register` | — | `{ username, password, confirm_password }` → `201` UserDTO |
| POST | `/api/v1/auth/login` | — | `{ username, password }` → `200` `{ access_token, expires_in, user }` |
| GET | `/api/v1/me` | Bearer | → `200` UserDTO |
| GET | `/healthz` `/readyz` | — | liveness / readiness |

---

## 3. Frontend

```bash
cd frontend
cp .env.example .env          # VITE_API_BASE_URL (ค่าเริ่มต้นชี้ http://localhost:8080/api/v1)
npm install
npm run dev                   # เปิด http://localhost:5173
```

ขั้นตอนใช้งาน: IT 02-1 → กด "สมัครสมาชิก" → IT 02-2 กรอก User/Password/Confirm → กลับ IT 02-1 → ล็อกอิน → IT 02-3 เห็น `Welcome User : xxx`

---

## 4. Quick test (curl)

```bash
# สมัคร
curl -s -X POST localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"password123","confirm_password":"password123"}'

# ล็อกอิน → เก็บ token
TOKEN=$(curl -s -X POST localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"password123"}' | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

# /me
curl -s localhost:8080/api/v1/me -H "Authorization: Bearer $TOKEN"
```

## Testing
```bash
cd backend && go test ./...     # unit tests (usecase)
```

## Security notes
- Password เก็บเป็น **Argon2id** เท่านั้น (ไม่มี plaintext)
- JWT เป็น **RS256**; private key อยู่ใน `backend/secrets/` (gitignored) สร้างด้วย `go run ./cmd/keys`
- ไฟล์ `.env` ทั้งสองฝั่งถูก gitignore — commit เฉพาะ `.env.example`
