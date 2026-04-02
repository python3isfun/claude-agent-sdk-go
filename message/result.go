package message

import "encoding/json"

// ResultMessage represents the final result of a query or session.
type ResultMessage struct {
	Type_            MessageType     `json:"type"`
	DurationMS       int64           `json:"duration_ms,omitempty"`
	DurationAPIMS    int64           `json:"duration_api_ms,omitempty"`
	Cost             float64         `json:"cost,omitempty"`
	Usage_           *Usage          `json:"usage,omitempty"`
	StopReason       string          `json:"stop_reason,omitempty"`
	StructuredOutput json.RawMessage `json:"structured_output,omitempty"`
	SessionID        string          `json:"session_id,omitempty"`
	NumTurns         int             `json:"num_turns,omitempty"`
}

func (m *ResultMessage) Type() MessageType { return MessageTypeResult }
func (m *ResultMessage) Usage() *Usage     { return m.Usage_ }

// Duration returns the total duration in milliseconds.
func (m *ResultMessage) Duration() int64 {
	return m.DurationMS
}

// APIDuration returns the API-specific duration in milliseconds.
func (m *ResultMessage) APIDuration() int64 {
	return m.DurationAPIMS
}

// ParseStructuredOutput unmarshals the structured output into the given value.
func (m *ResultMessage) ParseStructuredOutput(v interface{}) error {
	if m.StructuredOutput == nil {
		return nil
	}
	return json.Unmarshal(m.StructuredOutput, v)
}

// IsSuccess returns true if the query completed successfully.
func (m *ResultMessage) IsSuccess() bool {
	return m.StopReason == "" || m.StopReason == "end_turn" || m.StopReason == "stop_sequence"
}

// TotalTokens returns the total number of tokens used.
func (m *ResultMessage) TotalTokens() int {
	if m.Usage_ == nil {
		return 0
	}
	return m.Usage_.InputTokens + m.Usage_.OutputTokens
}
