# KripDroid - Suite Kriptografi Material You untuk Android & Desktop

**KripDroid** adalah aplikasi enkripsi dan dekripsi multi-algoritma berkinerja tinggi yang ditulis dalam bahasa **Pure Go (Golang)** menggunakan framework **Gio UI (`gioui.org`)** dengan desain visual **Google Material You (Material Design 3)** serta antarmuka penuh berbahasa **Indonesia**. 

Aplikasi ini mendukung pemilihan file secara native menggunakan **File Explorer bawaan Windows** maupun **Storage Access Framework (SAF) / File Manager bawaan Android**, serta memiliki fitur **Pemulihan Ekstensi Otomatis** saat dekripsi file biner.

---

##  What's New in V8 (Pembaruan Utama)
- **Native Android Biometric Prompt**: Fitur Kunci Perangkat telah ditingkatkan dari simulasi menjadi integrasi **JNI CGO** dengan API `androidx.biometric.BiometricPrompt` yang sesungguhnya. KripDroid akan memblokir thread (terjeda) secara hardware dan hanya akan mendekripsi data jika sidik jari/wajah Anda berhasil dikenali oleh OS Android! Mencegah pembobolan oleh teman yang meminjam HP yang sedang terbuka.
- **Material You UI & Laci Navigasi (Hamburger Menu)**: Antarmuka yang sepenuhnya diremajakan. Menu Utama bersih dengan layar sambutan dan tutorial. Semua fitur kriptografi disembunyikan secara rapi di dalam laci samping (*Drawer*) bergaya Material Design dengan ikon Avatar yang elegan.
- **Perbaikan Resolusi Layar & Tata Letak**: Tombol aksi yang tumpang tindih dengan teks kini fleksibel dan rapi di semua ukuran resolusi berkat refaktorisasi `layout.Flexed(1)`. Laci menu juga sudah dimaksimalkan (*Expanded*) agar tingginya selaras dari ujung atas ke bawah layar.
- **Mode Rilis Windows GUI Tanpa CMD**: Aplikasi Windows (.exe) kini berjalan murni sebagai antarmuka pengguna grafis tanpa jendela hitam `cmd.exe` terminal di belakangnya (`-ldflags="-s -w -H windowsgui"`).

---

## 🚀 Fitur Unggulan

- **Pemulihan Ekstensi File Otomatis (*Auto-Extension Recovery*)**:
  - Saat file dienkripsi (misal: `video.mp4`, `laporan.pdf`, `foto.jpg`), KripDroid otomatis merekam ekstensi asli ke dalam header terenkripsi (`KRIPDROID\x02`).
  - Saat file didekripsi (walaupun nama file enkripsinya diubah menjadi `rahasia.krip`), KripDroid **secara otomatis mengenali dan mengembalikan format/ekstensi aslinya** (`.mp4`, `.pdf`, dll.) tanpa perlu diingat atau diketik manual oleh pengguna.
- **Antarmuka Bahasa Indonesia & Material You (Material Design 3)**:
  - Seluruh teks antarmuka, label, tombol, tooltip, dan pesan status disajikan dalam Bahasa Indonesia yang bersih dan profesional.
  - Palet warna tonal dinamis (*Emerald/Teal Dynamic Dark Theme*), kontainer rounded, elevated cards, indikator badge dinamis, dan navigasi tab responsif.
- **Dukungan Native File Explorer**:
  - Tombol **📁 Pilih File...** langsung memicu dialog File Explorer bawaan Windows di Desktop dan Storage Access Framework (SAF) di Android.
  - Tombol **💾 Simpan Hasil File** memungkinkan pengguna memilih lokasi penyimpanan dan nama file hasil secara leluasa dengan ekstensi asli yang otomatis terpasang.
  - Opsi pengetikan/penempelan jalur file manual tetap tersedia bagi pengguna tingkat lanjut.
- **100% Pure Go Implementation**: Dibangun tanpa dependensi kode Java manual di sisi aplikasi; seluruh logika kriptografi dan rendering GUI diproses langsung di layer Go & Gio CGO NDK.
- **Dukungan 13 Spektrum Algoritma Kriptografi**: Mulai dari cipher klasik untuk kebutuhan edukasi hingga cipher terotentikasi modern (AES-256-GCM, ChaCha20-Poly1305) berstandar militer.
- **Mode Teks Interaktif**:
  - Input field responsif dengan *live character & byte counter*.
  - Input Kunci Rahasia / Kata Sandi dengan fitur *toggle mask* (Tampilkan / Sembunyikan).
  - Mode IV / Nonce kustom atau otomatis acak aman (*crypto/rand*).
  - Tombol aksi cepat: *Enkripsi Teks*, *Dekripsi Teks*, *Salin ke Papan Klip*, *Tukar Masukan/Keluaran*, dan *Bersihkan*.
- **Mode File Asinkron**:
  - Pemrosesan file biner tanpa batasan format (.mp4, .pdf, .docx, .zip, .png, .mkv, .apk, dll.).
  - Indikator progres bar dan status proses secara *real-time* di background thread (*zero UI freeze*).
  - Generator file sampel terintegrasi untuk pengujian instan.
- **Panduan Keamanan Cipher**: Modul edukasi di dalam aplikasi yang merinci kekuatan komputasi, kelemahan kriptanalisis, dan rekomendasi penggunaan setiap algoritma.

---

## 🛡️ Tabel Hierarki Algoritma Kriptografi (Terlemah ke Terkuat)

| No | Kategori | Algoritma | Ukuran Kunci | Mode / Primitif | Tingkat Keamanan | Rekomendasi & Use-Case |
|:---|:---|:---|:---|:---|:---|:---|
| 1 | **Klasik** | **Caesar Cipher** | Integer (0–25) / String | Monoalphabetic Substitution | 🔴 Lemah / Edukasi | Pembelajaran dasar kriptografi dan pergeseran karakter. |
| 2 | **Klasik** | **ROT13** | Fixed (13) | Monoalphabetic Substitution | 🔴 Lemah / Edukasi | Obfuscation teks sederhana, spoiler tag forum online. |
| 3 | **Klasik** | **Atbash Cipher** | None (Alphabet Inversion) | Monoalphabetic Substitution | 🔴 Lemah / Edukasi | Studi kriptografi kuno (pembalikan alfabet A↔Z, B↔Y). |
| 4 | **Klasik** | **Vigenère Cipher** | Variabel (String/Kata Kunci) | Polyalphabetic Shift Table | 🟠 Lemah / Edukasi | Demonstrasi serangan analisis frekuensi Kasiski. |
| 5 | **Klasik** | **Simple XOR Cipher** | Variabel (Byte Key) | Bitwise XOR Stream | 🟠 Lemah / Edukasi | Pengaburan biner cepat, edukasi operasi bitwise. |
| 6 | **Legacy** | **DES (Data Encryption Standard)** | 56-bit (8 Byte) | 64-bit Block, CBC + PKCS7 | 🟡 Menengah / Legacy | Analisis cipher blok historis (rentan brute-force modern). |
| 7 | **Legacy** | **Triple DES (3DES / EDE)** | 168-bit (24 Byte) | 64-bit Block, CBC + PKCS7 | 🟡 Menengah / Legacy | Kompatibilitas sistem perbankan lama & smart card EMV. |
| 8 | **Legacy** | **Blowfish** | 32–448 bit (1–56 Byte) | 64-bit Block, CBC + PKCS7 | 🟡 Menengah / Legacy | Proteksi data lokal non-kritis dengan kecepatan tinggi. |
| 9 | **Modern** | **AES-128-CBC** | 128-bit (16 Byte) | 128-bit Block, CBC + PKCS7 | 🟢 Standar Keamanan Tinggi | Enkripsi data umum berstandar NIST dengan mode CBC. |
| 10 | **Modern** | **AES-128-GCM** | 128-bit (16 Byte) | AEAD (Galois/Counter Mode) | 🟢 Standar Keamanan Tinggi | Kerahasiaan + Integritas terotentikasi, throughput tinggi. |
| 11 | **Modern** | **AES-256-CBC** | 256-bit (32 Byte) | 128-bit Block, CBC + PKCS7 | 🛡️ Standar Militer & Rahasia | Perlindungan data jangka panjang tahan serangan brute-force. |
| 12 | **Modern** | **AES-256-GCM** | 256-bit (32 Byte) | AEAD (Galois/Counter Mode) | 🛡️ Standar Militer & Rahasia | Enkripsi dokumen sensitif, transaksi finansial, VPN/TLS. |
| 13 | **Modern** | **ChaCha20-Poly1305** | 256-bit (32 Byte) | Stream Cipher + Poly1305 AEAD | 🛡️ Standar Militer & TLS 1.3 | Kecepatan optimal pada CPU mobile tanpa akselerasi hardware AES. |

---

## 📐 Struktur Protokol Header File KripDroid (`KRIPDROID\x02`)

```
+--------------------------+-----------------------+--------------------+---------------------+------------------+---------------------+-------------------+
| Magic Header (10 Bytes)  | Algo Length (1 Byte)  | Algo ID (N Bytes)  | Ext Length (1 Byte) | Ext Str (M Bytes)| IV Length (1 Byte)  | IV Data (K Bytes) |
| "KRIPDROID\x02"          | misal: 0x09           | "aes256gcm"        | misal: 0x04         | ".mp4"           | misal: 0x0C         | [12 bytes Nonce]  |
+--------------------------+-----------------------+--------------------+---------------------+------------------+---------------------+-------------------+
| Payload Size (8 Bytes)   | Encrypted Data Body Payload ...                                                                                                |
| Big-Endian uint64        | [Ciphertext biner terenkripsi]                                                                                                 |
+--------------------------+--------------------------------------------------------------------------------------------------------------------------------+
```

---

## 🛠️ Prasyarat Build & Lingkungan Pengembangan

Sebelum melakukan kompilasi, pastikan lingkungan pengembangan memiliki perkakas berikut:

1. **Go Toolchain**: Go versi `1.22+` (disarankan Go `1.26+`).
2. **Gio Build Tool (`gogio`)**:
   ```bash
   go install gioui.org/cmd/gogio@latest
   ```
3. **Android SDK**: Android API level 24 ke atas (misal Android SDK 33 / 34).
4. **Android NDK**: NDK versi `r25+` atau `r26+` (misal `26.3.11579264`).
5. **Java Development Kit (JDK)**: OpenJDK 17 atau 21 LTS.

### Konfigurasi Environment Variables (PowerShell):
```powershell
$env:ANDROID_SDK_ROOT = "C:\Android\Sdk"
$env:ANDROID_HOME = "C:\Android\Sdk"
$env:ANDROID_NDK_HOME = "C:\Android\Sdk\ndk\26.3.11579264"
$env:JAVA_HOME = "C:\Program Files\Microsoft\jdk-17.0.20.8-hotspot"
$env:PATH = "$env:JAVA_HOME\bin;$env:PATH"
```

---

## 🚀 Panduan Kompilasi & Pengujian

### 1. Menjalankan Unit Test Kriptografi:
```powershell
go test -v ./...
```

### 2. Kompilasi Executable Desktop (Windows / Linux / macOS):
```powershell
go build -o KripDroid.exe .
```

### 3. Kompilasi Android APK (.apk):
Jalankan perintah `gogio` untuk membuat file APK siap instal:
```powershell
gogio -target android -arch arm64,arm -appid com.kripdroid.app -minsdk 24 -o KripDroid.apk .
```

---

## 📲 Panduan Instalasi ke Perangkat / Emulator Android

### Menggunakan ADB (Android Debug Bridge):
1. Hubungkan perangkat Android ke PC dengan mode **USB Debugging** aktif, atau jalankan emulator Android.
2. Verifikasi deteksi perangkat:
   ```bash
   adb devices
   ```
3. Instal file APK hasil kompilasi:
   ```bash
   adb install -r KripDroid.apk
   ```
4. Buka aplikasi **KripDroid** dari menu aplikasi Android Anda.

(Btw sebenarnya release binary nya ada exe, ga cuman apk, cuman saya males, hehe)
