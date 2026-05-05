package ports

type UUID interface {
	GenV4() (string, error)
}
