# Coding exercises

## Boundary

Fonzytooter has one in-app Python execution environment: Pyodide in the browser.

There is intentionally no backend Python runner.

When an exercise no longer fits comfortably in Pyodide, the curriculum should promote it to a repository-based lab/project completed with normal development tools.

## Planned UI

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

Exploratory execution. Capture stdout/stderr and, later, useful renderable artifacts such as plots where practical.

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

## Browser execution

Pyodide should run in a Web Worker to isolate Python execution from the React UI thread.

The worker should expose a small application-facing contract rather than leaking Pyodide APIs through components.

Conceptually:

```ts
interface PythonRunner {
  run(request: PythonRunRequest): Promise<PythonRunResult>
  check(request: PythonCheckRequest): Promise<PythonCheckResult>
}
```

## Persistence

Working code should survive navigation and device changes.

A reasonable approach is:

- local browser persistence during active editing;
- periodic or navigation-time sync to the Go API;
- SQLite as the cross-device source of truth for saved workspaces and attempts.

Do not optimize the synchronization protocol until real usage demonstrates a problem.

## Graduation to real projects

Examples that belong outside the embedded editor:

- train a CNN on a substantial dataset;
- build and train a transformer;
- LoRA fine-tuning;
- GPU memory profiling;
- quantization experiments;
- multi-file ML systems.

The app can track the assignment, objectives, repository URL, status, and reflection. The actual code lives in Git.
