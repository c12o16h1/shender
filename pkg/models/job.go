package models

const (
	JobOk     uint8 = 0
	JobFailed uint8 = 1
)

type Job struct {
	Token       string `json:"token"`
	Url         string `json:"url"`
	CallbackURL string `json:"callback_url"`
}

type JobResult struct {
	Job
	HTML string
	Status uint8
}
