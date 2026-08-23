import type { ComponentPropsWithoutRef } from 'react'
import type { MDXComponents } from 'mdx/types'
import { CompositionPipeline } from '../components/CompositionPipeline'
import { InverseExplorer } from '../components/InverseExplorer'
import { MappingLab } from '../components/MappingLab'
import { MappingPropertiesLab } from '../components/MappingPropertiesLab'
import { MutableDefaultExplorer } from '../components/MutableDefaultExplorer'
import { PythonMentalModelCheck } from '../components/PythonMentalModelCheck'
import { ReferenceBindingExplorer } from '../components/ReferenceBindingExplorer'
import { SliceExplorer } from '../components/SliceExplorer'

function withClassName(className: string, existingClassName?: string) {
  return existingClassName ? `${className} ${existingClassName}` : className
}

function LessonHeading1({ className, ...props }: ComponentPropsWithoutRef<'h1'>) {
  return (
    <h1
      className={withClassName('mb-6 text-3xl font-semibold tracking-tight text-ink', className)}
      {...props}
    />
  )
}

function LessonHeading2({ className, ...props }: ComponentPropsWithoutRef<'h2'>) {
  return (
    <h2
      className={withClassName(
        'mb-4 mt-10 text-2xl font-semibold tracking-tight text-ink',
        className,
      )}
      {...props}
    />
  )
}

function LessonHeading3({ className, ...props }: ComponentPropsWithoutRef<'h3'>) {
  return (
    <h3
      className={withClassName(
        'mb-3 mt-8 text-xl font-semibold tracking-tight text-ink',
        className,
      )}
      {...props}
    />
  )
}

function LessonParagraph({ className, ...props }: ComponentPropsWithoutRef<'p'>) {
  return <p className={withClassName('mb-5 leading-7 text-body', className)} {...props} />
}

function LessonStrong({ className, ...props }: ComponentPropsWithoutRef<'strong'>) {
  return <strong className={withClassName('font-semibold text-ink', className)} {...props} />
}

function LessonEmphasis({ className, ...props }: ComponentPropsWithoutRef<'em'>) {
  return <em className={withClassName('italic text-ink', className)} {...props} />
}

function LessonBlockquote({ className, ...props }: ComponentPropsWithoutRef<'blockquote'>) {
  return (
    <blockquote
      className={withClassName(
        'my-6 border-l-2 border-accent-teal/50 pl-4 italic text-muted',
        className,
      )}
      {...props}
    />
  )
}

function LessonUnorderedList({ className, ...props }: ComponentPropsWithoutRef<'ul'>) {
  return (
    <ul
      className={withClassName('mb-5 ml-6 list-disc space-y-2 text-body', className)}
      {...props}
    />
  )
}

function LessonOrderedList({ className, ...props }: ComponentPropsWithoutRef<'ol'>) {
  return (
    <ol
      className={withClassName('mb-5 ml-6 list-decimal space-y-2 text-body', className)}
      {...props}
    />
  )
}

function LessonListItem({ className, ...props }: ComponentPropsWithoutRef<'li'>) {
  return <li className={withClassName('pl-1 leading-7', className)} {...props} />
}

function LessonHorizontalRule({ className, ...props }: ComponentPropsWithoutRef<'hr'>) {
  return (
    <hr className={withClassName('my-10 border-0 border-t border-line', className)} {...props} />
  )
}

function LessonCode({ className, ...props }: ComponentPropsWithoutRef<'code'>) {
  const codeClassName = className
    ? withClassName('block font-mono text-sm leading-6', className)
    : 'rounded bg-panel-soft px-1.5 py-0.5 font-mono text-[0.9em] text-accent-teal'

  return <code className={codeClassName} {...props} />
}

function LessonPreformatted({ className, ...props }: ComponentPropsWithoutRef<'pre'>) {
  return (
    <pre
      className={withClassName(
        'my-6 max-w-full overflow-x-auto overscroll-x-contain rounded-lg border border-line bg-code-surface px-4 py-3 text-sm leading-6 text-code-ink',
        className,
      )}
      {...props}
    />
  )
}

function LessonLink({ className, ...props }: ComponentPropsWithoutRef<'a'>) {
  return (
    <a
      className={withClassName(
        'text-accent-teal underline decoration-accent-teal/50 underline-offset-4 hover:text-accent-teal-light',
        className,
      )}
      {...props}
    />
  )
}

/**
 * The only components available to trusted curriculum MDX. Keep this registry explicit: names
 * authored in MDX must resolve to components defined here rather than being imported dynamically.
 */
export const lessonMdxComponents = {
  a: LessonLink,
  blockquote: LessonBlockquote,
  code: LessonCode,
  em: LessonEmphasis,
  h1: LessonHeading1,
  h2: LessonHeading2,
  h3: LessonHeading3,
  hr: LessonHorizontalRule,
  li: LessonListItem,
  ol: LessonOrderedList,
  p: LessonParagraph,
  pre: LessonPreformatted,
  strong: LessonStrong,
  ul: LessonUnorderedList,
  CompositionPipeline,
  InverseExplorer,
  MappingLab,
  MappingPropertiesLab,
  MutableDefaultExplorer,
  PythonMentalModelCheck,
  ReferenceBindingExplorer,
  SliceExplorer,
} satisfies MDXComponents
