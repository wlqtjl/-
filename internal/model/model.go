package model

import "time"

// User represents a registered user.
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Nickname     string    `json:"nickname"`
	Avatar       string    `json:"avatar,omitempty"`
	Bio          string    `json:"bio,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Soul represents a digital soul persona.
type Soul struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Relation    string    `json:"relation"`
	Personality string    `json:"personality"`
	SpeechStyle string    `json:"speech_style"`
	Memory      string    `json:"memory"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Message represents a chat message.
type Message struct {
	ID        int64     `json:"id"`
	SoulID    int64     `json:"soul_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Sentiment string    `json:"sentiment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// SoulHistory records edits to a soul.
type SoulHistory struct {
	ID          int64     `json:"id"`
	SoulID      int64     `json:"soul_id"`
	UserID      int64     `json:"user_id"`
	FieldName   string    `json:"field_name"`
	OldValue    string    `json:"old_value"`
	NewValue    string    `json:"new_value"`
	CreatedAt   time.Time `json:"created_at"`
}

// AuditLog records auditable user actions.
type AuditLog struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Detail    string    `json:"detail,omitempty"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatStats holds aggregated statistics.
type ChatStats struct {
	TotalUsers    int64 `json:"total_users"`
	TotalSouls    int64 `json:"total_souls"`
	TotalMessages int64 `json:"total_messages"`
	ActiveToday   int64 `json:"active_today"`
}
