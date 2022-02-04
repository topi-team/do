// Package do leverages Go 1.18's generics to simplify error handling.
//
// As the Errors are values blogpost says:
//  > Use the language to simplify your error handling.
//  >
//  > But remember: Whatever you do, always check your errors!
//
// https://go.dev/blog/errors-are-values
package do

// Result encapsulates either a value of type T or an error.
type Result[T any] struct {
	val T
	err error
}

// WithReturn is a short-hand to create a Result wrapping a function that
// returns a value and an error.
//
// Example: r := do.WithReturn(os.Open("file"))
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

// Err returns the encapsulated error. It returns nil if the Result is an
// error.
func (r Result[T]) Err() error {
	return r.err
}

// Val returns the wrapped value. The returned value will be invalid if the
// Result contains an error.
func (r Result[T]) Val() T {
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
	}
}

// Fold takes an input Result and will call either okFn or errFn depending on
// whether input.IsError().
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
		err: newErr,
	}
}
