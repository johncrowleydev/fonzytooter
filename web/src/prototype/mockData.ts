import type { ActivityItem, CurriculumModule, Objective, Project, ReviewCard } from './types'

export const modules: CurriculumModule[] = [
  {
    id: 'scientific-python',
    eyebrow: 'Foundations',
    title: 'Scientific Python',
    description: 'The small, expressive toolkit you need to experiment with data and models.',
    status: 'completed',
    objectiveIds: ['python.arrays', 'python.visualization'],
    accent: 'teal',
    lessons: [
      {
        id: 'python-arrays',
        title: 'Thinking in arrays',
        kind: 'lesson',
        completed: true,
        objectiveIds: ['python.arrays'],
      },
      {
        id: 'python-visualization',
        title: 'A visual language for data',
        kind: 'video',
        completed: true,
        objectiveIds: ['python.visualization'],
      },
    ],
  },
  {
    id: 'mathematical-foundations',
    eyebrow: 'Foundations',
    title: 'Mathematical Foundations',
    description:
      'Build the calculus, linear algebra, and probability intuition that makes ML legible.',
    status: 'in-progress',
    objectiveIds: [
      'calculus.derivative-intuition',
      'linear-algebra.vectors',
      'probability.distributions',
    ],
    accent: 'gold',
    lessons: [
      {
        id: 'derivative-intuition',
        title: 'Derivatives as local change',
        kind: 'lesson',
        completed: true,
        objectiveIds: ['calculus.derivative-intuition'],
      },
      {
        id: 'vectors-and-geometry',
        title: 'Vectors as directions and data',
        kind: 'lesson',
        completed: false,
        objectiveIds: ['linear-algebra.vectors'],
      },
      {
        id: 'probability-distributions',
        title: 'Distributions as stories about uncertainty',
        kind: 'lesson',
        completed: false,
        objectiveIds: ['probability.distributions'],
      },
    ],
  },
  {
    id: 'classical-ml',
    eyebrow: 'Foundations',
    title: 'Classical Machine Learning',
    description: 'Move from mathematical ingredients to models that learn from examples.',
    status: 'available',
    objectiveIds: ['ml.train-test', 'ml.linear-regression'],
    accent: 'violet',
    lessons: [
      {
        id: 'train-test-thinking',
        title: 'What it means to learn from data',
        kind: 'lesson',
        completed: false,
        objectiveIds: ['ml.train-test'],
      },
      {
        id: 'linear-regression',
        title: 'Linear regression',
        kind: 'exercise',
        completed: false,
        objectiveIds: ['ml.linear-regression'],
      },
    ],
  },
  {
    id: 'neural-networks',
    eyebrow: 'Neural networks',
    title: 'Neural Networks From Scratch',
    description:
      'See the moving parts of a neural network clearly enough to implement them yourself.',
    status: 'in-progress',
    objectiveIds: ['nn.neuron', 'nn.activations', 'nn.chain-rule', 'nn.backpropagation'],
    prerequisites: ['Mathematical Foundations'],
    accent: 'coral',
    lessons: [
      {
        id: 'from-linear-models',
        title: 'From linear models to neurons',
        kind: 'lesson',
        completed: true,
        objectiveIds: ['nn.neuron'],
      },
      {
        id: 'activation-functions',
        title: 'Activation functions',
        kind: 'lesson',
        completed: true,
        objectiveIds: ['nn.activations'],
      },
      {
        id: 'computational-graphs',
        title: 'Computational graphs',
        kind: 'lesson',
        completed: false,
        objectiveIds: ['nn.chain-rule'],
      },
      {
        id: 'backpropagation',
        title: 'Backpropagation',
        kind: 'lesson',
        completed: false,
        objectiveIds: ['nn.backpropagation', 'nn.chain-rule'],
      },
      {
        id: 'gradient-descent-exercise',
        title: 'Implement gradient descent',
        kind: 'exercise',
        completed: false,
        objectiveIds: ['nn.backpropagation'],
      },
      {
        id: 'nn-playlist',
        title: 'Visual intuition for backpropagation',
        kind: 'video',
        completed: false,
        objectiveIds: ['nn.chain-rule'],
      },
      {
        id: 'nn-from-scratch-lab',
        title: 'Neural Network From Scratch',
        kind: 'lab',
        completed: false,
        objectiveIds: ['nn.neuron', 'nn.backpropagation'],
      },
    ],
  },
  {
    id: 'deep-learning',
    eyebrow: 'Neural networks',
    title: 'Deep Learning',
    description:
      'Understand depth, representation learning, and the training choices behind modern models.',
    status: 'locked',
    objectiveIds: ['deep.representations'],
    prerequisites: ['Neural Networks From Scratch'],
    accent: 'blue',
    lessons: [],
  },
  {
    id: 'transformers',
    eyebrow: 'Modern AI',
    title: 'Transformers',
    description:
      'Build an intuition for attention, sequence modeling, and the architecture behind LLMs.',
    status: 'locked',
    objectiveIds: ['transformers.attention'],
    prerequisites: ['Deep Learning'],
    accent: 'violet',
    lessons: [],
  },
  {
    id: 'modern-llms',
    eyebrow: 'Modern AI',
    title: 'Modern LLMs',
    description:
      'Connect transformer foundations to language-model training, alignment, and evaluation.',
    status: 'locked',
    objectiveIds: ['llm.training'],
    prerequisites: ['Transformers'],
    accent: 'gold',
    lessons: [],
  },
  {
    id: 'ml-systems',
    eyebrow: 'Modern AI',
    title: 'ML Systems',
    description:
      'Make thoughtful tradeoffs when models leave the notebook and meet real constraints.',
    status: 'locked',
    objectiveIds: ['systems.reliability'],
    prerequisites: ['Modern LLMs'],
    accent: 'teal',
    lessons: [],
  },
]

export const objectives: Objective[] = [
  {
    id: 'calculus.derivative-intuition',
    title: 'Explain a derivative as local change',
    description: 'Connect slope, local approximation, and the direction a function is changing.',
    moduleId: 'mathematical-foundations',
    prerequisiteIds: [],
    introduced: true,
    recall: 'strong',
    conceptual: 'strong',
    application: 'developing',
    transfer: 'not-assessed',
  },
  {
    id: 'linear-algebra.vectors',
    title: 'Use vectors as representations',
    description: 'Read vectors as both geometric directions and containers for structured data.',
    moduleId: 'mathematical-foundations',
    prerequisiteIds: [],
    introduced: true,
    recall: 'strong',
    conceptual: 'developing',
    application: 'developing',
    transfer: 'not-assessed',
  },
  {
    id: 'nn.neuron',
    title: 'Explain a neuron as a parameterized function',
    description: 'Describe how weights, bias, and an activation turn inputs into an output.',
    moduleId: 'neural-networks',
    prerequisiteIds: ['linear-algebra.vectors'],
    introduced: true,
    recall: 'strong',
    conceptual: 'strong',
    application: 'strong',
    transfer: 'developing',
  },
  {
    id: 'nn.activations',
    title: 'Choose and reason about activation functions',
    description: 'Recognize why non-linearities matter and how their shapes affect optimization.',
    moduleId: 'neural-networks',
    prerequisiteIds: ['nn.neuron'],
    introduced: true,
    recall: 'strong',
    conceptual: 'strong',
    application: 'developing',
    transfer: 'not-assessed',
  },
  {
    id: 'nn.chain-rule',
    title: 'Apply the chain rule to a computational graph',
    description:
      'Trace local derivatives through composed functions to find a useful global gradient.',
    moduleId: 'neural-networks',
    prerequisiteIds: ['calculus.derivative-intuition', 'nn.neuron'],
    introduced: true,
    recall: 'developing',
    conceptual: 'developing',
    application: 'developing',
    transfer: 'not-assessed',
  },
  {
    id: 'nn.backpropagation',
    title: 'Implement gradient descent and backpropagation',
    description:
      'Use gradients to reduce a loss, while preserving the structure of the computation.',
    moduleId: 'neural-networks',
    prerequisiteIds: ['nn.chain-rule'],
    introduced: true,
    recall: 'developing',
    conceptual: 'developing',
    application: 'developing',
    transfer: 'not-assessed',
  },
  {
    id: 'ml.train-test',
    title: 'Separate training evidence from generalization evidence',
    description: 'Use train, validation, and test sets to reason about what a model has learned.',
    moduleId: 'classical-ml',
    prerequisiteIds: [],
    introduced: false,
    recall: 'not-assessed',
    conceptual: 'not-assessed',
    application: 'not-assessed',
    transfer: 'not-assessed',
  },
]

export const reviewCards: ReviewCard[] = [
  {
    id: 'review-gradient',
    prompt: 'What does the gradient represent at a point on a loss surface?',
    answer:
      'It points in the direction of steepest increase. To reduce loss, an optimizer steps in the opposite direction.',
    objectiveId: 'nn.backpropagation',
    objectiveLabel: 'Gradients & optimization',
    lastReviewed: '3 days ago',
    hint: 'Think about slope, but in more than one dimension.',
  },
  {
    id: 'review-chain',
    prompt: 'Why can the chain rule be applied repeatedly in a neural network?',
    answer:
      'A network is a composition of functions. The derivative of the composition is the product of local derivatives along the path.',
    objectiveId: 'nn.chain-rule',
    objectiveLabel: 'Computational graphs',
    lastReviewed: '5 days ago',
    hint: 'Follow one output back through the operations that produced it.',
  },
  {
    id: 'review-neuron',
    prompt: 'What makes a neuron more than a plain linear transformation?',
    answer:
      'The activation function introduces a non-linearity after the weighted sum, allowing layers to compose richer functions.',
    objectiveId: 'nn.neuron',
    objectiveLabel: 'Neurons & activations',
    lastReviewed: '1 day ago',
    hint: 'What happens if every layer is only linear?',
  },
  {
    id: 'review-learning-rate',
    prompt: 'What is the role of the learning rate in gradient descent?',
    answer:
      'It scales each update. Too small can make progress slow; too large can overshoot or destabilize training.',
    objectiveId: 'nn.backpropagation',
    objectiveLabel: 'Gradients & optimization',
    lastReviewed: '6 days ago',
    hint: 'It controls how far each step travels.',
  },
]

export const recentActivity: ActivityItem[] = [
  {
    id: 'activity-1',
    label: 'Reviewed chain rule',
    detail: 'Computational graphs',
    time: '12 min ago',
    kind: 'review',
  },
  {
    id: 'activity-2',
    label: 'Completed “Gradient intuition”',
    detail: 'Neural Networks From Scratch',
    time: 'Yesterday',
    kind: 'lesson',
  },
  {
    id: 'activity-3',
    label: 'Attempted gradient descent exercise',
    detail: '1 test still failing',
    time: 'Yesterday',
    kind: 'exercise',
  },
  {
    id: 'activity-4',
    label: 'Asked tutor about learning rates',
    detail: 'Tutor conversation',
    time: '2 days ago',
    kind: 'tutor',
  },
]

export const projects: Project[] = [
  {
    id: 'nn-scratch',
    title: 'Neural Network From Scratch',
    description:
      'A repository-based lab that turns the pieces of this module into a small, inspectable implementation.',
    status: 'in-progress',
    repository: 'github.com/helix-academy/neural-net-from-scratch',
    objectives: [
      { label: 'Forward propagation', state: 'done' },
      { label: 'Gradient computation', state: 'working' },
      { label: 'Backpropagation', state: 'todo' },
      { label: 'Training loop', state: 'todo' },
    ],
    deliverables: ['implementation', 'tests', 'experiment notes', 'short reflection'],
    boundaryNote:
      'This is deliberate project work: the code belongs in a real repository and IDE, while Helix Academy keeps the objectives, prompts, and reflection visible.',
  },
  {
    id: 'linear-regression-lab',
    title: 'Linear Regression Lab',
    description: 'A small guided lab for comparing analytical and iterative solutions.',
    status: 'not-started',
    repository: 'github.com/helix-academy/linear-regression-lab',
    objectives: [
      { label: 'Fit a linear model', state: 'todo' },
      { label: 'Compare loss curves', state: 'todo' },
    ],
    deliverables: ['implementation', 'short reflection'],
    boundaryNote:
      'Use the browser exercises for tight practice. Move this lab to a normal repo when you are ready to experiment beyond the prompt.',
  },
]

export const activeModule = modules.find((module) => module.id === 'neural-networks')!
export const currentLesson = activeModule.lessons.find((lesson) => lesson.id === 'backpropagation')!
