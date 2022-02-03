package do

type Result[T any] struct {
	val T
	err error
}

func Check[T any](v T, err error) Result[T] {
	return Result[T]{
		val: v,
		err: err,
	}
}

func (r Result[T]) IsError() bool {
	return r.err != nil
}

func (r Result[T]) Err() error {
	return r.err
}

func (r Result[T]) Val() T {
	return r.val
}

func (r Result[T]) Return() (T, error) {
	return r.val, r.err
}

func Map[T, newT any](r Result[T], fn func(T) newT) Result[newT] {
	if r.IsError() {
		return Result[newT]{
			err: r.err,
		}
	}
	return Result[newT]{
		val: fn(r.val),
	}
}

func Fold[T any](r Result[T], okFn func(T), errFn func(error)) {
	if r.IsError() {
		errFn(r.err)
	} else {
		okFn(r.val)
	}
}

func MapErr[T, newT any](r Result[T], fn func(T) (newT, error)) Result[newT] {
	if r.IsError() {
		return Result[newT]{
			err: r.err,
		}
	}
	newVal, newErr := fn(r.val)
	return Result[newT]{
		val: newVal,
		err: newErr,
	}
}
