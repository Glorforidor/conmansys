package handler

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Glorforidor/conmansys/insservice/storage"
)

type serviceMock struct {
	closed bool
}

func (s *serviceMock) GetItems(modules ...storage.Module) ([]*storage.Item, error) {
	if s.closed {
		return nil, errors.New("")
	}

	return nil, nil
}

var (
	items = []*storage.Item{
		{Value: "Taxonomy"},
		{Value: "Management"},
		{Value: "Payment"},
	}

	service = &serviceMock{}
)

func TestResponseJSON(t *testing.T) {
	r := New(service)
	srv := httptest.NewServer(r)
	defer srv.Close()

	tt := map[string]struct {
	}{}
	_ = tt
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
				t.Fatalf("could not create request: %v", err)
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
