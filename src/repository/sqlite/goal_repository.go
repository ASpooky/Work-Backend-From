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
		`INSERT INTO goals (id, workspace_id, title, detail, achievement_condition, end_date, mode, status, created_at, postpone_count)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		goal.ID, goal.WorkspaceID, goal.Title, goal.Detail, goal.AchievementCondition,
		goal.EndDate.Format(dateLayout), string(goal.Mode), string(goal.Status), goal.CreatedAt.Format(time.RFC3339),
		goal.PostponeCount,
	)
	return err
}

// Update overwrites a goal's editable fields (title/detail/achievement
// condition/end date/mode) for the "目標の見直し" edit flow. Status,
// workspace, created_at, and postpone_count are never touched here.
func (r *GoalRepository) Update(goal *entity.Goal) error {
	_, err := r.db.Exec(
		`UPDATE goals SET title = ?, detail = ?, achievement_condition = ?, end_date = ?, mode = ? WHERE id = ?`,
		goal.Title, goal.Detail, goal.AchievementCondition, goal.EndDate.Format(dateLayout), string(goal.Mode), goal.ID,
	)
	return err
}

func (r *GoalRepository) UpdatePostponement(id string, endDate time.Time, postponeCount int) error {
	_, err := r.db.Exec(
		`UPDATE goals SET end_date = ?, postpone_count = ? WHERE id = ?`,
		endDate.Format(dateLayout), postponeCount, id,
	)
	return err
}

func (r *GoalRepository) FindByID(id string) (*entity.Goal, error) {
	row := r.db.QueryRow(
		`SELECT id, workspace_id, title, detail, achievement_condition, end_date, mode, status, created_at, postpone_count
		 FROM goals WHERE id = ?`,
		id,
	)

	goal, err := scanGoal(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return goal, err
}

// FindAll returns every goal across every workspace, for the cross-workspace
// "all" view.
func (r *GoalRepository) FindAll() ([]*entity.Goal, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, title, detail, achievement_condition, end_date, mode, status, created_at, postpone_count
		 FROM goals ORDER BY created_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanGoals(rows)
}

func (r *GoalRepository) FindByWorkspaceID(workspaceID string) ([]*entity.Goal, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, title, detail, achievement_condition, end_date, mode, status, created_at, postpone_count
		 FROM goals WHERE workspace_id = ? ORDER BY created_at`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanGoals(rows)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanGoal(row rowScanner) (*entity.Goal, error) {
	var id, wsID, title, detail, achievementCondition, endDate, mode, status, createdAt string
	var postponeCount int
	if err := row.Scan(&id, &wsID, &title, &detail, &achievementCondition, &endDate, &mode, &status, &createdAt, &postponeCount); err != nil {
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
	goal.PostponeCount = postponeCount
	return goal, nil
}

func scanGoals(rows *sql.Rows) ([]*entity.Goal, error) {
	goals := []*entity.Goal{}
	for rows.Next() {
		goal, err := scanGoal(rows)
		if err != nil {
			return nil, err
		}
		goals = append(goals, goal)
	}
	return goals, rows.Err()
}
