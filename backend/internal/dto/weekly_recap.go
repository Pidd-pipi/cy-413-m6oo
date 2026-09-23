package dto

type WeeklyRecapResponse struct {
	StartDate      string   `json:"start_date"`
	EndDate        string   `json:"end_date"`
	MoodCount      int      `json:"mood_count"`
	AverageMood    *float64 `json:"average_mood"`
	TopTag         string   `json:"top_tag"`
	JournalCount   int64    `json:"journal_count"`
	AssessmentDone int64    `json:"assessment_done"`
	MoodsToThree   int      `json:"moods_to_three"`
}
