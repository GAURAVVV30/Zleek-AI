// Package bootstrap seeds the platform with its authoritative baseline data,
// all derived from the roadmap.sh-derived domain graphs embedded in aiengine:
// base users (admin/curator), the 12 career role domains, their published
// knowledge structures and concepts, and the curator-approved resources. Every
// row uses a deterministic UUID so re-running is idempotent. Nothing here
// invents roadmap data — it mirrors the graph payloads 1:1.
package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/hcl-backend/services/api-go/internal/aiengine"
	"github.com/hcl-backend/services/api-go/internal/platform/keys"
)

const (
	adminEmail   = "admin@hclearn.app"
	curatorEmail = "curator@hclearn.app"
	demoEmail    = "learner.demo@gmail.com"
)

type DB interface {
	Pool() *pgxpool.Pool
}

func Seed(ctx context.Context, db *pgxpool.Pool, graph *aiengine.GraphEngine, log *slog.Logger) error {
	if err := seedUsers(ctx, db); err != nil {
		return fmt.Errorf("seed users: %w", err)
	}
	if err := seedKnowledge(ctx, db, graph); err != nil {
		return fmt.Errorf("seed knowledge: %w", err)
	}
	if err := seedResources(ctx, db, graph); err != nil {
		return fmt.Errorf("seed resources: %w", err)
	}
	log.Info("bootstrap seed complete")
	return nil
}

// SeedAdminUser returns the seeded admin id without re-seeding (used by main
// for the audit actor reference).
func SeedAdminID() string { return keys.User("admin") }

func seedUsers(ctx context.Context, db *pgxpool.Pool) error {
	adminPassword := "Admin_ChangeMe_2026"
	if v := os.Getenv("ADMIN_PASSWORD"); v != "" {
		adminPassword = v
	}

	users := []struct {
		key   string
		email string
		full  string
		role  string
	}{
		{"user:admin", adminEmail, "Platform Admin", "admin"},
		{"user:curator", curatorEmail, "Content Curator", "curator"},
		{"user:demo", demoEmail, "Demo Learner", "learner"},
	}
	for _, u := range users {
		pwHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		query := `
			INSERT INTO platform.users (id, email, password_hash, role, status, full_name, timezone, theme, created_at, updated_at)
			VALUES ($1, $2, $3, $4, 'active', $5, 'UTC', 'default', $6, $6)
			ON CONFLICT (id) DO NOTHING`
		if _, err := db.Exec(ctx, query, keys.UUID(u.key), u.email, string(pwHash), u.role, u.full, now); err != nil {
			return err
		}
	}
	// Ensure a learner profile row exists for the demo learner (idempotent).
	query := `
		INSERT INTO platform.learner_profiles (user_id, time_availability, format_preference, prior_experience)
		VALUES ($1, 'gt_20', 'video,article', 'intermediate')
		ON CONFLICT (user_id) DO NOTHING`
	if _, err := db.Exec(ctx, query, keys.User("demo")); err != nil {
		return err
	}
	return nil
}

func seedKnowledge(ctx context.Context, db *pgxpool.Pool, graph *aiengine.GraphEngine) error {
	for _, domainID := range graph.DomainList {
		g := graph.DomainGraphs[domainID]
		now := time.Now().UTC()

		domainKey := keys.Domain(domainID)
		ksID := keys.KnowledgeStructure(domainID)
		adminID := SeedAdminID()

		if _, err := db.Exec(ctx, `
			INSERT INTO platform.domains (id, slug, name, description)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO NOTHING`,
			domainKey, domainID, g.DomainName, fmt.Sprintf("Become a %s.", g.DomainName)); err != nil {
			return err
		}

		if _, err := db.Exec(ctx, `
			INSERT INTO platform.knowledge_structures
				(id, domain_id, version, status, created_by, published_at, created_at)
			VALUES ($1, $2, 1, 'published', $3, $4, $4)
			ON CONFLICT (domain_id, version) DO NOTHING`,
			ksID, domainKey, adminID, now); err != nil {
			return err
		}

		for _, node := range g.Nodes {
			desc := strings.TrimSpace(fmt.Sprintf("%s — %s | difficulty %s", name(node), category(node), difficulty(node)))
			if _, err := db.Exec(ctx, `
				INSERT INTO platform.concepts (id, node_id, knowledge_structure_id, name, description, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (id) DO NOTHING`,
				keys.Concept(node.ID), node.ID, ksID, name(node), desc, now); err != nil {
				return err
			}
		}
	}

	// Prerequisite edges (hard) between concepts — a second pass so every
	// concept PK exists. concept_prerequisites references concepts.id (UUID),
	// so node ids are mapped through the deterministic keys (globally unique).
	for _, domainID := range graph.DomainList {
		for _, node := range graph.DomainGraphs[domainID].Nodes {
			if prereqs, ok := nodePrereqs(node); ok {
				for _, dep := range prereqs {
					if _, err := db.Exec(ctx, `
						INSERT INTO platform.concept_prerequisites (concept_id, prerequisite_concept_id)
						VALUES ($1, $2)
						ON CONFLICT DO NOTHING`,
						keys.Concept(node.ID), keys.Concept(dep)); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func seedResources(ctx context.Context, db *pgxpool.Pool, graph *aiengine.GraphEngine) error {
	now := time.Now().UTC()
	adminID := SeedAdminID()
	seen := map[string]bool{}

	// resources_rag.json extended corpora (domain level) are merged in for the
	// two domains that ship them.
	ragByDomain := loadRagResources(graph)

	for _, domainID := range graph.DomainList {
		g := graph.DomainGraphs[domainID]

		insertResource := func(url, title, provider, resType, source string, authority float64, conceptKey string, seconds int) error {
			if url == "" || seen[url] {
				return nil
			}
			seen[url] = true
			resID := keys.Resource(url)
			_, err := db.Exec(ctx, `
				INSERT INTO platform.resources
					(id, url, source, author, resource_type, title, duration_minutes, authority_score, provenance_note,
					 status, freshness_status, curated_by, curated_at, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'published', 'fresh', $10, $11, $11)
				ON CONFLICT (id) DO NOTHING`,
				resID, url, source, provider, resType, title, seconds, authority,
				"roadmap.sh-curated node resource", adminID, now)
			if err != nil {
				return err
			}
			if _, err := db.Exec(ctx, `
				INSERT INTO platform.resource_concepts (resource_id, concept_id, relevance_note)
				VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING`,
				resID, keys.Concept(conceptKey), title); err != nil {
				return err
			}
			if _, err := db.Exec(ctx, `
				INSERT INTO platform.resource_quality_signals (resource_id, feedback_count, updated_at)
				VALUES ($1, 0, $2)
				ON CONFLICT (resource_id) DO NOTHING`,
				resID, now); err != nil {
				return err
			}
			return nil
		}

		for _, node := range g.Nodes {
			conceptID := node.ID
			resources, _ := node.Raw["resources"].([]any)
			for _, rr := range resources {
				rm, ok := rr.(map[string]any)
				if !ok {
					continue
				}
				url, _ := rm["url"].(string)
				title, _ := rm["title"].(string)
				provider, _ := rm["provider"].(string)
				typ, _ := rm["type"].(string)
				authority := floatAny(rm["authority_score"])
				if err := insertResource(url, title, provider, resourceType(typ), "roadmap.sh/"+domainID, authority, conceptID, resourceMinutes(node, resources)); err != nil {
					return err
				}
			}
		}

		// Domain-level RAG corpus entries.
		for _, rr := range ragByDomain[domainID] {
			url, _ := rr["url"].(string)
			title, _ := rr["title"].(string)
			typ, _ := rr["type"].(string)
			if url == "" {
				continue
			}
			// Attach to the node whose keywords match the tags best, else the
			// first node (keeps the resource reachable via resource_concepts).
			conceptKey := matchConcept(g, rr)
			if err := insertResource(url, title, "roadmap.sh", resourceType(typ), "roadmap.sh/"+domainID, 0.9, conceptKey, 20); err != nil {
				return err
			}
		}
	}
	return nil
}

func loadRagResources(graph *aiengine.GraphEngine) map[string][]map[string]any {
	out := map[string][]map[string]any{}
	for _, domainID := range graph.DomainList {
		g := graph.DomainGraphs[domainID]
		payload, err := loadRagPayload(g.SourceFile)
		if err != nil {
			continue
		}
		resources, _ := payload["resources"].([]any)
		for _, rr := range resources {
			if rm, ok := rr.(map[string]any); ok {
				out[domainID] = append(out[domainID], rm)
			}
		}
	}
	return out
}

func matchConcept(g *aiengine.DomainGraph, res map[string]any) string {
	tags := stringSlice(res["tags"])
	keywords := strings.ToLower(strings.Join(tags, " "))
	best := ""
	for _, node := range g.Nodes {
		core := strings.ToLower(strings.Join(stringSlice(node.Raw["core_concepts"]), " "))
		if keywords != "" && (strings.Contains(keywords, strings.ToLower(node.ID)) ||
			strings.Contains(core, keywords) || strings.Contains(keywords, core)) {
			best = node.ID
			break
		}
	}
	if best == "" && len(g.Nodes) > 0 {
		best = g.Nodes[0].ID
	}
	return best
}

func name(n *aiengine.GraphNode) string {
	if v := stringAny(n.Raw["name"]); v != "" {
		return v
	}
	return n.ID
}

func category(n *aiengine.GraphNode) string { return stringAny(n.Raw["category"]) }

func difficulty(n *aiengine.GraphNode) string {
	if v := stringAny(n.Raw["difficulty_level"]); v != "" {
		return v
	}
	return "zero"
}

func nodePrereqs(n *aiengine.GraphNode) ([]string, bool) {
	pre, ok := n.Raw["prerequisites"].(map[string]any)
	if !ok {
		return nil, false
	}
	hard, ok := pre["hard"].([]any)
	if !ok {
		return nil, false
	}
	out := []string{}
	for _, h := range hard {
		if s, ok := h.(string); ok {
			out = append(out, s)
		}
	}
	return out, true
}

func resourceMinutes(n *aiengine.GraphNode, resources []any) int {
	hours := intAny(n.Raw["estimated_hours"])
	if hours <= 0 {
		return 20
	}
	total := len(resources)
	if total <= 0 {
		total = 1
	}
	return max((hours*60)/total, 10)
}

func resourceType(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "video", "course", "tutorial":
		return "video"
	case "article", "doc", "documentation", "blog", "notes":
		return "article"
	case "interactive", "practice", "lab":
		return "interactive"
	case "book":
		return "article"
	default:
		return "article"
	}
}

// loadRagPayload reads the sibling resources_rag.json for a domain graph file.
func loadRagPayload(graphFilePath string) (map[string]any, error) {
	base := strings.TrimSuffix(graphFilePath, "_graph.json")
	dir := base[:strings.LastIndex(base, "/")+1]
	ragFile := dir + "resources_rag.json"
	data, err := os.ReadFile(ragFile)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func stringAny(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func intAny(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case json.Number:
		f, _ := t.Int64()
		return int(f)
	}
	return 0
}

func floatAny(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case json.Number:
		f, _ := t.Float64()
		return f
	}
	return 0
}

func stringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := []string{}
	for _, a := range arr {
		if s, ok := a.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
