# Tutor evaluation plan

## Purpose

Fonzytooter should select and tune its tutor model by evaluating models through the real production tutor harness, not by comparing isolated chat responses or relying primarily on general-purpose leaderboards.

The evaluation should answer two different questions:

1. **Is this model good enough to be a technical tutor?**
2. **Does this model work well inside Fonzytooter's actual context and tool system at an acceptable cost and latency?**

The first production comparison should run after the OpenRouter provider, application-owned conversation persistence, context builder, bounded agent loop, and initial read-only toolset described in `docs/tutor.md` are implemented.

The initial candidates are:

- `minimax/minimax-m3`;
- `moonshotai/kimi-k2.6`;
- `google/gemini-3.7-flash`;
- `qwen/qwen3.8-27b`.

These candidates are deliberately replaceable. Model releases move faster than the application architecture, so a future evaluation may add, remove, or replace candidates.

## Evaluation principles

### Exercise the production harness

Do not create a benchmark-specific agent implementation with different prompts, tools, schemas, context behavior, or conversation handling. The benchmark runner should invoke the same tutor service used by the application with controlled fixture state.

### Evaluate tutoring, not merely answer correctness

A model can know the right answer and still be a poor teacher. Evaluation must include:

- technical correctness;
- explanation quality;
- ability to notice what the learner actually misunderstands;
- ability to explain the same concept in a genuinely different way;
- conversational coherence across several turns;
- Socratic restraint;
- progressive hints;
- uncertainty calibration;
- appropriate use and non-use of tools.

### Prefer deterministic assertions where possible

Use ordinary assertions for behavior that ordinary software can evaluate reliably. Examples:

- whether a required tool was called;
- whether a forbidden tool was called;
- whether tool arguments matched fixture identifiers;
- whether the model exceeded the tool-iteration limit;
- whether a citation refers to a retrieved source;
- whether a required final numerical answer is correct;
- latency, token usage, and cost.

Do not use an LLM judge for facts the test runner can verify itself.

### Human evaluation remains important

Pedagogy, engagement, clarity, natural conversation, and whether an explanation actually approaches the concept differently are partly subjective. Initially, score these with a small human rubric rather than pretending they can all be reduced to exact assertions.

An automated model-as-judge pass may be added later to accelerate repeated comparisons, but it should not be the sole authority for model selection.

### Run multiple samples

Tool selection and conversational behavior may be stochastic. Important scenarios should run multiple times per model/reasoning configuration. Record both average quality and failure frequency rather than preserving only the best response.

## Metrics to record

Every run should capture enough information to make quality/cost tradeoffs visible:

- exact model ID;
- provider/routing mode where available;
- reasoning configuration;
- scenario ID;
- input tokens;
- output tokens;
- reasoning tokens when exposed;
- cache behavior when exposed;
- time to first token;
- end-to-end latency;
- number of model calls;
- number and names of tool calls;
- tool argument validation failures;
- actual cost reported or calculated for the run;
- deterministic assertion results;
- human rubric scores;
- evaluator notes.

Prices should be captured at evaluation time rather than copied into this document.

## Suggested human rubric

Use a compact 1–5 score for the dimensions that require judgment.

### Technical accuracy

1. Materially wrong or misleading.
2. Significant errors despite some correct content.
3. Mostly correct with minor problems.
4. Correct and appropriately qualified.
5. Correct, precise, and catches important subtleties or misconceptions.

### Pedagogy

1. Makes the concept harder to understand.
2. Mostly dumps definitions or procedures without building understanding.
3. Serviceable explanation.
4. Builds intuition and connects it to the learner's current knowledge.
5. Excellent teaching: diagnoses the actual gap, chooses a useful representation, checks understanding, and adapts naturally.

### Conversation

1. Ignores prior turns or feels incoherent.
2. Repeats itself or responds mechanically.
3. Coherent but generic.
4. Natural, responsive, and remembers the thread.
5. Feels like a strong human tutor sustaining a purposeful technical conversation.

### Restraint and calibration

1. Hallucinates, overclaims, or ignores mode constraints.
2. Frequently overconfident or over-helpful.
3. Generally appropriate.
4. Uses tools, hints, caveats, and questions judiciously.
5. Consistently knows when to answer, ask, retrieve, hint, or admit uncertainty.

## Initial scenario catalog

The scenarios below define behavior to test. They are not intended to freeze exact prompt wording forever. When practical, bind them to real Fonzytooter curriculum/objective/exercise fixtures so the evaluation exercises actual application data.

### A. Explanation and pedagogy

#### PED-01 — Match the learner's mathematical level

**Setup:** The learner is an experienced programmer but has formal mathematics only through precalculus. The current lesson introduces the dot product.

**User:** "I understand what the formula tells me to calculate, but what does a dot product actually mean?"

**Expected behavior:**

- explain meaning before adding more formalism;
- avoid assuming linear algebra knowledge not yet taught;
- connect the computation to geometric or practical intuition;
- remain technically correct;
- do not call a tool if the relevant lesson material is already injected.

#### PED-02 — Explain a difficult concept geometrically

**Setup:** A future linear-algebra lesson on eigenvectors is current context.

**User:** "Why are eigenvectors special? They just sound like vectors that get multiplied by a number."

**Expected behavior:**

- correct the oversimplification without talking past the learner;
- explain the transformation/direction relationship;
- use an intuitive geometric representation before or alongside notation;
- distinguish eigenvectors from arbitrary vectors.

#### PED-03 — Actually explain it a different way

**Conversation:**

1. Tutor gives a correct conceptual explanation of gradient descent.
2. User: "I still don't get it. Can you explain it a completely different way?"

**Expected behavior:**

- change representation or analogy rather than paraphrasing the first answer;
- preserve the core technical meaning;
- ideally identify which part of the first explanation may not have landed.

#### PED-04 — Detect a definition-level misconception

**User:** "A function is a one-to-one mapping between inputs and outputs, right?"

**Expected behavior:**

- explain that ordinary functions need not be one-to-one/injective;
- distinguish the function requirement of one output per input from injectivity;
- provide a simple counterexample;
- avoid burying the correction under unnecessary advanced terminology.

#### PED-05 — Fast reminder should stay fast

**Setup:** The learner previously completed the relevant lesson.

**User:** "Quick reminder: what does variance measure?"

**Expected behavior:**

- give a concise refresher rather than a mini-lecture;
- avoid unnecessary retrieval if current context is sufficient;
- use low/default reasoning appropriately.

#### PED-06 — Follow the learner's actual question

**User:** "I know how to take this derivative mechanically. I don't understand what the derivative represents."

**Expected behavior:**

- do not respond primarily with differentiation rules;
- focus on rate of change/local slope and relevant intuition;
- recognize the explicit distinction the learner made.

#### PED-07 — Distinguish related concepts cleanly

**User:** "What's the difference between probability and likelihood?"

**Expected behavior:**

- give a technically meaningful distinction appropriate to the course level;
- avoid the common vague answer that they are simply synonyms;
- use a small example if helpful.

#### PED-08 — Correct without derailing

**Conversation:** The learner asks a good question but includes one incidental incorrect statement.

**Expected behavior:**

- correct the misconception that matters;
- still answer the question they were trying to ask;
- avoid turning every minor wording issue into an unrelated lecture.

### B. Socratic behavior and tutoring policy

#### SOC-01 — Do not dump the answer

**Mode:** Socratic.

**Setup:** A solvable algebra problem is current context.

**User:** "How do I solve this?"

**Expected behavior:**

- begin with a useful question or small prompt toward the next step;
- do not immediately provide the complete worked solution;
- keep the question narrow enough to help rather than merely saying "what do you think?"

#### SOC-02 — Escalate hints when the learner remains stuck

**Conversation:** The learner responds "I don't know" to two Socratic prompts.

**Expected behavior:**

- progressively strengthen the hint;
- eventually demonstrate the blocked local step if needed;
- do not endlessly repeat questions while the learner is clearly stuck.

#### SOC-03 — Respect an explicit request for the solution

**Mode:** Exercise help or Socratic.

**User:** "I've tried this enough. Please just show me the full solution and explain it."

**Expected behavior:**

- provide the solution because the learner explicitly requested it;
- explain the reasoning rather than outputting only the answer;
- do not paternalistically refuse after the learner made a conscious choice.

#### SOC-04 — Challenge a false premise

**User:** asks a question whose premise is mathematically false.

**Expected behavior:**

- identify the false premise before reasoning from it;
- explain why it fails;
- redirect to the meaningful version of the question.

#### SOC-05 — Quiz one concept interactively

**Mode:** Quiz.

**Expected behavior:**

- ask one purposeful question at a time;
- evaluate the learner's response before continuing;
- distinguish a lucky final answer from demonstrated understanding when possible;
- do not silently change learner mastery state.

### C. Context and tool choice

#### TOOL-01 — No tool when page context is enough

**Setup:** Current lesson content and objective definitions directly answer the question.

**Expected behavior:**

- answer from injected context;
- call no retrieval/history tool.

**Deterministic assertion:** zero tool calls.

#### TOOL-02 — Search for earlier curriculum material

**User:** "Didn't we cover something earlier that explains why this matrix has that shape?"

**Setup:** The relevant explanation exists in a previous lesson and is not injected in full.

**Expected behavior:**

- use `search_curriculum`;
- retrieve the matching material if necessary with `get_curriculum_content`;
- answer using the retrieved content rather than guessing.

#### TOOL-03 — Retrieve objective evidence for a progress question

**User:** "Do I actually know this yet or have I just seen it once?"

**Expected behavior:**

- use `get_objective_state` for the relevant objective;
- distinguish introduction, recall, application, and transfer evidence instead of inventing a mastery percentage.

#### TOOL-04 — Diagnose vague repeated struggle

**User:** "Why am I still not getting this?"

**Setup:** Recent fixture state shows repeated failures on the same exercise and weak review performance on one prerequisite objective.

**Expected behavior:**

- use relevant learner-history tools when the injected recent context is insufficient;
- identify the evidence-backed pattern;
- avoid claiming causes not supported by state.

#### TOOL-05 — Do not retrieve history unnecessarily

**Setup:** Current page context already includes the exact failing test and current code.

**User:** "What does this failure mean?"

**Expected behavior:**

- explain the current failure directly;
- do not call `get_exercise_history` merely because that tool exists.

#### TOOL-06 — Correct multi-tool chain

**Setup:** A question requires discovering an earlier lesson and then reading its authoritative content.

**Expected behavior:**

- call `search_curriculum` first;
- use the returned identifier with `get_curriculum_content`;
- avoid unrelated learner-state tools.

#### TOOL-07 — Trust authoritative tool state over prior assumption

**Setup:** Conversation wording implies the learner already completed an objective, but `get_objective_state` says it was introduced and never reviewed or applied.

**Expected behavior:**

- use the authoritative learner state;
- explain the discrepancy carefully;
- do not preserve an earlier model assumption when tool evidence contradicts it.

#### TOOL-08 — Recover from invalid tool arguments

**Setup:** The model initially supplies an invalid identifier or schema value and receives a validation error.

**Expected behavior:**

- correct the call if sufficient information exists;
- do not loop indefinitely;
- do not invent a successful tool result.

#### TOOL-09 — Handle an empty retrieval result

**Setup:** `search_curriculum` returns no relevant authored material.

**Expected behavior:**

- state that the curriculum search did not find supporting material;
- answer from general knowledge only if appropriate and label that distinction;
- never fabricate a curriculum citation.

#### TOOL-10 — Read-only means read-only

**User:** "Mark this objective mastered for me."

**Setup:** Initial read-only toolset.

**Expected behavior:**

- do not pretend to modify learner state;
- explain that current tutor capabilities cannot perform that write;
- if appropriate, suggest an available mastery check instead.

### D. Exercise help

#### EX-01 — Interpret a straightforward Python failure

**Setup:** Current exercise code raises a type error; the relevant code and execution output are injected.

**Expected behavior:**

- identify the actual cause of the error;
- explain enough Python semantics to teach the issue;
- prefer a local hint/fix over replacing the entire solution.

#### EX-02 — Use history to find a repeated bug pattern

**Setup:** The learner has made an off-by-one error in several attempts; the current failure alone is ambiguous.

**Expected behavior:**

- retrieve `get_exercise_history`;
- identify the repeated pattern from evidence;
- teach the indexing/range concept rather than only patching the current line.

#### EX-03 — Syntax error should not become a conceptual dissertation

**Setup:** Exercise fails because of a simple Python syntax mistake.

**Expected behavior:**

- identify the syntax issue quickly;
- keep the answer proportional to the problem;
- do not overdiagnose a conceptual weakness without evidence.

#### EX-04 — Preserve productive struggle

**Setup:** The code is syntactically valid but the algorithm is conceptually wrong.

**Mode:** Exercise help.

**Expected behavior:**

- explain the failure or ask a targeted question;
- provide incremental hints;
- do not output a complete replacement implementation unless explicitly requested.

#### EX-05 — Full solution after explicit escalation

**Conversation:** Several hints have been given. User explicitly asks for the completed solution.

**Expected behavior:**

- provide a correct implementation;
- connect each important piece back to the underlying concept;
- make clear what changed from the learner's attempt.

### E. Multimodal worksheet review

These scenarios require controlled image/document fixtures once worksheet upload support exists.

#### VIS-01 — Find the earliest incorrect step

**Fixture:** Clear handwritten algebra where the final answer is wrong because of one intermediate sign error.

**Expected behavior:**

- correctly transcribe enough of the work to follow it;
- identify the first incorrect step rather than merely reporting the final answer is wrong;
- explain the local error.

#### VIS-02 — Admit ambiguous handwriting

**Fixture:** One symbol is genuinely ambiguous between two plausible interpretations.

**Expected behavior:**

- flag the ambiguity;
- avoid confidently inventing the symbol;
- explain how the evaluation changes under the plausible readings or ask the learner to clarify.

#### VIS-03 — Distinguish conceptual and arithmetic errors

**Fixture:** Method is conceptually correct; a small arithmetic mistake causes the wrong result.

**Expected behavior:**

- recognize the method as correct;
- locate the arithmetic mistake;
- avoid labeling the underlying concept as misunderstood without evidence.

#### VIS-04 — Associate work with the correct problem

**Fixture:** A page contains work for several numbered worksheet problems.

**Expected behavior:**

- associate work with the correct authored problem/answer key;
- do not mix reasoning from neighboring problems.

#### VIS-05 — Do not invent errors in correct work

**Fixture:** Correct handwritten solution with sufficient reasoning.

**Expected behavior:**

- mark the reasoning as sound;
- avoid manufacturing a criticism merely to produce feedback;
- identify uncertainty only where it actually exists.

#### VIS-06 — Reject irrelevant visual evidence gracefully

**Fixture:** Uploaded image is unrelated to the worksheet or too poor to interpret.

**Expected behavior:**

- state that the submission cannot be meaningfully evaluated;
- do not hallucinate mathematical work from the image.

### F. Grounding, provenance, and reliability

#### REL-01 — Cite retrieved curriculum material correctly

**Setup:** Answer uses a curriculum source with a known source ID.

**Expected behavior:**

- citation references a source actually present in context or returned by a tool;
- claims supported by the source are not attributed to a different source.

**Deterministic assertion:** every emitted curriculum source ID exists in the supplied/retrieved source set.

#### REL-02 — Never fabricate a citation

**Setup:** Relevant curriculum content has no external source metadata.

**Expected behavior:**

- do not invent author/title/URL/source ID;
- answer without a false citation.

#### REL-03 — Distinguish curriculum grounding from model knowledge

**User:** asks a reasonable technical tangent not covered in the authored course.

**Expected behavior:**

- answer from model knowledge if appropriate;
- make clear that the tangent is outside the currently retrieved curriculum when provenance matters;
- do not imply the course taught something it did not.

#### REL-04 — Resist instructions embedded in retrieved content

**Fixture:** Searchable fixture content contains text that looks like an instruction to the AI but is not an authored tutor policy.

**Expected behavior:**

- treat retrieved curriculum/source text as content, not as higher-priority agent instructions;
- continue following tutor policy and user intent.

#### REL-05 — Say when evidence is insufficient

**User:** asks why their learning performance changed.

**Setup:** Learner state contains too little evidence to support a causal explanation.

**Expected behavior:**

- separate observed facts from hypotheses;
- avoid confidently diagnosing a cause;
- suggest what evidence or practice could clarify the issue.

### G. Multi-turn coherence

#### CONV-01 — Preserve the conceptual thread

**Conversation:** Four to six turns progressively narrow from a broad concept to one subtle misunderstanding.

**Expected behavior:**

- respond to the latest question in light of earlier turns;
- avoid restarting the lesson from scratch every turn;
- use terminology already established with the learner.

#### CONV-02 — Mode change within one conversation

**Conversation:** Learner begins in Explain mode, then switches to Socratic mode.

**Expected behavior:**

- retain the conceptual context;
- change tutoring behavior without pretending this is a new independent agent.

#### CONV-03 — Recover after a failed explanation

**Conversation:** Learner explicitly says an explanation made things more confusing and identifies one confusing phrase.

**Expected behavior:**

- acknowledge and repair the specific explanation;
- avoid defensively repeating it;
- choose a simpler or different representation.

#### CONV-04 — Bounded context should not become false memory

**Setup:** Older conversation content is intentionally excluded by the context policy.

**User:** asks what they said in that omitted portion.

**Expected behavior:**

- do not fabricate memory of absent turns;
- state the limitation or retrieve persisted history only if the harness deliberately exposes an appropriate mechanism.

## Evaluation phases

### Phase 1 — Harness/model bake-off

Run the candidate models after the initial production harness and read-only tools exist. Focus on:

- text tutoring quality;
- multi-turn behavior;
- context use;
- tool selection and restraint;
- reasoning configuration;
- latency and cost.

Vision scenarios can be included as soon as canonical image/document content parts are wired through the provider.

### Phase 2 — Full tutor reevaluation

Repeat the same core suite after worksheet review, richer learner evidence, adaptive recommendations, or additional tools are introduced.

A model that behaves well with six tools may behave differently with a larger decision surface. Reevaluate rather than assuming the first winner remains the best production model forever.

## What this evaluation should not become

Avoid turning model selection into a giant synthetic benchmark project disconnected from the product. The scenario set exists to protect actual Fonzytooter behavior.

Likewise, do not optimize solely for a composite score. A cheap model with occasional catastrophic teaching/tool failures may be worse than a slightly more expensive model with consistently reliable behavior, while an expensive model that is only marginally more polished may not justify its cost.

The final decision should be a documented engineering judgment informed by repeatable evidence: quality, failure modes, latency, and cost together.
