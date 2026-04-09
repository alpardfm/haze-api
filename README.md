# README.md

# Admin Schedule Reminder API

Backend API untuk sistem jadwal admin + reminder.

Project ini fokus pada pengelolaan appointment oleh admin, reminder sebelum jadwal dimulai, auto status berdasarkan waktu, dan public checker untuk melihat jadwal yang sudah terisi pada tanggal tertentu.

## 1. Tujuan Project

Tujuan utama backend ini:

* membantu admin mengelola jadwal pertemuan secara terstruktur
* mencegah bentrok jadwal
* mengirim reminder ke admin sebelum appointment dimulai
* mengubah status appointment otomatis berdasarkan waktu
* menyediakan endpoint public untuk melihat rentang waktu yang sudah terisi

## 2. Scope V1

### Included

* admin login
* create appointment
* update appointment
* cancel appointment
* list appointment
* appointment detail
* reminder sebelum jadwal dimulai
* auto status `scheduled -> on_going -> done`
* public checker berdasarkan tanggal
* validasi overlap appointment

### Excluded

* booking online
* approval booking
* client login
* WhatsApp bot
* jam operasional admin
* availability absolut
* multi admin complex assignment
* analytics/dashboard kompleks
* custom duration meeting

## 3. Core Business Rules

### Appointment

* setiap jadwal adalah `appointment`
* admin bebas menentukan jam mulai
* durasi appointment v1 adalah **fix 120 menit / 2 jam**
* `start_at` dan `end_at` adalah source utama logic waktu
* `end_at = start_at + 120 menit`

### Status

Status yang digunakan:

* `scheduled`
* `on_going`
* `done`
* `cancelled`

Rule perubahan status:

* sebelum `start_at` = `scheduled`
* saat `start_at <= now < end_at` = `on_going`
* saat `now >= end_at` = `done`
* jika dibatalkan admin = `cancelled`

### Reminder

Reminder hanya berjalan jika:

* `is_reminder_enabled = true`
* status appointment masih `scheduled`

Reminder dikonfigurasi per appointment:

* `is_reminder_enabled`
* `reminder_start_at`
* `reminder_interval_hours`

Reminder berhenti jika status menjadi:

* `on_going`
* `done`
* `cancelled`

### Overlap Rule

Appointment tidak boleh overlap.

Jika ada appointment lain yang waktunya bertabrakan, proses create/update harus ditolak.

### Public Checker

Public checker hanya menampilkan:

* daftar rentang waktu yang sudah terisi pada tanggal tertentu

Public checker **tidak** boleh:

* menyatakan availability absolut
* menampilkan data sensitif internal appointment

Jika tidak ada jadwal pada tanggal tersebut, tampilkan response kosong dan frontend dapat merender:

`Belum ada jadwal tercatat pada tanggal ini`

## 4. Domain Model

### `admins`

Menyimpan akun admin.

Field utama:

* `id`
* `name`
* `email`
* `phone`
* `password_hash`
* `created_at`
* `updated_at`

### `appointments`

Entity utama jadwal.

Field utama:

* `id`
* `client_name`
* `address`
* `notes`
* `meeting_date`
* `meeting_time`
* `duration_minutes`
* `start_at`
* `end_at`
* `status`
* `is_reminder_enabled`
* `reminder_start_at`
* `reminder_interval_hours`
* `created_by_admin_id`
* `created_at`
* `updated_at`
* `cancelled_at`

Catatan:

* `duration_minutes` default `120`
* `start_at` dan `end_at` adalah source of truth

### `notification_logs`

Log reminder yang sudah dikirim.

Field utama:

* `id`
* `appointment_id`
* `notification_type`
* `scheduled_for`
* `sent_at`
* `recipient`
* `status`
* `message`
* `created_at`

## 5. API Endpoints V1

### Auth

#### `POST /auth/login`

Login admin.

### Appointments

#### `GET /appointments`

Ambil list appointment admin.

Optional query:

* `date`
* `status`

#### `POST /appointments`

Buat appointment baru.

#### `GET /appointments/:id`

Ambil detail appointment.

#### `PUT /appointments/:id`

Update appointment.

#### `PATCH /appointments/:id/cancel`

Batalkan appointment.

### Public Checker

#### `GET /public/schedules?date=YYYY-MM-DD`

Ambil daftar jadwal terisi pada tanggal tertentu.

Contoh response:

```json
{
  "date": "2026-04-12",
  "items": [
    {
      "start": "09:30",
      "end": "11:30",
      "status": "occupied"
    },
    {
      "start": "13:00",
      "end": "15:00",
      "status": "occupied"
    }
  ]
}
```

Kalau kosong:

```json
{
  "date": "2026-04-12",
  "items": []
}
```

## 6. Suggested Project Structure

```bash
.
├── cmd/
│   └── api/
├── internal/
│   ├── auth/
│   ├── appointment/
│   ├── notification/
│   ├── publicschedule/
│   ├── worker/
│   ├── shared/
│   └── config/
├── migrations/
├── docs/
│   └── logic/
│       ├── RULE.md
│       └── TODO.md
├── scripts/
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

Struktur final boleh menyesuaikan stack, tapi business flow harus tetap sama.

## 7. Suggested Tech Direction

Backend v1 cukup sederhana:

* REST API
* relational database: PostgreSQL atau MySQL
* worker / scheduler untuk reminder dan auto status
* auth admin sederhana berbasis login

Prinsip teknis:

* business logic harus mudah dibaca
* hindari overengineering
* prioritaskan flow yang stabil dulu
* siap dipakai AI agent untuk vibe coding

## 8. Minimal Flow

### Create Appointment

1. admin login
2. kirim request create appointment
3. backend validasi payload
4. backend hitung `start_at` dan `end_at`
5. backend cek overlap
6. backend simpan data dengan status `scheduled`

### Reminder Worker

1. worker scan appointment status `scheduled`
2. cek apakah reminder aktif
3. cek apakah sudah masuk waktu reminder
4. kirim reminder
5. simpan `notification_logs`

### Auto Status Worker / Fallback Read Logic

1. cek waktu saat ini
2. jika sudah masuk interval meeting, status dianggap `on_going`
3. jika lewat `end_at`, status dianggap `done`
4. jika `cancelled`, jangan diproses lagi

## 9. Non Goals

Hal-hal berikut sengaja tidak dibangun di v1:

* booking slot oleh client
* approval/reject booking
* public availability absolut
* custom durasi meeting
* multi admin scheduling yang kompleks
* reminder channel kompleks seperti WhatsApp bot

## 10. Development Principle

Saat mengembangkan project ini:

* jaga konsistensi dengan RULE.md
* hindari menambah fitur di luar scope v1
* utamakan domain clarity daripada banyak fitur
* prioritaskan implementasi yang realistis dan mudah di-maintain

## 11. Next Docs

Dokumen pendukung:

* `docs/logic/RULE.md` → aturan bisnis dan guardrail AI agent
* `docs/logic/TODO.md` → phase pengerjaan backend API

---