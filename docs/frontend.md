# Frontend conventions

Fonzytooter's frontend should remain simple, explicit, and easy for a human developer to read and evolve. These rules are intentionally opinionated because the application will be worked on heavily by coding agents, and generated code must remain maintainable after the prototype phase.

## Stack

The frontend stack is:

- React
- TypeScript with strict type checking
- Vite
- Tailwind CSS
- React Router
- MDX for authored lesson content

Add dependencies only when they solve a concrete current requirement.

## Styling: Tailwind first, custom CSS only by exception

Tailwind is the application's styling system, not merely a CSS dependency that happens to be installed.

Use Tailwind utility classes and Tailwind's extension mechanisms for normal application styling, including:

- layout;
- spacing;
- typography;
- responsive behavior;
- colors;
- borders and radii;
- hover/focus/active states;
- light/dark theme variants;
- transitions and ordinary animation;
- reusable design tokens through Tailwind configuration/theme facilities.

Do **not** create feature-specific global CSS classes simply to avoid writing Tailwind classes in JSX.

Custom CSS is acceptable only when there is a concrete requirement that Tailwind cannot reasonably express or when CSS itself is the subject of the implementation. Examples may include unusually complex generated visualizations, browser quirks that require selectors Tailwind cannot represent cleanly, or third-party integration overrides.

When custom CSS is added:

1. keep it as small and local as practical;
2. explain in a nearby comment or the change description why Tailwind was insufficient;
3. do not use the exception as a path toward a parallel handcrafted design system.

A large global stylesheet containing application layout, feature styling, component variants, and responsive rules is contrary to this architecture.

## Formatting: human readability is mandatory

Generated code is still source code intended for humans.

Use Prettier as the standard formatter for TypeScript, TSX, JavaScript, JSON, Markdown, and other supported frontend files. The repository should expose formatting commands suitable for local use and CI/check workflows.

Code must never be deliberately compressed to reduce physical line count. In particular, do not place large JSX trees, multiple nested components, long prop sets, or entire page layouts onto single lines.

Prefer formatting like:

```tsx
<Card className="p-6">
  <SectionHeading
    eyebrow="Current project"
    title="Neural Network From Scratch"
    action={
      <Link to="/projects/nn-scratch">
        Open project
      </Link>
    }
  />

  <p className="mt-3 text-sm text-slate-400">
    A repository-based lab for turning the pieces of this module into a small,
    inspectable implementation.
  </p>
</Card>
```

not a single-line equivalent.

Formatting should be automated rather than enforced by taste during review.

## Routing: use React Router

Routing is already a real application concern. Use React Router rather than implementing navigation with direct `window.history` manipulation, custom `popstate` listeners, path-prefix checks, or manual `split('/')` parameter parsing.

Use declarative routes and normal web navigation semantics:

```tsx
<Routes>
  <Route element={<AppShell />}>
    <Route index element={<Dashboard />} />
    <Route path="curriculum" element={<Curriculum />} />
    <Route path="curriculum/:moduleId" element={<ModuleDetail />} />
    <Route path="lesson/:lessonId" element={<Lesson />} />
    <Route path="review" element={<Review />} />
    <Route path="exercise/:exerciseId" element={<Exercise />} />
    <Route path="progress" element={<Progress />} />
    <Route path="projects" element={<Projects />} />
    <Route path="projects/:projectId" element={<ProjectDetail />} />
  </Route>
</Routes>
```

Prefer `Link`, `NavLink`, `useNavigate`, and `useParams` as appropriate. Navigation controls that semantically navigate should be links, preserving browser behavior such as opening in a new tab and correct back/forward history.

Do not grow a custom router incrementally.

## Component structure

Organize application code primarily by feature. The current direction is appropriate:

```text
src/
├── app/
├── components/
├── features/
│   ├── curriculum/
│   ├── dashboard/
│   ├── exercises/
│   ├── lessons/
│   ├── progress/
│   ├── projects/
│   ├── reviews/
│   └── tutor/
└── prototype/
```

Shared UI primitives belong in `components/` only when they are genuinely reused across features. Do not build a generic internal UI framework.

Within a feature, extract a component when it has a meaningful semantic or behavioral responsibility. Good examples include:

- `KnowledgeCheck`;
- `LessonSources`;
- `CodeWorkspace`;
- `ExerciseOutput`;
- `ReviewCard`;
- `ReviewRating`.

Do not extract components solely because a file reached an arbitrary number of lines. Conversely, do not hide a large page behind a deceptively small line count by compressing its JSX.

## State

Prefer the simplest state mechanism that fits the requirement:

- component-local state for local interactions;
- React context for genuinely cross-application concerns such as the global tutor;
- server-state libraries only when real server-state complexity justifies them.

Do not add a global state library preemptively.

## Tutor page context

Each active feature/page owns the semantic context it exposes to the tutor. Avoid duplicating route-to-context metadata in a top-level router and then immediately overwriting it inside the page.

A small helper hook such as `useTutorPageContext(...)` may be introduced when repetition becomes useful to remove, but the page remains authoritative for its own semantic context.

## Prototype data

Static prototype data should remain centralized and clearly identified as mock data. Do not build repositories, service abstractions, or persistence layers around fake prototype state.

When real data arrives, replace mock sources with real domain/API boundaries intentionally rather than allowing mock types to silently become permanent persistence contracts.

## Review checklist

Before considering frontend work complete, verify:

- TypeScript type checking passes;
- Prettier formatting passes;
- the production Vite build passes;
- application routing uses React Router rather than custom History API code;
- styling is expressed in Tailwind except for explicitly justified CSS exceptions;
- JSX is formatted for human readability;
- feature components have understandable responsibilities;
- navigation controls use correct link semantics;
- no unnecessary global state or framework abstractions were introduced.
