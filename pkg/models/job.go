package models

type Job struct {
	Token       string `json:"token"`
	Url         string `json:"url"`
	CallbackURL string `json:"callback_url"`
}
