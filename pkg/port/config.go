package port

type Config[T any] interface {
	GetConfig() *T
}
