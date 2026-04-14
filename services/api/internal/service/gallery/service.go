package galleryservice

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/media"
	"github.com/bhune/utsav/services/api/internal/repository/galleryrepo"
)

type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

type Service struct {
	repo   galleryrepo.Repository
	signer media.Signer
}

func NewService(repo galleryrepo.Repository, signer media.Signer) *Service {
	return &Service{repo: repo, signer: signer}
}

func sanitizeFileName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	repl := strings.NewReplacer(" ", "-", "..", "", "\\", "", "/", "", ":", "", ";", "")
	s = repl.Replace(s)
	if s == "" {
		return "asset"
	}
	return s
}

func normalizeStatus(status string) (string, bool) {
	s := strings.TrimSpace(strings.ToLower(status))
	if s == "" {
		return "pending", true
	}
	if s == "pending" || s == "approved" || s == "rejected" {
		return s, true
	}
	return "", false
}

func (s *Service) PresignPut(eventID uuid.UUID, fileName, contentType string) (any, *ServiceError) {
	key := fmt.Sprintf("events/%s/gallery/%d-%s", eventID.String(), time.Now().Unix(), sanitizeFileName(fileName))
	resp, err := s.signer.PresignPut(media.PresignRequest{
		ObjectKey:   key,
		ContentType: contentType,
		ExpiresIn:   10 * time.Minute,
	})
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "PRESIGN_FAILED", Message: "Failed to generate upload URL."}
	}
	return resp, nil
}

func (s *Service) CreateAsset(ctx context.Context, in galleryrepo.CreateAssetInput) *ServiceError {
	status, ok := normalizeStatus(in.Status)
	if !ok {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_STATUS", Message: "Gallery status is invalid."}
	}
	in.Status = status
	if err := s.repo.CreateAsset(ctx, in); err != nil {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INSERT_FAILED", Message: "Failed to register gallery asset."}
	}
	return nil
}

func (s *Service) ListAssets(ctx context.Context, eventID uuid.UUID, status string) ([]map[string]any, *ServiceError) {
	assets, err := s.repo.ListAssets(ctx, eventID, status)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to list gallery assets."}
	}
	out := make([]map[string]any, 0, len(assets))
	for _, a := range assets {
		out = append(out, map[string]any{
			"id":         a.ID,
			"section":    a.Section,
			"object_key": a.ObjectKey,
			"status":     a.Status,
			"mime_type":  a.MimeType,
			"bytes":      a.Bytes,
			"created_at": a.CreatedAt,
			"url":        s.signer.PublicObjectURL(a.ObjectKey),
		})
	}
	return out, nil
}

func (s *Service) ModerateAsset(ctx context.Context, eventID, assetID uuid.UUID, status string) *ServiceError {
	st, ok := normalizeStatus(status)
	if !ok {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_STATUS", Message: "Gallery status is invalid."}
	}
	found, err := s.repo.UpdateAssetStatus(ctx, eventID, assetID, st)
	if err != nil {
		return &ServiceError{Status: http.StatusBadRequest, Code: "UPDATE_FAILED", Message: "Failed to update gallery asset."}
	}
	if !found {
		return &ServiceError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: "Gallery asset not found."}
	}
	return nil
}
