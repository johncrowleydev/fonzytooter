# AI/ML detailed syllabus plan: Modules 08–14

This document expands Modules 08 through 14 from the canonical [`ai-ml-syllabus.md`](ai-ml-syllabus.md) into a lesson-level teaching plan.

It is a **planning document**, not runtime-authored curriculum. Exact lesson files, objective metadata, worksheets, interactives, notebook repositories, sources, project specifications, and exercise definitions are authored later under `curriculum/courses/ai-ml/`.

The purpose of this tranche is to bridge the first-model foundation into deep learning. Modules 08–09 broaden the learner's understanding of machine learning beyond linear/logistic models. Modules 10–13 then build neural networks from first principles before introducing framework abstractions. Module 14 gives enough architectural breadth in convolutional and recurrent networks to make the later emergence of attention technically and historically motivated.

## Scope guardrails

This tranche should broaden the learner's model vocabulary without becoming a survey course, and deepen neural-network understanding without front-loading specialist computer vision or NLP.

The design rules are:

- classical ML breadth should teach genuinely different modeling ideas, not enumerate estimators;
- every new model family should be connected to the experimental discipline established in Module 07;
- neural networks should emerge from affine transformations, nonlinear functions, composition, gradients, and optimization already learned earlier;
- backpropagation should be understood mechanically before automatic differentiation is treated as a framework primitive;
- the learner should personally build a small trainable neural network before PyTorch hides the mechanics;
- CNNs and recurrent networks should be taught deeply enough to understand the structural problems they solve and the limitations that lead toward attention;
- specialist depth belongs in optional follow-on paths unless it enables a later core objective.

The center of gravity is Modules **10–13**. Modules 08–09 provide important breadth, but they should not delay the neural-network spine with an encyclopedic classical-ML detour.

## Tranche overview

| Module | Planned lessons | Primary job |
| --- | ---: | --- |
| 08 — Classical ML Beyond Linear Models | 6 | Learn several fundamentally different ways models can fit data |
| 09 — Unsupervised Learning and Dimensionality Reduction | 6 | Introduce unlabeled learning and linear algebra II through PCA |
| 10 — Neural Networks as Composed Functions | 5 | Make neural networks a natural extension of earlier models |
| 11 — Computational Graphs and Backpropagation | 6 | Understand reverse-mode differentiation and gradient flow |
| 12 — Neural Network From Scratch | 5 | Build and train a small network without a deep-learning framework |
| 13 — PyTorch and Training Neural Networks | 7 | Map known mechanics onto modern framework abstractions |
| 14 — Deep Learning Architectures: CNNs and Sequences | 7 | Learn architectural inductive bias and reach the seq2seq bottleneck |

This is approximately 42 lessons or project-oriented lesson checkpoints. Lesson count is a planning estimate, not a product requirement. Project-heavy modules may ultimately contain fewer long-form explanatory lessons and more guided build checkpoints.

## Dependency shape

Recommended teaching order remains Modules 08 through 14, but objective metadata should encode only real prerequisites.

```text
                 07 Generalization & Evaluation
                    /                  \
                   v                    v
          08 Classical ML        09 Unsupervised / PCA
                   \                  /
                    \                /
                     \              /
                      \            /
                       v          v
                 broader ML experience

02 Linear Algebra + 04 Calculus + 06 Classification + 07 Evaluation
                              |
                              v
                 10 Neural Networks
                              |
                              v
                    11 Backpropagation
                              |
                              v
                 12 NN From Scratch
                              |
                              v
                      13 PyTorch
                              |
                              v
                  14 CNNs & Sequences
```

Important consequences:

- Modules 08 and 09 are part of the recommended sequence but should not become hard prerequisites for every neural-network objective.
- Module 10 genuinely depends on earlier affine-model, vector/matrix, loss, optimization, and classification concepts.
- Module 11 depends strongly on Module 10 and on the derivative/chain-rule foundation from Module 04.
- Module 12 depends on Module 11 because the project must not treat backpropagation as unexplained machinery.
- Module 13 depends on Module 12 pedagogically: framework abstractions should arrive after the learner has implemented the mechanics they automate.
- Module 14 depends on framework-level neural-network competence but should not require specialist vision or sequence-modeling knowledge.
- PCA-specific eigenvector knowledge from Module 09 is not a prerequisite for neural networks merely because Module 09 appears earlier.

## Learning-medium emphasis

| Module | Most important media |
| --- | --- |
| 08 | interactives, notebooks, small tracing worksheets |
| 09 | geometric interactives, worksheets, NumPy implementation, notebooks |
| 10 | hand forward-pass worksheets, shape interactives, embedded NumPy exercises |
| 11 | substantial hand-worked worksheets, computational-graph interactives, embedded backward-pass exercises |
| 12 | repository project, tests, debugging/reflection |
| 13 | notebooks and repository work using PyTorch |
| 14 | visual interactives plus focused image/sequence experiments |

---

# 08 — Classical ML Beyond Linear Models

## Purpose

Show that machine learning is not synonymous with differentiable linear models or neural networks. The learner should encounter several substantially different ways a model can make predictions, understand the inductive biases those methods introduce, and learn to compare model families under a common evaluation protocol.

This module should build breadth without becoming a catalog of scikit-learn estimators.

## Prerequisites

Required capabilities from earlier modules:

- supervised learning, features, targets, predictions, and loss;
- regression and classification;
- train/validation/test discipline;
- overfitting, regularization, metrics, leakage, and reproducibility;
- vector/distance intuition sufficient for nearest-neighbor methods;
- basic probability sufficient to interpret class proportions and impurity measures conceptually.

## Proposed core objectives

Objective IDs are provisional until authored. Capabilities should cover:

- explain how decision trees partition feature space and produce predictions;
- trace a small tree by hand and reason about a candidate split;
- explain why unrestricted trees overfit and how bagging/random forests reduce variance;
- explain boosting as sequential error-correction using weak learners at a conceptual level;
- use distance-based prediction and explain why feature scaling matters;
- compare linear, tree-based, ensemble, and neighbor-based models under a controlled experimental protocol;
- choose a model family for a simple problem using evidence rather than fashion.

## Lesson sequence

### 01 — Beyond Linear Decision Boundaries

**Job:** Motivate why additional model families are needed before naming algorithms.

Topics:

- what linear regression/logistic regression can and cannot represent;
- nonlinear relationships and nonlinear decision boundaries;
- model capacity and inductive bias revisited;
- global parametric rules versus local/partition-based behavior;
- why a more flexible model is not automatically a better model;
- maintaining the Module 07 discipline of held-out evaluation.

A simple 2D dataset should make the limitation visually obvious. The lesson should end with the question: if one global line is the wrong shape, what other ways could a model organize the space?

### 02 — Decision Trees

**Job:** Introduce recursive partitioning as a fundamentally different modeling strategy.

Topics:

- decision nodes and leaves;
- threshold splits on features;
- recursive partitioning of feature space;
- classification versus regression leaves;
- purity/impurity as a way to evaluate splits;
- Gini or entropy at the level needed to understand split quality;
- depth and minimum-sample constraints;
- tracing a prediction path.

The learner should calculate at least a few tiny split-quality examples, but the module should not derive every criterion or implement an industrial tree builder from scratch.

### 03 — From One Tree to a Random Forest

**Job:** Use tree instability to motivate ensemble averaging and bagging.

Topics:

- how a deep tree can fit idiosyncrasies of a dataset;
- variance and sensitivity to sampled data;
- bootstrap samples;
- bagging;
- averaging/voting across models;
- random feature subsets;
- random forests;
- out-of-bag intuition as optional enrichment, not a central objective;
- why many noisy models can collectively become more stable.

The emphasis should be on the statistical intuition rather than implementation details of a production random-forest library.

### 04 — Boosting

**Job:** Contrast parallel averaging ensembles with sequential ensembles that focus on previous errors.

Topics:

- weak learners;
- sequential model construction;
- reweighting or fitting residual/error signal conceptually;
- additive prediction;
- AdaBoost as historical/conceptual context;
- gradient boosting as the important modern family;
- why boosting can be extremely strong on tabular data;
- overfitting and hyperparameter tradeoffs.

Do not turn this into a full derivation of gradient boosting. The learner should understand the mechanism well enough to explain why boosting differs from bagging and when it might work well.

### 05 — Nearest Neighbors and Distance-Based Prediction

**Job:** Introduce a model whose prediction comes directly from similar observed examples.

Topics:

- nearest-neighbor classification and regression;
- choosing `k`;
- majority vote/averaging;
- Euclidean distance revisited;
- feature scale and why meters can dominate centimeters/normed variables if untreated;
- standardization at a practical level;
- local decision boundaries;
- computational and high-dimensional limitations;
- the curse of dimensionality preview without a mathematical detour.

This lesson should reinforce that preprocessing requirements depend on model family: scaling matters greatly for distance-based methods but not in the same way for ordinary decision trees.

### 06 — Comparing Model Families Fairly

**Job:** Synthesize the module through controlled model comparison rather than algorithm collection.

Topics:

- a common train/validation/test protocol;
- scikit-learn estimators and pipelines as practical abstractions;
- preprocessing tied to model requirements;
- choosing comparable metrics;
- hyperparameters versus learned parameters;
- simple hyperparameter search without introducing an optimization framework;
- interpretation of model behavior and failure cases;
- predictive performance, simplicity, interpretability, data size, and compute as competing considerations;
- why there is no universally best model.

The notebook should compare at least a linear model, tree, random forest or boosted tree, and nearest-neighbor model on the same problem.

## Mathematics

New mathematics should remain light:

- impurity measures as summaries of class mixture;
- weighted averages/votes;
- distance and scale revisited;
- simple bootstrap sampling intuition;
- bias/variance intuition revisited through ensembles.

Entropy may appear as one tree split criterion, but this is not the place for a general information-theory unit.

## History

History should explain why these ideas represent different traditions in machine learning:

- recursive partitioning and the development of modern classification/regression trees;
- ensemble methods and the insight that model diversity can improve generalization;
- bagging/random forests;
- boosting and the surprising power of combining weak learners;
- nearest-neighbor methods as a simple, long-lived example of instance-based learning.

Avoid a chronology of named algorithms that does not illuminate technical ideas.

## Interactives

Strong candidates:

- **Decision-boundary gallery** — same dataset, different model families;
- **Tree split explorer** — drag a threshold and see resulting partitions/impurity;
- **Random-forest ensemble visualizer** — several unstable trees versus averaged boundary;
- **k-nearest-neighbor explorer** — move a query point or change `k` and see neighbors/prediction.

## Worksheets

Useful worksheet topics:

- trace several examples through a small decision tree;
- calculate a tiny split impurity and compare candidate splits;
- reason about bootstrap samples and ensemble voting;
- calculate nearest-neighbor distances;
- identify when feature scaling will change neighbor relationships;
- diagnose unfair model-comparison setups.

This should be lighter on worksheets than Modules 02–06.

## Embedded exercises

Candidates:

- implement prediction for a pre-authored small tree representation;
- compute a simple impurity score;
- implement a tiny k-nearest-neighbor predictor;
- standardize a feature matrix from provided mean/std values;
- compare predictions for different `k` values;
- identify leakage or preprocessing mistakes in a small pipeline.

A full decision-tree training algorithm is deliberately not a core hand-written implementation.

## Jupyter lab

**One dataset, several inductive biases**

Use the same dataset and split strategy to train several model families. The learner should visualize boundaries where possible, compare held-out metrics, inspect obvious failure cases, and explain why the models behave differently.

The conclusion should be explanatory rather than leaderboard-oriented.

## Mastery expectations

After Module 08 the learner should be able to:

- explain at least four genuinely different modeling strategies;
- trace and reason about decision-tree behavior;
- explain bagging/random forests and boosting without treating them as magic;
- use and reason about nearest-neighbor prediction;
- identify preprocessing requirements tied to model family;
- compare models under a fair experimental protocol;
- resist equating "more sophisticated" with "better."

## Deliberately deferred

- exhaustive scikit-learn estimator coverage;
- SVMs as a core deep-dive;
- full derivations of tree criteria;
- implementing a production decision-tree or boosting library;
- advanced gradient-boosting systems and tuning practice;
- interpretability frameworks such as SHAP as a major topic;
- AutoML/hyperparameter-optimization platforms.

---

# 09 — Unsupervised Learning and Dimensionality Reduction

## Purpose

Introduce learning problems where no labeled target is supplied, then revisit linear algebra because high-dimensional structure creates a real need for projections, covariance geometry, eigenvectors, and principal components.

The module should make unsupervised learning feel less like "classification without labels" and more like a different problem: discovering useful structure under an authored objective.

## Prerequisites

Required capabilities:

- vectors, matrices, dot products, norms, distance, and matrix multiplication;
- NumPy arrays and vectorized operations;
- probability/statistical ideas including means, variance, and covariance intuition;
- experimental discipline from Module 07.

No neural-network knowledge is required.

## Proposed core objectives

Capabilities should cover:

- distinguish supervised and unsupervised learning objectives;
- explain and execute the k-means update procedure;
- personally implement a small k-means loop;
- reason about distance, scale, and cluster assignments;
- understand projection and orthogonality geometrically;
- interpret covariance as directional co-variation;
- explain eigenvectors/eigenvalues at the depth needed for PCA;
- use PCA to project data and interpret variance retained/lost;
- recognize that cluster and principal-component structure are modeling choices, not discovered ground truth.

## Lesson sequence

### 01 — Learning Without Targets

**Job:** Establish what an unsupervised objective is and what it can and cannot claim.

Topics:

- labeled versus unlabeled data;
- structure-finding objectives;
- clustering, density, and representation as broad families;
- similarity as an authored assumption;
- why different objectives can reveal different "structure" in the same data;
- evaluation challenges without ground-truth labels;
- examples of useful exploratory tasks.

This lesson should actively guard against anthropomorphic language such as "the algorithm discovers the real groups."

### 02 — K-Means: Clusters and Centroids

**Job:** Introduce a simple unsupervised optimization loop.

Topics:

- centroids;
- assignment to the nearest centroid;
- recomputing centroids;
- within-cluster squared distance objective;
- alternating assignment/update steps;
- convergence to a local solution;
- the role of `k`;
- sensitivity to initialization and scale;
- cluster-shape assumptions.

Small examples should be calculated by hand before code.

### 03 — K-Means From Scratch

**Job:** Consolidate k-means through a learner-written implementation.

Topics:

- representing centroids and assignments with arrays;
- vectorized distance calculations;
- assignment step;
- centroid update step;
- stopping criteria;
- random initialization and deterministic seeds;
- empty-cluster handling at a simple level;
- comparing the implementation against a library result.

This is the tranche's small classical hand-written algorithm checkpoint.

### 04 — Projection, Orthogonality, and Covariance Geometry

**Job:** Build the geometric machinery PCA actually needs before introducing PCA itself.

Topics:

- projecting a vector onto a direction;
- dot product as projection-related measurement revisited;
- unit vectors;
- orthogonal directions;
- centered data;
- variance along a direction;
- covariance matrix as a summary of how dimensions vary together;
- visualizing an elongated point cloud and asking which direction captures most variation.

The lesson should stay concrete and geometric. Abstract vector-space proofs remain unnecessary.

### 05 — Eigenvectors and Eigenvalues, Just in Time

**Job:** Introduce only the eigenvector/eigenvalue intuition needed to understand principal directions.

Topics:

- a matrix transformation usually changes a vector's direction;
- eigenvectors as special directions preserved up to scale;
- eigenvalues as the associated scale factor;
- covariance matrices and principal directions;
- orthogonal eigenvectors for the covariance/PCA setting at an intuitive level;
- interpreting large versus small eigenvalues as high versus low variance directions;
- numerical computation rather than hand-solving large characteristic polynomials.

The learner should work one or two tiny eigenvector examples, but the module should not become a determinant/characteristic-polynomial course.

### 06 — Principal Component Analysis and High-Dimensional Data

**Job:** Assemble centering, covariance, eigen-directions, and projection into PCA.

Topics:

- principal components;
- ordering components by explained variance;
- projecting into a lower-dimensional coordinate system;
- reconstruction intuition;
- explained variance ratio;
- visualization of high-dimensional data;
- information loss and tradeoffs;
- PCA as an unsupervised linear projection, not a semantic understanding engine;
- feature scaling choices before PCA;
- modern software often uses numerically robust decompositions internally without requiring the learner to derive them.

This lesson should also serve as the classical-ML synthesis point for Modules 08–09.

## Mathematics

This is **linear algebra II**, narrowly motivated by PCA:

- projection;
- unit vectors and orthogonality;
- centered data;
- covariance matrix;
- eigenvector/eigenvalue intuition;
- variance captured along directions;
- change of coordinates/projection into a lower-dimensional basis.

Determinants, diagonalization proofs, SVD derivations, and abstract spectral theory remain deferred.

## History

Useful thread:

- early clustering methods and the long history of grouping observations by similarity;
- Pearson/Hotelling and the development of principal-component methods;
- dimensionality reduction as a response to multivariate data rather than an invention of deep learning.

The history should reinforce that many modern representation problems have older statistical and geometric roots.

## Interactives

Strong candidates:

- **k-means stepper** — alternate assignment and centroid-update phases manually;
- **projection explorer** — rotate a direction through a point cloud and see projected coordinates/variance;
- **PCA variance explorer** — rotate candidate axes and watch explained variance change;
- **dimensionality slider** — reconstruct data using one versus more principal components.

## Worksheets

Likely worksheets:

- hand k-means assignments and centroid updates;
- vector projection practice;
- covariance interpretation from small datasets;
- tiny eigenvector/eigenvalue examples focused on meaning;
- PCA reasoning: choose principal directions and explain what is lost in projection.

## Embedded exercises

Candidates:

- compute pairwise distances to centroids;
- implement one k-means assignment/update step;
- complete the full small k-means loop;
- center a matrix of observations;
- project observations onto a supplied direction;
- compute explained-variance fractions from supplied eigenvalues.

## Jupyter work

Two useful labs may be warranted:

**Clustering lab:** explore initialization, scale, `k`, and failure cases of k-means on intentionally different cluster shapes.

**PCA lab:** visualize a 2D/3D dataset, project it to principal directions, compare reconstruction, then apply PCA to a moderately higher-dimensional real dataset.

## Classical-ML synthesis checkpoint

At the end of Module 09, complete a small study that combines the experimental habits of Module 07 with the breadth of Modules 08–09. It may include supervised model comparison, preprocessing, dimensionality reduction where appropriate, and an unsupervised exploratory component.

The goal is not to use every technique. The learner should justify which techniques are appropriate and which are not.

## Mastery expectations

After Module 09 the learner should be able to:

- implement and explain k-means;
- distinguish a useful clustering result from a claim that clusters are objectively "real";
- reason geometrically about projection and variance;
- explain eigenvectors/eigenvalues in the specific context of PCA;
- explain and use PCA without treating library output as magic;
- make defensible preprocessing and evaluation choices for classical ML experiments.

## Deliberately deferred

- spectral clustering;
- DBSCAN and clustering-algorithm catalogs;
- Gaussian mixture models/EM as a core topic;
- manifold-learning surveys such as t-SNE/UMAP as major units;
- matrix-factorization theory;
- SVD derivation;
- deep unsupervised/representation learning;
- abstract linear-algebra proofs.

---

# 10 — Neural Networks as Composed Functions

## Purpose

Make neural networks feel like a direct extension of concepts the learner already understands: affine transformations, parameters, nonlinear functions, vector/matrix operations, classification losses, and function composition.

Nothing in this module should be presented as a mysterious "brain-like" primitive. A network is a parameterized function built from ordinary mathematical operations.

## Prerequisites

Core prerequisites:

- affine functions and parameterized models;
- matrix-vector/matrix-matrix multiplication;
- function composition;
- regression/classification loss functions;
- gradients and gradient descent conceptually;
- probability/logistic-regression intuition;
- train/validation/test discipline.

Modules 08–09 are useful breadth but should not create artificial hard prerequisites for neural-network objectives.

## Proposed core objectives

Capabilities should cover:

- explain a neuron as an affine transformation followed by a nonlinear activation;
- explain why stacking only affine transformations cannot create a genuinely nonlinear model;
- compute a small neural-network forward pass by hand;
- translate a layer between scalar equations, matrix notation, and NumPy code;
- reason about weight, bias, activation, and batch shapes;
- explain hidden layers and learned representations at an introductory level;
- distinguish network architecture from learned parameter values.

## Lesson sequence

### 01 — From Linear Models to Artificial Neurons

**Job:** Reuse logistic/perceptron structure to introduce the neuron without mystique.

Topics:

- affine score `wx + b` revisited;
- multiple inputs and weighted sums;
- threshold/perceptron historical model;
- differentiable activations in modern networks;
- weights, biases, inputs, and outputs;
- one neuron as a tiny parameterized function;
- terminology versus biology: "neuron" is a historical metaphor, not a claim of biological realism.

### 02 — Why Neural Networks Need Nonlinearity

**Job:** Show mathematically why multiple linear/affine layers alone collapse to another affine transformation.

Topics:

- composition of affine transformations;
- why depth without nonlinear activation adds no expressive power of the desired kind;
- sigmoid, tanh, and ReLU;
- piecewise-linear intuition for ReLU networks;
- activation choice as part of architecture;
- saturation intuition for sigmoid/tanh;
- no need yet for exhaustive activation catalogs.

A simple XOR-style example can motivate why a nonlinear representation is needed.

### 03 — Layers as Vectorized Functions

**Job:** Move from individual neurons to matrix-based layers.

Topics:

- input vector;
- weight matrix;
- bias vector;
- `z = Wx + b`;
- activation applied element-wise;
- dimensions and parameter counts;
- multiple output units;
- batching and the distinction between feature/unit/batch axes;
- translating notation into NumPy.

Shape reasoning should be explicit and heavily practiced.

### 04 — Multilayer Perceptrons and Learned Representations

**Job:** Explain what hidden layers contribute conceptually.

Topics:

- stacking layers;
- hidden units;
- hidden representations/features;
- architecture versus parameters;
- output layers for regression and classification;
- logits and probability conversions revisited where needed;
- capacity and overfitting revisited;
- representation learning as learning intermediate transformations rather than hand-authoring all features.

Avoid universal-approximation theorem proof. The learner needs intuition about expressive composition, not a theorem detour.

### 05 — Forward Passes and Network Shape Reasoning

**Job:** Consolidate the module through substantial hand calculation and code.

Topics:

- tracing a complete two-layer network;
- intermediate pre-activations and activations;
- output/logit/loss pipeline;
- parameter shapes;
- batch shapes;
- parameter counting;
- separating forward computation from the later backward computation;
- debugging shape mismatches.

The module should end with the learner able to implement a forward-only MLP in NumPy before gradients are introduced.

## Mathematics

This module revisits rather than introduces most mathematics:

- affine transformations;
- matrix multiplication;
- vectorized function composition;
- nonlinear scalar functions applied element-wise;
- logits/probabilities/losses;
- tensor/shape reasoning at an introductory level.

The key new mathematical insight is that nonlinear composition creates a richer family of functions than affine composition alone.

## History

History should be tightly integrated and should preserve the parallel symbolic/connectionist story rather than reducing early AI to a single neural-network lineage:

- Turing and early questions about machine intelligence;
- Dartmouth and the emergence of artificial intelligence as a named field;
- symbolic reasoning/search as a dominant early paradigm;
- McCulloch–Pitts as an early formal neuron idea;
- Rosenblatt's perceptron and early connectionist optimism;
- symbolic and connectionist approaches as competing traditions rather than a clean chronological replacement;
- the real representational limitations of single-layer perceptrons without repeating the simplistic myth that one book single-handedly "killed neural networks";
- expert systems and the commercial success of symbolic knowledge-based AI;
- brittleness, knowledge-engineering cost, compute/data constraints, inflated expectations, and funding cycles as multiple contributors to AI winters;
- why multilayer differentiable networks change the representational story;
- the later statistical/neural resurgence as enabled by algorithms, data, compute, and engineering rather than one isolated breakthrough.

The recurring historical tension should remain visible through the rest of the course: hand-authored symbolic structure versus learned statistical representations. Module 21 can later revisit a modern version of this tension when learned models are combined with deterministic retrieval and tools.

## Interactives

Strong candidates:

- neuron weighted-sum/activation explorer;
- activation-function explorer;
- network-shape visualizer;
- small 2D MLP decision-boundary visualizer;
- layer composition visualizer showing intermediate transformed representations.

## Worksheets

This should be worksheet-heavy:

- neuron forward values;
- activation calculations;
- matrix layer forward passes;
- parameter shapes/counts;
- complete two-layer forward passes;
- classify architecture versus parameter questions;
- show algebraically that two affine transformations compose to another affine transformation.

## Embedded exercises

Candidates:

- implement ReLU;
- implement a dense layer forward pass;
- implement a two-layer network forward pass;
- handle a batch dimension correctly;
- calculate parameter counts from supplied dimensions;
- debug intentionally wrong matrix shapes.

## Mastery expectations

After Module 10 the learner should be able to:

- explain a neural network using ordinary mathematical functions;
- calculate forward passes by hand;
- implement a forward-only MLP with NumPy;
- reason confidently about layer and parameter shapes;
- explain why nonlinear activations are necessary;
- explain hidden representations without mystical language;
- identify exactly which remaining problem Module 11 must solve: how to efficiently compute parameter gradients through all these composed operations.

## Deliberately deferred

- backpropagation mechanics;
- automatic differentiation;
- PyTorch;
- convolution/recurrent architectures;
- universal-approximation proofs;
- large-network architecture catalogs;
- biological-neural-network detail.

---

# 11 — Computational Graphs and Backpropagation

## Purpose

Revisit the chain rule at the depth required to understand how a scalar loss produces gradients for every parameter in a composed neural network.

This is a foundational module. Backpropagation should not be reduced to a memorized algorithm or a call to `loss.backward()`.

## Prerequisites

Required capabilities:

- derivative, partial derivative, gradient, and chain-rule intuition from Module 04;
- composed functions and neural-network forward passes from Module 10;
- matrix/vector shape reasoning;
- loss functions and gradient-descent updates.

## Proposed core objectives

Capabilities should cover:

- represent a composed computation as a directed computational graph;
- calculate local derivatives for graph operations;
- apply the chain rule across multiple paths;
- explain reverse-mode automatic differentiation conceptually;
- manually backpropagate through a small neuron/network;
- reason about gradient accumulation when a value influences the loss through multiple paths;
- distinguish forward values from backward sensitivities/gradients;
- verify an analytic gradient with finite differences;
- explain what a modern autodiff system is doing at a conceptual level.

## Lesson sequence

### 01 — Computation as a Graph

**Job:** Make composition visually explicit.

Topics:

- values as graph nodes/edges and operations as transformations;
- forward evaluation order;
- intermediate values;
- scalar loss at the end of a graph;
- one expression represented as nested algebra, code, and graph;
- why graph structure makes dependency paths visible.

Use very small scalar expressions before neural-network graphs.

### 02 — Local Derivatives and the Chain Rule Revisited

**Job:** Turn the Module 04 chain rule into a graph procedure.

Topics:

- local derivative of an operation with respect to each input;
- upstream sensitivity;
- multiplying along a path;
- composed scalar examples;
- partial derivatives at branching operations;
- evaluating derivatives using stored forward values;
- keeping notation disciplined.

The learner should perform enough problems that the procedure becomes mechanical without becoming rote symbol manipulation detached from meaning.

### 03 — Reverse-Mode Differentiation

**Job:** Explain why gradients are propagated from the scalar loss backward.

Topics:

- forward mode versus reverse mode conceptually;
- many parameters feeding one scalar loss;
- reverse traversal;
- storing/interpreting adjoints or gradients;
- reuse of intermediate derivatives;
- computational efficiency intuition;
- automatic differentiation as exact chain-rule bookkeeping rather than symbolic algebra or numerical differencing.

Formal Jacobian machinery is not required as a separate unit. Where vector-valued operations appear, emphasize shape-aware gradient propagation and vector-Jacobian-product intuition rather than a full matrix-calculus course.

### 04 — Backpropagating Through a Neuron

**Job:** Apply reverse-mode reasoning to a familiar neural computation.

Topics:

- affine pre-activation;
- nonlinear activation;
- loss;
- local derivatives for multiply/add/ReLU or sigmoid;
- gradients with respect to weight, bias, and input;
- parameter update after the backward pass;
- relation to the derivative calculations already used in logistic regression.

A complete tiny example should be worked by hand from prediction through parameter update.

### 05 — Backpropagating Through Layers and Multiple Paths

**Job:** Scale the reasoning to a small multilayer network.

Topics:

- gradients through matrix/dense layers at a practical level;
- per-example and batched gradients conceptually;
- gradient accumulation when parameters/values contribute through multiple paths;
- broadcasting and gradient shape awareness;
- hidden-layer gradients;
- multiple outputs feeding one scalar loss;
- careful dimensional reasoning.

Avoid deriving every matrix-calculus identity in abstraction. Derive the operations needed for the upcoming from-scratch implementation.

### 06 — Gradient Checking and the Autodiff Mental Model

**Job:** Connect manual backpropagation to reliable implementation practice.

Topics:

- finite-difference derivative approximation;
- comparing numerical and analytic gradients;
- tolerance and floating-point error;
- common backward-pass bugs;
- why gradient checking is useful but expensive;
- what an autodiff framework records during forward computation;
- what `backward()` will later trigger;
- gradients as accumulated state in many frameworks.

This lesson should end with a tiny operation graph whose hand-written backward values are checked numerically.

## Mathematics

This is **calculus II for the course**, centered on reverse differentiation:

- chain rule at deeper operational level;
- partial derivatives through branches;
- gradient accumulation;
- derivative shape reasoning for vectors/matrices;
- finite-difference approximation;
- reverse-mode differentiation.

Formal multivariable-calculus theorem work, Hessians, and full Jacobian algebra remain deferred.

## History

The history should be accurate and nuanced:

- reverse-mode differentiation has roots older than modern deep learning;
- ideas closely related to backpropagation appeared in control/optimization and earlier neural-network work;
- backpropagation was developed/reintroduced in multiple places and became widely influential through 1980s neural-network research;
- the key historical point is not a single invention date but the combination of differentiable networks, efficient reverse differentiation, data, compute, and later engineering advances.

## Interactives

Strong candidates:

- computational-graph forward/backward stepper;
- chain-rule path highlighter;
- gradient-accumulation visualizer for branching graphs;
- numerical-versus-analytic gradient comparison.

## Worksheets

This should be one of the most worksheet-intensive modules in the course:

- local derivatives;
- scalar computational graphs;
- chain-rule paths;
- branching/gradient accumulation;
- one-neuron backpropagation;
- two-layer backpropagation;
- parameter updates after calculated gradients;
- gradient-check reasoning.

Showing work matters heavily here because the earliest wrong derivative or path is more informative than the final number alone.

## Embedded exercises

Candidates:

- implement backward methods for add, multiply, ReLU, and sigmoid operations;
- propagate gradients through a tiny graph;
- implement a neuron backward pass;
- accumulate gradients from multiple paths;
- compare an analytic gradient with finite differences;
- debug an intentionally incorrect backward rule.

## Mastery expectations

After Module 11 the learner should be able to:

- explain backpropagation as reverse-mode chain-rule computation;
- manually differentiate a small network;
- reason about why gradients accumulate;
- implement backward rules for basic operations;
- verify gradients numerically;
- explain what automatic differentiation will automate and what it will not.

## Deliberately deferred

- full matrix calculus;
- Hessians and second-order methods;
- advanced autodiff implementation internals;
- symbolic differentiation systems;
- implicit differentiation;
- distributed gradient computation.

---

# 12 — Neural Network From Scratch

## Purpose

Consolidate Modules 10–11 into a complete trainable neural network built without a deep-learning framework hiding the core mechanics.

This is a major transfer/mastery checkpoint. The learner should personally produce the foundational implementation before using agents for review, debugging, testing, or refactoring.

## Prerequisites

Required capabilities:

- forward-pass and shape reasoning from Module 10;
- backpropagation from Module 11;
- gradient descent and learning-rate behavior;
- loss functions;
- NumPy/vectorized computation;
- evaluation discipline from Module 07.

## Proposed core objectives

Capabilities should cover:

- design a small multilayer network with explicit parameter shapes;
- initialize parameters sensibly enough to train a small network;
- implement forward and backward passes;
- calculate loss and parameter gradients;
- train with mini-batches;
- update parameters with a basic optimizer;
- gradient-check critical operations;
- separate training and evaluation behavior;
- diagnose common failure modes using loss curves, predictions, and gradient checks;
- explain every major part of the implementation without framework terminology as a substitute for understanding.

## Lesson/project sequence

This module should feel more like a guided build than five conventional textbook chapters.

### 01 — Designing the Small Network

**Job:** Define the architecture and implementation boundaries before coding the whole system.

Topics:

- choose a deliberately small supervised task;
- input/output dimensions;
- one or two hidden layers;
- activation choice;
- output/loss pairing;
- parameter arrays and shapes;
- initialization scale intuition;
- minimal abstractions: enough structure to make forward/backward logic readable without recreating PyTorch.

The implementation should remain transparent. Avoid framework-like abstraction for its own sake.

### 02 — Forward Pass and Loss

**Job:** Build the complete prediction path first.

Topics:

- dense layer forward calculation;
- activation storage;
- intermediate-cache design for the backward pass;
- logits/output;
- loss computation;
- batched inputs;
- deterministic tests of known forward values;
- numerical stability where the chosen loss requires it.

### 03 — Backward Pass

**Job:** Turn the hand-worked Module 11 mechanics into code.

Topics:

- output-loss gradient;
- activation backward pass;
- dense-layer gradients;
- bias gradient;
- gradient propagation to earlier layers;
- storing parameter gradients;
- shape checks;
- comparing selected gradients against finite differences.

### 04 — Mini-Batches and the Training Loop

**Job:** Make the system actually learn.

Topics:

- mini-batch sampling/iteration;
- forward → loss → backward → update cycle;
- averaging gradients/loss over batches where appropriate;
- shuffling;
- learning-rate behavior;
- epoch-level metrics;
- train/evaluation split;
- simple reproducibility controls;
- recognizing a decreasing training loss as necessary but not sufficient evidence.

### 05 — Debugging, Validation, and Final From-Scratch Network

**Job:** Turn a working script into a trustworthy learning artifact.

Topics:

- gradient checking;
- overfit-a-tiny-batch sanity test;
- inspect activations/gradients;
- distinguish code bugs from optimization problems;
- compare with a simple baseline;
- held-out evaluation;
- document architecture, training behavior, known limitations, and lessons learned;
- optional refactoring only after mechanics are correct and understandable.

## Mathematics

No major new mathematics should be introduced. This is synthesis:

- affine transformations;
- nonlinear composition;
- loss functions;
- chain rule/backpropagation;
- gradient descent;
- mini-batch averages;
- initialization scale intuition.

The module's difficulty should come from integrating known ideas, not introducing another theoretical layer.

## History

History should be minimal here. The conceptual/history work belongs in Modules 10–11. Short context about why small hand-written neural-network implementations remain pedagogically valuable is sufficient.

## Worksheets/interactives

Use these sparingly. If the learner needs repeated backpropagation practice, Module 11 should supply it. Module 12 should prioritize implementation and transfer.

A useful pre-build architecture/shape worksheet may help verify parameter dimensions before coding.

## Repository project

**Neural Network From Scratch**

The learner writes a small NumPy-based neural network capable of fitting a nontrivial toy problem. The repository should include tests for important operations, a reproducible training entry point or notebook/script, held-out evaluation, and a short technical reflection.

Exact project dataset and repository specification belong to later content authoring, not this planning document.

## Mastery expectations

Completion should demonstrate **transfer**, not merely successful execution. The learner should be able to explain:

- what every parameter array represents;
- why each forward operation exists;
- where every backward gradient comes from;
- how gradients reach earlier layers;
- what the optimizer changes;
- how mini-batches affect the update process;
- how they know the implementation is probably correct;
- how they know the model generalizes beyond its training examples.

## Deliberately deferred

- PyTorch or other deep-learning framework abstractions;
- convolution/recurrent layers;
- GPU acceleration;
- advanced optimizers;
- production-grade neural-network libraries;
- distributed training;
- complex datasets.

---

# 13 — PyTorch and Training Neural Networks

## Purpose

Introduce a modern deep-learning framework only after the learner has implemented the mechanics it automates. Every important PyTorch abstraction should be mapped back to a concept already understood from the from-scratch network.

The theme is not "learn the PyTorch API." It is **understand what the framework is doing for you and learn the disciplined workflow required to train neural networks reliably.**

## Prerequisites

Required capabilities:

- complete small-network mechanics from Module 12;
- NumPy arrays and shape reasoning;
- backpropagation/autodiff mental model;
- gradient-based optimization;
- evaluation/reproducibility practices.

## Proposed core objectives

Capabilities should cover:

- use PyTorch tensors while reasoning about shape, dtype, device, and gradient tracking;
- explain how autograd records and computes gradients;
- define models with `nn.Module` and inspect parameters;
- construct datasets/dataloaders and mini-batch training loops;
- use SGD, momentum, Adam, and AdamW with conceptual understanding;
- distinguish ordinary L2 regularization from decoupled weight decay under adaptive optimization;
- reason about learning-rate warmup and decay schedules and record them as part of an experiment;
- reason about initialization, normalization, and regularization choices;
- move models/data between CPU and accelerator devices correctly;
- save/load checkpoints;
- run reproducible experiments and distinguish training/evaluation modes;
- explain mixed precision and numerical-stability concerns at an introductory level.

## Lesson sequence

### 01 — From NumPy Arrays to PyTorch Tensors

**Job:** Establish the tensor abstraction without pretending it is conceptually new.

Topics:

- tensor shape/dtype/indexing;
- similarities/differences from NumPy arrays;
- tensor creation and conversion;
- operations and broadcasting;
- contiguous/storage awareness only where needed;
- devices;
- CPU versus accelerator placement;
- explicit data movement;
- why framework tensors carry more training/runtime metadata than plain arrays.

### 02 — Autograd

**Job:** Map PyTorch automatic differentiation directly onto Module 11.

Topics:

- `requires_grad`;
- computation graphs built during forward execution;
- `backward()` from scalar loss;
- `.grad` accumulation;
- clearing gradients;
- disabling gradient tracking;
- leaf tensors/parameters at a useful level;
- inspect a tiny gradient and compare with the known hand-derived result.

The lesson should repeatedly ask: what part of the manual implementation is PyTorch doing now?

### 03 — `nn.Module`, Layers, and Parameters

**Job:** Introduce framework model structure as a convenience over the Module 12 architecture.

Topics:

- `nn.Module`;
- registered parameters;
- standard layers;
- `forward`;
- model composition;
- activation modules/functions;
- inspecting parameter names/shapes/counts;
- state dictionaries;
- keeping architecture explicit and readable.

### 04 — Datasets, DataLoaders, and the Training Loop

**Job:** Build the standard supervised-training workflow from known pieces.

Topics:

- dataset abstraction;
- batching;
- shuffling;
- DataLoader;
- training/evaluation loops;
- `model.train()` and `model.eval()`;
- loss functions;
- optimizer step cycle;
- validation;
- metric aggregation;
- device movement inside the loop;
- clean experiment structure.

### 05 — SGD, Momentum, Adam, and AdamW

**Job:** Extend optimization beyond plain gradient descent without turning this into an optimizer zoo, and make common modern training configurations legible.

Topics:

- stochastic/mini-batch gradient descent;
- momentum as accumulated direction/smoothing intuition;
- Adam as adaptive first/second-moment scaling intuition;
- optimizer state;
- ordinary L2 penalty in the objective versus decoupled weight decay under an adaptive optimizer;
- AdamW as the representative modern decoupled-weight-decay formulation;
- optimizer-state memory implications at a basic level;
- learning rate remains important;
- constant learning rate as a baseline;
- warmup and why earliest optimization steps can be unusually fragile;
- decay after warmup;
- cosine decay as a representative commonly encountered schedule;
- comparing optimization trajectories/curves;
- reasons an optimizer that trains faster is not necessarily a universally better choice.

Mathematics should be sufficient to read the update equations and explain the intuition, not derive convergence proofs. The learner should be able to inspect a normal modern training configuration containing AdamW, weight decay, warmup, and decay and explain what each setting is trying to control.

### 06 — Initialization, Normalization, and Regularization

**Job:** Introduce practical training techniques as solutions to concrete optimization/generalization problems.

Topics:

- why initialization scale matters;
- simple Xavier/Glorot and He/Kaiming intuition tied to activation variance;
- L2 regularization and its relationship—but not general equivalence—to weight decay;
- dropout;
- batch normalization concept and training/evaluation behavior;
- normalization as a broader design idea;
- inspecting activation/gradient distributions in a notebook;
- avoid cataloging every normalization variant.

### 07 — Devices, Checkpoints, Mixed Precision, and Reproducible Training

**Job:** Complete the practical framework workflow.

Topics:

- CPU/GPU device selection at an introductory level;
- checkpoints and state dictionaries;
- saving model plus optimizer state where appropriate;
- saving/restoring scheduler state when a schedule is part of training;
- deterministic seeds and limits of exact reproducibility;
- train/eval mode pitfalls;
- floating-point precision revisited;
- FP32/FP16/BF16 conceptually;
- mixed-precision training at a high level;
- overflow/underflow and loss scaling intuition where relevant;
- logging optimizer, weight-decay, warmup, scheduler, and other experiment metadata needed to reproduce a run.

GPU architecture/performance engineering remains for Module 22.

## Mathematics

Math is applied/revisited:

- stochastic gradients;
- momentum;
- moving first/second-moment intuition for Adam;
- AdamW/decoupled weight-decay intuition;
- learning-rate schedules as time-varying optimization step sizes;
- variance propagation intuition for initialization;
- normalization statistics;
- regularization;
- floating-point range/precision.

Do not introduce optimization proofs or numerical-analysis formalism.

## History

Useful context:

- early deep-learning frameworks and static computation graphs;
- the shift toward more imperative/dynamic workflows;
- PyTorch's role in making eager-style research workflows mainstream;
- framework evolution as engineering around the same core tensor/autodiff ideas rather than a replacement for them.

## Interactives

Fewer custom interactives are needed because code/notebooks are the natural medium. Useful candidates only if justified:

- autograd graph inspector tied to a tiny computation;
- optimizer/scheduler trajectory comparison;
- activation-distribution visualization for initialization choices.

## Worksheets

Light use:

- map PyTorch operations back to from-scratch operations;
- calculate one or two SGD/momentum/Adam/AdamW updates;
- distinguish L2 penalty from decoupled weight decay in supplied optimizer setups;
- interpret a warmup/decay schedule;
- parameter-shape/count exercises;
- identify train/eval/checkpoint pitfalls.

## Jupyter/repository work

Key work:

- rebuild the Module 12 model in PyTorch;
- compare manual versus autograd gradients on a tiny example;
- compare optimizers under a controlled setup;
- compare at least one constant-rate run with a simple warmup/decay schedule where useful;
- probe initialization behavior;
- train/evaluate a small framework model with checkpointing and reproducible optimizer/scheduler configuration.

## Mastery expectations

After Module 13 the learner should be able to:

- use PyTorch without treating autograd as magic;
- build and train a small model idiomatically;
- reason about tensor shapes, dtypes, and devices;
- explain the optimizer lifecycle;
- explain SGD, momentum, Adam, and AdamW at the level needed to read training code;
- distinguish L2 regularization from decoupled weight decay;
- explain the purpose of warmup and learning-rate decay and reproduce their configuration;
- explain why initialization/normalization/regularization matter;
- save and restore training state;
- structure a reproducible experiment;
- distinguish framework convenience from underlying ML mechanics.

## Deliberately deferred

- distributed training;
- custom CUDA/Triton kernels;
- compiler internals;
- advanced mixed-precision systems work;
- optimizer and scheduler catalogs;
- large-scale data pipelines;
- architecture-specific specialist tooling.

---

# 14 — Deep Learning Architectures: CNNs and Sequences

## Purpose

Show why dense feed-forward networks are not the only useful neural architecture. Introduce **architectural inductive bias** through convolutional networks for spatial data and recurrent networks for sequences, then follow the historical/technical sequence through LSTM/GRU and encoder-decoder models until the fixed-context bottleneck makes attention an obvious next question.

This module is intentionally broad but bounded. It is not a computer-vision course plus an RNN course compressed into one module.

## Prerequisites

Required capabilities:

- neural-network forward/backward/training mechanics;
- PyTorch tensors/modules/training loops;
- matrix/shape reasoning;
- gradients and optimization;
- generalization/evaluation discipline.

## Proposed core objectives

Capabilities should cover:

- explain architectural inductive bias and why data structure should influence model design;
- explain convolution, locality, receptive fields, and weight sharing;
- calculate a tiny 1D/2D convolution by hand;
- train and evaluate a small CNN;
- explain recurrent hidden state and unrolled sequence computation;
- trace a small RNN across timesteps;
- explain why long recurrent chains create optimization/memory problems;
- explain the purpose of LSTM/GRU gates at a conceptual/mechanical level;
- explain encoder-decoder sequence-to-sequence modeling;
- identify the fixed-context bottleneck that motivates attention.

## Lesson sequence

### 01 — Architecture as Inductive Bias

**Job:** Establish why the shape of a network matters before introducing CNNs/RNNs.

Topics:

- fully connected layers make weak assumptions about structure;
- useful data often has locality, translation structure, order, or temporal dependence;
- parameter sharing;
- architectural constraints as inductive bias;
- why fewer, better-structured parameters can outperform a generic dense network;
- images and sequences as two motivating cases.

### 02 — Convolution, Locality, and Weight Sharing

**Job:** Build convolution from a small sliding-kernel computation.

Topics:

- 1D convolution first if pedagogically cleaner;
- 2D kernels/filters;
- local receptive fields;
- stride;
- padding;
- channels;
- output-shape calculation;
- shared filter weights;
- convolution versus dense parameter count;
- translation-related behavior at an intuitive level.

The learner should calculate several tiny convolutions by hand and predict output shapes.

### 03 — CNNs and Learned Feature Hierarchies

**Job:** Assemble convolutional layers into a meaningful image-model experiment.

Topics:

- stacked convolutions;
- nonlinearities;
- pooling/downsampling at a useful level;
- increasing receptive field;
- feature maps;
- early/late feature hierarchy intuition;
- classifier head;
- train/eval workflow in PyTorch;
- inspect learned filters/activations where useful;
- one small image classification experiment.

Do not expand into detection, segmentation, vision transformers, or modern CNN architecture catalogs.

### 04 — Sequence Data and Recurrent State

**Job:** Explain what changes when input order matters and sequence length varies.

Topics:

- ordered observations;
- hidden state as a carried summary;
- recurrent update equation;
- same parameters reused at each timestep;
- unrolling through time;
- many-to-one and many-to-many patterns;
- recurrent computation versus feed-forward depth;
- small RNN forward trace.

The hidden state should be treated as a learned numerical summary, not a magical "memory."

### 05 — Training RNNs and the Long-Dependency Problem

**Job:** Connect recurrence back to backpropagation and expose why long sequences are difficult.

Topics:

- backpropagation through time conceptually;
- repeated Jacobian/derivative multiplication intuition without formal matrix-calculus overload;
- vanishing gradients;
- exploding gradients;
- gradient clipping conceptually;
- difficulty retaining information across long gaps;
- why naive recurrent state creates both optimization and information bottlenecks.

This is a conceptual explanation, not a full BPTT derivation exercise.

### 06 — LSTM and GRU: Gated Recurrent Memory

**Job:** Show how gated recurrent cells address important shortcomings of simple RNNs.

Topics:

- additive state path intuition;
- gates as learned continuous controls;
- LSTM cell state;
- input/forget/output gates conceptually;
- GRU reset/update gates at a high level;
- why gating helps preserve/use information;
- trace a deliberately simplified gated-cell example;
- tradeoff between understanding the mechanism and memorizing every equation.

The learner should understand what problem the gates solve; memorizing exact library gate ordering is not a core objective.

### 07 — Encoder–Decoder Sequence Models and the Bottleneck

**Job:** End the tranche at the exact problem attention will solve.

Topics:

- variable-length input/output tasks;
- encoder converts an input sequence into a representation;
- decoder generates an output sequence;
- sequence-to-sequence training conceptually;
- teacher forcing at a high level if useful;
- compressing an entire input into one fixed context vector;
- long-sequence degradation;
- inability of the decoder to directly revisit different source positions;
- alignment as the missing capability;
- historical motivation for learned attention.

The lesson should stop before teaching the attention mechanism itself. The desired final question is:

> What if the decoder did not have to rely on one fixed summary, and could learn which encoder states to consult at each output step?

That question opens Module 16 after Module 15 establishes embeddings and language-modeling foundations.

## Mathematics

New/revisited math:

- discrete convolution arithmetic;
- output-shape formulas;
- parameter sharing/counting;
- recurrent difference-equation intuition;
- repeated nonlinear composition through time;
- derivative multiplication and vanishing/exploding gradient intuition;
- gating as element-wise multiplication/interpolation.

No signal-processing Fourier treatment, full BPTT derivation, or advanced dynamical-systems theory is required.

## History

The historical spine should explain architectural evolution:

- early convolutional/neocognitron ideas and LeNet-style CNNs;
- ImageNet-era deep learning as a demonstration of representation learning at scale;
- recurrent networks as a natural approach to sequence data;
- LSTM/GRU as responses to recurrent memory/optimization limitations;
- encoder-decoder seq2seq models;
- the fixed-vector bottleneck and alignment problem immediately preceding attention.

History should make later architecture changes feel like solutions to known constraints rather than arbitrary fashion.

## Interactives

Strong candidates:

- **convolution explorer** — slide a kernel and show each output calculation;
- **CNN receptive-field visualizer** — show how stacked layers expand visible input region;
- **RNN unroller** — same recurrent cell reused across timesteps;
- **vanishing-gradient chain** — visualize repeated multiplication shrinking/growing sensitivity;
- **gated-state explorer** — change gate values and observe retained/updated state;
- **seq2seq bottleneck visualization** — entire source compressed into one vector versus access to per-token encoder states as a preview.

## Worksheets

Useful worksheet topics:

- tiny convolution calculations;
- convolution output shapes and parameter counts;
- trace a small RNN hidden state across timesteps;
- reason about repeated gradient multiplication;
- simplified LSTM/GRU gate calculations;
- compare dense, convolutional, and recurrent parameter-sharing structures;
- identify the encoder-decoder bottleneck in a concrete example.

## Embedded exercises

Candidates:

- implement a tiny 1D convolution directly;
- calculate convolution output dimensions;
- implement one recurrent-cell forward step;
- unroll a provided simple RNN over a sequence;
- inspect how clipping affects an exploding numerical recurrence;
- implement a simplified gated-state update.

Do not require a hand-written full CNN or LSTM library as a core checkpoint.

## Jupyter/repository experiments

Two focused experiments are appropriate:

**CNN experiment:** train a small CNN on a modest image dataset and compare it against a dense baseline while inspecting parameter count and generalization.

**Sequence experiment:** train a small recurrent model on a deliberately simple sequence task, inspect performance as dependency length changes, and compare a simple RNN with a gated recurrent model where practical.

The experiments should demonstrate architectural behavior, not chase benchmark performance.

## Mastery expectations

After Module 14 the learner should be able to:

- explain why architecture encodes assumptions about data structure;
- calculate and reason about convolutional operations;
- train and analyze a small CNN;
- explain recurrent hidden state and parameter sharing through time;
- explain vanishing/exploding gradients in recurrent chains;
- explain why LSTM/GRU gating helps;
- explain encoder-decoder sequence modeling;
- articulate precisely why a single fixed context vector is a bottleneck;
- understand why attention was a compelling next architectural step.

## Deliberately deferred

- object detection and segmentation;
- advanced CNN families and architecture search;
- vision transformers;
- signal-processing treatment of convolution/Fourier analysis;
- exhaustive RNN-cell implementation details;
- specialist sequence-modeling practice;
- attention itself;
- Transformers;
- modern state-space sequence models.

---

# Tranche-wide mathematics progression

| Module | Mathematical emphasis |
| --- | --- |
| 08 | distance/scaling, impurity, ensemble averaging, bias/variance revisited |
| 09 | projection, orthogonality, covariance geometry, eigenvectors/eigenvalues for PCA |
| 10 | affine matrix functions, nonlinear composition, tensor/shape reasoning |
| 11 | chain rule II, reverse-mode differentiation, gradient accumulation |
| 12 | synthesis of forward/backward/optimization mathematics |
| 13 | stochastic optimization, momentum/Adam/AdamW intuition, learning-rate scheduling, initialization/normalization, numerical precision |
| 14 | convolution arithmetic, recurrent equations, repeated gradient behavior, gating |

This tranche should not accidentally create detached courses in linear algebra II, multivariable calculus, numerical analysis, or signal processing. Mathematical depth should remain tied to the ML mechanism that motivates it.

# Historical spine

The historical story across Modules 08–14 should feel coherent:

```text
multiple classical modeling traditions
        ↓
ensembles and instance-based learning
        ↓
unsupervised structure + multivariate projection
        ↓
symbolic AI and early connectionist traditions
        ↓
early artificial neurons / perceptrons
        ↓
expert systems, AI winters, and changing expectations
        ↓
multilayer differentiable networks
        ↓
efficient reverse differentiation / backpropagation
        ↓
practical deep-learning frameworks
        ↓
architectures that exploit spatial and sequential structure
        ↓
RNN memory/optimization limitations
        ↓
seq2seq fixed-context bottleneck
        ↓
attention becomes motivated
```

History should be used to explain why technical ideas emerged, not as a parallel memorization track.

# Foundational implementation checkpoints

By the end of Module 14, the learner should personally have implemented or substantially reasoned through:

- a small k-nearest-neighbor predictor;
- k-means;
- dense neural-network forward passes;
- backward rules for basic computational-graph operations;
- a small neural network's backpropagation;
- a complete small neural network from scratch;
- a PyTorch version of a previously understood model;
- a tiny convolution operation;
- a simple recurrent-cell forward/unrolled computation.

A full tree trainer, random-forest implementation, LSTM library, or CNN framework is not required. The hand-written-first rule should target mechanisms whose implementation materially improves understanding.

# Mastery checkpoints

This tranche should contain three increasingly substantial checkpoints:

1. **Classical ML synthesis (end of Module 09):** choose and compare appropriate classical techniques under a defensible experimental protocol.
2. **Neural Network From Scratch (Module 12):** major learner-written repository project demonstrating forward/backward/training mechanics and debugging discipline.
3. **Framework deep-learning experiment (Modules 13–14):** use PyTorch to run a controlled architecture/training experiment while still explaining the underlying mechanics.

The third checkpoint should not become a large application project. Its job is to demonstrate that framework fluency has not replaced conceptual understanding.

# Handoff to the next tranche

Module 14 intentionally stops at the recurrent encoder-decoder bottleneck. The next detailed syllabus tranche should then proceed through:

- Module 15 — Embeddings and Language Modeling;
- Module 16 — Attention;
- Module 17 — Transformer From Scratch;
- Module 18 — Tiny Autoregressive Language Model.

That sequence should preserve the same pedagogical principle used throughout the course: **the learner should encounter the problem before being given the architectural solution.**