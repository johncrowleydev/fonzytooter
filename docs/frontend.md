# Frontend conventions

Fonzytooter's frontend should remain simple, explicit, and easy for a human developer to read and evolve. These rules are intentionally opinionated because the application will be worked on heavily by coding agents, and generated code must remain maintainable after the prototype phase.

This document covers how frontend **code** should be written. `docs/ui-design-system.md` covers what the interface should **look like and how it should behave**: the design tokens, the type scale, theme handling, and the interaction rules that apply across touch, touchpad, mouse, and keyboard. Read both before styling work — the token rules in particular are load bearing, because choosing the wrong color family is the difference between 9:1 and 1.8:1 contrast in light mode.

## Stack

The frontend stack is:

- React
- TypeScript with strict type checking
- Vite
- Tailwind CSS
- React Router
- MDX for authored lesson content

### Authored lesson MDX

`LessonMdx` in `web/src/features/lessons/LessonMdx.tsx` is the runtime boundary for curriculum MDX
authored in Git. It uses `@mdx-js/mdx` to asynchronously evaluate trusted lesson source with the
static registry in `web/src/features/lessons/mdx/components.tsx`.

MDX evaluation executes compiled JavaScript. `LessonMdx` is therefore not safe for user
submissions, tutor output, arbitrary remote content, or database-authored untrusted markup. Keep
those inputs outside this renderer. Validate authored lesson bodies with `cd web && npm run
curriculum:mdx`; the Go curriculum validator remains responsible for semantic curriculum checks.

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

### Use Tailwind design tokens for intentional project-wide values

The ban on unnecessary custom CSS does **not** mean that every project-specific value should be encoded as an arbitrary utility in JSX.

Values that are intentionally part of Fonzytooter's visual language should be defined centrally through Tailwind's theme/design-token facilities and then consumed through normal utilities. Examples include the application palette, semantic foreground/background colors, recurring radii, or other values that are deliberately reused across the product.

Prefer a small, explicit set of project design tokens over repeatedly writing utilities such as `text-[var(--teal)]`, `bg-[var(--panel)]`, or the same raw RGBA value throughout the component tree.

Theme configuration is part of using Tailwind correctly. It is not considered a parallel handcrafted CSS design system.

The tokens themselves — which families exist, what each is for, and which are theme-aware — are defined in `docs/ui-design-system.md`. Do not introduce a new color, surface, or font size without reading it: several of the tokens replaced a hardcoded value that measured below 3:1 in one of the two themes, and the guards described there will fail on a reintroduction.

### Arbitrary-value utilities are an escape hatch

Arbitrary-value utilities such as `w-[242px]`, `mb-[17px]`, `text-[10px]`, `rounded-[14px]`, and `bg-[rgba(...)]` are allowed when a value is genuinely exceptional or cannot reasonably be represented by the project's Tailwind scale or design tokens.

They must **not** be the default way to translate a visual design into Tailwind.

Before using an arbitrary value, prefer in this order:

1. an existing Tailwind utility from the normal spacing, sizing, typography, radius, color, or breakpoint scale;
2. an existing Fonzytooter design token;
3. a new shared Tailwind design token when the value is intentionally reused or semantically meaningful;
4. an arbitrary value only when the value is truly one-off or unusually specific.

Do not preserve prototype pixel values merely because they existed in a mockup or earlier stylesheet. Prefer the nearest sensible Tailwind scale value unless the exact value materially affects the design.

Good arbitrary-value cases include unusual visualization geometry, a genuinely specific grid calculation, or a one-off size that has a concrete design reason. Ordinary margins, gaps, font sizes, radii, colors, and component dimensions should normally use the shared scale or tokens.

### Never dynamically construct Tailwind utility names

Tailwind utility names must never be assembled from runtime fragments such as `` `bg-${tone}-500` ``, `` `text-${color}-600` ``, or string concatenation that produces a utility name only after JavaScript executes.

When props or state select styling, map them to **complete, statically detectable utility strings** instead.

Prefer:

```ts
const badgeTones = {
  teal: 'border-teal-300 bg-teal-50 text-teal-700',
  gold: 'border-amber-300 bg-amber-50 text-amber-700',
} as const

const className = badgeTones[tone]
```

over:

```ts
const className = `border-${tone}-300 bg-${tone}-50 text-${tone}-700`
```

Finite visual states such as tone, status, activity kind, mastery level, navigation state, and tutor mode should generally be represented with typed lookup maps or other explicit selection between complete class strings rather than nested string-building logic.

Conditional `className` expressions are acceptable when each branch contains complete literal utilities, but avoid large nested ternaries in JSX. Extract meaningful visual variants into typed maps when doing so makes the component easier to read.

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
- recurring project-wide visual values use Tailwind theme/design tokens rather than repeated arbitrary utilities;
- arbitrary-value utilities are limited to genuinely exceptional values rather than used as a mechanical translation of prototype CSS;
- no Tailwind utility names are dynamically constructed from runtime string fragments;
- finite visual variants use complete statically detectable class strings, preferably through typed maps when that improves readability;
- JSX is formatted for human readability;
- feature components have understandable responsibilities;
- navigation controls use correct link semantics;
- no unnecessary global state or framework abstractions were introduced;
- the design-system checklist in `docs/ui-design-system.md` passes, including its automated guards.
