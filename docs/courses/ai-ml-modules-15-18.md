# AI/ML detailed syllabus plan: Modules 15–18

This document expands Modules 15 through 18 from the canonical [`ai-ml-syllabus.md`](ai-ml-syllabus.md) into a lesson-level teaching plan.

It is a **planning document**, not runtime-authored curriculum. Exact lesson files, objective metadata, worksheets, interactives, notebook repositories, sources, project specifications, and exercise definitions are authored later under `curriculum/courses/ai-ml/`.

This tranche is the core course's transition from general deep learning into language models and Transformers. The sequencing is intentionally causal rather than architectural: first establish how language becomes numerical data and what a language model is, then encounter the recurrent sequence-modeling bottleneck, then learn attention as a solution, then assemble a Transformer, and finally train and sample from a complete tiny autoregressive language model.

## Scope guardrails

This tranche contains some of the most important material in the course and should favor mechanical understanding over breadth.

The design rules are:

- language modeling must be understood before Transformers are introduced;
- embeddings should be understood as learned numerical representations, not as a mysterious NLP-specific data type;
- tokenization should be understood well enough to reason about model inputs and context length, without becoming a tokenizer-engineering course;
- attention should first appear as a solution to the fixed-context/alignment problem created by recurrent encoder-decoder models;
- the learner should hand-calculate a complete tiny attention example before treating attention as matrix code;
- query/key/value semantics should be connected to ordinary similarity, weighted averages, and learned projections already available from earlier modules;
- the Transformer should be assembled from known components rather than presented as one indivisible architecture diagram;
- the learner should personally implement attention, a Transformer block, the autoregressive generation loop, and sampling logic before relying on equivalent high-level library abstractions;
- the TinyLM project should be genuinely trainable and inspectable on modest hardware, not a miniature attempt at frontier-scale pretraining;
- modern architectural variants, scaling laws, large-scale data engineering, alignment/post-training, and inference systems belong to Modules 19–22.

The center of gravity is **mechanistic comprehension**. The learner should leave Module 18 able to point at each major part of a small decoder-only Transformer language model and explain what computation it performs, why it exists, and how it participates in training and generation.

## Tranche overview

| Module | Planned lessons / checkpoints | Primary job |
| --- | ---: | --- |
| 15 — Embeddings and Language Modeling | 6 | Turn discrete language into learnable numerical prediction problems |
| 16 — Attention | 6 | Derive attention from sequence-model limitations and implement it directly |
| 17 — Transformer From Scratch | 6 | Assemble attention, position, residual paths, normalization, and feed-forward computation into a Transformer block |
| 18 — Tiny Autoregressive Language Model | 7 | Build, train, evaluate, and sample from a complete small decoder-only language model |

This is approximately 25 lessons or project-oriented checkpoints. Lesson count is a planning estimate, not a product requirement. Modules 17–18 are implementation-heavy and may ultimately use fewer long-form explanatory lessons plus larger guided build activities.

## Dependency shape

Recommended order is linear through this tranche because the conceptual dependencies are unusually strong:

```text
14 CNNs & Sequences
        |
        v
15 Embeddings & Language Modeling
        |
        v
16 Attention
        |
        v
17 Transformer From Scratch
        |
        v
18 Tiny Autoregressive Language Model
```

Important prerequisite details:

- Module 15 also depends directly on probability, softmax/categorical cross-entropy, matrix/vector operations, PyTorch, and neural-network training from earlier modules.
- Module 16 depends on the seq2seq bottleneck introduced at the end of Module 14 and on dot products, matrix multiplication, softmax, and weighted sums from earlier mathematics.
- Module 17 depends strongly on a working understanding of attention, tensor shapes, neural-network layers, normalization, residual computation, and training mechanics.
- Module 18 depends on the learner having personally implemented a Transformer block; the project should not substitute a high-level Transformer package for that understanding.
- Exact runtime objective metadata should still encode only real capability dependencies rather than making every prior module objective a prerequisite for every later one.

## Learning-medium emphasis

| Module | Most important media |
| --- | --- |
| 15 | diagrams, embedded exercises, notebooks, small representation/sampling worksheets |
| 16 | substantial hand calculation, attention interactives, embedded implementation, visualization notebooks |
| 17 | shape-tracing worksheets, architecture diagrams, learner-written code, repository work |
| 18 | repository project, training notebooks/logs, evaluation/reflection, generation experiments |

---

# 15 — Embeddings and Language Modeling

## Purpose

Establish how discrete symbols become numerical model inputs and what it means to train a model to predict language before attention or Transformers appear.

The learner should understand that a language model is fundamentally a model of conditional probability over sequences. A Transformer is one architecture for parameterizing those probabilities, not the definition of a language model.

This module is also the first explicit information-theory treatment in the core. Rather than creating a detached information-theory unit, surprisal, entropy, cross-entropy, and KL divergence should be introduced here because language modeling makes probabilistic prediction and average log loss concrete.

## Prerequisites

Required capabilities from earlier modules:

- vectors, matrices, matrix multiplication, dot products, and basic similarity;
- probability distributions, conditional probability, expectation, and sampling;
- logits, softmax, categorical cross-entropy, and maximum-likelihood intuition;
- neural networks as parameterized composed functions;
- PyTorch tensors, modules, autograd, optimizers, and training loops;
- recurrent/sequence-model context from Module 14.

## Proposed core objectives

Objective IDs are provisional until authored. Capabilities should cover:

- represent categorical symbols using one-hot vectors and learned embeddings;
- explain embedding lookup as selecting a learned row/vector from a parameter matrix;
- reason about similarity and geometry in an embedding space without over-interpreting it;
- explain tokens, vocabularies, token IDs, special tokens, and basic subword tokenization tradeoffs;
- explain a language model as a conditional probability model over sequences;
- use the probability chain rule to factor sequence probability into next-token conditionals;
- explain and compute next-token logits, softmax probabilities, cross-entropy loss, and perplexity at an introductory level;
- explain surprisal, entropy, cross-entropy, and KL divergence and relate them to model likelihood/log loss;
- explain why minimizing cross-entropy to a fixed target distribution also minimizes KL divergence to that target;
- train and sample from a simple pre-Transformer neural language model;
- distinguish model architecture, tokenization, training objective, and decoding procedure as separate concerns.

## Lesson sequence

### 01 — From Symbols to Vectors

**Job:** Show why categorical symbols need a numerical representation and introduce one-hot vectors as the simplest bridge.

Topics:

- categorical values versus numerical magnitudes;
- token IDs as identifiers rather than meaningful scalar quantities;
- one-hot vectors;
- vocabulary size and vector dimensionality;
- one-hot inputs as basis vectors;
- multiplying a one-hot vector by a weight matrix;
- why one-hot representations are sparse and do not encode useful similarity by themselves;
- the transition from fixed identity encoding to learned representation.

The key conceptual result is that assigning token ID `73` does **not** mean that token is numerically larger or more similar to token `72` than token `4`.

### 02 — Learned Embeddings and Representation Geometry

**Job:** Make embeddings a natural consequence of learning useful representations rather than NLP magic.

Topics:

- an embedding matrix as trainable parameters;
- embedding lookup as selecting a matrix row;
- equivalence between one-hot matrix multiplication and direct lookup;
- dense learned vectors;
- how gradients update the representations used by the task;
- dot-product and cosine-similarity intuition revisited;
- neighborhoods and directions in representation space;
- distributional similarity at a conceptual level;
- static word embeddings as historical context;
- limits of interpreting embedding dimensions or analogies too literally.

A small visualization should show tokens moving or grouping as a simple model learns, while explicitly warning that geometric patterns are learned task-dependent structure rather than human-readable semantic coordinates.

### 03 — Tokens, Vocabularies, and Tokenization

**Job:** Explain how raw text becomes the discrete sequence consumed by a language model.

Topics:

- characters, bytes, words, and subwords as possible token units;
- vocabulary construction;
- token IDs;
- special tokens where relevant;
- unknown-token problems in word-level vocabularies;
- subword tokenization motivation;
- byte-pair encoding or a similar merge-based tokenizer at the conceptual level;
- how tokenization affects sequence length and context-window usage;
- why the same text can produce very different token counts under different tokenizers;
- decoding token IDs back to text;
- tokenizer/model vocabulary coupling.

The learner should inspect real tokenizations, but implementing a production tokenizer is not a core requirement.

### 04 — What Is a Language Model?

**Job:** Define language modeling independently of neural-network architecture and give information-theory quantities a concrete probabilistic meaning.

Topics:

- sequence probability;
- conditional next-token prediction;
- the probability chain rule:

  `P(x1, x2, ..., xn) = P(x1) P(x2|x1) ... P(xn|x1,...,x(n-1))`;

- context/history;
- n-gram models as finite-context historical examples;
- counts and smoothing intuition at a high level;
- autoregressive factorization;
- training examples created by shifting a sequence;
- teacher forcing terminology where useful;
- likelihood/cross-entropy connection revisited;
- surprisal/information content `-log p(x)`: an unlikely event carries more information under a model;
- entropy as expected surprisal under a reference/true distribution;
- cross-entropy as expected surprisal when predictions come from another distribution/model;
- KL divergence as excess cross-entropy incurred by using `q` when data are distributed as `p`;
- the relationship `H(p, q) = H(p) + D_KL(p || q)` at a conceptual/calculational level;
- why minimizing cross-entropy with respect to model parameters also minimizes KL to the target distribution when the target entropy is fixed;
- why KL divergence is asymmetric;
- perplexity as exponentiated average negative log-likelihood/cross-entropy and as a relative predictive measure, not a universal quality score.

Use tiny categorical distributions that can be calculated by hand. Base-2 versus natural logarithms may be mentioned, but coding theory, channel capacity, mutual-information derivations, and formal information theory remain outside the core.

This lesson should make clear that next-token prediction is not a Transformer-specific objective and that information-theoretic quantities are expectations over distributions, not mystical measures of model "understanding."

### 05 — A Neural Next-Token Predictor

**Job:** Build the simplest neural language model that uses learned representations and produces a probability distribution over the vocabulary.

Topics:

- context tokens and embeddings;
- combining or concatenating a fixed context window;
- hidden layer(s);
- vocabulary-sized output logits;
- softmax distribution over the next token;
- target token and cross-entropy loss;
- batches of context/target pairs;
- training with gradient descent;
- inspecting learned embeddings and predictions;
- contrast with recurrent sequence models from Module 14.

The exact model can be deliberately small and historically inspired. The point is to see the entire path:

```text
context token IDs
      -> embeddings
      -> neural computation
      -> vocabulary logits
      -> softmax probabilities
      -> next-token target
      -> cross-entropy loss
```

### 06 — Autoregressive Modeling Experiment

**Job:** Turn the previous pieces into a complete pre-Transformer language-model experiment and introduce generation as repeated conditional sampling.

Topics:

- training versus generation;
- using a model's predicted distribution to choose the next token;
- appending that token and predicting again;
- greedy choice versus stochastic sampling at a basic level;
- random seeds;
- generated sequence failure modes;
- train/validation loss;
- memorization versus generalization on a tiny corpus;
- qualitative inspection without treating a few impressive samples as evaluation.

The learner should train a genuinely small model, generate sequences, and explain why generation can diverge from the training distribution even when one-step prediction loss is reasonable.

## Mathematics

Module 15 introduces or revisits:

- one-hot/basis vectors;
- embedding matrices and lookup;
- dot product and cosine similarity;
- categorical probability distributions;
- conditional probability;
- the probability chain rule for sequences;
- logits and softmax;
- surprisal/negative log probability;
- entropy;
- cross-entropy/negative log-likelihood;
- KL divergence and `H(p,q) = H(p) + D_KL(p || q)`;
- exponentials/logarithms revisited through perplexity;
- stochastic sampling from a categorical distribution.

The information-theory treatment should be compact and motivated by language-model prediction. It should be deep enough that later KL terms in post-training are a revisit, not a first encounter, without expanding into a general information-theory course.

## History

The history spine should connect technical developments:

- n-gram/statistical language models;
- distributed representations and the idea that useful representations can be learned from usage context;
- early neural probabilistic language models;
- word embeddings such as Word2Vec/GloVe as important historical representations;
- the transition from hand-engineered linguistic features toward learned representations and end-to-end neural language modeling.

Do not imply that modern embedding spaces are simply larger versions of static word vectors; contextual representations arrive later through sequence models and Transformers.

## Interactives

Strong candidates:

- **One-hot to embedding explorer** — select a token and visualize matrix-row lookup;
- **Embedding-space explorer** — inspect neighbors/similarities in a tiny learned representation;
- **Tokenizer explorer** — compare character/word/subword segmentations and token counts;
- **Information-measure explorer** — vary tiny reference/predicted categorical distributions and inspect surprisal, entropy, cross-entropy, and KL;
- **Next-token distribution explorer** — modify context and inspect probability changes;
- **Autoregressive generation stepper** — show context, distribution, selected token, and updated context one step at a time.

## Worksheets

Useful worksheet topics:

- one-hot vectors and embedding lookup;
- dot/cosine similarity between tiny embedding vectors;
- sequence-probability factorization;
- construct shifted context/target training examples;
- calculate softmax/cross-entropy for a tiny vocabulary;
- calculate surprisal, entropy, cross-entropy, and KL for tiny categorical distributions;
- use the `H(p,q) = H(p) + D_KL(p || q)` relationship on a small example;
- trace several autoregressive generation steps from supplied distributions.

Worksheets should be lighter than Module 16's attention work but still make the sequence-probability mechanics explicit.

## Embedded exercises

Candidates:

- convert token IDs to one-hot vectors;
- implement embedding lookup using array indexing;
- show equivalence between one-hot matrix multiplication and row lookup;
- compute cosine similarity;
- construct context/target windows from token sequences;
- compute next-token cross-entropy from logits;
- compute entropy/cross-entropy/KL for supplied tiny distributions;
- implement categorical sampling from a supplied probability vector;
- implement a simple autoregressive loop around a provided predictor.

## Jupyter lab

**Small neural language model before Transformers**

The learner should:

1. tokenize a tiny text corpus using a deliberately simple tokenizer;
2. build context/target examples;
3. train a small embedding-based neural language model;
4. plot train/validation loss;
5. inspect learned embeddings or nearest neighbors;
6. generate several samples;
7. record what the model appears to capture and where it fails.

The experiment should establish a baseline mental model that later attention/Transformer work can improve upon.

## Mastery expectations

After Module 15 the learner should be able to:

- explain how discrete tokens become trainable vectors;
- reason about an embedding matrix in ordinary linear-algebra terms;
- explain what tokenization does and why tokenizer choices affect context length;
- define a language model without mentioning Transformers;
- factor sequence probability autoregressively;
- explain the complete next-token training objective;
- explain the practical relationship among negative log-likelihood, surprisal, entropy, cross-entropy, perplexity, and KL divergence;
- calculate those information quantities for tiny categorical distributions and explain what they measure;
- train and sample from a simple neural language model;
- distinguish tokenization, architecture, objective, and decoding.

## Deliberately deferred

- production tokenizer implementation and tokenizer training details;
- exhaustive comparison of tokenizer algorithms;
- contextual embedding architectures as a standalone survey;
- masked language modeling/BERT as a major topic;
- advanced information theory such as coding theory, channel capacity, and mutual-information derivations;
- attention and Transformers;
- large-scale pretraining data engineering;
- modern decoding algorithms beyond basic sampling.

---

# 16 — Attention

## Purpose

Teach attention as a solution to a concrete sequence-modeling problem before presenting the Transformer architecture.

The learner should first understand why compressing an entire source sequence into one recurrent hidden representation is limiting. Attention then emerges as a mechanism that lets a prediction dynamically retrieve and combine the parts of a sequence that matter for the current computation.

## Prerequisites

Required capabilities:

- recurrent encoder-decoder and seq2seq concepts from Module 14;
- vectors, matrices, dot products, matrix multiplication, and weighted sums;
- softmax and categorical distributions;
- embeddings and sequence representations from Module 15;
- PyTorch tensor/shape reasoning.

## Proposed core objectives

Capabilities should cover:

- explain the fixed-context/alignment problem that motivated neural attention;
- explain attention as scoring candidates, normalizing scores, and computing a weighted combination;
- explain query, key, and value roles without anthropomorphic overstatement;
- hand-calculate a complete scaled dot-product attention example;
- translate single-query attention into matrix-form attention;
- predict tensor shapes through `QK^T`, softmax, and multiplication by `V`;
- explain the purpose of scaling by `sqrt(d_k)`;
- apply padding and causal masks conceptually and computationally;
- distinguish self-attention from cross-attention;
- explain multi-head attention as multiple learned projection/attention subspaces;
- personally implement scaled dot-product and multi-head attention without relying on a high-level attention module.

## Lesson sequence

### 01 — The Sequence Bottleneck

**Job:** Reopen the problem left at the end of Module 14 and create the motivation for attention.

Topics:

- recurrent encoder-decoder recap;
- source sequence encoded into a fixed context representation;
- long-distance information loss and capacity pressure;
- alignment in translation/sequence prediction;
- why different output steps may need different source information;
- the idea of retaining all encoder states rather than collapsing everything into one vector;
- attention as dynamic access to those states.

The lesson should end with a concrete question: if the decoder could inspect every source representation, how should it decide which ones matter right now?

### 02 — Scores, Weights, and Context Vectors

**Job:** Introduce attention without query/key/value terminology first.

Topics:

- a current decoder state or request representation;
- candidate source representations;
- similarity/compatibility scores;
- softmax normalization into nonnegative weights that sum to one;
- weighted sums;
- context vector;
- different output positions receiving different attention distributions;
- visualization as an alignment matrix.

This lesson should use tiny numbers and ordinary weighted averages before introducing projection matrices.

### 03 — Queries, Keys, and Values

**Job:** Generalize the previous mechanism into the query/key/value abstraction.

Topics:

- query as the representation used to ask what is relevant;
- keys as representations used for matching/scoring;
- values as representations that are actually combined;
- why keys and values can come from the same source while serving different computational roles;
- learned linear projections `W_Q`, `W_K`, and `W_V`;
- dot-product similarity;
- shape reasoning for one query against multiple keys;
- avoiding anthropomorphic claims such as keys "knowing" what they contain.

Relate Q/K/V to familiar software/data-retrieval metaphors only as intuition; keep the actual computation explicit.

### 04 — Scaled Dot-Product Attention by Hand

**Job:** Make the full modern attention calculation mechanically transparent.

Core computation:

```text
Attention(Q, K, V) = softmax(QK^T / sqrt(d_k)) V
```

Topics:

- constructing small `Q`, `K`, and `V` matrices;
- computing `QK^T`;
- interpreting each score row;
- why large dot products can make softmax overly sharp and gradients poorly behaved;
- scaling by `sqrt(d_k)`;
- applying softmax row-wise;
- multiplying attention weights by `V`;
- tracing input and output shapes.

A substantial worksheet should require at least one complete end-to-end attention calculation with numbers small enough to calculate manually.

### 05 — Masks, Self-Attention, and Cross-Attention

**Job:** Show how the same calculation supports different information-flow constraints.

Topics:

- self-attention: queries, keys, and values derived from the same sequence;
- cross-attention: queries from one sequence/representation and keys/values from another;
- padding masks;
- causal masks;
- setting disallowed score entries to effectively negative infinity before softmax;
- why future tokens must be hidden during autoregressive training;
- attention matrices and directionality;
- shape comparisons between self- and cross-attention.

The learner should be able to draw a small causal mask and predict which token positions may influence each output position.

### 06 — Multi-Head Attention and Implementation

**Job:** Extend single attention into the form used by Transformers and personally implement it.

Topics:

- multiple learned Q/K/V projection sets;
- splitting model dimension across heads;
- independent attention calculations;
- concatenating head outputs;
- final output projection;
- head/model dimensions and divisibility;
- why multiple heads increase representational flexibility without implying one fixed human-interpretable role per head;
- implementation using ordinary tensor operations;
- comparing manual output against a trusted reference where useful;
- visualizing attention matrices from a trained toy model.

The learner should personally implement the attention computation rather than delegating it to `nn.MultiheadAttention` or an equivalent high-level primitive for the foundational exercise.

## Mathematics

Module 16 reinforces rather than introducing a large new mathematical branch:

- dot products as similarity scores;
- matrix multiplication;
- transposition;
- weighted averages;
- softmax normalization;
- square-root scaling;
- masks represented through score manipulation;
- tensor shapes and batched matrix operations.

The reason for `1/sqrt(d_k)` should be motivated through variance/magnitude intuition. A formal probabilistic derivation is unnecessary.

## History

Historical context should connect the mechanism to the problem it solved:

- encoder-decoder seq2seq models;
- alignment challenges in neural machine translation;
- Bahdanau-style learned attention and related early mechanisms;
- dot-product/multiplicative attention;
- the conceptual transition from recurrence plus attention toward self-attention as the main sequence operation.

The Transformer itself belongs to Module 17.

## Interactives

This module strongly justifies interactive work:

- **Alignment explorer** — change decoder/query state and watch source attention weights change;
- **Q/K similarity explorer** — manipulate vectors and inspect scores;
- **Attention matrix builder** — step through `QK^T`, scaling, mask, softmax, and `×V`;
- **Causal-mask explorer** — reveal which positions may attend to which;
- **Multi-head shape explorer** — show projection, head splitting, attention, concat, and output projection.

## Worksheets

This should be one of the worksheet-heavy modules in the course.

Core worksheet topics:

- weighted context calculations;
- query/key dot products;
- full scaled dot-product attention calculation;
- row-wise softmax and weighted `V` combinations;
- tensor-shape tracing;
- causal-mask construction;
- self-attention versus cross-attention identification;
- multi-head dimension calculations.

The learner should show intermediate matrices rather than only final answers.

## Embedded exercises

Candidates:

- compute raw attention scores;
- implement stable row-wise softmax for a supplied score matrix;
- apply a causal mask;
- implement scaled dot-product attention;
- validate probability rows sum approximately to one;
- implement head splitting/combining;
- implement multi-head attention using learned projection matrices supplied by the exercise;
- compare implementation against known expected tensors within numerical tolerance.

## Jupyter lab

**Visualizing attention behavior**

Use a small trained sequence model or toy attention module to inspect attention matrices across inputs. The learner should vary tokens/positions, inspect masks, compare heads, and record what patterns are robust versus what interpretations are speculative.

The lab should explicitly discourage assuming that an attention visualization is a complete explanation of model reasoning.

## Mastery expectations

After Module 16 the learner should be able to:

- explain why attention arose historically and technically;
- derive the attention computation from scoring plus weighted combination;
- calculate scaled dot-product attention by hand;
- reason about Q/K/V and tensor shapes;
- apply and explain masks;
- distinguish self- from cross-attention;
- explain and implement multi-head attention;
- inspect attention visualizations without over-interpreting them.

## Deliberately deferred

- complete Transformer architecture;
- positional representations;
- flash/fused attention kernels;
- sparse/linear attention variants;
- long-context architecture surveys;
- interpretability claims based solely on attention maps;
- production attention optimization.

---

# 17 — Transformer From Scratch

## Purpose

Assemble the attention mechanism and ordinary neural-network components into a complete Transformer block that the learner can explain operation by operation and implement without a high-level Transformer library.

The learner should experience the architecture as a composition of already-understood parts: representations, self-attention, residual paths, normalization, feed-forward transformations, position information, and repeated blocks.

## Prerequisites

Required capabilities:

- embeddings and language-modeling objective from Module 15;
- scaled dot-product and multi-head attention from Module 16;
- MLPs, activations, residual intuition, optimization, and PyTorch;
- mean/variance and normalization intuition;
- strong tensor-shape reasoning.

## Proposed core objectives

Capabilities should cover:

- explain why self-attention alone does not encode token order;
- construct and reason about positional representations;
- explain residual connections as additive information/gradient paths;
- explain LayerNorm at the level needed to implement/use it correctly;
- explain the position-wise feed-forward network inside a Transformer block;
- trace tensor shapes through a complete Transformer block;
- personally implement a Transformer block from lower-level PyTorch primitives;
- distinguish encoder-only, decoder-only, and encoder-decoder Transformer arrangements;
- explain causal decoder-only stacking as the architecture used for the following TinyLM project.

## Lesson sequence

### 01 — From Attention to a Transformer

**Job:** Introduce the architecture as a response to recurrence's sequential computation and connect it explicitly to the original Transformer design.

Topics:

- recurrence versus parallel sequence processing;
- self-attention as a way for each position to combine information from other positions;
- the original encoder-decoder Transformer at a high level;
- repeated blocks rather than one giant layer;
- residual stream intuition;
- encoder self-attention, decoder causal self-attention, and encoder-decoder cross-attention;
- why the core course will next focus on a decoder-only causal language model;
- architecture diagrams as data-flow diagrams rather than boxes to memorize.

Do not re-teach attention mechanics here; use them as known components.

### 02 — Position: Giving Order to Self-Attention

**Job:** Solve the fact that content-only self-attention does not inherently know sequence order.

Topics:

- permutation sensitivity/invariance intuition;
- token representations versus position representations;
- adding positional vectors to token embeddings;
- learned position embeddings;
- sinusoidal positional encoding from the original Transformer;
- a brief sine/cosine refresher sufficient to read the construction;
- multiple frequencies providing distinguishable position patterns;
- position representation as an architectural design choice rather than an absolute rule.

The learner should calculate or inspect a tiny positional-encoding example but should not get diverted into a trigonometry unit. Rotary position embeddings and other modern schemes belong to Module 19.

### 03 — Residual Connections and Normalization

**Job:** Explain the structural components surrounding attention/MLP sublayers.

Topics:

- residual/additive skip connections;
- preserving an information path while learning a transformation/update;
- gradient-flow intuition;
- LayerNorm inputs, mean, variance, normalization, learned scale/shift;
- normalization across feature dimensions for each token representation;
- normalization placement as an architecture detail that varies across Transformer families;
- conceptual difference between LayerNorm and batch-oriented normalization encountered earlier;
- numerical stability and training behavior at an intuitive level.

The goal is to understand why these operations exist, not to survey every normalization variant.

### 04 — The Position-Wise Feed-Forward Network

**Job:** Show that a Transformer block alternates token-to-token information mixing with per-token nonlinear transformation.

Topics:

- attention mixes information across sequence positions;
- feed-forward network processes each position independently using shared weights;
- expansion to a larger hidden dimension and projection back;
- activation function;
- shape preservation at block boundaries;
- parameter count intuition;
- relationship to the MLPs already studied;
- residual and normalization around the sublayer.

This lesson should reinforce a useful decomposition:

```text
attention = communication across positions
feed-forward network = nonlinear computation within each position
```

while noting that this is an intuition, not a complete theory of learned representation.

### 05 — Build a Transformer Block

**Job:** Personally assemble the known components into one working block.

Implementation should include:

- multi-head causal self-attention built from the learner's lower-level implementation;
- output projection;
- LayerNorm;
- feed-forward network;
- residual additions;
- explicit shape assertions/tests;
- dropout only if useful for the chosen tiny training setup;
- deterministic small test cases.

The learner should be able to trace one batch of shape `(batch, sequence, model_dimension)` through every operation.

High-level `TransformerEncoderLayer`/`TransformerDecoderLayer` style modules should not replace this foundational implementation.

### 06 — Stacking Blocks and Transformer Families

**Job:** Connect one block to a model architecture and prepare the exact structure used in Module 18.

Topics:

- repeated Transformer blocks;
- token/position embeddings at input;
- final normalization where appropriate;
- vocabulary projection/language-model head;
- causal decoder-only data flow;
- encoder-only models conceptually;
- encoder-decoder models conceptually;
- decoder-only models conceptually;
- weight sharing/tied embeddings as optional enrichment;
- context length and computational growth preview;
- where modern models differ, deferred to Module 19.

The lesson should end with a complete causal-language-model architecture diagram that contains no unexplained major computational box.

## Mathematics

Module 17 uses familiar mathematics in a deeper composition:

- addition of representation vectors;
- sine/cosine at a limited practical level for sinusoidal positions;
- mean and variance for LayerNorm;
- affine transformations and nonlinear activations;
- matrix multiplication and tensor reshaping;
- residual addition;
- repeated function composition.

A major objective is **shape fluency**. Worksheets and code should repeatedly require dimensions to be predicted before execution.

## History

Historical context should include:

- *Attention Is All You Need* and the move away from recurrence as the central sequence operation;
- encoder-decoder Transformer structure;
- the importance of parallelizable training across sequence positions;
- the later emergence of encoder-only, encoder-decoder, and decoder-only model families.

Do not yet turn this into a history of BERT, GPT versions, T5, LLaMA, and every architectural variant. Module 19 will connect the original Transformer to modern LLM design.

## Interactives

Strong candidates:

- **Transformer block data-flow explorer** — animate one tensor through attention, residual, norm, FFN, residual, norm;
- **Positional encoding explorer** — visualize several sinusoidal dimensions across positions;
- **Residual-path explorer** — compare transformed branch and additive residual stream;
- **Shape tracer** — enter batch/sequence/model/head dimensions and inspect all intermediate shapes;
- **Architecture family comparison** — visually distinguish encoder-only, decoder-only, and encoder-decoder information flow.

## Worksheets

Useful worksheet topics:

- positional encoding values for tiny positions/dimensions;
- LayerNorm calculation on a tiny vector;
- FFN shape calculations;
- residual-addition tracing;
- complete Transformer-block shape tracing;
- parameter-shape identification;
- causal decoder information-flow diagrams.

Attention arithmetic itself belongs primarily to Module 16 rather than being repeated extensively here.

## Embedded exercises

Candidates:

- construct positional encodings;
- implement a tiny LayerNorm calculation from primitives before using the framework module;
- implement a position-wise feed-forward sublayer;
- combine residual and sublayer outputs correctly;
- trace/validate tensor shapes;
- implement a complete Transformer block from supplied lower-level attention code;
- property tests ensuring causal outputs at earlier positions do not depend on changed future tokens.

## Repository work

**Transformer block from scratch**

This is a foundational learner-written implementation checkpoint, not a separate major course project. The code should be small enough to inspect in full and include tests for:

- output shape;
- causal masking behavior;
- deterministic behavior when dropout is disabled;
- gradient flow to expected parameters;
- compatibility with stacking multiple blocks.

The learner should write the core implementation before AI agents are used for refactoring or debugging assistance under the course's human-written-foundation rule.

## Mastery expectations

After Module 17 the learner should be able to:

- draw and explain a Transformer block from memory at the component level;
- explain how position information enters the model;
- explain residual and normalization roles;
- distinguish attention from the feed-forward sublayer;
- trace all important tensor dimensions through a block;
- implement a Transformer block without a high-level Transformer abstraction;
- explain the structural differences among encoder-only, encoder-decoder, and decoder-only Transformers;
- explain why the next module can produce a language model simply by stacking known blocks and adding a vocabulary prediction head.

## Deliberately deferred

- RoPE, ALiBi, and other modern position methods;
- RMSNorm and modern normalization variants;
- SwiGLU/gated FFNs;
- grouped-query/multi-query attention;
- mixture-of-experts layers;
- long-context architecture variants;
- fused kernels/FlashAttention;
- large-scale distributed Transformer training;
- detailed scaling-law analysis.

---

# 18 — Tiny Autoregressive Language Model

## Purpose

Turn the learner's Transformer block into a complete small GPT-style causal language model that can be trained, validated, checkpointed, sampled from, and analyzed end to end.

This is a major synthesis project. The learner should encounter the same conceptual pipeline used by much larger autoregressive models while keeping the data, model, and compute small enough that every part remains inspectable.

## Prerequisites

Required capabilities:

- tokenization, embeddings, sequence probability, cross-entropy, and autoregressive modeling from Module 15;
- attention and causal masking from Module 16;
- a personally implemented Transformer block from Module 17;
- PyTorch training loops, devices, optimizers, checkpoints, and reproducibility;
- evaluation/generalization discipline from Module 07.

## Proposed core objectives

Capabilities should cover:

- turn a token sequence into causal input/target training windows;
- batch language-model examples correctly;
- assemble token/position representations, stacked Transformer blocks, normalization, and vocabulary logits into a complete model;
- personally write the initial language-model training loop;
- monitor training and validation loss and diagnose basic training failures;
- explain exposure differences between training and autoregressive generation;
- personally implement the autoregressive generation loop;
- implement and explain greedy decoding, temperature scaling, top-k sampling, and top-p/nucleus sampling at the depth needed for controlled experiments;
- evaluate a tiny model quantitatively and qualitatively without overclaiming from sample quality;
- save/reload checkpoints and reproduce a training/generation experiment;
- explain how this small model is structurally related to modern decoder-only LLMs and where the scale/architecture differences begin.

## Project framing

The project should favor transparency over benchmark performance.

A suitable default is a small character-level, byte-level, or intentionally simple subword corpus whose tokenizer and dataset pipeline are easy to inspect. The model should be small enough to train on ordinary local hardware or an inexpensive modest GPU if desired.

The learner must not replace the personally implemented attention/block stack with a high-level pretrained-model library for the core project.

## Lesson / checkpoint sequence

### 01 — Turn Text Into Training Examples

**Job:** Build the causal dataset pipeline.

Topics:

- choose/construct a tiny corpus;
- train/validation split before window construction where appropriate;
- tokenizer/vocabulary selection;
- encode text to token IDs;
- context length/block size;
- sliding or sampled windows;
- input sequence `x[0:T]` and shifted target `x[1:T+1]`;
- batching sequences;
- batch, time, and vocabulary dimensions;
- deterministic sampling/seeds;
- inspecting decoded examples to verify correctness.

The learner should manually inspect examples before any model training begins.

### 02 — Assemble the TinyLM

**Job:** Build the complete decoder-only model from known pieces.

Components:

- token embedding table;
- positional representation;
- stack of learner-written Transformer blocks;
- final normalization where appropriate;
- vocabulary projection/language-model head;
- logits of shape `(batch, sequence, vocabulary)`;
- parameter count inspection;
- optional tied token/output embeddings if chosen explicitly;
- forward pass returning logits and optionally loss.

The implementation should remain small enough that the learner can read the entire model definition without navigating a framework architecture forest.

### 03 — Train the Model

**Job:** Write and understand the end-to-end optimization loop.

Topics:

- forward pass;
- flattening/alignment for token-level cross-entropy where needed;
- loss calculation across predicted positions;
- zero gradients;
- backward pass;
- optimizer step;
- minibatches;
- learning rate;
- gradient clipping if recurrently useful, explained rather than cargo-culted;
- periodic validation;
- checkpoint saving;
- logging experiment configuration;
- estimating tokens processed and basic compute awareness without a systems detour.

The first training loop should be learner-written explicitly before convenience abstractions are considered.

### 04 — Debugging and Understanding Training Behavior

**Job:** Treat training curves and failure cases as evidence rather than waiting for generated text to look impressive.

Topics:

- deliberate overfit-a-tiny-batch sanity test;
- train versus validation loss;
- learning-rate failures;
- initialization/normalization problems at the practical level;
- exploding/NaN loss diagnostics;
- context length and model capacity tradeoffs;
- memorization on tiny corpora;
- gradient/activation inspection where useful;
- reproducibility;
- checkpoint resume tests.

This checkpoint should reinforce research-engineering habits that will later be formalized in Module 23.

### 05 — Autoregressive Generation

**Job:** Personally implement the generation loop around the trained model.

Topics:

- seed/prompt tokens;
- crop to context window where required;
- forward pass;
- select logits from the final position;
- convert logits to a distribution;
- choose/sample next token;
- append token;
- repeat;
- decode tokens to text;
- inference mode/no-grad behavior;
- why generation is sequential even though training processes many positions in parallel.

The learner should be able to trace a generation iteration line by line.

### 06 — Sampling and Decoding

**Job:** Show that a trained probability model and a decoding policy are separate systems.

Topics:

- greedy argmax decoding;
- stochastic categorical sampling;
- temperature scaling of logits;
- low versus high temperature behavior;
- top-k filtering;
- top-p/nucleus filtering;
- random seeds and sample variability;
- repetition/pathological generation at a basic level;
- why decoding settings can change apparent model quality without changing model weights;
- controlled comparison of decoding strategies.

The learner should personally implement temperature, top-k, and top-p filtering rather than only call a generation API.

Beam search may be mentioned historically/for contrast but is not a central autoregressive-chat-model skill in this course.

### 07 — Evaluate, Explain, and Package the TinyLM

**Job:** Complete the project as an engineering/learning artifact rather than stopping after the first amusing sample.

Deliverables should include:

- reproducible dataset/tokenizer configuration;
- model architecture/configuration;
- parameter count;
- training and validation loss curves;
- checkpoint/reload instructions;
- several controlled generation comparisons;
- at least one deliberate failure analysis;
- explanation of every major architecture component;
- reflection comparing this model with the recurrent/simple neural language models from Modules 14–15;
- a short account of what remains different between the TinyLM and modern LLMs.

The final explanation should be judged as importantly as the generated samples.

## Mathematics

Module 18 synthesizes rather than opening a new mathematical branch:

- sequence indexing and shifted targets;
- cross-entropy over many token positions;
- average negative log-likelihood/perplexity;
- categorical sampling;
- temperature scaling of logits;
- probability renormalization after top-k/top-p filtering;
- parameter counts and tensor dimensions;
- optimization curves and basic descriptive experiment statistics.

## History

Historical context should connect the project to the decoder-only autoregressive line without turning into a model-version catalog:

- autoregressive neural language models;
- the Transformer enabling highly parallel next-token training;
- decoder-only Transformer language modeling;
- early GPT-style pretraining as an important architectural/training lineage;
- the observation that the learner's TinyLM shares the same broad computational skeleton while differing enormously in scale, data, architecture refinements, training regime, post-training, and systems engineering.

Those differences become the subject of Module 19 onward.

## Interactives

Fewer new interactives are necessary because this module is project-heavy. Useful candidates:

- **Causal window builder** — show input/target shifts across a sequence;
- **Generation loop stepper** — context → logits → filtering → probability → sample → append;
- **Sampling explorer** — manipulate temperature/top-k/top-p against one fixed logit vector;
- **Parameter/shape explorer** — change model dimensions and inspect rough parameter growth.

## Worksheets

Worksheets should be selective:

- construct causal input/target windows;
- trace tensor shapes through the full TinyLM;
- calculate one token-level cross-entropy example;
- calculate temperature-adjusted probabilities for a tiny logit vector;
- perform top-k/top-p filtering by hand on a tiny distribution;
- trace several autoregressive generation steps.

Most mastery evidence should come from the actual project rather than extensive paper drills.

## Embedded exercises

Useful constrained exercises before/in support of the repository project:

- create shifted language-model targets;
- reshape logits/targets for cross-entropy correctly;
- implement one autoregressive generation step;
- temperature-scale logits;
- implement top-k filtering;
- implement top-p filtering;
- deterministic sampling with a supplied RNG seed;
- verify probabilities remain normalized after filtering.

## Repository project

**TinyLM**

This is one of the course's major learner-written implementation checkpoints.

Minimum project capabilities:

- encode/decode text;
- build causal training batches;
- learner-written multi-head attention and Transformer block;
- stack blocks into a decoder-only language model;
- train from scratch on a small corpus;
- track train/validation loss;
- save/reload checkpoints;
- generate autoregressively;
- support greedy, temperature, top-k, and top-p decoding;
- reproduce at least one training run from configuration/seed;
- produce a concise technical report/reflection.

The project should avoid unnecessary application scaffolding. The educational value is in the model, training loop, experiments, and explanation.

## Mastery expectations

After Module 18 the learner should be able to:

- explain a decoder-only autoregressive language model end to end;
- create its training data from raw token sequences;
- implement its major architectural mechanics personally;
- train it and diagnose basic failures;
- distinguish parallel training from sequential autoregressive generation;
- implement common sampling strategies and explain their behavioral effects;
- interpret train/validation loss separately from anecdotal generation quality;
- reproduce a small experiment;
- explain precisely which ideas scale directly toward modern LLMs and which important ingredients have not yet been covered.

## Deliberately deferred

- frontier-scale datasets and pretraining infrastructure;
- distributed/data/model/pipeline parallelism;
- modern architecture refinements such as RoPE, RMSNorm, GQA, SwiGLU, and MoE;
- scaling laws in depth;
- supervised instruction tuning;
- preference optimization/RLHF/DPO;
- retrieval/tool use;
- production inference serving and KV-cache optimization;
- quantization and local-model systems work;
- benchmark/evaluation methodology at modern LLM scale.

---

# Tranche-level progression

## Conceptual story

The four modules should feel like one continuous derivation:

```text
discrete language
    -> token IDs
    -> learned embeddings
    -> next-token probability model
    -> recurrent/seq2seq bottleneck already encountered
    -> dynamic weighted retrieval (attention)
    -> self-attention
    -> multi-head attention
    -> position + residual + norm + FFN
    -> Transformer block
    -> stacked causal decoder
    -> next-token logits
    -> autoregressive sampling
    -> TinyLM
```

At no point should a major component exist only because "Transformers have one."

## Mathematical spiral

The mathematics should accumulate/recur as follows:

| Module | Mathematical emphasis |
| --- | --- |
| 15 | categorical vectors, embeddings, similarity, conditional probability, sequence factorization, surprisal/entropy/cross-entropy/KL |
| 16 | dot products, matrix products, weighted sums, softmax, scaling, masks, tensor shapes |
| 17 | positional vectors, limited trigonometry, mean/variance normalization, residual composition, tensor shapes |
| 18 | sequence indexing, token-level loss, categorical sampling, logit transformations, experiment statistics |

This tranche should not require a new detached math module. Earlier linear algebra, probability, calculus, and neural-network mathematics should now feel like working tools.

## Historical spine

The historical progression should mirror the technical motivation:

```text
statistical / n-gram language modeling
        -> neural probabilistic language models
        -> learned word representations
        -> recurrent neural language/seq2seq models
        -> neural attention
        -> self-attention
        -> 2017 Transformer
        -> decoder-only autoregressive Transformer language models
```

History should explain why architectures changed rather than simply attach dates and names to lessons.

## Foundational implementation checkpoints

By the end of this tranche the learner should personally have written the essential mechanics of:

- embedding lookup/representation handling;
- a simple neural next-token model;
- autoregressive generation;
- scaled dot-product attention;
- causal masking;
- multi-head attention;
- a Transformer block;
- causal language-model batching/training;
- temperature sampling;
- top-k sampling;
- top-p sampling;
- a complete TinyLM training/generation path.

## Tranche mastery checkpoints

### Attention calculation + implementation — Module 16

This is a **foundational implementation checkpoint**, not a separate major course project. Evidence should include both hand-worked mathematics and executable code. A learner who can call an attention library but cannot calculate a tiny example has not met the foundational objective.

### Transformer block from scratch — Module 17

This is another **foundational implementation checkpoint**, not a separate major course project. The learner should implement and test one complete block using lower-level framework primitives, with explicit shape and causal-behavior reasoning.

### TinyLM — Module 18

This is the tranche's **major transfer/synthesis project**. Completion should provide strong evidence across representation, attention, architecture, optimization, experimentation, and generation objectives.

## Explicit boundary before Module 19

After Module 18 the learner understands the **basic computational skeleton** of a Transformer language model. They do **not** yet understand why a modern 2026 LLM differs from that tiny model in architecture, data, scale, post-training, evaluation, inference, or systems behavior.

That gap is intentional. Module 19 should now be able to ask:

> We understand the original ingredients. What changed between this understandable TinyLM and modern large language models, and what problem did each change solve?

That is the correct handoff into the final core tranche.