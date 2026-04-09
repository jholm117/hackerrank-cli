package api

type Test struct {
	ID        string   `json:"id"`
	UniqueID  string   `json:"unique_id"`
	Name      string   `json:"name"`
	Duration  int      `json:"duration"`
	State     string   `json:"state"`
	Draft     bool     `json:"draft"`
	CreatedAt string   `json:"created_at"`
	Questions []string `json:"questions"`
}

// Candidate is used for list endpoints where questions is a map of scores.
type Candidate struct {
	ID              string             `json:"id"`
	Email           string             `json:"email"`
	FullName        string             `json:"full_name"`
	Score           float64            `json:"score"`
	PercentageScore float64            `json:"percentage_score"`
	Status          int                `json:"status"`
	AttemptStart    string             `json:"attempt_starttime"`
	AttemptEnd      string             `json:"attempt_endtime"`
	AttemptID       string             `json:"attempt_id"`
	Questions       map[string]float64 `json:"questions"`
	PDFURL          string             `json:"pdf_url"`
	ReportURL       string             `json:"report_url"`
}

// CandidateDetail is used with additional_fields where questions has full results.
type CandidateDetail struct {
	ID              string                    `json:"id"`
	Email           string                    `json:"email"`
	FullName        string                    `json:"full_name"`
	Score           float64                   `json:"score"`
	PercentageScore float64                   `json:"percentage_score"`
	Status          int                       `json:"status"`
	AttemptStart    string                    `json:"attempt_starttime"`
	AttemptEnd      string                    `json:"attempt_endtime"`
	AttemptID       string                    `json:"attempt_id"`
	Questions       map[string]QuestionResult `json:"questions"`
	PDFURL          string                    `json:"pdf_url"`
	ReportURL       string                    `json:"report_url"`
}

type QuestionResult struct {
	Answered    bool         `json:"answered"`
	Answer      Answer       `json:"answer"`
	Score       float64      `json:"score"`
	Submissions []Submission `json:"submissions"`
}

type Answer struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

type Submission struct {
	ID        int64          `json:"id"`
	Answer    Answer         `json:"answer"`
	Score     float64        `json:"score"`
	IsValid   bool           `json:"is_valid"`
	CreatedAt string         `json:"created_at"`
	Metadata  SubmissionMeta `json:"metadata"`
}

type SubmissionMeta struct {
	Result         int   `json:"result"`
	TestcaseStatus []int `json:"testcase_status"`
}

// InterviewRecording holds the code recordings from an interview.
type InterviewRecording struct {
	Data struct {
		Questions []InterviewQuestion `json:"questions"`
	} `json:"data"`
}

// InterviewQuestion holds a single question pad from an interview.
type InterviewQuestion struct {
	Question string           `json:"question"`
	QHash    string           `json:"qhash"`
	QType    string           `json:"qtype"`
	Runs     []InterviewRun   `json:"runs"`
}

// InterviewRun holds one code snapshot from an interview pad.
type InterviewRun struct {
	Code  string `json:"code"`
	Lang  string `json:"lang"`
	Input string `json:"input"`
}

type Interview struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	URL       string `json:"url"`
}

type Transcript struct {
	Messages []TranscriptMessage `json:"messages"`
}

type TranscriptMessage struct {
	Author    string `json:"author"`
	Email     string `json:"email"`
	Candidate bool   `json:"candidate"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}
