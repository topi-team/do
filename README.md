# do: cleaner error handling

Package do leverages Go 1.18's generics to simplify error handling.

As the Errors are values blogpost says:

```go
> Use the language to simplify your error handling.
>
> But remember: Whatever you do, always check your errors!
```

[https://go.dev/blog/errors-are-values](https://go.dev/blog/errors-are-values)

## Examples

```golang
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/topi-team/do"
)

const maxBodyLength = 10240

type InvalidRequest struct {
	msg  string
	code int
}

func (err InvalidRequest) Error() string {
	return err.msg
}

type User struct {
	Email   string
	IsAdmin bool
}

func testRequest(method, path, body string) (*httptest.ResponseRecorder, *http.Request) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Add("content-type", "application/json")
	return rec, req
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
		r = do.Check(r, requestMatch(req.Method, method, http.StatusNotFound))
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
