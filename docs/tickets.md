# SkillSphere API Linear Tickets

This document tracks every implementation ticket we plan to push into Linear. Tickets are split into two buckets so they can be copied directly into Linear projects:
- **TODO (MVP Scope):** must-ship features for the first public launch (Weeks 1-4).
- **Backlog (Post-MVP):** V1/V2 enhancements that will sit in the backlog until the MVP is stable.

Every ticket keeps the same structure Linear expects (priority, complexity, description, tasks, acceptance criteria, dependencies, related files).

---

## TODO (MVP Scope)

### AUTH-001 – Complete Auth Service
**Status:** TODO · **Priority:** P0 (Critical) · **Complexity:** Medium · **Service:** `auth.v1.AuthService`

**Description:** Finalize the authentication surface (password + OAuth) and harden it with interceptors, rate limiting, and tests.

**Tasks:**
- [ ] Verify each RPC implementation (Register, Login, Logout, RefreshToken, ValidateToken, OAuthLogin, RequestPasswordReset, ResetPassword, ChangePassword, VerifyEmail, ResendVerificationEmail)
- [ ] Add JWT validation interceptor and rate limiting
- [ ] Wire OAuth (Google, GitHub) + password reset + email verification flows
- [ ] Improve error handling/logging
- [ ] Unit + integration tests

**Acceptance Criteria:** All RPCs succeed end-to-end, JWTs validate correctly, OAuth + password reset + email verification flows work, rate limiting blocks brute-force attempts, and tests pass.

**Dependencies:** None

**Related Files:** `internal/domain/auth/**`, `gen/.../auth/v1`

---

### USER-001 – Implement User Service
**Status:** TODO · **Priority:** P0 (Critical) · **Complexity:** Large · **Service:** `user.v1.UserService`

**Description:** Deliver complete profile management, stats, availability, notifications, embeddings, and list/query endpoints.

**Tasks:**
- [ ] Create DDD folders (handler/service/repository/presenter)
- [ ] Implement every RPC (Create/Get/BatchGet/Update/Delete/List, stats, availability, notification prefs, sessions, ratings, verification)
- [ ] Generate embeddings + avatar uploads + caching + search filters
- [ ] Unit + integration tests

**Acceptance Criteria:** CRUD + stats + prefs work, embeddings/avatars/caching function, pagination/filtering works, tests >80% coverage.

**Dependencies:** AUTH-001

**Related Files:** `pkg/db/migrations/002,008,010-013`, `gen/.../user/v1`

---

### SKILL-001 – Implement Skill Service
**Status:** TODO · **Priority:** P0 (Critical) · **Complexity:** Large · **Service:** `skill.v1.SkillService`

**Description:** Manage the skill catalog, categories/tags, user skills, embeddings, and search/trending endpoints.

**Tasks:**
- [ ] Create DDD folders
- [ ] Implement RPCs (Create/Get/List/Update/Delete Skill, user-skill CRUD, trending, category, search)
- [ ] Integrate Gemini embeddings, category/tag management, pg_trgm search, trending calc
- [ ] Unit + integration tests

**Acceptance Criteria:** Catalog CRUD & user skill CRUD work, proficiency validated, embeddings/trending/search operate, tests >80%.

**Dependencies:** USER-001

**Related Files:** `pkg/db/migrations/009`, `gen/.../skill/v1`

---

### MATCH-001 – Implement Matching Service
**Status:** TODO · **Priority:** P0 (Critical) · **Complexity:** Large · **Service:** `matching.v1.MatchingService`

**Description:** Build the matching engine (algorithms, scoring, caching, history, vector search).

**Tasks:**
- [ ] Create DDD folders + dedicated algorithms package
- [ ] Implement RPCs (FindMatches, GetMatchScore, GetRecommendations, GetSimilarUsers, GenerateEmbeddings)
- [ ] Deliver Euclidean + cosine + AI embeddings, vector search via pgvector, recommendation cache/history
- [ ] Unit + integration tests

**Acceptance Criteria:** Users receive relevant matches backed by multiple algorithms, embeddings/caching/history working, tests >80%.

**Dependencies:** USER-001, SKILL-001, AI-001 (embeddings can stub initially)

**Related Files:** `pkg/db/migrations/015`, `gen/.../matching/v1`

---

### SEARCH-001 – Implement Search Service
**Status:** TODO · **Priority:** P0 (Critical) · **Complexity:** Medium · **Service:** `search.v1.SearchService`

**Description:** Provide user/skill search with ranking, suggestions, filters, and history.

**Tasks:**
- [ ] Create DDD folders
- [ ] Implement RPCs (SearchUsers/Skills, TrendingSkills, FeaturedUsers, Suggestions, AdvancedSearch)
- [ ] Build pg_trgm search, ranking, autocomplete, filters, history logging, featured management
- [ ] Unit + integration tests

**Acceptance Criteria:** Search works with filters + ranking + suggestions, trending & featured visible, tests >80%.

**Dependencies:** USER-001, SKILL-001

**Related Files:** `pkg/db/migrations/017`, `docs/SEARCH_*.md`, `gen/.../search/v1`

---

### SESSION-001 – Implement Session Service
**Status:** TODO · **Priority:** P0 (Critical) · **Complexity:** Large · **Service:** `session.v1.SessionService`

**Description:** Implement session scheduling, lifecycle, reminders, analytics, ratings.

**Tasks:**
- [ ] Create DDD folders
- [ ] Implement RPCs (Create/Get/Update/Cancel/Start/Complete/ListUserSessions/Rate/Report/GetUpcoming)
- [ ] Add scheduling logic, status state machine, reminders + meeting URL generation, analytics
- [ ] Unit + integration tests

**Acceptance Criteria:** Sessions progress through lifecycle, cancellations & ratings work, reminders + analytics available, tests >80%.

**Dependencies:** USER-001, SKILL-001

**Related Files:** `pkg/db/migrations/014`, `gen/.../session/v1`

---

### CHAT-001 – Implement Chat Service
**Status:** TODO · **Priority:** P0 (Critical) · **Complexity:** Large · **Service:** `chat.v1.ChatService`

**Description:** Real-time messaging with conversations, read receipts, reactions, moderation.

**Tasks:**
- [ ] Create DDD folders
- [ ] Implement RPCs (Send/Get/Stream messages, MarkAsRead, ListConversations, Delete, React, Report)
- [ ] Add WebSocket/streaming support, conversations, read status, reactions, attachments, search, typing indicators, moderation
- [ ] Unit + integration tests

**Acceptance Criteria:** Messages deliver in real time, histories/reactions/attachments/typing/ reporting work, tests >80%.

**Dependencies:** USER-001

**Related Files:** `pkg/db/migrations/018`, `gen/.../chat/v1`

---

### INFRA-001 – Implement Interceptors/Middleware
**Status:** TODO · **Priority:** P0 (Critical) · **Complexity:** Medium

**Description:** Cross-cutting Connect interceptors for auth, logging, rate limiting, metrics, validation, errors, tracing.

**Tasks:** Build interceptor files, add JWT validation, structured logging, rate limiting, Prometheus metrics, validation, error mapping, tracing hooks, tests.

**Acceptance Criteria:** All requests flow through consistent middleware stack; metrics/logging/auth/rate limiting validated.

**Dependencies:** AUTH-001

---

### INFRA-002 – Implement Database Layer
**Status:** TODO · **Priority:** P0 (Critical) · **Complexity:** Medium

**Description:** Shared DB utilities: pooling, migrations, transactions, health, metrics.

**Tasks:** Build `pkg/db` helpers (connection, transaction, migration runner, health), add retries/query logging/metrics, tests.

**Acceptance Criteria:** Reliable pooled connections, migrations + health + metrics in place, tests green.

**Dependencies:** None

---

### TEST-001 – Integration Test Suite
**Status:** TODO · **Priority:** P0 (Critical) · **Complexity:** Large

**Description:** Shared integration testing harness + CI automation.

**Tasks:** Add integration test scaffolding, fixtures, isolated DB setup/teardown, contract/perf tests, CI job.

**Acceptance Criteria:** Every service has integration coverage (>80%), tests run in CI per commit.

**Dependencies:** Service implementations

---

### PAYMENT-001 – Implement Payment Service
**Status:** TODO · **Priority:** P0 (Critical) · **Complexity:** Large · **Service:** `payment.v1.PaymentService`

**Description:** Stripe-powered subscriptions, one-time payments, escrow, payouts, invoices.

**Tasks:** Build DDD stack + Stripe client, implement subscription/payment/escrow/payout RPCs, webhook handlers, tier management, retries, invoices, tests.

**Acceptance Criteria:** Subscriptions + payments + escrow + payouts + invoices run via Stripe with error handling + tests.

**Dependencies:** USER-001

**Related Files:** `pkg/db/migrations/020-021`, `gen/.../payment/v1`

---

### AI-001 – Implement AI Service
**Status:** TODO · **Priority:** P0 (Critical) · **Complexity:** Medium · **Service:** `ai.v1.AIService`

**Description:** Gemini-powered embeddings, semantic recommendations, learning/advice APIs.

**Tasks:** Build DDD stack + Gemini client, implement AI RPCs (GenerateSkillEmbedding, GenerateUserEmbedding, GetSkillRecommendations, GenerateLearningPath, GetSessionAdvice, AnalyzeSkillGap), caching, rate limiting, tests.

**Acceptance Criteria:** Embeddings + AI recommendations run in production with caching/cost controls; learning paths/advice/skill gap outputs are usable; tests >80%.

**Dependencies:** SKILL-001, USER-001

**Related Files:** `gen/.../ai/v1`

---

## Backlog (Post-MVP)

---

### GIG-001 – Implement Gig Service
**Status:** Backlog · **Priority:** P1 · **Complexity:** Large · **Service:** `gig.v1.GigService`

**Description:** Freelance gig marketplace with applications, submissions, revisions, disputes.

**Tasks:** Build DDD stack, implement gig/app/submission/revision workflows + search/filtering + analytics + tests.

**Acceptance Criteria:** Creators list gigs, freelancers apply/deliver, workflows/disputes/search operate, tests >80%.

**Dependencies:** USER-001, SKILL-001, PAYMENT-001

**Related Files:** `pkg/db/migrations/019`, `gen/.../gig/v1`

---

### REVIEW-001 – Implement Review Service
**Status:** Backlog · **Priority:** P1 · **Complexity:** Medium · **Service:** `review.v1.ReviewService`

**Description:** Paid/asynchronous reviews with content, sections, revisions, ratings.

**Tasks:** Build DDD layers, implement review RPCs, support attachments/sections/revisions/ratings, tests.

**Acceptance Criteria:** Reviews requested/delivered with revisions & ratings; tests >80%.

**Dependencies:** USER-001, SKILL-001, PAYMENT-001

**Related Files:** `pkg/db/migrations/016`, `gen/.../review/v1`

---


### INFRA-003 – Cloud Storage Integration
**Status:** Backlog · **Priority:** P1 · **Complexity:** Small

**Description:** Unified storage interface (S3/GCS/local) for avatars, chat/media, workshop assets.

**Tasks:** Implement storage interface + providers, upload validation, signed URLs, deletion, CDN support, tests.

**Acceptance Criteria:** Files upload via configured provider with validation + signed URLs + CDN; tests pass.

**Dependencies:** None

---

### INFRA-004 – Email Service
**Status:** Backlog · **Priority:** P1 · **Complexity:** Small

**Description:** Email abstraction (SMTP/SendGrid) with templates, queueing, tracking.

**Tasks:** Build email interface + providers, templates (verification/reset/etc.), queueing + tracking, tests.

**Acceptance Criteria:** Emails send via providers with templating + queueing + tracking; tests pass.

**Dependencies:** AUTH-001

---

### INFRA-005 – Background Jobs
**Status:** Backlog · **Priority:** P1 · **Complexity:** Medium

**Description:** Job queue + worker + scheduler for async tasks (emails, embeddings, analytics, cleanup).

**Tasks:** Build job interface/worker/scheduler/queue, concurrency + retries + monitoring, seed core jobs, tests.

**Acceptance Criteria:** Jobs enqueue/process/retry/schedule with monitoring; tests pass.

**Dependencies:** None

---

### WORKSHOP-001 – Implement Workshop Service
**Status:** Backlog · **Priority:** P2 · **Complexity:** Large · **Service:** `workshop.v1.WorkshopService`

**Description:** Group workshops with ticketing, capacity, materials, recordings.

**Tasks:** Build DDD stack, implement workshop RPCs, manage capacity/registrations/material uploads/recordings, integrate payments for tickets, tests.

**Acceptance Criteria:** Hosts create/manage workshops, attendees register/pay, materials stored, recordings tracked, tests >80%.

**Dependencies:** USER-001, SKILL-001, PAYMENT-001

**Related Files:** `gen/.../workshop/v1`

---

### CERT-001 – Implement Certification Service
**Status:** Backlog · **Priority:** P2 · **Complexity:** Medium · **Service:** `certification.v1.CertificationService`

**Description:** Issue/verify badges (optional blockchain), templates, PDFs, sharing.

**Tasks:** Build DDD stack, implement certification RPCs (issue/get/verify/revoke/list/templates), PDF generation, blockchain hooks, LinkedIn share, tests.

**Acceptance Criteria:** Certifications issued/verified/shared with templates + optional blockchain; tests >80%.

**Dependencies:** USER-001, SESSION-001, PAYMENT-001

**Related Files:** `gen/.../certification/v1`

---

### CHALLENGE-001 – Implement Challenge Service
**Status:** Backlog · **Priority:** P2 · **Complexity:** Medium · **Service:** `challenge.v1.ChallengeService`

**Description:** Gamified challenges with leaderboards, rewards, sponsors.

**Tasks:** Build DDD stack, implement challenge RPCs (create/list/join/submit/leaderboard/trending), leaderboards, rewards, sponsor hooks, tests.

**Acceptance Criteria:** Users join/compete, leaderboards/rewards work, trending challenges visible, tests >80%.

**Dependencies:** USER-001, SKILL-001

**Related Files:** `gen/.../challenge/v1`

---

### ANALYTICS-001 – Implement Analytics Service
**Status:** Backlog · **Priority:** P2 · **Complexity:** Medium · **Service:** `analytics.v1.AnalyticsService`

**Description:** Platform + revenue + engagement analytics with TimescaleDB aggregation + exports.

**Tasks:** Build DDD stack, implement analytics RPCs (platform stats, user analytics, skill trends, session + revenue + engagement metrics), timeseries storage, aggregation jobs, exports, tests.

**Acceptance Criteria:** Accurate platform/user/skill/revenue metrics + export support; tests >80%.

**Dependencies:** USER-001, SKILL-001, SESSION-001, PAYMENT-001

**Related Files:** `gen/.../analytics/v1`

---

### ADMIN-001 – Implement Admin Service
**Status:** Backlog · **Priority:** P2 · **Complexity:** Large · **Service:** `admin.v1.AdminService`

**Description:** Moderation tools (users/content/disputes), featured management, audit logs.

**Tasks:** Build DDD stack, implement admin RPCs (user/content moderation, report handling, dispute resolution, featured management, queues, platform health), audit logging, tests.

**Acceptance Criteria:** Admins moderate/manage disputes/reports with auditing; tests >80%.

**Dependencies:** USER-001 + downstream services

**Related Files:** `pkg/db/migrations/005-006`, `gen/.../admin/v1`

---

### DOC-001 – API Documentation
**Status:** Backlog · **Priority:** P1 · **Complexity:** Small

**Description:** Publish API docs (buf-generated), examples, tutorials, auth docs, error codes.

**Tasks:** Generate buf docs, host doc site, add examples/tutorials/auth/error sections, ensure continuous updates.

**Acceptance Criteria:** Full API docs live with examples/auth guidance and ongoing updates.

**Dependencies:** None

---

## Summary
- **TODO (MVP):** 12 tickets (Auth, User, Skill, Matching, Search, Session, Chat, Infra-001, Infra-002, Test suite, Payments, AI)
- **Backlog:** 10 tickets (Gig, Review, Infra-003/4/5, Workshop, Certification, Challenge, Analytics, Admin, Docs)

Copy any ticket block straight into Linear with minimal edits (status/priority already set). Update statuses in this doc as work progresses.
