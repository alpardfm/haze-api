# TODO.md

## Backend API Roadmap

Dokumen ini membagi pengerjaan backend API ke dalam phase yang realistis untuk v1.

Prinsip utama:

* utamakan fondasi domain dan business logic
* backend dulu, UI belakangan
* jangan lompat ke fitur phase lanjut
* setiap phase harus menghasilkan progress yang benar-benar bisa dipakai

## Phase 0 — Foundation Setup

### Goal

Membuat pondasi project backend yang siap dikembangkan.

### Tasks

* [ ] tentukan stack final backend
* [ ] buat struktur project repo
* [ ] setup config environment
* [ ] setup database connection
* [ ] setup migration system
* [ ] setup basic HTTP server
* [ ] setup base routing
* [ ] setup standard response format
* [ ] setup error handling dasar
* [ ] setup logger dasar
* [ ] buat docs folder untuk logic project

### Output

* project bisa dijalankan local
* koneksi database hidup
* migration bisa jalan
* server punya health/basic route

## Phase 1 — Database Schema & Domain Model

### Goal

Menyiapkan schema inti sesuai business flow v1.

### Tasks

* [ ] buat migration table `admins`
* [ ] buat migration table `appointments`
* [ ] buat migration table `notification_logs`
* [ ] tentukan index yang relevan
* [ ] pastikan field inti sesuai pondasi project
* [ ] pastikan `duration_minutes` default 120
* [ ] pastikan `cancelled_at` nullable
* [ ] buat model/entity internal
* [ ] pastikan `start_at` dan `end_at` jadi source utama logic waktu

### Output

* schema database final v1 siap dipakai
* model domain inti sudah ada

## Phase 2 — Auth Admin

### Goal

Admin bisa login ke sistem.

### Tasks

* [ ] buat seed / cara insert admin awal
* [ ] implement login admin
* [ ] validasi email dan password
* [ ] hashing password
* [ ] generate auth token / session strategy sederhana
* [ ] middleware auth untuk endpoint private
* [ ] buat endpoint `POST /auth/login`
* [ ] buat test minimal login success / failed

### Output

* admin bisa login
* endpoint private bisa diamankan

## Phase 3 — Create Appointment

### Goal

Admin bisa menambah appointment baru dengan rule v1.

### Tasks

* [ ] buat request DTO create appointment
* [ ] validasi field wajib
* [ ] parse `meeting_date` + `meeting_time` menjadi `start_at`
* [ ] hitung `end_at = start_at + 120 menit`
* [ ] set `duration_minutes = 120`
* [ ] set status default `scheduled`
* [ ] validasi reminder config
* [ ] implement overlap checking
* [ ] simpan appointment
* [ ] buat endpoint `POST /appointments`
* [ ] buat test create success
* [ ] buat test create ditolak jika overlap

### Output

* appointment baru bisa dibuat
* overlap sudah dicegah dari awal

## Phase 4 — Read/List Appointment

### Goal

Admin bisa melihat daftar dan detail appointment.

### Tasks

* [ ] implement endpoint `GET /appointments`
* [ ] filter by tanggal
* [ ] filter by status
* [ ] sorting default by `start_at`
* [ ] implement endpoint `GET /appointments/:id`
* [ ] mapping response detail
* [ ] fallback compute status saat read/list bila perlu
* [ ] pastikan appointment cancelled tetap bisa dibaca di admin area jika dibutuhkan
* [ ] buat test list dan detail

### Output

* admin bisa lihat daftar jadwal
* admin bisa lihat detail jadwal

## Phase 5 — Update Appointment

### Goal

Admin bisa mengubah appointment tanpa merusak rule inti.

### Tasks

* [ ] implement endpoint `PUT /appointments/:id`
* [ ] validasi field update
* [ ] recompute `start_at` dan `end_at` bila waktu diubah
* [ ] overlap checking saat update
* [ ] jaga `duration_minutes` tetap 120
* [ ] jaga agar status valid tetap konsisten
* [ ] aturan edit cancelled appointment ditentukan jelas
* [ ] buat test update success
* [ ] buat test update ditolak jika overlap

### Output

* appointment bisa diedit dengan aman
* rule no-overlap tetap terjaga

## Phase 6 — Cancel Appointment

### Goal

Admin bisa membatalkan appointment.

### Tasks

* [ ] implement endpoint `PATCH /appointments/:id/cancel`
* [ ] set status `cancelled`
* [ ] isi `cancelled_at`
* [ ] pastikan appointment cancelled tidak ikut public checker aktif
* [ ] pastikan reminder berhenti untuk cancelled appointment
* [ ] buat test cancel success
* [ ] buat test cancel idempotency / duplicate cancel behavior

### Output

* appointment bisa dibatalkan dengan aman

## Phase 7 — Public Schedule Checker

### Goal

Public/client bisa melihat jadwal yang sudah terisi per tanggal.

### Tasks

* [ ] implement endpoint `GET /public/schedules?date=YYYY-MM-DD`
* [ ] validasi query `date`
* [ ] ambil appointment aktif pada tanggal tersebut
* [ ] exclude cancelled appointment
* [ ] urutkan berdasarkan `start_at`
* [ ] mapping response public hanya `start`, `end`, `status`
* [ ] jangan tampilkan `client_name`, `address`, `notes`
* [ ] buat test response kosong
* [ ] buat test response berisi occupied ranges

### Output

* public checker berjalan sesuai scope v1

## Phase 8 — Reminder Worker

### Goal

Sistem bisa mengirim reminder appointment sebelum dimulai.

### Tasks

* [ ] buat worker/scheduler sederhana
* [ ] query appointment status `scheduled`
* [ ] filter `is_reminder_enabled = true`
* [ ] cek apakah sudah masuk `reminder_start_at`
* [ ] cek interval reminder berdasarkan `reminder_interval_hours`
* [ ] kirim reminder ke admin
* [ ] simpan log ke `notification_logs`
* [ ] cegah double-send pada slot yang sama
* [ ] skip appointment dengan status `on_going`, `done`, `cancelled`
* [ ] buat test / simulation reminder flow

### Output

* reminder berjalan sesuai config appointment

## Phase 9 — Auto Status Update

### Goal

Status appointment konsisten terhadap waktu.

### Tasks

* [ ] buat worker untuk update status berkala
* [ ] ubah `scheduled -> on_going` saat masuk waktu mulai
* [ ] ubah `on_going -> done` saat lewat `end_at`
* [ ] skip `cancelled`
* [ ] tambah fallback compute status saat read/list bila worker telat
* [ ] buat test transisi status

### Output

* status appointment konsisten baik dari worker maupun fallback read logic

## Phase 10 — Hardening

### Goal

Merapikan backend agar siap dipakai untuk development lanjut.

### Tasks

* [ ] rapikan struktur service/repository/handler
* [ ] rapikan validation error messages
* [ ] rapikan auth middleware
* [ ] tambahkan pagination jika benar-benar perlu
* [ ] tambahkan request logging dasar
* [ ] tambahkan integration test penting
* [ ] review naming dan consistency field
* [ ] rapikan README API repo
* [ ] sinkronkan README, RULE, TODO

### Output

* backend v1 lebih stabil dan enak dilanjutkan

## Suggested Build Order

Urutan implementasi yang disarankan:

1. phase 0
2. phase 1
3. phase 2
4. phase 3
5. phase 4
6. phase 5
7. phase 6
8. phase 7
9. phase 8
10. phase 9
11. phase 10

## Acceptance Checklist V1

Project backend v1 dianggap cukup ketika:

* [ ] admin bisa login
* [ ] admin bisa create appointment
* [ ] admin bisa list appointment
* [ ] admin bisa lihat detail appointment
* [ ] admin bisa update appointment
* [ ] admin bisa cancel appointment
* [ ] overlap appointment ditolak
* [ ] public checker per tanggal berjalan
* [ ] reminder berjalan untuk appointment scheduled
* [ ] status otomatis berubah sesuai waktu
* [ ] logic tetap konsisten dengan pondasi v1

## Important Reminder

Jangan menambah fitur berikut ke TODO v1 kecuali ada keputusan baru:

* booking public
* approval flow
* client account
* availability absolut
* custom meeting duration
* WhatsApp bot
* analytics besar
* multi admin scheduling kompleks
