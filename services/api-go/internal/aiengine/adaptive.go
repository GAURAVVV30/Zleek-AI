package aiengine

// AdaptiveEngine is a faithful Go port of app/core/adaptive_engine.py.

import (
	"math"
	"strconv"
)

const (
	BktMasteryThreshold  = 0.95
	LegacyScoreThreshold = 0.80
)

type NextActionResult struct {
	NodeID        string   `json:"node_id"`
	Action        string   `json:"action"`
	Message       string   `json:"message"`
	DecisionBasis string   `json:"decision_basis"`
	PMastery      *float64 `json:"p_mastery"`
}

func pf(v float64) *float64 { return &v }

// DetermineNextAction mirrors AdaptiveEngine.determine_next_action().
func DetermineNextAction(nodeID string, score float64, failedAttempts int, pMastery *float64) NextActionResult {
	// ---- Determine mastery using BKT if available ----
	if pMastery != nil {
		v := *pMastery
		if math.IsNaN(v) { // TypeError-equivalent
			v = 0.0
		}
		if v >= BktMasteryThreshold {
			return NextActionResult{
				NodeID:        nodeID,
				Action:        "advance",
				Message:       "Mastery confirmed — P(mastery) = " + strconv.Itoa(int(math.Round(v)*100)) + "%. You're ready for the next concept.",
				DecisionBasis: "bkt_mastery",
				PMastery:      pf(round4(v)),
			}
		}
	} else {
		// ---- Legacy: fall back to raw score ----
		if score >= LegacyScoreThreshold {
			return NextActionResult{
				NodeID:        nodeID,
				Action:        "advance",
				Message:       "Competency proven. Ready for the next node.",
				DecisionBasis: "legacy_score",
				PMastery:      nil,
			}
		}
	}

	// ---- Learner has not yet reached mastery ----
	if failedAttempts <= 0 {
		failedAttempts = 1
	}
	basis := "legacy_score"
	if pMastery != nil {
		basis = "bkt_mastery"
	}
	var pM *float64
	if pMastery != nil {
		pM = pf(round4(*pMastery))
	}

	if failedAttempts == 1 {
		return NextActionResult{
			NodeID:        nodeID,
			Action:        "remediate",
			Message:       "Let's review the foundational concepts again.",
			DecisionBasis: basis,
			PMastery:      pM,
		}
	}
	// failed_attempts >= 2
	return NextActionResult{
		NodeID:        nodeID,
		Action:        "human_intervention",
		Message:       "You seem stuck. Let's try a completely different approach or project.",
		DecisionBasis: basis,
		PMastery:      pM,
	}
}
