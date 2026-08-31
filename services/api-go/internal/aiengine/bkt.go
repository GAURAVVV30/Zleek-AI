package aiengine

// BktEstimator is a faithful Go port of the FastAPI app/core/bkt_estimator.py.
// Bayesian Knowledge Tracing forward pass; mastery declared at >= 0.95.

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	DefaultPInit            = 0.30
	DefaultPLearn           = 0.20
	DefaultPGuess           = 0.25
	DefaultPSlip            = 0.10
	DefaultMasteryThreshold = 0.95
)

// BktParams mirrors the four BKT skill parameters.
type BktParams struct {
	PInit  float64
	PLearn float64
	PGuess float64
	PSlip  float64
}

func DefaultBktParams() BktParams {
	return BktParams{PInit: DefaultPInit, PLearn: DefaultPLearn, PGuess: DefaultPGuess, PSlip: DefaultPSlip}
}

func (b BktParams) toDict() map[string]float64 {
	return map[string]float64{"p_init": b.PInit, "p_learn": b.PLearn, "p_guess": b.PGuess, "p_slip": b.PSlip}
}

// BKTResult mirrors the dict returned by BKTEstimator.estimate().
type BKTResult struct {
	PMastery  float64            `json:"p_mastery"`
	Mastered  bool               `json:"mastered"`
	Threshold float64            `json:"threshold"`
	Attempts  int                `json:"attempts"`
	Params    map[string]float64 `json:"params"`
	History   []int              `json:"history"`
}

// skillParams is the per-node parameter registry (Python BKTEstimator._skill_params).
var skillParams = map[string]BktParams{}

// RegisterSkill mirrors register_skill().
func RegisterSkill(nodeID string, p BktParams) {
	skillParams[nodeID] = p
}

// GetSkillParams mirrors get_params(): registered params or the defaults.
func GetSkillParams(nodeID string) BktParams {
	if p, ok := skillParams[nodeID]; ok {
		return p
	}
	return DefaultBktParams()
}

// bktForward mirrors _bkt_forward().
func bktForward(responses []int, pInit, pLearn, pGuess, pSlip float64) float64 {
	pMastery := pInit
	for _, obs := range responses {
		var pObsGivenMastered, pObsGivenUnmastered float64
		if obs == 1 {
			pObsGivenMastered = 1.0 - pSlip
			pObsGivenUnmastered = pGuess
		} else {
			pObsGivenMastered = pSlip
			pObsGivenUnmastered = 1.0 - pGuess
		}
		pObsMastered := pObsGivenMastered * pMastery
		pObsUnmastered := pObsGivenUnmastered * (1.0 - pMastery)
		pObs := pObsMastered + pObsUnmastered

		var pMasteryGivenObs float64
		if pObs == 0.0 {
			pMasteryGivenObs = pMastery
		} else {
			pMasteryGivenObs = pObsMastered / pObs
		}
		pMastery = pMasteryGivenObs + (1.0-pMasteryGivenObs)*pLearn
	}
	return math.Min(math.Max(pMastery, 0.0), 1.0)
}

// BktEstimate mirrors estimate(): full attempt-history estimation for a node.
func BktEstimate(nodeID string, attemptHistory []int, customParams *BktParams) BKTResult {
	params := DefaultBktParams()
	if customParams != nil {
		params = *customParams
	} else {
		params = GetSkillParams(nodeID)
	}

	var pMastery float64
	if len(attemptHistory) == 0 {
		pMastery = params.PInit
	} else {
		pMastery = bktForward(attemptHistory, params.PInit, params.PLearn, params.PGuess, params.PSlip)
	}

	return BKTResult{
		PMastery:  round4(pMastery),
		Mastered:  pMastery >= DefaultMasteryThreshold,
		Threshold: DefaultMasteryThreshold,
		Attempts:  len(attemptHistory),
		Params:    params.toDict(),
		History:   append([]int(nil), attemptHistory...),
	}
}

// BktEstimateIncremental mirrors estimate_incremental().
func BktEstimateIncremental(nodeID string, currentPMastery float64, newResponse int, customParams *BktParams) BKTResult {
	params := DefaultBktParams()
	if customParams != nil {
		params = *customParams
	} else {
		params = GetSkillParams(nodeID)
	}
	pMastery := bktForward([]int{newResponse}, currentPMastery, params.PLearn, params.PGuess, params.PSlip)
	return BKTResult{
		PMastery:  round4(pMastery),
		Mastered:  pMastery >= DefaultMasteryThreshold,
		Threshold: DefaultMasteryThreshold,
		Attempts:  1,
		Params:    params.toDict(),
		History:   []int{newResponse},
	}
}

// RegisteredSkills returns a snapshot of the per-node BKT parameter registry.
func RegisteredSkills() map[string]BktParams {
	out := make(map[string]BktParams, len(skillParams))
	for k, v := range skillParams {
		out[k] = v
	}
	return out
}

// BktParamsToList mirrors get_params()/estimate() "params" dict shape.
func BktParamsToList(p BktParams) map[string]float64 {
	return p.toDict()
}

// round4 mirrors round(value, 4).
func round4(v float64) float64 {
	return math.Round(v*1e4) / 1e4
}

// ParseBktParams parses string overrides for the guardrails/debug params endpoint.
func ParseBktParams(pInit, pLearn, pGuess, pSlip string) (BktParams, error) {
	bp := DefaultBktParams()
	parse := func(s string, target *float64, name string) error {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		if v < 0 || v > 1 {
			return fmt.Errorf("%s out of range (0..1): %v", name, v)
		}
		*target = v
		return nil
	}
	if err := parse(pInit, &bp.PInit, "p_init"); err != nil {
		return bp, err
	}
	if err := parse(pLearn, &bp.PLearn, "p_learn"); err != nil {
		return bp, err
	}
	if err := parse(pGuess, &bp.PGuess, "p_guess"); err != nil {
		return bp, err
	}
	if err := parse(pSlip, &bp.PSlip, "p_slip"); err != nil {
		return bp, err
	}
	return bp, nil
}
