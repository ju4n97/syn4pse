package relic

import "time"

// MessageRole defines the role of a message in a chat conversation.
type MessageRole string

// Message roles for chat conversations.
const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleTool      MessageRole = "tool"
	MessageRoleDeveloper MessageRole = "developer"
)

// Message represents a single message in a chat conversation.
type Message struct {
	Role      MessageRole `json:"role"`
	Content   string      `json:"content"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewMessage creates a new message with the current timestamp.
func NewMessage(role MessageRole, content string) Message {
	return Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// NewSystemMessage creates a new system message.
func NewSystemMessage(content string) Message { return NewMessage(MessageRoleSystem, content) }

// NewUserMessage creates a new user message.
func NewUserMessage(content string) Message { return NewMessage(MessageRoleUser, content) }

// NewAssistantMessage creates a new assistant message.
func NewAssistantMessage(content string) Message { return NewMessage(MessageRoleAssistant, content) }

// NewToolMessage creates a new tool message.
func NewToolMessage(content string) Message { return NewMessage(MessageRoleTool, content) }

// NewDeveloperMessage creates a new developer message.
func NewDeveloperMessage(content string) Message { return NewMessage(MessageRoleDeveloper, content) }

// StreamChunk represents a chunk of a stream response.
type StreamChunk struct {
	Done    bool   `json:"done"`
	Content string `json:"content"`
	Error   error  `json:"error,omitempty"`
}
