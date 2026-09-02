package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hcl-backend/services/api-go/internal/roadmap/domain"
)

type PostgresRoadmapRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRoadmapRepository(db *pgxpool.Pool) *PostgresRoadmapRepository {
	ctx := context.Background()
	_, _ = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS platform.daily_tasks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			learner_id UUID NOT NULL,
			task_date DATE NOT NULL,
			title TEXT NOT NULL,
			category TEXT NOT NULL,
			duration INT NOT NULL,
			completed BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			CONSTRAINT uq_learner_date_title UNIQUE(learner_id, task_date, title)
		);
	`)
	return &PostgresRoadmapRepository{db: db}
}

// GetRoadmap assembles the frontend roadmap view: goal context + ordered
// milestones with public node ids, titles, states and unlock hints.
func (r *PostgresRoadmapRepository) GetRoadmap(ctx context.Context, learnerID string) (*domain.Roadmap, error) {
	query := `
		SELECT p.id, g.id, g.goal_text, ks.id, d.name AS domain_name, d.slug AS domain_slug,
		       p.status, p.created_at, p.updated_at
		FROM platform.paths p
		JOIN platform.goals g ON g.id = p.goal_id
		JOIN platform.knowledge_structures ks ON ks.id = p.knowledge_structure_id
		JOIN platform.domains d ON d.id = ks.domain_id
		WHERE p.learner_id = $1 AND p.status = 'active'
		ORDER BY p.created_at DESC
		LIMIT 1
	`
	var pathID, goalID, goalTitle, structureID, domainName, domainSlug, status string
	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(ctx, query, learnerID).Scan(
		&pathID, &goalID, &goalTitle, &structureID, &domainName, &domainSlug, &status, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrActivePathNotFound
		}
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT c.node_id, c.name, COALESCE(c.description, ''), pi.state,
		       pi.sequence_order, pi.is_remediation,
		       COALESCE((SELECT res.duration_minutes
		                 FROM platform.resource_concepts rc
		                 JOIN platform.resources res ON res.id = rc.resource_id
		                 WHERE rc.concept_id = c.id AND res.status = 'published'
		                 ORDER BY res.duration_minutes LIMIT 1), 45)
		FROM platform.path_items pi
		JOIN platform.concepts c ON c.id = pi.concept_id
		WHERE pi.path_id = $1
		ORDER BY pi.sequence_order ASC`, pathID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []domain.RoadmapNode
	var prevTitle string
	var inProgress string
	competent := 0

	for rows.Next() {
		var nodeID, title, desc, state string
		var order int
		var isRem bool
		var minutes int
		if err := rows.Scan(&nodeID, &title, &desc, &state, &order, &isRem, &minutes); err != nil {
			return nil, err
		}
		node := domain.RoadmapNode{
			ID:               nodeID,
			Title:            title,
			Description:      desc,
			Domain:           domainSlug,
			DomainID:         domainSlug,
			State:            state,
			Order:            order,
			EstimatedMinutes: minutes,
			IsRemediation:    isRem,
		}
		if state == "locked" && prevTitle != "" {
			node.UnlockRequirement = "Unlocks after: " + prevTitle
		}
		if state == "in_progress" && inProgress == "" {
			inProgress = nodeID
		}
		if state == "competent" {
			competent++
		}
		prevTitle = title
		nodes = append(nodes, node)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if nodes == nil {
		nodes = []domain.RoadmapNode{}
	}

	if inProgress == "" && len(nodes) > 0 {
		inProgress = nodes[0].ID
	}

	progress := 0
	if len(nodes) > 0 {
		progress = competent * 100 / len(nodes)
	}

	return &domain.Roadmap{
		GoalID:             goalID,
		GoalTitle:          goalTitle,
		Domain:             domainSlug,
		DomainID:           domainSlug,
		ProgressPercentage: progress,
		CurrentNodeID:      inProgress,
		Nodes:              nodes,
	}, nil
}

func (r *PostgresRoadmapRepository) DeactivatePaths(ctx context.Context, tx pgx.Tx, learnerID, goalID string) error {
	_, err := tx.Exec(ctx, `
		UPDATE platform.paths
		SET status = 'completed', updated_at = now()
		WHERE goal_id = $1 AND status = 'active'`, goalID)
	return err
}

func (r *PostgresRoadmapRepository) CreatePath(ctx context.Context, tx pgx.Tx, path *domain.Path, items []domain.PathItem) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO platform.paths (id, learner_id, goal_id, knowledge_structure_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		path.ID, path.LearnerID, path.GoalID, path.KnowledgeStructureID, path.Status, path.CreatedAt, path.UpdatedAt)
	if err != nil {
		return err
	}
	for _, item := range items {
		_, err := tx.Exec(ctx, `
			INSERT INTO platform.path_items (id, path_id, concept_id, resource_id, sequence_order, state, is_remediation, inserted_at)
			VALUES ($1, $2, (SELECT id FROM platform.concepts WHERE node_id = $3), $4, $5, $6, $7, $8)`,
			item.ID, item.PathID, item.ConceptID, item.ResourceID, item.SequenceOrder, item.State, item.IsRemediation, item.InsertedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRoadmapRepository) GetDailyTasks(ctx context.Context, learnerID string, start, end time.Time) ([]domain.DailyTaskDay, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, task_date, title, category, duration, completed
		FROM platform.daily_tasks
		WHERE learner_id = $1 AND task_date >= $2 AND task_date <= $3
		ORDER BY task_date ASC, title ASC`, learnerID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dayMap := make(map[string][]domain.DailyTask)
	for rows.Next() {
		var id, title, cat string
		var duration int
		var comp bool
		var date time.Time
		if err := rows.Scan(&id, &date, &title, &cat, &duration, &comp); err != nil {
			return nil, err
		}
		dateStr := date.Format("2006-01-02")
		dayMap[dateStr] = append(dayMap[dateStr], domain.DailyTask{
			ID:        id,
			Title:     title,
			Category:  cat,
			Duration:  duration,
			Completed: comp,
		})
	}

	var out []domain.DailyTaskDay
	totalTasks := 0
	for i := 0; i < 7; i++ {
		dayDate := start.AddDate(0, 0, i)
		dateStr := dayDate.Format("2006-01-02")
		tasks := dayMap[dateStr]
		if tasks == nil {
			tasks = []domain.DailyTask{}
		}
		totalTasks += len(tasks)
		out = append(out, domain.DailyTaskDay{
			Date:  dayDate,
			Tasks: tasks,
		})
	}

	if totalTasks > 0 {
		return out, nil
	}

	// Generate adaptively if no tasks exist
	var availability, formatPref, experience string
	err = r.db.QueryRow(ctx, `
		SELECT time_availability, format_preference, prior_experience
		FROM platform.learner_profiles
		WHERE user_id::text = $1`, learnerID).Scan(&availability, &formatPref, &experience)
	if err != nil {
		availability = "10_20"
		formatPref = "article"
		experience = "beginner"
	}

	var conceptID, conceptName string
	err = r.db.QueryRow(ctx, `
		SELECT c.node_id, c.name
		FROM platform.path_items pi
		JOIN platform.concepts c ON c.id = pi.concept_id
		JOIN platform.paths p ON p.id = pi.path_id
		WHERE p.learner_id::text = $1 AND p.status = 'active' AND pi.state = 'in_progress'
		ORDER BY pi.sequence_order LIMIT 1`, learnerID).Scan(&conceptID, &conceptName)
	if err != nil {
		err = r.db.QueryRow(ctx, `
			SELECT c.node_id, c.name
			FROM platform.path_items pi
			JOIN platform.concepts c ON c.id = pi.concept_id
			JOIN platform.paths p ON p.id = pi.path_id
			WHERE p.learner_id::text = $1 AND p.status = 'active' AND pi.state = 'available'
			ORDER BY pi.sequence_order LIMIT 1`, learnerID).Scan(&conceptID, &conceptName)
		if err != nil {
			conceptID = "general_review"
			conceptName = "General Learning Review"
		}
	}

	dailyBudget := 60
	switch availability {
	case "lt_5":
		dailyBudget = 30
	case "5_10":
		dailyBudget = 60
	case "10_20":
		dailyBudget = 120
	case "gt_20":
		dailyBudget = 180
	}

	formats := strings.Split(formatPref, ",")
	primaryFormat := "article"
	if len(formats) > 0 && formats[0] != "" {
		primaryFormat = strings.TrimSpace(formats[0])
	}

	generatedDays := make([]domain.DailyTaskDay, 7)
	for i := 0; i < 7; i++ {
		dayDate := start.AddDate(0, 0, i)
		var tasks []domain.DailyTask
		
		switch i {
		case 0:
			tasks = []domain.DailyTask{
				{Title: fmt.Sprintf("Intro to %s", conceptName), Category: "Foundations", Duration: dailyBudget / 2, Completed: false},
				{Title: fmt.Sprintf("Read documentation for %s", conceptName), Category: "Core", Duration: dailyBudget / 2, Completed: false},
			}
		case 1:
			tasks = []domain.DailyTask{
				{Title: fmt.Sprintf("Review %s core concepts", conceptName), Category: "Foundations", Duration: dailyBudget / 3, Completed: false},
				{Title: fmt.Sprintf("Hands-on exercise: %s", conceptName), Category: "Practice", Duration: dailyBudget * 2 / 3, Completed: false},
			}
		case 2:
			tasks = []domain.DailyTask{
				{Title: fmt.Sprintf("Watch advanced tutorial on %s", conceptName), Category: "Core", Duration: dailyBudget / 2, Completed: false},
				{Title: fmt.Sprintf("Coding lab for %s", conceptName), Category: "Project", Duration: dailyBudget / 2, Completed: false},
			}
		case 3:
			tasks = []domain.DailyTask{
				{Title: fmt.Sprintf("Optimize implementation of %s", conceptName), Category: "Practice", Duration: dailyBudget / 2, Completed: false},
				{Title: fmt.Sprintf("Take checkpoint quiz for %s", conceptName), Category: "Assessment", Duration: dailyBudget / 2, Completed: false},
			}
		case 4:
			tasks = []domain.DailyTask{
				{Title: fmt.Sprintf("Deploy first project for %s", conceptName), Category: "Project", Duration: dailyBudget * 3 / 4, Completed: false},
				{Title: fmt.Sprintf("Verify weak areas in %s", conceptName), Category: "Assessment", Duration: dailyBudget / 4, Completed: false},
			}
		case 5:
			tasks = []domain.DailyTask{
				{Title: fmt.Sprintf("Deep-dive articles on %s", conceptName), Category: "Foundations", Duration: dailyBudget, Completed: false},
			}
		case 6:
			tasks = []domain.DailyTask{
				{Title: fmt.Sprintf("Self-assessment check for %s", conceptName), Category: "Assessment", Duration: dailyBudget, Completed: false},
			}
		}

		for idx := range tasks {
			if strings.Contains(strings.ToLower(tasks[idx].Title), "read") || strings.Contains(strings.ToLower(tasks[idx].Title), "documentation") {
				if primaryFormat == "video" {
					tasks[idx].Title = strings.Replace(tasks[idx].Title, "Read", "Watch video about", 1)
				} else if primaryFormat == "interactive" {
					tasks[idx].Title = strings.Replace(tasks[idx].Title, "Read", "Interactive lab on", 1)
					tasks[idx].Title = strings.Replace(tasks[idx].Title, "read", "interactive lab on", 1)
				}
			}
		}

		generatedDays[i] = domain.DailyTaskDay{
			Date:  dayDate,
			Tasks: tasks,
		}
	}

	err = r.SaveDailyTasks(ctx, learnerID, generatedDays)
	if err != nil {
		return nil, err
	}

	// Fetch again now that IDs exist
	return r.GetDailyTasks(ctx, learnerID, start, end)
}

func (r *PostgresRoadmapRepository) SaveDailyTasks(ctx context.Context, learnerID string, tasks []domain.DailyTaskDay) error {
	for _, day := range tasks {
		for _, t := range day.Tasks {
			_, err := r.db.Exec(ctx, `
				INSERT INTO platform.daily_tasks (learner_id, task_date, title, category, duration, completed)
				VALUES ($1::uuid, $2, $3, $4, $5, $6)
				ON CONFLICT (learner_id, task_date, title) DO NOTHING`,
				learnerID, day.Date.Format("2006-01-02"), t.Title, t.Category, t.Duration, t.Completed)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *PostgresRoadmapRepository) ToggleDailyTask(ctx context.Context, learnerID string, taskID string, completed bool) error {
	_, err := r.db.Exec(ctx, `
		UPDATE platform.daily_tasks
		SET completed = $1
		WHERE id::text = $2 AND learner_id = $3`, completed, taskID, learnerID)
	return err
}
