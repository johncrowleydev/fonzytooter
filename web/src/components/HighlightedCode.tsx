import { useMemo } from 'react'
import { classHighlighter, highlightTree } from '@lezer/highlight'
import { parser as pythonParser } from '@lezer/python'

/**
 * Static syntax highlighting for read-only code.
 *
 * This reuses the Lezer parser that already ships with `@codemirror/lang-python` for the exercise
 * editor rather than adding a highlighter of its own, so it costs no new install. `@lezer/python`
 * carries its own style tags, and `classHighlighter` emits stable `tok-*` class names that are
 * themed in styles.css.
 *
 * Python is the only grammar wired up, which covers 250 of the 263 highlightable curriculum code
 * fences. Everything else -- `text` blocks, plus a handful of js/ts/csharp/bash -- renders as plain
 * code, which is why an unknown language degrades to a single untagged token rather than throwing.
 */
type CodeToken = {
  text: string
  className?: string
}

const parsers = {
  python: pythonParser,
  py: pythonParser,
} as const

/** Strips the `language-` prefix MDX and Markdown put on a fenced block's `code` element. */
export function languageFromClassName(className?: string) {
  return className?.match(/(?:^|\s)language-([\w+-]+)/)?.[1]?.toLowerCase()
}

export function tokenizeCode(code: string, language?: string): CodeToken[] {
  const parser = language ? parsers[language as keyof typeof parsers] : undefined

  if (!parser) {
    return [{ text: code }]
  }

  const tokens: CodeToken[] = []
  let position = 0

  // highlightTree only reports styled ranges, so the gaps between them -- indentation, blank
  // lines, anything the grammar leaves untagged -- have to be filled in or the code comes out
  // with characters missing.
  highlightTree(parser.parse(code), classHighlighter, (from, to, className) => {
    if (from > position) {
      tokens.push({ text: code.slice(position, from) })
    }

    tokens.push({ text: code.slice(from, to), className })
    position = to
  })

  if (position < code.length) {
    tokens.push({ text: code.slice(position) })
  }

  return tokens
}

export function HighlightedCode({ code, language }: { code: string; language?: string }) {
  const tokens = useMemo(() => tokenizeCode(code, language), [code, language])

  return (
    <>
      {tokens.map((token, index) =>
        token.className ? (
          <span className={token.className} key={index}>
            {token.text}
          </span>
        ) : (
          token.text
        ),
      )}
    </>
  )
}
