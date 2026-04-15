package publicservice

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/media"
	"github.com/bhune/utsav/services/api/internal/repository/publicrepo"
)

type mockPublicRepo struct {
	getUPIContextBySlugFn    func(ctx context.Context, slug string) (*publicrepo.UPIContext, error)
	getEventBySlugFn         func(ctx context.Context, slug string) (*publicrepo.PublicEvent, error)
	insertGuestShagunReportFn func(ctx context.Context, in publicrepo.GuestShagunReportInput) error
}

func (m *mockPublicRepo) GetSlugByEventID(context.Context, uuid.UUID) (string, error) {
	return "", errors.New("not implemented")
}

func (m *mockPublicRepo) ResolveEventIDBySlug(ctx context.Context, slug string) (uuid.UUID, error) {
	if m.getEventBySlugFn == nil {
		return uuid.Nil, errors.New("not implemented")
	}
	ev, err := m.getEventBySlugFn(ctx, slug)
	if err != nil {
		return uuid.Nil, err
	}
	return ev.ID, nil
}

func (m *mockPublicRepo) GetEventBySlug(ctx context.Context, slug string) (*publicrepo.PublicEvent, error) {
	if m.getEventBySlugFn != nil {
		return m.getEventBySlugFn(ctx, slug)
	}
	return nil, errors.New("not implemented")
}
func (m *mockPublicRepo) ListSubEvents(context.Context, uuid.UUID) ([]publicrepo.PublicSubEvent, error) {
	return nil, nil
}
func (m *mockPublicRepo) ListBroadcasts(context.Context, uuid.UUID) ([]publicrepo.PublicBroadcast, error) {
	return nil, nil
}
func (m *mockPublicRepo) ListApprovedGallery(context.Context, uuid.UUID) ([]publicrepo.PublicGalleryAsset, error) {
	return nil, nil
}
func (m *mockPublicRepo) GetUPIContextBySlug(ctx context.Context, slug string) (*publicrepo.UPIContext, error) {
	if m.getUPIContextBySlugFn != nil {
		return m.getUPIContextBySlugFn(ctx, slug)
	}
	return nil, errors.New("not implemented")
}
func (m *mockPublicRepo) InsertGuestShagunReport(ctx context.Context, in publicrepo.GuestShagunReportInput) error {
	if m.insertGuestShagunReportFn != nil {
		return m.insertGuestShagunReportFn(ctx, in)
	}
	return nil
}

func TestBuildUPILinkSuccess(t *testing.T) {
	eid := uuid.New()
	svc := NewService(&mockPublicRepo{
		getUPIContextBySlugFn: func(_ context.Context, slug string) (*publicrepo.UPIContext, error) {
			if slug != "wedding-slug" {
				t.Fatalf("unexpected slug: %s", slug)
			}
			return &publicrepo.UPIContext{EventID: eid, VPA: "host@upi", Title: "Utsav Wedding"}, nil
		},
	}, media.URLSigner{BaseURL: "https://cdn.example.com"}, nil)

	out, svcErr := svc.BuildUPILink(context.Background(), "Wedding-Slug", eid, "9876543210")
	if svcErr != nil {
		t.Fatalf("unexpected error: %+v", svcErr)
	}
	if out["payee_vpa"] != "host@upi" {
		t.Fatalf("expected payee_vpa host@upi, got %v", out["payee_vpa"])
	}
	if out["guest_phone_masked"] != "******3210" {
		t.Fatalf("unexpected phone mask: %v", out["guest_phone_masked"])
	}
}

func TestReportShagunRejectsWrongEvent(t *testing.T) {
	eventID := uuid.New()
	otherEventID := uuid.New()
	svc := NewService(&mockPublicRepo{
		getEventBySlugFn: func(_ context.Context, _ string) (*publicrepo.PublicEvent, error) {
			return &publicrepo.PublicEvent{ID: eventID, Slug: "slug"}, nil
		},
	}, media.URLSigner{}, nil)

	err := svc.ReportShagun(context.Background(), "slug", otherEventID, "9999999999", 501, "blessing", nil)
	if err == nil {
		t.Fatal("expected wrong event error")
	}
	if err.Code != "WRONG_EVENT" {
		t.Fatalf("expected WRONG_EVENT, got %s", err.Code)
	}
}
