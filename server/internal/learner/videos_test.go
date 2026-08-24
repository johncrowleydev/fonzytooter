package learner

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
	"github.com/johncrowleydev/fonzytooter/server/internal/database"
)

func TestVideoProgressLifecycleIsolationAndActivityIdempotency(t *testing.T) {
	service, db, catalog := videoTestService(t)
	insertLearnerTestUser(t, db, secondLearnerUserID)
	var courseID, moduleID, videoID string
	for _, course := range catalog.Courses() {
		for _, module := range course.Modules {
			if len(module.Videos) > 0 {
				courseID, moduleID, videoID = course.ID, module.ID, module.Videos[0].ID
				break
			}
		}
	}
	if videoID == "" {
		t.Fatal("test curriculum needs a video")
	}
	now := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	ctx := context.Background()

	initial, err := service.VideoProgress(ctx, testUserID, courseID, moduleID, videoID)
	if err != nil || initial.Completed {
		t.Fatalf("default progress: %#v, %v", initial, err)
	}
	for range 2 {
		completed, err := service.SetVideoProgress(ctx, testUserID, courseID, moduleID, videoID, true)
		if err != nil || !completed.Completed || completed.CompletedAt == nil || !completed.CompletedAt.Equal(now) {
			t.Fatalf("completed progress: %#v, %v", completed, err)
		}
	}
	second, err := service.VideoProgress(ctx, secondLearnerUserID, courseID, moduleID, videoID)
	if err != nil || second.Completed {
		t.Fatalf("second learner observed completion: %#v, %v", second, err)
	}
	assertRowCount(t, db, "video_progress", 1)
	assertRowCount(t, db, "activities", 1)
	activities, err := service.Activities(ctx, testUserID, courseID, 10)
	if err != nil || len(activities) != 1 || activities[0].VideoTitle == nil {
		t.Fatalf("video activity was not enriched: %#v, %v", activities, err)
	}
}

func TestVideoProgressRejectsInvalidOrRetiredIdentity(t *testing.T) {
	service, _, catalog := videoTestService(t)
	course := catalog.Courses()[0]
	if _, err := service.VideoProgress(context.Background(), testUserID, course.ID, course.Modules[0].ID, "retired-video"); !errors.Is(err, ErrVideoNotFound) {
		t.Fatalf("expected video not found, got %v", err)
	}
}

func videoTestService(t *testing.T) (*Service, *sql.DB, *curriculum.Catalog) {
	t.Helper()
	catalog, err := curriculum.Load(fstest.MapFS{
		"sources.yaml":                              {Data: []byte("sources: {}\n")},
		"courses/course/course.yaml":                {Data: []byte("id: course\ntitle: Course\ndescription: Course.\norder: 0\n")},
		"courses/course/modules/module/module.yaml": {Data: []byte("id: module\ntitle: Module\norder: 0\nobjectives:\n  - id: objective\n    title: Objective\n    description: Objective.\n    prerequisites: []\nvideos:\n  - id: video\n    youtubeId: dQw4w9WgXcQ\n    title: Video title\n    channel: Channel\n    durationMinutes: 5\n    order: 0\n    objectiveIds: [objective]\nlessons: []\n")},
	})
	if err != nil {
		t.Fatalf("load video catalog: %v", err)
	}
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open state database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewService(db, catalog), db, catalog
}
