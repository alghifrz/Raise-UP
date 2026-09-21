package bot

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/diuk/raiseup/internal/activity"
	"github.com/diuk/raiseup/internal/announcement"
	"github.com/diuk/raiseup/internal/finance"
)

// FormatWelcome builds the main menu greeting (mirrors n8n King Nasir menu).
func FormatWelcome(name string) string {
	return fmt.Sprintf(
		"Halo %s, ada yang bisa dibantu?\n"+
			"Silakan pilih menu dibawah ini:\n\n"+
			"*1.* Pengumuman terbaru\n"+
			"*2.* Jadwal kegiatan RT\n"+
			"*3.* Laporan Keuangan\n"+
			"*4.* Kirim Pengaduan\n\n"+
			"Ketik angka *1–4* untuk memilih.\n"+
			"Contoh : 1",
		name,
	)
}

// FormatMenuHint reminds citizens of the menu after an unclear message.
func FormatMenuHint() string {
	return "Maaf, saya belum memahami pesan itu.\n" +
		"Ketik *menu* atau angka *1–4* untuk layanan RT."
}

// FormatMenuAck is a short confirmation before data replies (phase 2).
func FormatMenuAck(choice string) string {
	labels := map[string]string{
		"1": "pengumuman terbaru",
		"2": "jadwal kegiatan",
		"3": "laporan keuangan",
		"4": "form pengaduan",
	}
	label := labels[choice]
	if label == "" {
		label = "informasinya"
	}
	return fmt.Sprintf(
		"Baik, saya siapkan *%s*.\n"+
			"_Balasan detail menyusul di update berikutnya._",
		label,
	)
}

// FormatAnnouncements builds a WhatsApp-friendly list of recent public announcements.
func FormatAnnouncements(items []announcement.AnnouncementSummary) string {
	if len(items) == 0 {
		return "📢 Belum ada pengumuman terbaru.\n\nKetik *menu* untuk kembali."
	}

	var b strings.Builder
	b.WriteString("📢 *Pengumuman Terbaru*\n")
	for i, item := range items {
		b.WriteString("\n")
		if len(items) > 1 {
			fmt.Fprintf(&b, "*%d. %s*\n", i+1, strings.TrimSpace(item.Title))
		} else {
			fmt.Fprintf(&b, "*%s*\n", strings.TrimSpace(item.Title))
		}

		content := strings.TrimSpace(item.Body)
		if content == "" {
			content = strings.TrimSpace(item.Excerpt)
		}
		if content != "" {
			b.WriteString(content)
			b.WriteString("\n")
		}
		if date := formatAnnouncementDate(item.PublishedAt); date != "" {
			fmt.Fprintf(&b, "📅 %s\n", date)
		}
	}
	b.WriteString("\nKetik *menu* untuk kembali.")
	return strings.TrimSpace(b.String())
}

// FormatActivities builds a WhatsApp-friendly list of upcoming activities.
func FormatActivities(items []activity.Activity) string {
	if len(items) == 0 {
		return "📅 Belum ada kegiatan yang akan datang.\n\nKetik *menu* untuk kembali."
	}

	var b strings.Builder
	b.WriteString("📅 *Jadwal Kegiatan RT*\n")
	for i, item := range items {
		fmt.Fprintf(&b, "\n*%d. %s*\n", i+1, strings.TrimSpace(item.Name))
		if date := formatActivityDate(item.Date); date != "" {
			fmt.Fprintf(&b, "🗓️ %s\n", date)
		}
		if description := strings.TrimSpace(item.Description); description != "" {
			b.WriteString(description)
			b.WriteString("\n")
		}
	}
	b.WriteString("\nKetik *menu* untuk kembali.")
	return strings.TrimSpace(b.String())
}

// FormatFinanceReport builds the current monthly cash report.
func FormatFinanceReport(period time.Time, summary finance.Summary, items []finance.Transaction) string {
	var b strings.Builder
	fmt.Fprintf(&b, "💰 *Laporan Keuangan RT*\n*Periode %s*\n\n", formatIndonesianMonth(period))
	fmt.Fprintf(&b, "*Saldo*\n%s\n\n", formatIDR(summary.Balance))
	fmt.Fprintf(&b, "📈 Pemasukan: %s\n", formatIDR(summary.TotalIncome))
	fmt.Fprintf(&b, "📉 Pengeluaran: %s\n", formatIDR(summary.TotalExpense))

	if len(items) > 0 {
		incomeByCategory := make(map[string]int64)
		expenseByCategory := make(map[string]int64)
		for _, item := range items {
			category := strings.ToUpper(strings.TrimSpace(item.Category))
			if category == "" {
				category = "LAINNYA"
			}
			if string(item.Type) == "EXPENSE" {
				expenseByCategory[category] += item.Amount
			} else {
				incomeByCategory[category] += item.Amount
			}
		}

		if len(incomeByCategory) > 0 {
			b.WriteString("\n*Rincian pemasukan:*\n")
			for _, category := range sortedCategories(incomeByCategory) {
				fmt.Fprintf(&b, "+ %s: %s\n", category, formatIDR(incomeByCategory[category]))
			}
		}
		if len(expenseByCategory) > 0 {
			b.WriteString("\n*Rincian pengeluaran:*\n")
			for _, category := range sortedCategories(expenseByCategory) {
				fmt.Fprintf(&b, "- %s: %s\n", category, formatIDR(expenseByCategory[category]))
			}
		}
	} else {
		b.WriteString("\nBelum ada transaksi pada periode ini.\n")
	}

	b.WriteString("\nKetik *menu* untuk kembali.")
	return strings.TrimSpace(b.String())
}

func sortedCategories(values map[string]int64) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// FormatDataUnavailable is a safe reply when a menu data source fails.
func FormatDataUnavailable(subject string) string {
	return fmt.Sprintf(
		"Maaf, data %s sedang tidak dapat dimuat. Silakan coba lagi nanti atau ketik *menu*.",
		subject,
	)
}

func formatAnnouncementDate(value *string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return strings.TrimSpace(*value)
	}
	return parsed.In(time.FixedZone("WIB", 7*60*60)).Format("02-01-2006")
}

func formatActivityDate(value string) string {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return strings.TrimSpace(value)
	}
	return parsed.In(activity.Jakarta).Format("02-01-2006")
}

func formatIndonesianMonth(value time.Time) string {
	months := [...]string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	return fmt.Sprintf("%s %d", months[value.Month()-1], value.Year())
}

func formatIDR(amount int64) string {
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	raw := strconv.FormatInt(amount, 10)
	var b strings.Builder
	for i, r := range raw {
		if i > 0 && (len(raw)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}
	return sign + "Rp" + b.String()
}

// FormatAskComplaint asks the citizen to describe their complaint.
func FormatAskComplaint() string {
	return "Baik, silakan jelaskan pengaduan Anda dalam *satu pesan* " +
		"(lokasi + masalah + kapan terjadi).\n" +
		"Nanti saya bantu catat ke sistem.\n\nKetik *batal* untuk membatalkan."
}

// FormatComplaintCreated confirms that a complaint was stored.
func FormatComplaintCreated(ref string) string {
	return fmt.Sprintf(
		"✅ Pengaduan Anda berhasil diterima.\n\nNomor laporan: *%s*\nStatus: *BARU*\n\nPengurus akan menindaklanjuti laporan Anda.",
		strings.TrimSpace(ref),
	)
}

// FormatComplaintFailed keeps the flow open so the citizen can retry.
func FormatComplaintFailed() string {
	return "Maaf, pengaduan belum dapat disimpan karena terjadi gangguan. Silakan kirim ulang detail pengaduan Anda beberapa saat lagi atau ketik *batal*."
}

// FormatComplaintCancelled confirms cancellation.
func FormatComplaintCancelled() string {
	return "Pengaduan dibatalkan.\n\nKetik *menu* untuk melihat pilihan layanan."
}
