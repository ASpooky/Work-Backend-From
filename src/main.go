package main

import (
	"log"
	"net/http"

	"github.com/ASpooky/Work-Backend-From/src/handler/httpapi"
	"github.com/ASpooky/Work-Backend-From/src/infra/clock"
	"github.com/ASpooky/Work-Backend-From/src/infra/idgen"
	"github.com/ASpooky/Work-Backend-From/src/repository/sqlite"
	"github.com/ASpooky/Work-Backend-From/src/usecase"
	"github.com/ASpooky/Work-Backend-From/src/usecase/dailytask"
	"github.com/ASpooky/Work-Backend-From/src/usecase/goal"
	"github.com/ASpooky/Work-Backend-From/src/usecase/workspace"
)

func main() {
	db, err := sqlite.Open("app.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	clk := clock.SystemClock{}
	ids := idgen.UUIDGenerator{}

	workspaceRepo := sqlite.NewWorkspaceRepository(db)
	goalRepo := sqlite.NewGoalRepository(db)
	dailyTaskRepo := sqlite.NewDailyTaskRepository(db)

	createWorkspace := workspace.NewCreateWorkspaceUsecase(workspaceRepo, ids, clk)
	listWorkspaces := workspace.NewListWorkspacesUsecase(workspaceRepo)
	createGoal := goal.NewCreateGoalUsecase(goalRepo, ids, clk)
	listGoals := goal.NewListGoalsUsecase(goalRepo)
	createDailyTask := dailytask.NewCreateDailyTaskUsecase(dailyTaskRepo, ids, clk)
	createRecurringDailyTasks := dailytask.NewCreateRecurringDailyTasksUsecase(dailyTaskRepo, ids, clk)
	listDailyTasks := dailytask.NewListDailyTasksUsecase(dailyTaskRepo)
	updateDailyTaskDone := dailytask.NewUpdateDailyTaskDoneUsecase(dailyTaskRepo)
	getCalendar := usecase.NewGetCalendarUsecase(goalRepo, dailyTaskRepo)

	if err := ensureDefaultWorkspace(createWorkspace, listWorkspaces); err != nil {
		log.Fatalf("failed to ensure default workspace: %v", err)
	}

	catchUpMissedTasks := usecase.NewCatchUpMissedTasksUsecase(goalRepo, goalRepo, dailyTaskRepo, dailyTaskRepo, clk)
	if err := catchUpAllWorkspaces(listWorkspaces, catchUpMissedTasks); err != nil {
		log.Fatalf("failed to catch up missed tasks: %v", err)
	}

	workspaceHandler := httpapi.NewWorkspaceHandler(createWorkspace, listWorkspaces)
	goalHandler := httpapi.NewGoalHandler(createGoal, listGoals)
	dailyTaskHandler := httpapi.NewDailyTaskHandler(createDailyTask, createRecurringDailyTasks, listDailyTasks, updateDailyTaskDone)
	calendarHandler := httpapi.NewCalendarHandler(getCalendar)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /workspaces", workspaceHandler.Create)
	mux.HandleFunc("GET /workspaces", workspaceHandler.List)
	mux.HandleFunc("POST /goals", goalHandler.Create)
	mux.HandleFunc("GET /goals", goalHandler.List)
	mux.HandleFunc("POST /daily-tasks", dailyTaskHandler.Create)
	mux.HandleFunc("POST /daily-tasks/recurring", dailyTaskHandler.CreateRecurring)
	mux.HandleFunc("GET /daily-tasks", dailyTaskHandler.List)
	mux.HandleFunc("PATCH /daily-tasks/{id}", dailyTaskHandler.UpdateDone)
	mux.HandleFunc("GET /calendar", calendarHandler.Get)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", httpapi.WithCORS(mux)))
}

func ensureDefaultWorkspace(create *workspace.CreateWorkspaceUsecase, list *workspace.ListWorkspacesUsecase) error {
	existing, err := list.Execute()
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	_, err = create.Execute(workspace.CreateWorkspaceInput{
		UserID: sqlite.DefaultUserID,
		Name:   "private",
	})
	return err
}

func catchUpAllWorkspaces(list *workspace.ListWorkspacesUsecase, catchUp *usecase.CatchUpMissedTasksUsecase) error {
	workspaces, err := list.Execute()
	if err != nil {
		return err
	}

	for _, ws := range workspaces {
		if err := catchUp.Execute(ws.ID); err != nil {
			return err
		}
	}
	return nil
}
