package bot

import (
	"context"
	"strings"
	"testing"

	"github.com/diuk/raiseup/internal/complaint"
)

type botInboxFake struct {
	replies []string
}

func (f *botInboxFake) WhatsAppMessageCount(context.Context, string) (int64, error) {
	return 2, nil
}

func (f *botInboxFake) WhatsAppDisplayName(context.Context, string, string) (string, error) {
	return "Budi", nil
}

func (f *botInboxFake) WhatsAppResidentIdentity(context.Context, string) (string, string, string, bool, error) {
	return "11111111-1111-1111-1111-111111111111", "Budi", "08123456789", true, nil
}

func (f *botInboxFake) SendBotReply(_ context.Context, _ string, body string) error {
	f.replies = append(f.replies, body)
	return nil
}

type botSessionFake struct {
	awaiting bool
}

func (f *botSessionFake) IsAwaitingComplaint(context.Context, string) (bool, error) {
	return f.awaiting, nil
}

func (f *botSessionFake) SetAwaitingComplaint(context.Context, string) error {
	f.awaiting = true
	return nil
}

func (f *botSessionFake) Clear(context.Context, string) error {
	f.awaiting = false
	return nil
}

type complaintCreatorFake struct {
	requests []complaint.CreateRequest
}

func (f *complaintCreatorFake) Create(_ context.Context, req complaint.CreateRequest) (*complaint.Complaint, error) {
	f.requests = append(f.requests, req)
	return &complaint.Complaint{Ref: "PGD-20260921-0001"}, nil
}

func TestComplaintMenuCollectsAndCreatesComplaint(t *testing.T) {
	inbox := &botInboxFake{}
	sessions := &botSessionFake{}
	complaints := &complaintCreatorFake{}
	service := NewService(inbox, nil, nil, nil, nil, complaints, sessions, true, nil)

	if err := service.HandleInbound(context.Background(), "628123456789", "Budi WA", "wamid-1", "4"); err != nil {
		t.Fatalf("choose complaint menu: %v", err)
	}
	if !sessions.awaiting {
		t.Fatal("expected complaint session to be active")
	}
	if len(inbox.replies) != 1 || !strings.Contains(inbox.replies[0], "jelaskan pengaduan") {
		t.Fatalf("unexpected prompt: %#v", inbox.replies)
	}

	detail := "Lampu jalan rusak di blok A sejak tadi malam"
	if err := service.HandleInbound(context.Background(), "628123456789", "Budi WA", "wamid-2", detail); err != nil {
		t.Fatalf("submit complaint: %v", err)
	}
	if sessions.awaiting {
		t.Fatal("expected complaint session to be cleared")
	}
	if len(complaints.requests) != 1 {
		t.Fatalf("got %d complaint requests, want 1", len(complaints.requests))
	}
	req := complaints.requests[0]
	if req.Message != detail || req.Category != "UTILITAS" || req.ResidentID == nil {
		t.Fatalf("unexpected complaint request: %#v", req)
	}
	if len(inbox.replies) != 2 || !strings.Contains(inbox.replies[1], "PGD-20260921-0001") {
		t.Fatalf("unexpected confirmation: %#v", inbox.replies)
	}
}

func TestPendingComplaintCanBeCancelled(t *testing.T) {
	inbox := &botInboxFake{}
	sessions := &botSessionFake{awaiting: true}
	service := NewService(inbox, nil, nil, nil, nil, &complaintCreatorFake{}, sessions, true, nil)

	if err := service.HandleInbound(context.Background(), "628123456789", "", "", "batal"); err != nil {
		t.Fatalf("cancel complaint: %v", err)
	}
	if sessions.awaiting {
		t.Fatal("expected complaint session to be cleared")
	}
	if len(inbox.replies) != 1 || !strings.Contains(inbox.replies[0], "dibatalkan") {
		t.Fatalf("unexpected cancellation reply: %#v", inbox.replies)
	}
}
