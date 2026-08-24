# AI/ML detailed syllabus plan: Modules 19–23

This document expands Modules 19 through 23 from the canonical [`ai-ml-syllabus.md`](ai-ml-syllabus.md) into a lesson-level teaching plan.

It is a **planning document**, not runtime-authored curriculum. Exact lesson files, objective metadata, worksheets, interactives, notebook repositories, sources, project specifications, datasets, and exercise definitions are authored later under `curriculum/courses/ai-ml/`.

The purpose of this final core tranche is to connect the learner's small, understandable Transformer and TinyLM to contemporary large-language-model practice without abandoning the first-principles approach established earlier in the course. The learner should understand what changes when language models become large, how post-training changes model behavior, how LLM systems are evaluated and grounded, why inference performance is constrained by hardware and numerical representation, and how to conduct a defensible research-engineering project.

This tranche completes the mandatory deep-primer spine. Specialist depth remains available through optional follow-on paths.

## Scope guardrails

These modules sit close to fast-moving industry practice, so they are especially vulnerable to turning into a catalog of fashionable techniques. The syllabus should instead organize modern methods around the problems they solve.

The design rules are:

- Module 19 should explain why important modern Transformer changes exist rather than enumerate model-family trivia;
- scaling should be treated as an empirical relationship among model size, data, compute, and performance, not as a mystical source of intelligence;
- pretraining data quality, filtering, deduplication, and mixture design are part of the model-development problem, not incidental plumbing;
- post-training should distinguish supervised adaptation, parameter-efficient adaptation, preference learning, and reinforcement-learning-style optimization;
- LoRA should be understood mathematically and implemented at small scale before it is treated as a library switch;
- reinforcement learning should be taught only to the depth needed to understand RLHF-style post-training in the core course;
- LLM evaluation should emphasize task definition, contamination, variance, baselines, human judgment, and failure analysis rather than benchmark leaderboard worship;
- retrieval and tool use should be taught as explicit system mechanisms with observable failure modes, not as evidence that a model has become an autonomous agent;
- reasoning should be discussed in terms of elicitation, inference-time computation, search, verification, and measurable task performance rather than assuming generated reasoning traces reveal the model's internal causal process;
- Module 21 should not become a prompt-engineering or agent-framework course;
- Module 22 should teach enough GPU, memory, precision, quantization, and inference architecture to reason about real performance without becoming a CUDA, compiler, or distributed-systems specialization;
- the final capstone should require a reproducible question, baseline, controlled experiment, evidence, and limitations, not performative novelty.

## Tranche overview

| Module | Planned lessons | Primary job |
| --- | ---: | --- |
| 19 — From the Original Transformer to Modern LLMs | 7 | Explain how scale, data, and architecture connect TinyLM to modern LLMs |
| 20 — Fine-Tuning, LoRA, and Post-Training | 7 | Understand how pretrained models are adapted into useful assistants and specialists |
| 21 — Evaluation, Reasoning, Retrieval, and Tool Use | 7 | Evaluate generative systems and understand grounding/reasoning/tool mechanisms |
| 22 — GPUs, Quantization, and LLM Inference | 8 | Understand the hardware and numerical constraints that govern inference |
| 23 — Research Engineering / Capstone | 7 | Conduct and communicate a reproducible ML experiment |

This is approximately 36 lessons or project-oriented checkpoints. Lesson count is a planning estimate, not a product requirement. Modules 20, 22, and 23 should contain substantial hands-on work and may ultimately be authored as a mixture of shorter explanations and longer experiment/project checkpoints.

## Dependency shape

Recommended teaching order remains Modules 19 through 23, but objective metadata should encode only real prerequisites.

```text
18 Tiny Autoregressive Language Model
              |
              v
19 Modern LLMs
        /            \
       v              v
20 Post-Training    22 Inference Systems
       |              |
       v              |
21 Evaluation /      |
   Retrieval / Tools |
        \            /
         \          /
          v        v
        23 Research Engineering
             Capstone
```

Important consequences:

- Module 19 genuinely depends on the Transformer and language-model foundations from Modules 15–18.
- Module 20 depends on modern language-model architecture/training concepts but its LoRA objective also reaches back to matrix multiplication and low-rank intuition from earlier linear algebra.
- Module 21 uses concepts from Module 20 when evaluating post-trained models, but retrieval objectives also depend directly on embeddings/similarity from Module 15 and need not logically depend on every post-training objective.
- Module 22 depends strongly on tensor/device/numerical concepts from Module 13 and autoregressive generation from Module 18. It does not require mastery of retrieval or tool use.
- Module 23 is a synthesis module. Individual capstone topics may draw more heavily from one branch than another, but the learner should enter it with the full core sequence completed.

## Learning-medium emphasis

| Module | Most important media |
| --- | --- |
| 19 | architecture/config inspection, notebooks, reading primary papers/excerpts, quantitative exercises |
| 20 | hand matrix reasoning, embedded LoRA implementation, fine-tuning notebooks/repository work |
| 21 | evaluation notebooks, retrieval/tool-use experiments, failure-analysis reports |
| 22 | quantitative worksheets, profiling notebooks, local/GPU inference experiments |
| 23 | paper reading, experiment repository, research log, reproducible report |

Later modules should rely less on worksheets than the mathematical foundation modules, but hand calculations remain useful where they expose otherwise invisible systems constraints: parameter memory, KV-cache size, quantization storage, LoRA parameter counts, throughput, and simple uncertainty estimates.

---

# 19 — From the Original Transformer to Modern LLMs

## Purpose

Explain what changes when the learner's TinyLM becomes a modern large language model. The key story is not merely "make the Transformer bigger." Modern LLMs emerge from the interaction of larger compute budgets, much larger and more carefully constructed datasets, empirical scaling behavior, architectural changes that improve stability or efficiency, and engineering choices that make training and inference feasible.

The learner should finish this module able to inspect a modern decoder-only model configuration and explain why most of its unfamiliar components exist.

## Prerequisites

Required capabilities:

- tokenization, embeddings, next-token prediction, and autoregressive factorization;
- self-attention, multi-head attention, causal masking, residual connections, normalization, and feed-forward networks;
- a personally implemented Transformer block and TinyLM;
- training/validation loss and generalization;
- PyTorch tensors, devices, mixed precision, and ordinary training loops.

## Proposed core objectives

Objective IDs are provisional until authored. Capabilities should cover:

- distinguish architectural scale from training-data and compute scale;
- explain the role of corpus composition, filtering, deduplication, contamination control, and data quality in pretraining;
- interpret scaling-law relationships and compute/data tradeoffs at a practical level;
- explain why common modern block changes such as pre-normalization/RMSNorm, gated feed-forward networks, and rotary position representations exist;
- distinguish ordinary multi-head attention from multi-query/grouped-query attention and explain the inference-memory motivation;
- explain mixture-of-experts as conditional computation and reason about total versus active parameters;
- inspect a contemporary model configuration and map unfamiliar terms back to known Transformer components;
- avoid treating any particular current architecture as the final form of the Transformer.

## Lesson sequence

### 01 — From TinyLM to Large Language Model

**Job:** Establish which dimensions actually change when a small educational model becomes an LLM.

Topics:

- parameter count versus architecture family;
- training-token count;
- context length;
- vocabulary and tokenizer choices;
- batch size and optimization scale;
- compute budget;
- training duration;
- data diversity;
- why a larger model trained on too little or poor-quality data is not automatically better;
- model capability as an empirical result of the whole training system rather than parameter count alone.

The learner should compare the rough dimensions of the TinyLM project with representative larger-model configurations without turning the lesson into a leaderboard.

### 02 — Pretraining Data Is Part of the Model

**Job:** Make the dataset pipeline conceptually central to language-model training.

Topics:

- raw text/code/data sources versus training examples;
- document extraction and normalization;
- language/source mixtures;
- quality filtering;
- near-duplicate and exact-duplicate removal;
- contamination between training data and evaluation sets;
- toxic, private, copyrighted, low-quality, or otherwise problematic data as data-governance concerns;
- mixture weighting and curriculum-like choices;
- tokenizer/data interactions;
- why "more tokens" and "better data" are distinct dimensions.

The course should discuss provenance and governance conceptually without becoming a legal-compliance course.

### 03 — Scaling Laws and Compute-Optimal Training

**Job:** Introduce scaling laws as empirical measurement tools for allocating finite training budgets.

Topics:

- power-law relationships conceptually;
- log-log plots;
- validation loss as a function of model size, data, and compute;
- diminishing returns;
- undertrained versus over-parameterized-for-data regimes;
- compute-optimal model/data tradeoffs;
- extrapolation and its risks;
- why scaling laws are empirical regularities, not guarantees of qualitative capability;
- using small experiments to estimate trends.

### Mathematics

This lesson should give enough mathematical treatment to read a simple relationship such as

```text
L(N) ≈ A N^-α + C
```

or a multi-variable scaling relation and reason about what the exponent means. The learner should practice logarithms and slopes on log-log axes rather than receive a detached power-law theory unit.

### 04 — The Modern Transformer Block

**Job:** Compare the learner's basic Transformer block with common modern decoder blocks and explain each change by function.

Topics:

- post-norm versus pre-norm placement;
- LayerNorm revisited;
- RMSNorm intuition;
- ordinary feed-forward networks versus gated variants;
- GLU-family intuition;
- SwiGLU as a representative modern gated feed-forward design;
- activation/parameter-count tradeoffs;
- training stability and optimization as architectural design pressures;
- residual-stream continuity.

The lesson should use side-by-side block diagrams and tensor-shape traces. It should not require deriving every normalization gradient.

### 05 — Position and Long Context

**Job:** Revisit positional information now that context length is an important practical constraint.

Topics:

- why self-attention itself has no notion of token order;
- learned/sinusoidal position approaches revisited;
- relative-position intuition;
- rotary positional embeddings (RoPE) as the central modern example;
- applying rotations to query/key representations conceptually;
- context-window limits;
- training context versus inference context;
- context extension/interpolation ideas at a survey level;
- lost-in-the-middle and effective-context limitations;
- quadratic attention cost as motivation for later efficiency work.

The objective is not to catalog every positional method. RoPE should be understood well enough that seeing it in a model implementation is no longer mysterious.

### 06 — Attention Efficiency: MHA, MQA, and GQA

**Job:** Show how inference constraints reshape attention architecture.

Topics:

- standard multi-head query/key/value projections;
- repeated key/value state across attention heads;
- multi-query attention;
- grouped-query attention;
- quality/efficiency tradeoffs;
- key/value state as an inference-memory concern;
- a preview of the KV cache, with the full inference treatment deferred to Module 22;
- why architectural choices can target serving cost even when the mathematical attention idea is unchanged.

The learner should reason through tensor shapes and approximate key/value storage for a tiny example.

### 07 — Mixture of Experts and Reading Modern Model Architectures

**Job:** Introduce conditional computation, then synthesize the module by reading a real model configuration as an engineer rather than memorizing model brands.

Topics:

- dense feed-forward layers versus expert banks;
- routing/gating;
- top-k expert selection;
- total parameters versus active parameters per token;
- load-balancing intuition;
- expert specialization claims versus what can actually be measured;
- communication/serving complications at a high level;
- recognizing architecture fields for hidden size, head count, grouped-query heads, normalization, RoPE, feed-forward width, expert count, active experts, vocabulary, and context length;
- asking "what problem does this component solve?"

The module should finish with an architecture-autopsy exercise using one or more open model configuration files or architecture diagrams.

## Mathematics

Module 19 introduces or revisits:

- power laws and log-log relationships;
- ratios and order-of-magnitude reasoning;
- parameter-count arithmetic;
- tensor-shape arithmetic;
- normalization statistics conceptually;
- rotational/relative-position intuition without a trigonometry detour;
- sparse/conditional selection in mixture-of-experts systems.

No new formal mathematics course is required.

## History

History should connect technical changes to scaling pressure:

- the original Transformer as an encoder-decoder sequence model;
- the rise of large-scale pretrained Transformers and decoder-only language models;
- empirical scaling-law work and compute/data allocation;
- architecture evolution driven by training stability, context length, and inference cost;
- mixture-of-experts as an older conditional-computation idea made newly relevant at large scale.

Do not turn the lesson into a chronology of GPT/version numbers or vendor announcements.

## Interactives

Strong candidates:

- **Scaling-law explorer** — vary model/data/compute assumptions and inspect diminishing returns on linear versus log axes;
- **Modern block comparer** — switch between basic post-norm Transformer and a representative modern pre-norm/RMSNorm/gated block;
- **RoPE visualizer** — show token-position-dependent rotation and resulting query/key geometry;
- **MHA/MQA/GQA shape explorer** — compare key/value head counts and state size;
- **MoE router explorer** — visualize token-to-expert routing and active versus total parameters.

## Worksheets

Useful worksheet topics:

- estimate parameter-count changes from width/depth changes;
- interpret simple scaling-law plots;
- compare training-data/model-size tradeoffs;
- trace modern block tensor shapes;
- compare MHA/MQA/GQA key/value storage for small dimensions;
- distinguish active and total parameters in an MoE layer;
- annotate a modern architecture diagram with the problem each component addresses.

## Embedded exercises

Candidates:

- compute RMS normalization for a tiny vector;
- implement a simplified gated feed-forward operation;
- implement or manipulate a tiny RoPE transformation;
- reshape key/value tensors for grouped-query attention;
- calculate expert routing masks for a toy MoE gate;
- parse a model configuration and compute selected derived quantities.

The module does not require a learner-written production MoE implementation.

## Jupyter lab

**Architecture autopsy**

Choose a small open decoder-only model whose configuration and implementation are inspectable. The learner should:

1. map its configuration fields to known Transformer concepts;
2. compute selected parameter counts;
3. identify its normalization, position, feed-forward, and attention variants;
4. estimate which choices primarily target training stability, representational capacity, or inference efficiency;
5. compare it with the learner's TinyLM;
6. document which differences are architectural and which are simply scale.

## Mastery expectations

After Module 19 the learner should be able to:

- explain why modern LLM development is a data/compute/architecture problem rather than merely "more parameters";
- read simple scaling-law evidence;
- explain common modern Transformer changes by the engineering or optimization problem they address;
- reason about MHA/MQA/GQA and MoE at the mechanism level;
- inspect an unfamiliar open-model configuration without treating its terminology as magic;
- separate enduring Transformer concepts from implementation fashions likely to change.

## Deliberately deferred

- full distributed pretraining systems;
- optimizer/state sharding and ZeRO/FSDP internals;
- tensor/pipeline/data parallelism in depth;
- kernel implementation of RoPE/attention/MoE;
- exhaustive architecture catalogs;
- alternative sequence architectures as a major branch;
- dataset licensing/legal analysis;
- advanced long-context research.

---

# 20 — Fine-Tuning, LoRA, and Post-Training

## Purpose

Explain how a pretrained next-token predictor becomes a useful specialist or assistant. The learner should distinguish pretraining from adaptation and post-training, understand supervised fine-tuning, personally implement the core LoRA mechanism, and acquire enough preference-learning and reinforcement-learning background to understand RLHF and modern direct-preference methods.

This module should make clear that "post-training" is not one algorithm. It is a family of data, optimization, and evaluation stages that alter model behavior after pretraining.

## Prerequisites

Required capabilities:

- pretrained autoregressive language models;
- Transformer parameter matrices and forward passes;
- cross-entropy training;
- gradient-based optimization and PyTorch autograd;
- train/validation/evaluation discipline;
- basic probability and expectation;
- modern LLM architecture concepts from Module 19.

## Proposed core objectives

Capabilities should cover:

- distinguish pretraining, continued pretraining, supervised fine-tuning, instruction tuning, preference optimization, and inference-time prompting;
- prepare and reason about supervised instruction/response training examples;
- explain catastrophic forgetting/over-adaptation risks and the role of validation data;
- distinguish full-parameter fine-tuning from parameter-efficient fine-tuning;
- derive the basic LoRA update and reason about rank, parameter count, and mergeability;
- personally implement a small LoRA layer/adapter before relying on a PEFT library;
- explain preference data and reward-model training conceptually;
- understand reward, policy, action, trajectory, expected return, and policy-gradient intuition deeply enough to follow RLHF;
- explain the purpose of PPO-style constraints/KL penalties in RLHF at a conceptual level;
- explain direct preference optimization as an alternative route from preference pairs to policy updates;
- design a small adaptation experiment with meaningful baseline and evaluation.

## Lesson sequence

### 01 — Pretraining Is Not an Assistant

**Job:** Motivate post-training by separating language modeling competence from desired interaction behavior.

Topics:

- next-token pretraining objective revisited;
- base-model completion behavior;
- prompting versus changing weights;
- domain adaptation;
- continued pretraining;
- supervised fine-tuning;
- instruction following;
- preference alignment;
- behavioral objectives that are not directly represented by raw next-token pretraining;
- the post-training stack as stages, not a single magic alignment step.

### 02 — Supervised Fine-Tuning and Instruction Data

**Job:** Show how ordinary supervised learning is reused to shape language-model behavior.

Topics:

- prompt/instruction/input/response formats;
- chat templates and role tokens conceptually;
- token-level targets and masking portions of the sequence where appropriate;
- teacher forcing revisited;
- dataset quality and diversity;
- domain-specific versus general instruction data;
- overfitting and catastrophic forgetting intuition;
- validation sets and task-specific evaluation;
- full-model fine-tuning mechanics.

The learner should be able to inspect a tokenized instruction example and identify which tokens contribute to the chosen loss.

### 03 — Full Fine-Tuning Versus Parameter-Efficient Adaptation

**Job:** Motivate PEFT from memory, compute, and storage constraints rather than introducing it as a trendy library feature.

Topics:

- which parameters receive gradients in full fine-tuning;
- optimizer-state and gradient memory costs;
- storing one full model per task;
- freezing base weights;
- adapters and low-dimensional updates conceptually;
- parameter-efficient fine-tuning as a family;
- when full fine-tuning may still be appropriate;
- adaptation capacity versus efficiency.

The course may mention prompt/prefix tuning and adapters, but LoRA is the technique to understand deeply.

### 04 — LoRA From First Principles

**Job:** Derive and implement the core low-rank adaptation mechanism.

Topics:

- a frozen weight matrix `W`;
- replacing a learned full update `ΔW` with a low-rank product;
- `W' = W + sBA` or equivalent convention;
- rank `r`;
- matrix shapes;
- initialization so the adapter initially produces no/limited update;
- scaling factor;
- selecting target matrices;
- training only adapter parameters;
- merging/unmerging adapters;
- parameter-count savings;
- why low rank is an empirical modeling assumption rather than a universal theorem about fine-tuning.

### Required learner-written implementation

Before using a high-level LoRA/PEFT library, the learner should personally implement a small LoRA-wrapped linear layer, verify that frozen base weights do not receive updates, train the adapter on a tiny task, and demonstrate that merging the adapter produces equivalent inference outputs within numerical tolerance.

### 05 — Preference Data and Reward Models

**Job:** Introduce the idea of learning from comparisons rather than a single target response.

Topics:

- demonstrations versus preferences;
- pairwise chosen/rejected responses;
- human or synthetic preference labels;
- preference noise and annotator disagreement;
- scoring responses;
- reward-model training conceptually;
- pairwise ranking loss intuition;
- reward hacking and proxy objectives;
- distribution shift between preference data and deployed behavior;
- why a learned reward model is itself an imperfect model.

### 06 — Reinforcement Learning for RLHF, Just Enough

**Job:** Build the minimum reinforcement-learning foundation needed to understand RLHF mechanically.

Topics:

- agent/policy/environment/action/reward terminology;
- trajectories/episodes in the language-generation setting;
- policy as a token distribution;
- expected return;
- exploration at a conceptual level;
- policy-gradient intuition: increase probability of actions associated with higher-than-expected reward;
- credit assignment at a high level;
- baseline/value-function intuition;
- PPO-style clipped updates conceptually;
- KL penalties/reference models as constraints against moving too far from a known policy;
- why RL training can be unstable and expensive.

The course should not expand into a general reinforcement-learning sequence. Markov decision-process theory, temporal-difference learning, Q-learning, actor-critic families, and advanced RL belong in an optional specialization.

### 07 — Direct Preference Optimization and the Post-Training Experiment

**Job:** Show that preference optimization can often be expressed without an explicit online RL loop, then synthesize the module experimentally.

Topics:

- policy/reference-model preference ratios conceptually;
- direct preference optimization (DPO) as a representative direct-preference method;
- contrast with reward-model-plus-PPO pipelines;
- the role of a reference policy;
- hyperparameter sensitivity and preference-data quality;
- evaluation before/after adaptation;
- capability tradeoffs and regressions;
- base model versus SFT versus preference-tuned comparisons;
- why post-training objectives change behavior rather than create arbitrary new knowledge on demand.

## Mathematics

Module 20 introduces or revisits:

- low-rank matrix products and parameter-count arithmetic;
- pairwise/logistic ranking losses;
- expectation and weighted objectives;
- policy probability ratios conceptually;
- KL divergence at the depth needed to understand "do not move too far from the reference distribution";
- gradient direction as probability reweighting;
- no full RL theory or matrix-factorization proof sequence.

## History

Historical context should connect methods to practical model adaptation:

- task-specific fine-tuning of pretrained language models;
- instruction tuning and the transition from completion models toward assistants;
- human-preference optimization and RLHF;
- parameter-efficient methods such as LoRA;
- direct preference optimization and the move toward simpler offline preference objectives.

Avoid treating current post-training recipes as settled or universal.

## Interactives

Strong candidates:

- **Fine-tuning loss-mask explorer** — inspect which tokens contribute to instruction-tuning loss;
- **LoRA matrix explorer** — change rank and see parameter counts/possible update structure;
- **Preference-pair explorer** — compare demonstrations, rankings, and reward scores;
- **Policy-update intuition** — visualize a tiny categorical policy before/after positive/negative reward;
- **KL constraint explorer** — adjust policy probabilities and see divergence from a reference distribution.

## Worksheets

Useful worksheet topics:

- classify scenarios as prompting, continued pretraining, SFT, PEFT, or preference optimization;
- calculate full versus LoRA trainable parameter counts;
- trace LoRA matrix shapes;
- compute a tiny low-rank update;
- reason about pairwise preferences and reward scores;
- calculate simple expected returns;
- reason about how a policy update changes token probabilities;
- identify evaluation designs that cannot distinguish genuine improvement from overfitting to the training preference data.

## Embedded exercises

Candidates:

- implement a LoRA linear layer;
- freeze/unfreeze specified parameters correctly;
- verify adapter merging;
- construct loss masks for instruction examples;
- compute a pairwise preference loss;
- perform a tiny categorical policy update or preference-objective calculation.

## Repository/Jupyter project

**Adapt a small language model**

The learner should adapt a genuinely manageable model to a narrow task or response style using LoRA or another approved PEFT setup.

Required structure:

1. define the adaptation goal and evaluation criteria before training;
2. preserve a base-model baseline;
3. construct or curate a small train/validation/evaluation dataset;
4. calculate trainable versus frozen parameter counts;
5. train the adapter;
6. inspect training and validation behavior;
7. compare base and adapted model on held-out examples;
8. identify at least one regression or failure mode;
9. document what the experiment demonstrates and what it does not.

The project should not require expensive frontier-model training.

## Mastery expectations

After Module 20 the learner should be able to:

- describe the major stages between a base pretrained model and a post-trained assistant;
- fine-tune a small open model under a controlled experimental setup;
- explain and personally implement LoRA's core mechanism;
- reason about preference data and reward models;
- follow an RLHF diagram/paper without reinforcement-learning terminology being opaque;
- explain why PPO/KL-style constraints exist;
- contrast RLHF with direct-preference approaches such as DPO;
- evaluate adaptation as an empirical tradeoff rather than assuming post-training simply makes a model "better."

## Deliberately deferred

- full reinforcement-learning curriculum;
- advanced PPO derivations;
- online RL environments;
- process reward models in depth;
- constitutional/self-improvement methods as a major unit;
- large-scale synthetic-data pipelines;
- model merging as a specialization;
- advanced PEFT catalogs;
- distributed fine-tuning systems.

---

# 21 — Evaluation, Reasoning, Retrieval, and Tool Use

## Purpose

Teach how to determine whether a generative language-model system actually works, then introduce reasoning-oriented inference, retrieval augmentation, and tool use as explicit mechanisms that change what the system can do at inference time.

This module should connect model evaluation to system evaluation. A model can have strong benchmark scores and still fail a particular application; a retrieval or tool-using system can fail even when the underlying model is capable.

## Prerequisites

Required capabilities:

- held-out evaluation, metrics, leakage, baselines, and reproducibility;
- autoregressive generation and sampling;
- embeddings and similarity;
- modern LLMs and post-training;
- probability distributions and calibration intuition;
- ordinary software/API reasoning sufficient to understand structured tool calls.

## Proposed core objectives

Capabilities should cover:

- define an LLM evaluation around a concrete task and failure cost rather than a generic "intelligence" score;
- distinguish benchmark, human, model-judge, and task-specific evaluation methods;
- identify contamination, leakage, prompt sensitivity, sampling variance, and judge bias risks;
- evaluate stochastic outputs with repeated trials and appropriate aggregation;
- distinguish answer quality from generated reasoning text and avoid treating chain-of-thought as privileged access to internal mechanism;
- explain inference-time computation, sampling/search, self-consistency, and verification as ways of spending additional compute on a problem;
- build and evaluate a basic retrieval-augmented generation pipeline;
- reason about chunking, embeddings, nearest-neighbor retrieval, reranking, context construction, and citations/grounding;
- explain structured tool/function calling as model-generated actions embedded in an external control loop;
- identify tool-use failure modes and basic trust/permission boundaries;
- design end-to-end evaluations that localize failures to model, retrieval, tool, or orchestration components.

## Lesson sequence

### 01 — What Does "Good" Mean for a Generative Model?

**Job:** Reframe evaluation around tasks, distributions, and failure costs.

Topics:

- why loss/perplexity is useful but insufficient for downstream behavior;
- closed-form versus open-ended tasks;
- exact-match, unit-test, rubric, ranking, and semantic-quality evaluations;
- task distributions;
- coverage of normal and adversarial/edge cases;
- false-positive/false-negative asymmetry where relevant;
- pass@k and repeated sampling conceptually;
- latency/cost as possible system metrics;
- defining success before running the experiment.

The learner should create an evaluation specification before being shown model outputs.

### 02 — Benchmarks, Contamination, and Judges

**Job:** Build skepticism and practical discipline around modern LLM benchmark claims.

Topics:

- benchmark datasets and standardized comparison;
- development/test separation revisited;
- training-data contamination;
- benchmark saturation;
- prompt/template sensitivity;
- decoding-setting sensitivity;
- stochastic variance;
- human evaluation;
- pairwise preference evaluation;
- model-as-judge methods;
- judge position/style/length/self-preference biases;
- using multiple methods and inspecting disagreements;
- statistical uncertainty around small score differences.

A leaderboard number should always be tied back to a specific protocol.

### 03 — Reasoning and Inference-Time Computation

**Job:** Explain why spending more inference computation can improve some tasks without implying that visible reasoning text is a faithful microscope into model internals.

Topics:

- direct answering versus decomposition;
- generated intermediate work/scratchpads;
- chain-of-thought prompting as an elicitation method;
- self-consistency and multiple candidate samples;
- search over candidate solutions;
- verification/checking;
- external calculators/code/tools as deterministic support;
- inference-time compute as a budget;
- reasoning-model behavior conceptually;
- outcome-based versus process-based evaluation;
- unfaithfulness or post-hoc rationalization of generated reasoning traces;
- hidden/internal reasoning versus user-visible explanations as distinct concepts.

The learner should compare methods on problems with objective checkers when possible.

### 04 — Retrieval-Augmented Generation: Why Retrieve?

**Job:** Motivate retrieval as a way to supply external, current, private, or domain-specific context without expecting the model weights to contain everything.

Topics:

- knowledge in parameters versus knowledge supplied in context;
- retrieval-augmented generation (RAG);
- corpus/document ingestion;
- chunking;
- indexing;
- query representation;
- retrieval;
- context assembly;
- generation grounded in retrieved evidence;
- citations/source traceability;
- recall versus precision in retrieved context;
- stale or contradictory sources;
- context-window competition.

The lesson should clearly separate retrieval from fine-tuning: they solve different problems.

### 05 — Embedding Retrieval, Ranking, and Reranking

**Job:** Connect the embedding/similarity foundation from Module 15 to practical retrieval systems.

Topics:

- document/chunk embeddings;
- query embeddings;
- cosine/dot-product similarity revisited;
- nearest-neighbor search conceptually;
- top-k retrieval;
- chunk size/overlap tradeoffs;
- metadata filtering;
- lexical versus semantic retrieval at a high level;
- hybrid retrieval conceptually;
- reranking;
- retrieval metrics such as recall@k/precision@k where useful;
- evaluating retrieval separately from generation;
- why a vector database is an implementation choice rather than a conceptual requirement.

### 06 — Tool Use and the Agent Loop

**Job:** Explain tool calling as a structured interaction between a probabilistic model and deterministic/external systems.

Topics:

- tool/function schemas;
- choosing a tool;
- producing structured arguments;
- validating arguments;
- executing outside the model;
- returning tool results as observations;
- multi-step model → action → observation loops;
- stopping conditions;
- deterministic tools versus probabilistic language output;
- retries/error handling;
- permissions and least privilege;
- untrusted retrieved/web/tool content and prompt-injection risk;
- human approval for consequential actions;
- why framework abstractions do not remove these underlying responsibilities.

This lesson should be implementation-neutral. It should not teach a specific agent framework.

### 07 — Evaluate the Whole LLM System

**Job:** Synthesize evaluation, retrieval, reasoning, and tool use into component-level failure analysis.

Topics:

- end-to-end success metrics;
- component metrics;
- retrieval failure versus reasoning failure versus tool-selection failure versus tool-execution failure;
- trace/log inspection;
- controlled ablations: model-only, retrieval-enabled, tool-enabled;
- test fixtures for tools;
- deterministic checks where possible;
- adversarial/untrusted-context cases;
- regression suites;
- cost/latency/quality tradeoffs;
- deciding whether a more complex system actually beats a simpler baseline.

## Mathematics

Module 21 primarily reuses earlier mathematics:

- probability and repeated-sampling intuition;
- means/variance and uncertainty around evaluations;
- ranking metrics;
- cosine/dot-product similarity;
- precision/recall adapted to retrieval;
- expected-cost tradeoffs;
- no new detached statistics or information-retrieval mathematics course.

Where small score differences matter, the course should introduce practical confidence intervals or bootstrap-style uncertainty lightly rather than presenting a single benchmark number as exact.

## History

Historical context should explain why these system patterns emerged:

- standardized NLP benchmarks and their transition into broad LLM evaluations;
- retrieval-augmented generation as a continuation of information-retrieval plus neural-generation ideas;
- prompting, sampling, search, and verification as forms of inference-time computation;
- structured function/tool calling and the modern agent-loop pattern.

Avoid an agent-framework chronology.

## Interactives

Strong candidates:

- **Evaluation variance explorer** — sample stochastic outcomes and watch score uncertainty change with trial count;
- **Judge bias explorer** — alter ordering/style/verbosity in toy pairwise evaluations;
- **Retrieval pipeline explorer** — change chunking/top-k and inspect retrieved context;
- **Embedding search visualizer** — move a query in a small embedding space;
- **Tool loop visualizer** — inspect model action, validated call, result, and next model step;
- **Failure localizer** — classify observed failures by component.

## Worksheets

Useful worksheet topics:

- design an evaluation protocol for a concrete use case;
- detect contamination/leakage or invalid benchmark comparisons;
- calculate basic repeated-sampling metrics;
- compare retrieval precision/recall tradeoffs;
- rank toy embedding vectors by similarity;
- trace a tool-call loop and identify validation boundaries;
- classify system failures and propose the minimum experiment needed to isolate each cause.

## Embedded exercises

Candidates:

- implement cosine retrieval over a small matrix of embeddings;
- calculate retrieval metrics;
- select context under a token/length budget;
- validate structured tool arguments against a simple schema;
- implement a deterministic tool dispatcher for provided calls;
- score repeated model outputs against a deterministic checker;
- compare system ablations from supplied results.

## Jupyter/repository project

**Build and evaluate a grounded LLM system**

The learner should build a small system around an accessible model using either retrieval, one or more tools, or both.

Required structure:

1. define the task and evaluation set first;
2. establish a model-only baseline;
3. add retrieval and/or tool use;
4. evaluate the added component separately where possible;
5. compare end-to-end performance against the baseline;
6. record latency/cost/complexity where meaningful;
7. inspect failures manually;
8. include at least one adversarial or malformed-input case;
9. decide whether the added system complexity is justified by evidence.

The emphasis is evaluation and mechanism, not building a generalized autonomous-agent platform.

## Mastery expectations

After Module 21 the learner should be able to:

- design meaningful evaluations for generative model behavior;
- interpret benchmark claims skeptically and identify common invalid comparisons;
- reason about stochastic evaluation uncertainty;
- explain inference-time reasoning/search/verification without anthropomorphizing generated traces;
- build and evaluate a basic retrieval pipeline;
- distinguish retrieval failure from generation failure;
- explain tool calling as an external control loop with explicit trust boundaries;
- evaluate an LLM application as a system rather than attributing every outcome to "the model."

## Deliberately deferred

- prompt-engineering catalogs;
- vendor-specific agent frameworks;
- multi-agent orchestration as a core topic;
- production vector-database administration;
- advanced approximate-nearest-neighbor index algorithms;
- formal information-retrieval theory;
- deep mechanistic study of reasoning-model internals;
- automated judge research as a specialization;
- enterprise agent security architecture.

---

# 22 — GPUs, Quantization, and LLM Inference

## Purpose

Make LLM inference performance understandable from first principles. The learner should connect matrix-heavy neural computation to GPUs, understand memory hierarchy and numerical formats, explain why autoregressive decoding behaves differently from training, calculate the major contributors to model/KV-cache memory, understand quantization as a numerical approximation, and profile real inference tradeoffs.

This module is systems-oriented but remains an ML deep-primer module, not a GPU-programming or distributed-serving course.

## Prerequisites

Required capabilities:

- matrix multiplication and tensor shapes;
- PyTorch tensors, devices, dtypes, mixed precision, and batch processing;
- Transformer attention and decoder-only language models;
- modern attention variants from Module 19;
- autoregressive generation and sampling;
- basic floating-point awareness.

## Proposed core objectives

Capabilities should cover:

- explain why neural-network workloads map well to massively parallel accelerators;
- distinguish CPU/GPU execution strengths conceptually;
- explain kernels, matrix multiplication, tensor cores/specialized matrix hardware, threads/warps/blocks at a useful architectural level without requiring CUDA programming;
- distinguish compute throughput from memory bandwidth and reason about which is likely limiting a workload;
- estimate parameter-memory requirements from parameter count and numerical format;
- explain prefill versus decode phases of autoregressive inference;
- explain the purpose, growth, and memory cost of the KV cache;
- distinguish latency, time-to-first-token, inter-token latency, throughput, and batch throughput;
- understand floating-point formats and integer/low-bit representations relevant to ML;
- explain post-training quantization and the basic scale/zero-point/grouping ideas behind mapping high-precision values to lower-bit representations;
- distinguish weight-only, weight-and-activation, static/dynamic, and quantization-aware approaches at a useful level;
- measure memory/speed/quality tradeoffs of quantizing a small model;
- explain continuous batching, paged KV-cache management, prefix caching, and speculative decoding by the bottleneck they address;
- recognize distributed inference/parallelism as a later specialization rather than opaque magic.

## Lesson sequence

### 01 — Why GPUs Fit Neural Networks

**Job:** Connect the arithmetic learned throughout the course to accelerator hardware.

Topics:

- CPUs optimized for flexible low-latency serial/general computation;
- GPUs optimized for large numbers of similar parallel operations;
- matrix multiplication as the dominant primitive in dense neural layers;
- convolution/attention as compositions of highly parallel numerical operations;
- SIMD/SIMT intuition;
- throughput versus single-thread latency;
- GPU cores versus specialized matrix/tensor hardware conceptually;
- CPU-GPU data transfer as a real cost;
- why "GPU" persists as a name despite general-purpose computation.

The learner should leave able to explain why linear algebra workloads—not graphics specifically—fit the hardware organization.

### 02 — Kernels, Memory Hierarchy, and the Cost of Moving Data

**Job:** Establish the minimal hardware model required to reason about performance.

Topics:

- host memory versus device memory;
- VRAM;
- registers/local/shared/cache/global memory at a conceptual level;
- kernels;
- threads, warps/wavefront-like groups, and blocks as a useful execution hierarchy;
- coalesced/reused data conceptually;
- arithmetic throughput;
- memory bandwidth;
- arithmetic intensity;
- compute-bound versus memory-bandwidth-bound workloads;
- kernel-launch and synchronization overhead;
- operation fusion as a way to reduce data movement/launches.

A simplified roofline-style intuition is useful; a formal performance-modeling course is not.

### 03 — How Much Memory Does a Model Need?

**Job:** Make model-memory discussions calculable instead of anecdotal.

Topics:

- parameter count;
- bytes per parameter for FP32, FP16/BF16, INT8, 4-bit-style storage as first approximation;
- model weights versus activations;
- inference versus training memory;
- gradients and optimizer state as training-only differences;
- temporary workspace/buffers;
- allocator overhead;
- batch size and sequence length effects;
- why file size, loaded memory, and peak runtime memory can differ;
- simple memory-budget planning.

### Required calculations

The learner should repeatedly calculate rough memory budgets from parameter count and dtype, then compare those estimates with observed runtime measurements and explain the gap.

### 04 — Prefill, Decode, and the KV Cache

**Job:** Explain the two computational regimes inside autoregressive LLM inference.

Topics:

- prompt processing/prefill;
- parallel processing across prompt tokens;
- autoregressive decode dependency;
- one/few new token positions per decode step;
- recomputing prior keys/values versus caching them;
- KV-cache structure by layer/head/token;
- effect of MHA versus GQA/MQA on KV-cache size;
- sequence length and batch size;
- memory growth with context;
- time-to-first-token versus steady decode speed;
- why long prompts and long generations stress different resources.

The learner should calculate a toy KV-cache size from layer/head/dimension/sequence/dtype information.

### 05 — Latency, Throughput, Batching, and Bottlenecks

**Job:** Give the learner the vocabulary to profile and reason about inference rather than comparing tokens-per-second numbers blindly.

Topics:

- end-to-end latency;
- time to first token (TTFT);
- inter-token latency/time per output token;
- tokens per second;
- requests per second;
- batch throughput;
- prompt versus output token accounting;
- static batching;
- why batching can improve hardware utilization while increasing individual latency;
- compute-bound prefill versus often bandwidth-sensitive decode intuition;
- model size, batch size, context length, and hardware as interacting variables;
- warmup and measurement methodology.

### 06 — Numerical Precision and Quantization Fundamentals

**Job:** Connect numerical representation to memory, bandwidth, speed, and approximation error.

Topics:

- FP32, FP16, BF16 revisited;
- exponent/mantissa range intuition;
- integer and low-bit representations;
- dynamic range;
- scale and zero-point intuition;
- symmetric versus asymmetric quantization conceptually;
- per-tensor/per-channel/group-wise scaling;
- rounding/clipping;
- quantization error;
- outliers;
- dequantization during computation;
- weight-only versus activation quantization;
- post-training quantization versus quantization-aware training;
- why "4-bit model" can describe several materially different schemes.

### 07 — Quantize and Measure a Language Model

**Job:** Turn quantization from terminology into an empirical tradeoff.

The learner should run a controlled experiment using an appropriately small open model and one or more supported quantization methods.

Required measurements may include:

- model artifact/storage size;
- loaded RAM/VRAM where observable;
- time to first token;
- steady generation throughput;
- selected task/quality metrics or perplexity where practical;
- output comparisons on fixed prompts;
- CPU versus GPU behavior where hardware permits.

Topics:

- calibration data where relevant;
- group size/bit-width tradeoffs;
- measurement noise;
- quality degradation;
- hardware/backend-specific acceleration;
- why smaller memory footprint does not guarantee proportional speedup;
- documenting exact inference settings.

The goal is not to crown one quantization format as universally best.

### 08 — Efficient Serving: What Problem Does Each Technique Solve?

**Job:** Survey modern inference optimizations by bottleneck instead of by product/library.

Topics:

- continuous/dynamic batching;
- paged or block-based KV-cache management;
- prefix/prompt caching;
- speculative decoding;
- draft/verifier model intuition;
- kernel fusion/optimized attention implementations at a high level;
- tensor/model parallel inference as a scale-out concept;
- pipeline parallelism at a survey level;
- expert parallelism for MoE at a survey level;
- CPU offload and heterogeneous memory as tradeoffs;
- server throughput versus interactive latency;
- choosing an optimization only after identifying the bottleneck.

Distributed inference should be recognizable after this lesson, but implementing a multi-node serving stack is explicitly outside core scope.

## Mathematics

Module 22 emphasizes quantitative engineering:

- dimensional/tensor-shape arithmetic;
- bytes/bits and unit conversion;
- bandwidth, throughput, and latency relationships;
- parameter and KV-cache memory formulas;
- ratios/speedups;
- arithmetic-intensity intuition;
- linear quantization mappings;
- approximation/error measurement.

These calculations should be performed repeatedly by hand before being hidden behind profiler dashboards.

## History

Historical context should connect hardware changes to ML practice:

- GPUs from graphics accelerators to programmable general-purpose parallel processors;
- CUDA/general GPU computing and the deep-learning acceleration era;
- specialized tensor/matrix units;
- mixed-precision training/inference;
- low-bit quantization and local/edge inference;
- serving-system innovations driven by autoregressive decoder workloads.

Avoid a hardware-product catalog.

## Interactives

Strong candidates:

- **Matrix-parallelism visualizer** — distribute a toy GEMM across many workers;
- **Memory hierarchy animation** — illustrate reuse versus repeated global-memory traffic;
- **Model memory calculator** — parameter count/dtype/batch/context inputs;
- **KV-cache calculator/visualizer** — inspect growth by layer/token/head strategy;
- **Prefill/decode timeline** — contrast parallel prompt processing and sequential generation;
- **Quantization mapping explorer** — change bit width/scale and inspect rounding error;
- **Serving queue visualizer** — compare static versus continuous batching;
- **Speculative decoding visualizer** — draft several tokens and accept/reject against a target model.

## Worksheets

Useful worksheet topics:

- parameter-memory calculations;
- KV-cache calculations under MHA/GQA/MQA;
- classify workloads as likely compute- or bandwidth-sensitive from simplified data;
- compare TTFT and decode throughput measurements;
- convert high-precision values to a toy integer quantization grid;
- reason about clipping/rounding error;
- compare batching tradeoffs;
- identify which serving optimization targets a stated bottleneck.

## Embedded exercises

Candidates:

- write parameter/KV-memory calculator functions;
- implement a toy symmetric quantizer/dequantizer;
- compute group-wise scales;
- benchmark small matrix operations across dtypes where environment support allows;
- parse profiler measurements and compute derived throughput;
- calculate acceptance/speedup behavior for a toy speculative-decoding example.

## Jupyter/repository project

**Inference performance study**

The learner should choose a manageable open model and investigate an inference question such as:

- how precision/quantization affects memory and throughput;
- how context length affects TTFT and KV-cache memory;
- how batch size changes throughput/latency;
- how CPU and GPU inference differ for the same model;
- how an available attention/backend optimization changes performance.

Requirements:

1. state a hypothesis;
2. define hardware/software/model configuration precisely;
3. define metrics before measuring;
4. include warmup/repeated measurements where needed;
5. vary one primary factor systematically;
6. record raw results;
7. visualize at least one relationship;
8. explain the result using compute/memory/inference concepts from the module;
9. distinguish measured facts from speculative explanations.

## Mastery expectations

After Module 22 the learner should be able to:

- explain why GPUs are effective for neural-network inference;
- reason about kernels, parallelism, memory bandwidth, and data movement at a useful level;
- estimate whether a model fits in available memory;
- explain prefill/decode and calculate KV-cache growth;
- interpret latency/throughput benchmarks correctly;
- explain quantization as a numerical approximation and distinguish major categories;
- run a controlled quantization/performance experiment;
- recognize common serving optimizations and the bottleneck each addresses;
- approach inference performance as an empirical systems problem rather than a collection of magic flags.

## Deliberately deferred

- CUDA kernel programming;
- GPU assembly/microarchitecture;
- compiler internals;
- writing custom fused kernels;
- advanced roofline/performance modeling;
- multi-node distributed serving implementation;
- NCCL/collective communication internals;
- production autoscaling/load balancing;
- exhaustive quantization-format catalogs;
- hardware-specific tuning as mandatory core work.

---

# 23 — Research Engineering / Capstone

## Purpose

Synthesize the entire course into the habits of an ML research engineer: turn a vague curiosity into a falsifiable or answerable question, understand relevant prior work, establish a baseline, design a controlled experiment, track provenance, measure uncertainty, perform useful ablations, reproduce results, and communicate conclusions honestly.

The capstone does **not** require novel publishable research. A careful reproduction, extension, or comparative study is more valuable than a flashy project with weak evidence.

## Prerequisites

The learner should enter Module 23 having completed the mandatory core sequence through Module 22. The capstone may emphasize a subset of those capabilities, but the learner should be capable of selecting methods and evaluating tradeoffs across classical ML, neural networks, language models, post-training, LLM systems, and inference.

## Proposed core objectives

Capabilities should cover:

- turn an informal ML question into a precise research question or hypothesis;
- find and read relevant papers/documentation critically;
- distinguish claims, evidence, assumptions, and limitations in a paper;
- distinguish reproduction, replication, extension, and ablation;
- establish meaningful baselines;
- vary factors systematically and avoid changing multiple important variables without reason;
- account for stochasticity using repeated runs/seeds and practical uncertainty estimates where needed;
- preserve datasets, preprocessing, code, configurations, environments, checkpoints/results, and provenance well enough for another person to rerun the work;
- use experiment tracking appropriate to project scale without substituting tooling for scientific discipline;
- interpret negative/null results constructively;
- avoid p-hacking/metric shopping/cherry-picking;
- communicate results with tables/plots and a clear separation between observation and explanation;
- identify limitations and plausible alternative explanations;
- produce a reproducible research-engineering capstone artifact.

## Lesson sequence

### 01 — From Curiosity to Research Question

**Job:** Turn "I wonder whether..." into an experiment that can actually answer something.

Topics:

- exploratory versus confirmatory questions;
- hypotheses;
- independent/dependent variables;
- operational definitions;
- primary metric;
- target data/task distribution;
- controlled variables;
- confounders;
- falsifiability/answerability;
- selecting a question small enough to finish;
- expected information gain from an experiment;
- precommitting to what result would count as support, contradiction, or inconclusive evidence.

The learner should draft several candidate questions and reject those that cannot be evaluated cleanly.

### 02 — Reading ML Papers Like an Engineer

**Job:** Build a repeatable method for extracting the parts of a paper needed to understand and reproduce it.

Topics:

- title/abstract/figures first-pass reading;
- problem statement;
- claimed contribution;
- related work;
- method/architecture/objective;
- dataset;
- training/evaluation protocol;
- baselines;
- ablations;
- result tables;
- statistical uncertainty;
- limitations;
- appendix/supplementary details;
- distinguishing what is specified from what must be inferred;
- following citations backward to prerequisites and forward to later critiques/replications;
- code/model/data artifacts as part of the evidence.

The learner should produce a structured one-page paper teardown rather than a generic summary.

### 03 — Reproduction, Replication, Baselines, and Ablations

**Job:** Establish the experimental structures that let research claims be tested and interpreted.

Topics:

- reproduction using the same/similar method and setup;
- replication under independently constructed conditions;
- extension of a prior result;
- baseline selection;
- naive/simple baselines;
- strong prior-art baselines;
- ablations;
- controlled substitution of one component;
- negative controls where useful;
- implementation bugs masquerading as research findings;
- matching evaluation protocol before comparing numbers;
- what to do when a published result cannot be reproduced.

### 04 — Stochasticity, Uncertainty, and Honest Measurement

**Job:** Revisit statistics at the level needed to make ML experiment results defensible.

Topics:

- random initialization, data order, sampling, and nondeterminism;
- seeds as reproducibility aids, not evidence that variance disappeared;
- repeated runs;
- mean/median and dispersion;
- confidence intervals conceptually;
- bootstrap-style uncertainty where appropriate;
- paired comparisons where the same examples can be evaluated under two conditions;
- effect size versus mere score difference;
- multiple comparisons/metric shopping conceptually;
- reporting all planned primary results;
- recognizing when compute constraints make uncertainty estimates weak.

No detached mathematical-statistics unit is needed, but the learner should no longer report a 0.3-point improvement from one run as if it were exact.

### 05 — Reproducible Research Engineering

**Job:** Treat the experiment as an artifact another engineer could inspect and rerun.

Topics:

- source control;
- immutable/raw versus processed data where practical;
- data provenance;
- environment/dependency capture;
- configuration files;
- random seeds;
- checkpoint provenance;
- experiment IDs;
- raw result retention;
- generated tables/plots from source results;
- logs and failure notes;
- hardware/runtime metadata where performance matters;
- secrets/private data boundaries;
- README/run instructions;
- notebooks versus scripts/modules;
- when experiment-tracking tools help and when a simple structured directory/CSV/JSON is sufficient.

The course should favor boring, inspectable reproducibility over elaborate MLOps infrastructure.

### 06 — Capstone Proposal and Experiment Plan

**Job:** Convert the preceding habits into an approved, bounded final project plan before implementation begins.

The proposal should include:

- research question;
- motivation;
- relevant prior work/sources;
- hypothesis or expected outcomes;
- dataset/task;
- baseline;
- experimental variable(s);
- primary metric(s);
- secondary diagnostics;
- compute/hardware budget;
- planned repeated runs/seeds;
- planned ablation or controlled comparison;
- failure criteria/pivot rule;
- artifact/reproducibility plan;
- expected limitations.

A capstone proposal should be rejected or narrowed if its answer depends primarily on unbounded engineering work, proprietary inaccessible data, or compute far beyond the learner's resources.

### 07 — Execute, Analyze, and Communicate the Capstone

**Job:** Complete the core course with a research artifact whose claims are supported by evidence.

The final deliverable should contain:

- reproducible repository;
- exact experiment configuration;
- raw or appropriately summarized results;
- analysis notebook/script where useful;
- tables/plots generated from results;
- written report;
- discussion of failed runs/negative findings where relevant;
- limitations;
- follow-up questions;
- short retrospective on what changed between proposal and execution.

The report structure should normally be:

```text
Question
Context / related work
Hypothesis
Method
Data
Baseline
Experimental design
Results
Ablations / diagnostics
Interpretation
Limitations
Reproduction instructions
Further work
```

The learner should be able to defend every major claim by pointing to a specific experiment, measurement, or cited source.

## Capstone topic families

The capstone should permit choice while preserving rigorous structure. Good examples include:

### Small-model architecture study

Modify one component of TinyLM or another manageable Transformer and test a specific hypothesis, such as normalization placement, feed-forward gating, attention-head configuration, or context behavior.

### Post-training study

Compare a base model and one or more controlled adaptation strategies, LoRA ranks, data mixtures, or preference-training variants on a narrow evaluation.

### Retrieval/tool-system study

Test how chunking, retrieval strategy, reranking, or tool availability changes a well-defined end-to-end task while separately measuring component performance.

### Inference/quantization study

Measure a specific quality-memory-latency tradeoff across precision, quantization, context length, batching, or hardware.

### Reproduction/extension of a small published experiment

Reproduce a result whose scale is feasible, then change one meaningful factor and determine whether the claim/generalization still holds.

### Classical/deep-learning comparison

For a suitable problem, compare a simpler classical approach with a neural approach under controlled evaluation and analyze the conditions under which complexity is or is not justified.

The project should not require the learner to invent a new architecture or train a large foundation model.

## Mathematics

Module 23 consolidates rather than introduces a new mathematical field. It should use:

- descriptive statistics;
- variance/standard deviation;
- repeated-run uncertainty;
- practical confidence intervals/bootstrap intuition;
- effect sizes/relative differences;
- plotting and trend analysis;
- whichever earlier ML mathematics the selected capstone requires.

If a chosen paper requires mathematics outside the core syllabus, that material can be introduced just in time for the project rather than retroactively expanding the mandatory math curriculum.

## History and research culture

Historical/cultural context should support good research behavior:

- reproducibility and replication as scientific norms;
- benchmark culture and its failure modes;
- open-source code/data/model artifacts as accelerators of verification;
- negative results and failed replications as useful evidence;
- the distinction between engineering demonstration and scientific claim.

The course should avoid romanticizing research as lone-genius discovery. Modern ML research is heavily empirical, collaborative, infrastructural, and iterative.

## Worksheets/templates

Rather than many calculation worksheets, Module 23 should use structured research templates:

- research-question worksheet;
- paper teardown worksheet;
- experiment-design checklist;
- baseline/ablation matrix;
- reproducibility checklist;
- result-interpretation worksheet separating observations from explanations;
- capstone proposal template;
- final-report checklist.

## Embedded exercises

Only small exercises are needed:

- identify confounders in flawed experiment descriptions;
- choose valid baselines;
- interpret repeated-run results;
- calculate simple confidence intervals/bootstrap summaries from supplied samples;
- identify cherry-picking or invalid comparisons;
- reconstruct a missing experiment configuration from an incomplete research log and explain why reproduction fails.

The capstone itself belongs in a normal repository/IDE environment.

## Mastery expectations

After Module 23 the learner should be able to:

- formulate a tractable ML research question;
- read technical papers for methods and evidence rather than prestige or conclusions alone;
- reproduce and critically test a small research claim;
- design baselines and ablations;
- account for stochasticity and practical uncertainty;
- preserve experiment provenance and reproducibility;
- analyze negative or ambiguous results without forcing a story;
- communicate results with calibrated claims and explicit limitations;
- independently continue into specialized ML literature and projects after the core course.

## Deliberately deferred

- expectations of publishable novelty;
- a graduate-level research-methods/statistics sequence;
- large-scale MLOps platforms;
- distributed training infrastructure as a mandatory capstone requirement;
- academic publication mechanics as a major topic;
- formal causal inference unless a chosen specialization requires it;
- specialist mathematics unrelated to the selected question.

---

# Tranche-wide progression

Modules 19–23 should feel like a transition from **understanding a model** to **understanding the complete lifecycle of a modern model system**:

```text
Tiny language model
        |
        v
scale + data + modern architecture
        |
        v
adapt behavior through post-training
        |
        +-------------------+
        |                   |
        v                   v
evaluate/ground/use tools   run efficiently on hardware
        |                   |
        +---------+---------+
                  |
                  v
       design and execute research
```

## Mathematical progression across the final tranche

| Module | Mathematical emphasis |
| --- | --- |
| 19 | power laws, logarithms, parameter/tensor arithmetic, normalization/position intuition |
| 20 | low-rank matrix updates, ranking losses, expectation, KL/policy-probability intuition |
| 21 | repeated-sampling uncertainty, ranking/retrieval metrics, vector similarity |
| 22 | memory/throughput arithmetic, bit widths, KV-cache formulas, linear quantization |
| 23 | repeated-run statistics, confidence/uncertainty, effect-size reasoning, experiment-specific math |

The final tranche should not introduce mathematics simply because modern LLM papers contain it. Mathematics enters when it is required to understand a core mechanism or conduct a defensible experiment.

## History progression

The historical thread should close the course's broader story:

1. Transformers become scalable pretrained language models.
2. Empirical scaling and data/compute engineering change model development.
3. Post-training turns base models into instruction-following systems.
4. Evaluation evolves from fixed NLP benchmarks toward broader generative/system evaluation.
5. Retrieval, tools, search, and inference-time computation extend what a frozen model can accomplish.
6. GPU hardware, numerical formats, and serving systems shape what models can be deployed economically.
7. Research engineering ties the entire field back to repeatable empirical evidence.

History should remain integrated into technical explanations rather than becoming a separate chronology module.

## Final core mastery checkpoints

The final tranche adds three important synthesis checkpoints:

### Post-training checkpoint — Module 20

The learner adapts a small language model using a controlled PEFT/LoRA experiment and can explain the mathematical update being trained.

### Systems checkpoint — Modules 21–22

The learner runs at least one controlled LLM-system or inference experiment in which complexity/performance is measured rather than assumed.

### Research-engineering capstone — Module 23

The learner formulates a question, reads relevant prior work, establishes a baseline, executes a controlled experiment, accounts for uncertainty, preserves reproducibility, and communicates appropriately bounded conclusions.

These checkpoints should emphasize **transfer**. The learner should make experimental choices rather than merely follow a recipe whose expected output is known in advance.

## Core-completion standard

Completion of Module 23 should mean that the learner can move independently among three levels of reasoning:

1. **Mathematical/mechanistic:** explain what the important operations and learning objectives do;
2. **Implementation/systems:** build, adapt, run, profile, and debug manageable models and LLM systems;
3. **Empirical/research:** determine whether a claimed improvement is actually supported by controlled evidence.

The learner is not expected to be a specialist in every ML subfield. The goal of the mandatory core is to make later specialization legible: a computer-vision, reinforcement-learning, multimodal, diffusion, distributed-training, advanced inference, Bayesian, or advanced-mathematics path should now deepen an existing mental model rather than begin from mysterious terminology.

## Optional specialization boundary

After the core, specialist paths may include:

- computer vision;
- diffusion and generative image/video models;
- reinforcement learning;
- advanced LLM post-training and preference/reward modeling;
- advanced inference/performance engineering;
- distributed training and large-scale ML systems;
- multimodal models;
- speech/audio ML;
- Bayesian/probabilistic ML;
- advanced mathematics for ML;
- other domains added only when a concrete learning goal justifies them.

Optional paths should reuse the same objective/evidence/research discipline as the core rather than becoming disconnected topic collections.
