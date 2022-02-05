# do: cleaner error handling

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/topi-team/do)

Package do leverages Go 1.18's generics to simplify error handling.

## Goals

The intended goals:

```go
1. Simplify error handling for functions with lots of `if err != nil { return err }`.
2. Allow decomposing checks and maps in smaller testable functions.
3. Maintain type safety through the checks.
4. Package can be used without imposing this style in the rest of your codebase.
```

Why?

As the "Errors are values" blogpost says:

```go
> Use the language to simplify your error handling.
>
> But remember: Whatever you do, always check your errors!
```

[https://go.dev/blog/errors-are-values](https://go.dev/blog/errors-are-values)

## Concepts

At the core of the `do` package is the type `Result[T]`, which encapsulates
either a value of type T or an error.

You can initialise a result with the folloing two functions:

```go
// WithJust creates a Result encapsulating the given value
r := do.WithJust("foo") // type of r: Result[string]

// WithReturn creates a Result from encapsulating a typical Go func which can
// return an error
r := do.WithReturn(os.Open("file")) // type of r: Result[*os.File]
```

You can then work with `Result[T]` using `do.Map`, `do.MapOrErr`, `do.Check`.
The beauty of this pattern comes with the fact that, once one of these
functions catches an error, all subsequent functions will just skip and forward
the error they received.

Here is an example of a hypothetical http.HandlerFunc to create a blogpost.

```go
func PostCreate(rw http.ResponseWriter, req *http.Request) {
	r := do.WithJust(req)
	r = do.Check(r, validRequest("POST", "/posts"))
	r = do.Map(r, authenticatedRequest)
	r = do.Check(r, requiredScope("posts:write"))
	post = do.MapOrErr(r, decode[Post])
	post = do.MapOrErr(post, storePost(req.Context()))

	do.fold(
		post,
		encode[Post](rw),
		encodeError(rw),
	)
}
```

Here is an example of a function that can be used with `MapOrErr` or as a
stand-alone validator:

```go
// notice the function signature makes no reference to this package.
func validRequest(method, path string) func(req *http.Request) error {
	return func(req *http.Request) error {
		r := do.WithJust(req)
		r = do.Check(r, acceptsJSON)
		r = do.Check(r, contentTypeJSON)
		r = do.Check(r, requestMatch(req.URL.Path, path, http.StatusNotFound))
		r = do.Check(r, requestMatch(req.Method, method, http.StatusMethodNotAllowed))
		return r.Err()
	}
}
```

All of this is inspired by other functiona-programming patterns that have
become popular in many languages. Before Go 1.18 supporting these type of
patterns required using interface{} and doing away with type safety. Now that
generics are supported, we can leverage these patterns without sacrificing type
safety.

## Caveats

Due to type inference, `WithJust` and `WithReturn` will return a Result with
the concrete type you passed to it. That means you must either use the concrete
type in your functions rather than interfaces or explicitly specify the type of
the result.

As an example, the follwing code wouldn't compile:

```go
r := do.WithReturn(os.Open("file"))
limit := do.Map(r, func(r io.Reader) io.Reader { return io.LimitReader(r, 10) })
// Build error: type func(r io.Reader) io.Reader of func(r io.Reader) io.Reader {…}
//  does not match inferred type func(*os.File) newT for func(T) newT
```

Instead, you must manually set the types:

```go
r := do.WithReturn[io.Reader](os.Open("file"))
limit := do.Map(r, func(r io.Reader) io.Reader { return io.LimitReader(r, 10) })
```

## Functions

### func [Fold](https://github.com/topi-team/do/blob/main/result.go#L114)

`func Fold[T any](input Result[T], okFn func(T), errFn func(error))`

Fold takes an input Result and will call either okFn or errFn depending on
whether input.IsError(). Fold is convenient in functions that don't return
anything such as http.HandlerFunc.

It's a short-hand for the following code:

```go
if input.IsError() {
  errFn(input.Err())
} else {
  okFn(input.Val())
}
```

```golang
echoUser := func(rw http.ResponseWriter, r *http.Request) {
    req := do.WithJust(r)
    req = do.Check(req, validRequest("POST", "/echo"))
    body := do.Map(req, bodyWithLimit)
    user := do.MapOrErr(body, decodeUser)

    do.Fold(
        user,
        encode[User](rw),
        encodeError(rw),
    )
}

func() {
    rec, req := testRequest("POST", "/echo", "invalid")
    echoUser(rec, req)
    fmt.Printf("Status: %d\nBody: %s\n", rec.Result().StatusCode, rec.Body.String())
}()

func() {
    rec, req := testRequest("POST", "/echo", `{"email":"ernesto@topi.eu"}`)
    echoUser(rec, req)
    fmt.Printf("Status: %d\nBody: %s\n", rec.Result().StatusCode, rec.Body.String())
}()
```

 Output:

```
Status: 400
Body: invalid character 'i' looking for beginning of value

Status: 200
Body: {"Email":"ernesto@topi.eu","IsAdmin":false}
```

## Types

### type [Result](https://github.com/topi-team/do/blob/main/result.go#L4)

`type Result[T any] struct { ... }`

Result encapsulates either a value of type T or an error.

#### func [Check](https://github.com/topi-team/do/blob/main/result.go#L126)

`func Check[T any](input Result[T], checkFn func(T) error) Result[T]`

Check will call the checkFn with the given value as long as the input result
is not an error.

The returned Result will contain the error returned by checkFn.

#### func [Map](https://github.com/topi-team/do/blob/main/result.go#L75)

`func Map[T, newT any](input Result[T], mapFn func(T) newT) Result[newT]`

Map will call the given mapFn with the input Result and return a new Result
with the value returned by it.

When input.IsError(), Map returns a result with input.Err() without calling
the mapping function.

#### func [MapOrErr](https://github.com/topi-team/do/blob/main/result.go#L90)

`func MapOrErr[T, newT any](input Result[T], mapFn func(T) (newT, error)) Result[newT]`

MapOrErr is equivalent to Map, but mapFn can return an error.

When mapFn returns an error, the returning Result will include that error.

#### func [WithErrHandler](https://github.com/topi-team/do/blob/main/result.go#L138)

`func WithErrHandler[T any](input Result[T], wrapFn func(error) error) Result[T]`

WithErrHandler returns a Result that will call wrapFn when an error is later
stored in the result.

Subsequent calls to WithErrHandler will replace the existing handler.

#### func [WithJust](https://github.com/topi-team/do/blob/main/result.go#L33)

`func WithJust[T any](val T) Result[T]`

WithJust initialises a Result the given value.

#### func [WithReturn](https://github.com/topi-team/do/blob/main/result.go#L25)

`func WithReturn[T any](val T, err error) Result[T]`

WithReturn is a short-hand to create a Result wrapping a function that
returns a value and an error.

Example:

```go
r := do.WithReturn(os.Open("file"))
```

#### func (Result[T]) [Err](https://github.com/topi-team/do/blob/main/result.go#L46)

`func (r Result[T]) Err() error`

Err returns the encapsulated error. It returns nil if the Result
is not an error.

#### func (Result[T]) [IsError](https://github.com/topi-team/do/blob/main/result.go#L40)

`func (r Result[T]) IsError() bool`

IsError returns true if the Result contains an error.

#### func (Result[T]) [Return](https://github.com/topi-team/do/blob/main/result.go#L66)

`func (r Result[T]) Return() (T, error)`

Return is a convenient short-hand to get both the value and error to return.

Example:

```go
func(file string) (io.Reader, error) {
  r := do.WithReturn(os.Open(file))
  // ...
  return r.Return()
}
```

#### func (Result[T]) [Val](https://github.com/topi-team/do/blob/main/result.go#L51)

`func (r Result[T]) Val() T`

Val returns the wrapped value. Val will panic when result is an error.

## Examples

```golang
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/topi-team/do"
)

type User struct {
	Email   string
	IsAdmin bool
}

func main() {
	echoUser := func(rw http.ResponseWriter, r *http.Request) {
		req := do.WithJust(r)
		req = do.Check(req, validRequest("POST", "/echo"))
		body := do.Map(req, bodyWithLimit)
		user := do.MapOrErr(body, decodeUser)

		do.Fold(
			user,
			encode[User](rw),
			encodeError(rw),
		)
	}

	func() {
		rec, req := testRequest("POST", "/echo", "invalid")
		echoUser(rec, req)
		fmt.Printf("Status: %d\nBody: %s\n", rec.Result().StatusCode, rec.Body.String())
	}()

	func() {
		rec, req := testRequest("POST", "/echo", `{"email":"ernesto@topi.eu"}`)
		echoUser(rec, req)
		fmt.Printf("Status: %d\nBody: %s\n", rec.Result().StatusCode, rec.Body.String())
	}()
}

func encodeError(rw http.ResponseWriter) func(err error) {
	return func(err error) {
		switch err := err.(type) {
		case InvalidRequest:
			http.Error(rw, err.Error(), err.code)
		default:
			log.Println(err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	}
}

func encode[T any](rw http.ResponseWriter) func(v T) {
	return func(v T) {
		err := json.NewEncoder(rw).Encode(v)
		if err != nil {
			log.Println(err)
		}
	}
}

func decodeUser(r io.Reader) (User, error) {
	var u User
	err := json.NewDecoder(r).Decode(&u)
	if err != nil {
		return u, InvalidRequest{
			msg:  err.Error(),
			code: http.StatusBadRequest,
		}
	}
	return u, nil
}

func bodyWithLimit(r *http.Request) io.Reader {
	return io.LimitReader(r.Body, maxBodyLength)
}

func validRequest(method, path string) func(req *http.Request) error {
	return func(req *http.Request) error {
		r := do.WithJust(req)
		r = do.Check(r, acceptsJSON)
		r = do.Check(r, contentTypeJSON)
		r = do.Check(r, requestMatch(req.URL.Path, path, http.StatusNotFound))
		r = do.Check(r, requestMatch(req.Method, method, http.StatusMethodNotAllowed))
		return r.Err()
	}
}

func requestMatch(got, want string, code int) func(*http.Request) error {
	return func(*http.Request) error {
		if got != want {
			return InvalidRequest{"invalid request", code}
		}
		return nil
	}
}

func acceptsJSON(r *http.Request) error {
	accepts := r.Header.Get("accepts")
	if accepts != "" && accepts != "application/json" {
		return InvalidRequest{
			msg:  fmt.Sprintf(`expected "application/json" Accepts header. Got: %s`, accepts),
			code: http.StatusBadRequest,
		}
	}
	return nil
}

func contentTypeJSON(r *http.Request) error {
	accepts := r.Header.Get("content-type")
	if accepts != "" && accepts != "application/json" {
		return InvalidRequest{
			msg:  fmt.Sprintf(`expected "application/json" Content-Type header. Got: %s`, accepts),
			code: http.StatusBadRequest,
		}
	}
	return nil
}

```

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
