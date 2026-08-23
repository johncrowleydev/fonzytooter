import type { ComponentPropsWithoutRef } from 'react'
import Markdown, { type Components } from 'react-markdown'
import rehypeKatex from 'rehype-katex'
import remarkMath from 'remark-math'
import 'katex/dist/katex.min.css'

function withClassName(className: string, existingClassName?: string) {
  return existingClassName ? `${className} ${existingClassName}` : className
}

function WorksheetParagraph({ className, ...props }: ComponentPropsWithoutRef<'p'>) {
  return <p className={withClassName('mb-4 leading-7 text-body', className)} {...props} />
}

function WorksheetUnorderedList({ className, ...props }: ComponentPropsWithoutRef<'ul'>) {
  return (
    <ul
      className={withClassName('mb-4 ml-6 list-disc space-y-2 text-body', className)}
      {...props}
    />
  )
}

function WorksheetOrderedList({ className, ...props }: ComponentPropsWithoutRef<'ol'>) {
  return (
    <ol
      className={withClassName('mb-4 ml-6 list-decimal space-y-2 text-body', className)}
      {...props}
    />
  )
}

function WorksheetListItem({ className, ...props }: ComponentPropsWithoutRef<'li'>) {
  return <li className={withClassName('pl-1 leading-7', className)} {...props} />
}

function WorksheetStrong({ className, ...props }: ComponentPropsWithoutRef<'strong'>) {
  return <strong className={withClassName('font-semibold text-ink', className)} {...props} />
}

function WorksheetEmphasis({ className, ...props }: ComponentPropsWithoutRef<'em'>) {
  return <em className={withClassName('italic text-ink', className)} {...props} />
}

function WorksheetBlockquote({ className, ...props }: ComponentPropsWithoutRef<'blockquote'>) {
  return (
    <blockquote
      className={withClassName(
        'my-5 border-l-2 border-accent-teal/50 pl-4 italic text-muted',
        className,
      )}
      {...props}
    />
  )
}

function WorksheetCode({ className, ...props }: ComponentPropsWithoutRef<'code'>) {
  const codeClassName = className
    ? withClassName('font-mono text-sm leading-6 text-code-ink', className)
    : 'rounded bg-panel-soft px-1.5 py-0.5 font-mono text-sm text-accent-teal'

  return <code className={codeClassName} {...props} />
}

function WorksheetPreformatted({ className, ...props }: ComponentPropsWithoutRef<'pre'>) {
  return (
    <pre
      className={withClassName(
        'my-5 max-w-full overflow-x-auto rounded-lg border border-line bg-code-surface px-4 py-3 text-sm leading-6 text-code-ink',
        className,
      )}
      {...props}
    />
  )
}

const worksheetMarkdownComponents = {
  blockquote: WorksheetBlockquote,
  code: WorksheetCode,
  em: WorksheetEmphasis,
  li: WorksheetListItem,
  ol: WorksheetOrderedList,
  p: WorksheetParagraph,
  pre: WorksheetPreformatted,
  strong: WorksheetStrong,
  ul: WorksheetUnorderedList,
} satisfies Components

/**
 * Worksheet prose is trusted Markdown authored in the Git curriculum. This renderer must not be
 * reused for learner submissions, tutor output, or other untrusted content.
 */
export function WorksheetMarkup({ source }: { source: string }) {
  return (
    <Markdown
      components={worksheetMarkdownComponents}
      remarkPlugins={[remarkMath]}
      rehypePlugins={[rehypeKatex]}
    >
      {source}
    </Markdown>
  )
}
