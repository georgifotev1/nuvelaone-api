package mailer

type Mailer interface {
	Send(email EmailData) error
}

type EmailData struct {
	From     string
	To       []string
	Subject  string
	Template string
	Data     any
}
