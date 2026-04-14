package httpserver

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWriteAPIErrorEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)

	writeAPIError(ctx, 401, "UNAUTHORIZED", "Missing token.")

	if rec.Code != 401 {
		t.Fatalf("unexpected status: got %d", rec.Code)
	}

	var payload struct {
		Success bool `json:"success"`
		Error   struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Success {
		t.Fatal("expected success=false")
	}
	if payload.Error.Code != "UNAUTHORIZED" {
		t.Fatalf("unexpected error code: %s", payload.Error.Code)
	}
	if payload.Error.Message != "Missing token." {
		t.Fatalf("unexpected error message: %s", payload.Error.Message)
	}
}
