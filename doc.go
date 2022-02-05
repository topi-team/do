/*
Package do leverages Go 1.18's generics to simplify error handling.

Goals

The intended goals:

 1. Simplify error handling for functions with lots of `if err != nil { return err }`.
 2. Allow decomposing checks and maps in smaller testable functions.
 3. Maintain type safety through the checks.
 4. Package can be used without imposing this style in the rest of your codebase.

Why?

As the "Errors are values" blogpost says:
 > Use the language to simplify your error handling.
 >
 > But remember: Whatever you do, always check your errors!

https://go.dev/blog/errors-are-values

Concepts

At the core of the `do` package is the type `Result[T]`, which encapsulates
either a value of type T or an error.

You can initialise a result with the folloing two functions:

 // WithJust creates a Result encapsulating the given value
 r := do.WithJust("foo") // type of r: Result[string]

 // WithReturn creates a Result from encapsulating a typical Go func which can
 // return an error
 r := do.WithReturn(os.Open("file")) // type of r: Result[*os.File]

You can then work with `Result[T]` using `do.Map`, `do.MapOrErr`, `do.Check`.
The beauty of this pattern comes with the fact that, once one of these
functions catches an error, all subsequent functions will just skip and forward
the error they received.

Here is an example of a hypothetical http.HandlerFunc to create a blogpost.

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

Here is an example of a function that can be used with `MapOrErr` or as a
stand-alone validator:

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

All of this is inspired by other functiona-programming patterns that have
become popular in many languages. Before Go 1.18 supporting these type of
patterns required using interface{} and doing away with type safety. Now that
generics are supported, we can leverage these patterns without sacrificing type
safety.

Caveats

Due to type inference, `WithJust` and `WithReturn` will return a Result with
the concrete type you passed to it. That means you must either use the concrete
type in your functions rather than interfaces or explicitly specify the type of
the result.

As an example, the follwing code wouldn't compile:

 r := do.WithReturn(os.Open("file"))
 limit := do.Map(r, func(r io.Reader) io.Reader { return io.LimitReader(r, 10) })
 // Build error: type func(r io.Reader) io.Reader of func(r io.Reader) io.Reader {…}
 //  does not match inferred type func(*os.File) newT for func(T) newT

Instead, you must manually set the types:

 r := do.WithReturn[io.Reader](os.Open("file"))
 limit := do.Map(r, func(r io.Reader) io.Reader { return io.LimitReader(r, 10) })
*/
package do

// README.md generated with goreadme:
// goreadme -credit=false -title="do: cleaner error handling" -types -methods -functions -factories > README.md"
