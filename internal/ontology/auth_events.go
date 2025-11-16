package ontology

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/FACorreiaa/skillsphere-api/internal/domain/auth/repository"
)

const (
	userTypeIRI = "sk:User"
)

func userIRI(id uuid.UUID) string {
	return fmt.Sprintf("sk:User/%s", id.String())
}

// NewUserRegisteredEvent builds a JSON-LD document for a created user.
func NewUserRegisteredEvent(user *repository.User) Event {
	if user == nil {
		return NewEvent("", userTypeIRI)
	}

	evt := NewEvent(userIRI(user.ID), userTypeIRI)
	evt.SetTimestamp(user.CreatedAt)
	evt.Set("schema:email", user.Email)
	evt.Set("schema:name", user.DisplayName)
	if user.Username != "" {
		evt.Set("schema:alternateName", user.Username)
	}
	if user.AvatarURL != nil && *user.AvatarURL != "" {
		evt.Set("schema:image", *user.AvatarURL)
	}
	if !user.UpdatedAt.IsZero() && !user.UpdatedAt.Equal(user.CreatedAt) {
		evt.Set("schema:dateModified", user.UpdatedAt.UTC().Format(time.RFC3339Nano))
	}
	if user.LastLoginAt != nil {
		evt.Set("schema:lastReviewed", user.LastLoginAt.UTC().Format(time.RFC3339Nano))
	}
	if user.EmailVerifiedAt != nil {
		evt.Set("sk:isVerified", true)
	}
	evt.Set("sk:isActive", user.IsActive)
	evt.Set("sk:role", user.Role)
	return evt
}
