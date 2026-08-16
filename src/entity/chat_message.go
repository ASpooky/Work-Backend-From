package entity

type ChatRole string

const (
	ChatRoleUser  ChatRole = "user"
	ChatRoleModel ChatRole = "model"
)

// ChatMessage is one turn in an AI planning conversation. It carries no
// business rules and is never persisted — it exists here (rather than in a
// usecase subpackage) purely so both usecase/ai and infra/ai can share the
// same type without infra depending on a usecase package.
type ChatMessage struct {
	Role    ChatRole `json:"role"`
	Content string   `json:"content"`
}
