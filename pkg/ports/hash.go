package ports

type Hash interface {
	Encode(id int64) (string, error)
	Decode(code string) (int64, error)
}
