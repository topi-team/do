package do_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/topi-team/do"
)

func TestWithReturn(t *testing.T) {
	t.Run("Without Error", func(t *testing.T) {
		r := do.WithReturn(testStub("example"))
		val := r.Val()
		require.Equal(t, "example", val.String())
		require.NoError(t, r.Err())
	})

	t.Run("With Error", func(t *testing.T) {
		r := do.WithReturn(testStub("fail"))
		require.Error(t, r.Err())
	})

	t.Run("Specifying interface", func(t *testing.T) {
		r := do.WithReturn[io.Reader](testStub("example"))
		b := do.MapOrErr(r, func(r io.Reader) ([]byte, error) {
			return ioutil.ReadAll(r)
		})
		var res []byte = b.Val()
		require.Equal(t, "example", string(res))
	})
}

func TestWithJust(t *testing.T) {
	r := do.WithJust("example")
	val := r.Val()
	require.Equal(t, "example", val)
	require.NoError(t, r.Err())
}

func TestReturn(t *testing.T) {
	t.Run("With Value", func(t *testing.T) {
		r := do.WithReturn(testStub("example"))
		val, err := r.Return()
		require.Equal(t, "example", val.String())
		require.NoError(t, err)
	})
}

func TestMap(t *testing.T) {
	t.Run("To same type", func(t *testing.T) {
		r := do.WithJust("example")
		upper := do.Map(r, func(s string) string { return strings.ToUpper(s) })
		require.Equal(t, "EXAMPLE", upper.Val())
	})

	t.Run("To another type", func(t *testing.T) {
		r := do.WithJust("example")
		reader := do.Map(r, func(s string) io.Reader { return bytes.NewBufferString(s) })
		limit := do.Map(reader, func(r io.Reader) io.Reader { return io.LimitReader(r, 2) })
		read := do.Map(limit, func(r io.Reader) string {
			b, _ := ioutil.ReadAll(r)
			return string(b)
		})
		require.Equal(t, "ex", read.Val())
	})

	t.Run("func not called when result is an error", func(t *testing.T) {
		r := do.WithReturn[io.Reader](testStub("fail"))
		r = do.Map(r, func(io.Reader) io.Reader { panic("unexpected call") })
		require.True(t, r.IsError())
	})
}

func TestIsError(t *testing.T) {
	t.Run("Without Error", func(t *testing.T) {
		r := do.WithReturn(testStub("example"))
		require.False(t, r.IsError())
	})

	t.Run("With Error", func(t *testing.T) {
		r := do.WithReturn(testStub("fail"))
		require.True(t, r.IsError())
	})
}

func TestMapOrErr(t *testing.T) {
	t.Run("Successful", func(t *testing.T) {
		r := do.WithJust("example")
		reader := do.MapOrErr(r, func(s string) (io.Reader, error) { return testStub(s) })
		limit := do.Map(reader, func(r io.Reader) io.Reader { return io.LimitReader(r, 2) })
		bytes := do.MapOrErr(limit, func(r io.Reader) ([]byte, error) {
			return ioutil.ReadAll(r)
		})
		require.Equal(t, "ex", string(bytes.Val()))
	})

	t.Run("Sets error", func(t *testing.T) {
		r := do.WithJust("fail")
		reader := do.MapOrErr(r, func(s string) (io.Reader, error) { return testStub(s) })
		limit := do.Map(reader, func(r io.Reader) io.Reader { panic("unexpected call") })
		bytes := do.MapOrErr(limit, func(r io.Reader) ([]byte, error) { panic("unexpected call") })
		require.True(t, bytes.IsError())
	})

	t.Run("func not called when result is an error", func(t *testing.T) {
		r := do.WithReturn[io.Reader](testStub("fail"))
		r = do.MapOrErr(r, func(io.Reader) (io.Reader, error) { panic("unexpected call") })
		require.True(t, r.IsError())
	})
}

func TestFold(t *testing.T) {
	t.Run("Without Error", func(t *testing.T) {
		r := do.WithJust("example")
		var calls int
		do.Fold(
			r,
			func(s string) {
				calls++
				require.Equal(t, "example", s)
			},
			func(error) { panic("unexpected call") },
		)
		require.Equal(t, 1, calls)
	})

	t.Run("With Error", func(t *testing.T) {
		r := do.WithReturn(testStub("fail"))
		var calls int
		do.Fold(
			r,
			func(*bytes.Buffer) {
				panic("unexpected call")
			},
			func(err error) {
				calls++
				require.Error(t, err)
			},
		)
		require.Equal(t, 1, calls)
	})
}

func ExampleFold() {
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
	// Output:
	// Status: 400
	// Body: invalid character 'i' looking for beginning of value
	//
	// Status: 200
	// Body: {"Email":"ernesto@topi.eu","IsAdmin":false}
}

func TestCheck(t *testing.T) {
	t.Run("Successful check", func(t *testing.T) {
		r := do.WithJust("example")
		r = do.Check(r, func(s string) error {
			require.Equal(t, "example", s)
			return nil
		})
		require.False(t, r.IsError())
		require.Equal(t, "example", r.Val())
	})

	t.Run("Function skipped if result is error", func(t *testing.T) {
		r := do.WithReturn(testStub("fail"))
		r = do.Check(r, func(*bytes.Buffer) error {
			panic("unexpected call")
		})
		require.True(t, r.IsError())
	})

	t.Run("Failing check", func(t *testing.T) {
		r := do.WithJust("example")
		r = do.Check(r, func(s string) error {
			require.Equal(t, "example", s)
			return errors.New("fail")
		})
		require.True(t, r.IsError())
	})
}

func TestVal(t *testing.T) {
	t.Run("Panics with error", func(t *testing.T) {
		r := do.WithReturn(testStub("fail"))
		require.Panics(t, func() {
			r.Val()
		})
	})
}

func TestWithErrHandler(t *testing.T) {
	t.Run("Wrap MapOrErr error", func(t *testing.T) {
		var calls int
		r := do.WithJust(true)
		r = do.WithErrHandler(r, func(err error) error {
			calls++
			return fmt.Errorf("wrapped: %w", err)
		})
		res := do.MapOrErr(r, func(bool) (io.Reader, error) {
			return testStub("fail")
		})
		require.Equal(t, 1, calls)
		require.True(t, res.IsError())
		require.True(t, strings.HasPrefix(res.Err().Error(), "wrapped: "))
	})

	t.Run("Wrap Check error", func(t *testing.T) {
		var calls int
		r := do.WithJust(true)
		r = do.WithErrHandler(r, func(err error) error {
			calls++
			return fmt.Errorf("wrapped: %w", err)
		})
		r = do.Check(r, func(bool) error {
			return errors.New("fail")
		})
		require.Equal(t, 1, calls)
		require.True(t, r.IsError())
		require.Equal(t, r.Err().Error(), "wrapped: fail")
	})

	t.Run("Wrap only once", func(t *testing.T) {
		var calls int
		r := do.WithJust(true)
		r = do.WithErrHandler(r, func(err error) error {
			calls++
			return fmt.Errorf("wrapped: %w", err)
		})
		r = do.Check(r, func(bool) error { return errors.New("fail") })
		r = do.Check(r, func(bool) error { return errors.New("fail") })
		res := do.MapOrErr(r, func(bool) (io.Reader, error) { return testStub("fail") })
		res = do.MapOrErr(res, func(io.Reader) (io.Reader, error) { return testStub("fail") })
		require.Equal(t, 1, calls)
		require.True(t, res.IsError())
		require.Equal(t, res.Err().Error(), "wrapped: fail")
	})

	t.Run("Handler passed on by Map, MapOrErr, and Check", func(t *testing.T) {
		var calls int
		r := do.WithJust(true)
		r = do.WithErrHandler(r, func(err error) error {
			calls++
			return fmt.Errorf("wrapped: %w", err)
		})
		r = do.Check(r, func(bool) error { return nil })
		res := do.MapOrErr(r, func(bool) (io.Reader, error) { return testStub("example") })
		res = do.Map(res, func(r io.Reader) io.Reader { return io.LimitReader(r, 10) })
		res = do.Check(res, func(io.Reader) error { return errors.New("fail") })
		require.Equal(t, 1, calls)
		require.True(t, res.IsError())
		require.Equal(t, res.Err().Error(), "wrapped: fail")
	})
}

func testStub(s string) (*bytes.Buffer, error) {
	if s == "fail" {
		return nil, errors.New("fail")
	}
	return bytes.NewBufferString(s), nil
}
