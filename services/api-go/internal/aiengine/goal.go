package aiengine

// GoalAnalyzer is a faithful Go port of app/core/goal_intelligence.py.

import (
	"fmt"
	"strings"
)

// AnalyzeUserIntent mirrors GoalAnalyzer.analyze_user_intent().
func (engine *GraphEngine) AnalyzeUserIntent(userText string, llm *LLMClient) map[string]any {
	domains := engine.DomainList
	domainsText := ""
	for i, d := range domains {
		if i > 0 {
			domainsText += ", "
		}
		domainsText += d
	}

	systemPrompt := "You are a career mapping assistant. The user will state their goal. " +
		fmt.Sprintf("You must map their intent to one of the following available domain IDs: %s. ", domainsText) +
		"You must also extract any technical skills they claim to already know."

	schema := map[string]any{
		"mapped_domain_id":       map[string]any{"type": "string"},
		"extracted_prior_skills": map[string]any{"type": "array"},
		"reasoning":              map[string]any{"type": "string"},
	}

	result := llm.GenerateStructuredJSON(systemPrompt, userText, schema)
	if e, ok := result["error"]; ok && e != nil {
		return map[string]any{"error": e, "raw": result["raw"]}
	}

	mapped, _ := result["mapped_domain_id"].(string)
	skillsRaw, _ := result["extracted_prior_skills"].([]any)
	reasoning, _ := result["reasoning"].(string)

	if mapped == "" {
		return map[string]any{"error": "Invalid or missing 'mapped_domain_id' from LLM.", "raw": result}
	}

	skills := append([]string(nil), anyStringSlice(result["extracted_prior_skills"])...)
	_ = skillsRaw

	if !engine.DomainExists(mapped) {
		return map[string]any{
			"mapped_domain_id":       mapped,
			"extracted_prior_skills": skills,
			"reasoning":              reasoning,
			"warning":                "Mapped domain is not in available domains.",
		}
	}

	return map[string]any{
		"mapped_domain_id":       mapped,
		"extracted_prior_skills": skills,
		"reasoning":              reasoning,
	}
}

// domainKeywords is a deterministic fallback mapper used when the LLM is not
// reachable so roadmap/goal flows still produce a valid, seeded domain.
var domainKeywords = map[string][]string{
	"machine_learning":   {"machine learning", "ml", "data scientist", "data science", "model", "deep learning"},
	"ai_engineer":        {"ai engineer", "artificial intelligence", "generative ai", "llm", "agent"},
	"ai_data_scientist":  {"ai data scientist", "ai analyst", "intelligence data"},
	"data_analyst":       {"data analyst", "analytics", "dashboard", "power bi", "excel analyst"},
	"data_engineer":      {"data engineer", "etl", "pipeline", "warehouse"},
	"backend_engineer":   {"backend", "back end", "api", "server side", "microservices"},
	"frontend_engineer":  {"frontend", "front end", "react", "ui", "javascript engineer"},
	"full_stack":         {"full stack", "fullstack", "full-stack"},
	"devops_sre":         {"devops", "sre", "site reliability", "infrastructure", "kubernetes", "terraform", "ci/cd"},
	"software_architect": {"software architect", "system design", "architecture", "architect"},
	"product_manager":    {"product manager", "product management", "pm"},
	"mobile_engineer":    {"mobile", "ios", "android", "mobile app"},
}

// MatchDomainByKeywords maps goal text to a seeded domain ID by keyword score.
// Returns "" when nothing matches. Highest keyword-count match wins.
func (engine *GraphEngine) MatchDomainByKeywords(goalText string) string {
	lower := strings.ToLower(goalText)
	best := ""
	bestScore := 0
	for domain, keys := range domainKeywords {
		if !engine.DomainExists(domain) {
			continue
		}
		score := 0
		for _, k := range keys {
			if strings.Contains(lower, k) {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			best = domain
		}
	}
	return best
}
