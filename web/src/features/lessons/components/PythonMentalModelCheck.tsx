import { useState } from 'react'
import { Badge, Card } from '../../../components/ui'

const mentalModelQuestions = [
  {
    topic: 'Name rebinding vs object type',
    prompt: 'Given x = 10 followed by x = "ten", what changed?',
    options: [
      { id: 'rebind', label: 'The name x was rebound to a different object.' },
      { id: 'mutate', label: 'The integer object changed its type into a string.' },
      { id: 'annotation', label: 'Python inferred a permanent type for x.' },
    ],
    correctOption: 'rebind',
    explanation:
      'The integer object did not become a string. The name x now refers to a different object.',
    confirmation: 'Names can be rebound; objects keep their own runtime types.',
  },
  {
    topic: 'Reference assignment vs copying',
    prompt: 'Why does a also contain 3 after b = a and b.append(3)?',
    options: [
      { id: 'shared', label: 'a and b refer to the same mutable list object.' },
      { id: 'primitive', label: 'Python copies every list before a method call.' },
      { id: 'syntax', label: 'append changes every variable with a short name.' },
    ],
    correctOption: 'shared',
    explanation:
      'Assignment creates another reference to the same list. Use a.copy() when you need independent list storage.',
    confirmation: 'Another reference is not the same thing as a copied object.',
  },
  {
    topic: 'Annotations are not runtime enforcement',
    prompt: 'What does value: float guarantee when ordinary Python calls scale(value)?',
    options: [
      { id: 'runtime', label: 'Python itself rejects every non-float argument.' },
      {
        id: 'information',
        label: 'It records useful type information, but does not enforce it by itself.',
      },
      { id: 'conversion', label: 'Python automatically converts the argument to a float.' },
    ],
    correctOption: 'information',
    explanation:
      'Annotations help tools and readers. The ordinary Python runtime does not turn them into automatic argument checks.',
    confirmation:
      'Treat annotations as guidance for tools and humans unless extra runtime machinery is added.',
  },
  {
    topic: 'Slicing, filtering, and transformation',
    prompt:
      'What is the result of [value * 10 for value in values[1:4] if value % 2 == 0] when values is [1, 2, 3, 4, 5]?',
    options: [
      { id: 'twenty-forty', label: '[20, 40]' },
      { id: 'ten-forty', label: '[10, 40]' },
      { id: 'twenty-thirty-forty', label: '[20, 30, 40]' },
    ],
    correctOption: 'twenty-forty',
    explanation:
      'The slice keeps indices 1, 2, and 3: [2, 3, 4]. Filtering keeps [2, 4], then multiplying by 10 gives [20, 40].',
    confirmation: 'Trace the slice, then the filter, then the transformation.',
  },
  {
    topic: 'Equality vs identity',
    prompt: 'For two separately created lists with the same contents, which statement can be true?',
    options: [
      { id: 'equal-not-identical', label: 'a == b and a is not b' },
      { id: 'identity-only', label: 'a is b and a is not b' },
      { id: 'equality-only', label: 'a == b means both names must share one object' },
    ],
    correctOption: 'equal-not-identical',
    explanation:
      '== compares value equality. is compares object identity, so equal contents do not require shared storage.',
    confirmation:
      'Equality asks about value; identity asks whether the references reach one object.',
  },
] as const

type AnswerResult = 'correct' | 'incorrect' | null

type AnswerState = {
  selectedOption: string | null
  result: AnswerResult
  checked: boolean
}

const idleOptionClass =
  'border-line bg-raised text-muted hover:border-accent-teal/50 hover:bg-accent-teal/10 hover:text-ink'
const selectedOptionClass = 'border-accent-blue/50 bg-accent-blue/10 text-ink'
const incorrectOptionClass = 'border-accent-coral/50 bg-accent-coral/10 text-ink'
const correctOptionClass = 'border-accent-teal/50 bg-accent-teal/10 text-ink'

function createInitialAnswers(): AnswerState[] {
  return mentalModelQuestions.map(() => ({
    selectedOption: null,
    result: null,
    checked: false,
  }))
}

function getOptionClass(response: AnswerState, optionId: string) {
  if (response.checked && response.selectedOption === optionId) return correctOptionClass
  if (response.result === 'incorrect' && response.selectedOption === optionId) {
    return incorrectOptionClass
  }
  if (response.selectedOption === optionId) return selectedOptionClass
  return idleOptionClass
}

export function PythonMentalModelCheck() {
  const [answers, setAnswers] = useState<AnswerState[]>(createInitialAnswers)
  const checkedCount = answers.filter((answer) => answer.checked).length

  function chooseAnswer(questionIndex: number, optionId: string) {
    const question = mentalModelQuestions[questionIndex]
    const isCorrect = optionId === question.correctOption

    setAnswers((current) =>
      current.map((answer, index) => {
        if (index !== questionIndex) return answer

        return {
          selectedOption: optionId,
          result: isCorrect ? 'correct' : 'incorrect',
          checked: answer.checked || isCorrect,
        }
      }),
    )
  }

  return (
    <Card className="grid gap-5 p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-xs font-bold uppercase tracking-widest text-faint">Local check</p>
          <h2 className="mt-1.5 text-lg font-semibold tracking-tight">
            Check your Python mental model
          </h2>
          <p className="mt-2 max-w-2xl text-sm leading-relaxed text-muted">
            Choose an answer for each question. Incorrect choices include an explanation and remain
            retryable.
          </p>
        </div>
        <Badge tone={checkedCount === mentalModelQuestions.length ? 'teal' : 'neutral'}>
          {checkedCount} / {mentalModelQuestions.length} checked
        </Badge>
      </div>

      <div className="grid gap-4">
        {mentalModelQuestions.map((question, questionIndex) => {
          const response = answers[questionIndex]
          const questionId = `mental-model-question-${questionIndex}`

          return (
            <section key={question.topic} className="rounded-lg border border-line bg-raised p-4">
              <div className="flex items-start gap-3">
                <span className="grid size-7 shrink-0 place-items-center rounded-full border border-line-strong text-sm font-semibold text-ink">
                  {questionIndex + 1}
                </span>
                <div className="min-w-0 flex-1">
                  <p className="text-xs font-bold uppercase tracking-widest text-faint">
                    {question.topic}
                  </p>
                  <h3 className="mt-1.5 text-sm leading-relaxed text-ink" id={questionId}>
                    {question.prompt}
                  </h3>
                </div>
                <span className="shrink-0 text-sm text-muted">
                  {response.checked ? 'Checked' : 'Open'}
                </span>
              </div>

              <div className="mt-4 grid gap-2" role="group" aria-labelledby={questionId}>
                {question.options.map((option) => {
                  const selected = response.selectedOption === option.id

                  return (
                    <button
                      key={option.id}
                      className={`rounded-lg border px-3 py-2.5 text-left text-sm leading-relaxed transition focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent-teal ${getOptionClass(response, option.id)}`}
                      type="button"
                      onClick={() => chooseAnswer(questionIndex, option.id)}
                      disabled={response.checked}
                      aria-pressed={selected}
                      aria-label={`${option.label}${selected ? ', selected' : ''}`}
                    >
                      {option.label}
                    </button>
                  )
                })}
              </div>

              {response.result === 'incorrect' ? (
                <div
                  className="mt-3 rounded-md border border-accent-coral/30 bg-accent-coral/10 p-3 text-sm leading-relaxed text-ink"
                  role="alert"
                >
                  <strong>Not quite.</strong> {question.explanation} Try another answer.
                </div>
              ) : null}
              {response.result === 'correct' ? (
                <div
                  className="mt-3 rounded-md border border-accent-teal/30 bg-accent-teal/10 p-3 text-sm leading-relaxed text-ink"
                  role="status"
                  aria-live="polite"
                >
                  <strong>Correct.</strong> {question.confirmation}
                </div>
              ) : null}
            </section>
          )
        })}
      </div>

      <p className="border-t border-line pt-4 text-sm text-muted" role="status" aria-live="polite">
        {checkedCount === mentalModelQuestions.length
          ? '5 / 5 checked. This completion state is local to the component.'
          : `${checkedCount} of ${mentalModelQuestions.length} questions checked locally.`}
      </p>
    </Card>
  )
}
