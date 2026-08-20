import { useEffect, useRef } from 'react'
import { python } from '@codemirror/lang-python'
import { oneDark } from '@codemirror/theme-one-dark'
import { Compartment } from '@codemirror/state'
import { basicSetup, EditorView } from 'codemirror'

type CodeEditorProps = {
  value: string
  onChange: (value: string) => void
  disabled?: boolean
}

export function CodeEditor({ value, onChange, disabled = false }: CodeEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView>(null)
  const editable = useRef(new Compartment())
  const onChangeRef = useRef(onChange)
  onChangeRef.current = onChange

  useEffect(() => {
    if (!containerRef.current) return
    const view = new EditorView({
      parent: containerRef.current,
      doc: value,
      extensions: [
        basicSetup,
        python(),
        oneDark,
        editable.current.of(EditorView.editable.of(!disabled)),
        EditorView.lineWrapping,
        EditorView.updateListener.of((update) => {
          if (update.docChanged) onChangeRef.current(update.state.doc.toString())
        }),
        EditorView.theme({
          '&': { minHeight: '18rem' },
          '.cm-scroller': { overflow: 'auto', fontFamily: 'ui-monospace, monospace' },
          '.cm-content': { minHeight: '18rem', padding: '1rem 0' },
        }),
      ],
    })
    viewRef.current = view
    return () => {
      viewRef.current = null
      view.destroy()
    }
  }, [])

  useEffect(() => {
    viewRef.current?.dispatch({
      effects: editable.current.reconfigure(EditorView.editable.of(!disabled)),
    })
  }, [disabled])

  useEffect(() => {
    const view = viewRef.current
    if (!view || view.state.doc.toString() === value) return
    view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: value } })
  }, [value])

  return <div aria-label="Python exercise editor" ref={containerRef} />
}
