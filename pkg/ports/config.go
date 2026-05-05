package ports

type Config[T any] interface {
	GetConfig() *T
}
