# RULE.md

## Project Rule: Admin Schedule Reminder API

Dokumen ini adalah pondasi tetap untuk semua pekerjaan AI agent pada project ini.

AI agent **wajib** menjaga output tetap konsisten dengan business flow v1, scope v1, dan domain model yang sudah disepakati.

## 1. Identity Project

Project ini adalah **sistem jadwal admin + reminder**.

Fokus v1:

* admin mengelola appointment lewat web
* sistem mengirim reminder sebelum jadwal dimulai
* status appointment berubah otomatis berdasarkan waktu
* public/client hanya dapat melihat jadwal yang sudah terisi pada tanggal tertentu
* booking online belum ada di v1

## 2. Mandatory Thinking Rule Before Answering

Sebelum memberikan output apa pun, AI agent harus selalu melakukan pengecekan berikut:

1. cek apakah output konsisten dengan business flow v1
2. cek apakah ada miss logic, benturan domain, atau scope creep
3. jika ada benturan, jelaskan dulu sebelum lanjut
4. jangan mengambil keputusan arsitektur besar tanpa dasar yang jelas dari konteks project
5. prioritaskan implementasi sederhana, realistis, dan cocok untuk v1
6. hasil harus siap dipakai untuk vibe coding

## 3. Scope Guardrail

### Included in V1

* admin login
* tambah jadwal
* edit jadwal
* cancel jadwal
* list jadwal admin
* detail jadwal
* reminder sebelum jadwal dimulai
* auto status `scheduled -> on_going -> done`
* public checker berdasarkan tanggal
* validasi bentrok jadwal

### Excluded from V1

* booking online
* approval booking
* client login
* WhatsApp bot
* jam operasional admin
* perhitungan availability absolut
* multi admin complex assignment
* analytics/dashboard kompleks
* custom durasi meeting

Jika output mulai masuk ke area excluded, AI agent harus menganggap itu sebagai **scope creep** dan menolaknya atau mengembalikannya ke bentuk yang sesuai v1.

## 4. Core Domain Rule

### Appointment is the center

Setiap jadwal adalah `appointment`.

Admin membuat appointment dengan data utama:

* client_name
* address
* notes
* meeting_date
* meeting_time
* reminder config

### Fixed duration

Untuk v1:

* `duration_minutes` default `120`
* jangan buat custom durasi
* jangan buat pilihan slot 30 menit / 1 jam / 90 menit

### Source of truth

Untuk semua logic waktu:

* `start_at` adalah source utama waktu mulai
* `end_at` adalah source utama waktu selesai
* `meeting_date` dan `meeting_time` hanya pendukung untuk form/UI/readability

AI agent tidak boleh memindahkan source of truth ke field lain tanpa alasan yang sangat kuat.

## 5. Status Rule

Status yang valid hanya:

* `scheduled`
* `on_going`
* `done`
* `cancelled`

Jangan menambah status lain seperti:

* `pending`
* `approved`
* `rejected`
* `rescheduled`
* `expired`

### Status transition

* sebelum `start_at` => `scheduled`
* saat `start_at <= now < end_at` => `on_going`
* saat `now >= end_at` => `done`
* jika dibatalkan admin => `cancelled`

### Important behavior

* status `cancelled` tidak boleh kembali aktif otomatis
* status harus tetap bisa dihitung ulang dari waktu saat read/list sebagai fallback jika worker telat
* worker boleh update status persistently, tapi read logic tetap harus menjaga konsistensi

## 6. Reminder Rule

Reminder hanya berlaku jika:

* `is_reminder_enabled = true`
* status appointment masih `scheduled`

Reminder config per appointment:

* `is_reminder_enabled`
* `reminder_start_at`
* `reminder_interval_hours`

Reminder berhenti jika appointment menjadi:

* `on_going`
* `done`
* `cancelled`

### AI Agent must not do this

AI agent tidak boleh menambahkan asumsi berikut ke v1 tanpa diminta:

* multi channel reminder kompleks
* reminder ke client
* reminder via WhatsApp
* reminder berdasarkan hari kerja/jam kerja
* escalated reminder flow

Untuk v1, reminder cukup sederhana dan fokus ke admin.

## 7. Overlap Rule

Appointment tidak boleh overlap.

Rule validasi overlap harus berlaku saat:

* create appointment
* update appointment

Jika waktu bertabrakan dengan appointment lain yang masih aktif, request harus ditolak.

### Active overlap checking

Saat cek overlap, minimal pertimbangkan bahwa appointment dengan status berikut masih memblok jadwal:

* `scheduled`
* `on_going`

Appointment `cancelled` tidak memblok.

Appointment `done` yang berada di masa lalu secara natural tidak bentrok dengan waktu baru, tetapi logic utama tetap berbasis `start_at` dan `end_at`.

## 8. Public Checker Rule

Public checker hanya boleh:

* menerima input tanggal
* menampilkan daftar rentang waktu yang sudah terisi pada tanggal tersebut

Public checker tidak boleh:

* menyatakan slot lain pasti available
* menghitung jam kosong absolut
* membeberkan data sensitif appointment
* menampilkan nama client, alamat, atau catatan internal

Response public cukup sederhana, misalnya:

* `date`
* `items[]`

  * `start`
  * `end`
  * `status = occupied`

Jika kosong:

* kembalikan `items: []`
* frontend menampilkan teks: `Belum ada jadwal tercatat pada tanggal ini`

## 9. Entity Rule

### admins

Minimal field:

* `id`
* `name`
* `email`
* `phone`
* `password_hash`
* `created_at`
* `updated_at`

### appointments

Minimal field:

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
* `cancelled_at nullable`

### notification_logs

Minimal field:

* `id`
* `appointment_id`
* `notification_type`
* `scheduled_for`
* `sent_at`
* `recipient`
* `status`
* `message`
* `created_at`

AI agent boleh menambah field teknis minor jika benar-benar perlu implementasi, tapi tidak boleh merusak domain inti.

## 10. API Rule

Endpoint v1 yang dijaga:

* `POST /auth/login`
* `GET /appointments`
* `POST /appointments`
* `GET /appointments/:id`
* `PUT /appointments/:id`
* `PATCH /appointments/:id/cancel`
* `GET /public/schedules?date=YYYY-MM-DD`

AI agent jangan langsung menambah endpoint baru yang mengarah ke phase lanjut, misalnya:

* endpoint booking public
* endpoint approve/reject booking
* endpoint slot generator kompleks
* endpoint analytics besar

Jika ingin mengusulkan endpoint tambahan, harus dibuktikan dulu bahwa endpoint itu benar-benar diperlukan untuk v1.

## 11. Technical Direction Rule

Arah teknis project:

* fokus backend dan business logic dulu
* clean and simple architecture is preferred
* jangan overengineering
* jangan memaksa event-driven atau distributed design untuk v1
* worker/scheduler sederhana sudah cukup untuk reminder dan auto status
* implementasi harus mudah dibaca, dites, dan dimodifikasi AI agent

AI agent harus mengutamakan:

* clarity of flow
* domain consistency
* readable code
* practical implementation

Bukan mengutamakan:

* kompleksitas arsitektur
* abstraction berlapis yang belum dibutuhkan
* premature optimization

## 12. Conflict Handling Rule

Jika user meminta sesuatu yang bentrok dengan pondasi v1, AI agent harus:

1. tunjukkan letak benturannya
2. jelaskan kenapa itu keluar dari scope atau merusak domain v1
3. berikan versi alternatif yang tetap sesuai v1

Contoh benturan:

* user minta client booking langsung dari public page
* user minta slot available absolut
* user minta durasi meeting custom
* user minta multi admin assignment kompleks

Jangan langsung iya tanpa guardrail.

## 13. Output Style Rule

Semua output untuk project ini harus:

* rapi
* langsung bisa dipakai kerja
* tidak bertele-tele
* jelas boundary v1 nya
* cocok dipakai AI agent / vibe coding

Jika diminta membuat dokumen, prioritaskan format:

* jelas
* modular
* mudah dipecah jadi task implementasi
* bisa dipakai sebagai source of truth ringan

## 14. Final Principle

Untuk project ini, prinsip utama AI agent adalah:

> lebih baik sederhana tapi konsisten daripada canggih tapi keluar scope

> lebih baik realistis untuk v1 daripada terlihat lengkap tapi sulit dibangun

> lebih baik domain kuat dulu daripada fitur banyak

---