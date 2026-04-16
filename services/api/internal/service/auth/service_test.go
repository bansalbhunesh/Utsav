package authservice

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/bhune/utsav/services/api/internal/repository/authrepo"
)

type mockAuthRepo struct {
	getUserProfileByIDFn func(ctx context.Context, userID uuid.UUID) (string, string, error)
}

func (m *mockAuthRepo) DeletePhoneOTPChallenges(context.Context, string) error        { return nil }
func (m *mockAuthRepo) InsertPhoneOTPChallenge(context.Context, string, string) error { return nil }
func (m *mockAuthRepo) GetLatestPhoneOTPChallenge(context.Context, string) (*authrepo.OTPChallenge, error) {
	return nil, nil
}
func (m *mockAuthRepo) IncrementPhoneOTPAttempts(context.Context, uuid.UUID) error   { return nil }
func (m *mockAuthRepo) DeletePhoneOTPChallengeByID(context.Context, uuid.UUID) error { return nil }
func (m *mockAuthRepo) FindUserIDByPhone(context.Context, string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *mockAuthRepo) CreateUserWithPhone(context.Context, string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *mockAuthRepo) InsertRefreshTokenHash(context.Context, uuid.UUID, string) error { return nil }
func (m *mockAuthRepo) PruneRefreshTokensForUser(context.Context, uuid.UUID, int) error { return nil }
func (m *mockAuthRepo) ConsumeRefreshTokenHash(context.Context, string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *mockAuthRepo) GetRefreshTokenUserID(context.Context, string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *mockAuthRepo) RotateRefreshToken(context.Context, string, string, uuid.UUID) error {
	return nil
}
func (m *mockAuthRepo) RevokeRefreshTokenHash(context.Context, string) error { return nil }
func (m *mockAuthRepo) GetUserProfileByID(ctx context.Context, userID uuid.UUID) (string, string, error) {
	if m.getUserProfileByIDFn != nil {
		return m.getUserProfileByIDFn(ctx, userID)
	}
	return "", "", pgx.ErrNoRows
}

func TestGetMeSuccess(t *testing.T) {
	uid := uuid.New()
	svc := NewService(&mockAuthRepo{
		getUserProfileByIDFn: func(_ context.Context, got uuid.UUID) (string, string, error) {
			if got != uid {
				t.Fatalf("unexpected user id: %s", got.String())
			}
			return "9876543210", "Utsav", nil
		},
	}, nil, nil, "123456", "secret", "otp-secret", "test", nil, 5)

	me, err := svc.GetMe(context.Background(), uid)
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	if me.ID != uid.String() || me.Phone != "9876543210" || me.DisplayName != "Utsav" {
		t.Fatalf("unexpected me payload: %+v", me)
	}
}

func TestGetMeNotFound(t *testing.T) {
	svc := NewService(&mockAuthRepo{
		getUserProfileByIDFn: func(_ context.Context, _ uuid.UUID) (string, string, error) {
			return "", "", pgx.ErrNoRows
		},
	}, nil, nil, "123456", "secret", "otp-secret", "test", nil, 5)

	_, err := svc.GetMe(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
	if err.Code != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND code, got %s", err.Code)
	}
}
