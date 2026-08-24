package learner

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/database"
)

func TestVideoRecommendationsRankCurrentLessonPrerequisiteAndModuleFallback(t *testing.T) {
	service, _, _ := recommendationTestService(t, true)
	recommendations, err := service.VideoRecommendations(context.Background(), testUserID, "course", 3)
	if err != nil {
		t.Fatal(err)
	}
	want := []struct{ id, kind string }{
		{"current-video", VideoRecommendationCurrentLesson},
		{"prerequisite-video", VideoRecommendationNextPrerequisite},
		{"weak-video", VideoRecommendationCurrentModule},
	}
	if len(recommendations) != len(want) {
		t.Fatalf("recommendations = %#v", recommendations)
	}
	for index, expected := range want {
		if recommendations[index].Video.ID != expected.id || recommendations[index].ReasonKind != expected.kind {
			t.Fatalf("recommendation %d = %#v, want %#v", index, recommendations[index], expected)
		}
		if recommendations[index].Reason == "" {
			t.Fatalf("recommendation %d has no explanation", index)
		}
	}
}

func TestVideoRecommendationsUseRealRepeatedDifficultyForUnwatchedAndRevisit(t *testing.T) {
	service, _, _ := recommendationTestService(t, true)
	ctx := context.Background()
	now := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	for _, videoID := range []string{"current-video", "prerequisite-video", "revisit-video"} {
		if _, err := service.SetVideoProgress(ctx, testUserID, "course", "module", videoID, true); err != nil {
			t.Fatal(err)
		}
	}
	failed := []ExerciseTestResult{{TestID: "check", Status: "failed", Message: "try again"}}
	if _, err := service.CreateExerciseAttempt(ctx, testUserID, "course", "module", "weak-exercise", "pass", 1, failed); err != nil {
		t.Fatal(err)
	}
	oneAttempt, err := service.VideoRecommendations(ctx, testUserID, "course", 3)
	if err != nil {
		t.Fatal(err)
	}
	for _, recommendation := range oneAttempt {
		if recommendation.ReasonKind == VideoRecommendationWeakEvidence || recommendation.ReasonKind == VideoRecommendationRevisit {
			t.Fatalf("one failed attempt was treated as repeated difficulty: %#v", oneAttempt)
		}
	}
	if _, err := service.CreateExerciseAttempt(ctx, testUserID, "course", "module", "weak-exercise", "pass", 1, failed); err != nil {
		t.Fatal(err)
	}
	recommendations, err := service.VideoRecommendations(ctx, testUserID, "course", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(recommendations) != 3 || recommendations[0].Video.ID != "weak-video" || recommendations[0].ReasonKind != VideoRecommendationWeakEvidence {
		t.Fatalf("unwatched weak-evidence recommendation did not win: %#v", recommendations)
	}
	if recommendations[1].Video.ID != "revisit-video" || recommendations[1].ReasonKind != VideoRecommendationRevisit || !recommendations[1].Watched {
		t.Fatalf("watched revisit recommendation missing: %#v", recommendations)
	}
	if recommendations[2].Video.ID != "module-video" || recommendations[2].Watched {
		t.Fatalf("completed videos crowded out relevant unwatched material: %#v", recommendations)
	}
}

func TestVideoRecommendationsAreUserScoped(t *testing.T) {
	service, db, _ := recommendationTestService(t, true)
	insertLearnerTestUser(t, db, secondLearnerUserID)
	if _, err := service.SetVideoProgress(context.Background(), testUserID, "course", "module", "current-video", true); err != nil {
		t.Fatal(err)
	}
	first, err := service.VideoRecommendations(context.Background(), testUserID, "course", 1)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.VideoRecommendations(context.Background(), secondLearnerUserID, "course", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 1 || first[0].Video.ID != "prerequisite-video" {
		t.Fatalf("unexpected first learner recommendation: %#v", first)
	}
	if len(second) != 1 || second[0].Video.ID != "current-video" {
		t.Fatalf("second learner inherited watched state: %#v", second)
	}
}

func TestVideoRecommendationsHandleNoVideosAndAllWatched(t *testing.T) {
	emptyService, _, _ := recommendationTestService(t, false)
	empty, err := emptyService.VideoRecommendations(context.Background(), testUserID, "course", 3)
	if err != nil || len(empty) != 0 {
		t.Fatalf("no-video recommendations = %#v, %v", empty, err)
	}

	service, _, catalog := recommendationTestService(t, true)
	for _, video := range catalog.Courses()[0].Modules[0].Videos {
		if _, err := service.SetVideoProgress(context.Background(), testUserID, "course", "module", video.ID, true); err != nil {
			t.Fatal(err)
		}
	}
	allWatched, err := service.VideoRecommendations(context.Background(), testUserID, "course", 3)
	if err != nil || len(allWatched) != 0 {
		t.Fatalf("all-watched recommendations without difficulty = %#v, %v", allWatched, err)
	}
}

func recommendationTestService(t *testing.T, withVideos bool) (*Service, *sql.DB, *curriculum.Catalog) {
	t.Helper()
	videos := "videos: []\n"
	if withVideos {
		videos = `videos:
  - id: current-video
    youtubeId: dQw4w9WgXcQ
    title: Current video
    channel: Current creator
    durationMinutes: 5
    order: 0
    objectiveIds: [current]
    lessonIds: [current-lesson]
  - id: prerequisite-video
    youtubeId: 9bZkp7q19f0
    title: Prerequisite video
    channel: Prerequisite creator
    durationMinutes: 6
    order: 1
    objectiveIds: [prerequisite]
  - id: weak-video
    youtubeId: M7lc1UVf-VE
    title: Weak objective video
    channel: Practice creator
    durationMinutes: 7
    order: 2
    objectiveIds: [weak]
  - id: revisit-video
    youtubeId: aqz-KE-bpKQ
    title: Revisit video
    channel: Review creator
    durationMinutes: 8
    order: 3
    objectiveIds: [weak]
  - id: module-video
    youtubeId: ScMzIvxBSi4
    title: Module video
    channel: Module creator
    durationMinutes: 9
    order: 4
    objectiveIds: [current]
`
	}
	catalog, err := curriculum.Load(fstest.MapFS{
		"sources.yaml":                                      {Data: []byte("sources: {}\n")},
		"courses/course/course.yaml":                        {Data: []byte("id: course\ntitle: Course\ndescription: Course.\norder: 0\n")},
		"courses/course/modules/module/module.yaml":         {Data: []byte("id: module\ntitle: Module\norder: 0\nobjectives:\n  - id: prerequisite\n    title: Prerequisite objective\n    description: Prerequisite.\n    prerequisites: []\n  - id: current\n    title: Current objective\n    description: Current.\n    prerequisites: [prerequisite]\n  - id: weak\n    title: Weak objective\n    description: Weak.\n    prerequisites: []\n" + videos + "lessons: [current-lesson]\n")},
		"courses/course/modules/module/current.mdx":         {Data: []byte("---\nid: current-lesson\ntitle: Current lesson\nobjectiveIds: [current, weak]\nsourceIds: []\n---\n# Current\n")},
		"courses/course/modules/module/exercises/weak.yaml": {Data: []byte("id: weak-exercise\ntitle: Weak exercise\nlessonId: current-lesson\norder: 0\nobjectiveIds: [weak]\nprompt: Try it.\nstarterCode: pass\ntests:\n  - id: check\n    title: Check\n    visibility: visible\n    code: assert true\n")},
	})
	if err != nil {
		t.Fatalf("load recommendation catalog: %v", err)
	}
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewService(db, catalog), db, catalog
}
