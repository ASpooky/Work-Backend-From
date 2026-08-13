package entity

import "time"

type GoalMode string

const (
	ModeStrict GoalMode = "strict"
	ModeWant   GoalMode = "want"
)

type GoalStatus string

const (
	StatusActive    GoalStatus = "active"
	StatusAchieved  GoalStatus = "achieved"
	StatusAbandoned GoalStatus = "abandoned"
)

type Goal struct {
	ID                   string
	WorkspaceID          string
	Title                string
	Detail               string
	AchievementCondition string
	EndDate              time.Time
	Mode                 GoalMode
	Status               GoalStatus
	CreatedAt            time.Time
}

func NewGoal(id, workspaceID, title, detail, achievementCondition string, endDate time.Time, mode GoalMode, createdAt time.Time) *Goal {
	return &Goal{
		ID:                   id,
		WorkspaceID:          workspaceID,
		Title:                title,
		Detail:               detail,
		AchievementCondition: achievementCondition,
		EndDate:              endDate,
		Mode:                 mode,
		Status:               StatusActive,
		CreatedAt:            createdAt,
	}
}

// Postpone shifts EndDate by one day when a strict-mode goal misses a day.
// want-mode goals never shift, per entity.md's confirmed rules.
func (g *Goal) Postpone() {
	if g.Mode == ModeStrict {
		g.EndDate = g.EndDate.AddDate(0, 0, 1)
	}
}

// UpdateAchievement marks the goal achieved once every one of its daily
// tasks is done, per entity.md's confirmed rules.
func (g *Goal) UpdateAchievement(tasks []DailyTask) {
	if len(tasks) == 0 {
		return
	}

	for _, task := range tasks {
		if !task.Done {
			return
		}
	}

	g.Status = StatusAchieved
}
