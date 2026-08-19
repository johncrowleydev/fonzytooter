package tutor

import (
	"context"
)

type EventType string

const (
	EventTextDelta     EventType = "text_delta"
	EventToolStarted   EventType = "tool_started"
	EventToolCompleted EventType = "tool_completed"
	EventCitation      EventType = "citation"
	EventUsage         EventType = "usage"
	EventCompleted     EventType = "completed"
	EventError         EventType = "error"
)

type Event struct {
	Type     EventType `json:"type" enum:"text_delta,tool_started,tool_completed,citation,usage,completed,error"`
	Text     string    `json:"text,omitempty"`
	Tool     string    `json:"tool,omitempty"`
	SourceID string    `json:"sourceId,omitempty"`
	Usage    *Usage    `json:"usage,omitempty"`
	Error    string    `json:"error,omitempty"`
}

type Usage struct {
	InputTokens  int `json:"inputTokens,omitempty" minimum:"0"`
	OutputTokens int `json:"outputTokens,omitempty" minimum:"0"`
	TotalTokens  int `json:"totalTokens,omitempty" minimum:"0"`
}

type TurnRequest struct {
	ConversationID string       `json:"conversationId,omitempty" maxLength:"200"`
	Message        string       `json:"message" minLength:"1" maxLength:"10000" pattern:"[^\\s]" patternDescription:"contain at least one non-whitespace character"`
	Mode           TutorMode    `json:"mode,omitempty" enum:"explain,socratic,exercise,quiz,explore"`
	PageContext    *PageContext `json:"pageContext,omitempty"`
}

type TutorMode string

type PageContext struct {
	Type           PageType   `json:"type" enum:"dashboard,curriculum,lesson,exercise,review,progress,project"`
	Title          string     `json:"title,omitempty" maxLength:"300"`
	CourseID       string     `json:"courseId,omitempty" maxLength:"200"`
	CourseTitle    string     `json:"courseTitle,omitempty" maxLength:"300"`
	ModuleID       string     `json:"moduleId,omitempty" maxLength:"200"`
	ModuleTitle    string     `json:"moduleTitle,omitempty" maxLength:"300"`
	LessonID       string     `json:"lessonId,omitempty" maxLength:"200"`
	LessonTitle    string     `json:"lessonTitle,omitempty" maxLength:"300"`
	ObjectiveIDs   []string   `json:"objectiveIds,omitempty" maxItems:"50"`
	SectionID      string     `json:"sectionId,omitempty" maxLength:"200"`
	ExerciseID     string     `json:"exerciseId,omitempty" maxLength:"200"`
	ExerciseTitle  string     `json:"exerciseTitle,omitempty" maxLength:"300"`
	SelectedText   string     `json:"selectedText,omitempty" maxLength:"10000"`
	Code           string     `json:"code,omitempty" maxLength:"30000"`
	LastExecution  *Execution `json:"lastExecution,omitempty"`
	ObjectiveID    string     `json:"objectiveId,omitempty" maxLength:"200"`
	ObjectiveTitle string     `json:"objectiveTitle,omitempty" maxLength:"300"`
	ProjectID      string     `json:"projectId,omitempty" maxLength:"200"`
	ProjectTitle   string     `json:"projectTitle,omitempty" maxLength:"300"`
}

type PageType string

type Execution struct {
	Passed  int    `json:"passed" minimum:"0"`
	Failed  int    `json:"failed" minimum:"0"`
	Summary string `json:"summary,omitempty" maxLength:"1000"`
}

type Provider interface {
	StreamTurn(ctx context.Context, request TurnRequest) (<-chan Event, error)
}

type Service struct {
	provider Provider
}

func NewService(provider Provider) *Service {
	return &Service{provider: provider}
}

func (s *Service) StreamTurn(ctx context.Context, request TurnRequest) (<-chan Event, error) {
	return s.provider.StreamTurn(ctx, request)
}
