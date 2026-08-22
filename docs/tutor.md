# Embedded AI tutor

## Purpose

The embedded tutor is the next major product subsystem after the core learn, practice, and recall flows. It should make Fonzytooter feel like a learning system rather than a collection of course pages by helping the learner explain, reflect, recover from mistakes, and decide what to do next.

The tutor is not intended to be an unconstrained autonomous agent. Fonzytooter should continue to prefer deterministic application logic wherever ordinary software can make a reliable decision. The model supplies language understanding, explanation, multimodal interpretation, and judgment where those capabilities are actually useful.

## UX requirement

The tutor must be available from every screen through the application shell.

Desktop can use a side overlay/drawer. Tablet can use a wide drawer. Phone can use a full-screen overlay. These are presentation differences over the same global tutor state.

A dedicated tutor/history page may exist later, but it is not the primary access pattern.

## Initial implementation decision

Build the tutor around a small Go-native agent harness owned by Fonzytooter.

Do not introduce LangChain, LangGraph, or another general agent framework for the initial implementation. The required orchestration is intentionally straightforward and is easier to understand, test, and evolve as ordinary Go code.

The initial runtime should contain these responsibilities:

```text
TutorService
    |
    +-- ContextBuilder
    |     +-- current page and selected content
    |     +-- current lesson and objectives
    |     +-- relevant learner evidence
    |     +-- bounded conversation history
    |
    +-- ContextManager
    |     +-- provider-independent token budgeting
    |     +-- recent verbatim conversation tail
    |     +-- persistent compacted conversation memory
    |     +-- automatic pre-limit compaction
    |     +-- output and tool-call headroom
    |
    +-- ToolRegistry
    |     +-- typed definitions
    |     +-- JSON schemas
    |     +-- argument validation
    |     +-- execution
    |
    +-- AgentLoop
    |     +-- model call
    |     +-- zero or more tool calls
    |     +-- tool results
    |     +-- bounded iteration
    |
    +-- ConversationStore
    |
    +-- Provider
          +-- OpenRouter first
          +-- additional adapters later
```

The agent loop must be bounded. A model must not be allowed to recurse through tools indefinitely. The exact production limit should be chosen from observed behavior and evaluation results rather than treated as a model-specific constant.

## Provider strategy

OpenRouter is the first provider integration. It gives the project one consistent API surface for comparing multiple model families while preserving the option to add direct provider adapters later.

Provider-specific details must stay inside provider adapters. The rest of the tutor should not depend on OpenRouter model payloads, reasoning formats, or streaming wire details.

The provider boundary is responsible for:

- serializing canonical conversation messages and multimodal content parts;
- sending tool definitions;
- decoding streamed text and tool calls;
- preserving provider-required reasoning state across turns when necessary;
- reporting usage, latency-relevant metadata, and errors;
- mapping a provider's reasoning controls onto Fonzytooter's internal reasoning policy.

The tutor service is responsible for:

- conversation persistence;
- context construction;
- deciding which tools are available for a turn;
- executing tools requested by the model;
- enforcing iteration and permission limits;
- citations and source provenance;
- tutor-mode policy.

## Model selection

Do not choose the production tutor model before the production harness exists. Models should be evaluated through the same Fonzytooter context, tools, prompts, conversation persistence, and multimodal message format that the shipped tutor will use.

The initial candidate set is:

| Model | OpenRouter model ID | Reason for inclusion |
| --- | --- | --- |
| MiniMax M3 | `minimax/minimax-m3` | Strong cost/capability profile, multimodal support, tool use, and promising real-world coding/agent performance. |
| Kimi K2.6 | `moonshotai/kimi-k2.6` | Strong conversational reputation plus promising coding, reasoning, multimodal, and tool-use capability. |
| Gemini 3.7 Flash | `google/gemini-3.7-flash` | Strong multimodal and agentic baseline with configurable reasoning and broad provider support. |
| Qwen3.8 27B | `qwen/qwen3.8-27b` | Current compact Qwen vision-language model with coding, reasoning, and agentic capability. |

This list is an evaluation set, not a commitment to any model family. Exact model IDs should be pinned for reproducible evaluation and production behavior; do not use moving `latest` aliases for the shipped tutor.

Prices, discounts, provider availability, latency, and routing quality change too quickly to encode as architectural constants. Each evaluation run should record the actual model, provider/routing mode, token usage, latency, and cost observed at that time.

The production model does not need to be a frontier flagship. The target is the least expensive model that meets Fonzytooter's quality bar for technical accuracy, pedagogy, conversation, vision, and tool use.

See `docs/tutor-evaluation.md` for the initial evaluation dimensions and scenarios.

## Reasoning policy

Do not run maximum reasoning for every tutor message.

The internal request model should expose a provider-neutral reasoning policy such as none/low/medium/high, or an equivalent abstraction that can be mapped cleanly onto supported providers. Simple conversational questions should favor low latency and low cost. Difficult derivations, ambiguous worksheet work, or multi-step debugging may justify deeper reasoning.

The evaluation should measure both quality and the cost/latency effects of reasoning settings.

## Conversation ownership and persistence

Fonzytooter owns conversation state. Do not make a provider thread ID the authoritative record of a tutor conversation.

Persist the canonical conversation in SQLite so that providers can be changed without losing history or coupling the application to one vendor's state model. The exact schema may evolve, but it will likely need explicit records for conversations, messages, tool calls/results, and later attachments.

Inference requests should use bounded recent history plus deliberately selected older context or summaries when needed. Do not blindly resend an unbounded conversation forever merely because a model advertises a very large context window.

## Context management and automatic compaction

Context management is a first-class runtime responsibility. Every model request is assembled within a provider-independent budget that explicitly accounts for the system/tutor policy, compacted memory, recent verbatim messages, fresh application context, deterministic curriculum or learner context, tool definitions, and the current user message. The runtime reserves output and tool-call headroom rather than filling the provider's entire advertised context window with input.

The system/tutor policy and current user message are always retained. A bounded recent conversation tail remains verbatim. When older unsummarized history approaches a configured safety threshold—or grows beyond that retained tail—the runtime compacts it before the hard context limit and persists both the resulting memory and the exact message sequence summarized through. Repeated compaction advances that marker so transcript content is neither summarized twice nor included both verbatim and through memory.

Compacted memory should preserve salient learning state such as the learner's current goal, established understanding, known misconceptions, corrections already made, explanations that did not work, unresolved questions, active longer-lived task context, relevant tool findings, and source or citation IDs. The representation remains bounded and versioned so it can evolve without becoming an opaque replacement for the canonical transcript.

Production compaction uses the configured model to produce a strictly decoded structured-memory document. A bounded deterministic compactor remains the failure fallback and the test implementation; it is not the normal semantic-memory path. If the ordinary recent verbatim tail is still too large, context preparation progressively compacts more of its oldest messages while always retaining the current user message, stopping only when the request fits or the irreducible current request exceeds the hard budget.

Current page state is different: it is fresh, ephemeral input rebuilt from application state on every turn. Route identifiers, selected text, current code, and execution results must not become durable truth merely because they appeared in an older request, and stale page state must not be reintroduced through compaction memory.

Token estimation is an interface rather than a commitment to one tokenizer or model family. Providers may later supply more accurate model-aware estimates and context limits, while the runtime retains responsibility for enforcing its safety threshold and reserved response/tool capacity.

## Canonical multimodal messages

The tutor needs text and vision. No other input modality is required for the initial product.

Represent user/model input internally as provider-neutral content parts, for example:

```text
text
image
document
```

The provider adapter converts these parts into the provider's required request shape.

Allow text and vision model configuration to diverge later if evaluation shows a meaningful advantage. The initial configuration may use one model for both, but the architecture should not require that forever.

Audio, speech, and other modalities are not initial requirements.

## Semantic screen context

The tutor should understand the current screen through structured application state, not through routine screenshots.

Example page context:

```ts
type TutorPageContext = {
  type: 'dashboard' | 'curriculum' | 'lesson' | 'exercise' | 'review' | 'project'
  courseId?: string
  courseTitle?: string
  moduleId?: string
  moduleTitle?: string
  lessonId?: string
  lessonTitle?: string
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

## Context before tools

Information the application already knows and expects the model to need should generally be injected as context rather than exposed as a tool the model must discover and call.

For example, when the learner asks for help from an exercise page, Fonzytooter already knows the exercise ID, current code, latest execution result, objective IDs, and relevant lesson. The model should not need to call tools such as `get_current_page` merely to reconstruct state the application already possesses.

This is both cheaper and more reliable than forcing the model to retrieve obvious context through tool calls.

## Retrieval strategy

Do not begin with generic RAG infrastructure or a vector database.

The curriculum already contains strong deterministic structure:

```text
course
  module
    lesson
      objectives
      worksheets
      exercises
      reviews
      sources
```

The current route, objective IDs, lesson relationships, and prerequisite relationships should be used first to construct relevant context. Search is useful when the learner asks about material outside the immediate context or refers vaguely to something elsewhere in the course.

Semantic/vector retrieval can be added later if real usage demonstrates that structured retrieval is insufficient.

## Initial read-only toolset

Keep the first toolset intentionally small. Every additional tool increases the model's decision surface and creates another opportunity for incorrect or unnecessary calls.

The initial set should cover approximately these capabilities:

### `search_curriculum`

Search authored curriculum metadata/content for material relevant to a query when the current page context is insufficient.

### `get_curriculum_content`

Retrieve authoritative authored content for a known course/module/lesson/section or another supported curriculum reference.

### `get_objective_state`

Return factual learner evidence for one or more objectives, including introduction, recall evidence, application evidence, and relevant prerequisite state where available.

### `get_recent_learning_activity`

Return a bounded recent activity window for questions whose meaning depends on recent behavior or repeated struggle.

### `get_exercise_history`

Return previous saved attempts, checks, and failure information for an exercise when current context alone is insufficient.

### `get_review_history`

Return relevant spaced-repetition history for one or more objectives/review items.

Tool definitions should use typed Go argument/result structures. Prefer deriving tool JSON schemas from typed definitions using existing project schema machinery when that can be done cleanly; avoid maintaining redundant hand-written schemas without a concrete reason.

The registry should validate model-supplied arguments before invoking a tool.

## Tool implementation boundary

Tools are ordinary Go application capabilities, not miniature agents.

A tool should have a clear name, narrow responsibility, typed arguments, typed or serializable results, and a deterministic execution path. The tool layer should not contain provider-specific logic.

Conceptually:

```go
type Tool interface {
    Name() string
    Description() string
    Schema() JSONSchema
    Execute(ctx context.Context, args json.RawMessage) (any, error)
}
```

The exact interfaces should fit the existing server architecture rather than copying this sketch literally.

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

## Recent activity

The context builder may retrieve a bounded slice of recent learner activity when it is predictably useful. The model may use `get_recent_learning_activity` for additional history when the initial bounded context is insufficient.

This enables useful interpretations of vague questions such as:

> Why am I still not getting this?

when the system knows the learner has just failed the same convergence test three times.

## Worksheet review

A completed worksheet is a deliberate multimodal input, not routine screen context. The learner may upload images or a document containing handwritten or typed work for review.

The tutor should evaluate submitted work against authored problem answers or rubrics when those are available. For deterministic problems, the vision model should interpret the learner's work rather than invent the expected answer itself.

Worksheet review should consider both results and demonstrated reasoning. Useful feedback includes:

- whether the final answer is correct;
- whether sufficient work is shown when reasoning is part of the task;
- which intermediate steps are correct;
- the earliest identifiable conceptual or computational error;
- recurring misconceptions across several problems;
- concepts that appear strong;
- concepts that appear to need more practice or prerequisite review.

If handwriting, problem association, or an intermediate step is ambiguous, the tutor should state the uncertainty rather than fabricate a confident interpretation.

Worksheet feedback may contribute evidence about objective-level application or conceptual understanding and may recommend follow-up practice or a mastery check. It should not directly award mastery from model judgment alone.

Worksheet upload storage, grading persistence, and answer-key access are later implementation slices after the core text/vision harness has been evaluated.

See `docs/worksheets.md` for the curriculum role and assessment model for printable practice.

## Provider event boundary

OpenRouter and future direct providers may have materially different execution models. Preserve those differences inside provider adapters.

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

The UI should not need to understand OpenRouter streaming payloads or another provider's protocol details.

## Grounding and citations

For curriculum questions, prefer curriculum sources and attach source IDs to supported claims when possible.

Never invent source metadata. The content source registry is authoritative.

For discussion beyond the curriculum, the tutor should distinguish sourced curriculum material from unsourced model knowledge.

Do not give the initial tutor unrestricted web search. Core tutoring should rely on the authored curriculum, curated sources, learner state, and model knowledge. A deliberate Explore/research tool can be considered later if there is a real learning need and the product has a clear provenance and prompt-injection policy for external content.

## Writes to learner state

The initial toolset is read-only.

Future write operations can be useful, but they must be explicit and auditable. Examples include:

- "Save that as a note."
- "Make a review-card candidate from this."
- "Record that I completed the exercise."

The tutor may recommend a mastery check. It should not mark an objective mastered based only on conversational confidence.

## Evaluation gates

The first serious model comparison should happen after the production harness, OpenRouter adapter, conversation persistence, context builder, and initial read-only toolset exist.

That comparison should exercise the real harness rather than a parallel benchmark-only implementation. It should evaluate at least:

- factual and mathematical correctness;
- pedagogical quality;
- engaging multi-turn conversation;
- misconception detection and correction;
- adherence to tutor modes;
- tool selection and argument correctness;
- unnecessary tool calls;
- use of retrieved evidence rather than hallucinated state;
- vision/worksheet interpretation;
- uncertainty calibration;
- citation/provenance behavior;
- latency;
- token consumption;
- actual cost.

The selected model remains replaceable. The harness and application-owned conversation state are the durable product architecture; the model is a configured dependency.
