# Absensi CN API

Backend REST API untuk `Absensi CN`, sistem absensi sekolah `SMK Citra Negara`.

Project ini dibangun menggunakan `Golang` dengan arsitektur modular yang tetap sederhana, realistis, dan mudah dikembangkan untuk kebutuhan tugas sekolah maupun portfolio.

## Kepemilikan dan Kredit

Website absensi ini dibuat oleh `Randhu Paksi Membumi`.

Copyright 2026 `Randhu Paksi Membumi`. All rights reserved.

Project ini dibuat untuk kebutuhan ujian praktik SMK dan portfolio. Penggunaan untuk operasional sekolah, distribusi ulang, rebranding, penghapusan atribusi pembuat, atau pengembangan turunan membutuhkan izin tertulis dari `Randhu Paksi Membumi`.

## Tentang Project

`Absensi CN API` menangani proses inti sistem:

- autentikasi
- validasi role
- manajemen data user dan siswa
- proses absensi
- pengajuan izin atau sakit
- monitoring data kehadiran

Role utama yang dipakai:

- `STUDENT`
- `HOMEROOM_TEACHER`
- `BK`
- `ADMIN`

## Konsep Login

Backend memakai konsep `portal-aware login` agar sinkron dengan frontend.

Portal yang didukung:

- `student`
- `staff`

Aturannya:

- akun `STUDENT` hanya boleh login lewat portal `student`
- akun `HOMEROOM_TEACHER`, `BK`, dan `ADMIN` hanya boleh login lewat portal `staff`

Identifier login:

- siswa login dengan `NIS + password`
- staff login dengan `username + password`

Validasi `NIS`:

- wajib diisi
- angka only
- tepat 10 digit
- disimpan sebagai `VARCHAR(10)`

## Tujuan Backend

- Menyediakan REST API yang clean dan mudah diintegrasikan
- Menjaga validasi role dan permission tetap jelas
- Menangani flow absensi harian secara konsisten
- Menjadi fondasi data untuk dashboard siswa, walas, BK, dan admin

## Stack

- `Go`
- `Gin`
- `GORM`
- `MySQL`
- `JWT`
- `bcrypt`
- `validator`
- `CORS`
- `godotenv`
- `golang-migrate`

## Status Saat Ini

Yang sudah tersedia:

- fondasi backend project
- struktur folder modular
- koneksi MySQL
- setup config environment
- auth endpoint awal
- portal-aware login mechanism
- middleware dan utilitas dasar

Yang akan dilanjutkan:

- tabel dan migration inti
- auth berbasis data database nyata
- endpoint siswa, kelas, absensi, izin/sakit, dan monitoring BK
- middleware otorisasi per role

## Struktur Folder

```text
cmd/
  api/
internal/
  app/
  config/
  database/
  middleware/
  modules/
  router/
pkg/
  password/
  response/
  token/
```

## Menjalankan Project

1. Siapkan file environment dari `.env.example`

2. Jalankan API

```bash
go run ./cmd/api
```

3. Default local server

```bash
http://localhost:8080
```

## Script Berguna

```bash
go run ./cmd/api
go test ./...
go build ./...
go mod tidy
```

## Environment

Contoh konfigurasi lokal:

```env
APP_NAME=Absensi CN API
APP_ENV=development
APP_PORT=8080
API_PREFIX=/api/v1
APP_ALLOWED_ORIGINS=http://localhost:3000

DB_ENABLED=true
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=absensi_cn
DB_USER=root
DB_PASSWORD=

JWT_SECRET=replace-with-secure-secret
JWT_EXPIRES_IN_HOURS=24
```

## Endpoint Awal

Beberapa endpoint yang sudah disiapkan:

- `GET /api/v1/health`
- `POST /api/v1/auth/login`
- `GET /api/v1/attendance`
- `POST /api/v1/attendance/check-in`
- `POST /api/v1/attendance/check-out`

## Integrasi Frontend

API ini dirancang untuk dipakai oleh repository frontend:

- `absensi-cn-app`

Pastikan frontend mengarah ke base URL API yang sesuai pada environment lokal.

## Catatan

Project ini sengaja dijaga:

- cukup profesional untuk portfolio
- cukup sederhana untuk tugas sekolah
- tidak berlebihan dalam arsitektur
- fokus pada alur absensi yang jelas dan role yang tegas
