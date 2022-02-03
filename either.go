package is

type Either[T any] struct {
	val T
	err error
}

func Check[T any](v T, err error) Either[T] {
	return Either[T]{
		val: v,
		err: err,
	}
}

func (e Either[T]) IsError() bool {
	return e.err != nil
}

func (e Either[T]) Err() error {
	return e.err
}

func (e Either[T]) Val() T {
	return e.val
}

func (e Either[T]) Fold() (T, error) {
	return e.val, e.err
}

func Map[T, newT any](e Either[T], fn func(T) newT) Either[newT] {
	if e.IsError() {
		return Either[newT]{
			err: e.err,
		}
	}
	return Either[newT]{
		val: fn(e.val),
	}
}

func MapErr[T, newT any](e Either[T], fn func(T) (newT, error)) Either[newT] {
	if e.IsError() {
		return Either[newT]{
			err: e.err,
		}
	}
	newVal, newErr := fn(e.val)
	return Either[newT]{
		val: newVal,
		err: newErr,
	}
}
