# AI/ML detailed syllabus plan: Modules 01–07

This document expands Modules 01 through 07 from the canonical [`ai-ml-syllabus.md`](ai-ml-syllabus.md) into a lesson-level teaching plan.

It is a **planning document**, not runtime-authored curriculum. Exact lesson files, objective metadata, worksheets, interactives, notebook repositories, sources, and exercise definitions are created later under `curriculum/courses/ai-ml/`.

The purpose of this tranche is to establish the mathematical and experimental foundation required before the course reaches broader classical ML and neural networks. The intended learner is already an experienced software engineer, so programming instruction should concentrate on scientific Python, numerical reasoning, and implementing ML ideas rather than introductory programming.

## Scope guardrails

Modules 01–07 should be deep where later ML depends on the material, but they should not become miniature university courses in Python, linear algebra, calculus, probability, or statistics.

The design rule is:

> Teach the minimum complete mathematical foundation needed to understand and use the next important ML idea correctly, practice it enough to make it reliable, and revisit it later when a deeper topic demands more sophistication.

This tranche should therefore:

- preserve the existing bridge from programming functions to mathematical functions;
- develop concrete linear-algebra intuition before matrix-heavy ML notation appears;
- introduce the first trainable model before calculus so derivatives solve a problem the learner already understands;
- introduce probability before probabilistic classification;
- give exponentials and logarithms a real just-in-time treatment when logistic regression needs them;
- teach evaluation and experimental discipline before adding more model families;
- use worksheets heavily for mathematical operations and written reasoning;
- use embedded Pyodide exercises for constrained implementation practice;
- use Jupyter notebooks for exploration, visualization, and experiment habits;
- reserve repository projects for synthesis and transfer.

The tranche should **not** add decision trees, random forests, k-means, PCA, SVMs, neural networks, or other named model families merely for breadth. Those have later homes in the core syllabus.

## Tranche overview

| Module | Planned lessons | Primary job |
| --- | ---: | --- |
| 01 — Scientific Python Foundations | 6 | Become comfortable conducting numerical experiments |
| 02 — Vectors, Matrices, and Numerical Data | 5 | Build linear algebra I |
| 03 — Linear Regression: Your First ML Model | 5 | Understand what a trainable model is |
| 04 — Derivatives and Gradient Descent | 6 | Build calculus I and optimization intuition |
| 05 — Probability and Statistical Thinking for ML | 6 | Build probability/statistics I |
| 06 — Classification and Logistic Regression | 5 | Build the first probabilistic classifier |
| 07 — Generalization, Evaluation, and Good Experiments | 6 | Learn how to determine whether ML actually worked |

This is approximately 39 lessons total, including the two Module 01 lessons that already exist. Lesson count is a planning estimate, not a product requirement; concepts should be split or combined when authoring improves the learning experience.

## Dependency shape

Recommended teaching order remains Modules 01 through 07, but runtime objective metadata should encode only real prerequisites.

```text
01 Scientific Python
        |
        v
02 Linear Algebra I
        |
        v
03 Linear Regression
      /   \
     v     v
04 Calculus   05 Probability
     \       /
      \     /
       v   v
06 Logistic Regression
        |
        v
07 Generalization & Evaluation
```

Important consequences:

- Module 05 does not need Module 04 as a mathematical prerequisite merely because it appears later in the recommended sequence.
- Module 06 genuinely draws on both gradient-based optimization and probability.
- Module 07 depends on having trained and evaluated actual regression/classification models, but many of its individual concepts may have narrower prerequisite edges.
- Later objective definitions should avoid false edges such as making every Module 02 objective a prerequisite for every Module 03 objective.

## Learning-medium roles

The existing platform boundaries should remain explicit throughout this plan:

```text
lesson / explanation
    -> understand the idea

interactive visualization
    -> build intuition

worksheet
    -> practice calculations and written reasoning

embedded Python exercise
    -> implement or apply a constrained idea

Jupyter notebook
    -> explore behavior and conduct an experiment

Git repository + IDE
    -> synthesize and transfer the ideas in substantial work
```

A module does not need every medium. The medium should match the learning job.

---

# 01 — Scientific Python Foundations

## Purpose

Make Python, NumPy, and Jupyter comfortable enough that the language and tooling stop obstructing numerical reasoning, while establishing the function vocabulary used throughout the mathematical and ML curriculum.

This module already exists as `curriculum/courses/ai-ml/modules/01-scientific-python/`. Its first two lessons and current worksheet set should be preserved rather than redesigned.

## Existing objectives

The current module defines:

- `python.execution-model` — read and reason about ordinary scientific Python;
- `python.functions` — express transformations as Python functions;
- `numpy.arrays` — represent numerical data with NumPy arrays;
- `numpy.vectorization` — reason about vectorized array operations;
- `numpy.broadcasting` — predict NumPy broadcasting behavior;
- `python.numerical-experiments` — use Python for small numerical experiments.

These remain appropriate.

A likely objective gap is the Jupyter/reproducibility capability. During authoring, consider adding a durable objective such as:

- conduct a reproducible notebook-based numerical experiment.

Plotting may either support that objective or receive a separate objective if later evidence/mastery requirements justify one. Do not create objectives simply to mirror every lesson heading.

## Prerequisites

- Module 00 orientation.
- Professional programming experience is assumed by the course learner profile.
- No calculus, linear algebra, probability, or ML knowledge is required.

## Lesson sequence

### 01 — Python for a Programmer

**Status:** already authored.

**Job:** Translate existing programming knowledge into Python's particular syntax, object model, mutation/reference behavior, iteration style, typing conventions, and idioms without reteaching basic programming.

Keep its experienced-programmer framing. The lesson's names-refer-to-objects model should be treated as groundwork for later NumPy views/storage and tensor semantics.

### 02 — Functions: Code and Mathematics

**Status:** already authored.

**Job:** Bridge programming functions to mathematical functions and establish mapping, domain, codomain, image, injectivity, surjectivity, bijectivity, inverses, composition, parameters, and transformation-pipeline thinking.

The existing ML foreshadowing should remain. In particular, the current worksheets already use affine functions, ReLU, nested transformations, and `L(f_theta(x), y)` to establish structural familiarity without prematurely teaching neural networks.

### 03 — NumPy Arrays: Shape, Dimension, and Dtype

**Job:** Establish the core data structure used for numerical computing.

Topics:

- Python scalar/list intuition versus NumPy arrays;
- array creation;
- shape and dimensionality;
- axes;
- dtype;
- indexing and slicing;
- mutation;
- views versus copies at the level needed to avoid common surprises;
- basic storage intuition without a memory-layout detour.

By the end, the learner should be able to inspect an unfamiliar array and explain what its shape and dtype mean, retrieve or modify intended elements, and predict whether common operations change a view or independent copy where the lesson has explicitly covered that behavior.

### 04 — Vectorized Thinking

**Job:** Shift from element-by-element programming to array-level numerical operations.

Topics:

- element-wise operations;
- universal functions;
- reductions;
- axis-aware reductions at an introductory level;
- translating ordinary loops into array expressions;
- why vectorized operations are idiomatic and computationally important;
- recognizing when vectorization changes the shape of the reasoning, not merely the syntax.

Do not turn this into a performance-engineering lesson. The goal is the mental model required for later NumPy/PyTorch code.

### 05 — Broadcasting and Shape Reasoning

**Job:** Make shape compatibility predictable rather than magical.

Topics:

- broadcasting rules;
- aligning dimensions from the trailing axes;
- singleton dimensions;
- predicting result shapes;
- distinguishing broadcasting from matrix multiplication;
- common broadcasting bugs;
- using explicit reshaping when intent would otherwise be unclear.

Shape reasoning should be treated as a durable ML skill because later tensor code will rely on it constantly.

### 06 — Jupyter and Numerical Experiments

**Job:** Introduce the primary environment for exploratory work and establish reproducible experiment habits early.

Topics:

- code cells and Markdown cells;
- mathematical notation in Markdown;
- kernels and persistent state;
- out-of-order execution hazards;
- restart-and-run-all discipline;
- plotting with a mainstream Python plotting library;
- generating and inspecting small datasets/arrays;
- random seeds and deterministic reruns where practical;
- package/environment workflow at the minimum useful level;
- basic floating-point awareness: computer numbers are finite approximations, not mathematical real numbers;
- when exploratory code should graduate from a notebook into normal modules/repositories.

The lesson should end in an actual small experiment rather than a notebook UI tour.

## Mathematics

Module 01 introduces or reinforces:

- functions as mappings;
- domain/codomain/image;
- composition;
- inverses;
- parameterized functions;
- reading basic notation;
- shape/dimension language as preparation for linear algebra.

It should not introduce calculus or formal linear algebra.

## History

History should be light and connected to the technical material:

- development of the mathematical function concept where useful in Lesson 02;
- why Python became prominent in scientific computing;
- NumPy and array-oriented scientific Python;
- notebooks as part of a longer interactive-computing tradition.

Do not create a separate history lesson.

## Interactives

Existing Lesson 02 function interactives should remain.

New candidates:

- array shape/axis explorer;
- indexing/slicing visualizer;
- broadcasting shape predictor.

Only build interactives that add spatial or dynamic intuition beyond prose and code.

## Worksheets

Existing worksheets should remain:

- Python execution model;
- function mappings and properties;
- inverses and composition;
- functions in code and machine learning.

New likely worksheets:

- **Array shapes and indexing** — read shapes, identify axes, predict indexing results;
- **Broadcasting and shape reasoning** — determine compatibility and resulting shapes before running code.

Vectorization is likely better assessed in code than on paper unless a specific tracing problem benefits from a worksheet.

## Embedded exercises

Candidates:

- create and reshape arrays to satisfy a target shape;
- index/select requested values;
- replace a loop with a vectorized expression;
- perform reductions over specified axes;
- predict and then use broadcasting;
- debug an incorrect shape/broadcast assumption.

## Jupyter lab

**First numerical experiment**

The learner should:

1. generate or load a small numeric dataset;
2. inspect shape/dtype/basic summary values;
3. compute one or more transformations;
4. plot the result;
5. change a meaningful parameter;
6. rerun from a clean kernel;
7. record an observation in Markdown.

The scientific question should be deliberately simple. The learning objective is experimental workflow.

## Mastery expectations

After Module 01 the learner should be able to:

- read ordinary scientific Python without translating every line back into another language;
- reason about Python names, objects, mutation, and function behavior;
- connect code functions to mathematical mappings and composition;
- manipulate NumPy arrays confidently;
- predict common vectorization/broadcasting behavior;
- use a notebook as a reproducible exploratory document rather than an opaque scratchpad.

## Deliberately deferred

- Python web/backend development;
- advanced Python OOP/application architecture;
- async programming;
- metaprogramming and decorators as standalone topics;
- advanced packaging/tooling;
- Pandas as a major curriculum subject;
- performance engineering;
- calculus, probability, and formal linear algebra.

---

# 02 — Vectors, Matrices, and Numerical Data

## Purpose

Build the first layer of linear algebra by connecting mathematical objects directly to NumPy arrays, data, geometry, and transformations.

This is not a standalone linear algebra course. Every concept should earn its place by supporting later ML reasoning.

## Prerequisites

Required capabilities from Module 01:

- mathematical functions and composition;
- NumPy arrays, shapes, dimensions, indexing, and vectorized arithmetic;
- basic notebook/plotting workflow.

No calculus is required.

## Proposed core objectives

Objective IDs are provisional until authored. Capabilities should cover:

- represent numerical observations as vectors and datasets/transformations as matrices;
- perform and interpret basic vector operations;
- compute and interpret dot products;
- perform matrix-vector and matrix-matrix multiplication and reason about compatible dimensions;
- interpret matrices as linear transformations and distinguish linear from affine transformations;
- translate between mathematical notation and NumPy expressions.

## Lesson sequence

### 01 — Vectors: Numbers With Structure

**Job:** Make vectors concrete before formal notation becomes dense.

Topics:

- scalar versus vector;
- components/coordinates;
- dimension;
- vectors as points, directions, measurements, feature collections, and model inputs;
- column-vector notation versus one-dimensional NumPy arrays;
- vector equality and basic notation.

Emphasize that a vector is not merely a Python list with a fancy name; its mathematical structure determines which operations mean something.

### 02 — Vector Arithmetic, Distance, and Dot Products

**Job:** Establish the vector operations that recur throughout ML.

Topics:

- vector addition/subtraction;
- scalar multiplication;
- magnitude/norm at introductory depth;
- distance;
- dot product mechanically and geometrically;
- angle/cosine intuition using the learner's existing basic trigonometry;
- similarity intuition without prematurely teaching embeddings.

The learner should calculate these by hand on small examples before relying on NumPy.

### 03 — Matrices: Data and Transformations

**Job:** Give matrices two useful mental models: rectangular data and transformations.

Topics:

- rows, columns, and matrix dimensions;
- indexing;
- transpose;
- dataset convention examples and the need to state orientation explicitly;
- matrices as collections of numbers;
- matrices as functions that transform vectors;
- reading matrix notation alongside NumPy shapes.

### 04 — Matrix Multiplication

**Job:** Make matrix multiplication understandable as composition rather than a memorized algorithm.

Topics:

- matrix-vector multiplication;
- matrix-matrix multiplication;
- row/column mechanics;
- dimensional compatibility;
- result shapes;
- matrix multiplication versus element-wise multiplication;
- composition of transformations;
- non-commutativity through concrete examples.

This lesson deserves substantial hand practice.

### 05 — Linear and Affine Transformations

**Job:** Connect linear algebra directly to the form used by ML models.

Topics:

- scaling;
- reflection;
- rotation in 2D;
- simple shear where useful;
- origin-preserving property of linear transformations;
- why translation is not linear;
- affine transformation as linear transformation plus translation;
- `Wx + b` as the canonical bridge to later regression and neural-network layers.

This lesson should explicitly connect back to Lesson 01.02's scalar affine function and expand it to vectors/matrices.

## Mathematics

New mathematical layer:

- scalars, vectors, matrices;
- vector arithmetic;
- Euclidean norm/distance;
- dot product;
- matrix multiplication;
- transpose;
- linear and affine transformations;
- elementary geometric interpretation.

Formal proof language should remain light. Understanding and calculation come first.

## History

Use short historical context around:

- coordinate geometry;
- development of vector concepts;
- matrix notation and linear transformations.

History should support the idea that modern ML notation grew from mathematical tools developed for geometry, systems, and transformations long before machine learning.

## Interactives

Strong candidates:

- **Vector addition explorer** — move component vectors and see the resultant;
- **Dot product geometry** — change angle/magnitudes and observe dot-product sign/magnitude;
- **2D matrix transformation lab** — apply a matrix to a grid/shape and watch scaling/rotation/reflection/shear;
- **Matrix multiplication composition** — apply transformations A then B and compare with the product.

## Worksheets

This module should be worksheet-heavy:

- vector components and arithmetic;
- magnitude and distance;
- dot products;
- matrix shapes and transpose;
- matrix-vector multiplication;
- matrix-matrix multiplication;
- dimension compatibility;
- linear versus affine transformation reasoning.

Problems should require work, not only final answers.

## Embedded exercises

Candidates:

- manually implement vector addition and dot product once without `np.dot`;
- manually implement small matrix-vector multiplication;
- manually implement matrix multiplication once;
- compare manual implementations to NumPy;
- write shape checks that reject invalid multiplication;
- apply an affine transformation to a batch of vectors.

Manual implementation is pedagogical; after the mechanics are understood, normal NumPy operations should be preferred.

## Jupyter lab

**Visualizing transformations**

The learner creates a small set of 2D points or a grid and applies several matrices, plotting before/after coordinates. The notebook should ask the learner to predict a transformation before executing it, then explain discrepancies.

## Mastery expectations

After Module 02 the learner should be able to:

- look at a vector/matrix expression and identify dimensions;
- perform small vector/matrix calculations by hand;
- explain the dot product mechanically and geometrically;
- predict whether matrix multiplication is dimensionally valid and what shape it produces;
- explain matrix multiplication as composition;
- recognize `Wx + b` as an affine transformation rather than mysterious ML notation.

## Deliberately deferred

- abstract vector spaces and proof-heavy axioms;
- basis/change of basis in depth;
- determinants as a major topic;
- rank/null space in depth;
- eigenvalues/eigenvectors;
- matrix decompositions;
- SVD;
- abstract linear algebra proofs.

Eigen concepts and projections return later when PCA gives them a concrete purpose.

---

# 03 — Linear Regression: Your First ML Model

## Purpose

Introduce machine learning through the simplest useful trainable model and establish the recurring structure of data, parameters, predictions, error, loss, and fitting.

This module should answer the foundational question:

> What does it mean for a machine to learn a model from examples rather than for a programmer to write the rule directly?

## Prerequisites

From Module 01:

- functions and parameterized functions;
- NumPy numerical experimentation.

From Module 02:

- vectors/matrices at introductory depth;
- affine transformations;
- dot products/matrix operations sufficient to read vectorized model notation.

Calculus is intentionally **not** a prerequisite.

## Proposed core objectives

Capabilities should cover:

- distinguish data, features, targets, parameters, predictions, and training procedures;
- compute predictions for a linear/affine regression model;
- compute residuals and mean squared error;
- explain a loss function as an objective over model parameters;
- distinguish a model from the algorithm used to fit it;
- conduct a small parameter-search training experiment and interpret the result.

## Lesson sequence

### 01 — From Programs to Models

**Job:** Establish the machine-learning problem before presenting equations.

Topics:

- explicit programmed rules versus learned parameters;
- examples/observations;
- features/input variables;
- targets/labels;
- supervised learning;
- regression as predicting a quantity;
- training versus using a trained model;
- parameters versus ordinary program configuration.

Avoid a broad taxonomy of ML. Supervised versus unsupervised can be named, but the lesson's job is the concrete regression problem.

### 02 — A Model With Parameters

**Job:** Turn the familiar line equation into a trainable model.

Topics:

- `y_hat = wx + b`;
- input `x`;
- target `y`;
- prediction `y_hat`;
- weight/slope `w`;
- bias/intercept `b`;
- parameterized family of functions;
- how changing parameters changes predictions;
- extension from a single feature to vector notation only as far as useful.

Use notation carefully and repeatedly translate it into code.

### 03 — Measuring Error

**Job:** Explain why training needs a numerical definition of what counts as better.

Topics:

- residuals;
- signed error and cancellation problems;
- absolute error intuition;
- squared error intuition;
- mean squared error;
- indexed observations;
- summation notation `Sigma`;
- average over a dataset;
- loss/objective terminology.

Do not derive gradients yet.

### 04 — Loss and Fitting

**Job:** Reframe loss as a function of the parameters.

Topics:

- fixed dataset, variable parameters;
- `L(w,b)`;
- candidate parameter sets;
- one-dimensional and two-dimensional loss views;
- loss surfaces;
- fitting as choosing parameters that reduce loss;
- model versus optimizer/training algorithm;
- local visual intuition about a better/worse parameter choice.

### 05 — Your First Training Experiment

**Job:** Make a model improve from data without calculus so the need for optimization becomes obvious.

The learner personally implements:

- prediction;
- residual calculation;
- MSE;
- a simple grid/brute-force parameter search;
- selection of the best candidate parameters.

The experiment should visualize candidate lines and/or the loss surface.

The inefficiency of brute-force search is pedagogically useful. The module should end with the motivating question for Module 04:

> Can the shape of the loss function tell us which direction to move the parameters instead of trying many possibilities blindly?

## Mathematics

New mathematical layer:

- indexed observations such as `x_i` and `y_i`;
- summation notation;
- arithmetic mean;
- residuals;
- squared quantities;
- a loss/objective as a function of parameters;
- reading simple vectorized linear-model notation.

No derivatives yet.

## History

Integrate a concise least-squares story:

- fitting observations predates modern ML by centuries;
- Legendre/Gauss and least squares;
- modern supervised learning inherits ideas from statistics and numerical fitting rather than appearing from nowhere.

Avoid priority disputes or a full history of regression.

## Interactives

Strong candidates:

- draggable regression line with residuals shown visually;
- live MSE as `w` and `b` change;
- loss curve for one parameter;
- small 2D loss-surface explorer for `(w,b)`.

## Worksheets

Candidates:

- identify features/targets/parameters/predictions from scenarios;
- compute predictions and residuals;
- calculate MSE by hand for a tiny dataset;
- translate summation notation into explicit arithmetic;
- compare several parameter sets by loss;
- distinguish model definition from fitting algorithm.

## Embedded exercises

Foundational learner-written implementations:

- `predict(x, w, b)`;
- `mse(predictions, targets)`;
- vectorized residual/MSE calculation;
- evaluate candidate parameter grids;
- return the best candidate parameter set.

Do not use scikit-learn for the foundational implementation.

## Jupyter lab

**Explore a regression loss surface**

The learner should:

1. generate a small noisy linear dataset;
2. plot the observations;
3. evaluate many `(w,b)` candidates;
4. visualize loss across candidates;
5. inspect the best candidate;
6. compare predicted line and observations;
7. write a short interpretation of what fitting accomplished.

## Mastery expectations

After Module 03 the learner should be able to:

- explain training without anthropomorphic language;
- identify data, parameters, predictions, residuals, and loss in a simple model;
- calculate a tiny regression example by hand;
- explain why MSE defines a preference over parameter choices;
- distinguish the regression model from the procedure used to find its parameters;
- implement the model and loss without a machine-learning library.

## Deliberately deferred

- gradient descent;
- derivatives;
- normal-equation closed-form solutions;
- statistical inference for regression coefficients;
- hypothesis tests/confidence intervals;
- extensive multiple-regression theory;
- polynomial regression as a standalone topic;
- scikit-learn abstractions until the underlying model has been built manually.

---

# 04 — Derivatives and Gradient Descent

## Purpose

Introduce calculus because Module 03 has created a concrete optimization problem:

> If the loss is too high, how does changing a parameter change the loss, and which direction should the parameter move?

This module is the first major calculus foundation and should receive substantial practice. It should not become a full Calculus I course detached from ML.

## Prerequisites

From Module 01:

- functions and composition.

From Module 02:

- vectors and basic geometry sufficient to interpret a gradient vector.

From Module 03:

- parameterized models;
- loss as a function of parameters;
- linear regression and MSE.

## Proposed core objectives

Capabilities should cover:

- interpret a derivative as local rate of change/slope;
- compute derivatives of the common elementary forms needed in early ML;
- compute and interpret partial derivatives;
- interpret a gradient as a vector of sensitivities/directions of steepest increase;
- apply the chain rule to simple compositions;
- implement gradient descent and reason about learning-rate behavior;
- compare numerical and analytic derivatives as a correctness check.

## Lesson sequence

### 01 — From Slope to Instantaneous Change

**Job:** Begin with an existing geometric idea rather than derivative notation.

Topics:

- average rate of change;
- slope between two points;
- secant line;
- curved functions where one global slope is insufficient;
- tangent/local slope intuition;
- shrinking the interval;
- rate-of-change examples connected to model loss.

### 02 — The Derivative

**Job:** Name and formalize the local-change concept without turning limits into real analysis.

Topics:

- derivative at a point;
- derivative as another function;
- limit intuition to the depth needed to understand what is being approximated;
- common notations such as `f'(x)` and `df/dx`;
- positive/negative/zero derivative interpretation;
- local linear approximation intuition where useful.

### 03 — Calculating Derivatives

**Job:** Give the learner enough procedural fluency to differentiate early ML expressions.

Topics:

- derivative of a constant;
- power rule;
- constant multiple rule;
- sum/difference rule;
- simple polynomials;
- derivatives of basic exponentials/logarithms may be previewed only if useful, with their deeper role revisited in Module 06;
- careful algebra before and after differentiation.

Product/quotient rules should be included only if required by representative problems. They should not expand the module merely because they are traditional Calculus I checklist items.

### 04 — Partial Derivatives and Gradients

**Job:** Extend derivative intuition to models with multiple parameters.

Topics:

- functions of several variables;
- partial derivative: vary one input while holding others fixed;
- notation;
- gradient vector;
- gradient components and shapes;
- geometric interpretation of direction of steepest increase;
- negative gradient as a natural descent direction;
- apply to a small `L(w,b)` example.

### 05 — The Chain Rule

**Job:** Explain how change propagates through composed functions.

Topics:

- revisit function composition from Module 01;
- outer and inner functions;
- derivative of a composition;
- simple chain-rule calculations;
- dependency diagrams/computational-chain intuition;
- why this will matter dramatically for neural networks later.

Keep the depth sufficient for early gradient calculations. Backpropagation will revisit the chain rule with computational graphs and many parameters.

### 06 — Gradient Descent

**Job:** Use all previous concepts to build the first efficient training algorithm.

Topics:

- current parameters;
- calculate gradient;
- move opposite the gradient;
- learning rate;
- repeated updates;
- convergence intuition;
- too-small learning rate;
- too-large learning rate/divergence;
- initialization at a basic level;
- analytic versus numerical derivatives;
- finite-difference gradient checking;
- train the Module 03 linear-regression model with gradient descent.

## Mathematics

New mathematical layer:

- average and instantaneous rates of change;
- derivative;
- limit intuition;
- elementary derivative rules;
- partial derivatives;
- gradient vectors;
- chain rule;
- iterative numerical optimization.

The course should emphasize meaning and reliable calculation over proof formalism.

## History

Short connected context:

- Newton and Leibniz and the development of calculus;
- derivatives as a general tool for change, not an ML invention;
- later steepest-descent/gradient methods in numerical optimization;
- modern model training as an application of a much older mathematical idea.

## Interactives

High-value candidates:

- secant line approaching a tangent line;
- draggable point showing local derivative sign/magnitude;
- two-parameter loss surface with gradient arrow;
- step-by-step gradient-descent path;
- learning-rate comparison showing convergence, slow progress, oscillation, and divergence.

## Worksheets

This should be one of the most worksheet-intensive modules.

Candidates:

- slope and average-rate calculations;
- derivative interpretation from graphs;
- derivative-rule practice;
- partial derivatives;
- gradient construction;
- introductory chain rule;
- hand-computed gradient-descent steps;
- diagnose sign/learning-rate mistakes.

## Embedded exercises

Foundational learner-written implementations:

- numerical derivative with finite differences;
- analytic derivatives for selected functions;
- compare analytic and numerical results within tolerance;
- compute a gradient for a tiny loss;
- implement one gradient-descent update;
- implement the full iterative linear-regression optimizer.

Tests should use numerical tolerances and may verify that loss decreases for sensible inputs.

## Jupyter lab

**How learning rate changes optimization**

The learner trains the same regression model from several starting points and learning rates, plotting:

- loss over iterations;
- parameter trajectory where practical;
- final fitted line.

The notebook should intentionally include at least one learning rate that is too small and one that is unstable/too large.

## Mastery expectations

After Module 04 the learner should be able to:

- explain what a derivative means geometrically and operationally;
- calculate derivatives of simple expressions reliably;
- compute partial derivatives and assemble a gradient;
- explain the chain rule as change through composition;
- derive or closely work through the gradient for a simple regression loss;
- implement gradient descent without a library optimizer;
- explain learning-rate failures rather than merely tuning until something works.

## Deliberately deferred

- integration as a broad calculus subject;
- epsilon-delta proofs;
- rigorous real analysis;
- Hessians and second derivatives as a major topic;
- Newton/quasi-Newton methods;
- formal convex analysis;
- matrix calculus;
- advanced optimization theory.

---

# 05 — Probability and Statistical Thinking for ML

## Purpose

Build the probability foundation needed to reason correctly about uncertain observations, sampling, and probabilistic predictions.

Probability is potentially enormous. This module should teach the concepts that recur in ML rather than survey every classical probability topic.

## Prerequisites

Required:

- algebra/functions from Module 01;
- basic numerical experimentation from Module 01.

Helpful but not mathematically required:

- experience with noisy data from Module 03.

Module 04 calculus is **not** a hard prerequisite for the core probability concepts in this module.

## Proposed core objectives

Capabilities should cover:

- reason about events and basic probability rules;
- interpret random variables and probability distributions;
- compute and interpret expectation, variance, and standard deviation;
- distinguish covariance/correlation from independence and causation;
- compute conditional probabilities and apply Bayes' rule;
- distinguish a population from a sample and use simulation to reason about empirical estimates;
- explain sampling variability and uncertainty without pretending a finite sample is the population.

## Lesson sequence

### 01 — Probability and Uncertainty

**Job:** Establish probability as a mathematical language for uncertain outcomes.

Topics:

- experiments/outcomes;
- sample spaces;
- events;
- probability values;
- complements;
- mutually exclusive events where useful;
- unions/intersections at practical depth;
- simple addition/multiplication reasoning;
- empirical frequency versus theoretical probability.

Avoid a large combinatorics detour.

### 02 — Random Variables and Distributions

**Job:** Move from events to numerical quantities whose values are uncertain.

Topics:

- random variable;
- discrete versus continuous;
- probability mass/density intuition;
- distribution as a description of possible values and their relative likelihood;
- Bernoulli distribution;
- categorical outcomes;
- normal distribution as an important recurring model;
- distribution parameters at an introductory level.

Do not teach a catalog of named distributions.

### 03 — Expectation and Variability

**Job:** Separate typical value from spread.

Topics:

- expected value;
- mean;
- weighted averages;
- variance;
- standard deviation;
- intuitive effect of outliers;
- why two distributions can share a mean and behave very differently;
- expectation as a recurring ML quantity.

### 04 — Variables Together

**Job:** Introduce relationships between uncertain variables without conflating association and causation.

Topics:

- joint behavior;
- covariance intuition and sign;
- correlation and scale normalization;
- independence;
- independence versus zero correlation at conceptual depth;
- correlation does not imply causation;
- visual scatterplot reasoning.

### 05 — Conditional Probability and Bayes' Rule

**Job:** Make conditioning and base rates intuitive enough for later probabilistic modeling.

Topics:

- `P(A | B)`;
- conditioning changes the reference population;
- joint and conditional probability relationship;
- Bayes' rule;
- prior/base rate, likelihood/evidence terminology only where it clarifies the calculation;
- diagnostic/test examples;
- why `P(A | B)` and `P(B | A)` are not interchangeable.

### 06 — Sampling and Learning From Data

**Job:** Connect probability to what ML actually observes: finite samples from a larger process/population.

Topics:

- population versus sample;
- sample statistic versus population quantity;
- empirical estimates;
- repeated sampling;
- sampling variability;
- sample mean behavior through simulation;
- law-of-large-numbers intuition without proof;
- random seeds and reproducibility;
- Monte Carlo/simulation as a way to reason about probability.

This lesson should naturally prepare the learner for evaluation/generalization without attempting formal mathematical statistics.

## Mathematics

New mathematical layer:

- events and probability rules;
- random variables;
- probability distributions;
- expected value;
- variance and standard deviation;
- covariance and correlation;
- conditional probability;
- Bayes' rule;
- samples and empirical estimates.

## History

Short connected context:

- Pascal/Fermat and the early mathematical treatment of chance;
- Bayes and Laplace;
- probability/statistics becoming tools for learning about uncertain processes from observations.

Do not create a chronological survey of statistics.

## Interactives

Strong candidates:

- repeated Bernoulli/coin-trial simulator showing empirical frequency;
- adjustable distribution showing mean/spread;
- sample-size explorer showing variability of estimates;
- conditional-probability/base-rate visualization;
- covariance/correlation scatterplot explorer.

## Worksheets

Candidates:

- event/complement probabilities;
- simple joint probabilities;
- expectation;
- variance/standard deviation on tiny distributions;
- covariance/correlation interpretation;
- conditional probability;
- Bayes' rule with explicit base rates;
- population/sample identification.

## Embedded exercises

Candidates:

- simulate Bernoulli trials;
- calculate empirical probability;
- compute expected value for a discrete distribution;
- compute mean/variance from a small sample manually in code;
- implement a simple conditional-probability calculation;
- simulate repeated sample means.

## Jupyter lab

**Probability through simulation**

The learner runs repeated simulated experiments and investigates how empirical estimates behave as sample size changes.

Possible questions:

- How quickly does a coin's empirical success rate stabilize?
- How variable are sample means at different sample sizes?
- Can two distributions with the same mean have visibly different behavior?

The notebook should require predictions before simulation and written interpretation afterward.

## Mastery expectations

After Module 05 the learner should be able to:

- interpret common probability notation;
- distinguish outcome, event, random variable, distribution, population, and sample;
- compute expectation and variability for small examples;
- reason about conditional probability and Bayes' rule;
- explain why finite samples vary;
- use simulation to check probabilistic intuition;
- recognize that model probabilities and sampled observations require different kinds of reasoning than deterministic outputs.

## Deliberately deferred

- measure-theoretic probability;
- combinatorics as a broad unit;
- exhaustive named-distribution catalogs;
- moment-generating functions;
- stochastic processes;
- formal estimator theory;
- confidence intervals and hypothesis-testing machinery in depth;
- Bayesian computation;
- MCMC.

---

# 06 — Classification and Logistic Regression

## Purpose

Extend the first-model story from predicting continuous quantities to predicting classes and probabilities.

This is the first major convergence point of the curriculum:

```text
functions
+ vectors
+ loss
+ derivatives
+ gradient descent
+ probability
        |
        v
logistic regression
```

## Prerequisites

From Module 02:

- vector/dot-product and affine-transformation intuition.

From Module 04:

- derivatives, gradients, chain rule, gradient descent.

From Module 05:

- probability, expectation basics, conditional-probability intuition.

From Module 03:

- model/parameter/loss/training vocabulary.

## Proposed core objectives

Capabilities should cover:

- distinguish regression and classification outputs;
- explain why an unbounded linear score is not directly a probability;
- reason with exponentials, logarithms, odds, log-odds, logits, and sigmoid at the depth needed for logistic regression;
- interpret likelihood as a function of model parameters;
- compute and interpret binary cross-entropy/log loss;
- implement and train binary logistic regression manually;
- interpret decision boundaries and distinguish model probability from decision threshold.

## Lesson sequence

### 01 — From Regression to Classification

**Job:** Define the classification problem and motivate a probabilistic output.

Topics:

- binary classes/labels;
- classification versus regression;
- linear score;
- why ordinary linear regression can produce invalid probabilities;
- score versus probability versus final class decision;
- decision boundaries;
- vector form of a linear classifier at introductory depth.

### 02 — Odds, Log-Odds, and the Sigmoid

**Job:** Give exponentials/logarithms the just-in-time treatment needed to understand logistic regression rather than treating sigmoid as a magical formula.

Topics:

- exponential functions;
- logarithms as inverse functions;
- key log identities, especially products becoming sums;
- probability and odds;
- log-odds;
- logits;
- sigmoid/logistic function;
- sigmoid range and shape;
- converting an unbounded score into a `(0,1)` value.

The goal is conceptual/operational fluency, not an exhaustive precalculus review.

### 03 — Likelihood

**Job:** Introduce the idea that a probabilistic model can be judged by the probability it assigns to the observations that actually occurred.

Topics:

- model parameters determine predicted probabilities;
- probability of observed outcomes;
- likelihood as a function of parameters with data fixed;
- product of probabilities across independent observations as a motivating simplification;
- maximum-likelihood intuition;
- distinction between probability and likelihood in context.

Avoid turning this into formal statistical inference.

### 04 — Log Loss and Cross-Entropy

**Job:** Connect likelihood to the loss actually optimized for binary classification.

Topics:

- logarithms convert products into sums;
- negative log-likelihood intuition;
- binary cross-entropy/log loss;
- confident correct versus confident wrong predictions;
- shape of loss as predicted probability changes;
- why MSE is not the natural default here;
- cross-entropy terminology without launching a detached information-theory unit.

### 05 — Logistic Regression From Scratch

**Job:** Build and train the complete classifier with transparent mechanics before relying on scikit-learn.

The learner personally implements:

- linear score/logit;
- sigmoid;
- predicted probabilities;
- binary cross-entropy;
- gradient calculation or closely guided derivation using existing calculus;
- gradient-descent updates;
- vectorized training loop;
- thresholded predictions;
- decision-boundary visualization.

After the manual implementation works, fit the same problem with scikit-learn and compare behavior/API abstractions.

## Mathematics

New/revisited mathematical layer:

- exponentials;
- logarithms and inverse relationship;
- key log identities;
- odds and log-odds;
- sigmoid;
- likelihood;
- negative log-likelihood;
- binary cross-entropy;
- gradient-based optimization revisited.

Information-theoretic language may be previewed, but entropy/KL should wait until later modules have a real need.

## History

Useful context:

- logistic models/statistical classification predate modern deep learning;
- the shift from deterministic scores to probabilistic statistical models;
- statistical machine learning as part of the broader historical path toward modern AI.

Avoid a catalog of classifier history.

## Interactives

Strong candidates:

- exponential/log inverse explorer if existing generic math visualization is insufficient;
- sigmoid explorer showing score-to-probability mapping;
- probability/odds/log-odds converter;
- binary log-loss curve;
- 2D logistic decision-boundary explorer with adjustable parameters/threshold.

## Worksheets

Candidates:

- exponentials/logarithms refresher tied directly to ML expressions;
- probability to odds/log-odds conversions;
- evaluate sigmoid for manageable values;
- likelihood comparison for tiny datasets;
- binary cross-entropy calculations;
- one hand gradient/update step;
- distinguish score, probability, threshold, and predicted class.

## Embedded exercises

Foundational learner-written implementations:

- `sigmoid` with basic numerical-care discussion;
- binary cross-entropy;
- predict probabilities;
- classify with a supplied threshold;
- one logistic-regression gradient step;
- full training loop on a tiny dataset.

Afterward, a separate exercise may reproduce the workflow using scikit-learn.

## Jupyter lab

**Binary classifier from scratch**

Use a synthetic two-dimensional dataset so the learner can visualize:

- observations/classes;
- learned decision boundary;
- predicted probability field where practical;
- training loss;
- effects of threshold changes.

Compare the manual implementation to scikit-learn and explain what the library hides.

## Mastery expectations

After Module 06 the learner should be able to:

- explain how logistic regression turns a linear score into a probability;
- work with exponentials/logarithms well enough to follow the model and loss;
- explain likelihood and log loss without treating them as arbitrary formulas;
- implement binary logistic regression and train it with gradient descent;
- interpret a decision boundary;
- distinguish predicted probability from the application decision threshold;
- use scikit-learn with an understanding of the underlying model.

## Deliberately deferred

- multiclass softmax until later neural/language-model needs make it useful;
- full information theory;
- KL divergence;
- advanced generalized-linear-model theory;
- Bayesian logistic regression;
- multiclass strategy catalogs;
- broad classifier surveys.

---

# 07 — Generalization, Evaluation, and Good Experiments

## Purpose

Teach how to determine whether a model learned something useful rather than merely fitting the examples it was shown.

The central conceptual shift is:

> Training loss going down is not the goal. The goal is useful performance on data the training procedure did not get to exploit.

This module should establish experimental habits before the course introduces more model families.

## Prerequisites

From Modules 03 and 06:

- experience training regression and classification models;
- loss functions and fitted parameters;
- probabilistic classifier outputs.

From Module 05:

- samples, randomness, empirical estimates, and uncertainty.

From Module 04:

- optimization knowledge is helpful for understanding training behavior but is not a prerequisite for every evaluation objective.

## Proposed core objectives

Capabilities should cover:

- distinguish training performance from generalization;
- design appropriate train/validation/test splits and explain each role;
- identify common data leakage and test-set misuse;
- diagnose overfitting/underfitting and explain bias/variance intuition;
- explain introductory regularization and its purpose;
- choose and interpret appropriate regression/classification metrics;
- reason about class imbalance, thresholds, and calibration;
- run a reproducible controlled ML experiment with baseline, fixed evaluation protocol, and honest conclusions.

## Lesson sequence

### 01 — Training Is Not the Goal

**Job:** Establish generalization as the real target of predictive ML.

Topics:

- fitting versus memorization;
- seen versus unseen examples;
- training error;
- generalization/test error;
- overfitting and underfitting at first intuition;
- why a model can achieve excellent training loss and still be useless.

### 02 — Train, Validation, and Test

**Job:** Make data splitting an experimental-design concept rather than library boilerplate.

Topics:

- training set role;
- validation set role;
- test set role;
- model/hyperparameter selection;
- repeated peeking at test results as information leakage;
- preprocessing leakage;
- cross-validation conceptually;
- stratification where useful;
- data splitting when observations are not independent, mentioned as an important caveat without expanding into time-series/grouped-data methodology.

### 03 — Overfitting, Underfitting, and Regularization

**Job:** Explain model complexity and constraints without launching statistical learning theory.

Topics:

- capacity/complexity intuition;
- underfit versus overfit patterns;
- training/validation curves;
- bias/variance intuition;
- regularization as a preference/penalty that constrains solutions;
- L2 regularization at enough mathematical depth to understand its effect on the objective;
- L1 mentioned for contrast without a full sparse-model unit.

### 04 — Metrics and Confusion Matrices

**Job:** Show that evaluation requires defining what kind of error matters.

Topics:

Regression:

- MAE;
- MSE/RMSE revisited;
- baseline comparison.

Classification:

- confusion matrix;
- true/false positives/negatives;
- accuracy;
- precision;
- recall;
- F1;
- metric choice as a consequence of problem costs, not leaderboard taste.

ROC/PR should be introduced only to the level needed for the next lesson.

### 05 — Thresholds, Imbalance, and Calibration

**Job:** Separate model scoring from application decisions and expose common metric traps.

Topics:

- decision thresholds;
- threshold movement and precision/recall tradeoff;
- class imbalance;
- majority-class baseline;
- why accuracy can mislead;
- ROC intuition;
- precision-recall intuition and why it matters under imbalance;
- calibration: what it means for a predicted `0.8` probability to behave like 80% over comparable cases;
- discrimination versus calibration.

### 06 — How to Run a Trustworthy Experiment

**Job:** Turn all prior material into a reusable experimental discipline.

Topics:

- clear question/hypothesis;
- define dataset and target;
- inspect data before modeling;
- choose split strategy before tuning;
- establish a simple baseline;
- fix the evaluation protocol;
- control randomness/seeds where practical;
- compare one change at a time when investigating causes;
- record configuration/results;
- error analysis;
- report uncertainty/limitations;
- avoid claims not supported by the experiment;
- reproducible notebook/repository habits.

This is the first explicit research-engineering habit module, but the habits should continue throughout the rest of the course.

## Mathematics

New/revisited mathematical layer:

- sampling/generalization intuition;
- regularization penalty terms;
- confusion-matrix arithmetic;
- precision/recall/F1;
- threshold-dependent metrics;
- calibration intuition;
- bias/variance as conceptual language rather than a formal decomposition proof.

## History

Rather than another chronology lesson, connect the development of statistical/ML practice to the shift from merely fitting observed data toward out-of-sample prediction, held-out evaluation, and controlled empirical comparison.

Historical context may mention the growth of benchmark-driven empirical ML, but modern benchmark pathologies belong later in LLM evaluation.

## Interactives

Strong candidates:

- underfit/appropriate/overfit model explorer;
- train-versus-validation curve explorer;
- confusion-matrix metric calculator;
- classification threshold slider with precision/recall updates;
- class-imbalance simulator;
- calibration/reliability diagram explorer.

## Worksheets

Candidates:

- identify leakage in experiment descriptions;
- choose train/validation/test roles;
- diagnose overfitting from curves;
- compute L2-regularized objective for tiny examples;
- confusion-matrix arithmetic;
- precision/recall/F1 calculations;
- choose a metric for a stated cost scenario;
- reason about threshold and class imbalance;
- critique unsupported experiment conclusions.

## Embedded exercises

Candidates:

- implement regression/classification metrics;
- construct a confusion matrix from labels/predictions;
- calculate precision/recall/F1;
- apply different thresholds to supplied probabilities;
- demonstrate a leakage bug and correct it;
- compare an unregularized and L2-regularized objective on supplied values.

## Jupyter/repository capstone

### Can You Trust This Model?

This should be the first meaningful **transfer checkpoint** for the course, but not a giant engineering project.

The learner receives a small tabular binary-classification problem and several intentionally flawed experimental approaches. The flaws may include:

- evaluating on training data;
- preprocessing before the split in a way that leaks information;
- selecting decisions based repeatedly on the test set;
- misleading accuracy under class imbalance;
- uncontrolled randomness;
- no baseline;
- drawing a stronger conclusion than the evidence supports.

The learner should diagnose the flawed approaches and then produce a clean experiment using logistic regression.

Suggested deliverables:

```text
experiment.ipynb
README.md   # or a short experiment report
```

The report should cover:

```text
question
  -> data
  -> split strategy
  -> baseline
  -> model
  -> metric choice
  -> results
  -> error analysis
  -> limitations
  -> conclusion
```

The primary assessment target is experimental reasoning, not model novelty or best possible score.

This capstone should produce transfer evidence across several objectives from Modules 03–07.

## Mastery expectations

After Module 07 the learner should be able to:

- explain why training performance alone is not evidence of useful learning;
- construct and defend a train/validation/test strategy;
- identify common leakage and test-set contamination;
- diagnose basic underfitting/overfitting;
- explain why regularization can improve generalization;
- compute and choose common metrics intentionally;
- reason about thresholds, imbalance, and calibration;
- establish a baseline and controlled comparison;
- run and document a small reproducible ML experiment;
- state what an experiment supports and what it does not support.

## Deliberately deferred

- deep statistical learning theory;
- VC dimension/PAC learning;
- formal bias-variance derivations;
- statistical significance testing in depth;
- extensive confidence-interval theory;
- automated hyperparameter-optimization frameworks;
- causal inference;
- advanced dataset-shift methodology;
- LLM-scale benchmark design/contamination;
- production monitoring and MLOps.

---

# Mathematical progression across Modules 01–07

The mathematics should feel cumulative rather than like isolated prerequisite courses.

| Module | New mathematical layer |
| --- | --- |
| 01 | Functions, mappings, composition, parameterized functions, basic notation |
| 02 | Vectors, matrices, dot products, matrix multiplication, transformations |
| 03 | Indexed data, summation, means, residuals, squared loss |
| 04 | Derivatives, partial derivatives, gradients, chain rule, gradient descent |
| 05 | Probability, random variables, expectation, variance, conditional probability |
| 06 | Exponentials, logarithms, odds/log-odds, likelihood, probabilistic loss |
| 07 | Generalization, regularization, evaluation metrics, calibration |

By the end of Module 07, the learner should have enough mathematical foundation that the later sequence

```text
neural network
    -> forward pass
    -> loss
    -> computational graph
    -> backpropagation
```

does not introduce any fundamental mathematical ingredient from nowhere. Backpropagation can then deepen and recombine familiar ideas rather than serving as the learner's first encounter with gradients or the chain rule.

## Revisit rather than front-load

Several topics intentionally receive more than one encounter:

- **functions/composition** — Module 01, then throughout models and later backprop;
- **vectors/matrices** — Module 02, then every model afterward, with deeper linear algebra deferred until PCA;
- **chain rule** — Module 04 introduction, then Module 11 backpropagation at much greater depth;
- **probability** — Module 05 foundation, then likelihood/classification in Module 06 and later language modeling/evaluation;
- **exponentials/logarithms** — reviewed when logistic regression makes them necessary, then reused in softmax/cross-entropy later;
- **numerical/reproducibility habits** — Module 01, reinforced in every notebook and experiment;
- **evaluation** — Module 07 foundation, then increasingly sophisticated evaluation in later classical ML, neural networks, and LLM modules.

This spiral is intentional. A first encounter should be complete enough to use correctly, not exhaustive enough to eliminate every future revisit.

# Historical progression across Modules 01–07

History should remain integrated rather than becoming separate survey lessons.

The first tranche can tell a coherent story:

1. mathematical functions provide a language for transformations;
2. coordinate/vector/matrix mathematics provides a language for structured numerical transformations;
3. least squares shows that fitting data with parameterized models long predates the term machine learning;
4. calculus and numerical optimization provide a way to improve parameters systematically;
5. probability provides a language for uncertain observations and inference from samples;
6. statistical classification turns model outputs into probabilistic predictions;
7. held-out evaluation and experimental discipline distinguish useful generalization from merely fitting observed data.

Later modules can then introduce the AI-specific historical arc—perceptrons, symbolic AI, neural-network winters/revivals, ImageNet, embeddings, seq2seq, attention, and Transformers—on top of a learner who already understands the mathematical machinery those systems use.

# Foundational implementation policy in this tranche

The learner should personally write the first meaningful implementations of:

- vector/dot/matrix operations in pedagogically small form;
- linear-regression prediction;
- MSE;
- simple parameter search;
- numerical differentiation;
- gradient descent;
- logistic sigmoid;
- binary cross-entropy;
- logistic-regression training;
- common evaluation metrics.

Libraries should appear **after** those mechanics are understood. Scikit-learn becomes useful in Module 06 and Module 07 because it can then be recognized as a production-quality abstraction over concepts the learner has already implemented.

Agents may review, test, debug, refactor, or extend foundational work after the learner has produced the first implementation, consistent with the course-wide policy.

# Tranche exit checkpoint

Completing Module 07 should mean substantially more than having encountered the vocabulary.

The learner should be able to take a modest numerical prediction problem and independently reason through:

```text
data representation
    -> model choice at the learned level
    -> parameterized prediction
    -> loss
    -> gradient-based fitting
    -> probabilistic interpretation where appropriate
    -> held-out evaluation
    -> metric choice
    -> error/limitation analysis
    -> reproducible conclusion
```

They should also be mathematically prepared for the next major parts of the core syllabus without having taken detached semester-length courses in linear algebra, calculus, or statistics.

# What comes next

After this tranche is authored and validated at lesson resolution, detailed planning should continue with:

- **08 — Classical ML Beyond Linear Models**;
- **09 — Unsupervised Learning and Dimensionality Reduction**;
- **10 — Neural Networks as Composed Functions**;
- **11 — Computational Graphs and Backpropagation**;
- **12 — Neural Network From Scratch**;
- **13 — PyTorch and Training Neural Networks**.

Modules 08–09 provide selective classical-ML breadth. Modules 10–13 then form the next major conceptual arc: neural networks from understandable mathematical pieces to framework-based deep learning.
