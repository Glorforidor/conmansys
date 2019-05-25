package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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

	r.HandleFunc("/insfile", responseJSON(h.insfile)).Methods(http.MethodGet) // default json
	r.HandleFunc("/insfile/text", responseText(h.insfile)).Methods(http.MethodGet)

	return r
}

// responseText packs data into plain/text format.
func responseText(h func(r *http.Request) (interface{}, int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, status, err := h(r)
		if err != nil {
			data = err.Error()
		}

		items, ok := data.([]*storage.Item)
		if !ok {
			// TODO: might be good idea to handle this situation?
		}

		var b bytes.Buffer
		for _, item := range items {
			// https://www.ietf.org/rfc/rfc2046.txt says that the newline of plain/text is CRLF
			fmt.Fprintf(&b, "%v\r\n", item.Value)
		}

		w.Header().Set("Content-Type", "plain/text")
		w.WriteHeader(status)
		w.Write(b.Bytes())
	}
}

// responseJSON packs data in application/json format.
func responseJSON(h func(r *http.Request) (interface{}, int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, status, err := h(r)
		if err != nil {
			data = err.Error()
		}

		type response struct {
			Data interface{} `json:"data"`
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		err = json.NewEncoder(w).Encode(response{Data: data})
		if err != nil {
			log.Printf("could not decode to output: %v", err)
		}
	}
}

func (h handler) insfile(r *http.Request) (interface{}, int, error) {
	var modules []storage.Module
	t, err := ioutil.ReadAll(r.Body)
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

	items, err := h.storage.GetItems(modules...)
	if err != nil {
		log.Println(fmt.Errorf("could not retrieve data from database: %v", err))
		return nil, http.StatusInternalServerError, fmt.Errorf("Ups something went wrong")
	}

	return items, http.StatusOK, nil
}
