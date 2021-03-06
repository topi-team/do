package do

// Result encapsulates either a value of type T or an error.
type Result[T any] struct {
	val T
	err error
	errHandler func(error) error
}

func (r Result[T]) handleErr(err error) error {
	if err == nil {
		return nil
	}
	if r.errHandler == nil {
		return err
	}
	return r.errHandler(err)
}

// WithReturn is a short-hand to create a Result wrapping a function that
// returns a value and an error.
//
// Example:
//  r := do.WithReturn(os.Open("file"))
func WithReturn[T any](val T, err error) Result[T] {
	return Result[T]{
		val: val,
		err: err,
	}
}

// WithJust initialises a Result the given value.
func WithJust[T any](val T) Result[T] {
	return Result[T]{
		val: val,
	}
}

// IsError returns true if the Result contains an error.
func (r Result[T]) IsError() bool {
	return r.err != nil
}

// Err returns the encapsulated error. It returns nil if the Result
// is not an error.
func (r Result[T]) Err() error {
	return r.err
}

// Val returns the wrapped value. Val will panic when result is an error.
func (r Result[T]) Val() T {
	if r.IsError() {
		panic("called Val for result that IsError")
	}
	return r.val
}

// Return is a convenient short-hand to get both the value and error to return.
//
// Example:
//  func(file string) (io.Reader, error) {
//    r := do.WithReturn(os.Open(file))
//    // ...
//    return r.Return()
//  }
func (r Result[T]) Return() (T, error) {
	return r.val, r.err
}

// Map will call the given mapFn with the input Result and return a new Result
// with the value returned by it.
//
// When input.IsError(), Map returns a result with input.Err() without calling
// the mapping function.
func Map[T, newT any](input Result[T], mapFn func(T) newT) Result[newT] {
	if input.IsError() {
		return Result[newT]{
			err: input.err,
		}
	}
	return Result[newT]{
		val: mapFn(input.val),
		errHandler: input.errHandler,
	}
}

// MapOrErr is equivalent to Map, but mapFn can return an error.
//
// When mapFn returns an error, the returning Result will include that error.
func MapOrErr[T, newT any](input Result[T], mapFn func(T) (newT, error)) Result[newT] {
	if input.IsError() {
		return Result[newT]{
			err: input.err,
		}
	}
	newVal, newErr := mapFn(input.val)
	return Result[newT]{
		val: newVal,
		err: input.handleErr(newErr),
		errHandler: input.errHandler,
	}
}

// Fold takes an input Result and will call either okFn or errFn depending on
// whether input.IsError(). Fold is convenient in functions that don't return
// anything such as http.HandlerFunc.
//
// It's a short-hand for the following code:
//  if input.IsError() {
//    errFn(input.Err())
//  } else {
//    okFn(input.Val())
//  }
func Fold[T any](input Result[T], okFn func(T), errFn func(error)) {
	if input.IsError() {
		errFn(input.err)
	} else {
		okFn(input.val)
	}
}

// Check will call the checkFn with the given value as long as the input result
// is not an error.
//
// The returned Result will contain the error returned by checkFn.
func Check[T any](input Result[T], checkFn func(T) error) Result[T] {
	if input.IsError() {
		return input
	}
	input.err = input.handleErr(checkFn(input.val))
	return input
}

// WithErrHandler returns a Result that will call wrapFn when an error is later
// stored in the result.
//
// Subsequent calls to WithErrHandler will replace the existing handler.
func WithErrHandler[T any](input Result[T], wrapFn func(error) error) Result[T] {
	return Result[T]{
		val: input.val,
		err: input.err,
		errHandler: wrapFn,
	}
}
