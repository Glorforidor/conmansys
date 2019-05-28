package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Glorforidor/conmansys/insservice/storage"
	"github.com/gorilla/mux"
)

type handler struct {
	storage storage.Service
}

// New registers the service to the handler and registers the "/insfile"
// endpoint to the handler.
func New(service storage.Service) http.Handler {
	r := mux.NewRouter()

	h := handler{service}

	r.HandleFunc("/health", health)
	r.HandleFunc("/insfile", responseJSONWithModules(h.insfileWithModules)).Methods(http.MethodPost)
	r.HandleFunc("/insfile/text", responseTextWithModules(h.insfileWithModules)).Methods(http.MethodPost)
	r.HandleFunc("/insfile/traverse", responseJSON(h.insfile)).Methods(http.MethodPost)
	r.HandleFunc("/insfile/traverse/text", responseText(h.insfile)).Methods(http.MethodPost)

	return r
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// response is the response body for the client. Items and Modules should be set
// to an empty slice instead of nil slice. An empty error should be stored as
// nil
type response struct {
	Items   []*storage.Item   `json:"items"`
	Modules []*storage.Module `json:"modules"`
	Error   *string           `json:"error"`
}

// responseText packs data into text/plain format.
func responseText(h func(r *http.Request) (interface{}, int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, status, err := h(r)
		if err != nil {
			data = err.Error()
		}

		var b bytes.Buffer
		switch v := data.(type) {
		case []*storage.Item:
			for _, item := range v {
				// https://www.ietf.org/rfc/rfc2046.txt says that the newline of
				// text is CRLF
				fmt.Fprintf(&b, "%v\r\n", item.Value)
			}
		default:
			fmt.Fprintf(&b, "%v\r\n", data)
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(status)
		w.Write(b.Bytes())
	}
}

// responseJSON packs data in application/json format.
func responseJSON(h func(r *http.Request) (interface{}, int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp response
		data, status, err := h(r)
		if err != nil {
			v := err.Error()
			resp.Error = &v
		}

		// lets make sure we have an empty collection otherwise json will make
		// it null
		resp.Items = []*storage.Item{}
		resp.Modules = []*storage.Module{}

		switch v := data.(type) {
		case []*storage.Item:
			resp.Items = v
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Printf("could not decode to output: %v", err)
		}
	}
}

func readModules(r io.Reader) ([]storage.Module, int, error) {
	var modules []storage.Module
	t, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("could not read request body: %v", err)
	}

	err = json.Unmarshal(t, &modules)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("could not decode JSON request body: %v", err)
	}

	for _, mod := range modules {
		if mod.ID == 0 {
			return nil, http.StatusBadRequest, fmt.Errorf("one or more values are incorrect: %v", string(t))
		}
	}

	return modules, 0, nil
}

func (h handler) insfile(r *http.Request) (interface{}, int, error) {
	modules, status, err := readModules(r.Body)
	if err != nil {
		return nil, status, err
	}

	items, err := h.storage.GetItems(modules...)
	if err != nil {
		log.Println(fmt.Errorf("could not retrieve data from database: %v", err))
		return nil, http.StatusInternalServerError, fmt.Errorf("Ups something went wrong")
	}

	return items, http.StatusOK, nil
}

func responseTextWithModules(h func(r *http.Request) ([]interface{}, int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, status, err := h(r)

		var b bytes.Buffer

		if err != nil {
			fmt.Fprintf(&b, "%v\r\n", err.Error())
		} else {
			sep := strings.Repeat("-", 20) + "\r\n"
			for _, d := range data {
				switch v := d.(type) {
				case []*storage.Item:
					fmt.Fprint(&b, "items\r\n")
					fmt.Fprint(&b, sep)
					for _, item := range v {
						// https://www.ietf.org/rfc/rfc2046.txt says that the newline of
						// text is CRLF
						fmt.Fprintf(&b, "%v\r\n", item.Value)
					}
					fmt.Fprint(&b, sep)
				case []*storage.Module:
					fmt.Fprint(&b, "modules\r\n")
					fmt.Fprint(&b, sep)
					for _, mod := range v {
						// https://www.ietf.org/rfc/rfc2046.txt says that the newline of
						// text is CRLF
						fmt.Fprintf(&b, "%v\r\n", mod.ID)
					}
					fmt.Fprint(&b, sep)
				}
			}
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(status)
		w.Write(b.Bytes())
	}
}

func responseJSONWithModules(h func(r *http.Request) ([]interface{}, int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp response
		data, status, err := h(r)
		if err != nil {
			v := err.Error()
			resp.Error = &v
		}

		resp.Items = []*storage.Item{}
		resp.Modules = []*storage.Module{}

		for _, d := range data {
			switch v := d.(type) {
			case []*storage.Item:
				resp.Items = v
			case []*storage.Module:
				resp.Modules = v
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(resp)
	}
}

func (h handler) insfileWithModules(r *http.Request) ([]interface{}, int, error) {
	modules, status, err := readModules(r.Body)
	if err != nil {
		return nil, status, err
	}

	items, mods, err := h.storage.GetItemsAndModules(modules...)

	if err != nil {
		log.Println(fmt.Errorf("could not retrieve data from database: %v", err))
		return nil, http.StatusInternalServerError, fmt.Errorf("Ups something went wrong")
	}

	return []interface{}{items, mods}, http.StatusOK, nil
}
