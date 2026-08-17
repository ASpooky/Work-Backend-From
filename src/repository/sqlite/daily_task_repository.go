package sqlite

import (
	"database/sql"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type DailyTaskRepository struct {
	db *sql.DB
}

func NewDailyTaskRepository(db *sql.DB) *DailyTaskRepository {
	return &DailyTaskRepository{db: db}
}

func (r *DailyTaskRepository) Save(task *entity.DailyTask) error {
	done := 0
	if task.Done {
		done = 1
	}
	_, err := r.db.Exec(
		`INSERT INTO daily_tasks (id, goal_id, date, content, done, created_at, completed_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		task.ID, task.GoalID, task.Date.Format(dateLayout), task.Content, done, task.CreatedAt.Format(time.RFC3339),
		nullableTimeString(task.CompletedAt),
	)
	return err
}

func (r *DailyTaskRepository) FindByGoalIDAndDateRange(goalID string, from, to time.Time) ([]*entity.DailyTask, error) {
	rows, err := r.db.Query(
		`SELECT id, goal_id, date, content, done, created_at, completed_at, memo FROM daily_tasks
		 WHERE goal_id = ? AND date >= ? AND date <= ? ORDER BY date`,
		goalID, from.Format(dateLayout), to.Format(dateLayout),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDailyTasks(rows)
}

// UpdateDone records not just whether a task is done but when: completedAt
// is nil when marking a task not-done, clearing any previous timestamp
// rather than leaving it stale.
func (r *DailyTaskRepository) UpdateDone(id string, done bool, completedAt *time.Time) error {
	doneInt := 0
	if done {
		doneInt = 1
	}
	_, err := r.db.Exec(
		`UPDATE daily_tasks SET done = ?, completed_at = ? WHERE id = ?`,
		doneInt, nullableTimeString(completedAt), id,
	)
	return err
}

// Delete removes a single daily task — used when an AI goal-revision
// proposal removes an existing task from the schedule.
func (r *DailyTaskRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM daily_tasks WHERE id = ?`, id)
	return err
}

// UpdateMemo sets or clears (memo == nil) a task's free-text note. Editable
// independent of Done/CompletedAt — a memo can be added, changed, or
// removed at any time, not just at completion.
func (r *DailyTaskRepository) UpdateMemo(id string, memo *string) error {
	var value any
	if memo != nil {
		value = *memo
	}
	_, err := r.db.Exec(`UPDATE daily_tasks SET memo = ? WHERE id = ?`, value, id)
	return err
}

// FindOldestPendingBefore returns the earliest not-done task for goalID with
// a date strictly before `before`, or nil if there is none.
func (r *DailyTaskRepository) FindOldestPendingBefore(goalID string, before time.Time) (*entity.DailyTask, error) {
	row := r.db.QueryRow(
		`SELECT id, goal_id, date, content, done, created_at, completed_at, memo FROM daily_tasks
		 WHERE goal_id = ? AND done = 0 AND date < ?
		 ORDER BY date ASC LIMIT 1`,
		goalID, before.Format(dateLayout),
	)

	task, err := scanDailyTask(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return task, err
}

// ShiftPendingForward pushes every not-done task for goalID with a date on
// or after fromDate forward by one day. Done tasks are left untouched so
// historical completions are never rewritten.
func (r *DailyTaskRepository) ShiftPendingForward(goalID string, fromDate time.Time) error {
	_, err := r.db.Exec(
		`UPDATE daily_tasks SET date = date(date, '+1 day')
		 WHERE goal_id = ? AND done = 0 AND date >= ?`,
		goalID, fromDate.Format(dateLayout),
	)
	return err
}

// FindByDateAndWorkspaceID returns only the tasks scheduled for date whose
// goal belongs to workspaceID, unlike FindByDate which ignores workspace
// boundaries entirely.
func (r *DailyTaskRepository) FindByDateAndWorkspaceID(date time.Time, workspaceID string) ([]*entity.DailyTask, error) {
	rows, err := r.db.Query(
		`SELECT dt.id, dt.goal_id, dt.date, dt.content, dt.done, dt.created_at, dt.completed_at, dt.memo
		 FROM daily_tasks dt
		 JOIN goals g ON g.id = dt.goal_id
		 WHERE dt.date = ? AND g.workspace_id = ?
		 ORDER BY dt.created_at`,
		date.Format(dateLayout), workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDailyTasks(rows)
}

// FindByWorkspaceIDAndDateRange returns every task in [from, to] whose goal
// belongs to workspaceID, across all of that workspace's goals — used to
// build a "here's what's already scheduled elsewhere" summary for the AI
// planning/review prompts, not scoped to any single goal.
func (r *DailyTaskRepository) FindByWorkspaceIDAndDateRange(workspaceID string, from, to time.Time) ([]*entity.DailyTask, error) {
	rows, err := r.db.Query(
		`SELECT dt.id, dt.goal_id, dt.date, dt.content, dt.done, dt.created_at, dt.completed_at, dt.memo
		 FROM daily_tasks dt
		 JOIN goals g ON g.id = dt.goal_id
		 WHERE g.workspace_id = ? AND dt.date >= ? AND dt.date <= ?
		 ORDER BY dt.date`,
		workspaceID, from.Format(dateLayout), to.Format(dateLayout),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDailyTasks(rows)
}

func (r *DailyTaskRepository) FindByDate(date time.Time) ([]*entity.DailyTask, error) {
	rows, err := r.db.Query(
		`SELECT id, goal_id, date, content, done, created_at, completed_at, memo FROM daily_tasks WHERE date = ? ORDER BY created_at`,
		date.Format(dateLayout),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDailyTasks(rows)
}

func nullableTimeString(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Format(time.RFC3339)
}

func scanDailyTask(row rowScanner) (*entity.DailyTask, error) {
	var id, goalID, dateStr, content, createdAt string
	var done int
	var completedAt, memo sql.NullString
	if err := row.Scan(&id, &goalID, &dateStr, &content, &done, &createdAt, &completedAt, &memo); err != nil {
		return nil, err
	}

	parsedDate, err := time.Parse(dateLayout, dateStr)
	if err != nil {
		return nil, err
	}
	parsedCreatedAt, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, err
	}

	task := entity.NewDailyTask(id, goalID, parsedDate, content, parsedCreatedAt)
	task.Done = done != 0
	if completedAt.Valid {
		parsedCompletedAt, err := time.Parse(time.RFC3339, completedAt.String)
		if err != nil {
			return nil, err
		}
		task.CompletedAt = &parsedCompletedAt
	}
	if memo.Valid {
		task.Memo = &memo.String
	}
	return task, nil
}

func scanDailyTasks(rows *sql.Rows) ([]*entity.DailyTask, error) {
	tasks := []*entity.DailyTask{}
	for rows.Next() {
		task, err := scanDailyTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}
