# SkillSphere

![SkillSphere Logo](https://via.placeholder.com/150?text=SkillSphere) <!-- Replace with actual logo URL if available -->

## Overview

SkillSphere is a peer-to-peer (P2P) skill exchange platform that connects users to trade skills in real-time. Whether you're teaching Python programming or learning guitar, SkillSphere facilitates 1:1 exchanges through profiles, matching algorithms, and interactive sessions. Built as a lightweight, scalable web app with modern RPC architecture, it emphasizes simplicity for solo developers while supporting growth into a monetized service.

The platform uses skill matching algorithms (e.g., cosine similarity on skill vectors or embedding-based semantic matching via AI) to recommend partners. It's designed for the gig economy, where users can discover, exchange, and even certify skills in a freemium model.

### Key Features
- **User Profiles**: Create detailed profiles with bio, offered/wanted skills (e.g., programming, languages, arts, cooking, fitness), proficiency levels (1-10 scale), location, and availability.
- **Skill Matching**: Search and get recommendations using algorithms like Euclidean distance, cosine similarity, or AI embeddings for semantic matches (e.g., "ML" matches "machine learning").
- **Discovery**: Users find new skills via a search bar (keyword or category-based), personalized recommendations, trending skills feed, or browsing user listings. Categories include Tech, Languages, Creative Arts, Professional Development, Hobbies, and more.
- **Sessions and Chat**: Schedule 1:1 exchanges with real-time chat via WebSockets. Basic exchanges are free; premium users get priority scheduling and unlimited chats.
- **Monetization Features**: Freemium—free basic access; premium subscriptions for advanced matching, certifications (e.g., badges for completed exchanges), and ad-free experience. Users pay for access to high-demand 1:1 chats with verified experts.
- **Admin Tools**: Moderation for profiles, dispute resolution, and analytics.

### Target Audience
- Learners seeking affordable, personalized skill-building (e.g., students, career changers).
- Experts monetizing niche skills in the gig economy.
- Focus on global users, with potential localization (e.g., for Brazil's growing edtech market).

## Technology Stack

SkillSphere is built with a modern, type-safe RPC architecture optimized for server-side rendering and minimal JavaScript.

- **Backend**: Go with **Connect RPC** (by Buf) for type-safe, protocol buffer-based APIs. Connect provides gRPC, gRPC-Web, and Connect protocol support with better browser compatibility than pure gRPC. Authentication via JWT/OAuth (e.g., integrate with Google/Auth0) using Connect interceptors. Database: PostgreSQL with GORM ORM for data persistence; add TimescaleDB for time-series data (e.g., session logs) and pgvector for vector storage in AI-based matching (e.g., skill embeddings). (Alternative: Start with SQLite via Turso for MVP simplicity if not needing vectors immediately.)
- **API Layer**: Protocol Buffers (proto3) define services and messages. Buf CLI manages protobuf generation and breaking change detection. Connect handlers serve both RPC clients and traditional HTTP endpoints.
- **Frontend**: Templ for server-rendered HTML templates, HTMX for asynchronous updates (e.g., dynamic search results without page reloads via Connect RPC calls), and Alpine.js for lightweight client-side interactivity (e.g., modals, toggles). Connect's HTTP/JSON support integrates seamlessly with HTMX for partial page updates.
- **Real-Time Features**: Gorilla WebSocket for chat and session notifications (Connect RPC supports streaming for real-time updates as an alternative).
- **AI Integration**: Optional but recommended for advanced matching—use Google Gemini SDK (Go client) to generate skill embeddings for semantic similarity. This enhances discovery by handling synonyms and related skills.
- **Deployment/Cloud**: Fly.io for easy, global deployment (scales well with Go's efficiency, low-cost tiers). Alternatives: Hetzner for budget VPS (if self-managed) or Google Cloud Platform (GCP) for seamless Gemini integration and managed Postgres.
- **Other Tools**: Stripe for payments, Prometheus for metrics, Docker for containerization, and Buf for protobuf workflow.

### Why This Stack?
- **Connect RPC Benefits**: Type-safe APIs across client/server, automatic code generation, built-in middleware (interceptors), supports gRPC and browser-friendly protocols, excellent for microservices evolution.
- **Pros**: Fast development (2-4 weeks MVP), low overhead, scalable. Go handles concurrency for real-time features; server-side focus reduces JS complexity. Connect's protobuf contracts prevent API drift.
- **Cons**: For complex UIs (e.g., drag-and-drop), add more JS sparingly. Start as a Progressive Web App (PWA) for mobile; native apps later can reuse protobuf definitions.
- **AI Decision**: Yes, for better matching—Gemini provides cost-effective embeddings without heavy ML training. Skip for MVP if focusing on basic algorithms.

### Connect RPC Architecture
SkillSphere uses Connect RPC to define services in `.proto` files, generating type-safe Go server handlers and TypeScript/JavaScript clients (optional for future native apps). Key services:

- **UserService**: CreateUser, GetUser, UpdateProfile, ListUsers
- **MatchingService**: FindMatches, GetRecommendations (algorithms run server-side)
- **SessionService**: CreateSession, GetSessions, UpdateStatus
- **ChatService**: SendMessage (with streaming support via Connect)
- **PaymentService**: CreateSubscription, ProcessPayment (Stripe integration)

Connect interceptors handle auth (JWT validation), logging, rate limiting, and error mapping. The frontend calls Connect endpoints via fetch/HTMX (JSON over HTTP) or gRPC-Web for streaming.

## Installation and Setup

### Prerequisites
- Go 1.21+
- PostgreSQL 15+ (with pgvector extension installed via `CREATE EXTENSION vector;`)
- Buf CLI (`brew install bufbuild/buf/buf` or see [buf.build](https://buf.build/docs/installation))
- Node.js 18+ (optional, for generating TS clients)

### Steps
1. Clone the repo:
   ```bash
   git clone https://github.com/yourusername/skillsphere.git
   cd skillsphere
   ```

2. Install Go dependencies:
   ```bash
   go mod tidy
   ```

3. Install Buf and generate code from protobuf:
   ```bash
   # Install Buf plugins for Go
   go install github.com/bufbuild/buf/cmd/buf@latest
   go install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go@latest
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

   # Generate Go code from proto files
   buf generate
   ```
   This generates type-safe server stubs and client code in `gen/` directory.

4. Set up environment variables (`.env` file):
   ```env
   DATABASE_URL=postgres://user:pass@localhost:5432/skillsphere
   JWT_SECRET=your-secret-key
   GEMINI_API_KEY=your-gemini-key
   STRIPE_KEY=sk_test_...
   SERVER_PORT=8080
   ```

5. Initialize database:
   ```bash
   go run cmd/migrate/main.go  # Run GORM migrations
   ```

6. Run the server:
   ```bash
   air  # For hot-reloading (install via go install github.com/air-verse/air@latest)
   # Or: go run cmd/server/main.go
   ```
   Access at `http://localhost:8080`. Connect RPC services available at `/connect/*` paths.

7. Deploy to Fly.io:
   ```bash
   # Install Fly CLI
   curl -L https://fly.io/install.sh | sh

   # Launch app (follow prompts for Postgres add-on)
   fly launch

   # Add secrets
   fly secrets set DATABASE_URL=... JWT_SECRET=... GEMINI_API_KEY=... STRIPE_KEY=...

   # Deploy
   fly deploy
   ```

### Project Structure
```
skillsphere/
├── api/                    # Proto definitions
│   └── skillsphere/
│       └── v1/
│           ├── user.proto
│           ├── matching.proto
│           ├── session.proto
│           └── chat.proto
├── buf.gen.yaml           # Buf codegen config
├── buf.yaml               # Buf workspace config
├── cmd/
│   ├── server/            # Main server entry point
│   └── migrate/           # Database migrations
├── gen/                   # Generated code (gitignored)
│   └── skillsphere/
│       └── v1/
├── internal/
│   ├── server/            # Connect RPC handlers
│   ├── service/           # Business logic
│   ├── matching/          # Matching algorithms
│   ├── db/                # Database models (GORM)
│   └── middleware/        # Connect interceptors
├── web/
│   ├── templates/         # Templ files
│   └── static/            # CSS, JS (Alpine, HTMX)
├── go.mod
└── README.md
```

## Usage

### API Interaction
Connect RPC services are accessible via:
- **Connect Protocol**: `POST /connect/skillsphere.v1.UserService/GetUser` (JSON body)
- **gRPC**: Standard gRPC clients (port 8080)
- **gRPC-Web**: Browser-compatible gRPC

Example with curl (Connect protocol):
```bash
# Create user
curl -X POST http://localhost:8080/connect/skillsphere.v1.UserService/CreateUser \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","name":"Alice"}'

# Find matches
curl -X POST http://localhost:8080/connect/skillsphere.v1.MatchingService/FindMatches \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt-token>" \
  -d '{"user_id":"123","wanted_skills":["python","ml"]}'
```

### Frontend Integration with HTMX
HTMX can call Connect endpoints directly (they're HTTP/JSON):
```html
<!-- In Templ template -->
<button hx-post="/connect/skillsphere.v1.MatchingService/FindMatches"
        hx-vals='{"user_id":"123","wanted_skills":["go","rust"]}'
        hx-target="#results"
        hx-swap="innerHTML">
  Find Matches
</button>
<div id="results"></div>
```
Server returns HTML fragments (generated via Templ) from Connect handlers.

### User Flow
- **Sign Up/Login**: Via email/OAuth. JWT tokens issued via UserService.
- **Create Profile**: Add skills (e.g., "Go Programming - Level 8" offered, "Spanish - Level 3" wanted) via UpdateProfile RPC.
- **Search & Match**: Call FindMatches RPC; get sorted recommendations via algorithms.
- **Exchange**: Schedule sessions via SessionService; chat in real-time (WebSocket or Connect streaming).
- **Premium**: Subscribe via PaymentService (Stripe integration) for perks like exclusive 1:1 access.

For code examples of matching algorithms, see `internal/matching/` directory (e.g., cosine similarity in Go using gonum).

## Connect RPC Example

### Proto Definition (api/skillsphere/v1/matching.proto)
```protobuf
syntax = "proto3";

package skillsphere.v1;

option go_package = "github.com/yourusername/skillsphere/gen/skillsphere/v1;skillsphere";

message Skill {
  string name = 1;
  int32 proficiency = 2; // 1-10 scale
}

message FindMatchesRequest {
  string user_id = 1;
  repeated Skill wanted_skills = 2;
  int32 limit = 3;
}

message UserMatch {
  string user_id = 1;
  string name = 2;
  repeated Skill offered_skills = 3;
  double similarity_score = 4;
}

message FindMatchesResponse {
  repeated UserMatch matches = 1;
}

service MatchingService {
  rpc FindMatches(FindMatchesRequest) returns (FindMatchesResponse) {}
}
```

### Server Handler (internal/server/matching.go)
```go
package server

import (
    "context"
    "connectrpc.com/connect"
    matchingv1 "github.com/yourusername/skillsphere/gen/skillsphere/v1"
    "github.com/yourusername/skillsphere/internal/service"
)

type MatchingServer struct {
    svc *service.MatchingService
}

func NewMatchingServer(svc *service.MatchingService) *MatchingServer {
    return &MatchingServer{svc: svc}
}

func (s *MatchingServer) FindMatches(
    ctx context.Context,
    req *connect.Request[matchingv1.FindMatchesRequest],
) (*connect.Response[matchingv1.FindMatchesResponse], error) {
    // Business logic in service layer
    matches, err := s.svc.FindMatches(ctx, req.Msg.UserId, req.Msg.WantedSkills, req.Msg.Limit)
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }

    // Convert to proto response
    resp := &matchingv1.FindMatchesResponse{
        Matches: matches,
    }
    return connect.NewResponse(resp), nil
}
```

### Main Server Setup (cmd/server/main.go)
```go
package main

import (
    "log"
    "net/http"
    "golang.org/x/net/http2"
    "golang.org/x/net/http2/h2c"
    "connectrpc.com/connect"
    "github.com/yourusername/skillsphere/gen/skillsphere/v1/skillspherev1connect"
    "github.com/yourusername/skillsphere/internal/server"
    "github.com/yourusername/skillsphere/internal/service"
)

func main() {
    // Initialize services
    matchingSvc := service.NewMatchingService()
    matchingServer := server.NewMatchingServer(matchingSvc)

    // Create Connect handlers
    mux := http.NewServeMux()
    path, handler := skillspherev1connect.NewMatchingServiceHandler(matchingServer)
    mux.Handle(path, handler)

    // Add CORS and auth interceptors
    interceptors := connect.WithInterceptors(
        NewAuthInterceptor(),
        NewLoggingInterceptor(),
    )

    // Serve with HTTP/2 support (for gRPC compatibility)
    addr := ":8080"
    log.Printf("Server listening on %s", addr)
    log.Fatal(http.ListenAndServe(addr, h2c.NewHandler(mux, &http2.Server{})))
}
```

## Roadmap
- **MVP (Weeks 1-4)**: Profiles (UserService), basic matching (cosine similarity via MatchingService), chat (WebSocket/ChatService), protobuf API definitions.
- **V1 (Months 2-3)**: AI integration (Gemini embeddings), payments (PaymentService with Stripe), auth interceptors.
- **V2 (Months 4-6)**: Mobile app (reuse proto definitions for native clients), certifications via blockchain badges, streaming chat via Connect.
- **Future**: Microservices split (separate matching service), GraphQL gateway over Connect, analytics dashboard.

For questions, open an issue or contact [your.email@example.com].

## License

MIT License. See [LICENSE](LICENSE) for details.

---

## SkillSphere Business Plan

### Executive Summary
SkillSphere is a P2P skill exchange platform launching as an MVP in 2-4 weeks, targeting the intersection of the gig economy and online learning markets. With a freemium model, it connects users for skill trades (e.g., tech, languages) using advanced matching algorithms. Built on modern Connect RPC architecture for type-safe, scalable APIs. Projected revenue from subscriptions, premium chats, and ads/partnerships.

Market opportunity: The global gig economy is valued at ~$582B in 2025, online learning at ~$353B, and sharing economy at ~$246B. SkillSphere differentiates with real-time P2P focus, AI-driven discovery, type-safe RPC APIs, and low-barrier entry.

Goal: Achieve 10K users in Year 1, $500K revenue by Year 2. Solo-founder viable, with scalable tech.

### Market Analysis
- **Size & Growth**: Gig economy: $582.2B by 2025, with 70M+ US workers (36% of workforce). Online learning: $320-353B in 2025, growing 12-14% CAGR to $842B by 2030. P2P/sharing subsets: $194B in 2024 to $246B in 2025.
- **Trends**: Rise in remote work, skill gaps (e.g., AI/tech demand). In emerging markets like Brazil, edtech booms (~$3B by 2025). Users seek personalized, affordable alternatives to courses.
- **Target Segments**: 18-45 year-olds in tech, education, creative fields. Initial focus: Global, with SEO for niches like "P2P coding exchange."
- **Competitive Analysis**: Direct competitors include Preply (language tutoring), Skillshare (class-based, not pure P2P), and alternatives like Udemy, Coursera, LinkedIn Learning, MasterClass. P2P platforms: Teachable/Kajabi for creators, but less exchange-focused. Differentiation: Free basic P2P, AI matching, real-time chat, type-safe APIs for mobile/web. Weaknesses: Established players have scale; SkillSphere starts lean.

### Product Description
- **Core Offering**: P2P exchanges of any skills (e.g., programming, foreign languages, music, cooking, business). Profiles include bio, skill lists with proficiencies, ratings.
- **User Flow**: Sign up → Build profile → Search/discover (algorithms recommend based on wanted/offered skills) → Match & chat (pay for premium 1:1 with high-demand users) → Exchange session → Rate/certify.
- **MVP Scope**: Profiles (UserService), basic matching (cosine similarity via MatchingService), chat (ChatService), search. Expand to AI (Gemini for embeddings), categories.
- **Tech & Operations**: Connect RPC for type-safe APIs; HTMX/Templ for frontend; Go backend. Development: Solo, 2-4 weeks MVP. Hosting: Fly.io (~$5-20/month initially). Support: Community forums, email.

### Monetization Strategy
- **Freemium Model**: Basic free (limited matches/chats). Premium: $5-20/month for unlimited chats, priority matching, certifications.
- **Additional Revenue**:
    - Pay-per-chat: Users pay $1-10 for 1:1 access to premium profiles (e.g., verified experts).
    - Ads/Partnerships: Targeted ads from edtech firms; affiliate links (e.g., tools for skills).
    - Certifications: $10-50 for badges post-exchange.
    - API Access: Sell access to SkillSphere Connect APIs for third-party integrations.
- **Projections**: Year 1: 10K users, 10% premium conversion → $60K revenue. Scale to $500K by Year 2 via marketing.

### Marketing & Growth Strategy
- **Acquisition**: SEO (e.g., "free skill exchange app"), social media (Reddit, LinkedIn, X), content marketing (blogs on skill-building). Partnerships with influencers in niches.
- **Retention**: Email newsletters with recommendations, gamification (badges, streaks).
- **Metrics**: Track sign-ups, retention, match success rate. Budget: $1K/month initial (ads/SEO tools).
- **Launch Plan**: Beta via landing page (built with Templ), invite-only, then public.

### Operations & Team
- **Founder-Led**: Solo initially; outsource design if needed.
- **Risks**: Data privacy (GDPR compliance), moderation (AI flags). Mitigation: JWT auth via interceptors, user reports.
- **Legal**: Incorporate as LLC; terms for exchanges (no liability).

### Financial Projections
- **Startup Costs**: $1-5K (domain, hosting, API keys, Buf licenses if needed).
- **Revenue Model**: Subscriptions (70%), pay-per-chat (20%), ads/API (10%).
- **Break-Even**: Month 6 at 1K premium users.
- **3-Year Forecast**:
    - Year 1: Revenue $100K, Expenses $50K (Profit $50K).
    - Year 2: $500K revenue.
    - Year 3: $2M (with 100K users).
    - Assumptions: 20% MoM growth, low churn.

This plan is adaptable—validate MVP feedback for pivots. For refinements, consult advisors.

---

## Overview of Skill Matching Algorithms

Skill matching algorithms are computational methods used to pair individuals, jobs, or resources based on skills, proficiencies, or requirements. In SkillSphere's Connect RPC architecture, matching algorithms run server-side in the MatchingService, exposing results via type-safe protobuf APIs. They range from simple rule-based systems to advanced AI-driven ones, balancing factors like accuracy, scalability, and computational cost.

Key considerations in design:
- **Input Representation**: Skills as vectors (e.g., proficiency levels from 1-10) or embeddings (semantic vectors from NLP models).
- **Scoring**: Measure similarity or distance between profiles.
- **Additional Factors**: Incorporate location, availability, or user ratings for refined matches.
- **Challenges**: Handling synonyms (e.g., "ML" vs. "machine learning"), sparse data (missing skills), and scalability for large user bases.

Below, I'll detail common algorithms, with examples of how they work and implementations in Go (using libraries like gonum for vector math) within Connect RPC services.

### 1. Distance-Based Algorithms (e.g., Euclidean/Cartesian Distance)
These treat skills as points in multi-dimensional space, where each skill is a dimension and proficiency is a coordinate. The "closeness" of two profiles is the geometric distance—smaller distances indicate better matches.

- **How It Works**:
    - Represent user profiles as vectors: e.g., User A: [Java: 8, SQL: 5, Python: 0] → Vector [8, 5, 0].
    - Compute distance to a query vector (e.g., [7, 6, 3]).
    - Formula: Euclidean Distance = √Σ (user_i - query_i)². For efficiency, use squared distance (skip the square root).
    - Missing skills default to 0.
    - Sort results by ascending distance for top matches.

- **Pros**: Simple, fast for databases; no training data needed.
- **Cons**: Doesn't handle semantic similarities (e.g., "Java" and "C#" as related).
- **Implementation Example**: In Go within MatchingService, compute distances after fetching profiles via GORM. Use gonum/floats for vector operations.
  ```go
  package matching

  import (
      "math"
      "gonum.org/v1/gonum/floats"
  )

  // UserProfile represents a skill vector (e.g., [Java, SQL, Python] proficiencies)
  type UserProfile struct {
      ID     string
      Vector []float64
  }

  func euclideanDistance(a, b []float64) float64 {
      if len(a) != len(b) {
          panic("Vectors must be same length")
      }
      diff := make([]float64, len(a))
      floats.SubTo(diff, a, b) // diff = a - b
      floats.Mul(diff, diff)   // square each element
      return math.Sqrt(floats.Sum(diff))
  }

  func FindMatchesByDistance(query []float64, users []UserProfile, limit int) []UserProfile {
      type match struct {
          user UserProfile
          dist float64
      }
      matches := make([]match, 0, len(users))

      for _, u := range users {
          dist := euclideanDistance(u.Vector, query)
          matches = append(matches, match{user: u, dist: dist})
      }

      // Sort by distance and return top N
      sort.Slice(matches, func(i, j int) bool {
          return matches[i].dist < matches[j].dist
      })

      result := make([]UserProfile, 0, limit)
      for i := 0; i < limit && i < len(matches); i++ {
          result = append(result, matches[i].user)
      }
      return result
  }
  ```
- **Use in SkillSphere**: Ideal for MVP with fixed skill lists; integrate into MatchingService RPC handler. Fetch users from DB via GORM, compute distances, return as FindMatchesResponse.

### 2. Similarity-Based Algorithms (e.g., Cosine Similarity)
These measure the angle between vectors, focusing on direction rather than magnitude—useful for sparse or varying-length profiles.

- **How It Works**:
    - Vectors as above, but normalize for length.
    - Formula: Cosine Similarity = (A · B) / (||A|| ||B||), where · is dot product (Σ A_i * B_i), and || || is magnitude (√Σ X_i²).
    - Scores range from -1 (opposite) to 1 (identical); threshold e.g., >0.7 for matches.
    - Handles zeros well (e.g., irrelevant skills don't penalize).

- **Pros**: Robust to scale differences; integrates with search engines like Elasticsearch.
- **Cons**: Ignores absolute proficiency levels if not weighted.
- **Implementation Example**: In Go with gonum/floats for dot product and norms.
  ```go
  package matching

  import (
      "gonum.org/v1/gonum/floats"
  )

  func cosineSimilarity(a, b []float64) float64 {
      if len(a) != len(b) {
          panic("Vectors must be same length")
      }
      dot := floats.Dot(a, b)
      normA := floats.Norm(a, 2) // L2 norm
      normB := floats.Norm(b, 2)
      if normA == 0 || normB == 0 {
          return 0
      }
      return dot / (normA * normB)
  }

  func FindMatchesBySimilarity(query []float64, users []UserProfile, threshold float64, limit int) []UserProfile {
      type match struct {
          user UserProfile
          sim  float64
      }
      matches := make([]match, 0, len(users))

      for _, u := range users {
          sim := cosineSimilarity(u.Vector, query)
          if sim >= threshold {
              matches = append(matches, match{user: u, sim: sim})
          }
      }

      // Sort by similarity (descending) and return top N
      sort.Slice(matches, func(i, j int) bool {
          return matches[i].sim > matches[j].sim
      })

      result := make([]UserProfile, 0, limit)
      for i := 0; i < limit && i < len(matches); i++ {
          result = append(result, matches[i].user)
      }
      return result
  }
  ```
- **Use in SkillSphere**: For fuzzy matching; combine with full-text search for skill keywords. In Connect RPC MatchingService, add to FindMatches handler with configurable threshold.

### 3. Machine Learning/Embedding-Based Algorithms (e.g., Word2Vec, Gemini Embeddings)
These use NLP to create dense vector representations (embeddings) of skills, capturing semantic relationships (e.g., "Python" close to "programming").

- **How It Works** (Using Gemini API):
    - **Step 1: Skill Extraction**: Parse skill names from user profiles.
    - **Step 2: Generate Embeddings**: Call Gemini API to get 768D (or similar) vectors per skill.
    - **Step 3: Profile Aggregation**: Average embeddings of all skills in a profile to create a profile vector.
    - **Step 4: Matching**: Compute cosine similarity between query profile embedding and candidate embeddings.
    - **Step 5: Threshold & Rank**: Return matches above similarity threshold (e.g., >0.62), sorted by score.

- **Pros**: Semantic understanding; handles variations like abbreviations, synonyms ("ML" = "machine learning").
- **Cons**: Requires API calls (latency, cost); cache embeddings in DB for performance.
- **Implementation Example**: Go with Gemini SDK for embeddings, gonum for similarity.
  ```go
  package matching

  import (
      "context"
      "gonum.org/v1/gonum/floats"
      "github.com/google/generative-ai-go/genai"
      "google.golang.org/api/option"
  )

  type EmbeddingMatcher struct {
      client *genai.Client
      cache  map[string][]float64 // Cache skill embeddings
  }

  func NewEmbeddingMatcher(apiKey string) (*EmbeddingMatcher, error) {
      ctx := context.Background()
      client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
      if err != nil {
          return nil, err
      }
      return &EmbeddingMatcher{
          client: client,
          cache:  make(map[string][]float64),
      }, nil
  }

  func (m *EmbeddingMatcher) GetEmbedding(ctx context.Context, skill string) ([]float64, error) {
      // Check cache first
      if emb, ok := m.cache[skill]; ok {
          return emb, nil
      }

      // Generate embedding via Gemini
      model := m.client.EmbeddingModel("text-embedding-004")
      result, err := model.EmbedContent(ctx, genai.Text(skill))
      if err != nil {
          return nil, err
      }

      // Store in cache
      m.cache[skill] = result.Embedding.Values
      return result.Embedding.Values, nil
  }

  func (m *EmbeddingMatcher) GetProfileEmbedding(ctx context.Context, skills []string) ([]float64, error) {
      embeddings := make([][]float64, 0, len(skills))

      for _, skill := range skills {
          emb, err := m.GetEmbedding(ctx, skill)
          if err != nil {
              return nil, err
          }
          embeddings = append(embeddings, emb)
      }

      // Average embeddings to get profile vector
      if len(embeddings) == 0 {
          return nil, fmt.Errorf("no skills provided")
      }

      dim := len(embeddings[0])
      avg := make([]float64, dim)
      for _, emb := range embeddings {
          floats.Add(avg, emb)
      }
      floats.Scale(1.0/float64(len(embeddings)), avg)

      return avg, nil
  }

  func (m *EmbeddingMatcher) FindSemanticMatches(
      ctx context.Context,
      querySkills []string,
      users []UserProfile, // Assume profiles have pre-computed embeddings
      threshold float64,
      limit int,
  ) ([]UserProfile, error) {
      queryEmb, err := m.GetProfileEmbedding(ctx, querySkills)
      if err != nil {
          return nil, err
      }

      type match struct {
          user UserProfile
          sim  float64
      }
      matches := make([]match, 0, len(users))

      for _, u := range users {
          sim := cosineSimilarity(u.Vector, queryEmb) // Vector is pre-computed embedding
          if sim >= threshold {
              matches = append(matches, match{user: u, sim: sim})
          }
      }

      sort.Slice(matches, func(i, j int) bool {
          return matches[i].sim > matches[j].sim
      })

      result := make([]UserProfile, 0, limit)
      for i := 0; i < limit && i < len(matches); i++ {
          result = append(result, matches[i].user)
      }
      return result, nil
  }
  ```
- **Use in SkillSphere**: Integrate into MatchingService RPC handler. Pre-compute and cache embeddings in PostgreSQL (use pgvector for efficient similarity queries). Call via FindMatches RPC; returns semantic matches in FindMatchesResponse.

### 4. Other Advanced Algorithms
- **Ontology-Based**: Use knowledge graphs (e.g., skill hierarchies like "programming > Python") for semantic matching. Metric: Graph distance or custom similarity. Store in Neo4j or PostgreSQL with pg_graph extension. Query via Connect RPC GraphService.
- **Clustering (e.g., k-Means)**: Group users into skill clusters, then match queries to nearest clusters. Useful for discovery at scale. Implement with gonum/cluster.
- **Rule-Based/Hybrid**: Simple thresholds (e.g., match if >70% skills overlap) combined with ML for ties. Implement as a Connect interceptor that applies business rules before calling embedding matcher.
- **AI-Enhanced with LLMs**: Use Gemini to build dynamic skill ontologies, improving accuracy over time. Background job updates ontology weekly.

### Recommendations for SkillSphere
1. **MVP**: Start with **Cosine Similarity** on vectorized profiles—it's balanced for your Go backend (use gonum). Define MatchingService in proto with FindMatches RPC.
2. **V1**: Add **Gemini Embeddings** for semantic matching. Cache embeddings in PostgreSQL with pgvector extension; query via SQL for performance.
3. **Scaling**: Use Connect RPC streaming for real-time match updates. Split MatchingService into separate microservice if needed (reuse proto definitions).
4. **Testing**: Create synthetic data (100-1000 users); measure precision/recall. Use Connect's built-in testing tools for RPC handlers.

Install gonum: `go get gonum.org/v1/gonum`
Install Connect: `go get connectrpc.com/connect`

---

## Extending SkillSphere

To grow beyond MVP, leverage 2025 edtech/gig trends like AI personalization, microlearning, blockchain creds, and hybrid work demands.

Here are targeted extensions, prioritized for monetization and feasibility:

### 1. AI-Enhanced Matching and Coaching
- Integrate LLM coaches (e.g., Gemini bots for session prep/tips via ChatService) or AI-driven microlearning paths (e.g., generate personalized skill roadmaps).
- Monetize as premium: $5/month for "AI Mentor" access.
- Add ontology-based matching: Build skill graphs (e.g., "Python" → "Data Science") using knowledge bases; compute graph distances for recommendations via new GraphService RPC.
- Implement via Connect RPC: Define CoachService with StreamAdvice RPC for real-time chat with AI.

### 2. Blockchain Certifications and Badges
- Issue verifiable badges/NFTs for completed exchanges via Polygon/Ethereum (low-fee chains).
- Users link to LinkedIn/resumes. Charge $10-50 per cert; partner with credential platforms.
- Define CertificationService RPC: IssueBadge, VerifyBadge methods.

### 3. Gig Economy Integrations
- **Freelance Marketplace**: Allow paid P2P gigs (e.g., "Teach Python for $20/hr") with escrow via Stripe.
- Add job boards for skill-based freelance (e.g., integrate with Upwork APIs).
- Define GigService RPC: CreateGig, AcceptGig, CompleteGig.
- **On-the-Move/AR Features**: Mobile app (reuse proto definitions for native clients) with AR overlays for in-person exchanges (e.g., scan QR for skill badges).

### 4. Community and Social Extensions
- **Group Sessions**: Scale 1:1 to 1:many workshops; monetize tickets. Add GroupSessionService RPC with streaming for live events.
- **Trending/Global Challenges**: Weekly skill challenges (e.g., "Learn AI basics") with leaderboards; sponsor with partners.
- Define ChallengeService RPC: GetTrendingChallenges, JoinChallenge.

### 5. B2B Enterprise Features
- Enterprise mode for companies (e.g., internal skill-sharing); white-label for schools.
- Multi-tenancy via Connect interceptors (tenant ID in request headers).
- Define OrganizationService RPC: CreateOrg, ManageUsers, GenerateReports.

### 6. Tech/Infra Extensions
- **Microservices**: Split matching into a separate Connect service for scalability. Use Buf to manage proto dependencies across services.
- **Mobile/PWA**: Enhance PWA with offline matching (cache embeddings in IndexedDB); add native push via Firebase. Generate TypeScript Connect clients for React Native.
- **Analytics**: User heatmaps for trending skills; sell insights to edtech firms. Define AnalyticsService RPC: GetSkillTrends, GetUserMetrics.
- **API Marketplace**: Sell access to SkillSphere Connect APIs for third-party integrations (e.g., LMS platforms). Use API keys via interceptors.

These build on your freemium model, targeting $500K+ revenue by Year 2 via 20% MoM growth. Validate with beta users for pivots. Connect RPC's type-safe contracts make it easy to evolve APIs without breaking clients.

---

## Connect RPC Best Practices for SkillSphere

1. **Versioning**: Use semantic proto packages (e.g., `skillsphere.v1`, `skillsphere.v2`). Buf checks for breaking changes via `buf breaking`.
2. **Interceptors**: Chain auth (JWT validation), logging, rate limiting, and error mapping in Connect middleware.
3. **Error Handling**: Return Connect errors with appropriate codes (e.g., `connect.CodeInvalidArgument`, `connect.CodeUnauthenticated`).
4. **Streaming**: Use streaming RPCs for real-time features (e.g., ChatService.StreamMessages) instead of WebSockets where possible.
5. **Testing**: Write unit tests for handlers using `connecttest` package. Mock dependencies (e.g., GORM DB, Gemini client).
6. **Documentation**: Generate docs from proto files using `buf generate` with protoc-gen-doc plugin. Publish to Buf Schema Registry for client discovery.
7. **Observability**: Add Prometheus metrics via interceptors. Track RPC call counts, latency, error rates per service/method.

For more Connect RPC examples and SkillSphere-specific implementations, see the `internal/server/` and `internal/service/` directories.