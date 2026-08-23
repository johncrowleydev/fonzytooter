import { useState } from 'react'
import { Badge, Button, Card } from '../../../components/ui'
import { InteractiveHeader } from './ArrayVisual'

const questions = [
  {
    prompt: 'For shape (2, 3, 4), which pair is correct?',
    answers: ['ndim = 3, size = 24', 'ndim = 24, size = 3', 'ndim = 4, size = 9'],
    correct: 0,
    explanation: 'The shape has three axes and 2 × 3 × 4 = 24 elements.',
  },
  {
    prompt: 'In shape (2, 3, 4), how long is axis 1?',
    answers: ['2', '3', '4'],
    correct: 1,
    explanation: 'Axis numbers index the shape tuple starting at zero.',
  },
  {
    prompt: 'If x.shape is (3, 4), what is x[0:1].shape?',
    answers: ['(4,)', '(1, 4)', 'scalar'],
    correct: 1,
    explanation: 'A slice preserves axis 0, even when it selects one position.',
  },
  {
    prompt: 'What commonly happens when a basic slice is mutated?',
    answers: [
      'Its source can change because storage is shared.',
      'NumPy always creates an independent copy.',
      'The dtype changes instead of the values.',
    ],
    correct: 0,
    explanation: 'Basic slices commonly return views. Use .copy() when independence matters.',
  },
  {
    prompt: 'Which statement about dtype is durable?',
    answers: [
      'It describes element representation, separately from shape.',
      'It is another name for ndim.',
      'Every element normally has an unrelated dtype.',
    ],
    correct: 0,
    explanation: 'Shape describes arrangement; dtype describes the common element representation.',
  },
] as const

export function ArrayMentalModelCheck() {
  const [questionIndex, setQuestionIndex] = useState(0)
  const [selected, setSelected] = useState<number | null>(null)
  const [completed, setCompleted] = useState<Set<number>>(() => new Set())
  const question = questions[questionIndex]
  const correct = selected === question.correct

  function answer(index: number) {
    setSelected(index)
    if (index === question.correct) setCompleted((current) => new Set(current).add(questionIndex))
  }

  function goTo(index: number) {
    setQuestionIndex(index)
    setSelected(null)
  }

  return (
    <Card className="my-8 grid gap-5 p-5">
      <InteractiveHeader
        title="Check your array mental model"
        description="Predict structure and storage behavior; retry any missed question."
        badge={
          <Badge tone={completed.size === questions.length ? 'teal' : 'neutral'}>
            {completed.size} / {questions.length} checked
          </Badge>
        }
      />
      <div className="flex flex-wrap gap-2" aria-label="Question navigation">
        {questions.map((_, index) => (
          <Button
            key={index}
            variant={index === questionIndex ? 'secondary' : 'outline'}
            onClick={() => goTo(index)}
          >
            Question {index + 1}
            {completed.has(index) ? ' ✓' : ''}
          </Button>
        ))}
      </div>
      <section className="grid gap-4 rounded-lg border border-line bg-white/5 p-4">
        <h3 className="text-sm leading-relaxed text-ink">{question.prompt}</h3>
        <div
          className="grid gap-2"
          role="group"
          aria-label={`Answers for question ${questionIndex + 1}`}
        >
          {question.answers.map((answerText, index) => (
            <Button
              key={answerText}
              variant={selected === index ? 'secondary' : 'outline'}
              onClick={() => answer(index)}
              disabled={correct}
            >
              {answerText}
            </Button>
          ))}
        </div>
        {selected !== null ? (
          <p
            className={
              correct
                ? 'rounded-md border border-brand-teal/30 bg-brand-teal/10 p-3 text-xs text-ink'
                : 'rounded-md border border-brand-coral/30 bg-brand-coral/10 p-3 text-xs text-ink'
            }
            role={correct ? 'status' : 'alert'}
          >
            <strong>{correct ? 'Correct.' : 'Not yet.'}</strong> {question.explanation}
          </p>
        ) : null}
      </section>
      <Button
        variant="quiet"
        onClick={() => {
          setCompleted(new Set())
          goTo(0)
        }}
      >
        Reset all answers
      </Button>
    </Card>
  )
}
