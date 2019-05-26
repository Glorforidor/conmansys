package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Glorforidor/conmansys/insservice/storage"
)

type serviceMock struct {
	items  []*storage.Item
	closed bool
}

func (s *serviceMock) GetItems(modules ...storage.Module) ([]*storage.Item, error) {
	if s.closed {
		return nil, errors.New("")
	}

	return s.items, nil
}

var (
	items = []*storage.Item{
		{Value: "Taxonomy"},
		{Value: "Management"},
		{Value: "Payment"},
	}

	service = &serviceMock{items: items}
)

func TestResponseJSON(t *testing.T) {
	r := New(service)
	srv := httptest.NewServer(r)
	defer srv.Close()

	url := fmt.Sprintf("%v/insfile", srv.URL)

	tt := map[string]struct {
		data   []byte
		status int
	}{
		"success": {
			data:   []byte("[{\"id\": 1}]\r\n"),
			status: http.StatusOK,
		},
		"fail": {
			data:   []byte("[{\"d\": 1}]\r\n"),
			status: http.StatusBadRequest,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			req, _ := http.NewRequest(
				http.MethodGet,
				url,
				bytes.NewReader(tc.data),
			)
			resp, err := srv.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.status {
				t.Fatalf("expected status: %v, got: %v", http.StatusOK, resp.StatusCode)
			}

			ct := resp.Header.Get("Content-Type")
			if ct != "application/json" {
				t.Fatalf("expected Content-Type: application/json, got: %v", ct)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("expected to be able to read body: %v", err)
			}

			if !json.Valid(body) {
				t.Fatalf("expected valid JSON encoding, got: %v", string(body))
			}

			m := make(map[string]interface{})
			if err := json.Unmarshal(body, &m); err != nil {
				t.Fatalf("%v", err)
			}

			data, ok := m["data"]
			if !ok {
				t.Fatal("response body is missing data field")
			}

			switch v := data.(type) {
			case string:
			case interface{}:
			default:
				t.Fatalf("data was of unexpected type: %v", v)
			}
		})
	}
}

func TestResponseText(t *testing.T) {
	r := New(service)
	srv := httptest.NewServer(r)
	defer srv.Close()

	url := fmt.Sprintf("%v/insfile/text", srv.URL)

	tt := map[string]struct {
		data   []byte
		status int
	}{
		"success": {
			data:   []byte("[{\"id\": 1}]\r\n"), // correct module format
			status: http.StatusOK,
		},
		"fail": {
			data:   []byte("[{\"d\": 1}]\r\n"), // incorrect module format
			status: http.StatusBadRequest,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			req, _ := http.NewRequest(
				http.MethodGet,
				url,
				bytes.NewReader(tc.data),
			)
			resp, err := srv.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != tc.status {
				t.Fatalf("expected status: %v, got: %v", tc.status, resp.StatusCode)
			}

			ct := resp.Header.Get("Content-Type")
			if ct != "plain/text" {
				t.Fatalf("expected Content-Type: application/json, got: %v", ct)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("expected to be able to read body: %v", err)
			}

			s := string(body)
			if tc.status != http.StatusBadRequest {
				for _, item := range items {
					if !strings.Contains(s, item.Value) {
						t.Fatalf("missing value: %v in text body: %v", item.Value, s)
					}
				}
			}

			// other status than OK should result in a body with an error
			// message
			if s == "" {
				t.Fatalf("expected an error message in body")
			}
		})
	}
}

func TestInsfile(t *testing.T) {

	tt := map[string]struct {
		body   io.Reader
		status int
		err    bool
		closed bool
	}{
		"Success": {
			body:   bytes.NewReader([]byte("[{\"id\": 1}]\r\n")),
			status: http.StatusOK,
		},
		"wrong JSON format": {
			body:   bytes.NewReader([]byte("{\"id\": 1}\r\n")),
			status: http.StatusBadRequest,
			err:    true,
		},
		"wrong JSON field format": {
			body:   bytes.NewReader([]byte("[{\"i\": 1}]\r\n")),
			status: http.StatusBadRequest,
			err:    true,
		},
		"closed storage": {
			body:   bytes.NewReader([]byte("[{\"id\": 1}]\r\n")),
			status: http.StatusInternalServerError,
			err:    true,
			closed: true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/", tc.body)
			if err != nil {
				t.Fatalf("could not create GET request: %v", err)
			}

			h := handler{service}
			if tc.closed {
				service.closed = true
				defer func() { service.closed = false }()
			}
			_, status, err := h.insfile(req)
			if err != nil {
				if !tc.err {
					t.Errorf("expected no error, got: %v", err)
				}
			}

			if status != tc.status {
				t.Errorf("expected status: %v, got: %v", tc.status, status)
			}
		})
	}
}
