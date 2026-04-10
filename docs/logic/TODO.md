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

* [x] tentukan stack final backend
* [x] buat struktur project repo
* [x] setup config environment
* [x] setup database connection
* [x] setup migration system
* [x] setup basic HTTP server
* [x] setup base routing
* [x] setup standard response format
* [x] setup error handling dasar
* [x] setup logger dasar
* [x] buat docs folder untuk logic project

### Output

* project bisa dijalankan local
* koneksi database hidup
* migration bisa jalan
* server punya health/basic route

## Phase 1 — Database Schema & Domain Model

### Goal

Menyiapkan schema inti sesuai business flow v1.

### Tasks

* [x] buat migration table `admins`
* [x] buat migration table `appointments`
* [x] buat migration table `notification_logs`
* [x] tentukan index yang relevan
* [x] pastikan field inti sesuai pondasi project
* [x] pastikan `duration_minutes` default 120
* [x] pastikan `cancelled_at` nullable
* [x] buat model/entity internal
* [x] pastikan `start_at` dan `end_at` jadi source utama logic waktu

### Output

* schema database final v1 siap dipakai
* model domain inti sudah ada

## Phase 2 — Auth Admin

### Goal

Admin bisa login ke sistem.

### Tasks

* [x] buat seed / cara insert admin awal
* [x] implement login admin
* [x] validasi email dan password
* [x] hashing password
* [x] generate auth token / session strategy sederhana
* [x] middleware auth untuk endpoint private
* [x] buat endpoint `POST /auth/login`
* [x] buat test minimal login success / failed

### Output

* admin bisa login
* endpoint private bisa diamankan

## Phase 3 — Create Appointment

### Goal

Admin bisa menambah appointment baru dengan rule v1.

### Tasks

* [x] buat request DTO create appointment
* [x] validasi field wajib
* [x] parse `meeting_date` + `meeting_time` menjadi `start_at`
* [x] hitung `end_at = start_at + 120 menit`
* [x] set `duration_minutes = 120`
* [x] set status default `scheduled`
* [x] validasi reminder config
* [x] implement overlap checking
* [x] simpan appointment
* [x] buat endpoint `POST /appointments`
* [x] buat test create success
* [x] buat test create ditolak jika overlap

### Output

* appointment baru bisa dibuat
* overlap sudah dicegah dari awal

## Phase 4 — Read/List Appointment

### Goal

Admin bisa melihat daftar dan detail appointment.

### Tasks

* [x] implement endpoint `GET /appointments`
* [x] filter by tanggal
* [x] filter by status
* [x] sorting default by `start_at`
* [x] implement endpoint `GET /appointments/:id`
* [x] mapping response detail
* [x] fallback compute status saat read/list bila perlu
* [x] pastikan appointment cancelled tetap bisa dibaca di admin area jika dibutuhkan
* [x] buat test list dan detail

### Output

* admin bisa lihat daftar jadwal
* admin bisa lihat detail jadwal

## Phase 5 — Update Appointment

### Goal

Admin bisa mengubah appointment tanpa merusak rule inti.

### Tasks

* [x] implement endpoint `PUT /appointments/:id`
* [x] validasi field update
* [x] recompute `start_at` dan `end_at` bila waktu diubah
* [x] overlap checking saat update
* [x] jaga `duration_minutes` tetap 120
* [x] jaga agar status valid tetap konsisten
* [x] aturan edit cancelled appointment ditentukan jelas
* [x] buat test update success
* [x] buat test update ditolak jika overlap

### Output

* appointment bisa diedit dengan aman
* rule no-overlap tetap terjaga

## Phase 6 — Cancel Appointment

### Goal

Admin bisa membatalkan appointment.

### Tasks

* [x] implement endpoint `PATCH /appointments/:id/cancel`
* [x] set status `cancelled`
* [x] isi `cancelled_at`
* [x] pastikan appointment cancelled tidak ikut public checker aktif
* [x] pastikan reminder berhenti untuk cancelled appointment
* [x] buat test cancel success
* [x] buat test cancel idempotency / duplicate cancel behavior

### Output

* appointment bisa dibatalkan dengan aman

## Phase 7 — Public Schedule Checker

### Goal

Public/client bisa melihat jadwal yang sudah terisi per tanggal.

### Tasks

* [x] implement endpoint `GET /public/schedules?date=YYYY-MM-DD`
* [x] validasi query `date`
* [x] ambil appointment aktif pada tanggal tersebut
* [x] exclude cancelled appointment
* [x] urutkan berdasarkan `start_at`
* [x] mapping response public hanya `start`, `end`, `status`
* [x] jangan tampilkan `client_name`, `address`, `notes`
* [x] buat test response kosong
* [x] buat test response berisi occupied ranges

### Output

* public checker berjalan sesuai scope v1

## Phase 8 — Reminder Worker

### Goal

Sistem bisa mengirim reminder appointment sebelum dimulai.

### Tasks

* [x] buat worker/scheduler sederhana
* [x] query appointment status `scheduled`
* [x] filter `is_reminder_enabled = true`
* [x] cek apakah sudah masuk `reminder_start_at`
* [x] cek interval reminder berdasarkan `reminder_interval_hours`
* [x] kirim reminder ke admin
* [x] simpan log ke `notification_logs`
* [x] cegah double-send pada slot yang sama
* [x] skip appointment dengan status `on_going`, `done`, `cancelled`
* [x] buat test / simulation reminder flow

### Output

* reminder berjalan sesuai config appointment

## Phase 9 — Auto Status Update

### Goal

Status appointment konsisten terhadap waktu.

### Tasks

* [x] buat worker untuk update status berkala
* [x] ubah `scheduled -> on_going` saat masuk waktu mulai
* [x] ubah `on_going -> done` saat lewat `end_at`
* [x] skip `cancelled`
* [x] tambah fallback compute status saat read/list bila worker telat
* [x] buat test transisi status

### Output

* status appointment konsisten baik dari worker maupun fallback read logic

## Phase 10 — Hardening

### Goal

Merapikan backend agar siap dipakai untuk development lanjut.

### Tasks

* [x] rapikan struktur service/repository/handler
* [x] rapikan validation error messages
* [x] rapikan auth middleware
* [x] tambahkan pagination jika benar-benar perlu
* [x] tambahkan request logging dasar
* [x] tambahkan integration test penting
* [x] review naming dan consistency field
* [x] rapikan README API repo
* [x] sinkronkan README, RULE, TODO

Catatan:

* pagination belum ditambahkan karena belum benar-benar diperlukan untuk list v1 saat ini

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

* [x] admin bisa login
* [x] admin bisa create appointment
* [x] admin bisa list appointment
* [x] admin bisa lihat detail appointment
* [x] admin bisa update appointment
* [x] admin bisa cancel appointment
* [x] overlap appointment ditolak
* [x] public checker per tanggal berjalan
* [x] reminder berjalan untuk appointment scheduled
* [x] status otomatis berubah sesuai waktu
* [x] logic tetap konsisten dengan pondasi v1

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
