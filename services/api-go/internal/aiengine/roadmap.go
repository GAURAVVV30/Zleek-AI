package aiengine

// RoadmapStore is a Go port of the FastAPI api/v1/roadmap.py persistent store.
// The embedded roadmaps.json acts as the seed; runtime submissions are kept in
// memory and best-effort persisted to AI_ROADMAP_DATA_FILE (the embedded file
// itself is read-only).

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

var sampleRoadmap = map[string]any{
	"domain": "software_architect",
	"nodes": []any{
		map[string]any{"id": "intro", "label": "Intro to Architecture"},
		map[string]any{"id": "design", "label": "System Design"},
		map[string]any{"id": "cloud", "label": "Cloud Patterns"},
		map[string]any{"id": "security", "label": "Security"},
	},
	"edges": []any{
		map[string]any{"source": "intro", "target": "design"},
		map[string]any{"source": "design", "target": "cloud"},
		map[string]any{"source": "cloud", "target": "security"},
	},
}

type RoadmapStore struct {
	mu       sync.RWMutex
	store    map[string]any
	dataFile string
}

// NewRoadmapStore seeds from the embedded roadmaps.json (Python _load_store on
// first access) unless an overridden AI_ROADMAP_DATA_FILE already exists.
func NewRoadmapStore() *RoadmapStore {
	dataFile := strings.TrimSpace(os.Getenv("AI_ROADMAP_DATA_FILE"))
	store := map[string]any{}
	var err error
	if dataFile != "" {
		if raw, readErr := os.ReadFile(dataFile); readErr == nil {
			err = json.Unmarshal(raw, &store)
		}
	}
	if err != nil || len(store) == 0 {
		if raw, readErr := dataFS.ReadFile(roadmapDataFile); readErr == nil {
			_ = json.Unmarshal(raw, &store)
		}
	}
	if dataFile == "" {
		dataFile = "roadmaps.json"
	}
	return &RoadmapStore{store: store, dataFile: dataFile}
}

func normalizeDomainKey(domain string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(domain)), " ", "_")
}

// Get returns a stored roadmap for a domain, else the sample (like Python).
func (rs *RoadmapStore) Get(domain string) map[string]any {
	key := normalizeDomainKey(domain)
	rs.mu.RLock()
	if v, ok := rs.store[key]; ok {
		if vm, isMap := v.(map[string]any); isMap {
			rs.mu.RUnlock()
			return vm
		}
	}
	rs.mu.RUnlock()

	result := make(map[string]any, len(sampleRoadmap)+1)
	for k, v := range sampleRoadmap {
		result[k] = v
	}
	result["domain"] = domain
	return result
}

// List returns the stored domain keys (Python: list(store.keys())).
func (rs *RoadmapStore) List() []string {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	keys := make([]string, 0, len(rs.store))
	for k := range rs.store {
		keys = append(keys, k)
	}
	return keys
}

// Put stores a roadmap record under the normalized domain key.
func (rs *RoadmapStore) Put(domain string, record map[string]any) error {
	rs.mu.Lock()
	rs.store[normalizeDomainKey(domain)] = record
	rs.mu.Unlock()
	return rs.save()
}

func (rs *RoadmapStore) save() error {
	rs.mu.RLock()
	data, err := json.MarshalIndent(rs.store, "", "  ")
	rs.mu.RUnlock()
	if err != nil {
		return err
	}
	return writeAtomic(rs.dataFile, data)
}

func writeAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// BuildRoadmapRecord mirrors the POST /roadmap record assembly.
func BuildRoadmapRecord(domain string, nodes, edges []map[string]any) map[string]any {
	return map[string]any{"domain": domain, "nodes": nodes, "edges": edges}
}
