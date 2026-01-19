Ini adalah Functional Specification Document (FSD) lengkap untuk membangun SQDesk versi TUI menggunakan Golang. Dokumen ini dirancang sebagai panduan langkah-demi-langkah bagi pengembang.
Functional Specification Document (FSD): SQDesk TUI
1. Ringkasan Produk
SQDesk adalah database client berbasis terminal (TUI) yang ringan namun cerdas. Berfokus pada kecepatan eksekusi kueri, kenyamanan mengetik dengan bantuan AI, dan kemudahan navigasi data tanpa harus menulis kueri manual.
2. Arsitektur Teknis (Stack)
Language: Golang 1.21+
TUI Framework: Bubble Tea (State management)
Styling & Layout: Lip Gloss
Input Handling: Bubbles (Textarea, TextInput, List, Table)
Database Drivers: sqlx (mendukung PostgreSQL, MySQL, SQLite)
AI Integration: SDK resmi dari Google (Gemini) atau Anthropic (Claude) via HTTP client.
3. Flow: Penggunaan Pertama (First-Run Experience)
Saat user menjalankan sqdesk pertama kali dan file konfigurasi belum ada:
Welcome Screen: Animasi ASCII art SQDesk.
Set-up Style: User memilih tema (misal: Dracula, Nord, atau System Default).
LLM Configuration:
Pilih Provider: Gemini, Claude, OpenAI, atau Skip.
Input API Token (disimpan secara lokal dan terenkripsi/aman di ~/.config/sqdesk/config.yaml).
Connection Setup: Input awal untuk satu database (Host, Port, User, Pass, DB Name).
4. Fitur Utama & Spesifikasi Fungsional
4.1. Dashboard Utama (Workspace)
Terdiri dari 3 panel utama (menggunakan layout Bubble Tea):
Sidebar (Left): Daftar Tabel (Tree view/List). User bisa klik/enter untuk langsung "Preview Data".
Editor Pane (Top Right): Area menulis kueri SQL dengan syntax highlighting sederhana.
Result/Preview Pane (Bottom Right): Tabel interaktif hasil kueri atau preview tabel.
4.2. Smart Editor (Auto-Suggestion & Completion)
Schema Indexing: Saat koneksi aktif, SQDesk melakukan SELECT table_name, column_name ke metadata DB dan menyimpannya di memori (Map/Struct).
Real-time Suggestion: Saat user mengetik SELECT * FROM u, muncul teks ghost (abu-abu) di depan kursor: users. Tekan Tab untuk melengkapi.
Context Aware: Jika sudah mengetik SELECT email FROM users WHERE , suggestion hanya akan memunculkan kolom yang ada di tabel users.
4.3. AI Integration (The "Desk Pad" Intelligence)
Direct Prompt (NL2SQL): Tekan Ctrl+G -> Muncul prompt input -> Ketik "tampilkan 10 transaksi terakhir user budi" -> AI mengembalikan kode SQL langsung ke editor.
Selection & Refactor: * User memblok baris kode di editor.
Tekan Ctrl+K.
Input prompt: "tambahkan limit dan urutkan berdasarkan tgl".
AI memperbarui teks yang diblok tersebut.
4.4. GUI-Style Data Browsing
User navigasi ke Sidebar -> Pilih tabel orders.
Tekan Enter.
SQDesk secara otomatis menjalankan SELECT * FROM orders LIMIT 100 dan menampilkannya di Result Pane tanpa user mengetik apapun.
5. Flow Pengaturan (Settings)
User dapat menekan Ctrl+, untuk membuka panel modal setting:
Style: Mengubah skema warna terminal.
LLM: Mengganti model (misal dari Gemini ke Claude) atau update API Key.
Database Driver: Menambah koneksi baru (multi-connection support).
Connections Management: Edit/Hapus kredensial database yang tersimpan.
6. Desain Antarmuka (TUI Layout Plan)

Plaintext


+-------------------------------------------------------------+
| SQDesk CLI | [DB: Production_v1] | Model: Gemini 1.5 Pro    |
+----------+--------------------------------------------------+
| TABLES   | 1  SELECT * |
| > users  | 2  FROM transactions                             |
|   orders | 3  WHERE status = 'pending'                      |
|   items  | 4  _                                             |
|          +--------------------------------------------------+
|          | AI Prompt: [ Tambahkan filter tanggal hari ini ] |
+----------+--------------------------------------------------+
| RESULTS (3 rows found)                                      |
+----+------------+----------+-----------+--------------------+
| ID | USER_ID    | AMOUNT   | STATUS    | DATE               |
+----+------------+----------+-----------+--------------------+
| 1  | 99         | 50000    | pending   | 2024-01-18         |
+----+------------+----------+-----------+--------------------+
| [F1] Help [Ctrl+S] Run [Ctrl+G] AI [Ctrl+N] New Conn        |
+-------------------------------------------------------------+


7. Roadmap Pengembangan (WBS)
Minggu 1: Core TUI & DB Connection
Inisialisasi project Go & Bubble Tea.
Implementasi sistem konfigurasi (YAML/Viper).
Membuat konektor database generic (Postgres/SQLite).
Membuat UI Sidebar (List tabel).
Minggu 2: Editor & Preview
Integrasi bubbles/textarea untuk SQL Editor.
Implementasi bubbles/table untuk Result Pane.
Optimasi fetching data besar (pagination/limit otomatis).
Minggu 3: Intelligence Layer
Membuat "Schema Crawler" untuk menyimpan metadata kolom/tabel.
Implementasi logika Auto-completion (Trie algorithm atau fuzzy match).
Integrasi API Gemini/Claude untuk fitur Ctrl+G.
Minggu 4: Polishing & CLI Distribution
Sistem per-tema (Styling dengan Lip Gloss).
Fitur First-run setup.
Build biner untuk berbagai OS (Linux, Mac, Windows).
8. Key Shortcuts (Shortcut Keyboard)
Ctrl + Enter : Jalankan kueri.
Ctrl + G : Tanya AI (Generate SQL baru).
Ctrl + K : Refactor SQL yang diblok dengan AI.
Tab : Terima suggestion/Pindah antar panel.
Ctrl + P : Cari tabel dengan cepat (Fuzzy search).
Esc : Tutup modal atau batalkan input.
Langkah Selanjutnya:
Apakah Anda ingin saya buatkan struktur folder boilerplate (file main.go, internal/db, internal/tui) untuk memulai project ini berdasarkan FSD di atas?
