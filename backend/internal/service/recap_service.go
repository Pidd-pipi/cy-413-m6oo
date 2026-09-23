package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"
	_ "time/tzdata"

	"github.com/blueship581/mindgarden/backend/internal/constants"
	"github.com/blueship581/mindgarden/backend/internal/dto"
	"github.com/blueship581/mindgarden/backend/internal/model"
	"github.com/blueship581/mindgarden/backend/internal/repository"
)

const (
	moodRecordGoal  = 3
	beijingLocation = "Asia/Shanghai"
)

type RecapService struct {
	moods       repository.MoodRepository
	journals    repository.JournalRepository
	assessments repository.AssessmentRepository
	logger      *slog.Logger
}

func NewRecapService(m repository.MoodRepository, j repository.JournalRepository, a repository.AssessmentRepository, l *slog.Logger) *RecapService {
	return &RecapService{m, j, a, l}
}

// beijingWeek returns [Monday 00:00, tomorrow 00:00) in Beijing time, with both
// bounds expressed as time.Time; the upper bound covers everything up to now.
func beijingWeek(now time.Time) (start, end time.Time, err error) {
	loc, err := time.LoadLocation(beijingLocation)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	today := now.In(loc)
	sinceMonday := (int(today.Weekday()) + 6) % 7
	start = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -sinceMonday)
	end = time.Date(today.Year(), today.Month(), today.Day()+1, 0, 0, 0, 0, loc)
	return
}

func (s *RecapService) WeeklyRecap(uid uint, now time.Time) (*dto.WeeklyRecapResponse, error) {
	start, end, e := beijingWeek(now)
	if e != nil {
		return nil, fmt.Errorf("WeeklyRecap[timezone] range failed: %w", e)
	}
	moods, e := s.moods.Range(uid, start, end)
	if e != nil {
		return nil, fmt.Errorf("Mood[user_id] weekly recap failed: %w", e)
	}
	// newest first, so the tie break reads the most recent record
	sort.SliceStable(moods, func(i, j int) bool {
		if moods[i].RecordDate.Equal(moods[j].RecordDate) {
			return moods[i].ID > moods[j].ID
		}
		return moods[i].RecordDate.After(moods[j].RecordDate)
	})
	journals, e := s.journals.CountRange(uid, start, end)
	if e != nil {
		return nil, fmt.Errorf("Journal[user_id] weekly recap failed: %w", e)
	}
	assessments, e := s.assessments.CountUserRange(uid, start, end)
	if e != nil {
		return nil, fmt.Errorf("UserAssessment[user_id] weekly recap failed: %w", e)
	}
	res := &dto.WeeklyRecapResponse{
		StartDate:       start.Format("2006-01-02"),
		EndDate:         end.AddDate(0, 0, -1).Format("2006-01-02"),
		MoodCount:       len(moods),
		TopTag:          mostCommonTag(moods),
		JournalCount:    int(journals),
		AssessmentCount: int(assessments),
		MoodsToThree:    moodRecordGoal - len(moods),
	}
	if res.MoodsToThree < 0 {
		res.MoodsToThree = 0
	}
	if len(moods) > 0 {
		sum := 0
		for _, m := range moods {
			sum += m.MoodLevel
		}
		avg := math.Round(float64(sum)/float64(len(moods))*10) / 10
		res.AverageMood = &avg
	}
	s.logger.Info(constants.LogWeeklyRecapRead, "user_id", uid, "start_date", res.StartDate, "end_date", res.EndDate)
	return res, nil
}

// mostCommonTag picks the most frequent tag of the week; ties are broken by
// the tag that appears in the most recent mood record (moods are newest first).
func mostCommonTag(moods []model.Mood) string {
	counts := map[string]int{}
	for _, m := range moods {
		var tags []string
		if e := json.Unmarshal([]byte(m.MoodTags), &tags); e != nil {
			continue
		}
		for _, t := range tags {
			counts[t]++
		}
	}
	top := ""
	topCount := 0
	for _, n := range counts {
		if n > topCount {
			topCount = n
		}
	}
	if topCount == 0 {
		return top
	}
	for _, m := range moods {
		var tags []string
		if e := json.Unmarshal([]byte(m.MoodTags), &tags); e != nil {
			continue
		}
		for _, t := range tags {
			if counts[t] == topCount {
				return t
			}
		}
	}
	return top
}
