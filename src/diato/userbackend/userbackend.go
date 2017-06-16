package userbackend

type Userbackend interface {
	GetServerForUser(string) (string, uint32, error)
}
