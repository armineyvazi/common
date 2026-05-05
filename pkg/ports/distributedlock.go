package ports

import "context"

type DistributedLock interface {
	Lock() error
	LockContext(ctx context.Context) error
	TryLock() error
	TryLockContext(ctx context.Context) error
	Unlock() (bool, error)
	UnlockContext(ctx context.Context) (bool, error)
}
