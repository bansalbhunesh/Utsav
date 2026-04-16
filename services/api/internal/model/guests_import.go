package model

// Guests CSV import payload used by JSON import mode.
type GuestsImportBody struct {
	CSV string `json:"csv" binding:"required"`
}

