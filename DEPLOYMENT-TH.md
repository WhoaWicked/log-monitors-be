# คู่มือ Deploy — Log Monitor

Deploy `log-monitors-be` (Go) + `log-monitors-fe` (React/Vite) ขึ้น VPS ของ DigitalOcean
โดยใช้ Docker (Postgres/Redis), systemd (backend), Nginx (reverse proxy) และ
Let's Encrypt (HTTPS)

อ้างอิงจากที่ deploy จริง: DigitalOcean droplet, Singapore, $6/เดือน (1 vCPU / 1GB RAM), Ubuntu 24.04

---

## 0. เตรียมของก่อน (บนเครื่องตัวเอง)

สร้าง SSH key ถ้ายังไม่มี:
```bash
ssh-keygen -t ed25519
cat ~/.ssh/id_ed25519.pub
```
Copy ผลลัพธ์เก็บไว้ — เอาไปวางตอนสร้าง droplet บน DigitalOcean

## 1. สร้าง VPS

1. สมัคร DigitalOcean, ผูก payment method
2. **Create → Droplets**:
   - Region: Singapore (หรือใกล้ผู้ใช้ที่สุด)
   - Image: Ubuntu 24.04 (LTS) x64
   - Plan: Basic / Regular SSD, **$6/เดือน (1GB RAM)** — 512MB แน่นเกินไปสำหรับ `go build`/`npm install`
   - Authentication: SSH Keys → วาง public key ที่ copy ไว้
   - ข้าม Volumes, Backups, Managed Database ไป
3. SSH เข้า:
```bash
ssh root@YOUR_IP
```
4. อัปเดตระบบ:
```bash
apt update && apt upgrade -y
```
   ถ้าถามเรื่อง `sshd_config` เลือก **"keep the local version currently installed"**

5. **เพิ่ม swap** (RAM 1GB ไม่พอสำหรับ `go build` / `npm install` เดี่ยวๆ — ถ้าไม่ทำขั้นนี้
   VPS อาจค้างจน RAM เต็ม ต้อง hard reboot จาก DigitalOcean dashboard):
```bash
fallocate -l 2G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo '/swapfile none swap sw 0 0' >> /etc/fstab
```

## 2. Docker Compose — Postgres + Redis

```bash
mkdir -p ~/log-monitor && cd ~/log-monitor
openssl rand -base64 24   # สร้าง password ของ DB เลี่ยงอักขระพิเศษอย่าง "/" ถ้าเป็นไปได้
nano docker-compose.yml
```

```yaml
services:
  postgres:
    image: postgres:16-alpine   # ระบุเวอร์ชันให้ชัด — "alpine" เฉยๆ อาจดึง major version ใหม่กว่า
                                 # ที่ format ข้อมูลไม่ตรงกับที่มีอยู่เดิม
    restart: always
    environment:
      - POSTGRES_USER=logmon
      - POSTGRES_PASSWORD=<password ที่ generate ไว้>
      - POSTGRES_DB=logmon_db
    ports:
      - "127.0.0.1:5555:5432"   # bind localhost เท่านั้น — ห้าม expose port ของ DB ออกสู่สาธารณะ
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:alpine
    restart: always
    ports:
      - "127.0.0.1:6380:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

```bash
docker compose up -d
docker compose ps   # ทั้งคู่ควรขึ้น "Up"
```

## 3. Backend — build + ตั้งเป็น systemd service

```bash
# ติดตั้ง Go (ให้ตรงกับเวอร์ชันใน go.mod เช็คด้วย: cat go.mod | head -5)
curl -fsSL https://go.dev/dl/go1.26.1.linux-amd64.tar.gz -o go.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

git clone https://github.com/WhoaWicked/log-monitors-be.git
cd log-monitors-be
```

สร้าง `.env` (ค่าสำหรับ production — ต่างจากตอน dev ที่ใช้ `123456`/placeholder secret):
```bash
nano .env
```
```
DB_HOST=localhost
DB_PORT=5555
DB_USER=logmon
DB_PASSWORD=<เหมือนกับใน docker-compose.yml>
DB_NAME=logmon_db
REDIS_ADDR=localhost:6380
JWT_SECRET=<openssl rand -base64 32>
JWT_EXPIRY_HOUR=2
APP_PORT=3000
```

Build แล้วลองรันตรงๆ ก่อนเพื่อทดสอบ:
```bash
go build -o log-monitor-server ./cmd/...
./log-monitor-server        # กด Ctrl+C หลังเช็คว่า start สำเร็จไม่มี error
```

สร้าง systemd unit file:
```bash
nano /etc/systemd/system/log-monitor.service
```
```ini
[Unit]
Description=Log Monitor Backend
After=network.target docker.service

[Service]
Type=simple
User=root
WorkingDirectory=/root/log-monitors-be
ExecStart=/root/log-monitors-be/log-monitor-server
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
systemctl daemon-reload
systemctl enable log-monitor
systemctl start log-monitor
systemctl status log-monitor   # ควรขึ้น "active (running)"
```

### Database migrations (ทำแค่ครั้งแรกที่ deploy)

```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xvz
mv migrate /usr/local/bin/migrate

# URL-encode อักขระพิเศษ (เช่น "/" -> "%2F") ใน password ของ DB
migrate -path pkg/databases/migrations \
  -database "postgres://logmon:<password ที่ encode แล้ว>@localhost:5555/logmon_db?sslmode=disable" up
```

## 4. Frontend — build สำหรับ production

```bash
curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
apt install -y nodejs

git clone https://github.com/WhoaWicked/log-monitors-fe.git log-monitor-frontend
cd log-monitor-frontend
npm install
```

สร้าง `.env.production`:
```bash
nano .env.production
```
```
VITE_API_URL=
VITE_WS_URL=wss://YOUR_DOMAIN
```
- `VITE_API_URL` ปล่อย**ว่าง** → frontend เรียก relative path (`/logs`, `/login`)
  ซึ่งจะไปที่ origin เดียวกับที่ Nginx เสิร์ฟอยู่ — ไม่ต้องยุ่งกับ CORS และไม่ต้อง build ใหม่ถ้า IP/domain เปลี่ยน
- `VITE_WS_URL` ต้องเป็น **URL เต็ม** เสมอ (`WebSocket()` ไม่รองรับ relative path)
  และต้องใช้ `wss://` เมื่อเปิด HTTPS แล้ว ไม่งั้น browser จะบล็อกเพราะ mixed content

> เช็คให้ครบว่าไฟล์ `src/api/*.js` และ WebSocket hook ทุกไฟล์ใช้ `import.meta.env.VITE_API_URL` /
> `VITE_WS_URL` แทนที่จะ hardcode `localhost:3000` — ไล่เช็คทุกไฟล์ให้ดี พลาดไฟล์เดียวก็เจอปัญหา
> ได้ง่ายๆ (เจอจริงตอน deploy รอบอ้างอิง เสียเวลา debug ไปพอสมควร)

```bash
npm run build
mkdir -p /var/www/log-monitor
cp -r dist/* /var/www/log-monitor/
```
> อย่าให้ Nginx ชี้ไปที่ `/root/...` — `/root` ถูกล็อกไว้ไม่ให้ user `www-data` ของ Nginx
> เข้าถึงได้ (จะได้ error 500) ให้เสิร์ฟจาก `/var/www/` แทน

## 5. Nginx — reverse proxy

```bash
apt install -y nginx
nano /etc/nginx/sites-available/log-monitor
```
```nginx
server {
    listen 80;
    server_name YOUR_DOMAIN;

    root /var/www/log-monitor;
    index index.html;

    location /login { proxy_pass http://localhost:3000; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; }
    location /logs { proxy_pass http://localhost:3000; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; }
    location /alert-rules { proxy_pass http://localhost:3000; proxy_set_header Host $host; proxy_set_header X-Real-IP $remote_addr; }

    location /ws {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }

    location / {
        try_files $uri $uri/ /index.html;   # ให้ client-side routing (React Router) จัดการ path ที่ไม่รู้จักเอง
    }
}
```

```bash
ln -s /etc/nginx/sites-available/log-monitor /etc/nginx/sites-enabled/
rm /etc/nginx/sites-enabled/default
nginx -t
systemctl restart nginx
```

## 6. HTTPS — Let's Encrypt

Let's Encrypt ต้องการ domain จริง ไม่รับ IP ตรงๆ ถ้าไม่มี domain ของตัวเอง ใช้บริการ
free wildcard DNS ที่ทำให้ `<ip>.nip.io` ชี้ไปที่ IP นั้นได้เลยโดยไม่ต้องสมัครอะไร:
```
YOUR_IP.nip.io  →  resolve ไปที่ YOUR_IP
```

แก้ `server_name` ใน Nginx config ให้เป็น domain นี้ก่อน แล้วรัน:
```bash
apt install -y certbot python3-certbot-nginx
certbot --nginx -d YOUR_IP.nip.io
```
Certbot จะแก้ Nginx config ให้เองอัตโนมัติ (เพิ่ม `listen 443 ssl`, redirect
80→443) และตั้ง auto-renewal ให้ด้วย (certificate หมดอายุทุก 90 วัน)

**หลังจากนี้ ต้องอัปเดต `VITE_WS_URL` ใน `.env.production` ให้ใช้ domain (ไม่ใช่ IP)
แล้ว build frontend ใหม่** — SSL cert ออกให้กับ domain เท่านั้น เพราะฉะนั้น `wss://`
ที่ยิงไปที่ IP ตรงๆ จะไม่ทำงาน ถึงแม้ `https://` ของตัวเว็บเองจะใช้ได้ปกติก็ตาม

```bash
cd ~/log-monitor-frontend
nano .env.production   # VITE_WS_URL=wss://YOUR_IP.nip.io
npm run build
cp -r dist/* /var/www/log-monitor/
```

---

## ปัญหาที่เจอตอน deploy จริง (และวิธีแก้)

- **`go build` / `npm install` ค้าง VPS ไม่ตอบสนอง** — RAM เต็มบน droplet 1GB
  ต้องเพิ่ม swap (ขั้น 1) **ก่อน** build อะไรที่หนัก ถ้าเกิดไปแล้ว: DigitalOcean
  dashboard → droplet → Actions → **Restart** (force restart แบบ hard reboot
  ใช้ได้แม้ SSH เข้าไม่ได้แล้ว)
- **`bind: address already in use`** — มี process เก่าของ binary ยังจับ port ค้างอยู่
  ใช้ `pkill -9 -f log-monitor-server` (`pkill -9 log-monitor-server` เฉยๆ
  จะล้มเหลวเงียบๆ ถ้าชื่อยาวเกิน 15 ตัวอักษร — ต้องมี `-f`)
- **Nginx ขึ้น 500, error log บอก `Permission denied`** — เสิร์ฟไฟล์ frontend จาก
  `/root/...` ซึ่ง user `www-data` ของ Nginx เข้าไม่ถึง ย้ายไฟล์ `dist/` ไปไว้ที่ `/var/www/` แทน
- **Connection string ของ DB parse ไม่ผ่าน** — มี `/` (หรืออักขระพิเศษอื่น) อยู่ใน
  password ที่ generate มา ทำให้ URL `postgres://` แบบธรรมดา parse ผิด ต้อง
  URL-encode ก่อน (`/` → `%2F`)
- **Frontend ยังยิงไป `localhost:3000` อยู่ทั้งที่ deploy แล้ว** — ปกติมาจากสาเหตุใดสาเหตุหนึ่ง:
  ยังไม่ได้สร้าง `.env.production` บน VPS (ไฟล์ env ไม่ได้ push ขึ้น git),
  มีไฟล์ที่ยังเขียน URL แบบ hardcode แทนที่จะใช้ `import.meta.env`, commit การแก้ไข
  ไปอยู่ที่ feature branch แทน `main`, หรือลืม `git push` ก่อนไป `git pull` บน VPS
  เช็ค `git branch` / `git status` ทั้งสองฝั่งให้ดี
- **WebSocket ต่อผ่าน `wss://` ไม่ติด** — certificate ออกให้กับ domain ไม่ใช่ IP
  พอเปิด HTTPS แล้ว `VITE_WS_URL` ต้องใช้ domain ด้วย

## ค่าใช้จ่าย

- Droplet: ~$6/เดือน (คิดเป็นรายชั่วโมง ~$0.009/ชม.) ช่วงแรกใช้ signup credit
  ของ DigitalOcean ก่อนได้สักพัก
- Domain (nip.io) + SSL (Let's Encrypt): ฟรี
- **ถ้าอยากหยุดเสียค่าใช้จ่ายทั้งหมด**: destroy droplet ทิ้งจาก DigitalOcean
  dashboard ได้เลย ไม่มีอะไรหายไปไหน — โค้ดอยู่บน GitHub ครบ และถ้าทำตามคู่มือนี้
  ซ้ำบน droplet ใหม่ จะใช้เวลาน้อยกว่าตอน deploy ครั้งแรกมาก
