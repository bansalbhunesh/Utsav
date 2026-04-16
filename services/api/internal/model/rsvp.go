package model

// Public RSVP OTP request/verify payloads.
type RSVPOTPRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type RSVPOTPVerify struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// RSVP submit payload.
type RSVPItem struct {
	SubEventID           string `json:"sub_event_id" binding:"required"`
	Status               string `json:"status" binding:"required"`
	MealPref             string `json:"meal_pref"`
	Dietary              string `json:"dietary"`
	AccommodationNeeded  bool   `json:"accommodation_needed"`
	TravelMode           string `json:"travel_mode"`
	PlusOneNames         string `json:"plus_one_names"`
}

type RSVPSubmit struct {
	Items []RSVPItem `json:"items" binding:"required"`
}

