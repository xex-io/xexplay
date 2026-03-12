package domain

import "time"

type Setting struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description string    `json:"description,omitempty"`
	IsSecret    bool      `json:"is_secret"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SettingView is the client-facing representation that masks secret values.
type SettingView struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description string    `json:"description,omitempty"`
	IsSecret    bool      `json:"is_secret"`
	HasValue    bool      `json:"has_value"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToView returns a client-safe view, masking secret values.
func (s *Setting) ToView() *SettingView {
	v := &SettingView{
		Key:         s.Key,
		Description: s.Description,
		IsSecret:    s.IsSecret,
		HasValue:    s.Value != "",
		UpdatedAt:   s.UpdatedAt,
	}
	if s.IsSecret {
		if s.Value != "" {
			v.Value = "••••••••" + s.Value[max(0, len(s.Value)-4):]
		}
	} else {
		v.Value = s.Value
	}
	return v
}
