package httpserver

import (
	"encoding/csv"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type importGuestsBody struct {
	CSV string `json:"csv" binding:"required"`
}

// postGuestsImport accepts CSV with header row optional. Columns: name, phone (required);
// optional: email, relationship, side (matched by header name).
func (s *Server) postGuestsImport(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var body importGuestsBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	raw := strings.TrimSpace(body.CSV)
	if strings.HasPrefix(raw, "\ufeff") {
		raw = strings.TrimPrefix(raw, "\ufeff")
	}
	r := csv.NewReader(strings.NewReader(raw))
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	records, err := r.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "csv_parse", "detail": err.Error()})
		return
	}
	if len(records) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty_csv"})
		return
	}

	start := 0
	nameIdx, phoneIdx, emailIdx, relIdx, sideIdx := 0, 1, -1, -1, -1
	first := records[0]
	joined := strings.ToLower(strings.Join(first, "|"))
	if strings.Contains(joined, "phone") && strings.Contains(joined, "name") {
		start = 1
		nameIdx, phoneIdx, emailIdx, relIdx, sideIdx = -1, -1, -1, -1, -1
		for i, h := range first {
			hl := strings.ToLower(strings.TrimSpace(h))
			switch hl {
			case "name", "guest", "guest_name":
				nameIdx = i
			case "phone", "mobile", "contact":
				phoneIdx = i
			case "email", "e-mail":
				emailIdx = i
			case "relationship", "relation":
				relIdx = i
			case "side", "bride_groom", "family_side":
				sideIdx = i
			}
		}
		if nameIdx < 0 || phoneIdx < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "csv_header", "detail": "need name and phone columns"})
			return
		}
	} else {
		if len(first) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "csv_row1", "detail": "first data row needs at least name,phone"})
			return
		}
	}

	ctx := c.Request.Context()
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tx_begin"})
		return
	}
	defer tx.Rollback(ctx)

	imported := 0
	errs := []gin.H{}
	for i := start; i < len(records); i++ {
		row := records[i]
		lineNo := i + 1
		maxIdx := phoneIdx
		if nameIdx > maxIdx {
			maxIdx = nameIdx
		}
		if emailIdx > maxIdx {
			maxIdx = emailIdx
		}
		if relIdx > maxIdx {
			maxIdx = relIdx
		}
		if sideIdx > maxIdx {
			maxIdx = sideIdx
		}
		if len(row) <= maxIdx {
			errs = append(errs, gin.H{"line": lineNo, "error": "too_few_columns"})
			continue
		}
		name := strings.TrimSpace(row[nameIdx])
		phone := strings.TrimSpace(row[phoneIdx])
		if phone == "" {
			errs = append(errs, gin.H{"line": lineNo, "error": "missing_phone"})
			continue
		}
		if name == "" {
			name = phone
		}
		email := ""
		if emailIdx >= 0 && emailIdx < len(row) {
			email = strings.TrimSpace(row[emailIdx])
		}
		rel := ""
		if relIdx >= 0 && relIdx < len(row) {
			rel = strings.TrimSpace(row[relIdx])
		}
		side := ""
		if sideIdx >= 0 && sideIdx < len(row) {
			side = strings.TrimSpace(row[sideIdx])
		}
		tags := []string{}
		_, err := tx.Exec(ctx, `
			INSERT INTO guests (event_id, group_id, name, phone, email, relationship, side, tags)
			VALUES ($1,NULL,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),$7::text[])
			ON CONFLICT (event_id, phone) DO UPDATE SET name=EXCLUDED.name, email=EXCLUDED.email,
				relationship=EXCLUDED.relationship, side=EXCLUDED.side, updated_at=now()`,
			eventID, name, phone, email, rel, side, tags)
		if err != nil {
			errs = append(errs, gin.H{"line": lineNo, "error": err.Error()})
			continue
		}
		imported++
	}
	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"imported": imported, "errors": errs})
}
