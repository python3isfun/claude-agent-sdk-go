package message

import "encoding/json"

// UserMessage represents a message from the user.
type UserMessage struct {
	Type_           MessageType       `json:"type"`
	UUID_           string            `json:"uuid,omitempty"`
	Content_        []ContentBlock    `json:"-"`
	RawContent      []json.RawMessage `json:"content,omitempty"`
	ParentToolUseID string            `json:"parent_tool_use_id,omitempty"`
	ToolUseResult   *ToolResultInfo   `json:"tool_use_result,omitempty"`
}

func (m *UserMessage) Type() MessageType { return MessageTypeUser }
func (m *UserMessage) UUID() string      { return m.UUID_ }

// Content returns the parsed content blocks.
func (m *UserMessage) Content() []ContentBlock {
	if m.Content_ == nil && len(m.RawContent) > 0 {
		blocks, _ := ParseContentBlocks(m.RawContent)
		m.Content_ = blocks
	}
	return m.Content_
}

// TextContent returns all text content concatenated.
func (m *UserMessage) TextContent() string {
	var result string
	for _, block := range m.Content() {
		if text, ok := block.(*TextBlock); ok {
			result += text.Text()
		}
	}
	return result
}

// ToolResultInfo contains information about a tool result.
type ToolResultInfo struct {
	ToolUseID string `json:"tool_use_id"`
	IsError   bool   `json:"is_error,omitempty"`
}

// NewUserMessage creates a new user message with text content.
func NewUserMessage(text string) *UserMessage {
	return &UserMessage{
		Type_: MessageTypeUser,
		Content_: []ContentBlock{
			&TextBlock{Type_: BlockTypeText, Text_: text},
		},
	}
}

// NewUserMessageWithToolResult creates a user message containing a tool result.
func NewUserMessageWithToolResult(toolUseID string, content []ContentBlock, isError bool) *UserMessage {
	return &UserMessage{
		Type_: MessageTypeUser,
		Content_: []ContentBlock{
			&ToolResultBlock{
				Type_:      BlockTypeToolResult,
				ToolUseID_: toolUseID,
				Content_:   content,
				IsError_:   isError,
			},
		},
	}
}
