package authrepo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OTPChallenge struct {
	ID        uuid.UUID
	CodeHash  string
	ExpiresAt time.Time
}

type Repository interface {
	DeletePhoneOTPChallenges(ctx context.Context, phone string) error
	InsertPhoneOTPChallenge(ctx context.Context, phone, codeHash string) error
	GetLatestPhoneOTPChallenge(ctx context.Context, phone string) (*OTPChallenge, error)
	IncrementPhoneOTPAttempts(ctx context.Context, id uuid.UUID) error
	DeletePhoneOTPChallengeByID(ctx context.Context, id uuid.UUID) error
	FindUserIDByPhone(ctx context.Context, phone string) (uuid.UUID, error)
	CreateUserWithPhone(ctx context.Context, phone string) (uuid.UUID, error)
	InsertRefreshTokenHash(ctx context.Context, userID uuid.UUID, tokenHash string) error
	ConsumeRefreshTokenHash(ctx context.Context, tokenHash string) (uuid.UUID, error)
	GetUserProfileByID(ctx context.Context, userID uuid.UUID) (string, string, error)
}

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

func (r *PGRepository) DeletePhoneOTPChallenges(ctx context.Context, phone string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM phone_otp_challenges WHERE phone=$1`, phone)
	return err
}

func (r *PGRepository) InsertPhoneOTPChallenge(ctx context.Context, phone, codeHash string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO phone_otp_challenges (phone, code_hash, expires_at)
		VALUES ($1, $2, now() + interval '10 minutes')`, phone, codeHash)
	return err
}

func (r *PGRepository) GetLatestPhoneOTPChallenge(ctx context.Context, phone string) (*OTPChallenge, error) {
	var ch OTPChallenge
	err := r.pool.QueryRow(ctx, `
		SELECT id, code_hash, expires_at FROM phone_otp_challenges
		WHERE phone=$1 ORDER BY created_at DESC LIMIT 1`, phone).
		Scan(&ch.ID, &ch.CodeHash, &ch.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *PGRepository) IncrementPhoneOTPAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE phone_otp_challenges SET attempts=attempts+1 WHERE id=$1`, id)
	return err
}

func (r *PGRepository) DeletePhoneOTPChallengeByID(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM phone_otp_challenges WHERE id=$1`, id)
	return err
}

func (r *PGRepository) FindUserIDByPhone(ctx context.Context, phone string) (uuid.UUID, error) {
	var userID uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM users WHERE phone=$1`, phone).Scan(&userID)
	return userID, err
}

func (r *PGRepository) CreateUserWithPhone(ctx context.Context, phone string) (uuid.UUID, error) {
	var userID uuid.UUID
	err := r.pool.QueryRow(ctx, `INSERT INTO users (phone) VALUES ($1) RETURNING id`, phone).Scan(&userID)
	return userID, err
}

func (r *PGRepository) InsertRefreshTokenHash(ctx context.Context, userID uuid.UUID, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, now() + interval '30 days')`,
		userID, tokenHash)
	return err
}

func (r *PGRepository) ConsumeRefreshTokenHash(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	var userID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		DELETE FROM refresh_tokens WHERE token_hash=$1 AND expires_at > now()
		RETURNING user_id`, tokenHash).Scan(&userID)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func (r *PGRepository) GetUserProfileByID(ctx context.Context, userID uuid.UUID) (string, string, error) {
	var phone, displayName string
	err := r.pool.QueryRow(ctx, `SELECT phone, COALESCE(display_name,'') FROM users WHERE id=$1`, userID).Scan(&phone, &displayName)
	return phone, displayName, err
}

func IsNoRows(err error) bool {
	return err == pgx.ErrNoRows
}
