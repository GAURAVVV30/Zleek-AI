// Package keys provides deterministic UUIDs for seeded platform rows. Node
// concepts, resources, users, domains and knowledge structures must be
// addressable across modules (bootstrap seeds, knowledge resolves a concept
// back to its roadmap.sh graph node, roadmap maps graph nodes to concepts).
package keys

import (
	"github.com/google/uuid"
)

func namespace() uuid.UUID { return uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8") }

func UUID(key string) string { return uuid.NewSHA1(namespace(), []byte("hcl:"+key)).String() }

func User(key string) string { return UUID("user:" + key) }

func Domain(domainID string) string { return UUID("domain:" + domainID) }

func KnowledgeStructure(domainID string) string { return UUID("ks:" + domainID) }

func Concept(nodeID string) string { return UUID("concept:" + nodeID) }

func Resource(url string) string { return UUID("resource:" + url) }
