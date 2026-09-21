package bot_test

import (
	"strings"
	"testing"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/internal/activity"
	"github.com/diuk/raiseup/internal/announcement"
	"github.com/diuk/raiseup/internal/bot"
	"github.com/diuk/raiseup/internal/finance"
)

func TestClassifyFirstMessageWelcome(t *testing.T) {
	d := bot.Classify("halo", bot.Session{DisplayName: "Budi", IsFirstMessage: true})
	if d.Intent != bot.IntentWelcome || !d.ShouldWelcome {
		t.Fatalf("expected welcome, got %+v", d)
	}
	if !strings.Contains(d.Reply, "Budi") || !strings.Contains(d.Reply, "*1.*") {
		t.Fatalf("welcome reply missing name/menu: %q", d.Reply)
	}
}

func TestClassifyMenuDigits(t *testing.T) {
	d := bot.Classify("2", bot.Session{IsFirstMessage: true})
	if d.Intent != bot.IntentMenu || d.MenuChoice != "2" || d.ShouldWelcome {
		t.Fatalf("expected menu 2, got %+v", d)
	}
	if !strings.Contains(d.Reply, "jadwal kegiatan") {
		t.Fatalf("unexpected ack: %q", d.Reply)
	}
}

func TestClassifyMenuRequest(t *testing.T) {
	d := bot.Classify("menu", bot.Session{DisplayName: "Siti", IsFirstMessage: false})
	if d.Intent != bot.IntentWelcome || !d.ShouldWelcome {
		t.Fatalf("expected welcome menu, got %+v", d)
	}
}

func TestClassifyComplaintKeyword(t *testing.T) {
	d := bot.Classify("Mau lapor sampah menumpuk di gang A", bot.Session{IsFirstMessage: false})
	if d.Intent != bot.IntentComplaint {
		t.Fatalf("expected complaint, got %+v", d)
	}
	if !strings.Contains(strings.ToLower(d.Reply), "pengaduan") {
		t.Fatalf("unexpected reply: %q", d.Reply)
	}
}

func TestClassifyOtherHint(t *testing.T) {
	d := bot.Classify("asdfgh", bot.Session{IsFirstMessage: false})
	if d.Intent != bot.IntentOther {
		t.Fatalf("expected other, got %+v", d)
	}
	if !strings.Contains(d.Reply, "1–4") && !strings.Contains(d.Reply, "1-4") {
		t.Fatalf("hint should mention menu digits: %q", d.Reply)
	}
}

func TestFormatAnnouncements(t *testing.T) {
	publishedAt := "2026-09-21T02:00:00Z"
	reply := bot.FormatAnnouncements([]announcement.AnnouncementSummary{
		{
			Title:       "Kerja Bakti",
			Body:        "Kerja bakti dimulai pukul 07.00 WIB.",
			Status:      db.AnnouncementStatusPUBLISHED,
			Visibility:  db.AnnouncementVisibilityPUBLIC,
			PublishedAt: &publishedAt,
		},
	})

	for _, want := range []string{"Pengumuman Terbaru", "Kerja Bakti", "07.00 WIB", "21-09-2026"} {
		if !strings.Contains(reply, want) {
			t.Fatalf("reply %q does not contain %q", reply, want)
		}
	}
}

func TestFormatAnnouncementsEmpty(t *testing.T) {
	reply := bot.FormatAnnouncements(nil)
	if !strings.Contains(reply, "Belum ada pengumuman") {
		t.Fatalf("unexpected empty reply: %q", reply)
	}
}

func TestFormatActivities(t *testing.T) {
	reply := bot.FormatActivities([]activity.Activity{
		{
			Name:        "Kerja Bakti",
			Description: "Berkumpul di balai RT.",
			Date:        "2026-09-27T00:00:00Z",
		},
	})
	for _, want := range []string{"Jadwal Kegiatan RT", "Kerja Bakti", "27-09-2026", "balai RT"} {
		if !strings.Contains(reply, want) {
			t.Fatalf("reply %q does not contain %q", reply, want)
		}
	}
}

func TestFormatFinanceReport(t *testing.T) {
	reply := bot.FormatFinanceReport(
		time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
		finance.Summary{
			TotalIncome:  2500000,
			TotalExpense: 750000,
			Balance:      1750000,
		},
		[]finance.Transaction{
			{Type: db.CashTransactionTypeINCOME, Title: "Iuran Budi", Category: "IURAN", Amount: 1500000},
			{Type: db.CashTransactionTypeINCOME, Title: "Iuran Siti", Category: "IURAN", Amount: 1000000},
			{Type: db.CashTransactionTypeEXPENSE, Title: "Angkut sampah", Category: "KEBERSIHAN", Amount: 750000},
		},
	)
	for _, want := range []string{
		"September 2026",
		"Rp1.750.000",
		"Pemasukan: Rp2.500.000",
		"Pengeluaran: Rp750.000",
		"+ IURAN: Rp2.500.000",
		"- KEBERSIHAN: Rp750.000",
	} {
		if !strings.Contains(reply, want) {
			t.Fatalf("reply %q does not contain %q", reply, want)
		}
	}
	if strings.Contains(reply, "Budi") || strings.Contains(reply, "Siti") {
		t.Fatalf("report must not expose per-resident transaction names: %q", reply)
	}
}
