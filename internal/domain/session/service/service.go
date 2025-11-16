package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/FACorreiaa/skillsphere-api/internal/ontology"
)

// Repository defines the persistence operations required by the service.
type Repository interface {
	CreateSession(ctx context.Context, session Session) (*Session, error)
}

// Session models the persisted session.
type Session struct {
	ID              uuid.UUID
	InitiatorID     uuid.UUID
	PartnerID       uuid.UUID
	InitiatorOffers []string
	PartnerOffers   []string
	ScheduledStart  time.Time
	ScheduledEnd    time.Time
	MeetingURL      string
	Notes           string
	StatusIRI       string
	CreatedAt       time.Time
	IsPremium       bool
}

// Service coordinates session orchestration and ontology emission.
type Service struct {
	repo    Repository
	emitter ontology.Emitter
}

// NewService builds a session Service.
func NewService(repo Repository, emitter ontology.Emitter) *Service {
	if emitter == nil {
		emitter = ontology.NopEmitter{}
	}
	return &Service{repo: repo, emitter: emitter}
}

// CreateSession persists the session and emits a JSON-LD envelope.
func (s *Service) CreateSession(ctx context.Context, session Session) (*Session, error) {
	saved, err := s.repo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	event := ontology.NewSessionScheduledEvent(ontology.SessionEvent{
		SessionID:       saved.ID,
		InitiatorID:     saved.InitiatorID,
		PartnerID:       saved.PartnerID,
		InitiatorOffers: saved.InitiatorOffers,
		PartnerOffers:   saved.PartnerOffers,
		ScheduledStart:  saved.ScheduledStart,
		ScheduledEnd:    saved.ScheduledEnd,
		MeetingURL:      saved.MeetingURL,
		StatusIRI:       saved.StatusIRI,
		IsPremium:       saved.IsPremium,
		Notes:           saved.Notes,
	})
	_ = s.emitter.Emit(ctx, event)

	return saved, nil
}
