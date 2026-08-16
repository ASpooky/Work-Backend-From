package sqlite

import (
	"database/sql"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

const dateLayout = "2006-01-02"

type GoalRepository struct {
	db *sql.DB
}

func NewGoalRepository(db *sql.DB) *GoalRepository {
	return &GoalRepository{db: db}
}

func (r *GoalRepository) Save(goal *entity.Goal) error {
	_, err := r.db.Exec(
		`INSERT INTO goals (id, workspace_id, title, detail, achievement_condition, end_date, mode, status, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		goal.ID, goal.WorkspaceID, goal.Title, goal.Detail, goal.AchievementCondition,
		goal.EndDate.Format(dateLayout), string(goal.Mode), string(goal.Status), goal.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *GoalRepository) UpdateEndDate(id string, endDate time.Time) error {
	_, err := r.db.Exec(`UPDATE goals SET end_date = ? WHERE id = ?`, endDate.Format(dateLayout), id)
	return err
}

func (r *GoalRepository) FindByWorkspaceID(workspaceID string) ([]*entity.Goal, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, title, detail, achievement_condition, end_date, mode, status, created_at
		 FROM goals WHERE workspace_id = ? ORDER BY created_at`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	goals := []*entity.Goal{}
	for rows.Next() {
		var id, wsID, title, detail, achievementCondition, endDate, mode, status, createdAt string
		if err := rows.Scan(&id, &wsID, &title, &detail, &achievementCondition, &endDate, &mode, &status, &createdAt); err != nil {
			return nil, err
		}

		parsedEndDate, err := time.Parse(dateLayout, endDate)
		if err != nil {
			return nil, err
		}
		parsedCreatedAt, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, err
		}

		goal := entity.NewGoal(id, wsID, title, detail, achievementCondition, parsedEndDate, entity.GoalMode(mode), parsedCreatedAt)
		goal.Status = entity.GoalStatus(status)
		goals = append(goals, goal)
	}
	return goals, rows.Err()
}
