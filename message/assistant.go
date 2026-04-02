package message

import (
	"encoding/json"
	"sync"
)

// AssistantMessageInner represents the inner message structure from Claude CLI.
type AssistantMessageInner struct {
	Model      string            `json:"model,omitempty"`
	ID         string            `json:"id,omitempty"`
	Type       string            `json:"type,omitempty"`
	Role       string            `json:"role,omitempty"`
	Content    []json.RawMessage `json:"content,omitempty"`
	StopReason string            `json:"stop_reason,omitempty"`
	Usage      *Usage            `json:"usage,omitempty"`
}

// AssistantMessage represents a response from Claude.
type AssistantMessage struct {
	Type_     MessageType            `json:"type"`
	Message_  *AssistantMessageInner `json:"message,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	UUID      string                 `json:"uuid,omitempty"`

	Content_    []ContentBlock `json:"-"` // Parsed content blocks
	contentOnce sync.Once      // ensures Content_ is parsed only once
}

func (m *AssistantMessage) Type() MessageType { return MessageTypeAssistant }

func (m *AssistantMessage) ID() string {
	if m.Message_ != nil {
		return m.Message_.ID
	}
	return ""
}

func (m *AssistantMessage) Model() string {
	if m.Message_ != nil {
		return m.Message_.Model
	}
	return ""
}

func (m *AssistantMessage) Usage() *Usage {
	if m.Message_ != nil {
		return m.Message_.Usage
	}
	return nil
}

// Content returns the parsed content blocks.
// If not yet parsed, it parses the raw content.
// This method is thread-safe.
func (m *AssistantMessage) Content() []ContentBlock {
	m.contentOnce.Do(func() {
		if m.Content_ == nil && m.Message_ != nil && len(m.Message_.Content) > 0 {
			blocks, _ := ParseContentBlocks(m.Message_.Content)
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
