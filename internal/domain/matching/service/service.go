package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/FACorreiaa/skillsphere-api/internal/ontology"
)

// RecommendationRepository persists match results for auditing/caching.
type RecommendationRepository interface {
	SaveMatch(ctx context.Context, match Match) (*Match, error)
}

// Match represents a computed recommendation.
type Match struct {
	ID           uuid.UUID
	RequesterID  uuid.UUID
	CandidateID  uuid.UUID
	AlgorithmIRI string
	Score        float64
	Explanation  string
	SkillMatches []ontology.SkillMatchEvent
}

// Service coordinates match generation and ontology emission.
type Service struct {
	repo    RecommendationRepository
	emitter ontology.Emitter
}

// NewService constructs the matching service.
func NewService(repo RecommendationRepository, emitter ontology.Emitter) *Service {
	if emitter == nil {
		emitter = ontology.NopEmitter{}
	}
	return &Service{repo: repo, emitter: emitter}
}

// RecordMatch writes a match and emits the ontology payload.
func (s *Service) RecordMatch(ctx context.Context, match Match) (*Match, error) {
	saved, err := s.repo.SaveMatch(ctx, match)
	if err != nil {
		return nil, err
	}
	event := ontology.NewMatchEvent(ontology.MatchEvent{
		MatchID:      saved.ID,
		RequesterID:  saved.RequesterID,
		CandidateID:  saved.CandidateID,
		AlgorithmIRI: saved.AlgorithmIRI,
		Score:        saved.Score,
		SkillMatches: saved.SkillMatches,
		Explanation:  saved.Explanation,
	})
	_ = s.emitter.Emit(ctx, event)
	return saved, nil
}
