# Public Curriculum and Authenticated Learner Features

Fonzytooter separates **public curriculum access** from **authenticated learner features**.

The curriculum is intended to remain openly readable. Authentication is not a gate around the site or the course content. It is a gate around capabilities that create or expose learner-specific state, mutate persistent application data, or consume metered external services.

## Product boundary

An anonymous visitor is a **curriculum reader**. An authenticated user is a **learner**.

### Public curriculum capabilities

The following should remain available without signing in:

- Browse courses, modules, and lessons.
- Read all lesson content.
- View curriculum diagrams, examples, and interactive teaching material.
- View worksheet and exercise prompts.
- Download public curriculum materials such as worksheets and workbooks.
- Run functionality that is entirely client-side and does not persist learner state, such as browser-only code execution or interactive examples.

Public curriculum access should behave like a normal public website. A visitor should be able to follow links directly to lessons and move through the curriculum without encountering a login wall.

### Authenticated learner capabilities

Authentication is required for features that belong to an individual learner or incur service cost, including:

- Lesson completion and progress tracking.
- Saved exercise attempts and results.
- Learner dashboards, mastery state, and activity history.
- Spaced-repetition scheduling and review history.
- Tutor conversations and AI tutor usage.
- Personalized or AI-assisted grading, generation, or other future metered services.
- Other learner-specific saved state such as preferences, notes, or bookmarks if added in the future.

This includes both writes to learner state and reads of private learner state. Learner-specific data must never be exposed to anonymous users.

## Core rule

> Curriculum is public. Learner state is private.

Reading or executing static/client-side curriculum functionality must not require authentication. Any operation that accesses learner-specific state, mutates persistent application state, or consumes metered external services requires an authenticated learner identity.

## Anonymous behavior

Anonymous visitors should not receive pseudo-accounts or anonymous learner records. Anonymous use is intentionally stateless from the learner system's perspective.

This avoids creating guest progress, guest review queues, guest tutor histories, or account-merging behavior.

If an anonymous visitor uses a public exercise or interactive successfully, the result may be shown in the browser, but it is not saved as learner progress. The UI may invite the visitor to sign in to save or track the result.

## Authentication UX

Authentication should be introduced at the point where a visitor tries to use a learner capability, not as a blanket redirect before viewing curriculum.

Examples:

- A lesson remains fully readable while an anonymous visitor sees a sign-in affordance instead of progress controls.
- An AI tutor entry point may remain visible, but using it requires sign-in.
- A browser-only exercise may remain usable, while saving the result requires sign-in.
- Learner-only destinations such as progress dashboards or review queues may prompt for authentication.

After signing in, the user should return to the context they were already using whenever possible.

## Metered services

Authentication is necessary for metered services such as AI, but authenticated access alone should not imply unlimited consumption.

Metered capabilities may additionally require an entitlement, quota, or other spending control. The product boundary is therefore:

1. Anonymous users cannot invoke metered services.
2. Authenticated users may invoke only the metered services they are authorized to use.

This separation allows future limits, plans, administrative access, or quotas without changing the public-curriculum principle.

## Guidance for future features

When deciding whether a new capability should require authentication, classify it by what it does rather than where it appears in the UI.

Typical public features:

- New lesson content.
- New curriculum visualizations.
- Public worksheet downloads.
- Browser-only examples or code execution.

Typical authenticated features:

- Saving any learner result.
- Reading personalized progress or mastery data.
- Scheduling or completing spaced-repetition reviews.
- Using an AI tutor or personalized AI generation.
- Persisting learner-specific notes, preferences, or history.

The default expectation is that curriculum remains public unless a feature crosses into learner-specific state or service consumption.
