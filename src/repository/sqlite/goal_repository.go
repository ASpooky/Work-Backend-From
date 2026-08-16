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
		`INSERT INTO goals (id, workspace_id, title, detail, achievement_condition, end_date, mode, status, created_at, postpone_count, priority)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		goal.ID, goal.WorkspaceID, goal.Title, goal.Detail, goal.AchievementCondition,
		goal.EndDate.Format(dateLayout), string(goal.Mode), string(goal.Status), goal.CreatedAt.Format(time.RFC3339),
		goal.PostponeCount, goal.Priority,
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

// Delete removes a goal and everything scoped to it (its daily tasks, and
// any goal-review AI conversations plus their messages) in a single
// transaction, mirroring WorkspaceRepository.Delete's approach since SQLite
// foreign keys aren't enforced here. A goal-unscoped (general) conversation
// in the same workspace is untouched — it isn't this goal's.
func (r *GoalRepository) Delete(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	statements := []struct {
		query string
		args  []any
	}{
		{`DELETE FROM conversation_messages WHERE conversation_id IN (SELECT id FROM conversations WHERE goal_id = ?)`, []any{id}},
		{`DELETE FROM conversations WHERE goal_id = ?`, []any{id}},
		{`DELETE FROM daily_tasks WHERE goal_id = ?`, []any{id}},
		{`DELETE FROM goals WHERE id = ?`, []any{id}},
	}
	for _, s := range statements {
		if _, err := tx.Exec(s.query, s.args...); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *GoalRepository) UpdatePostponement(id string, endDate time.Time, postponeCount int) error {
	_, err := r.db.Exec(
		`UPDATE goals SET end_date = ?, postpone_count = ? WHERE id = ?`,
		endDate.Format(dateLayout), postponeCount, id,
	)
	return err
}

// UpdatePriority sets a goal's manual sort position (lower sorts first). A
// separate narrow method from Update, same reasoning as UpdatePostponement —
// this is a list-ordering concern, not part of the "目標の見直し" edit flow.
func (r *GoalRepository) UpdatePriority(id string, priority int) error {
	_, err := r.db.Exec(`UPDATE goals SET priority = ? WHERE id = ?`, priority, id)
	return err
}

func (r *GoalRepository) FindByID(id string) (*entity.Goal, error) {
	row := r.db.QueryRow(
		`SELECT id, workspace_id, title, detail, achievement_condition, end_date, mode, status, created_at, postpone_count, priority
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
		`SELECT id, workspace_id, title, detail, achievement_condition, end_date, mode, status, created_at, postpone_count, priority
		 FROM goals ORDER BY priority ASC, created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanGoals(rows)
}

func (r *GoalRepository) FindByWorkspaceID(workspaceID string) ([]*entity.Goal, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, title, detail, achievement_condition, end_date, mode, status, created_at, postpone_count, priority
		 FROM goals WHERE workspace_id = ? ORDER BY priority ASC, created_at ASC`,
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
	var postponeCount, priority int
	if err := row.Scan(&id, &wsID, &title, &detail, &achievementCondition, &endDate, &mode, &status, &createdAt, &postponeCount, &priority); err != nil {
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
	goal.Priority = priority
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
