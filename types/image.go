package types

import (
	"github.com/google/uuid"
	"time"
)

type ImageStatus int

const (
	ImageStatusPending ImageStatus = iota
	ImageStatusFailed
	ImageStatusCompleted
)

type Image struct {
	ID            int `bun:"id,pk,autoincrement"`
	UserID        uuid.UUID
	BatchID       uuid.UUID
	Status        ImageStatus
	ImageLocation string
	Prompt        string
	Deleted       bool      `bun:"default:'false'"`
	CreatedAt     time.Time `bun:"default:'now()'"`
	DeletedAt     time.Time
}
