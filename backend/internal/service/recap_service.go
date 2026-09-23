package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/blueship581/mindgarden/backend/internal/constants"
	"github.com/blueship581/mindgarden/backend/internal/dto"
	"github.com/blueship581/mindgarden/backend/internal/model"
	"github.com/blueship581/mindgarden/backend/internal/repository"
)

var beijingZone = time.FixedZone("CST", 8*60*60)

const weeklyRecapMoodTarget = 3

type RecapService struct {
	moods       repository.MoodRepository
	journals    repository.JournalRepository
	assessments repository.AssessmentRepository
	logger      *slog.Logger
}

func NewRecapService(m repository.MoodRepository, j repository.JournalRepository, a repository.AssessmentRepository, l *slog.Logger) *RecapService {
	return &RecapService{m, j, a, l}
}

// weeklyBounds returns the Beijing-time Monday 00:00 and tomorrow 00:00
// (i.e. just past today) expressed as absolute time instants, plus the
// Monday and today calendar dates for display.
func weeklyBounds(now time.Time) (from, to time.Time, start, end time.Time) {
	n := now.In(beijingZone)
	end = time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, beijingZone)
	daysSinceMonday := (int(n.Weekday()) + 6) % 7
	start = end.AddDate(0, 0, -daysSinceMonday)
	return start, end.AddDate(0, 0, 1), start, end
}

func roundOneDecimal(v float64) float64 {
	return math.Round(v*10) / 10
}

// topTag counts tag occurrences; ties are resolved in favour of the tag
// carried by the most recent mood record (records arrive newest-first).
func topTag(records []model.Mood) string {
	counts := map[string]int{}
	maxN := 0
	for _, r := range records {
		var tags []string
		if e := json.Unmarshal([]byte(r.MoodTags), &tags); e != nil {
			continue
		}
		for _, t := range tags {
			counts[t]++
			if counts[t] > maxN {
				maxN = counts[t]
			}
		}
	}
	for _, r := range records {
		var tags []string
		if e := json.Unmarshal([]byte(r.MoodTags), &tags); e != nil {
			continue
		}
		for _, t := range tags {
			if counts[t] == maxN {
				return t
			}
		}
	}
	return ""
}

func (s *RecapService) Weekly(uid uint) (*dto.WeeklyRecapResponse, error) {
	from, to, start, end := weeklyBounds(time.Now())
	moods, e := s.moods.Range(uid, from, to)
	if e != nil {
		return nil, fmt.Errorf("Mood[user_id] weekly recap failed: %w", e)
	}
	journals, e := s.journals.CountRange(uid, from, to)
	if e != nil {
		return nil, fmt.Errorf("Journal[user_id] weekly recap failed: %w", e)
	}
	done, e := s.assessments.CountRange(uid, from, to)
	if e != nil {
		return nil, fmt.Errorf("UserAssessment[user_id] weekly recap failed: %w", e)
	}
	res := &dto.WeeklyRecapResponse{
		StartDate:      start.Format("2006-01-02"),
		EndDate:        end.Format("2006-01-02"),
		MoodCount:      len(moods),
		TopTag:         topTag(moods),
		JournalCount:   journals,
		AssessmentDone: done,
	}
	if len(moods) > 0 {
		sum := 0
		for _, m := range moods {
			sum += m.MoodLevel
		}
		avg := roundOneDecimal(float64(sum) / float64(len(moods)))
		res.AverageMood = &avg
	}
	if n := weeklyRecapMoodTarget - len(moods); n > 0 {
		res.MoodsToThree = n
	}
	s.logger.Info(constants.LogWeeklyRecap, "user_id", uid, "start", res.StartDate, "end", res.EndDate)
	return res, nil
}
