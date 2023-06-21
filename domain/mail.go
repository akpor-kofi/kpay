package domain

type Mail struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
	To      string `json:"to"`
}
