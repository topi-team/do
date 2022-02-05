package do_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
)

const maxBodyLength = 10240

type InvalidRequest struct {
	msg  string
	code int
}

func (err InvalidRequest) Error() string {
	return err.msg
}

func testRequest(method, path, body string) (*httptest.ResponseRecorder, *http.Request) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Add("content-type", "application/json")
	return rec, req
}
