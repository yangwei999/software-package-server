package email

type Email interface {
	Send(subject, content string) error
}
