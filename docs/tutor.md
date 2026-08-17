# Embedded AI tutor

## UX requirement

The tutor must be available from every screen through the application shell.

Desktop can use a side overlay/drawer. Tablet can use a wide drawer. Phone can use a full-screen overlay. These are presentation differences over the same global tutor state.

A dedicated tutor/history page may exist later, but it is not the primary access pattern.

## Semantic screen context

The tutor should understand the current screen through structured application state, not through routine screenshots.

Example page context:

```ts
type TutorPageContext = {
  type: 'dashboard' | 'lesson' | 'exercise' | 'review' | 'project'
  moduleId?: string
  lessonId?: string
  objectiveIds?: string[]
  sectionId?: string
  exerciseId?: string
  selectedText?: string
  code?: string
  lastExecution?: {
    passed: number
    failed: number
    summary?: string
  }
}
```

Do not send every possible field on every turn. Include data relevant to what the learner is asking about.

## Recent activity

The context builder may also retrieve a bounded slice of recent learner activity. This enables useful interpretations of vague questions such as:

> Why am I still not getting this?

when the system knows the learner has just failed the same convergence test three times.

## Tutor modes

Initial conceptual modes:

- **Explain** — direct teaching and explanation.
- **Socratic** — guide through questions rather than immediately supplying the answer.
- **Exercise help** — interpret failures and provide progressively stronger hints.
- **Quiz** — assess an objective interactively.
- **Explore** — allow broad tangents and connections beyond the immediate lesson.

These are policies over one tutor system, not separate autonomous agents.

## Exercise behavior

When invoked from an exercise, the tutor should prefer:

- questions;
- conceptual clarification;
- explanation of test failures;
- references back to the relevant lesson;
- incremental hints.

It should not casually dump the completed solution merely because it can. The learner may explicitly request a solution, but the product should make that a conscious choice.

## Provider boundary

OpenRouter and Codex have materially different execution models. Preserve those differences inside provider adapters.

The internal service consumes a normalized stream:

```go
type Event struct {
    Type     string `json:"type"`
    Text     string `json:"text,omitempty"`
    Tool     string `json:"tool,omitempty"`
    SourceID string `json:"sourceId,omitempty"`
    Error    string `json:"error,omitempty"`
}
```

Likely event types include:

- `text_delta`;
- `tool_started`;
- `tool_completed`;
- `citation`;
- `usage`;
- `completed`;
- `error`.

The UI should not need to understand OpenRouter SSE payloads or Codex JSON-RPC details.

## Grounding and citations

For curriculum questions, prefer curriculum sources and attach source IDs to supported claims when possible.

Never invent source metadata. The content source registry is authoritative.

For discussion beyond the curriculum, the tutor should distinguish sourced curriculum material from unsourced model knowledge unless a future web/research tool is deliberately added.

## Writes to learner state

Read operations can be automatic. Writes should be explicit and auditable.

Examples:

- "Save that as a note."
- "Make a review-card candidate from this."
- "Record that I completed the exercise."

The tutor may recommend a mastery check. It should not mark an objective mastered based only on conversational confidence.
