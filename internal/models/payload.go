// File: internal/models/payload.go
package models

// ReservationPayload represents the request payload for reserving a server name
type ReservationPayload struct {
	UnitCode    string `json:"unitCode,omitempty" validate:"omitempty,max=3"`
	Type        string `json:"type,omitempty" validate:"omitempty,max=1"`
	Provider    string `json:"provider,omitempty" validate:"omitempty,max=1"`
	Region      string `json:"region,omitempty" validate:"omitempty,max=4"`
	Environment string `json:"environment,omitempty" validate:"omitempty,max=1"`
	Function    string `json:"function,omitempty" validate:"omitempty,max=2"`
}

// CommitPayload represents the request payload for committing a reservation
type CommitPayload struct {
	ReservationID string `json:"reservationId" validate:"required,uuid"`
}

// ReleasePayload represents the request payload for releasing a reservation
type ReleasePayload struct {
	ReservationID string `json:"reservationId" validate:"required,uuid"`
}
