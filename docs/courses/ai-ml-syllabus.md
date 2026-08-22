# AI/ML core syllabus skeleton

This document defines the canonical **core teaching sequence** for Fonzytooter's AI & Machine Learning course.

It sits one level below [`ai-ml.md`](ai-ml.md). That document defines the course's long-range philosophy, learning model, environments, and broad subject areas. This document defines the mandatory deep-primer spine: the modules every learner on this course path should complete before choosing a specialization.

This is intentionally a **syllabus skeleton**, not a lesson specification. Exact objectives, lesson counts, worksheets, interactives, notebook labs, source assignments, and repository projects are authored later in `curriculum/courses/ai-ml/`.

## Scope guardrail

The core course should be deep without becoming a degree-equivalent survey of every branch of machine learning.

The mandatory path should:

- build the mathematics needed to reason about important ML ideas rather than treating formulas as incantations;
- teach foundational algorithms deeply enough that the learner can implement the important mechanics by hand;
- cover classical machine learning well enough to understand the broader field before neural networks dominate the course;
- build neural networks, attention, a Transformer, and a tiny autoregressive language model from understandable pieces;
- connect those foundations to modern LLM training, adaptation, evaluation, tooling, and inference systems;
- develop experimental and research habits throughout the course;
- leave specialist depth such as advanced reinforcement learning, diffusion models, distributed training, and specialized application domains for optional follow-on paths.

The thing to protect is **depth of foundations**, not exhaustive breadth. If future syllabus work grows a module merely to include another named technique, architecture, or theorem, that material should have to justify its place by enabling a later core objective.

## Learner starting point

The intended learner is already an experienced professional software engineer but does not have a formal university foundation in calculus, linear algebra, probability/statistics, optimization, or related ML mathematics.

Programming instruction should therefore focus on scientific Python, numerical reasoning, and ML implementation rather than introductory programming concepts. Mathematics should be introduced just in time, practiced explicitly, and revisited at increasing depth as later ML topics require it.

## Core sequence

The recommended teaching order is Modules 00 through 23. The numbering is a curriculum sequence, not a claim that every earlier module is a hard prerequisite for every later one. Runtime objective metadata should encode only real prerequisite relationships.

### Foundation

#### 00 — Orientation

**Purpose:** Establish how Fonzytooter teaches, how mastery evidence works, and when to use embedded exercises, worksheets, Jupyter, repositories, and AI assistance.

Core ideas:

- self-paced progress and objective-based mastery;
- the roles of explanation, practice, experimentation, and projects;
- expectations for personally written foundational implementations;
- reproducibility, reflection, and evidence of understanding.

This module already exists and should remain lightweight.

#### 01 — Scientific Python Foundations

**Purpose:** Make Python and NumPy comfortable enough for numerical experimentation without turning the course into general Python application training.

Core ideas:

- Python's execution and object model for an experienced programmer;
- programming functions and mathematical functions;
- NumPy arrays, shapes, dimensions, dtypes, indexing, and storage intuition;
- vectorized operations and broadcasting;
- Jupyter notebooks, plotting, kernels, and reproducible notebook habits;
- packages/environments and a disciplined scientific-computing workflow;
- basic floating-point and numerical-computing awareness.

Representative work:

- embedded Python/NumPy exercises;
- short numerical experiments;
- first Jupyter notebooks and plots.

**Deliberately deferred:** Python web development, deep OOP patterns, framework architecture, and other Python topics that do not support the ML path.

### First models and just-in-time mathematics

#### 02 — Vectors, Matrices, and Numerical Data

**Purpose:** Build the first layer of linear algebra by connecting mathematical objects directly to arrays, data, geometry, and transformations.

Core ideas:

- scalars, vectors, matrices, dimensions, and coordinates;
- vector addition and scalar multiplication;
- dot products and geometric intuition;
- matrix-vector and matrix-matrix multiplication;
- matrices as transformations;
- affine transformations;
- norms and distance at an introductory level;
- translating between notation and NumPy.

Representative work:

- hand matrix arithmetic worksheets;
- geometric interactives;
- NumPy implementations of core operations;
- notebooks visualizing transformations.

**Deliberately deferred:** abstract vector-space proofs, determinants as a major topic, eigenvectors, decompositions, and other advanced linear algebra until an ML problem motivates them.

#### 03 — Linear Regression: Your First ML Model

**Purpose:** Introduce machine learning through the simplest useful trainable model and establish the recurring pattern of data, parameters, predictions, error, and fitting.

Core ideas:

- supervised learning, features, targets, and parameters;
- linear and affine models;
- predictions and residuals;
- mean squared error;
- loss/objective functions;
- fitting as choosing parameters that reduce loss;
- loss surfaces and the distinction between model and training algorithm.

Representative work:

- fit tiny models by hand;
- visualize lines and loss surfaces;
- personally implement prediction and MSE;
- personally implement a small linear-regression training experiment before relying on scikit-learn.

Historical thread: least squares and the long prehistory of statistical fitting before modern machine learning.

#### 04 — Derivatives and Gradient Descent

**Purpose:** Introduce calculus because the learner now has a concrete reason to ask how changing a parameter changes model error.

Core ideas:

- change, slope, and derivative intuition;
- limits only to the depth needed to make derivatives meaningful;
- derivative rules for common functions;
- partial derivatives;
- gradients;
- the chain rule at an introductory level;
- gradient descent;
- learning rate and convergence behavior;
- numerical versus analytic derivatives.

Representative work:

- derivative and gradient worksheets;
- tangent/slope interactives;
- personally implement gradient descent;
- notebooks comparing learning rates and visualizing optimization paths.

**Deliberately deferred:** integration as a general calculus topic, formal epsilon-delta analysis, and advanced optimization theory.

#### 05 — Probability and Statistical Thinking for ML

**Purpose:** Build the probability foundation needed to reason about uncertain data and probabilistic predictions.

Core ideas:

- events and probability;
- random variables;
- common discrete and continuous distributions at a useful conceptual level;
- expectation, variance, and standard deviation;
- covariance and correlation intuition;
- conditional probability;
- Bayes' rule;
- sampling and empirical estimates;
- population versus sample;
- uncertainty and noisy observations.

Representative work:

- probability worksheets;
- distribution/sampling interactives;
- simulation notebooks that connect formulas to repeated experiments.

**Deliberately deferred:** measure-theoretic probability, exhaustive distribution catalogs, and formal mathematical statistics.

#### 06 — Classification and Logistic Regression

**Purpose:** Extend the first-model story from predicting quantities to predicting classes and probabilities.

Core ideas:

- classification versus regression;
- logits, odds, and log-odds;
- exponentials and logarithms where they become useful;
- sigmoid and probabilistic binary outputs;
- likelihood and maximum-likelihood intuition;
- binary cross-entropy/log loss;
- decision boundaries;
- thresholding and the difference between scores, probabilities, and decisions;
- multiclass logits and categorical distributions;
- softmax;
- categorical cross-entropy/negative log-likelihood;
- numerically stable softmax and practical log-sum-exp intuition.

Representative work:

- hand sigmoid/log-loss exercises;
- decision-boundary interactives;
- personally implement logistic regression and its training loop;
- calculate a tiny multiclass softmax/categorical-cross-entropy example;
- personally implement numerically stable softmax;
- compare the manual logistic-regression implementation with scikit-learn.

Cross-entropy terminology begins here operationally. Entropy and KL divergence receive their first explicit, compact information-theory treatment in Module 15 when language modeling gives those quantities a concrete probabilistic context.

#### 07 — Generalization, Evaluation, and Good Experiments

**Purpose:** Teach how to know whether a model actually learned something useful rather than merely fitting the examples it saw.

Core ideas:

- training, validation, and test data;
- overfitting and underfitting;
- generalization;
- bias/variance intuition;
- regularization at a practical introductory level;
- cross-validation;
- data leakage;
- class imbalance;
- regression and classification metrics;
- precision, recall, F1, ROC/PR intuition where appropriate;
- calibration;
- baselines and experimental controls;
- randomness, seeds, and reproducibility.

Representative work:

- diagnose intentionally flawed experiments;
- notebooks comparing metrics and regularization behavior;
- a bounded evaluation-focused transfer study whose main goal is experimental discipline rather than model novelty.

### Classical ML breadth without a survey-course detour

#### 08 — Classical ML Beyond Linear Models

**Purpose:** Show that useful ML is not synonymous with differentiable neural networks and develop intuition for alternative ways models can partition and fit data.

Core ideas:

- decision trees and recursive splitting;
- impurity and split quality conceptually;
- random forests and bagging;
- boosting at the level needed to understand why ensembles work;
- nearest-neighbor methods and similarity-based prediction;
- feature scaling where distance matters;
- scikit-learn pipelines and model comparison.

Representative work:

- trace a small decision tree by hand;
- visualize decision boundaries;
- compare linear, tree-based, and neighbor-based models on the same data.

**Deliberately deferred:** exhaustive ensemble variants, a full derivation of every tree criterion, and a survey of every classical estimator in scikit-learn.

#### 09 — Unsupervised Learning and Dimensionality Reduction

**Purpose:** Introduce learning without labeled targets and revisit linear algebra when higher-dimensional data makes projections and representations useful.

Core ideas:

- supervised versus unsupervised objectives;
- distance and similarity;
- k-means and clustering intuition;
- personally implement a small k-means loop;
- covariance geometry;
- projections and orthogonality;
- eigenvectors/eigenvalues at the depth needed for PCA;
- principal component analysis;
- high-dimensional intuition and visualization limits.

Representative work:

- clustering exercises and notebooks;
- PCA visualizations;
- a classical-ML synthesis project using appropriate preprocessing, model selection, and held-out evaluation.

**Deliberately deferred:** spectral methods, advanced matrix decompositions, manifold learning catalogs, and deep unsupervised learning.

### Neural networks from first principles

#### 10 — Neural Networks as Composed Functions

**Purpose:** Make neural networks feel like a natural extension of functions and linear models rather than a new kind of magic.

Core ideas:

- historical perceptron;
- neurons as affine transformations plus nonlinearities;
- activation functions;
- layers and multilayer perceptrons;
- forward passes;
- parameter tensors and shapes;
- composition of functions;
- representation learning intuition.

Representative work:

- hand forward-pass worksheets;
- network-shape interactives;
- embedded exercises implementing neurons and small layers with NumPy.

Historical thread: Turing/Dartmouth context, McCulloch-Pitts, the perceptron, symbolic and connectionist traditions, expert systems, multifactor AI winters, and the later statistical/neural resurgence. The point is to explain competing approaches and changing constraints, not repeat a one-book or one-failure myth about neural-network history.

#### 11 — Computational Graphs and Backpropagation

**Purpose:** Revisit the chain rule at the depth required to understand how neural-network gradients are actually computed.

Core ideas:

- computational graphs;
- local derivatives;
- partial derivatives through composed functions;
- reverse-mode differentiation;
- backpropagation;
- gradient accumulation;
- why automatic differentiation works conceptually.

Representative work:

- substantial hand-worked backpropagation worksheets;
- computational-graph interactives;
- personally implement backward passes for a small set of operations.

Historical thread: the development and rediscovery/popularization of backpropagation and its role in the revival of neural networks.

#### 12 — Neural Network From Scratch

**Purpose:** Consolidate the previous modules by building a small trainable neural network without a deep-learning framework hiding the core mechanics.

Core ideas:

- layer abstractions;
- forward and backward passes;
- parameter initialization;
- loss calculation;
- minibatches at a simple level;
- gradient-based updates;
- training and evaluation loops;
- gradient checking and debugging.

Representative work:

- major learner-written repository project using NumPy or similarly transparent numerical tools;
- agents may review, test, debug, and refactor after the learner has produced the foundational implementation.

This project is a core mastery checkpoint.

#### 13 — PyTorch and Training Neural Networks

**Purpose:** Introduce modern framework abstractions only after the learner understands the mechanics they automate.

Core ideas:

- tensors, shapes, dtypes, and devices;
- autograd;
- `nn.Module` and parameters;
- datasets, dataloaders, and minibatches;
- training/evaluation loops;
- SGD, momentum, Adam, and AdamW;
- L2 regularization versus decoupled weight decay;
- learning-rate warmup and representative decay schedules;
- initialization and normalization;
- regularization;
- checkpoints, including optimizer/scheduler state where relevant;
- mixed precision and numerical stability at an introductory level;
- reproducible experiment organization.

Representative work:

- rebuild a previous small model in PyTorch;
- notebooks probing optimizer, schedule, and initialization behavior;
- explain what `backward()`, optimizer steps, weight decay, and learning-rate scheduling are doing beneath the framework abstraction.

### Deep learning beyond dense networks

#### 14 — Deep Learning Architectures: CNNs and Sequences

**Purpose:** Give meaningful breadth beyond MLPs and create the historical/technical runway for attention without turning the core into separate full courses on vision and recurrent networks.

Core ideas:

- convolution, locality, and weight sharing;
- feature hierarchies and CNN intuition;
- one meaningful image-model experiment;
- sequence data and recurrent hidden state;
- RNNs conceptually and computationally;
- vanishing/exploding gradient motivation;
- LSTM/GRU ideas without exhaustive implementation detail;
- encoder-decoder/sequence-to-sequence models;
- the fixed-context bottleneck that motivated attention.

Historical thread: ImageNet-era deep learning, learned representations, recurrent sequence modeling, and seq2seq.

**Deliberately deferred:** advanced computer-vision architectures, object detection/segmentation, exhaustive recurrent-cell variants, and specialist vision/NLP practice.

#### 15 — Embeddings and Language Modeling

**Purpose:** Establish how discrete symbols become learnable numerical representations, what it means to train a model to predict language, and the compact information-theory foundation needed to interpret language-model objectives.

Core ideas:

- categorical data and one-hot representations;
- learned embeddings;
- similarity in embedding spaces;
- tokens, vocabularies, and token IDs;
- tokenization concepts;
- n-gram and simple neural language-model context;
- next-token prediction;
- logits, softmax, probability distributions, and cross-entropy revisited;
- autoregressive factorization conceptually;
- surprisal/information content;
- entropy as expected surprisal;
- cross-entropy as expected surprisal under a model;
- KL divergence as excess cross-entropy and its asymmetry;
- the relationship among negative log-likelihood, cross-entropy, KL divergence, and perplexity.

Representative work:

- embedding visualizations;
- simple language-model notebooks;
- inspect how tokenization changes model inputs;
- hand-calculate surprisal, entropy, cross-entropy, and KL for tiny categorical distributions.

Historical thread: distributed representations, word embeddings, and the transition from feature engineering toward learned representations.

### Attention and Transformers

#### 16 — Attention

**Purpose:** Teach attention as a solution to a concrete sequence-modeling problem before presenting the Transformer architecture.

Core ideas:

- alignment and weighted context;
- queries, keys, and values;
- dot-product similarity;
- scaling;
- softmax weighting;
- masks;
- attention matrices;
- self-attention versus cross-attention.

Representative work should deliberately use all four learning modes:

- hand-calculate a tiny attention example;
- manipulate an attention interactive;
- personally implement attention in an embedded exercise or small script;
- visualize attention behavior in Jupyter.

Historical thread: seq2seq attention and the path from recurrent attention mechanisms to self-attention.

#### 17 — Transformer From Scratch

**Purpose:** Assemble attention and ordinary neural-network components into a complete Transformer block that the learner can explain operation by operation.

Core ideas:

- self-attention;
- causal masking;
- multi-head attention;
- residual connections;
- normalization;
- position-wise feed-forward networks;
- positional information;
- tensor shapes through a block;
- stacking blocks.

Representative work:

- personally implement a Transformer block before relying on a high-level Transformer library;
- shape-tracing worksheets and exercises;
- notebook visualizations of intermediate representations.

Historical thread: *Attention Is All You Need* and why removing recurrence changed sequence-model scaling and parallelism.

#### 18 — Tiny Autoregressive Language Model

**Purpose:** Turn the Transformer block into a complete small GPT-style model that can be trained, evaluated, and sampled from end to end.

Core ideas:

- context windows;
- token and positional representations;
- stacked Transformer blocks;
- output logits;
- cross-entropy training objective;
- batches of token sequences;
- autoregressive training loops;
- validation loss;
- sampling and decoding;
- temperature and top-k/top-p intuition;
- generation failure modes.

Representative work:

- major learner-written TinyLM repository project;
- personally write the first training loop and sampling path;
- train a genuinely small model and inspect what it learns.

This is a second major core mastery checkpoint.

### From Transformers to modern LLMs

#### 19 — From the Original Transformer to Modern LLMs

**Purpose:** Explain the most important architectural and training changes that connect the learner's tiny model to contemporary open and frontier language models without becoming a catalog of every published variant.

Core ideas:

- pretraining data and next-token objectives at scale;
- dataset quality and filtering;
- compute/data tradeoffs and scaling-law intuition;
- modern normalization such as RMSNorm;
- rotary positional embeddings;
- gated feed-forward layers such as SwiGLU-like designs;
- grouped-query/multi-query attention;
- mixture-of-experts conceptually;
- long-context challenges and representative approaches;
- why architectural changes often target quality, stability, memory, or throughput.

The selection rule for this module is problem-oriented: each modern technique should be taught because it solves an important limitation, not merely because it has a name.

#### 20 — Fine-Tuning, LoRA, and Post-Training

**Purpose:** Show how a pretrained model is adapted into a useful assistant or specialized model and provide only the reinforcement-learning foundation needed to understand modern post-training.

Core ideas:

- pretraining versus supervised fine-tuning;
- full fine-tuning versus parameter-efficient adaptation;
- low-rank intuition and LoRA;
- personally implement the core LoRA mechanism before using a PEFT library;
- instruction data;
- preferences and reward signals;
- minimal RL vocabulary: state, action, reward, policy, return, value;
- policy-gradient intuition;
- reward models conceptually;
- RLHF and why PPO-style constraints/KL penalties are used;
- DPO as a representative direct preference-optimization method;
- the distinction between training-time and inference-time behavior shaping.

Representative work:

- fine-tune a small open model with LoRA using standard tooling after the mechanism is understood;
- compare base and adapted model behavior with a controlled evaluation set.

**Deliberately deferred:** a full reinforcement-learning curriculum, deep PPO derivations, advanced offline RL, constitutional/self-improvement methods as a major unit, large-scale synthetic/preference-data pipelines, and exhaustive post-training algorithm catalogs. Those belong in a specialization if desired.

#### 21 — LLM Evaluation, Reasoning, Retrieval, and Tool Use

**Purpose:** Teach how modern language-model systems are evaluated and how models are combined with context, computation, and deterministic tools without turning the course into an SDK tutorial.

Core ideas:

- evaluation sets and benchmark design;
- contamination and benchmark validity;
- automatic versus human evaluation;
- task-specific metrics and rubric-based evaluation;
- calibration, uncertainty, robustness, and failure analysis;
- inference-time/test-time compute concepts;
- retrieval-augmented generation as a system pattern;
- tool calling and structured interfaces;
- agent loops and orchestration at a conceptual/system-design level;
- learned models combined with deterministic software.

Representative work:

- design a small evaluation harness;
- compare prompting/retrieval/tool-use variants under controlled conditions;
- write conclusions supported by evidence rather than anecdotes.

**Deliberately deferred:** vendor-specific API walkthroughs and exhaustive agent-framework coverage.

### Hardware and inference systems

#### 22 — GPUs, Quantization, and LLM Inference

**Purpose:** Give the learner a practical mental model for why ML workloads behave the way they do on real hardware and why local/server inference systems make particular tradeoffs.

Core ideas:

- CPU versus GPU execution models;
- parallelism and tensor operations;
- memory hierarchy and bandwidth intuition;
- FLOPs versus memory movement;
- tensor layout at a useful conceptual level;
- floating-point formats and reduced precision;
- numerical range, overflow/underflow, and stability;
- model weights, activations, and memory use;
- inference prefill versus decode;
- KV cache;
- batching;
- latency versus throughput;
- quantization and quantization error;
- context length and memory cost;
- speculative decoding and continuous batching conceptually;
- distributed inference/training only to the level needed to understand why multiple accelerators are used.

Representative work:

- profile real tensor/model operations;
- estimate memory requirements;
- compare model precision/quantization configurations;
- benchmark inference and explain measured behavior from hardware principles.

**Deliberately deferred:** CUDA/Triton kernel engineering, compiler internals, and full distributed-training architecture unless pursued later as a systems specialization.

### Research synthesis

#### 23 — Research Engineering and Capstone

**Purpose:** Remove most of the instructional scaffolding and require the learner to investigate a real question using the habits developed throughout the course.

Core ideas:

- reading papers critically;
- identifying a paper's actual claim, method, assumptions, and evidence;
- turning claims into reproducible experiments;
- baselines and controlled comparisons;
- ablations;
- interpreting negative results;
- reproducibility and experiment logs;
- technical writing;
- distinguishing evidence from speculation;
- deciding what to learn next from the literature.

Representative work:

- reproduce or partially reproduce a tractable published result, or conduct a similarly rigorous original small experiment;
- modify one meaningful variable and form a hypothesis about the result;
- document method, evidence, limitations, failures, and conclusions in a technical report.

Research practice should **not begin here**. Earlier modules should progressively introduce historical papers, figures, method sections, small reproductions, controlled experiments, and written conclusions. Module 23 is the point where those habits become an independent project.

## Dependency shape

The recommended reading/teaching sequence is mostly linear because a single clear path is easier for a learner to follow. The conceptual prerequisite graph is less rigid.

A simplified view is:

```text
00 Orientation
  -> 01 Scientific Python
      -> 02 Vectors & Matrices
          -> 03 Linear Regression
              -> 04 Derivatives & Gradient Descent
              -> 05 Probability
                  -> 06 Logistic Regression
                      -> 07 Generalization & Evaluation
                          -> 08 Classical ML
                          -> 09 Unsupervised Learning & PCA

02 + 04 + 06 + 07
  -> 10 Neural Networks
      -> 11 Backpropagation
          -> 12 Neural Network From Scratch
              -> 13 PyTorch & Training
                  -> 14 CNNs & Sequences
                  -> 15 Embeddings & Language Modeling
                      -> 16 Attention
                          -> 17 Transformer From Scratch
                              -> 18 TinyLM
                                  -> 19 Modern LLMs
                                      -> 20 Fine-Tuning & Post-Training
                                      -> 21 Evaluation, Retrieval & Tools
                                      -> 22 GPU & Inference Systems
                                          -> 23 Research Capstone
```

This diagram is guidance, not the runtime objective DAG. In particular:

- decision trees are not a mathematical prerequisite for neural networks;
- PCA is not a prerequisite for attention;
- a module may be recommended earlier for breadth even when later modules do not strictly depend on it;
- detailed syllabus work should avoid inventing false prerequisite edges simply to reproduce module order.

## Mathematical spine

Mathematics should spiral through the course instead of appearing as a separate prerequisite semester.

### Functions, algebra, and notation

Begin in Module 01 and reuse continuously. Exponentials and logarithms become important in Module 06 and recur in softmax, cross-entropy, probability, and language modeling.

### Linear algebra

- **Module 02:** vectors, matrices, dot products, multiplication, transformations, norms, geometry;
- **Module 09:** projections, covariance geometry, eigenvectors/eigenvalues, PCA;
- **Modules 10-18:** repeated application through layers, embeddings, attention, and Transformer tensor operations;
- **Module 20:** low-rank structure revisited through LoRA.

### Calculus and optimization

- **Module 04:** derivatives, partial derivatives, gradients, chain rule introduction, gradient descent;
- **Module 11:** chain rule and multivariable differentiation revisited through backpropagation;
- **Module 13:** optimization revisited through SGD, momentum, Adam/AdamW, decoupled weight decay, learning-rate warmup/decay, initialization, and training behavior.

Integration should be introduced later only if a core topic actually needs it. The course does not need a detached full calculus sequence before ML can begin.

### Probability and statistics

- **Module 05:** probability, random variables, distributions, expectation, variance, conditional probability, Bayes, sampling;
- **Modules 06-07:** likelihood, probabilistic classification, metrics, calibration, uncertainty, experimental design;
- **Modules 15-21:** probability distributions, language-model likelihood, sampling, evaluation, preference data, and uncertainty in modern model behavior.

### Information theory

- **Module 06:** cross-entropy is introduced operationally through binary and multiclass classification;
- **Module 15:** surprisal, entropy, cross-entropy, KL divergence, and their relationships receive the first explicit compact information-theory treatment in the concrete setting of language modeling;
- **Module 20:** KL divergence is revisited as a way to reason about reference-policy constraints in post-training.

Coding theory, channel capacity, mutual-information derivations, and other deeper information-theory topics remain optional unless a later specialization requires them.

### Numerical computing

Begin in Module 01 and revisit through:

- stable implementations of sigmoid, softmax, logs, and losses;
- floating-point behavior and reproducibility;
- mixed precision in PyTorch;
- precision, memory, and quantization in Module 22.

## Historical spine

History should explain why ideas appeared, what problem they were responding to, and how the field's assumptions changed. It should not become a detached chronology course.

Key placements include:

- early computing, Turing, and the emergence of AI during orientation/foundational context;
- statistical fitting and least squares with linear regression;
- McCulloch-Pitts, the perceptron, symbolic AI, expert systems, and AI winters around Module 10;
- backpropagation and the neural-network revival around Module 11;
- statistical ML and the shift toward data-driven methods through Modules 07-09;
- ImageNet and the deep-learning resurgence around Modules 13-14;
- word embeddings and neural language modeling around Module 15;
- seq2seq and attention around Modules 14-16;
- the Transformer around Module 17;
- GPT-style autoregressive scaling around Modules 18-19;
- scaling laws, instruction tuning, RLHF, and modern post-training around Modules 19-20;
- reasoning systems, retrieval, structured tool use, and agent-loop patterns around Module 21.

Multimodal ML remains an optional post-core specialization rather than a required Module 21 capability.

Recurring tensions worth revisiting include symbolic versus statistical systems, hand-engineered versus learned representations, specialized versus general models, and learned components versus deterministic tools.

## Foundational implementations the learner should write first

Before agents or high-level libraries take over the mechanics, the learner should personally produce the first meaningful implementation of:

- linear regression prediction/loss;
- gradient descent;
- logistic regression;
- numerically stable softmax;
- k-means;
- a small neural network;
- backpropagation for that network;
- attention;
- a Transformer block;
- an autoregressive language-model training loop;
- sampling/decoding;
- the core low-rank adaptation mechanism used by LoRA.

This does not mean every implementation must become production-quality software. The point is to cross the conceptual boundary personally once. After that, agents and libraries may be used normally for review, testing, refactoring, extension, and engineering work.

## Major mastery checkpoints

The core should contain a small number of substantial projects rather than a new large project for every topic.

1. **Classical ML experiment** — frame a problem, establish a baseline, split data correctly, compare appropriate models, evaluate honestly, and explain results.
2. **Neural network from scratch** — implement forward propagation, backpropagation, training, and evaluation without a deep-learning framework hiding the core mechanics.
3. **TinyLM** — implement and train a small autoregressive Transformer language model and write the first sampling path.
4. **Open-model adaptation/evaluation experiment** — adapt a small open model with LoRA and compare base/adapted behavior with a controlled evaluation process.
5. **Inference systems experiment** — profile or benchmark real inference configurations and explain the results using hardware, precision, memory, and algorithmic reasoning.
6. **Research capstone** — reproduce or rigorously investigate a tractable result and report method, evidence, failures, limitations, and conclusions.

Bounded labs and implementation/transfer checkpoints—including the Module 07 trustworthiness study, Module 13–14 framework experiment, Module 16 attention calculation/implementation, Module 17 Transformer-block build, Module 19 architecture autopsy, and Module 21 grounded-system study—remain important but are not additional major projects or capstones.

Smaller notebooks, worksheets, interactives, and embedded exercises should prepare for these checkpoints rather than compete with them.

## Learning-medium emphasis

Different topics should use different forms of practice.

- **Worksheets:** especially valuable for vectors/matrices, derivatives, probability, loss calculations, neural-network forward passes, backpropagation, and small attention calculations.
- **Interactives:** especially valuable for geometry, loss surfaces, gradient descent, probability distributions, decision boundaries, network shapes, attention weights, and sampling behavior.
- **Embedded Pyodide exercises:** small implementations and checks such as NumPy operations, losses, gradients, sigmoid/softmax, one optimization step, and small attention components.
- **Jupyter notebooks:** exploratory work, visualization, dataset inspection, training-behavior experiments, attention inspection, and quantization/profiling investigations.
- **Git repositories:** neural-network-from-scratch work, TinyLM, GPU/data-heavy experiments, open-model adaptation, inference benchmarking, and paper reproduction.

The same concept may appear in several media when each medium adds a distinct kind of understanding.

## Optional specialization paths after the core

The following are valuable subjects, but they should not expand the mandatory syllabus unless the core later develops a genuine dependency on them:

- advanced computer vision;
- diffusion, score-based, flow-matching, and other generative-model families;
- reinforcement learning beyond the post-training foundation;
- advanced LLM post-training and alignment;
- multimodal ML;
- speech and audio ML;
- Bayesian/probabilistic ML;
- time-series ML;
- graph neural networks;
- causal inference;
- advanced inference/performance engineering, CUDA, Triton, and ML compilers;
- distributed training and large-scale ML infrastructure;
- advanced mathematics for learners who want deeper theoretical study.

These are extensions of the foundation, not prerequisites for being able to say the learner has completed a serious AI/ML primer.

## Contract for detailed syllabus work

When a module is expanded from this skeleton, the detailed design should define:

- the module's purpose and why it occurs at that point in the sequence;
- real prerequisite objective IDs;
- learning objectives;
- ordered lessons;
- mathematics introduced and revisited;
- historical context that improves understanding;
- interactives;
- printable worksheets;
- embedded coding exercises;
- Jupyter labs;
- repository-based work where appropriate;
- what the learner should be able to do after the module;
- what is deliberately deferred.

Detailed design should be done incrementally. The architecture should be tested at lesson resolution on the early modules before every later module is exhaustively specified.

## Current authored state

At the time this skeleton was established, authored runtime curriculum already contains:

- Module 00, **Orientation**;
- Module 01, **Scientific Python Foundations**;
- Lesson 01, **Python for a Programmer**;
- Lesson 02, **Functions: Code and Mathematics**;
- initial worksheets supporting the first two Scientific Python lessons.

The next authored work should continue Module 01 rather than renumbering or replacing the material that already exists. The syllabus skeleton defines the destination; `curriculum/courses/ai-ml/` remains the source of truth for content that has actually been authored and is loadable by the application.