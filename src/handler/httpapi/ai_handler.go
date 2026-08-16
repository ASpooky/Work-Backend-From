package httpapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ASpooky/Work-Backend-From/src/usecase/ai"
)

type AIHandler struct {
	sendMessage           *ai.SendMessageUsecase
	listConversations     *ai.ListConversationsUsecase
	getConversation       *ai.GetConversationUsecase
	plan                  *ai.PlanGoalUsecase
	summarizeGoal         *ai.SummarizeGoalUsecase
	sendGoalReviewMessage *ai.SendMessageUsecase
	listGoalConversations *ai.ListGoalConversationsUsecase
	reviseGoal            *ai.ReviseGoalUsecase
}

func NewAIHandler(
	sendMessage *ai.SendMessageUsecase,
	listConversations *ai.ListConversationsUsecase,
	getConversation *ai.GetConversationUsecase,
	plan *ai.PlanGoalUsecase,
	summarizeGoal *ai.SummarizeGoalUsecase,
	sendGoalReviewMessage *ai.SendMessageUsecase,
	listGoalConversations *ai.ListGoalConversationsUsecase,
	reviseGoal *ai.ReviseGoalUsecase,
) *AIHandler {
	return &AIHandler{
		sendMessage:           sendMessage,
		listConversations:     listConversations,
		getConversation:       getConversation,
		plan:                  plan,
		summarizeGoal:         summarizeGoal,
		sendGoalReviewMessage: sendGoalReviewMessage,
		listGoalConversations: listGoalConversations,
		reviseGoal:            reviseGoal,
	}
}

func (h *AIHandler) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": true})
}

type sendMessageRequest struct {
	WorkspaceID    string `json:"workspace_id"`
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content"`
}

func (h *AIHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	got, err := h.sendMessage.Execute(r.Context(), ai.SendMessageInput{
		WorkspaceID:    req.WorkspaceID,
		ConversationID: req.ConversationID,
		Content:        req.Content,
	})
	if err != nil {
		log.Printf("AI send message failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"conversation_id": got.ConversationID,
		"reply":           got.Reply,
	})
}

func (h *AIHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")

	got, err := h.listConversations.Execute(workspaceID)
	if err != nil {
		log.Printf("AI list conversations failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, got)
}

func (h *AIHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	got, err := h.getConversation.Execute(id)
	if err != nil {
		log.Printf("AI get conversation failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, got)
}

type plannedGoalResponse struct {
	Title                string `json:"title"`
	Detail               string `json:"detail"`
	AchievementCondition string `json:"achievement_condition"`
	EndDate              string `json:"end_date"`
	Mode                 string `json:"mode"`
}

type plannedTaskResponse struct {
	Content      string `json:"content"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	RuleType     string `json:"rule_type"`
	IntervalDays int    `json:"interval_days,omitempty"`
	Weekdays     []int  `json:"weekdays,omitempty"`
}

type planResponse struct {
	Goal  plannedGoalResponse   `json:"goal"`
	Tasks []plannedTaskResponse `json:"tasks"`
}

type planRequest struct {
	ConversationID string `json:"conversation_id"`
}

func (h *AIHandler) Plan(w http.ResponseWriter, r *http.Request) {
	var req planRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	plan, err := h.plan.Execute(r.Context(), ai.PlanGoalInput{ConversationID: req.ConversationID})
	if err != nil {
		log.Printf("AI plan generation failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := planResponse{
		Goal: plannedGoalResponse{
			Title:                plan.Goal.Title,
			Detail:               plan.Goal.Detail,
			AchievementCondition: plan.Goal.AchievementCondition,
			EndDate:              plan.Goal.EndDate.Format(dateLayout),
			Mode:                 string(plan.Goal.Mode),
		},
		Tasks: []plannedTaskResponse{},
	}
	for _, t := range plan.Tasks {
		tr := plannedTaskResponse{
			Content:      t.Content,
			StartDate:    t.StartDate.Format(dateLayout),
			EndDate:      t.EndDate.Format(dateLayout),
			RuleType:     string(t.Rule.Type),
			IntervalDays: t.Rule.IntervalDays,
		}
		for _, wd := range t.Rule.Weekdays {
			tr.Weekdays = append(tr.Weekdays, int(wd))
		}
		resp.Tasks = append(resp.Tasks, tr)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *AIHandler) SummarizeGoal(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	got, err := h.summarizeGoal.Execute(r.Context(), ai.GoalSummaryInput{GoalID: id})
	if err != nil {
		log.Printf("AI goal summary failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"summary": got.Summary})
}

func (h *AIHandler) GoalChat(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	got, err := h.sendGoalReviewMessage.Execute(r.Context(), ai.SendMessageInput{
		WorkspaceID:    req.WorkspaceID,
		GoalID:         id,
		ConversationID: req.ConversationID,
		Content:        req.Content,
	})
	if err != nil {
		log.Printf("AI goal review chat failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"conversation_id": got.ConversationID,
		"reply":           got.Reply,
	})
}

func (h *AIHandler) ListGoalConversations(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	got, err := h.listGoalConversations.Execute(id)
	if err != nil {
		log.Printf("AI list goal conversations failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, got)
}

type reviseGoalRequest struct {
	ConversationID string `json:"conversation_id"`
}

func (h *AIHandler) ReviseGoal(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req reviseGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	revised, err := h.reviseGoal.Execute(r.Context(), ai.ReviseGoalInput{GoalID: id, ConversationID: req.ConversationID})
	if err != nil {
		log.Printf("AI goal revision failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, plannedGoalResponse{
		Title:                revised.Title,
		Detail:               revised.Detail,
		AchievementCondition: revised.AchievementCondition,
		EndDate:              revised.EndDate.Format(dateLayout),
		Mode:                 string(revised.Mode),
	})
}
