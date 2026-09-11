# Tokopedia Backend - Golang & MongoDB

Backend REST API untuk Tokopedia Clone berbasis Golang (Gin Framework), MongoDB, dan JWT Authentication.

## 🚀 Fitur Backend
- **Autentikasi & Otorisasi**: Register, Login, Bcrypt Password Hashing, JWT Bearer Tokens, Role Guard (User & Admin).
- **Katalog Produk**: CRUD Produk, Filter dinamis (kategori, search query, rentang harga, lokasi), Seeding data otomatis.
- **Voucher Diskon**: Validasi kode promo, minimum belanja, dan pemotongan harga otomatis.
- **Transaksi & Pesanan**: Checkout pesanan, nomor invoice unik, riwayat belanja pembeli, panel update status pesanan untuk Admin.
- **Ulasan Produk**: Review bintang 1-5 dan update rating produk otomatis.

---

## 🛠️ Panduan Menjalankan di Server VPS (Docker Compose)

### 1. Prasyarat di VPS:
- VPS Ubuntu / Debian (RAM minimal 1GB)
- Sudah terinstall Docker & Docker Compose:
  ```bash
  sudo apt update && sudo apt install -y docker.io docker-compose
  ```

### 2. Clone Repository di VPS:
```bash
git clone <URL_REPO_BACKEND_ANDA>.git tokped-backend
cd tokped-backend
```

### 3. Buat File Environment:
```bash
cp .env.example .env
```
*(Sesuaikan JWT_SECRET atau FRONTEND_URL dengan domain Vercel Anda jika diperlukan)*

### 4. Jalankan dengan 1 Perintah:
```bash
docker-compose up -d --build
```
Backend API otomatis aktif di port `8080` dan MongoDB aktif di port `27017`!

---

## 🔒 Konfigurasi Nginx Reverse Proxy & SSL (Domain di VPS)

Jika Anda memiliki domain (contoh: `api.domainanda.com`), pasang Nginx di VPS:

```nginx
server {
    server_name api.domainanda.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Pasang SSL gratis Let's Encrypt:
```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d api.domainanda.com
```

Setelah SSL aktif, masukkan URL `https://api.domainanda.com/api` ke Environment Variable Vercel pada frontend (`VITE_API_URL`).
