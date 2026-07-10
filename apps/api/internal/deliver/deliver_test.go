package deliver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDo_SuccessReturnsStatusCodeAndNoError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL, nil)
	statusCode, err := Do(srv.Client(), req, "test", func(resp *http.Response) error {
		t.Fatal("errFn should not be called for a 2xx response")
		return nil
	})
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if statusCode != http.StatusCreated {
		t.Fatalf("statusCode = %d, want 201", statusCode)
	}
}

func TestDo_NonSuccessCallsErrFnWithTheResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL, nil)
	statusCode, err := Do(srv.Client(), req, "test", func(resp *http.Response) error {
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("errFn got status %d, want 500", resp.StatusCode)
		}
		return errCustom
	})
	if err != errCustom {
		t.Fatalf("want errFn's error surfaced, got %v", err)
	}
	if statusCode != http.StatusInternalServerError {
		t.Fatalf("statusCode = %d, want 500", statusCode)
	}
}

func TestDo_ConnectionFailureReturnsZeroStatusAndWrapsChanName(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:0", nil)
	statusCode, err := Do(&http.Client{}, req, "test", func(*http.Response) error {
		t.Fatal("errFn should not be called for a connection failure")
		return nil
	})
	if err == nil {
		t.Fatal("want an error when the request never reaches a server")
	}
	if statusCode != 0 {
		t.Fatalf("statusCode = %d, want 0 for a connection failure", statusCode)
	}
}

var errCustom = &testError{"custom error"}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
