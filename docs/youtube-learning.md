# YouTube learning

## Purpose

YouTube is a first-class learning medium in Fonzytooter.

Curated videos are not bibliography entries, footnotes, or an undifferentiated list of external resources. A good video can explain, visualize, demonstrate, motivate, or synthesize a concept in a way that complements authored prose, interactive components, exercises, worksheets, reviews, and projects.

Fonzytooter should therefore use YouTube in three related ways:

1. embed selected videos at the exact point in a lesson where they improve understanding;
2. expose the curated videos for a module as a browsable module playlist;
3. recommend relevant unwatched or revisit-worthy videos from learner progress and curriculum context.

The same curated video entity should support all three experiences. Do not create separate ad hoc link lists for lessons, playlists, and recommendations.

## Product principles

### Videos are curriculum content

A curated video is part of the authored curriculum. It should have stable curriculum identity, objective associations, and enough metadata for the application to present it consistently.

The authoritative curriculum remains Git-authored. YouTube supplies the media; Fonzytooter supplies the pedagogical placement, curation, objective mapping, progress context, and surrounding learning experience.

### Placement must be pedagogical

Do not add a video merely because one exists for the topic. Add it when it materially improves the lesson.

Useful roles include:

- introducing visual intuition before formal treatment;
- giving a second explanation of a difficult idea;
- showing a process that prose describes poorly;
- demonstrating implementation or experimentation;
- providing historical or conceptual context;
- synthesizing material after several smaller ideas have been introduced;
- revisiting a prerequisite through a different explanation.

Some lessons should have no video. Difficult lessons may justify more than one. Coverage targets such as "one video per lesson" are contrary to this model.

### Curated beats algorithmic

Fonzytooter should recommend from its authored curriculum collection, not surface arbitrary YouTube recommendations.

The platform knows what the learner is studying, which objectives a video supports, what has already been watched, and what learning evidence exists. That curriculum context is more valuable than a generic engagement-oriented recommendation graph.

### Watching is exposure, not mastery

Watching or completing a video is evidence that the learner encountered another explanation of an objective. It is not evidence that the learner can recall, apply, or transfer the concept.

Video state may influence recommendations and recency context. It must not directly manufacture recall, application, transfer, or mastery status.

## Authored video model

Modules own curated videos. The existing `videos` collection in `module.yaml` is the module's authoritative video catalog.

A video should have a stable ID and YouTube identity rather than being represented only as an arbitrary URL. Conceptually:

```yaml
videos:
  - id: gradient-descent-step-by-step
    youtubeId: example-video-id
    title: Gradient Descent, Step-by-Step
    channel: StatQuest
    durationMinutes: 23
    order: 0
    objectiveIds:
      - optimization.gradient-descent
    lessonIds:
      - 04-gradient-descent
```

The exact validated schema is an implementation concern and may evolve, but the authored model should preserve these concepts:

- stable curriculum video ID;
- YouTube video identity;
- authored display title;
- channel/creator identity;
- useful duration metadata;
- one or more objective associations;
- one or more lesson associations when applicable.

Do not treat YouTube titles, thumbnails, descriptions, or other remotely fetched metadata as application identity.

Playlist order is an explicit non-negative `order` authored for each module video. Values are unique within the module, and the validated catalog sorts by that value. This keeps the learning sequence stable and independent of YAML incidental order or YouTube ranking.

### Objective associations

Every curated video should support at least one objective. Objective mappings make the video meaningful beyond a media URL and enable curriculum-aware recommendations.

A video may support multiple objectives when the content genuinely spans them. Avoid mapping broadly just to increase recommendation eligibility.

### Lesson associations and placement are different

A video may be associated with one or more lessons in module metadata, but exact pedagogical placement belongs in lesson MDX.

Module metadata answers:

> What is this video and what curriculum concepts does it support?

Lesson MDX answers:

> Why is the learner seeing this video here?

Do not encode a fragile paragraph number, section index, or other prose-placement coordinate in YAML.

## Lesson MDX embeds

A lesson should be able to embed a curated module video by stable video ID at the relevant point in the prose.

Conceptually:

```mdx
Broadcasting is easier to reason about once you can see dimensions expand along compatible axes.

<YouTubeVideo id="broadcasting-visualized">
  Pay particular attention to how singleton dimensions expand. We will use that
  exact shape reasoning in the examples immediately after the video.
</YouTubeVideo>

Now return to the shapes `(3, 1)` and `(1, 4)`...
```

The surrounding lesson remains responsible for teaching. A video embed is not permission to replace authored explanation with "watch this."

### Contextual guidance

An embed may include short authored guidance explaining why the video is here or what the learner should notice.

Good guidance is specific:

> Watch how the derivative determines the direction of the parameter update. We will implement that mechanism next.

Weak guidance adds no pedagogical value:

> Here is a helpful video.

The guidance belongs to the placement context rather than global video metadata because the same video can serve different purposes in different locations.

### Lesson flow

Place an embed where it supports the instructional sequence, including in the middle of a lesson. Do not force all videos into a resources section at the bottom.

A video may appear:

- before a formal explanation to establish intuition;
- after an explanation to reinforce it through another medium;
- between two sections when it bridges concepts;
- before an exercise when it demonstrates a technique the learner is about to use;
- near the end when it synthesizes the lesson.

The author should choose the placement intentionally.

## Module playlist experience

Every module with curated videos should expose a module-level video collection derived from the same authored `videos` metadata.

This is a Fonzytooter playlist view, not a requirement to create or synchronize an actual YouTube playlist.

The view should make the module's video curriculum easy to browse and revisit. Useful presentation includes:

- thumbnail;
- title;
- channel;
- duration;
- associated lesson or lessons;
- associated objectives;
- watched/completed state for authenticated learners;
- a clear way to return to the lesson context where the video appears.

Module playlist order should be authored or deterministically derived from curriculum order. Do not let YouTube ranking determine curriculum sequence.

For unauthenticated users, the playlist remains publicly browsable with the rest of the public curriculum, but learner-specific completion state is absent.

## Learner video state

Authenticated learners may have state associated with curated videos.

At minimum, the learning model needs to distinguish an unwatched video from one the learner has completed. Future UX may justify additional state such as partial playback position, but do not introduce more tracking than the product actually uses.

Completion should create normal learner activity context such as `video_completed` and be scoped to the authenticated learner.

Video progress is learner state and therefore follows the authentication boundary: public users may view the video and surrounding lesson, but the application should not persist personal watch state for anonymous visitors.

### Completion semantics

"Completed" should mean that the application has sufficient evidence that the learner watched substantially all of the curated video according to the implemented player behavior. The UI should not imply perfect attention or comprehension.

Manual completion may be appropriate if platform/embed limitations make reliable playback telemetry unavailable. Whatever rule is implemented should be understandable and consistent rather than pretending to measure attention precisely.

## Home-page recommendations

The authenticated home page should be able to recommend curated videos from recorded learner progress.

Initial recommendations should be deterministic and curriculum-aware. A machine-learning recommender is unnecessary.

Useful recommendation signals include:

- the learner's current course, module, and lesson;
- videos embedded in the current or next lesson;
- unwatched videos supporting objectives introduced recently;
- videos supporting prerequisites for upcoming objectives;
- videos supporting objectives where application or recall evidence is weak;
- videos associated with repeated exercise difficulty;
- previously watched videos that are useful to revisit after evidence of difficulty;
- module sequence and recency.

A recommendation should be explainable in learner-facing terms when useful. Examples:

```text
Worth watching next
Why matrix multiplication works — 3Blue1Brown
This supports the matrix-transformation objective you just started.
```

```text
Review this visually
Gradient Descent, Step-by-Step — StatQuest
Recommended because recent gradient-descent practice has been difficult.
```

Avoid false precision such as a recommendation score presented as a measurement of learning need.

### Recommendation priority

A reasonable initial ordering preference is:

1. a highly relevant unwatched video for the learner's current lesson;
2. an unwatched video for the next immediate objective or prerequisite;
3. a video supporting an objective with recent weak evidence;
4. a revisit recommendation when recent difficulty makes another explanation useful;
5. other unwatched videos in the current module.

The exact ranking may evolve with the product, but relevance to learning should dominate novelty or engagement.

## Curation standards

### Preferred creators

The curriculum may maintain a preferred creator pool because some channels consistently produce unusually strong educational material. For the AI/ML course, channels worth checking first include:

- 3Blue1Brown;
- StatQuest;
- The Gradient Descent;
- Andrej Karpathy;
- Welch Labs;
- Serrano.Academy;
- DeepLearning.AI / Andrew Ng;
- Reducible;
- Sebastian Lague;
- Computerphile;
- research-oriented channels such as Yannic Kilcher when the curriculum reaches material where paper walkthroughs are appropriate.

This is a discovery preference, not an allowlist. The individual video's pedagogical fit matters more than the creator name.

### Review the actual video

Do not curate from title and thumbnail alone.

Before adding a video, inspect enough of the actual content to verify:

- it teaches what the curriculum says it teaches;
- its level fits the lesson placement;
- it does not rely on unintroduced prerequisites without a good reason;
- its explanation is accurate enough for the role it plays;
- the useful material is not buried inside an unnecessarily poor recommendation;
- the video remains available and embeddable at review time.

When practical, record why the video was selected in the PR description or review discussion so future curriculum maintenance has context.

### Avoid resource dumping

A module playlist should remain curated. Do not add ten similar videos simply because they are all good.

Prefer complementary explanations. For example, one visual-intuition treatment and one implementation-oriented treatment may both deserve inclusion because they do different jobs.

## Embed and UX expectations

YouTube embeds should behave like normal educational content rather than advertising surfaces.

The implementation should:

- use responsive embeds;
- avoid autoplay;
- use YouTube privacy-enhanced embedding where practical;
- preserve ordinary player controls required for learning, including captions and playback speed when YouTube provides them;
- provide a normal route to open the original video or channel on YouTube;
- work in the application's supported light and dark themes;
- remain usable with keyboard navigation and assistive technology;
- avoid layout shifts that make lesson reading unpleasant;
- display authored title/channel/context consistently around the player rather than relying entirely on iframe chrome.

Do not copy or re-host video content or thumbnails in ways that conflict with YouTube's terms or creator rights.

## Availability and failure handling

External videos can disappear, become private, disable embedding, change title, or otherwise become unavailable.

The curriculum validator or maintenance tooling should eventually make broken curated videos discoverable, but an unavailable video must not make the surrounding lesson unreadable.

The lesson should degrade to a clear unavailable-resource state while preserving the authored prose before and after it.

Removing or replacing a previously curated video should be treated as a curriculum maintenance decision. If persisted learner state uses stable video IDs, identity changes must preserve or deliberately retire that state according to the repository's curriculum-identity policy.

## Public curriculum and authentication

YouTube learning follows the same public-content boundary as lessons.

Public users may:

- view embedded videos;
- browse module video playlists;
- follow video links to YouTube.

Authenticated users additionally receive learner-state features such as:

- recorded video completion;
- watched/unwatched presentation;
- progress-aware recommendations;
- recommendation explanations grounded in their recorded learning history.

Viewing public educational content must not require authentication merely because the content is a YouTube embed.

## Authoring workflow

When authoring or revising a lesson:

1. identify concepts where another medium could materially improve understanding;
2. search preferred creators and other high-quality sources for candidate videos;
3. review the actual candidate content;
4. choose only videos with a clear pedagogical role;
5. add the video to the owning module's curated catalog with stable identity and objective mappings;
6. embed it in lesson MDX at the exact instructional point where it belongs;
7. add short contextual viewing guidance when that improves the learning task;
8. verify the lesson still teaches coherently if the external video is temporarily unavailable;
9. review the module playlist as a whole for duplication, ordering, and balance.

Video curation is part of lesson authoring, not an after-the-fact resource sweep.

## Implementation boundaries

This document defines product and curriculum behavior, not the implementation decomposition.

Implementation should preserve the following boundaries:

- authored video identity and objective mapping live with curriculum data;
- exact lesson placement lives in MDX;
- public video content remains readable without authentication;
- personal watch/completion state is authenticated learner state;
- recommendations derive from curated curriculum metadata plus recorded learner state;
- watching a video never directly establishes recall, application, transfer, or mastery.
