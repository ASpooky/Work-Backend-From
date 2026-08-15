package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/repository/sqlite"
)

const dateLayout = "2006-01-02"

type seedGoal struct {
	title, detail, condition string
	mode                     entity.GoalMode
	doneRate                 float64 // probability a scheduled past/today day is marked done
	skipRate                 float64 // probability a day has no task at all
}

func main() {
	dbPath := flag.String("db", "app.db", "path to the sqlite database file")
	flag.Parse()

	db, err := sqlite.Open(*dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	workspaceRepo := sqlite.NewWorkspaceRepository(db)
	goalRepo := sqlite.NewGoalRepository(db)
	taskRepo := sqlite.NewDailyTaskRepository(db)

	workspaces, err := workspaceRepo.FindAll()
	if err != nil {
		log.Fatalf("failed to list workspaces: %v", err)
	}

	var ws *entity.WorkSpace
	if len(workspaces) == 0 {
		ws = entity.NewWorkSpace(uuid.NewString(), sqlite.DefaultUserID, "private", time.Now())
		if err := workspaceRepo.Save(ws); err != nil {
			log.Fatalf("failed to create default workspace: %v", err)
		}
	} else {
		ws = workspaces[0]
	}

	existingGoals, err := goalRepo.FindByWorkspaceID(ws.ID)
	if err != nil {
		log.Fatalf("failed to list goals: %v", err)
	}
	byTitle := make(map[string]*entity.Goal, len(existingGoals))
	for _, g := range existingGoals {
		byTitle[g.Title] = g
	}

	seeds := []seedGoal{
		{"Run 5km", "Keep up a running habit", "Run at least 5km", entity.ModeStrict, 0.8, 0.05},
		{"Read 30 minutes", "Read every day before bed", "Read for 30+ minutes", entity.ModeStrict, 0.6, 0.15},
		{"Learn Go", "Study Go in free time", "Write or read Go code", entity.ModeWant, 0.5, 0.35},
	}

	rng := rand.New(rand.NewSource(42))
	now := time.Now()
	from := now.AddDate(0, -3, 0)
	to := now.AddDate(0, 1, 0)

	for _, s := range seeds {
		goal, ok := byTitle[s.title]
		if !ok {
			goal = entity.NewGoal(uuid.NewString(), ws.ID, s.title, s.detail, s.condition, to, s.mode, from)
			if err := goalRepo.Save(goal); err != nil {
				log.Fatalf("failed to save goal %q: %v", s.title, err)
			}
			log.Printf("created goal %q", s.title)
		}

		existingTasks, err := taskRepo.FindByGoalIDAndDateRange(goal.ID, from, to)
		if err != nil {
			log.Fatalf("failed to list existing tasks for %q: %v", s.title, err)
		}
		hasTask := make(map[string]bool, len(existingTasks))
		for _, t := range existingTasks {
			hasTask[t.Date.Format(dateLayout)] = true
		}

		added := 0
		for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
			key := d.Format(dateLayout)
			if hasTask[key] {
				continue
			}
			if rng.Float64() < s.skipRate {
				continue
			}

			task := entity.NewDailyTask(uuid.NewString(), goal.ID, d, s.title, d)
			if !d.After(now) {
				task.Done = rng.Float64() < s.doneRate
			}
			if err := taskRepo.Save(task); err != nil {
				log.Fatalf("failed to save daily task for %q on %s: %v", s.title, key, err)
			}
			added++
		}

		log.Printf("goal %q: added %d daily task(s) (existing left untouched)", s.title, added)
	}

	log.Printf("seed complete: workspace=%s from=%s to=%s (re-run anytime to top up new days)", ws.Name, from.Format(dateLayout), to.Format(dateLayout))
}
