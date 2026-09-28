package models

// Group represents a user group
type Group struct {
	ID          int64        `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	Permissions []Permission `json:"permissions" db:"-"` // Loaded separately
}
