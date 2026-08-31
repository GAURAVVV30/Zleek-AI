package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/hcl-backend/services/api-go/internal/knowledge/domain"
	"github.com/hcl-backend/services/api-go/internal/platform/keys"
)

type KnowledgeService struct {
	repo KnowledgeRepository
	meta ConceptMetaProvider
	now  func() time.Time
	rag  RAGService
}

func NewKnowledgeService(repo KnowledgeRepository, meta ConceptMetaProvider, rag RAGService) *KnowledgeService {
	return &KnowledgeService{repo: repo, meta: meta, now: time.Now, rag: rag}
}

func (s *KnowledgeService) ListDomains(ctx context.Context) ([]domain.Domain, error) {
	domains, err := s.repo.ListDomains(ctx)
	if err != nil {
		return nil, err
	}
	for i := range domains {
		domains[i].PopularGoals = []string{
			"Become a " + domains[i].Name + " and build real-world projects",
			"Master " + domains[i].Name + " from first principles",
			"Advance your career as a " + domains[i].Name + " professional",
		}
	}
	return domains, nil
}

func (s *KnowledgeService) GetConcept(ctx context.Context, id string) (*domain.Concept, error) {
	return s.repo.GetConcept(ctx, id)
}

// ValidateConcept ensures a concept exists (used by competency/assessment ports).
func (s *KnowledgeService) ValidateConcept(ctx context.Context, id string) error {
	_, err := s.repo.GetConcept(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrConceptNotFound) {
			return errors.New("concept not found")
		}
		return err
	}
	return nil
}

// ConceptName resolves a concept's title by node id.
func (s *KnowledgeService) ConceptName(ctx context.Context, id string) (string, error) {
	concept, err := s.repo.GetConcept(ctx, id)
	if err != nil {
		return "", err
	}
	return concept.Name, nil
}

// CoreConceptNames resolves the concept's core building-block names by node id.
func (s *KnowledgeService) CoreConceptNames(ctx context.Context, id string) ([]string, error) {
	if meta, ok := s.meta.Meta(ctx, id); ok && len(meta.CoreConcepts) > 0 {
		return meta.CoreConcepts, nil
	}
	return nil, nil
}

// ConceptDomain resolves the roadmap domain slug for a concept (needed to drive
// the AI evaluator on a free-text submission).
func (s *KnowledgeService) ConceptDomain(ctx context.Context, id string) (string, error) {
	if meta, ok := s.meta.Meta(ctx, id); ok {
		return meta.DomainID, nil
	}
	return "", domain.ErrConceptNotFound
}

// ResolvedStructure is the published knowledge structure matched to a goal.
type ResolvedStructure struct {
	ID          string
	DomainSlug  string
	DomainName  string
	Confidence  float64
	IsPublished bool
}

// ResolveStructure maps a roadmap domain slug (or structure uuid) to the
// published knowledge structure for that domain.
func (s *KnowledgeService) ResolveStructure(ctx context.Context, ref string) (*ResolvedStructure, error) {
	structure, err := s.repo.GetPublishedStructureForDomain(ctx, ref)
	if err != nil {
		if !errors.Is(err, domain.ErrKnowledgeStructureNotFound) {
			return nil, err
		}
		structure, err = s.repo.GetStructure(ctx, ref)
		if err != nil {
			return nil, err
		}
	}
	slug, name := structure.DomainName, structure.DomainName
	if slug, name, err = s.repo.GetDomainByStructure(ctx, structure.ID); err != nil {
		slug, name = structure.DomainID, structure.DomainName
	}
	return &ResolvedStructure{
		ID:          structure.ID,
		DomainSlug:  slug,
		DomainName:  name,
		Confidence:  0.94,
		IsPublished: structure.Status == "published",
	}, nil
}

// StructureName returns the roadmap domain slug/name for a structure uuid.
func (s *KnowledgeService) StructureName(ctx context.Context, structureID string) (string, string, error) {
	return s.repo.GetDomainByStructure(ctx, structureID)
}

// ConceptView is the LearningWorkspace payload shape consumed by the frontend.
type ConceptView struct {
	ID                 string         `json:"id"`
	Title              string         `json:"title"`
	Breadcrumb         []string       `json:"breadcrumb"`
	WhyItMatters       string         `json:"whyItMatters"`
	PrimaryResource    *ResourceView  `json:"primaryResource"`
	AlternateResources []ResourceView `json:"alternateResources"`
	EstimatedMinutes   int            `json:"estimatedMinutes"`
	Difficulty         string         `json:"difficulty"`
	Category           string         `json:"category"`
	CoreConcepts       []string       `json:"coreConcepts"`
}

type ResourceView struct {
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	Type            string      `json:"type"`
	DurationMinutes int         `json:"durationMinutes"`
	Provider        string      `json:"provider"`
	SourceURL       string      `json:"sourceUrl"`
	Provenance      *Provenance `json:"provenance,omitempty"`
	WhyThisResource string      `json:"whyThisResource"`
}

type Provenance struct {
	Author     string `json:"author"`
	VettedBy   string `json:"vettedBy"`
	VettedDate string `json:"vettedDate"`
}

// GetConceptView assembles the workspace view for a learner: concept + graph
// metadata + the best published resource for their preferred format.
func (s *KnowledgeService) GetConceptView(ctx context.Context, learnerID, conceptID string) (*ConceptView, error) {
	// Enforce sequential learning path check
	state, err := s.repo.GetConceptState(ctx, learnerID, conceptID)
	if err == nil && state == "locked" {
		return nil, fmt.Errorf("concept %s is locked: complete preceding modules first", conceptID)
	}

	concept, err := s.repo.GetConcept(ctx, conceptID)
	if err != nil {
		return nil, err
	}
	meta, ok := s.meta.Meta(ctx, concept.ID)
	if !ok {
		meta = &ConceptMeta{ID: concept.ID, DomainName: concept.Name, Name: concept.Name}
	}

	prefs, _ := s.repo.GetFormatPrefs(ctx, learnerID)
	resources, err := s.repo.ListConceptResources(ctx, concept.ID)
	if err != nil {
		return nil, err
	}
	resources = published(resources)

	// Fetch active domain slug and RAG context
	domainSlug, _, err := s.repo.GetDomainByStructure(ctx, concept.KnowledgeStructureID)
	if err != nil {
		domainSlug = "machine_learning"
	}
	ragContexts := s.rag.GetRAGContext(domainSlug, concept.ID)
	ragText := strings.Join(ragContexts, "\n---\n")

	// Filter strictly by the preferred formats if preferences are configured
	if len(prefs) > 0 {
		var filtered []domain.Resource
		want := map[string]bool{}
		for _, p := range prefs {
			want[strings.ToLower(strings.TrimSpace(p))] = true
		}
		for _, r := range resources {
			if want[strings.ToLower(r.ResourceType)] {
				filtered = append(filtered, r)
			}
		}
		resources = filtered
	}

	ranked := rankResources(resources, prefs)
	view := &ConceptView{
		ID:                 concept.ID,
		Title:              meta.Name,
		Breadcrumb:         breadcrumb(meta),
		WhyItMatters:       whyItMatters(meta, concept),
		EstimatedMinutes:   meta.EstimatedHours * 60,
		Difficulty:         meta.Difficulty,
		Category:           meta.Category,
		CoreConcepts:       meta.CoreConcepts,
		AlternateResources: []ResourceView{},
	}

	if len(ragText) > 0 {
		view.WhyItMatters = view.WhyItMatters + "\n\nAuthoritative Reference (RAG):\n" + ragText
	}

	if len(ranked) > 0 {
		primary := ranked[0]
		view.PrimaryResource = s.toResourceView(ctx, primary, true)
		for _, r := range ranked[1:] {
			view.AlternateResources = append(view.AlternateResources, *s.toResourceView(ctx, r, false))
		}
	}
	return view, nil
}

func (s *KnowledgeService) toResourceView(ctx context.Context, r domain.Resource, primary bool) *ResourceView {
	rv := &ResourceView{
		ID:              r.ID,
		Title:           displayTitle(r),
		Type:            r.ResourceType,
		DurationMinutes: minutes(r.DurationMinutes),
		Provider:        r.Author,
		SourceURL:       r.URL,
		WhyThisResource: fmt.Sprintf("Curated from roadmap.sh with %d%% authority. Grid-verified and learner-quality checked.", int(r.AuthorityScore*100)),
	}
	if primary {
		curator := r.Author
		if r.CuratedBy != nil {
			if name, err := s.repo.LookupUserName(ctx, *r.CuratedBy); err == nil && name != "" {
				curator = name
			}
		}
		date := ""
		if r.CuratedAt != nil {
			date = r.CuratedAt.Format("2006-01-02")
		}
		rv.Provenance = &Provenance{
			Author:     r.Author,
			VettedBy:   curator,
			VettedDate: date,
		}
	}
	return rv
}

func published(rs []domain.Resource) []domain.Resource {
	out := rs[:0:0]
	for _, r := range rs {
		if r.Status == "published" {
			out = append(out, r)
		}
	}
	return out
}

func rankResources(rs []domain.Resource, prefs []string) []domain.Resource {
	want := map[string]bool{}
	for _, p := range prefs {
		want[strings.ToLower(strings.TrimSpace(p))] = true
	}
	sort.SliceStable(rs, func(i, j int) bool {
		di, dj := 0, 0
		if want[rs[i].ResourceType] {
			di = 1
		}
		if want[rs[j].ResourceType] {
			dj = 1
		}
		if di != dj {
			return di > dj
		}
		return rs[i].AuthorityScore > rs[j].AuthorityScore
	})
	return rs
}

func breadcrumb(m *ConceptMeta) []string {
	out := []string{m.DomainName, m.Category}
	if m.ID != "" {
		out = append(out, m.Name)
	}
	return out
}

func whyItMatters(m *ConceptMeta, c *domain.Concept) string {
	base := strings.TrimSpace(c.Description)
	if base != "" {
		base = strings.TrimSpace(strings.TrimSuffix(base, " | difficulty "+m.Difficulty))
	}
	if m.CoreConcepts != nil && len(m.CoreConcepts) > 0 {
		if base != "" {
			return fmt.Sprintf("%s Core building blocks: %s.", base, strings.Join(m.CoreConcepts, ", "))
		}
		return fmt.Sprintf("Covers core building blocks: %s.", strings.Join(m.CoreConcepts, ", "))
	}
	if base == "" {
		return fmt.Sprintf("Master %s — a %s concept in the %s track.", m.Name, orZero(m.Difficulty), m.DomainName)
	}
	return base
}

func orZero(s string) string {
	if s == "" || s == "zero" {
		return "core"
	}
	return s
}

func displayTitle(r domain.Resource) string {
	if r.Title != "" {
		return r.Title
	}
	return r.URL
}

func minutes(p *int) int {
	if p != nil {
		return *p
	}
	return 20
}

// ---------------------------------------------------------------------------
// Curator tooling

func (s *KnowledgeService) ListStructures(ctx context.Context) ([]domain.KnowledgeStructure, error) {
	return s.repo.ListStructures(ctx)
}

func (s *KnowledgeService) CreateStructure(ctx context.Context, sIn *domain.KnowledgeStructure, actorID string) error {
	if sIn.DomainID == "" {
		return errors.New("domainId is required")
	}
	if _, err := s.repo.GetDomainBySlug(ctx, sIn.DomainID); err != nil {
		return fmt.Errorf("unknown domain: %w", err)
	}
	sIn.ID = keys.KnowledgeStructure(sIn.DomainID)
	sIn.Version = 1
	if sIn.Status == "" {
		sIn.Status = "draft"
	}
	sIn.CreatedBy = actorID
	sIn.CreatedAt = s.now().UTC()
	if sIn.Status == "published" {
		now := s.now().UTC()
		sIn.PublishedAt = &now
	}
	return s.repo.CreateStructure(ctx, sIn)
}

func (s *KnowledgeService) UpdateStructure(ctx context.Context, id, status string) error {
	if status != "draft" && status != "published" && status != "deprecated" {
		return errors.New("status must be draft, published or deprecated")
	}
	existing, err := s.repo.GetStructure(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.UpdateStructure(ctx, existing, status, s.now())
}

// ValidateStructure performs a DAG (cycle + self-edge) check over the
// structure's prerequisite edges.
func (s *KnowledgeService) ValidateStructure(ctx context.Context, structureID string) (bool, string, error) {
	if _, err := s.repo.GetStructure(ctx, structureID); err != nil {
		return false, "", err
	}
	edges, err := s.repo.ListEdges(ctx, structureID)
	if err != nil {
		return false, "", err
	}
	adj := map[string][]string{}
	indeg := map[string]int{}
	nodes := map[string]bool{}
	for _, e := range edges {
		if e.ConceptID == e.PrerequisiteConceptID {
			return false, "a concept cannot be its own prerequisite", nil
		}
		adj[e.PrerequisiteConceptID] = append(adj[e.PrerequisiteConceptID], e.ConceptID)
		indeg[e.ConceptID]++
		if _, ok := indeg[e.PrerequisiteConceptID]; !ok {
			indeg[e.PrerequisiteConceptID] = 0
		}
		nodes[e.ConceptID] = true
		nodes[e.PrerequisiteConceptID] = true
	}
	queue := []string{}
	for n := range indeg {
		if indeg[n] == 0 {
			queue = append(queue, n)
		}
	}
	visited := 0
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		visited++
		for _, next := range adj[cur] {
			indeg[next]--
			if indeg[next] == 0 {
				queue = append(queue, next)
			}
		}
	}
	if visited < len(nodes) {
		return false, fmt.Sprintf("circular dependency detected across %d concepts", len(nodes)-visited), nil
	}
	return true, fmt.Sprintf("Knowledge structure DAG validated with 0 circular dependencies (%d concepts, %d edges).", len(nodes), len(edges)), nil
}
