package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/ASpooky/Work-Backend-From/src/handler/httpapi"
	infraai "github.com/ASpooky/Work-Backend-From/src/infra/ai"
	"github.com/ASpooky/Work-Backend-From/src/infra/clock"
	"github.com/ASpooky/Work-Backend-From/src/infra/idgen"
	"github.com/ASpooky/Work-Backend-From/src/repository/sqlite"
	"github.com/ASpooky/Work-Backend-From/src/usecase"
	"github.com/ASpooky/Work-Backend-From/src/usecase/ai"
	"github.com/ASpooky/Work-Backend-From/src/usecase/dailytask"
	"github.com/ASpooky/Work-Backend-From/src/usecase/goal"
	"github.com/ASpooky/Work-Backend-From/src/usecase/workspace"
)

const defaultGeminiModel = "gemini-3.7-flash"

func main() {
	loadEnvFiles(".env", "src/.env")

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
	renameWorkspace := workspace.NewRenameWorkspaceUsecase(workspaceRepo)
	deleteWorkspace := workspace.NewDeleteWorkspaceUsecase(workspaceRepo)
	createGoal := goal.NewCreateGoalUsecase(goalRepo, ids, clk)
	listGoals := goal.NewListGoalsUsecase(goalRepo)
	getGoalStats := usecase.NewGetGoalStatsUsecase(goalRepo, dailyTaskRepo, clk)
	updateGoal := goal.NewUpdateGoalUsecase(goalRepo)
	deleteGoal := goal.NewDeleteGoalUsecase(goalRepo)
	reorderGoal := goal.NewReorderGoalUsecase(goalRepo)
	createDailyTask := dailytask.NewCreateDailyTaskUsecase(dailyTaskRepo, ids, clk)
	createRecurringDailyTasks := dailytask.NewCreateRecurringDailyTasksUsecase(dailyTaskRepo, ids, clk)
	listDailyTasks := dailytask.NewListDailyTasksUsecase(dailyTaskRepo)
	updateDailyTaskDone := dailytask.NewUpdateDailyTaskDoneUsecase(dailyTaskRepo, clk)
	getCalendar := usecase.NewGetCalendarUsecase(goalRepo, dailyTaskRepo)

	if err := ensureDefaultWorkspace(createWorkspace, listWorkspaces); err != nil {
		log.Fatalf("failed to ensure default workspace: %v", err)
	}

	catchUpMissedTasks := usecase.NewCatchUpMissedTasksUsecase(goalRepo, goalRepo, dailyTaskRepo, dailyTaskRepo, clk)
	if err := catchUpAllWorkspaces(listWorkspaces, catchUpMissedTasks); err != nil {
		log.Fatalf("failed to catch up missed tasks: %v", err)
	}

	workspaceHandler := httpapi.NewWorkspaceHandler(createWorkspace, listWorkspaces, renameWorkspace, deleteWorkspace)
	goalHandler := httpapi.NewGoalHandler(createGoal, listGoals, getGoalStats, updateGoal, deleteGoal, reorderGoal)
	dailyTaskHandler := httpapi.NewDailyTaskHandler(createDailyTask, createRecurringDailyTasks, listDailyTasks, updateDailyTaskDone)
	calendarHandler := httpapi.NewCalendarHandler(getCalendar)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /workspaces", workspaceHandler.Create)
	mux.HandleFunc("GET /workspaces", workspaceHandler.List)
	mux.HandleFunc("PATCH /workspaces/{id}", workspaceHandler.Rename)
	mux.HandleFunc("DELETE /workspaces/{id}", workspaceHandler.Delete)
	mux.HandleFunc("POST /goals", goalHandler.Create)
	mux.HandleFunc("GET /goals", goalHandler.List)
	mux.HandleFunc("GET /goals/{id}", goalHandler.Get)
	mux.HandleFunc("PATCH /goals/{id}", goalHandler.Update)
	mux.HandleFunc("DELETE /goals/{id}", goalHandler.Delete)
	mux.HandleFunc("PATCH /goals/{id}/priority", goalHandler.Reorder)
	mux.HandleFunc("POST /daily-tasks", dailyTaskHandler.Create)
	mux.HandleFunc("POST /daily-tasks/recurring", dailyTaskHandler.CreateRecurring)
	mux.HandleFunc("GET /daily-tasks", dailyTaskHandler.List)
	mux.HandleFunc("PATCH /daily-tasks/{id}", dailyTaskHandler.UpdateDone)
	mux.HandleFunc("GET /calendar", calendarHandler.Get)

	if geminiAPIKey := os.Getenv("GEMINI_API_KEY"); geminiAPIKey != "" {
		model := os.Getenv("GEMINI_MODEL")
		if model == "" {
			model = defaultGeminiModel
		}
		geminiClient := infraai.NewGeminiClient(geminiAPIKey, model)
		conversationRepo := sqlite.NewConversationRepository(db)
		conversationMessageRepo := sqlite.NewConversationMessageRepository(db)

		chatUsecase := ai.NewChatUsecase(geminiClient, clk)
		sendMessage := ai.NewSendMessageUsecase(conversationRepo, conversationRepo, conversationMessageRepo, chatUsecase, ids, clk)
		listConversations := ai.NewListConversationsUsecase(conversationRepo)
		getConversation := ai.NewGetConversationUsecase(conversationMessageRepo)
		planGoal := ai.NewPlanGoalUsecase(geminiClient, conversationMessageRepo, dailyTaskRepo, clk)
		summarizeGoal := ai.NewSummarizeGoalUsecase(goalRepo, dailyTaskRepo, geminiClient, clk)
		goalReviewChat := ai.NewGoalReviewChatUsecase(geminiClient, goalRepo, dailyTaskRepo, dailyTaskRepo, clk)
		sendGoalReviewMessage := ai.NewSendMessageUsecase(conversationRepo, conversationRepo, conversationMessageRepo, goalReviewChat, ids, clk)
		listGoalConversations := ai.NewListGoalConversationsUsecase(conversationRepo)
		reviseGoal := ai.NewReviseGoalUsecase(geminiClient, goalRepo, conversationMessageRepo, clk)

		aiHandler := httpapi.NewAIHandler(
			sendMessage, listConversations, getConversation, planGoal, summarizeGoal,
			sendGoalReviewMessage, listGoalConversations, reviseGoal,
		)
		mux.HandleFunc("GET /ai/status", aiHandler.Status)
		mux.HandleFunc("POST /ai/chat", aiHandler.SendMessage)
		mux.HandleFunc("GET /ai/conversations", aiHandler.ListConversations)
		mux.HandleFunc("GET /ai/conversations/{id}/messages", aiHandler.GetConversation)
		mux.HandleFunc("POST /ai/plan", aiHandler.Plan)
		mux.HandleFunc("POST /ai/goals/{id}/summary", aiHandler.SummarizeGoal)
		mux.HandleFunc("POST /ai/goals/{id}/chat", aiHandler.GoalChat)
		mux.HandleFunc("GET /ai/goals/{id}/conversations", aiHandler.ListGoalConversations)
		mux.HandleFunc("POST /ai/goals/{id}/revise", aiHandler.ReviseGoal)
		log.Printf("AI features enabled (model=%s)", model)
	} else {
		log.Println("GEMINI_API_KEY not set; AI features disabled")
	}

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", httpapi.WithCORS(mux)))
}

// loadEnvFiles best-effort loads .env files from the given candidate paths
// (so `go run ./src` from the repo root and from src/ both work). A missing
// file is not an error; existing OS environment variables always win.
func loadEnvFiles(paths ...string) {
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if err := godotenv.Load(path); err != nil {
			log.Printf("warning: failed to load %s: %v", path, err)
			continue
		}
		log.Printf("loaded environment from %s", path)
	}
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
