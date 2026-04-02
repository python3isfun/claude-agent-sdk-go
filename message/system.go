package message

import "encoding/json"

// SystemMessageSubtype identifies the type of system message.
type SystemMessageSubtype string

const (
	SystemSubtypeTaskStarted      SystemMessageSubtype = "task_started"
	SystemSubtypeTaskProgress     SystemMessageSubtype = "task_progress"
	SystemSubtypeTaskNotification SystemMessageSubtype = "task_notification"
	SystemSubtypeInit             SystemMessageSubtype = "init"
)

// SystemMessage represents a system-level message.
type SystemMessage struct {
	Type_    MessageType          `json:"type"`
	Subtype_ SystemMessageSubtype `json:"subtype,omitempty"`
	Data_    json.RawMessage      `json:"data,omitempty"`
}

func (m *SystemMessage) Type() MessageType             { return MessageTypeSystem }
func (m *SystemMessage) Subtype() SystemMessageSubtype { return m.Subtype_ }
func (m *SystemMessage) RawData() json.RawMessage      { return m.Data_ }

// TaskStartedData returns the data for a task_started message.
func (m *SystemMessage) TaskStartedData() (*TaskStartedData, error) {
	if m.Subtype_ != SystemSubtypeTaskStarted {
		return nil, nil
	}
	var data TaskStartedData
	if err := json.Unmarshal(m.Data_, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// TaskProgressData returns the data for a task_progress message.
func (m *SystemMessage) TaskProgressData() (*TaskProgressData, error) {
	if m.Subtype_ != SystemSubtypeTaskProgress {
		return nil, nil
	}
	var data TaskProgressData
	if err := json.Unmarshal(m.Data_, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// TaskNotificationData returns the data for a task_notification message.
func (m *SystemMessage) TaskNotificationData() (*TaskNotificationData, error) {
	if m.Subtype_ != SystemSubtypeTaskNotification {
		return nil, nil
	}
	var data TaskNotificationData
	if err := json.Unmarshal(m.Data_, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// TaskStartedData contains information about a started task.
type TaskStartedData struct {
	TaskID      string `json:"task_id"`
	Description string `json:"description,omitempty"`
	UUID        string `json:"uuid,omitempty"`
	SessionID   string `json:"session_id,omitempty"`
}

// TaskProgressData contains progress information about a task.
type TaskProgressData struct {
	TaskID       string `json:"task_id"`
	LastToolName string `json:"last_tool_name,omitempty"`
	Usage        *Usage `json:"usage,omitempty"`
}

// TaskNotificationData contains notification about task status.
type TaskNotificationData struct {
	TaskID     string `json:"task_id"`
	Status     string `json:"status"` // "completed", "failed", "stopped"
	OutputFile string `json:"output_file,omitempty"`
	Summary    string `json:"summary,omitempty"`
	Usage      *Usage `json:"usage,omitempty"`
}
