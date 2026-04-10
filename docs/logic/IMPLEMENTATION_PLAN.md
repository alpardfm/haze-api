# IMPLEMENTATION_PLAN.md

## Purpose

Dokumen ini menerjemahkan `README.md`, `RULE.md`, dan `TODO.md` menjadi rencana implementasi teknis yang langsung bisa dipakai untuk membangun backend v1.

Dokumen ini bukan pengganti source of truth. Jika ada konflik, ikuti urutan prioritas berikut:

1. `docs/logic/RULE.md`
2. `README.md`
3. `docs/logic/TODO.md`
4. `docs/logic/IMPLEMENTATION_PLAN.md`

## Current Repo Audit

Status repo saat dokumen ini dibuat:

* project masih berada di fase dokumentasi awal
* belum ada struktur backend seperti `cmd/`, `internal/`, atau `migrations/`
* belum ada `go.mod`
* belum ada server HTTP
* belum ada database connection
* belum ada migration system
* `.gitignore` sudah mengarah ke project Go
* dokumen domain utama sudah tersedia di `README.md`, `docs/logic/RULE.md`, dan `docs/logic/TODO.md`

Kesimpulan:

* project siap masuk `Phase 0 - Foundation Setup`
* stack yang paling selaras dengan `.gitignore` dan struktur README adalah Go backend
* belum ada implementasi yang perlu di-refactor atau disinkronkan dengan rule v1

## Recommended V1 Technical Direction

Rekomendasi sederhana untuk v1:

* language: Go
* API style: REST
* database: PostgreSQL
* routing: Go standard library `net/http` atau router ringan
* migration: file SQL migration sederhana
* auth: login admin dengan token sederhana
* worker: scheduler sederhana dalam proses yang sama atau command terpisah
* config: environment variables
* response: JSON response konsisten

Prinsip implementasi:

* business logic appointment harus mudah dibaca
* `start_at` dan `end_at` tetap menjadi source of truth
* hindari abstraction besar sebelum dibutuhkan
* jangan menambah fitur di luar scope v1
* reminder dan auto status cukup sederhana untuk v1

## Domain Decisions To Lock Early

Keputusan berikut perlu ditetapkan sebelum atau saat implementasi phase terkait.

### Timezone

Gunakan timezone project yang konsisten untuk parse `meeting_date` + `meeting_time`.

Rekomendasi awal:

* gunakan `Asia/Jakarta` untuk local business logic
* simpan timestamp sebagai timestamptz di database jika memakai PostgreSQL
* response bisa mengembalikan format yang konsisten dari `start_at` dan `end_at`

### Cancelled Appointment Editing

Aturan ini belum final di dokumen utama.

Rekomendasi v1:

* appointment `cancelled` tidak boleh di-update kembali menjadi aktif
* jika perlu jadwal baru, admin membuat appointment baru
* `cancelled` tetap bisa dibaca di admin area

Alasan:

* menjaga status `cancelled` tidak kembali aktif otomatis
* menghindari flow `rescheduled` yang belum masuk scope v1

### Done Appointment Editing

Rekomendasi v1:

* appointment `done` tidak boleh diubah waktunya
* data historis tetap bisa dibaca

Alasan:

* mengurangi risiko perubahan history
* tetap sederhana untuk v1

### Reminder Delivery

Dokumen utama belum menentukan channel teknis reminder.

Rekomendasi v1:

* buat interface/service reminder sederhana
* implementasi awal boleh log-only atau email jika konfigurasi tersedia
* tetap simpan `notification_logs`
* jangan menambahkan WhatsApp bot atau multi-channel kompleks

## Suggested Project Structure

Struktur awal yang disarankan:

```text
.
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── appointment/
│   ├── auth/
│   ├── config/
│   ├── database/
│   ├── notification/
│   ├── publicschedule/
│   ├── shared/
│   └── worker/
├── migrations/
├── docs/
│   └── logic/
├── scripts/
├── go.mod
└── README.md
```

## Phase 0 Implementation Tasks

Target:

* backend bisa dijalankan lokal
* health route tersedia
* koneksi database bisa dites
* struktur awal siap untuk phase domain

Task teknis:

* buat `go.mod`
* buat `cmd/api/main.go`
* buat package config untuk membaca environment variables
* buat package database untuk koneksi PostgreSQL
* buat package shared response untuk JSON response standar
* buat package shared errors untuk error response dasar
* buat route `GET /health`
* buat script atau dokumentasi env lokal
* buat folder `migrations/`

Acceptance:

* `go test ./...` berjalan
* server bisa start lokal
* `GET /health` mengembalikan response sukses

## Phase 1 Implementation Tasks

Target:

* schema database inti siap dipakai
* domain model utama tersedia

Task teknis:

* buat migration `admins`
* buat migration `appointments`
* buat migration `notification_logs`
* tambahkan index untuk appointment date/time lookup
* tambahkan index untuk overlap query berbasis `start_at`, `end_at`, dan `status`
* buat model internal untuk admin
* buat model internal untuk appointment
* buat model internal untuk notification log
* pastikan `duration_minutes` default `120`
* pastikan `cancelled_at` nullable

Acceptance:

* migration bisa dijalankan dari database kosong
* field schema sesuai minimal entity rule
* tidak ada field domain yang keluar dari scope v1

## Phase 2 Implementation Tasks

Target:

* admin bisa login
* endpoint private bisa dilindungi

Task teknis:

* buat seed admin awal atau script insert admin
* implement hashing password
* implement `POST /auth/login`
* validasi email dan password
* generate token/session sederhana
* buat middleware auth
* buat test login success
* buat test login failed

Acceptance:

* credential valid menghasilkan token/session
* credential invalid ditolak
* endpoint private menolak request tanpa auth

## Phase 3 Implementation Tasks

Target:

* admin bisa create appointment sesuai rule v1

Task teknis:

* buat request DTO create appointment
* validasi `client_name`, `address`, `meeting_date`, dan `meeting_time`
* parse `meeting_date` + `meeting_time` menjadi `start_at`
* hitung `end_at = start_at + 120 menit`
* set `duration_minutes = 120`
* set status default `scheduled`
* validasi reminder config jika reminder aktif
* implement query overlap untuk appointment aktif
* buat endpoint `POST /appointments`
* buat test create success
* buat test overlap ditolak

Acceptance:

* appointment valid tersimpan
* appointment overlap ditolak
* custom duration tidak diterima

## Phase 4 Implementation Tasks

Target:

* admin bisa list dan detail appointment

Task teknis:

* implement `GET /appointments`
* support filter `date`
* support filter `status`
* sorting default by `start_at`
* implement `GET /appointments/:id`
* mapping response admin
* tambahkan fallback compute status saat read/list
* pastikan `cancelled` tidak otomatis berubah status
* buat test list
* buat test detail

Acceptance:

* appointment bisa dilihat oleh admin
* fallback status konsisten dengan waktu
* appointment `cancelled` tetap aman dibaca di admin area

## Phase 5 Implementation Tasks

Target:

* admin bisa update appointment tanpa merusak rule inti

Task teknis:

* implement `PUT /appointments/:id`
* validasi payload update
* recompute `start_at` dan `end_at` jika waktu berubah
* paksa `duration_minutes` tetap `120`
* overlap checking dengan exclude appointment saat ini
* tolak update appointment `cancelled`
* tolak update appointment `done` jika mengubah waktu
* buat test update success
* buat test update overlap ditolak
* buat test update cancelled ditolak

Acceptance:

* appointment aktif bisa diupdate
* no-overlap tetap terjaga
* status final seperti `cancelled` tidak kembali aktif

## Phase 6 Implementation Tasks

Target:

* admin bisa cancel appointment

Task teknis:

* implement `PATCH /appointments/:id/cancel`
* set status menjadi `cancelled`
* isi `cancelled_at`
* buat cancel idempotent atau return conflict secara konsisten
* pastikan cancelled tidak masuk public checker
* pastikan reminder worker skip cancelled
* buat test cancel success
* buat test duplicate cancel behavior

Acceptance:

* appointment cancelled tidak memblok jadwal baru
* appointment cancelled tidak menerima reminder
* behavior duplicate cancel jelas dan konsisten

## Phase 7 Implementation Tasks

Target:

* public checker menampilkan jadwal terisi per tanggal

Task teknis:

* implement `GET /public/schedules?date=YYYY-MM-DD`
* validasi query `date`
* query appointment aktif pada tanggal tersebut
* exclude `cancelled`
* sorting by `start_at`
* mapping response hanya `start`, `end`, `status`
* pastikan tidak ada `client_name`, `address`, atau `notes`
* buat test response kosong
* buat test occupied ranges

Acceptance:

* response kosong memakai `items: []`
* response berisi hanya occupied ranges
* public checker tidak menyatakan availability absolut

## Phase 8 Implementation Tasks

Target:

* reminder berjalan untuk appointment scheduled

Task teknis:

* buat worker/scheduler sederhana
* query appointment `scheduled`
* filter `is_reminder_enabled = true`
* cek `now >= reminder_start_at`
* cek interval berdasarkan `reminder_interval_hours`
* cegah double-send dengan `notification_logs`
* kirim reminder ke admin via implementasi v1
* simpan log notification
* skip `on_going`, `done`, dan `cancelled`
* buat test atau simulation reminder flow

Acceptance:

* reminder tidak dikirim sebelum waktunya
* reminder tidak double-send pada slot yang sama
* reminder berhenti saat appointment tidak lagi `scheduled`

## Phase 9 Implementation Tasks

Target:

* status appointment konsisten terhadap waktu

Task teknis:

* buat worker auto status berkala
* update `scheduled -> on_going` saat `start_at <= now < end_at`
* update `on_going -> done` saat `now >= end_at`
* skip `cancelled`
* pertahankan fallback compute saat read/list
* buat test transisi status

Acceptance:

* status berubah otomatis sesuai waktu
* worker telat tidak membuat response read/list inkonsisten
* `cancelled` tidak berubah otomatis

## Phase 10 Implementation Tasks

Target:

* backend v1 rapi dan siap dilanjutkan

Task teknis:

* review struktur service/repository/handler
* rapikan validation error messages
* rapikan auth middleware
* tambahkan pagination hanya jika list mulai butuh
* tambahkan request logging dasar
* tambahkan integration test untuk flow penting
* review naming field agar konsisten
* sinkronkan README, RULE, TODO, dan implementation plan

Acceptance:

* acceptance checklist v1 di `TODO.md` terpenuhi
* tidak ada scope creep ke fitur non-goal
* flow utama bisa dijalankan end-to-end

## Immediate Next Step

Langkah paling aman berikutnya:

1. implement `Phase 0 - Foundation Setup`
2. lock stack Go + PostgreSQL
3. buat server minimal dengan `GET /health`
4. siapkan config dan database connection
5. baru lanjut ke migration schema di Phase 1
