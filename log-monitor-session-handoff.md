# Log Monitoring Project — Session Handoff

> ใช้ไฟล์นี้เปิด session ใหม่ พร้อมกับ `log-monitor-requirements.md`
> (มี requirement เต็มอยู่ในไฟล์นั้นแล้ว ไฟล์นี้สรุปว่า "ทำถึงไหนแล้ว" กับ "ต่อไปทำอะไร")

---

## สถานะตอนนี้: ฝึกพื้นฐานครบแล้ว ยังไม่เริ่มประกอบเป็นระบบจริง

พื้นฐานทุกตัวที่ต้องใช้ในโปรเจกต์ ฝึกแยกจนคล่องแล้ว เหลือแค่เอามาประกอบเป็นระบบเดียวกัน

## Background ผู้เรียน

- Go beginner — ไม่เคยสร้างโปรเจกต์ Go เองมาก่อน มีแต่ตามคอร์สทำ
- ไม่เคยใช้ WebSocket มาก่อน
- มีพื้นฐาน Node.js/Express/Next.js/TypeScript แน่น (ใช้เทียบ concept บ่อย)
- สไตล์การเรียนที่ใช้ได้ผลดี: **อธิบาย concept ก่อน → ให้เขียนโค้ดเอง → แก้ทีละจุด** ไม่ให้เฉลยเต็มทันที ถามซ้ำจนเข้าใจจริงถึงจะไปต่อ

## สิ่งที่ฝึกจบแล้ว (คล่องแล้ว ไม่ต้องสอนใหม่)

**Go concurrency** (ฝึกผ่าน `url-checker` CLI):
- goroutine, WaitGroup (Add/Done/Wait), channel (buffered/unbuffered)
- race condition (เข้าใจลึกระดับเล่าเหตุการณ์ G1/G2 แย่งเขียนตัวแปรได้เอง)
- deadlock (เข้าใจว่าเกิดจากรอกันเป็นวงกลม ไม่ใช่แค่ "รอนาน")
- `context.WithTimeout`, `context.WithCancel` (parent-child ซ้อนกันได้)
- worker pool (จำกัดจำนวน goroutine, trade-off ความเร็ว vs resource)
- `sync.Mutex`, `sync/atomic` (นับ counter ปลอดภัย)

**WebSocket** (ฝึกผ่าน echo server + Hub):
- `gorilla/websocket` + Gin: `upgrader.Upgrade`, `ReadMessage`/`WriteMessage`
- Hub pattern: `map[*websocket.Conn]bool` + Mutex, broadcast ไปหลาย client พร้อมกัน
- ทดสอบผ่าน Postman WebSocket mode สำเร็จ (เห็น broadcast ทำงานจริงกับ 2 แท็บ)

**Database**:
- Postgres ผ่าน Docker (`postgres:alpine`, port `4444`)
- `golang-migrate` — สร้าง/รัน migration แล้ว 1 ตาราง (`logs`)
- `sqlx` + `pgx` — connect, insert, query กลับเป็น struct (ครบ flow แล้ว)
- Redis ผ่าน Docker (`redis:alpine`, port `6380`) — connect, `Set`/`Get`, เข้าใจ expiration

## การตัดสินใจสำคัญที่ล็อกไว้แล้ว (อย่าเสนอใหม่)

- Backend: **Go + Gin** (ไม่ใช่ Fiber — market share)
- Database: **Postgres + Redis ผ่าน Docker local** (เคยคิดจะใช้ Supabase/Upstash แต่เปลี่ยนใจกลับมา Docker แล้ว)
- ORM: ไม่ใช้ ORM เต็มรูปแบบ — ใช้ `sqlx` + `pgx` (เขียน SQL เอง)
- Auth: **เขียน JWT เอง** (`golang-jwt/jwt` + `bcrypt`) ไม่พึ่ง Firebase/service ภายนอก, ไม่มีหน้า signup (seed admin ใน migration)
- Frontend: Next.js + Tailwind (ของเดิมที่ถนัด)
- **ตัดออกจาก scope โดยตั้งใจ**: Kafka/RabbitMQ, RBAC หลาย role, email/Slack notification, cloud deploy (ยังไม่ต้องคิดตอนนี้)
- อยากให้ deploy ได้จริงตอนท้าย (ยังไม่ต้องวางแผนตอนนี้)

## งานที่ค้างอยู่ ยังไม่ได้ทำ

1. **Migration อีก 2 ตาราง** — `alert_rules` (service, level, threshold, window) และ `users` (email, password hash) — ยังไม่ได้ออกแบบ schema เลย
2. **วาง project structure** — มีร่างไว้แบบ domain-vertical:
   ```
   logmon/
   ├── cmd/main.go
   ├── internal/
   │   ├── hub/          # มีโค้ดแล้ว (จาก echo server)
   │   ├── db/            # มีโค้ด connect แล้ว (จากฝึก sqlx)
   │   ├── cache/          # มีโค้ด connect แล้ว (จากฝึก redis)
   │   ├── models/         # มี Log struct แล้ว
   │   ├── ingest/         # ยังไม่เขียน — worker pool รับ log event จริง
   │   └── handler/        # ยังไม่เขียน — ผูก hub เข้ากับ handler
   └── pkg/databases/migrations/  # มี migration logs table แล้ว
   ```
   ยังไม่ได้ย้ายไฟล์จริง โค้ดตอนนี้กระจัดกระจายอยู่ในไฟล์ทดลองแยกๆ
3. **`Server` struct** ต้องขยายจาก `{ hub *hub.Hub }` ให้มี `db`, `redis` ด้วย — ยังไม่ได้แก้
4. **Ingestion + worker pool จริง** — เอา pattern จาก `url-checker` มาปรับกับ log event (ยังไม่เริ่มเขียนเลย)
5. ทุกอย่างในไฟล์ requirement ข้อ 4 (Functional Requirements) เป็น checklist เปล่าทั้งหมด ยังไม่ติ๊กอะไรเลย

## ย้ำเรื่อง benchmark (ห้ามลืม)

Non-functional requirement ข้อ throughput/latency/`go test -race` **ห้ามตัดออกจาก scope** เพราะเป็นสิ่งที่ทำให้ resume bullet มีตัวเลขจริง ต่างจากโปรเจกต์เดิม (Equipment Borrowing) ที่ bullet point เป็นแค่ adjective ลอยๆ

## จุดเริ่มต้นที่แนะนำสำหรับ session ใหม่

เริ่มจากข้อ 1 (ออกแบบ + รัน migration ของ `alert_rules` และ `users`) แล้วค่อยไปข้อ 2 (วาง project structure จริง ย้ายโค้ด)
