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
		`INSERT INTO daily_tasks (id, goal_id, date, content, done, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		task.ID, task.GoalID, task.Date.Format(dateLayout), task.Content, done, task.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *DailyTaskRepository) FindByGoalIDAndDateRange(goalID string, from, to time.Time) ([]*entity.DailyTask, error) {
	rows, err := r.db.Query(
		`SELECT id, goal_id, date, content, done, created_at FROM daily_tasks
		 WHERE goal_id = ? AND date >= ? AND date <= ? ORDER BY date`,
		goalID, from.Format(dateLayout), to.Format(dateLayout),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []*entity.DailyTask{}
	for rows.Next() {
		var id, gID, dateStr, content, createdAt string
		var done int
		if err := rows.Scan(&id, &gID, &dateStr, &content, &done, &createdAt); err != nil {
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

		task := entity.NewDailyTask(id, gID, parsedDate, content, parsedCreatedAt)
		task.Done = done != 0
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *DailyTaskRepository) UpdateDone(id string, done bool) error {
	doneInt := 0
	if done {
		doneInt = 1
	}
	_, err := r.db.Exec(`UPDATE daily_tasks SET done = ? WHERE id = ?`, doneInt, id)
	return err
}

// FindOldestPendingBefore returns the earliest not-done task for goalID with
// a date strictly before `before`, or nil if there is none.
func (r *DailyTaskRepository) FindOldestPendingBefore(goalID string, before time.Time) (*entity.DailyTask, error) {
	row := r.db.QueryRow(
		`SELECT id, goal_id, date, content, done, created_at FROM daily_tasks
		 WHERE goal_id = ? AND done = 0 AND date < ?
		 ORDER BY date ASC LIMIT 1`,
		goalID, before.Format(dateLayout),
	)

	var id, gID, dateStr, content, createdAt string
	var done int
	if err := row.Scan(&id, &gID, &dateStr, &content, &done, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
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

	task := entity.NewDailyTask(id, gID, parsedDate, content, parsedCreatedAt)
	task.Done = done != 0
	return task, nil
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
		`SELECT dt.id, dt.goal_id, dt.date, dt.content, dt.done, dt.created_at
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

func (r *DailyTaskRepository) FindByDate(date time.Time) ([]*entity.DailyTask, error) {
	rows, err := r.db.Query(
		`SELECT id, goal_id, date, content, done, created_at FROM daily_tasks WHERE date = ? ORDER BY created_at`,
		date.Format(dateLayout),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDailyTasks(rows)
}

func scanDailyTasks(rows *sql.Rows) ([]*entity.DailyTask, error) {
	tasks := []*entity.DailyTask{}
	for rows.Next() {
		var id, goalID, dateStr, content, createdAt string
		var done int
		if err := rows.Scan(&id, &goalID, &dateStr, &content, &done, &createdAt); err != nil {
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
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}
