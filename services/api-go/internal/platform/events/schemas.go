package events

const (
	EventTypeCompetencyUpdated = "CompetencyUpdated"
	EventTypeConceptWeak       = "ConceptWeak"
	EventTypeGoalAchieved      = "GoalAchieved"
	EventTypeResourceFlagged   = "ResourceFlagged"
)

type CompetencyUpdatedPayload struct {
	LearnerID string `json:"learner_id"`
	ConceptID string `json:"concept_id"`
	NewState  string `json:"new_state"`
}

type ConceptWeakPayload struct {
	LearnerID string `json:"learner_id"`
	ConceptID string `json:"concept_id"`
}

type GoalAchievedPayload struct {
	LearnerID string `json:"learner_id"`
	GoalID    string `json:"goal_id"`
}

type ResourceFlaggedPayload struct {
	ResourceID string `json:"resource_id"`
	LearnerID  string `json:"learner_id"`
}
