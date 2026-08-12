package main

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

type User struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

type WorkSpace struct {
	ID        string
	UserID    string
	Name      string
	CreatedAt time.Time
}

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

type DailyTask struct {
	ID        string
	GoalID    string
	Date      time.Time
	Content   string
	Done      bool
	CreatedAt time.Time
}
