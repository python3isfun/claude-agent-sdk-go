package message

import "encoding/json"

// TextBlock represents text content.
type TextBlock struct {
	Type_ BlockType `json:"type"`
	Text_ string    `json:"text"`
}

func (b *TextBlock) BlockType() BlockType { return BlockTypeText }
func (b *TextBlock) Text() string         { return b.Text_ }

// ToolUseBlock represents a tool invocation request.
type ToolUseBlock struct {
	Type_  BlockType              `json:"type"`
	ID_    string                 `json:"id"`
	Name_  string                 `json:"name"`
	Input_ map[string]interface{} `json:"input"`
}

func (b *ToolUseBlock) BlockType() BlockType          { return BlockTypeToolUse }
func (b *ToolUseBlock) ID() string                    { return b.ID_ }
func (b *ToolUseBlock) Name() string                  { return b.Name_ }
func (b *ToolUseBlock) Input() map[string]interface{} { return b.Input_ }

// ToolResultBlock represents a tool execution result.
type ToolResultBlock struct {
	Type_      BlockType      `json:"type"`
	ToolUseID_ string         `json:"tool_use_id"`
	Content_   []ContentBlock `json:"content,omitempty"`
	IsError_   bool           `json:"is_error,omitempty"`
}

func (b *ToolResultBlock) BlockType() BlockType    { return BlockTypeToolResult }
func (b *ToolResultBlock) ToolUseID() string       { return b.ToolUseID_ }
func (b *ToolResultBlock) Content() []ContentBlock { return b.Content_ }
func (b *ToolResultBlock) IsError() bool           { return b.IsError_ }

// ThinkingBlock represents Claude's internal reasoning.
type ThinkingBlock struct {
	Type_      BlockType `json:"type"`
	Thinking_  string    `json:"thinking"`
	Signature_ string    `json:"signature,omitempty"`
}

func (b *ThinkingBlock) BlockType() BlockType { return BlockTypeThinking }
func (b *ThinkingBlock) Thinking() string     { return b.Thinking_ }
func (b *ThinkingBlock) Signature() string    { return b.Signature_ }

// ImageBlock represents image content.
type ImageBlock struct {
	Type_   BlockType   `json:"type"`
	Source_ ImageSource `json:"source"`
}

func (b *ImageBlock) BlockType() BlockType { return BlockTypeImage }
func (b *ImageBlock) Source() ImageSource  { return b.Source_ }

// ImageSource contains image data.
type ImageSource struct {
	Type      string `json:"type"` // "base64" or "url"
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
}

// rawContentBlock is used for parsing content blocks.
type rawContentBlock struct {
	Type string `json:"type"`
}

// ParseContentBlock parses a raw JSON content block into the appropriate type.
func ParseContentBlock(data json.RawMessage) (ContentBlock, error) {
	var raw rawContentBlock
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	switch BlockType(raw.Type) {
	case BlockTypeText:
		var block TextBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, err
		}
		return &block, nil

	case BlockTypeToolUse:
		var block ToolUseBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, err
		}
		return &block, nil

	case BlockTypeToolResult:
		var block ToolResultBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, err
		}
		return &block, nil

	case BlockTypeThinking:
		var block ThinkingBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, err
		}
		return &block, nil

	case BlockTypeImage:
		var block ImageBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, err
		}
		return &block, nil

	default:
		// Return text block as fallback for unknown types
		return &TextBlock{Type_: BlockType(raw.Type)}, nil
	}
}

// ParseContentBlocks parses an array of raw JSON content blocks.
func ParseContentBlocks(data []json.RawMessage) ([]ContentBlock, error) {
	blocks := make([]ContentBlock, 0, len(data))
	for _, raw := range data {
		block, err := ParseContentBlock(raw)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}
