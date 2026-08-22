# AI/ML curriculum plan

This document defines the **high-level philosophy and direction** of Fonzytooter's AI & Machine Learning course.

It is intentionally not the canonical lesson sequence. The authoritative core structure lives in [`ai-ml-syllabus.md`](ai-ml-syllabus.md), with lesson-level planning in the four detailed tranche documents and final cross-course amendments in [`ai-ml-syllabus-coherence.md`](ai-ml-syllabus-coherence.md).

The hierarchy is:

```text
ai-ml.md
    high-level purpose, learner model, pedagogy, and boundaries

ai-ml-syllabus.md
    canonical mandatory Modules 00–23 and course-wide spines

ai-ml-modules-01-07.md
ai-ml-modules-08-14.md
ai-ml-modules-15-18.md
ai-ml-modules-19-23.md
    lesson-level planning

ai-ml-syllabus-coherence.md
    final cross-tranche amendments; wins where an earlier detailed plan conflicts

curriculum/courses/ai-ml/
    runtime-authored curriculum that actually exists in the product
```

The course is designed for an experienced professional software engineer who is comfortable programming but does not have a formal university foundation in calculus, linear algebra, probability/statistics, optimization, or related ML mathematics.

The goal is not merely to use AI APIs. The goal is to understand machine learning deeply enough to:

- reason from first principles about important model mechanisms;
- implement foundational algorithms personally;
- train and adapt manageable models;
- understand Transformer language models rather than treating them as opaque services;
- reason about GPU/inference behavior and quantization;
- design trustworthy experiments;
- read papers critically;
- reproduce or extend tractable results;
- continue independently toward ML/research-engineering work.

## Core principles

### Self-paced, not calendar-paced

The curriculum has ordering and prerequisites, but no semesters, target completion dates, or "behind schedule" state.

Progress is structural: the learner advances by demonstrating capabilities, not by spending a prescribed number of weeks on a topic.

### Objectives before lessons

Learning objectives are the connective tissue of the curriculum. Reading a lesson is not equivalent to mastering an objective.

An objective may be supported by:

- authored explanation;
- curated primary/secondary readings and videos;
- interactive visualizations;
- printable worksheets;
- recall/review items;
- embedded coding exercises;
- Jupyter experiments;
- repository-based implementation work;
- transfer/mastery assessments.

Runtime objective metadata should encode **real prerequisite capabilities**, not mechanically treat every earlier module as a prerequisite for every later one.

### Math is just in time, but not hand-waved

The learner should not spend a year completing detached prerequisite mathematics before seeing machine learning. Mathematics should appear when a concrete ML mechanism creates the need for it, then recur at increasing depth.

The main mathematical spine is:

- functions, algebra, notation, exponentials, logarithms, and sequences;
- linear algebra;
- calculus and optimization;
- probability and statistics;
- information theory where it explains real objectives;
- numerical computing and finite-precision reasoning.

The course intentionally does **not** require detached semester-length sequences in calculus, abstract linear algebra, real analysis, mathematical statistics, or information theory before meaningful ML work begins.

### Mechanisms before abstractions

Where implementation creates understanding, the learner should personally cross the conceptual boundary once before relying on high-level libraries or coding agents.

Important examples include:

- linear regression prediction/loss;
- gradient descent;
- logistic regression;
- k-means;
- neural-network forward/backward mechanics;
- backpropagation;
- a complete small neural network;
- attention;
- a Transformer block;
- an autoregressive language-model training loop;
- sampling/decoding;
- LoRA's low-rank update.

After the foundational implementation is understood, libraries and agents may be used normally for testing, debugging, refactoring, extension, and larger engineering work.

### History explains why ideas exist

History should be integrated into technical explanations rather than taught as a detached chronology.

The recurring question is:

> What problem, limitation, or change in available data/compute caused this idea to become useful?

Important historical threads include:

- Turing and early machine-intelligence questions;
- Dartmouth and early symbolic AI;
- McCulloch–Pitts and the perceptron;
- symbolic/expert systems versus connectionist/statistical approaches;
- AI winters as multifactorial rather than a one-event story;
- statistical machine learning and held-out empirical evaluation;
- backpropagation and neural-network revival;
- ImageNet/deep-learning resurgence;
- word embeddings and neural language models;
- recurrent seq2seq models;
- attention;
- the Transformer;
- decoder-only autoregressive language models;
- scaling laws and data/compute engineering;
- instruction tuning, preference optimization, and RLHF;
- inference-time reasoning/search, retrieval, and tool use;
- GPU/inference systems and low-precision deployment.

Recurring tensions worth revisiting include:

- symbolic systems versus statistical learning;
- hand-engineered features versus learned representations;
- specialized models versus broad foundation models;
- training-time compute versus inference-time compute;
- monolithic learned behavior versus models combined with deterministic retrieval/tools.

## Three coding environments

The course deliberately uses three different environments for different learning jobs.

### 1. Fonzytooter embedded Python — constrained practice and assessment

Small, focused Python/NumPy exercises belong in the in-app CodeMirror + Pyodide runtime.

Examples:

- array/shape manipulation;
- matrix multiplication;
- MSE;
- sigmoid and stable softmax;
- numerical derivatives;
- one gradient-descent step;
- simple metric calculations;
- tiny attention components.

The purpose is **focused practice and automatically checkable assessment**.

Pyodide is the only Python interpreter built into Fonzytooter. The course should not depend on a hidden backend Python runner.

### 2. Jupyter — exploration and visualization

Notebooks are used when the learner should vary assumptions, visualize behavior, preserve intermediate results, or conduct a small numerical experiment.

Examples:

- loss surfaces and learning-rate behavior;
- matrix transformations;
- probability simulations;
- PCA projections;
- activation/gradient distributions;
- attention visualizations;
- sampling behavior;
- scaling-law fits;
- quantization error and performance measurements.

The learner should understand notebook state, kernel restarts, clean execution, plotting, Markdown/math, reproducibility, and when reusable logic belongs in normal modules rather than notebook cells.

Fonzytooter should track notebook work but should **not embed or host Jupyter**.

### 3. Git repository + normal IDE — engineering, synthesis, and transfer

Substantial implementations, GPU-dependent experiments, multi-file work, and research artifacts belong in ordinary Git repositories.

Examples:

- neural network from scratch;
- CNN/sequence experiments;
- Transformer block implementation;
- TinyLM;
- open-model LoRA adaptation;
- grounded LLM-system study;
- inference/quantization profiling;
- paper reproduction and capstone research.

Fonzytooter may track the assignment, objective evidence, repository, status, and reflection. The code itself belongs in Git and a normal development environment.

## Canonical core architecture

The mandatory core contains Modules 00–23. Module count does **not** imply equal module length.

### Foundation

- **00 — Orientation**
- **01 — Scientific Python Foundations**

### First ML models and mathematical foundations

- **02 — Vectors, Matrices, and Numerical Data**
- **03 — Linear Regression: Your First ML Model**
- **04 — Derivatives and Gradient Descent**
- **05 — Probability and Statistical Thinking for ML**
- **06 — Classification and Logistic Regression**
- **07 — Generalization, Evaluation, and Good Experiments**

### Selective classical-ML breadth

- **08 — Classical ML Beyond Linear Models**
- **09 — Unsupervised Learning and Dimensionality Reduction**

### Neural networks from first principles

- **10 — Neural Networks as Composed Functions**
- **11 — Computational Graphs and Backpropagation**
- **12 — Neural Network From Scratch**
- **13 — PyTorch and Training Neural Networks**

### Deep learning and the road to attention

- **14 — Deep Learning Architectures: CNNs and Sequences**
- **15 — Embeddings and Language Modeling**
- **16 — Attention**

### Transformers and language models from scratch

- **17 — Transformer From Scratch**
- **18 — Tiny Autoregressive Language Model**

### Modern LLM lifecycle and systems

- **19 — From the Original Transformer to Modern LLMs**
- **20 — Fine-Tuning, LoRA, and Post-Training**
- **21 — Evaluation, Reasoning, Retrieval, and Tool Use**
- **22 — GPUs, Quantization, and LLM Inference**

### Research synthesis

- **23 — Research Engineering / Capstone**

The detailed teaching contract for these modules lives in the canonical syllabus and tranche documents rather than this overview.

## Mathematical progression

The course should feel cumulative rather than like a sequence of disconnected math mini-courses.

### Functions and notation

Begin in Module 01 and reuse continuously through parameterized models, composition, computational graphs, and neural architectures.

### Linear algebra

- Module 02: vectors, matrices, dot products, multiplication, transformations, geometry;
- Module 09: projection, covariance geometry, eigenvectors/eigenvalues for PCA;
- Modules 10–19: layers, embeddings, attention, tensor shapes, Transformer architecture;
- Module 20: low-rank structure through LoRA.

### Calculus and optimization

- Module 04: derivatives, partial derivatives, gradients, chain rule, gradient descent;
- Module 11: deeper chain-rule/backpropagation and reverse-mode differentiation;
- Module 13: stochastic optimization, momentum, Adam/AdamW, initialization, normalization, learning-rate schedules.

### Probability, statistics, and information theory

- Module 05: probability, random variables, expectation, variance, conditional probability, Bayes, sampling;
- Modules 06–07: likelihood, classification, softmax/cross-entropy, metrics, calibration, generalization;
- Module 15: sequence probability, NLL/perplexity, surprisal, entropy, cross-entropy, KL divergence;
- Modules 20–23: preference objectives, KL constraints, stochastic evaluation, uncertainty, effect-size reasoning.

### Numerical/systems reasoning

- Module 01: floating-point awareness and reproducible numerical work;
- Modules 06/15: stable sigmoid/softmax/log calculations;
- Module 13: FP32/FP16/BF16 and mixed precision;
- Module 22: memory arithmetic, KV cache, bit widths, quantization, bandwidth/throughput/latency.

## Major mastery projects

The course should contain a small number of substantial projects, with smaller labs and implementation checkpoints preparing for them.

The six major synthesis projects are:

1. **Classical ML experiment** — frame a problem, establish a baseline, apply appropriate preprocessing/modeling, evaluate honestly, and explain results.
2. **Neural Network From Scratch** — implement forward propagation, backpropagation, optimization, and evaluation without a deep-learning framework hiding the mechanics.
3. **TinyLM** — implement and train a small autoregressive Transformer language model and personally write the initial sampling path.
4. **Open-model adaptation/evaluation experiment** — adapt a manageable open model with LoRA/PEFT under a controlled evaluation.
5. **Inference systems experiment** — profile or benchmark a real inference question and explain the result using hardware, precision, memory, and algorithmic reasoning.
6. **Research capstone** — investigate or reproduce a tractable result with prior work, baselines, controlled experiments, uncertainty, reproducibility, limitations, and bounded conclusions.

Smaller activities such as the Module 07 trustworthiness study, attention calculation/implementation, Transformer-block build, architecture autopsy, and grounded-system study are labs or transfer checkpoints rather than additional capstones.

## Scope boundary: a deep primer, not an exhaustive degree

The mandatory core protects **foundational depth**, not exhaustive breadth.

The learner should leave the core able to understand unfamiliar ML/LLM ideas from known primitives and continue into specialist literature. The learner is not expected to become a specialist in every subfield before completing the course.

Topics deliberately left primarily for optional post-core paths include:

- advanced computer vision;
- diffusion, flow matching, and image/video generative modeling;
- reinforcement learning beyond the minimum foundation required for RLHF;
- advanced LLM post-training/alignment;
- multimodal ML;
- speech/audio ML;
- Bayesian/probabilistic ML;
- time-series ML;
- graph neural networks;
- causal inference;
- advanced inference/performance engineering, CUDA, Triton, and ML compilers;
- distributed training and large-scale serving systems;
- advanced mathematics/theory.

Optional paths should deepen the same mental model and experimental discipline rather than becoming disconnected topic catalogs.

## Source-of-truth rule

If planning documents disagree, use this order:

1. runtime-authored `curriculum/courses/ai-ml/` for content that actually exists;
2. [`ai-ml-syllabus-coherence.md`](ai-ml-syllabus-coherence.md) for final cross-tranche amendments;
3. [`ai-ml-syllabus.md`](ai-ml-syllabus.md) for the canonical core architecture;
4. detailed tranche plans for lesson-level intended design;
5. this document for high-level philosophy and boundaries.

The purpose of this overview is to keep the course's intent legible. It should not drift into a second competing syllabus again.
