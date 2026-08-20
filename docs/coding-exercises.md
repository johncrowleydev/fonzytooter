# Coding exercises

## Boundary

Fonzytooter has one in-app Python execution environment: Pyodide in the browser.

There is intentionally no backend Python runner.

Small, constrained practice and assessment exercises belong in the embedded CodeMirror + Pyodide experience.

Jupyter notebooks are a separate **external learning format** used for exploratory scientific/ML work. They are not embedded in Fonzytooter and do not create a second in-app Python runtime.

When work becomes substantial, multi-file, GPU-dependent, data-heavy, or otherwise uncomfortable in Pyodide, it should become a repository-based lab/project completed with normal development tools.

The intended progression is:

```text
Fonzytooter embedded exercise
    -> practice and assessment

Jupyter notebook
    -> exploration and intuition-building

Git repository + IDE
    -> engineering, synthesis, and substantial experiments
```

See [`courses/ai-ml.md`](courses/ai-ml.md) for how those environments fit into the current AI/ML course plan.

## Embedded workflow

The implemented editor uses CodeMirror 6 and runs Pyodide 314.0.5 in a
module-type Web Worker. Pyodide is loaded once per worker and remains warm.
The application-facing `PythonRunner` boundary owns request correlation,
structured results, and timeout recovery; React components do not receive
Pyodide objects or proxies.

The normal exercise resource exposes only visible authored tests. The separate
course-qualified `check-definition` read resource supplies all test code to the
learner's own browser when Check is requested. Each authored check gets a fresh
Python namespace, and only stable test IDs, titles, statuses, diagnostics, and
durations are persisted. Hidden source is never rendered in the exercise UI.

Workspace code is saved to SQLite after a short debounce. A browser-local draft
is retained for failure recovery while the server workspace remains the
cross-device source of truth. Exploratory Run output is not assessment evidence;
Check creates an immutable attempt, normalized test-result rows, and an
`exercise_checked` learner activity.

The runtime pins the official stable Pyodide distribution and exact CDN base:

```text
Pyodide 314.0.5
https://cdn.jsdelivr.net/pyodide/v314.0.5/full/
```

## UI

```text
┌─────────────────────────────────────────────────────┐
│ Implement gradient descent                          │
│                                                     │
│ prompt / supporting explanation                     │
│                                                     │
│ ┌─────────────────────────────────────────────────┐ │
│ │ CodeMirror                                      │ │
│ │                                                 │ │
│ └─────────────────────────────────────────────────┘ │
│                                                     │
│ [ Run ] [ Check ]                     [ Ask tutor ] │
├─────────────────────────────────────────────────────┤
│ output / test results                               │
└─────────────────────────────────────────────────────┘
```

## Run versus Check

### Run

Exploratory execution inside the scope of a small embedded exercise. Capture stdout/stderr and, later, useful renderable artifacts such as plots where practical.

This is distinct from a Jupyter notebook lab, where open-ended exploration is the actual learning activity.

### Check

Assessment execution. Run the exercise's tests and persist an attempt/result.

Tests may include visible examples plus non-disclosed checks where useful.

## Test philosophy

ML/math problems often require numerical or property-based correctness rather than exact string outputs.

Useful checks include:

- `allclose`-style numerical tolerance;
- shape/dimension correctness;
- probabilities sum to approximately 1;
- loss decreases after an update;
- numerical and analytic gradients approximately agree;
- input is not unexpectedly mutated;
- deterministic result with a supplied random seed.

The tests themselves can reinforce what "correct" means mathematically.

## Authored definition

Embedded exercise definitions are optional Git-authored YAML files under a module's direct-child `exercises/` directory. They attach to one lesson, reference learning objectives, and contain a prompt, starter code, and visible or hidden trusted Python tests. The curriculum loader validates and indexes this definition without executing Python.

The course-qualified student API returns exercise metadata, prompt, starter code, and visible tests only. Hidden test code remains server-internal curriculum data. A later runtime workflow may use it for checks through a deliberately separate contract; the normal exercise read resource must never disclose it.

## Browser execution

Pyodide should run in a Web Worker to isolate Python execution from the React UI thread.

The worker should expose a small application-facing contract rather than leaking Pyodide APIs through components.

Conceptually:

```ts
interface PythonRunner {
  run(request: PythonRunRequest): Promise<PythonRunResult>;
  check(request: PythonCheckRequest): Promise<PythonCheckResult>;
}
```

## Persistence

Working code should survive navigation and device changes.

A reasonable approach is:

- local browser persistence during active editing;
- periodic or navigation-time sync to the Go API;
- SQLite as the cross-device source of truth for saved workspaces and attempts.

Do not optimize the synchronization protocol until real usage demonstrates a problem.

## Jupyter notebook labs

Jupyter should be introduced early in the Scientific Python curriculum and then used naturally for exploratory work.

Notebook labs are appropriate when the learner should vary parameters, inspect intermediate results, make plots, compare alternatives, or document an experiment rather than simply produce a function that passes tests.

Examples include:

- visualize gradient-descent learning-rate behavior;
- inspect distributions in a dataset;
- compare initialization strategies;
- plot training curves;
- inspect activations or attention patterns;
- explore numerical stability or quantization error.

Notebook files should live in Git repositories and be opened using normal Jupyter-compatible tooling. Fonzytooter may link to the lab and track its objectives/status, but should not become a Jupyter server or notebook host.

## Graduation to real projects

Examples that belong outside the embedded editor:

- train a CNN on a substantial dataset;
- build and train a transformer;
- LoRA fine-tuning;
- GPU memory profiling;
- quantization experiments;
- multi-file ML systems.

The app can track the assignment, objectives, repository URL, status, and reflection. The actual code lives in Git.
