// Package message defines the message types used in the Claude Agent SDK.
package message

import "encoding/json"

// MessageType identifies the type of message.
type MessageType string

const (
	MessageTypeAssistant MessageType = "assistant"
	MessageTypeUser      MessageType = "user"
	MessageTypeSystem    MessageType = "system"
	MessageTypeResult    MessageType = "result"
)

// BlockType identifies the type of content block.
type BlockType string

const (
	BlockTypeText       BlockType = "text"
	BlockTypeToolUse    BlockType = "tool_use"
	BlockTypeToolResult BlockType = "tool_result"
	BlockTypeThinking   BlockType = "thinking"
	BlockTypeImage      BlockType = "image"
)

// Message is the base interface for all message types.
type Message interface {
	// Type returns the message type.
	Type() MessageType
}

// ContentBlock is the base interface for content blocks.
type ContentBlock interface {
	// BlockType returns the type of this content block.
	BlockType() BlockType
}

// Usage represents token usage information.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	CacheRead    int `json:"cache_read_input_tokens,omitempty"`
	CacheWrite   int `json:"cache_creation_input_tokens,omitempty"`
}

// RawMessage is used for parsing messages before type determination.
type RawMessage struct {
	Type    string          `json:"type"`
	Subtype string          `json:"subtype,omitempty"`
	Message json.RawMessage `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// ParseMessage parses a raw JSON message into the appropriate Message type.
func ParseMessage(data []byte) (Message, error) {
	var raw RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	switch MessageType(raw.Type) {
	case MessageTypeAssistant:
		var msg AssistantMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		return &msg, nil

	case MessageTypeUser:
		var msg UserMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		return &msg, nil

	case MessageTypeSystem:
		var msg SystemMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		return &msg, nil

	case MessageTypeResult:
		var msg ResultMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		return &msg, nil

	default:
		// Return unknown message type for forward compatibility
		return &UnknownMessage{RawType: raw.Type, RawData: data}, nil
	}
}

// UnknownMessage represents a message type that is not recognized.
type UnknownMessage struct {
	RawType string
	RawData []byte
}

func (m *UnknownMessage) Type() MessageType { return MessageType(m.RawType) }
