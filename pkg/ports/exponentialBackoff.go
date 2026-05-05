package ports

type Operation[T any] func() (T, error)

type ExponentialBackoff[T any] interface {
	Retry(operation func() (T, error)) (T, error)
}
