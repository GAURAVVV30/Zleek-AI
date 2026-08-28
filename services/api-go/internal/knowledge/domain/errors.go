package domain

import "errors"

var (
	ErrDomainNotFound             = errors.New("domain not found")
	ErrKnowledgeStructureNotFound = errors.New("knowledge structure not found")
	ErrConceptNotFound            = errors.New("concept not found")
	ErrInvalidDAG                 = errors.New("invalid prerequisite DAG structure")
)
