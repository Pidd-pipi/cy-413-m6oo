package service

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/blueship581/mindgarden/backend/internal/model"
	"github.com/blueship581/mindgarden/backend/internal/repository"
)

type fakeRecapMoodRepo struct {
	repository.MoodRepository
	moods      []model.Mood
	gotStart   time.Time
	gotEnd     time.Time
	gotUID     uint
	emptyRange bool
}

func (f *fakeRecapMoodRepo) Range(uid uint, start, end time.Time) ([]model.Mood, error) {
	f.gotUID, f.gotStart, f.gotEnd = uid, start, end
	return f.moods, nil
}

type fakeRecapJournalRepo struct {
	repository.JournalRepository
	count int64
}

func (f *fakeRecapJournalRepo) CountRange(uid uint, start, end time.Time) (int64, error) {
	return f.count, nil
}

type fakeRecapAssessmentRepo struct {
	repository.AssessmentRepository
	count int64
}

func (f *fakeRecapAssessmentRepo) CountUserRange(uid uint, start, end time.Time) (int64, error) {
	return f.count, nil
}

func newRecapService(m repository.MoodRepository, jc, ac int64) *RecapService {
	return NewRecapService(m, &fakeRecapJournalRepo{count: jc}, &fakeRecapAssessmentRepo{count: ac}, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func tagsJSON(tags ...string) string {
	b := []byte{'['}
	for i, t := range tags {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, '"')
		b = append(b, t...)
		b = append(b, '"')
	}
	b = append(b, ']')
	return string(b)
}

func moodAt(beijingDate string, level int, tags ...string) model.Mood {
	d, _ := time.ParseInLocation("2006-01-02 15:04", beijingDate+" 12:00", time.FixedZone("CST", 8*3600))
	return model.Mood{RecordDate: d, MoodLevel: level, MoodTags: tagsJSON(tags...)}
}

func TestWeeklyRecapEmpty(t *testing.T) {
	svc := newRecapService(&fakeRecapMoodRepo{}, 0, 0)
	now := time.Date(2026, 9, 23, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
	r, e := svc.WeeklyRecap(7, now)
	if e != nil {
		t.Fatalf("unexpected err: %v", e)
	}
	if r.StartDate != "2026-09-21" || r.EndDate != "2026-09-23" {
		t.Fatalf("range = %s..%s", r.StartDate, r.EndDate)
	}
	if r.MoodCount != 0 || r.JournalCount != 0 || r.AssessmentCount != 0 {
		t.Fatalf("counts should be zero: %+v", r)
	}
	if r.AverageMood != nil {
		t.Fatalf("average_mood should be null on empty data, got %v", *r.AverageMood)
	}
	if r.TopTag != "" {
		t.Fatalf("top_tag should be empty, got %q", r.TopTag)
	}
	if r.MoodsToThree != 3 {
		t.Fatalf("moods_to_three = %d, want 3", r.MoodsToThree)
	}
}

func TestWeeklyRecapAggregationAndTieBreak(t *testing.T) {
	// newest first: calm in the latest record ties with happy at 2 each.
	moods := []model.Mood{
		moodAt("2026-09-23", 8, "calm", "happy"),
		moodAt("2026-09-22", 5, "happy"),
		moodAt("2026-09-21", 6, "calm", "tired"),
		moodAt("2026-09-21", 4, "tired"),
	}
	svc := newRecapService(&fakeRecapMoodRepo{moods: moods}, 2, 1)
	now := time.Date(2026, 9, 23, 22, 0, 0, 0, time.FixedZone("CST", 8*3600))
	r, e := svc.WeeklyRecap(1, now)
	if e != nil {
		t.Fatalf("unexpected err: %v", e)
	}
	if r.MoodCount != 4 {
		t.Fatalf("mood_count = %d", r.MoodCount)
	}
	// (8+5+6+4)/4 = 5.75 -> 5.8
	if r.AverageMood == nil || *r.AverageMood != 5.8 {
		t.Fatalf("average_mood = %v", r.AverageMood)
	}
	// happy and calm both appear twice; calm is in the most recent record first.
	if r.TopTag != "calm" {
		t.Fatalf("top_tag = %q, want calm (latest-record tie break)", r.TopTag)
	}
	if r.JournalCount != 2 || r.AssessmentCount != 1 {
		t.Fatalf("journal/assessment counts = %d/%d", r.JournalCount, r.AssessmentCount)
	}
	if r.MoodsToThree != 0 {
		t.Fatalf("moods_to_three = %d, want 0", r.MoodsToThree)
	}
}

func TestWeeklyRecapUsesBeijingMondayBounds(t *testing.T) {
	repo := &fakeRecapMoodRepo{}
	svc := newRecapService(repo, 0, 0)
	// Sunday 2026-09-20 18:00 UTC == Monday 2026-09-21 02:00 Beijing,
	// so the week is a single day long: Monday to Monday.
	now := time.Date(2026, 9, 20, 18, 0, 0, 0, time.UTC)
	r, e := svc.WeeklyRecap(3, now)
	if e != nil {
		t.Fatalf("unexpected err: %v", e)
	}
	if r.StartDate != "2026-09-21" || r.EndDate != "2026-09-21" {
		t.Fatalf("range = %s..%s, want 2026-09-21..2026-09-21", r.StartDate, r.EndDate)
	}
	if repo.gotUID != 3 {
		t.Fatalf("uid = %d", repo.gotUID)
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	wantStart := time.Date(2026, 9, 21, 0, 0, 0, 0, loc)
	wantEnd := time.Date(2026, 9, 22, 0, 0, 0, 0, loc)
	if !repo.gotStart.Equal(wantStart) || !repo.gotEnd.Equal(wantEnd) {
		t.Fatalf("bounds = %v..%v, want %v..%v", repo.gotStart, repo.gotEnd, wantStart, wantEnd)
	}
	if repo.gotStart.Location().String() != "Asia/Shanghai" {
		t.Fatalf("start location = %s", repo.gotStart.Location())
	}
}

func TestWeeklyRecapSundayBeijingStartsSameWeekMonday(t *testing.T) {
	repo := &fakeRecapMoodRepo{}
	svc := newRecapService(repo, 0, 0)
	// Sunday 2026-09-27 12:00 Beijing (Monday 9/21 is the week start).
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	r, e := svc.WeeklyRecap(1, now)
	if e != nil {
		t.Fatalf("unexpected err: %v", e)
	}
	if r.StartDate != "2026-09-21" || r.EndDate != "2026-09-27" {
		t.Fatalf("range = %s..%s, want 2026-09-21..2026-09-27", r.StartDate, r.EndDate)
	}
}

func TestWeeklyRecapTwoRecordsLeavesOne(t *testing.T) {
	moods := []model.Mood{moodAt("2026-09-22", 3, "tired"), moodAt("2026-09-21", 7, "happy")}
	svc := newRecapService(&fakeRecapMoodRepo{moods: moods}, 0, 0)
	now := time.Date(2026, 9, 23, 9, 0, 0, 0, time.FixedZone("CST", 8*3600))
	r, _ := svc.WeeklyRecap(1, now)
	if r.MoodsToThree != 1 {
		t.Fatalf("moods_to_three = %d, want 1", r.MoodsToThree)
	}
	if r.AverageMood == nil || *r.AverageMood != 5.0 {
		t.Fatalf("average_mood = %v, want 5.0", r.AverageMood)
	}
	if r.TopTag != "tired" {
		t.Fatalf("top_tag = %q, want tired (newest record wins the tie)", r.TopTag)
	}
}
