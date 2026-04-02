package message

import (
	"encoding/json"
	"sync"
)

// AssistantMessage represents a response from Claude.
type AssistantMessage struct {
	Type_      MessageType       `json:"type"`
	ID_        string            `json:"id,omitempty"`
	Content_   []ContentBlock    `json:"-"` // Parsed separately
	RawContent []json.RawMessage `json:"content,omitempty"`
	Model_     string            `json:"model,omitempty"`
	Usage_     *Usage            `json:"usage,omitempty"`
	StopReason string            `json:"stop_reason,omitempty"`
	SessionID  string            `json:"session_id,omitempty"`
	MessageID  string            `json:"message_id,omitempty"`

	contentOnce sync.Once // ensures Content_ is parsed only once
}

func (m *AssistantMessage) Type() MessageType { return MessageTypeAssistant }
func (m *AssistantMessage) ID() string        { return m.ID_ }
func (m *AssistantMessage) Model() string     { return m.Model_ }
func (m *AssistantMessage) Usage() *Usage     { return m.Usage_ }

// Content returns the parsed content blocks.
// If not yet parsed, it parses the raw content.
// This method is thread-safe.
func (m *AssistantMessage) Content() []ContentBlock {
	m.contentOnce.Do(func() {
		if m.Content_ == nil && len(m.RawContent) > 0 {
			blocks, _ := ParseContentBlocks(m.RawContent)
			m.Content_ = blocks
		}
	})
	return m.Content_
}

// TextContent returns all text content concatenated.
func (m *AssistantMessage) TextContent() string {
	var result string
	for _, block := range m.Content() {
		if text, ok := block.(*TextBlock); ok {
			result += text.Text()
		}
	}
	return result
}

// ToolUses returns all tool use blocks in this message.
func (m *AssistantMessage) ToolUses() []*ToolUseBlock {
	var tools []*ToolUseBlock
	for _, block := range m.Content() {
		if tool, ok := block.(*ToolUseBlock); ok {
			tools = append(tools, tool)
		}
	}
	return tools
}

// HasToolUse returns true if this message contains any tool use blocks.
func (m *AssistantMessage) HasToolUse() bool {
	for _, block := range m.Content() {
		if _, ok := block.(*ToolUseBlock); ok {
			return true
		}
	}
	return false
}
