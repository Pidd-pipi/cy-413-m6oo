package dto

type WeeklyRecapResponse struct {
	StartDate       string   `json:"start_date"`
	EndDate         string   `json:"end_date"`
	MoodCount       int      `json:"mood_count"`
	AverageMood     *float64 `json:"average_mood"`
	TopTag          string   `json:"top_tag"`
	JournalCount    int      `json:"journal_count"`
	AssessmentCount int      `json:"assessment_count"`
	MoodsToThree    int      `json:"moods_to_three"`
}
