# AI/ML syllabus coherence amendments

This document records the final cross-module coherence review of the AI/ML core syllabus after Modules 01–23 were expanded to lesson-level plans.

It is a **normative amendment** to the detailed tranche plans:

- [`ai-ml-modules-01-07.md`](ai-ml-modules-01-07.md)
- [`ai-ml-modules-08-14.md`](ai-ml-modules-08-14.md)
- [`ai-ml-modules-15-18.md`](ai-ml-modules-15-18.md)
- [`ai-ml-modules-19-23.md`](ai-ml-modules-19-23.md)

The canonical module architecture in [`ai-ml-syllabus.md`](ai-ml-syllabus.md) remains unchanged. Where this document conflicts with an earlier detailed tranche plan, **this document wins** and should be applied when runtime objectives, lessons, worksheets, exercises, notebooks, and projects are authored.

The review found no reason to add, remove, reorder, or split core modules. The 00–23 sequence is considered structurally complete. The amendments below close a small number of prerequisite, scope, and terminology gaps discovered only after reviewing the full course as one system.

## 1. Module 06 must teach multiclass softmax explicitly

The original detailed Module 06 plan teaches binary logistic regression and deliberately defers multiclass softmax. Later modules, however, legitimately need softmax before language modeling:

- multiclass neural-network outputs;
- CNN classification experiments;
- vocabulary distributions in language models;
- attention normalization.

Softmax therefore belongs in the first classification module rather than arriving implicitly later.

### Revised Module 06 sequence

Module 06 should contain **six** planned lessons/checkpoints rather than five:

1. **From Regression to Classification**
2. **Odds, Log-Odds, and the Sigmoid**
3. **Likelihood**
4. **Log Loss and Cross-Entropy**
5. **Logistic Regression From Scratch**
6. **From Binary to Multiclass: Softmax and Categorical Cross-Entropy**

The new Lesson 06 should cover:

- a vector of class logits rather than one binary logit;
- converting logits into a categorical probability distribution;
- the softmax function;
- invariance to adding the same constant to every logit;
- numerically stable softmax by subtracting the maximum logit;
- log-sum-exp intuition at the practical level needed to recognize stable implementations;
- categorical negative log-likelihood / cross-entropy;
- one-hot class targets and index-based class targets;
- binary sigmoid/BCE versus multiclass softmax/categorical cross-entropy;
- why softmax probabilities are coupled across classes;
- the distinction between logits, probabilities, and the final argmax/decision rule.

The learner should calculate a tiny softmax and categorical cross-entropy example by hand and implement a numerically stable softmax in an embedded exercise. A full multiclass classifier-from-scratch project is unnecessary; the existing binary logistic-regression implementation remains the foundational implementation checkpoint.

### Resulting prerequisite contract

After Module 06, later modules may assume the learner can:

- interpret a vector of logits;
- compute and reason about softmax probabilities;
- understand categorical cross-entropy;
- recognize stable softmax/log-sum-exp patterns.

The previous Module 06 deferral of multiclass softmax is superseded by this amendment.

## 2. Module 15 is the first real information-theory treatment

The high-level curriculum identifies information theory as part of the mathematical spine, but the detailed plans correctly avoid a detached information-theory course. The final review found that the course otherwise risks using cross-entropy repeatedly and then introducing KL divergence during post-training without ever establishing the relationships among these quantities.

Module 15 is the right place to fix this because language modeling makes probabilistic prediction and average log loss concrete.

### Expand Module 15 Lesson 04

The existing **What Is a Language Model?** lesson should add a compact, motivated information-theory layer after sequence probability and negative log-likelihood are established.

Teach:

- **surprisal / information content** as `-log p(x)`: unlikely events carry more information under a model;
- **entropy** as expected surprisal under the true/reference distribution;
- **cross-entropy** as expected surprisal when predictions come from another distribution/model;
- **KL divergence** as the excess cross-entropy incurred by using `q` when data are distributed as `p`;
- the relationship `H(p, q) = H(p) + D_KL(p || q)` at the conceptual/calculational level;
- why minimizing cross-entropy with respect to model parameters also minimizes KL to the target distribution when the target entropy is fixed;
- why KL is asymmetric;
- why these quantities are expectations over distributions rather than mystical measures of model "understanding."

Use tiny categorical distributions that can be calculated by hand. Base-2 versus natural logarithms may be mentioned, but coding theory, channel capacity, mutual information derivations, and formal information theory remain optional/specialist material.

### Consequence for later modules

Module 20 may now treat KL divergence in RLHF/reference-policy constraints as a **revisit in a new application**, not the learner's first encounter with the concept.

Module 15 mastery should include the ability to explain the practical relationship among negative log-likelihood, cross-entropy, entropy, perplexity, and KL divergence without requiring advanced information-theory proofs.

## 3. Module 13 needs AdamW and learning-rate scheduling

The detailed PyTorch module currently progresses from SGD to momentum to Adam and separately discusses L2/weight decay. Modern deep-learning and Transformer training code very commonly uses AdamW and explicit learning-rate schedules; leaving those out would make later training configurations contain unexplained machinery.

No new module or standalone optimization lesson is required.

### Expand Module 13 Lesson 05

The optimizer lesson should become conceptually:

**SGD, Momentum, Adam, and AdamW**

Add:

- Adam's adaptive moment estimates as already planned;
- the difference between adding an L2 penalty to the objective and applying decoupled weight decay under an adaptive optimizer;
- AdamW as the representative modern decoupled-weight-decay formulation;
- optimizer state and its memory implications at a basic level;
- why optimizer choice does not eliminate the need to choose and schedule a learning rate.

The learner does not need a convergence proof or optimizer catalog.

### Add learning-rate schedules to Module 13 training practice

Across Lessons 05–07, include:

- constant learning rate as a baseline;
- warmup and why the earliest optimization steps can be unusually fragile;
- decay after warmup;
- cosine decay as a representative commonly encountered schedule;
- schedule configuration as part of the experiment record;
- plotting learning rate beside training/validation curves where useful.

The objective is recognition and experimental understanding, not memorizing every scheduler API.

### Resulting mastery addition

After Module 13, the learner should be able to inspect a normal modern training configuration containing AdamW, weight decay, warmup, and a decay schedule and explain what each setting is trying to control.

## 4. Preserve the symbolic-AI / connectionist historical tension around Module 10

The detailed neural-network plan preserved McCulloch–Pitts, Rosenblatt, perceptron limitations, and early connectionist optimism, but the whole-course historical spine also intends to teach the parallel symbolic-AI tradition and the fact that AI winters had multiple causes.

Module 10's historical context should therefore explicitly connect:

- Turing and early machine-intelligence questions;
- Dartmouth and the emergence of AI as a named field;
- symbolic reasoning/search as a dominant early paradigm;
- connectionist/perceptron approaches as a competing tradition;
- the real representational limitations of single-layer perceptrons;
- expert systems and the later commercial success of symbolic knowledge-based AI;
- brittleness, knowledge-engineering cost, compute/data constraints, inflated expectations, and funding cycles as parts of the AI-winter story;
- the later statistical/neural resurgence as enabled by algorithms, data, compute, and engineering rather than one isolated theoretical breakthrough.

This history should remain integrated into technical lessons. Do **not** add a separate history-survey module or imply that one book, one failed technique, or one funding decision single-handedly caused an AI winter.

The recurring course-level tension should remain visible afterward:

```text
hand-authored rules / symbolic structure
        versus
learned statistical representations
```

Module 21 can then revisit a modern version of that tension when learned language models are combined with deterministic retrieval and tools.

## 5. Clarify project hierarchy across the course

The detailed planning process produced many valuable labs, repository exercises, foundational implementations, and transfer checkpoints. Not all of them should be described as "major projects" or "capstones."

The course should preserve a small number of substantial synthesis projects while allowing bounded studies between them.

### Major core projects/checkpoints

These are the six large synthesis projects identified by the canonical syllabus:

1. **Classical ML experiment** — Modules 07–09 progression, culminating around Module 09.
2. **Neural Network From Scratch** — Module 12.
3. **TinyLM** — Module 18.
4. **Open-model adaptation/evaluation experiment** — Module 20.
5. **Inference systems experiment** — Module 22.
6. **Research-engineering capstone** — Module 23.

These should receive the strongest transfer expectations, repository/report requirements, and multi-objective mastery evidence.

### Bounded labs and transfer assessments

Other planned work remains important but should be scoped and named accordingly. Examples include:

- Module 07 **Can You Trust This Model?** — a bounded first transfer study, **not** the course capstone;
- Module 13–14 framework deep-learning experiment;
- Module 16 hand-calculation + attention implementation checkpoint;
- Module 17 Transformer-block foundational implementation;
- Module 19 architecture autopsy;
- Module 21 grounded LLM-system study.

During content authoring, prefer terms such as **lab**, **study**, **implementation checkpoint**, or **transfer assessment** for these activities unless their scope truly rises to the level of the six major projects.

### Specific naming correction

The phrase "Jupyter/repository capstone" in the Module 07 detailed plan should be interpreted as **Jupyter/repository transfer study**. The activity itself remains valuable and unchanged; only its place in the project hierarchy changes.

## 6. Multimodality and full reinforcement learning remain post-core specializations

The final detailed syllabus intentionally narrowed the mandatory core. The following do **not** need to be pulled back into Modules 19–23 merely because they are important modern AI topics:

- multimodal model architecture/training in depth;
- diffusion/image/video generation;
- full reinforcement-learning foundations;
- Q-learning, Bellman equations, actor-critic families, general MDP theory;
- advanced CUDA/Triton/compiler work;
- distributed training and multi-node serving;
- exhaustive post-training/alignment method catalogs.

Module 20 teaches only the reinforcement-learning concepts required to understand RLHF-style post-training. Module 21 may mention multimodal and reasoning-system developments historically where useful, but multimodal ML is not a mandatory core capability.

These topics remain available as optional specialization paths after Module 23.

## 7. Module architecture and prerequisite shape remain unchanged

No module-number changes are required.

The final core remains:

```text
00 Orientation
01 Scientific Python Foundations
02 Vectors, Matrices, and Numerical Data
03 Linear Regression: Your First ML Model
04 Derivatives and Gradient Descent
05 Probability and Statistical Thinking for ML
06 Classification and Logistic Regression
07 Generalization, Evaluation, and Good Experiments
08 Classical ML Beyond Linear Models
09 Unsupervised Learning and Dimensionality Reduction
10 Neural Networks as Composed Functions
11 Computational Graphs and Backpropagation
12 Neural Network From Scratch
13 PyTorch and Training Neural Networks
14 Deep Learning Architectures: CNNs and Sequences
15 Embeddings and Language Modeling
16 Attention
17 Transformer From Scratch
18 Tiny Autoregressive Language Model
19 From the Original Transformer to Modern LLMs
20 Fine-Tuning, LoRA, and Post-Training
21 Evaluation, Reasoning, Retrieval, and Tool Use
22 GPUs, Quantization, and LLM Inference
23 Research Engineering / Capstone
```

Module sequence is recommended pedagogical order, not a claim that every objective depends on every previous module. Runtime objective metadata should continue to encode real capability prerequisites rather than the module sequence mechanically.

## 8. Revised cross-course mathematical spine

With the amendments above, the mathematical story is complete without adding detached prerequisite courses:

| Modules | Mathematical layer |
| --- | --- |
| 01 | functions, mappings, notation, numerical experimentation |
| 02 | vectors, matrices, dot products, matrix multiplication, transformations |
| 03 | indexed data, summation, means, residuals, squared loss |
| 04 | derivatives, partial derivatives, gradients, chain rule, gradient descent |
| 05 | probability, random variables, expectation, variance, conditional probability |
| 06 | exponentials/logarithms, likelihood, sigmoid, binary loss, **softmax and categorical cross-entropy** |
| 07 | generalization, regularization, metrics, calibration, experimental reasoning |
| 09 | projection, covariance geometry, eigenvectors/eigenvalues for PCA |
| 10–12 | nonlinear composition, deeper chain rule, reverse-mode differentiation, backpropagation |
| 13 | stochastic optimization, Adam/AdamW, initialization, normalization, learning-rate schedules, numerical precision |
| 14 | convolution arithmetic, recurrence, repeated-gradient behavior, gating |
| 15 | sequence probability, NLL/perplexity, **surprisal, entropy, cross-entropy, KL divergence** |
| 16–18 | attention mathematics, Transformer tensor shapes, categorical sampling |
| 19 | scaling relationships, parameter/tensor arithmetic |
| 20 | low-rank updates, preference/ranking objectives, KL/policy-probability revisit |
| 21 | repeated-sampling uncertainty, retrieval/ranking metrics, vector similarity |
| 22 | memory/throughput arithmetic, bits/dtypes, KV cache, quantization |
| 23 | repeated-run statistics, practical confidence intervals, effect-size reasoning |

Integration, formal real analysis, full matrix calculus, measure-theoretic probability, full information theory, and graduate-level statistics remain optional unless a later specialization creates a concrete need.

## 9. Final syllabus sign-off

After these amendments, the mandatory core should be treated as **syllabus-complete**.

Future work should happen in separate content-authoring threads and convert this plan into runtime curriculum. Authoring may still discover that one lesson should split, combine, or change medium, but those are implementation-level curriculum decisions. They should not reopen the 00–23 architecture without a concrete prerequisite or scope failure.

A learner who completes the core should be able to reason across three levels:

1. **mathematical/mechanistic** — explain the operations, objectives, and learning mechanisms;
2. **implementation/systems** — build, adapt, run, profile, and debug manageable ML models and LLM systems;
3. **empirical/research** — design controlled experiments and determine whether evidence supports a claim.

That is the completion standard for the mandatory deep primer. Specialist paths deepen it rather than repair missing fundamentals.
