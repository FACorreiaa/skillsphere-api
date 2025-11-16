package ontology

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	sessionTypeIRI = "sk:Session"
	matchTypeIRI   = "sk:Match"
)

func sessionIRI(id uuid.UUID) string {
	return fmt.Sprintf("sk:Session/%s", id.String())
}

// SessionEvent models the data required to describe a session resource.
type SessionEvent struct {
	SessionID       uuid.UUID
	InitiatorID     uuid.UUID
	PartnerID       uuid.UUID
	InitiatorOffers []string
	PartnerOffers   []string
	ScheduledStart  time.Time
	ScheduledEnd    time.Time
	MeetingURL      string
	StatusIRI       string
	IsPremium       bool
	Notes           string
}

// NewSessionScheduledEvent builds an ontology event for a session lifecycle change.
func NewSessionScheduledEvent(payload SessionEvent) Event {
	evt := NewEvent(sessionIRI(payload.SessionID), sessionTypeIRI)
	evt.SetTimestamp(time.Now())
	evt.Set("sk:initiatedBy", userIRI(payload.InitiatorID))
	evt.Set("sk:hasParticipant", []string{
		userIRI(payload.InitiatorID),
		userIRI(payload.PartnerID),
	})
	if payload.StatusIRI != "" {
		evt.Set("sk:sessionStatus", payload.StatusIRI)
	}
	if !payload.ScheduledStart.IsZero() {
		evt.Set("sk:scheduledStart", payload.ScheduledStart.UTC().Format(time.RFC3339Nano))
	}
	if !payload.ScheduledEnd.IsZero() {
		evt.Set("sk:scheduledEnd", payload.ScheduledEnd.UTC().Format(time.RFC3339Nano))
	}
	if payload.MeetingURL != "" {
		evt.Set("sk:meetingUrl", payload.MeetingURL)
	}
	if payload.Notes != "" {
		evt.Set("sk:sessionNotes", payload.Notes)
	}
	if len(payload.InitiatorOffers) > 0 {
		evt.Set("sk:initiatorOffers", payload.InitiatorOffers)
	}
	if len(payload.PartnerOffers) > 0 {
		evt.Set("sk:partnerOffers", payload.PartnerOffers)
	}
	evt.Set("sk:isPremium", payload.IsPremium)
	return evt
}

// SkillMatchEvent captures the overlap between two profiles.
type SkillMatchEvent struct {
	SkillName        string
	UserProficiency  int
	MatchProficiency int
	IsComplementary  bool
}

// MatchEvent describes the output from the Matching service.
type MatchEvent struct {
	MatchID      uuid.UUID
	RequesterID  uuid.UUID
	CandidateID  uuid.UUID
	AlgorithmIRI string
	Score        float64
	SkillMatches []SkillMatchEvent
	Explanation  string
}

// NewMatchEvent builds a JSON-LD representation for a computed match.
func NewMatchEvent(payload MatchEvent) Event {
	eventID := payload.MatchID.String()
	if eventID == "" {
		eventID = fmt.Sprintf("%s-%s-%s", payload.RequesterID, payload.CandidateID, time.Now().UTC().Format(time.RFC3339Nano))
	}
	evt := NewEvent(fmt.Sprintf("sk:Match/%s", eventID), matchTypeIRI)
	evt.SetTimestamp(time.Now())
	evt.Set("sk:hasMatch", userIRI(payload.RequesterID))
	evt.Set("sk:matchTarget", userIRI(payload.CandidateID))
	if payload.AlgorithmIRI != "" {
		evt.Set("sk:matchingAlgorithm", payload.AlgorithmIRI)
	}
	if payload.Score > 0 {
		evt.Set("sk:matchScore", payload.Score)
	}
	if payload.Explanation != "" {
		evt.Set("sk:explanation", payload.Explanation)
	}
	if len(payload.SkillMatches) > 0 {
		matches := make([]map[string]any, 0, len(payload.SkillMatches))
		for _, sm := range payload.SkillMatches {
			matches = append(matches, map[string]any{
				"sk:skillName":        sm.SkillName,
				"sk:userProficiency":  sm.UserProficiency,
				"sk:matchProficiency": sm.MatchProficiency,
				"sk:isComplementary":  sm.IsComplementary,
			})
		}
		evt.Set("sk:skillMatches", matches)
	}
	return evt
}
