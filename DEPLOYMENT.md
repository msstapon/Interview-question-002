# Deployment & CI/CD

CI/CD ด้วย GitHub Actions → build Docker images ขึ้น **GHCR** → **SSH deploy** ไป VPS แล้วรัน `docker compose` (มิเรอร์แนวทางของ `keyvc-backend`)

## Pipeline

```
push → main
  ├─ test-backend    (go vet + go test -race)
  ├─ test-frontend   (npm ci + vue-tsc + vite build)
  ├─ build           (matrix: backend + frontend → ghcr.io/<owner>/interview-question-002-*)
  └─ deploy          (scp compose+Caddyfile → VPS, docker compose pull + up -d)
```
- `pull_request` / branch อื่น → รัน **ci.yml** (test + build ตรวจ compile) เฉยๆ ไม่ push image ไม่ deploy
- `push main` / manual → รัน **deploy.yml** เต็ม pipeline

## Runtime stack บน VPS (`deployments/docker/docker-compose.yml`)

```
             ┌───────── caddy (:80/:443) ─────────┐
 internet ──▶│  /api/* /healthz → api:8080          │
             │  /*              → frontend:80 (nginx)│
             └──────────────────────────────────────┘
  keygen(one-shot) → RS256 keypair → volume:secrets ─▶ api (ro)
  migrate(one-shot) → migrate up ─▶ postgres:16 (volume:pgdata, internal only)
```
- `postgres` ไม่ publish ออก host (เข้าถึงเฉพาะ network ภายใน) — DB `test_tcc` ถูกสร้างอัตโนมัติจาก `POSTGRES_DB`
- `keygen` สร้าง JWT keypair ลง docker volume แบบ idempotent (ไม่ rotate ของเดิม) — ไม่ต้องจัดการไฟล์ secret เอง
- Caddy: ตั้ง `SITE_ADDRESS` เป็นโดเมน → auto-HTTPS, หรือ `:80` → HTTP บน IP

## สิ่งที่ต้องตั้งใน GitHub (ก่อน deploy ทำงาน)

**Settings → Secrets and variables → Actions → New secret:**
| Secret | ใช้ทำอะไร |
|--------|-----------|
| `VPS_HOST` | IP/hostname ของ VPS |
| `VPS_USER` | user สำหรับ SSH (เช่น `root` หรือ `deploy`) |
| `SSH_PRIVATE_KEY` | private key (ตัว public key ใส่ใน `~/.ssh/authorized_keys` บน VPS) |
| `GHCR_TOKEN` | PAT (scope `read:packages`) ให้ VPS pull image จาก GHCR *(ข้ามได้ถ้าตั้ง package เป็น public)* |

> `GITHUB_TOKEN` (สำหรับ push image ตอน build) GitHub ให้มาอัตโนมัติ ไม่ต้องตั้งเอง
> แนะนำสร้าง **Environment ชื่อ `production`** (job `deploy` ผูกไว้) เผื่อใส่ protection rule / required reviewers

## One-time setup บน VPS

ใช้สคริปต์ `deployments/server-setup.sh` รันทีเดียว (ติดตั้ง Docker + สร้างโฟลเดอร์ + `.env` สุ่ม DB_PASSWORD + ใส่ deploy key):

```bash
# บน VPS (รันในฐานะ root) — ก็อปสคริปต์ขึ้นไป หรือ curl จาก raw GitHub หลัง push แล้ว
sudo GHCR_OWNER=msstapon SITE_ADDRESS=:80 \
     DEPLOY_PUBKEY="ssh-ed25519 AAAA... gh-deploy" \
     bash server-setup.sh
```
สคริปต์เป็น idempotent (รันซ้ำได้ ไม่ทับ `.env` เดิม) จากนั้นตั้ง GitHub secrets แล้ว `git push origin main` ได้เลย

ทำเองทีละขั้นก็ได้:
```bash
curl -fsSL https://get.docker.com | sh
sudo mkdir -p /opt/interview-question-002/deployments/{docker,caddy}
sudo nano /opt/interview-question-002/deployments/docker/.env   # ก็อปจาก .env.deploy.example
# ใส่ SSH public key ที่คู่กับ SSH_PRIVATE_KEY ลงใน ~/.ssh/authorized_keys
```
docker-compose.yml + Caddyfile ไม่ต้องวางเอง — workflow `scp` ให้ทุกครั้งที่ deploy
ถ้าใช้โดเมน: ชี้ DNS A record มาที่ IP ของ VPS ก่อน (Caddy จะขอ cert Let's Encrypt ให้เอง)

## ทดสอบทั้ง stack ในเครื่อง (ไม่ต้องมี VPS / tunnel)

```bash
docker compose -f docker-compose.dev.yml up --build
# เปิด http://localhost:8088  → IT 02-1 → สมัคร → ล็อกอิน → IT 02-3
```
(dev ใช้ host port **8088** และ Postgres **5434** เพื่อไม่ชนกับ keyvc / tunnel)

## Port conflicts (deploy ร่วม VPS กับ keyvc)
- พอร์ต `8080`/`5432` เป็น **พอร์ตภายใน container** เท่านั้น — prod ไม่ publish ออก host จึงไม่ชน
- ตัวที่ publish ออก host จริงคือ **Caddy 80/443** เท่านั้น ถ้า VPS มี Caddy ของ keyvc อยู่แล้ว เลือกทางใดทางหนึ่ง:
  1. **deploy คนละ VPS** (ง่ายสุด)
  2. ตั้ง `HTTP_PORT`/`HTTPS_PORT` ใน `.env` ให้ใช้พอร์ตอื่น (เช่น 8081/8443) แล้วให้ Caddy ของ keyvc reverse_proxy มาที่พอร์ตนั้น
  3. รวมเป็น Caddy ตัวเดียว: เพิ่ม site block ใน Caddyfile ของ keyvc ชี้มาที่ container `iq002-api` / `iq002-frontend` (ต้องต่อ docker network เดียวกัน) แล้วเอา service `caddy` ออกจาก compose นี้

## Rollback
รัน workflow `deploy` ซ้ำจาก commit เดิม (Actions → Run workflow) หรือบน VPS:
```bash
cd /opt/interview-question-002/deployments/docker
export GHCR_OWNER=<owner> TAG=<commit-sha-เก่า>
docker compose pull && docker compose up -d
```

## หมายเหตุ
- GHCR package เริ่มต้นเป็น **private** — ถ้าไม่อยากตั้ง `GHCR_TOKEN` ให้ไปที่ package → Package settings → เปลี่ยนเป็น Public
- image tag ใช้ทั้ง `latest` และ `<full-commit-sha>` (deploy ผูกกับ sha เพื่อ rollback ได้แม่นยำ)
