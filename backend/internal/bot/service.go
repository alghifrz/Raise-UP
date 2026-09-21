package bot

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/diuk/raiseup/internal/activity"
	"github.com/diuk/raiseup/internal/announcement"
	"github.com/diuk/raiseup/internal/complaint"
	"github.com/diuk/raiseup/internal/finance"
)

// Inbox provides WhatsApp conversation helpers used by the bot.
type Inbox interface {
	WhatsAppMessageCount(ctx context.Context, phone string) (int64, error)
	WhatsAppDisplayName(ctx context.Context, phone, fallback string) (string, error)
	WhatsAppResidentIdentity(ctx context.Context, phone string) (residentID, name, normalizedPhone string, found bool, err error)
	SendBotReply(ctx context.Context, phone, body string) error
}

// Reader marks inbound WhatsApp messages as read (optional).
type Reader interface {
	MarkAsRead(ctx context.Context, waMessageID string) error
}

// ComplaintCreator records a complaint collected by the bot.
type ComplaintCreator interface {
	Create(ctx context.Context, req complaint.CreateRequest) (*complaint.Complaint, error)
}

// Service orchestrates rule-based auto-replies for public WhatsApp chat.
type Service struct {
	inbox         Inbox
	reader        Reader
	announcements *announcement.Service
	activities    *activity.Service
	finance       *finance.Service
	complaints    ComplaintCreator
	sessions      SessionStore
	enabled       bool
	log           *slog.Logger
}

// NewService creates a bot service. When enabled is false, HandleInbound is a no-op.
func NewService(
	inbox Inbox,
	reader Reader,
	announcements *announcement.Service,
	activities *activity.Service,
	financeService *finance.Service,
	complaints ComplaintCreator,
	sessions SessionStore,
	enabled bool,
	log *slog.Logger,
) *Service {
	if log == nil {
		log = slog.Default()
	}
	return &Service{
		inbox:         inbox,
		reader:        reader,
		announcements: announcements,
		activities:    activities,
		finance:       financeService,
		complaints:    complaints,
		sessions:      sessions,
		enabled:       enabled,
		log:           log,
	}
}

// Enabled reports whether auto-replies are active.
func (s *Service) Enabled() bool {
	return s != nil && s.enabled && s.inbox != nil
}

// HandleInbound classifies an inbound text message and sends a reply when needed.
// Errors are logged but returned so the caller may choose not to fail the webhook.
func (s *Service) HandleInbound(ctx context.Context, phone, contactName, waMessageID, body string) error {
	if !s.Enabled() {
		return nil
	}

	body = strings.TrimSpace(body)
	if body == "" {
		return nil
	}

	if s.reader != nil && waMessageID != "" {
		if err := s.reader.MarkAsRead(ctx, waMessageID); err != nil {
			s.log.Warn("bot mark-as-read failed", "error", err, "wa_message_id", waMessageID)
		}
	}

	count, err := s.inbox.WhatsAppMessageCount(ctx, phone)
	if err != nil {
		return err
	}

	name, err := s.inbox.WhatsAppDisplayName(ctx, phone, contactName)
	if err != nil {
		s.log.Warn("bot display name lookup failed", "error", err)
		name = strings.TrimSpace(contactName)
	}

	if handled, flowErr := s.handlePendingComplaint(ctx, phone, name, body); handled {
		return flowErr
	}

	decision := Classify(body, Session{
		DisplayName:    name,
		IsFirstMessage: count <= 1,
	})
	if decision.MenuChoice == "1" {
		decision.Reply, err = s.latestAnnouncementsReply(ctx)
		if err != nil {
			s.log.Error("bot announcement lookup failed", "error", err)
			decision.Reply = FormatDataUnavailable("pengumuman")
		}
	}
	if decision.MenuChoice == "2" {
		decision.Reply, err = s.upcomingActivitiesReply(ctx)
		if err != nil {
			s.log.Error("bot activity lookup failed", "error", err)
			decision.Reply = FormatDataUnavailable("jadwal kegiatan")
		}
	}
	if decision.MenuChoice == "3" {
		decision.Reply, err = s.monthlyFinanceReply(ctx)
		if err != nil {
			s.log.Error("bot finance lookup failed", "error", err)
			decision.Reply = FormatDataUnavailable("laporan keuangan")
		}
	}
	if decision.MenuChoice == "4" || decision.Intent == IntentComplaint {
		if s.complaints == nil || s.sessions == nil {
			decision.Reply = FormatDataUnavailable("layanan pengaduan")
		} else if err = s.sessions.SetAwaitingComplaint(ctx, phone); err != nil {
			s.log.Error("bot complaint session start failed", "error", err)
			decision.Reply = FormatDataUnavailable("layanan pengaduan")
		} else {
			decision.Reply = FormatAskComplaint()
		}
	}
	if strings.TrimSpace(decision.Reply) == "" {
		return nil
	}

	s.log.Info("bot reply",
		"intent", string(decision.Intent),
		"menu_choice", decision.MenuChoice,
		"phone", phone,
	)

	if err := s.inbox.SendBotReply(ctx, phone, decision.Reply); err != nil {
		return err
	}
	return nil
}

func (s *Service) handlePendingComplaint(ctx context.Context, phone, displayName, body string) (bool, error) {
	if s.sessions == nil {
		return false, nil
	}
	awaiting, err := s.sessions.IsAwaitingComplaint(ctx, phone)
	if err != nil {
		s.log.Error("bot complaint session lookup failed", "error", err)
		return true, s.inbox.SendBotReply(ctx, phone, FormatDataUnavailable("layanan pengaduan"))
	}
	if !awaiting {
		return false, nil
	}

	normalized := normalizeText(body)
	if normalized == "batal" || normalized == "cancel" {
		if err := s.sessions.Clear(ctx, phone); err != nil {
			return true, err
		}
		return true, s.inbox.SendBotReply(ctx, phone, FormatComplaintCancelled())
	}
	if isMenuRequest(normalized) {
		if err := s.sessions.Clear(ctx, phone); err != nil {
			return true, err
		}
		return false, nil
	}
	if _, menuNavigation := menuChoice(normalized); menuNavigation {
		if err := s.sessions.Clear(ctx, phone); err != nil {
			return true, err
		}
		return false, nil
	}
	if s.complaints == nil {
		return true, s.inbox.SendBotReply(ctx, phone, FormatDataUnavailable("layanan pengaduan"))
	}

	residentID, residentName, normalizedPhone, found, err := s.inbox.WhatsAppResidentIdentity(ctx, phone)
	if err != nil {
		return true, err
	}
	req := complaint.CreateRequest{
		ResidentName: displayName,
		Phone:        phone,
		Category:     complaintCategory(body),
		Urgency:      "NORMAL",
		Message:      body,
	}
	if found {
		req.ResidentID = &residentID
		req.ResidentName = residentName
		req.Phone = normalizedPhone
	}
	created, err := s.complaints.Create(ctx, req)
	if err != nil {
		s.log.Error("bot complaint creation failed", "error", err, "phone", phone)
		return true, s.inbox.SendBotReply(ctx, phone, FormatComplaintFailed())
	}
	if err := s.sessions.Clear(ctx, phone); err != nil {
		return true, err
	}
	return true, s.inbox.SendBotReply(ctx, phone, FormatComplaintCreated(created.Ref))
}

func complaintCategory(message string) string {
	normalized := normalizeText(message)
	categories := []struct {
		name     string
		keywords []string
	}{
		{"KEAMANAN", []string{"keamanan", "maling", "pencurian", "berkelahi"}},
		{"KEBERSIHAN", []string{"sampah", "kotor", "got", "selokan"}},
		{"UTILITAS", []string{"listrik", "air mati", "lampu"}},
		{"INFRASTRUKTUR", []string{"jalan", "rusak", "berlubang", "bocor", "banjir"}},
	}
	for _, category := range categories {
		for _, keyword := range category.keywords {
			if strings.Contains(normalized, keyword) {
				return category.name
			}
		}
	}
	return "LAINNYA"
}

func (s *Service) latestAnnouncementsReply(ctx context.Context) (string, error) {
	if s.announcements == nil {
		return "", ErrDataSourceUnavailable
	}
	result, err := s.announcements.List(ctx, 1, 3, "", "PUBLISHED", "PUBLIC", "")
	if err != nil {
		return "", err
	}
	return FormatAnnouncements(result.Items), nil
}

func (s *Service) upcomingActivitiesReply(ctx context.Context) (string, error) {
	if s.activities == nil {
		return "", ErrDataSourceUnavailable
	}
	from := time.Now().In(activity.Jakarta).Format("2006-01-02")
	result, err := s.activities.List(ctx, 1, 5, "", from, "")
	if err != nil {
		return "", err
	}
	return FormatActivities(result.Items), nil
}

func (s *Service) monthlyFinanceReply(ctx context.Context) (string, error) {
	if s.finance == nil {
		return "", ErrDataSourceUnavailable
	}

	now := time.Now().In(finance.Jakarta)
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, finance.Jakarta)
	lastDay := firstDay.AddDate(0, 1, -1)
	from := firstDay.Format("2006-01-02")
	to := lastDay.Format("2006-01-02")

	summary, err := s.finance.Summary(ctx, from, to)
	if err != nil {
		return "", err
	}
	transactions, err := s.finance.List(ctx, 1, 100, "", "", "", from, to)
	if err != nil {
		return "", err
	}
	items := transactions.Items
	for page := 2; page <= transactions.Meta.TotalPages; page++ {
		next, err := s.finance.List(ctx, page, 100, "", "", "", from, to)
		if err != nil {
			return "", err
		}
		items = append(items, next.Items...)
	}
	return FormatFinanceReport(now, *summary, items), nil
}
