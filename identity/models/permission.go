package models

// Permission represents a permission in the system
type Permission struct {
	ID          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Codename    string `json:"codename" db:"codename"`         // e.g., "add_post", "change_post"
	ContentType string `json:"content_type" db:"content_type"` // e.g., "post", "user"
	AppLabel    string `json:"app_label" db:"app_label"`       // e.g., "blog", "users"
}

// GetFullCodename returns the full permission codename (app_label.codename)
func (p *Permission) GetFullCodename() string {
	if p.AppLabel != "" {
		return p.AppLabel + "." + p.Codename
	}
	return p.Codename
}
