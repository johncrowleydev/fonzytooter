# AI/ML curriculum plan

This document defines the current high-level curriculum direction for Fonzytooter. It is intentionally a plan, not the final detailed syllabus. Individual modules, objectives, readings, videos, exercises, notebook labs, and projects will be authored and refined later.

The curriculum is designed for a strong professional software engineer who is largely self-taught in computer science and has not completed formal university coursework in calculus, linear algebra, probability/statistics, or related ML mathematics.

The goal is not merely to learn how to call AI APIs. The goal is to understand machine learning deeply enough to reason about models, reproduce important ideas, run meaningful experiments, read papers, and eventually operate at an ML/research-engineering level.

## Core principles

### Self-paced, not calendar-paced

The curriculum has ordering and prerequisites, but no weeks, semesters, due dates, target completion dates, or "behind schedule" state.

Spaced repetition uses elapsed time internally, but curriculum progress is structural rather than calendar-based.

### Objectives before lessons

Learning objectives are the connective tissue of the curriculum.

A concept may be supported by:

- authored lessons;
- readings;
- curated YouTube videos;
- interactive visualizations;
- recall/review items;
- mathematical exercises;
- embedded coding exercises;
- Jupyter notebook experiments;
- repository-based projects;
- mastery/transfer checks.

The curriculum should not confuse "read a lesson" with "learned the objective."

### Math is learned just in time

Do not require a long isolated mathematics curriculum before meaningful ML begins.

Teach mathematical ideas near the ML concepts that motivate them, while maintaining enough prerequisite structure to avoid hand-waving.

The main mathematical spine is:

- functions, algebra, exponentials, logarithms, sequences, and notation;
- linear algebra;
- probability and statistics;
- calculus;
- optimization;
- information theory;
- numerical computing.

Discrete mathematics remains useful for computer science generally, but it has lower priority than the above topics for the ML path.

### History runs in parallel

Technical ideas should be situated historically where that improves understanding.

Important historical landmarks include:

- McCulloch and Pitts;
- Turing;
- Dartmouth and early symbolic AI;
- the perceptron;
- expert systems;
- AI winters;
- backpropagation;
- statistical machine learning;
- ImageNet and deep learning;
- word embeddings;
- sequence-to-sequence models;
- attention;
- the Transformer;
- scaling laws;
- instruction tuning and RLHF;
- multimodal models;
- modern reasoning systems and agents.

Recurring historical/technical tensions are worth making explicit:

- symbolic systems vs statistical learning;
- hand-engineered features vs learned representations;
- specialized systems vs general models;
- training-time compute vs inference-time compute;
- monolithic learned systems vs models combined with deterministic tools.

### Foundational implementations should initially be human-written

Coding agents are part of the learning environment, but they should not erase the work that creates understanding.

For foundational concepts, the learner should personally write the first implementation before using agents aggressively for review, tests, debugging, refactoring, or extension.

Examples include:

- linear regression;
- logistic regression;
- gradient descent;
- backpropagation;
- a small neural network;
- attention;
- a transformer block;
- an autoregressive language-model training loop;
- sampling;
- LoRA.

Once the learner has demonstrated understanding, agents may be used normally.

## Three coding environments

The curriculum deliberately uses three different coding environments for three different kinds of work.

### 1. Fonzytooter embedded exercise: practice

Small, focused coding problems belong in the app's CodeMirror + Pyodide environment.

Examples:

- implement `sigmoid`;
- calculate softmax;
- implement matrix multiplication;
- write one gradient-descent update;
- calculate MSE;
- implement a small numerical derivative;
- manipulate NumPy arrays.

These exercises are constrained enough to run comfortably in the browser and can be automatically checked.

Their purpose is **practice and assessment**.

Pyodide is the only Python interpreter built into Fonzytooter. There is no backend Python runner.

### 2. Jupyter notebook: exploration

Jupyter notebooks are used when the goal is to investigate a concept experimentally rather than merely solve a small exercise.

A notebook mixes executable Python cells with Markdown, mathematical explanation, plots, tables, and retained outputs. A Python kernel keeps state between cell executions, making notebooks particularly useful for exploratory scientific and ML work.

Examples:

- visualize how different learning rates affect convergence;
- inspect a dataset and plot distributions;
- compare several loss functions;
- explore matrix transformations geometrically;
- inspect neural-network activations;
- vary initialization strategies and plot training curves;
- investigate attention patterns;
- compare quantization error experimentally.

The purpose is **experimentation and intuition-building**.

Jupyter should be introduced early in the Scientific Python phase. The learner should understand:

- code cells;
- Markdown cells;
- kernels and persistent state;
- inline plots and outputs;
- restarting the kernel;
- running all cells from a clean state;
- why out-of-order execution can produce misleading or irreproducible notebooks;
- when a notebook is appropriate;
- when reusable code should move into normal `.py` modules.

Fonzytooter should **not embed or host Jupyter**. Notebook files should live in Git repositories and be opened with normal Jupyter-compatible tooling.

This preserves a clean application boundary:

```text
embedded exercise
    -> practice a concept

Jupyter notebook
    -> explore a concept

Git repository + IDE
    -> build with a concept
```

### 3. Git repository + normal IDE: engineering and substantial experiments

When work becomes large, multi-file, GPU-dependent, data-heavy, or otherwise uncomfortable in Pyodide, it should be a real repository-based lab or project.

Examples:

- build a neural network from scratch;
- train a CNN;
- build and train a small Transformer;
- train a tiny language model;
- fine-tune an open model with LoRA;
- benchmark inference systems;
- profile GPU memory;
- compare quantization methods;
- reproduce a paper result.

The purpose is **engineering, synthesis, and transfer**.

Fonzytooter may track the assignment, objectives, repository, status, evidence, and reflection, but the code itself belongs in Git and a real editor/IDE.

## Curriculum phases

The phases below describe conceptual progression, not calendar periods.

## Phase 0: Scientific Python and computational experimentation

Goal: become comfortable using Python as the language of numerical and ML experimentation without requiring a detour through unrelated Python application development.

Topics include:

- Python syntax needed for scientific work;
- NumPy arrays and vectorized operations;
- shapes, dimensions, indexing, broadcasting;
- plotting;
- packages and environments;
- Jupyter notebooks;
- Markdown and mathematical notation in notebooks;
- reproducible notebook habits;
- type hints where useful;
- basic scientific-computing workflow;
- introductory numerical thinking.

Tooling should favor a typed, disciplined workflow where practical, including tools such as Pyright/mypy, Ruff, and modern package/environment tooling.

Do not spend major curriculum time on Python web frameworks, deep Python OOP patterns, or application architecture unrelated to ML.

Representative work:

- embedded NumPy exercises;
- notebook-based plotting and numerical experiments;
- small data-inspection tasks.

## Phase 1: Classical machine learning

Goal: understand the basic machine-learning problem before neural networks dominate the curriculum.

Topics include:

- supervised vs unsupervised learning;
- regression and classification;
- features and targets;
- loss/objective functions;
- train/validation/test splits;
- generalization;
- overfitting and underfitting;
- regularization;
- bias/variance;
- cross-validation;
- data leakage;
- class imbalance;
- calibration;
- metrics;
- linear regression;
- logistic regression;
- decision trees;
- clustering;
- k-means;
- PCA and dimensionality reduction;
- scikit-learn workflow.

Important implementations should be written manually before relying on libraries:

- linear regression;
- logistic regression;
- k-means;
- a small decision tree where pedagogically useful.

Then solve comparable problems using scikit-learn.

Representative work:

- browser exercises for losses, gradients, metrics, and small algorithms;
- notebooks comparing models and visualizing decision boundaries;
- a rigorous classical-ML project with train/validation/test discipline and clear evaluation.

## Phase 2: Neural networks from first principles

Goal: make neural networks understandable as compositions of ordinary mathematical operations rather than magic.

Topics include:

- perceptrons;
- layers;
- weights and biases;
- activation functions;
- loss functions;
- computational graphs;
- derivatives and partial derivatives;
- gradients;
- the chain rule;
- backpropagation;
- automatic differentiation conceptually;
- initialization;
- gradient descent;
- SGD;
- momentum;
- Adam;
- regularization;
- normalization;
- embeddings.

A central project should be a small neural-network implementation from scratch without relying on a deep-learning framework for the core mechanics.

Representative work:

- embedded derivative/backprop exercises;
- notebook experiments with loss surfaces, initialization, and optimization;
- repository project: neural network from scratch.

## Phase 3: Deep learning and PyTorch

Goal: transition from hand-built neural-network mechanics to modern framework-based model development and GPU computation.

Topics include:

- PyTorch tensors;
- tensor shapes and device placement;
- autograd;
- `nn.Module`;
- datasets and dataloaders;
- training loops;
- evaluation loops;
- checkpoints;
- mixed precision;
- CUDA concepts;
- profiling;
- memory constraints;
- reproducibility;
- experiment organization.

Architectural history should include:

- multilayer perceptrons;
- CNNs;
- RNNs;
- LSTMs/GRUs;
- why attention eventually displaced recurrent architectures for many sequence tasks.

Representative work:

- notebooks exploring tensor operations and training behavior;
- repo-based model training;
- experiments comparing architectures or optimization choices.

## Phase 4: Transformers from scratch

Goal: understand the Transformer deeply enough to implement a small language model rather than treating attention as an API primitive.

Topics include:

- tokenization;
- vocabulary and token IDs;
- embeddings;
- positional information;
- queries, keys, and values;
- dot-product attention;
- causal masks;
- multi-head attention;
- residual connections;
- normalization;
- feed-forward blocks;
- logits;
- softmax;
- cross-entropy;
- autoregressive training;
- sampling and decoding.

Build:

- attention manually;
- a transformer block manually;
- a tiny GPT-style language model;
- training and evaluation loops;
- autoregressive generation.

Stanford-style "language modeling from scratch" material is an appropriate target level for this phase.

Representative work:

- embedded attention/math exercises;
- notebooks visualizing attention and sampling behavior;
- repository project: TinyLM.

## Phase 5: Modern language models

Goal: connect the basic Transformer to the systems and techniques used by contemporary open and frontier language models.

Architecture topics include:

- RMSNorm;
- RoPE;
- modern gated feed-forward layers such as SwiGLU-like designs;
- grouped-query and multi-query attention;
- mixture-of-experts;
- long-context techniques;
- multimodal extensions.

Training/data topics include:

- dataset construction and filtering;
- pretraining objectives;
- scaling laws;
- compute/data tradeoffs;
- Chinchilla-style scaling considerations;
- evaluation design.

Post-training topics include:

- supervised fine-tuning;
- preference learning;
- reward models conceptually;
- RLHF conceptually;
- PPO conceptually;
- DPO;
- RLAIF;
- verifiable rewards;
- reasoning/test-time compute.

Adaptation topics include:

- full fine-tuning;
- LoRA;
- PEFT;
- quantization;
- low-resource local experiments.

The learner cannot reproduce frontier-scale training locally, but should perform the same *categories* of experiments at small scale where possible.

Representative project:

- take an open model through a small fine-tuning/adaptation/evaluation/serving workflow.

## Phase 5B: Inference and ML systems for LLMs

Goal: understand why inference behaves the way it does on real hardware.

Topics include:

- parameter memory;
- activation memory;
- FLOPs;
- memory bandwidth;
- arithmetic intensity;
- batching;
- prefill vs decode;
- KV cache;
- quantization;
- speculative decoding;
- continuous batching;
- tensor parallelism;
- pipeline parallelism;
- data parallelism;
- GPU/CPU tradeoffs;
- latency vs throughput.

Use multiple inference stacks experimentally where useful, for example:

- raw PyTorch/Transformers;
- llama.cpp;
- Ollama;
- LM Studio;
- vLLM or equivalent serving systems where hardware permits.

The purpose is not memorizing product APIs but understanding the underlying systems tradeoffs.

## Phase 6: Broaden beyond language models

Goal: avoid conflating "AI" with "LLMs."

Potential areas include:

- computer vision;
- representation learning;
- multimodal systems;
- diffusion and flow-based generative models;
- speech and audio;
- reinforcement learning;
- probabilistic methods;
- additional classical/specialized ML methods.

Reinforcement-learning foundations should include:

- states;
- actions;
- rewards;
- policies;
- value functions;
- Markov decision processes;
- Bellman equations;
- Q-learning;
- policy gradients;
- actor-critic methods.

This phase can become more elective based on interest.

## Phase 7: Research and ML systems practice

Goal: transition from following a curriculum to doing increasingly independent technical work.

Activities include:

- reading papers;
- reproducing selected results;
- implementing methods from papers;
- running controlled experiments;
- designing ablations;
- forming hypotheses;
- measuring failures;
- writing experiment reports;
- asking novel technical questions;
- contributing to or building ML systems.

This phase should increasingly resemble research-engineering practice rather than coursework.

## Mathematics spine

Math should be connected to concrete ML use wherever possible.

### Functions and algebra

Topics include:

- domain and codomain;
- mappings;
- composition;
- inverse functions;
- exponentials;
- logarithms;
- sequences;
- notation;
- asymptotic intuition where useful.

### Linear algebra

Topics include:

- vectors;
- matrices;
- dot products;
- matrix multiplication;
- linear transformations;
- basis and coordinates;
- norms;
- projections;
- eigenvalues/eigenvectors where motivated;
- decompositions where useful;
- high-dimensional geometric intuition.

### Probability and statistics

Topics include:

- random variables;
- common distributions;
- expectation;
- variance;
- covariance;
- conditional probability;
- Bayes' rule;
- sampling;
- estimators;
- uncertainty;
- likelihood;
- hypothesis/evaluation concepts needed for experiments.

### Calculus

Topics include:

- limits conceptually;
- derivatives;
- rates of change;
- partial derivatives;
- gradients;
- chain rule;
- multivariable differentiation;
- integration when useful rather than as an arbitrary prerequisite sequence.

### Optimization

Topics include:

- objective functions;
- local/global minima;
- convexity intuition;
- gradient-based optimization;
- conditioning;
- learning rates;
- stochastic optimization;
- momentum/adaptive optimizers.

### Information theory

Topics include:

- entropy;
- cross-entropy;
- KL divergence;
- mutual information where useful;
- relationship to model objectives and probabilistic prediction.

### Numerical computing

Topics include:

- floating-point limitations;
- numerical stability;
- vectorization;
- precision;
- overflow/underflow;
- stable softmax/log calculations;
- reproducibility and random seeds.

## Typical module composition

A module does not need every activity type, but a rich module may contain:

```text
Module
├── objectives
├── prerequisite review
├── sourced MDX lessons
├── curated readings
├── curated YouTube playlist
├── interactive visualization(s)
├── spaced-repetition items
├── conceptual/math exercises
├── embedded Pyodide coding exercises
├── Jupyter notebook lab(s)
├── mastery check
└── repository project or synthesis task
```

Example:

```text
Linear Regression
├── Lesson: what it means to fit a line
├── Math: squared error and derivatives
├── Video: geometric/statistical intuition
├── Embedded exercise: implement MSE
├── Embedded exercise: compute/update a gradient
├── Notebook lab: visualize loss and learning-rate behavior
├── Review cards: regression/generalization vocabulary
└── Repo lab: implement and evaluate linear regression
```

## Assessment and retention

Progress should use multiple forms of evidence.

### Recall

Spaced repetition assesses whether concepts remain retrievable over time.

### Conceptual understanding

Explanation, comparison, prediction, and reasoning tasks assess whether the learner understands why a technique behaves as it does.

### Mathematical application

Problems assess whether the learner can actually perform the relevant algebra/calculus/probability rather than recognizing terminology.

### Coding/application

Embedded exercises and tests assess constrained implementation ability.

### Transfer

Notebook investigations, open-ended tasks, projects, novel debugging scenarios, and synthesis work assess whether knowledge transfers beyond rehearsed examples.

Avoid collapsing all of these into a fake-precision single mastery percentage.

## AI tutor and agent policy

The tutor should know what the learner is currently doing and what relevant objectives they have struggled with, but it should not silently complete the core learning work.

For foundational coding exercises, default tutor behavior should scaffold:

- explain errors;
- ask guiding questions;
- point back to relevant concepts;
- offer progressively stronger hints;
- review an attempted implementation.

A full solution can be available when explicitly requested, but should not be the default interaction.

The tutor may suggest that an objective is ready for assessment, but it does not award mastery merely because a conversation sounded convincing.

## Sources and curriculum quality

All substantive lesson content must have an inspectable source basis.

Prefer:

- original papers;
- official framework/tool documentation;
- reputable textbooks;
- strong university course material;
- other high-quality educational sources where appropriate.

YouTube resources are curated for explanation and intuition but are not automatically authoritative citations.

AI-generated lesson text must follow the same workflow as human-authored content:

1. draft;
2. identify and verify reputable sources;
3. check claims against those sources;
4. review;
5. commit with stable source references.

## Portfolio direction

The curriculum should naturally produce a small portfolio of technically meaningful work rather than dozens of toy notebooks.

Representative milestone projects include:

- a rigorous classical-ML project;
- neural network from scratch;
- TinyLM / transformer from scratch;
- open-model fine-tuning + evaluation + inference project;
- one domain/research project based on learner interest.

The precise projects can evolve with interests and available hardware.

## Career direction

The curriculum should support a progression from existing software-engineering strengths toward ML/research engineering.

A rough distinction:

- **AI application engineer:** integrates models into products and systems;
- **ML/research engineer:** additionally understands training, data, model internals, evaluation, inference, experiments, and paper-to-code work;
- **researcher:** typically requires deeper mathematical specialization and sustained novel-method research.

The initial curriculum target is research/ML engineering depth while preserving strong software-engineering practice.

## Future syllabus work

This document deliberately stops at the high-level curriculum architecture.

The next curriculum-design step is to expand these phases into a detailed sequence of modules and objectives. A module may eventually follow a pattern such as:

```text
motivation/history
    -> mathematical prerequisite
    -> conceptual lesson
    -> reading/video
    -> embedded practice
    -> notebook exploration
    -> mastery check
    -> project/synthesis
    -> spaced review
```

The detailed syllabus should be authored incrementally so the learning-system implementation does not become a substitute for actually using the curriculum.