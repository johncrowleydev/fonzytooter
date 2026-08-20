package review

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	fsrs "github.com/open-spaced-repetition/go-fsrs/v4"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
)

const ActivityReviewCompleted = "review_completed"

var (
	ErrCourseNotFound     = errors.New("course not found")
	ErrReviewItemNotFound = errors.New("review item not found")
	ErrInvalidRating      = errors.New("invalid review rating")
)

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }

type Rating string

const (
	RatingAgain Rating = "again"
	RatingHard  Rating = "hard"
	RatingGood  Rating = "good"
	RatingEasy  Rating = "easy"
)

type Preview struct {
	Rating          Rating
	DueAt           time.Time
	IntervalSeconds int64
	IntervalDays    float64
}

type Card struct {
	Item      curriculum.ReviewItem
	Schedule  fsrs.Card
	Virtual   bool
	Previews  []Preview
	IsDue     bool
	New       bool
	LastRated *time.Time
}

type SubmittedReview struct {
	ID   int64
	Card Card
}

type Service struct {
	db        *sql.DB
	catalog   *curriculum.Catalog
	clock     Clock
	scheduler *fsrs.FSRS
}

func NewService(db *sql.DB, catalog *curriculum.Catalog, clock Clock) *Service {
	if db == nil {
		panic("review.NewService: nil database")
	}
	if catalog == nil {
		panic("review.NewService: nil curriculum catalog")
	}
	if clock == nil {
		panic("review.NewService: nil clock")
	}
	return &Service{
		db:        db,
		catalog:   catalog,
		clock:     clock,
		scheduler: fsrs.NewFSRS(fsrs.DefaultParam()),
	}
}

func (s *Service) Cards(ctx context.Context, courseID string, dueOnly bool) ([]Card, error) {
	course, ok := s.catalog.CourseByID(courseID)
	if !ok {
		return nil, ErrCourseNotFound
	}

	stored, err := s.loadCourseCards(ctx, courseID)
	if err != nil {
		return nil, err
	}

	now := s.clock.Now().UTC()
	type orderedCard struct {
		card          Card
		authoredIndex int
	}
	ordered := make([]orderedCard, 0)
	authoredIndex := 0
	for _, module := range course.Modules {
		for _, item := range module.ReviewItems {
			key := cardKey{moduleID: module.ID, reviewItemID: item.ID}
			schedule, exists := stored[key]
			if !exists {
				schedule = fsrs.NewCard(now)
			}
			isDue := !schedule.Due.After(now)
			if dueOnly && !isDue {
				authoredIndex++
				continue
			}
			previews, err := s.previews(schedule, now)
			if err != nil {
				return nil, fmt.Errorf("preview review item %q: %w", item.ID, err)
			}
			var lastRated *time.Time
			if !schedule.LastReview.IsZero() {
				value := schedule.LastReview.UTC()
				lastRated = &value
			}
			ordered = append(ordered, orderedCard{
				card: Card{
					Item: item, Schedule: schedule, Virtual: !exists, Previews: previews,
					IsDue: isDue, New: schedule.State == fsrs.New, LastRated: lastRated,
				},
				authoredIndex: authoredIndex,
			})
			authoredIndex++
		}
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := ordered[i], ordered[j]
		if left.card.Virtual != right.card.Virtual {
			return !left.card.Virtual
		}
		if !left.card.Virtual && !left.card.Schedule.Due.Equal(right.card.Schedule.Due) {
			return left.card.Schedule.Due.Before(right.card.Schedule.Due)
		}
		if left.authoredIndex != right.authoredIndex {
			return left.authoredIndex < right.authoredIndex
		}
		return left.card.Item.ID < right.card.Item.ID
	})

	result := make([]Card, len(ordered))
	for index := range ordered {
		result[index] = ordered[index].card
	}
	return result, nil
}

func (s *Service) Submit(ctx context.Context, courseID, reviewItemID string, rating Rating) (SubmittedReview, error) {
	item, err := s.findReviewItem(courseID, reviewItemID)
	if err != nil {
		return SubmittedReview{}, err
	}
	grade, err := fsrsRating(rating)
	if err != nil {
		return SubmittedReview{}, err
	}

	now := s.clock.Now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return SubmittedReview{}, fmt.Errorf("begin review transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	before, found, err := loadCard(ctx, tx, courseID, item.ModuleID, item.ID)
	if err != nil {
		return SubmittedReview{}, err
	}
	if !found {
		before = fsrs.NewCard(now)
	}
	after, err := s.scheduler.Next(before, now, grade)
	if err != nil {
		return SubmittedReview{}, fmt.Errorf("schedule review: %w", err)
	}
	if err := storeCard(ctx, tx, courseID, item.ModuleID, item.ID, after.Card, now); err != nil {
		return SubmittedReview{}, err
	}
	reviewID, err := insertReviewLog(ctx, tx, courseID, item.ModuleID, item.ID, rating, now, before, after.Card)
	if err != nil {
		return SubmittedReview{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO activities (kind, course_id, module_id, review_item_id, occurred_at)
		VALUES (?, ?, ?, ?, ?)
	`, ActivityReviewCompleted, courseID, item.ModuleID, item.ID, formatTime(now)); err != nil {
		return SubmittedReview{}, fmt.Errorf("record review activity: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return SubmittedReview{}, fmt.Errorf("commit review transaction: %w", err)
	}

	previews, err := s.previews(after.Card, now)
	if err != nil {
		return SubmittedReview{}, fmt.Errorf("preview resulting review card: %w", err)
	}
	lastRated := now
	return SubmittedReview{
		ID: reviewID,
		Card: Card{
			Item: item, Schedule: after.Card, Previews: previews,
			IsDue: !after.Card.Due.After(now), New: false, LastRated: &lastRated,
		},
	}, nil
}

func (s *Service) findReviewItem(courseID, reviewItemID string) (curriculum.ReviewItem, error) {
	course, ok := s.catalog.CourseByID(courseID)
	if !ok {
		return curriculum.ReviewItem{}, ErrCourseNotFound
	}
	var found *curriculum.ReviewItem
	for _, module := range course.Modules {
		for _, item := range module.ReviewItems {
			if item.ID != reviewItemID {
				continue
			}
			if found != nil {
				return curriculum.ReviewItem{}, fmt.Errorf("review item id %q is ambiguous within course %q", reviewItemID, courseID)
			}
			copy := item
			found = &copy
		}
	}
	if found == nil {
		return curriculum.ReviewItem{}, ErrReviewItemNotFound
	}
	return *found, nil
}

func (s *Service) previews(card fsrs.Card, now time.Time) ([]Preview, error) {
	records, err := s.scheduler.Repeat(card, now)
	if err != nil {
		return nil, err
	}
	previews := make([]Preview, 0, 4)
	for _, candidate := range []struct {
		rating Rating
		grade  fsrs.Rating
	}{
		{RatingAgain, fsrs.Again},
		{RatingHard, fsrs.Hard},
		{RatingGood, fsrs.Good},
		{RatingEasy, fsrs.Easy},
	} {
		dueAt := records[candidate.grade].Card.Due.UTC()
		interval := dueAt.Sub(now)
		if interval < 0 {
			interval = 0
		}
		previews = append(previews, Preview{
			Rating: candidate.rating, DueAt: dueAt,
			IntervalSeconds: int64(interval.Round(time.Second) / time.Second),
			IntervalDays:    interval.Hours() / 24,
		})
	}
	return previews, nil
}

type cardKey struct {
	moduleID     string
	reviewItemID string
}

func (s *Service) loadCourseCards(ctx context.Context, courseID string) (map[cardKey]fsrs.Card, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT module_id, review_item_id, due_at, stability, difficulty,
			scheduled_days, reps, lapses, state, last_review_at, remaining_steps
		FROM review_cards
		WHERE course_id = ?
	`, courseID)
	if err != nil {
		return nil, fmt.Errorf("read review cards: %w", err)
	}
	defer rows.Close()

	result := make(map[cardKey]fsrs.Card)
	for rows.Next() {
		var moduleID, reviewItemID string
		card, err := scanCard(rows, &moduleID, &reviewItemID)
		if err != nil {
			return nil, err
		}
		result[cardKey{moduleID: moduleID, reviewItemID: reviewItemID}] = card
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate review cards: %w", err)
	}
	return result, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCard(scanner rowScanner, prefix ...any) (fsrs.Card, error) {
	var dueAt, lastReviewAt sql.NullString
	var stability, difficulty float64
	var scheduledDays, reps, lapses uint64
	var state int
	var remainingSteps int
	dest := append(prefix,
		&dueAt, &stability, &difficulty, &scheduledDays, &reps, &lapses,
		&state, &lastReviewAt, &remainingSteps,
	)
	if err := scanner.Scan(dest...); err != nil {
		return fsrs.Card{}, fmt.Errorf("scan review card: %w", err)
	}
	due, err := parseRequiredTime("review card due", dueAt)
	if err != nil {
		return fsrs.Card{}, err
	}
	lastReview, err := parseOptionalTime("review card last review", lastReviewAt)
	if err != nil {
		return fsrs.Card{}, err
	}
	return fsrs.Card{
		Due: due, Stability: stability, Difficulty: difficulty,
		ScheduledDays: scheduledDays, Reps: reps, Lapses: lapses,
		State: fsrs.State(state), LastReview: lastReview, RemainingSteps: remainingSteps,
	}, nil
}

func loadCard(ctx context.Context, tx *sql.Tx, courseID, moduleID, reviewItemID string) (fsrs.Card, bool, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT due_at, stability, difficulty, scheduled_days, reps, lapses,
			state, last_review_at, remaining_steps
		FROM review_cards
		WHERE course_id = ? AND module_id = ? AND review_item_id = ?
	`, courseID, moduleID, reviewItemID)
	card, err := scanCard(row)
	if errors.Is(err, sql.ErrNoRows) {
		return fsrs.Card{}, false, nil
	}
	if err != nil {
		return fsrs.Card{}, false, err
	}
	return card, true, nil
}

func storeCard(ctx context.Context, tx *sql.Tx, courseID, moduleID, reviewItemID string, card fsrs.Card, now time.Time) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO review_cards (
			course_id, module_id, review_item_id, due_at, stability, difficulty,
			scheduled_days, reps, lapses, state, last_review_at, remaining_steps, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (course_id, module_id, review_item_id) DO UPDATE SET
			due_at = excluded.due_at,
			stability = excluded.stability,
			difficulty = excluded.difficulty,
			scheduled_days = excluded.scheduled_days,
			reps = excluded.reps,
			lapses = excluded.lapses,
			state = excluded.state,
			last_review_at = excluded.last_review_at,
			remaining_steps = excluded.remaining_steps,
			updated_at = excluded.updated_at
	`, courseID, moduleID, reviewItemID, formatTime(card.Due), card.Stability, card.Difficulty,
		card.ScheduledDays, card.Reps, card.Lapses, int(card.State), nullableTime(card.LastReview),
		card.RemainingSteps, formatTime(now))
	if err != nil {
		return fmt.Errorf("write review card: %w", err)
	}
	return nil
}

func insertReviewLog(ctx context.Context, tx *sql.Tx, courseID, moduleID, reviewItemID string, rating Rating, now time.Time, before, after fsrs.Card) (int64, error) {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO review_logs (
			course_id, module_id, review_item_id, reviewed_at, rating, previous_due, next_due,
			before_stability, after_stability, before_difficulty, after_difficulty,
			before_scheduled_days, after_scheduled_days, before_reps, after_reps,
			before_lapses, after_lapses, before_state, after_state,
			before_last_review_at, after_last_review_at,
			before_remaining_steps, after_remaining_steps
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, courseID, moduleID, reviewItemID, formatTime(now), rating,
		formatTime(before.Due), formatTime(after.Due), before.Stability, after.Stability,
		before.Difficulty, after.Difficulty, before.ScheduledDays, after.ScheduledDays,
		before.Reps, after.Reps, before.Lapses, after.Lapses, int(before.State), int(after.State),
		nullableTime(before.LastReview), nullableTime(after.LastReview),
		before.RemainingSteps, after.RemainingSteps)
	if err != nil {
		return 0, fmt.Errorf("write review log: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read review log id: %w", err)
	}
	return id, nil
}

func fsrsRating(rating Rating) (fsrs.Rating, error) {
	switch rating {
	case RatingAgain:
		return fsrs.Again, nil
	case RatingHard:
		return fsrs.Hard, nil
	case RatingGood:
		return fsrs.Good, nil
	case RatingEasy:
		return fsrs.Easy, nil
	default:
		return fsrs.Manual, ErrInvalidRating
	}
}

func parseRequiredTime(label string, value sql.NullString) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, fmt.Errorf("%s is missing", label)
	}
	parsed, err := time.Parse(time.RFC3339Nano, value.String)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse %s: %w", label, err)
	}
	return parsed.UTC(), nil
}

func parseOptionalTime(label string, value sql.NullString) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value.String)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse %s: %w", label, err)
	}
	return parsed.UTC(), nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return formatTime(value)
}
