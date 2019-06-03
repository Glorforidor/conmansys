package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Glorforidor/conmansys/confservice/storage"
	"github.com/gorilla/mux"
)

type handler struct {
	storage storage.Service
}

// New registers multiple endpoints, assoiciate the storage.Service to the
// handler for data creation and retrieval and returns the handler.
func New(service storage.Service) http.Handler {
	r := mux.NewRouter()

	h := handler{service}

	r.HandleFunc("/health", health)
	r.HandleFunc("/items", responseJSON(h.items)).Methods(http.MethodGet)
	r.HandleFunc("/items/{id:[0-9]+}", responseJSON(h.item)).Methods(http.MethodGet)
	r.HandleFunc("/items", responseJSON(h.createItem)).Methods(http.MethodPost)
	r.HandleFunc("/items/{id:[0-9]+}", responseJSON(h.deleteItem)).Methods(http.MethodDelete)
	r.HandleFunc("/modules", responseJSON(h.modules)).Methods(http.MethodGet)
	r.HandleFunc("/modules/{id:[0-9]+}", responseJSON(h.module)).Methods(http.MethodGet)
	r.HandleFunc("/modules", responseJSON(h.createModule)).Methods(http.MethodPost)
	r.HandleFunc("/modules/{id:[0-9]+}", responseJSON(h.deleteModule)).Methods(http.MethodDelete)
	r.HandleFunc("/itemmodules", responseJSON(h.itemModules)).Methods(http.MethodGet)
	r.HandleFunc("/itemmodules/{id:[0-9]+}", responseJSON(h.itemModule)).Methods(http.MethodGet)
	r.HandleFunc("/itemmodules", responseJSON(h.createItemModule)).Methods(http.MethodPost)
	r.HandleFunc("/itemmodules/{id:[0-9]+}", responseJSON(h.deleteItemModule)).Methods(http.MethodDelete)
	r.HandleFunc("/moduledependencies", responseJSON(h.moduleDependencies)).Methods(http.MethodGet)
	r.HandleFunc(
		"/moduledependencies/dependent/{id:[0-9]+}",
		responseJSON(h.moduleDependenciesByDependentID),
	).Methods(http.MethodGet)
	r.HandleFunc(
		"/moduledependencies/dependee/{id:[0-9]+}",
		responseJSON(h.moduleDependenciesByDependeeID),
	).Methods(http.MethodGet)
	r.HandleFunc("/moduledependencies", responseJSON(h.createModuleDependency)).Methods(http.MethodPost)
	r.HandleFunc(
		"/moduledependencies/dependent/{dependentID:[0-9]+}/dependee/{dependeeID:[0-9]+}",
		responseJSON(h.deleteModuleDependency),
	).Methods(http.MethodDelete)
	r.HandleFunc(
		"/moduledependencies/dependent/{id:[0-9]+}",
		responseJSON(h.deleteModuleDependencyByDependentID),
	).Methods(http.MethodDelete)
	r.HandleFunc(
		"/moduledependencies/dependee/{id:[0-9]+}",
		responseJSON(h.deleteModuleDependencyByDependeeID),
	).Methods(http.MethodDelete)

	return r
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

var (
	contentType = map[string]string{
		"json": "application/json",
	}

	errWrongFormat  = errors.New("wrong input format")
	errMissingValue = errors.New("missing value")
	errNaN          = errors.New("not a number")
	errInternal     = errors.New("Ups something went wrong")
)

func responseJSON(h func(*http.Request) (interface{}, int)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, status := h(r)

		w.Header().Set("Content-Type", contentType["json"])
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(data)
	}
}

type itemResponse struct {
	Item  *storage.Item `json:"item"`
	Error *string       `json:"error"`
}

// item retrieves a specifc item from storage packs it into a response and
// return it as an empty interface and http status.
func (h handler) item(r *http.Request) (data interface{}, status int) {
	params := mux.Vars(r)
	id := strings.TrimSpace(params["id"])
	var resp itemResponse
	var errMsg string

	// routing should prevent this, but might as well guard it
	if id == "" {
		errMsg = errMissingValue.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	i, err := strconv.ParseInt(id, 10, 64)
	// routing should prevent this, but might as well guard it
	if err != nil {
		errMsg = errNaN.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	item, err := h.storage.GetItem(i)
	if err != nil {
		errMsg = errInternal.Error()
		resp.Error = &errMsg
		log.Println(err)
		return resp, http.StatusInternalServerError
	}

	resp.Item = item
	return resp, http.StatusOK
}

type itemsResponse struct {
	Items []*storage.Item `json:"items"`
	Error *string         `json:"error"`
}

// items retrieves items from storage packs it into a response and returns it as
// an empty interface with http status.
func (h handler) items(r *http.Request) (data interface{}, status int) {
	var resp itemsResponse
	// ensure that there is an empty slice
	resp.Items = []*storage.Item{}

	i, err := h.storage.GetItems()
	if err != nil {
		errMsg := errInternal.Error()
		resp.Error = &errMsg
		log.Println(err)
		return resp, http.StatusInternalServerError
	}

	resp.Items = i
	return resp, http.StatusOK
}

// createItem creates an item in the storage and return the newly created item.
func (h handler) createItem(r *http.Request) (data interface{}, status int) {
	var resp itemResponse
	var item storage.Item
	var errMsg string

	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		errMsg = errWrongFormat.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	if item.Value == "" || item.Type == "" || item.Version == "" {
		errMsg = "missing values"
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	i, err := h.storage.CreateItem(item.Value, item.Type, item.Version)
	if err != nil {
		log.Println(err)
		errMsg = errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	// TODO: perhaps a better response besides the item?
	item.ID = i
	resp.Item = &item
	return resp, http.StatusCreated
}

type deleteResponse struct {
	RowsAffected int64   `json:"rows_affected"`
	Error        *string `json:"error"`
}

// deleteItem deletes the item in storage and packs the information about the
// deletion into an response. It returns the response as an empty interface and
// a http status.
func (h handler) deleteItem(r *http.Request) (data interface{}, status int) {
	params := mux.Vars(r)
	id := strings.TrimSpace(params["id"])
	var resp deleteResponse
	var errMsg string

	// routing should prevent this, but might as well guard it
	if id == "" {
		errMsg = errMissingValue.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	i, err := strconv.ParseInt(id, 10, 64)
	// routing should prevent this, but might as well guard it
	if err != nil {
		errMsg = errNaN.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	row, err := h.storage.DeleteItem(i)
	if err != nil {
		log.Println(err)
		errMsg = errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	resp.RowsAffected = row
	return resp, http.StatusOK
}

type moduleResponse struct {
	Module *storage.Module `json:"module"`
	Error  *string         `json:"error"`
}

// module retrieves a module from storage. It packs the information about the
// retrieval into a response. It returns the response as an empty interface and
// a http status.
func (h handler) module(r *http.Request) (data interface{}, status int) {
	params := mux.Vars(r)
	id := params["id"]
	var resp moduleResponse
	var errMsg string
	if id == "" {
		errMsg = errMissingValue.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		errMsg = errNaN.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	module, err := h.storage.GetModule(i)
	if err != nil {
		log.Println(err)
		errMsg = errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	resp.Module = module
	return resp, http.StatusOK
}

type modulesResponse struct {
	Modules []*storage.Module `json:"modules"`
	Error   *string           `json:"error"`
}

// modules retrieve all modules from storage and packs the information about the
// retrieval into a response. It returns the response as an empty interface and
// a http status.
func (h handler) modules(r *http.Request) (data interface{}, status int) {
	var resp modulesResponse
	modules, err := h.storage.GetModules()
	if err != nil {
		log.Println(err)
		errMsg := errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	resp.Modules = modules
	return resp, http.StatusOK
}

// createModule creates a module in the storage. It will respond with the newly
// create module and a http status.
func (h handler) createModule(r *http.Request) (data interface{}, status int) {
	var resp moduleResponse
	var errMsg string
	var module storage.Module
	err := json.NewDecoder(r.Body).Decode(&module)
	if err != nil {
		errMsg = errWrongFormat.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	if module.Value == "" || module.Version == "" {
		errMsg = errMissingValue.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	i, err := h.storage.CreateModule(module.Value, module.Version)
	if err != nil {
		log.Println(err)
		errMsg = errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	module.ID = i
	resp.Module = &module
	return resp, http.StatusCreated
}

// deleteModule deletes the item from storage and packs the deletion information
// into a response. It returns the response as an empty interface and a http
// status.
func (h handler) deleteModule(r *http.Request) (data interface{}, status int) {
	params := mux.Vars(r)
	var resp deleteResponse
	var errMsg string
	id := strings.TrimSpace(params["id"])
	if id == "" {
		errMsg = errMissingValue.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		errMsg = errNaN.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	row, err := h.storage.DeleteModule(i)
	if err != nil {
		log.Println(err)
		errMsg = errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	resp.RowsAffected = row
	return resp, http.StatusOK
}

type itemModuleResponse struct {
	ItemModule *storage.ItemModule `json:"item_module"`
	Error      *string             `json:"error"`
}

// itemModule retrieves a item module from storage and packs the retrivel
// information into a response. It returns the response as an empty interface
// and a http status.
func (h handler) itemModule(r *http.Request) (data interface{}, status int) {
	params := mux.Vars(r)
	var resp itemModuleResponse
	var errMsg string
	id := strings.TrimSpace(params["id"])
	if id == "" {
		errMsg = errMissingValue.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		errMsg = errNaN.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	im, err := h.storage.GetItemModule(i)
	if err != nil {
		log.Println(err)
		errMsg = errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	resp.ItemModule = im
	return resp, http.StatusOK
}

type itemModulesResponse struct {
	ItemModules []*storage.ItemModule `json:"item_modules"`
	Error       *string               `json:"error"`
}

// itemModules retrieve all modules from storage and packs the retrieval
// information into a response. It returns the response as an empty interface
// and a http status.
func (h handler) itemModules(r *http.Request) (data interface{}, status int) {
	var resp itemModulesResponse
	ims, err := h.storage.GetItemModules()
	if err != nil {
		log.Println(err)
		errMsg := errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	resp.ItemModules = ims
	return resp, http.StatusOK
}

// createItemModule inserts a new item module into storage and packs the
// creation information into a response. It returns the response as an empty
// interface and a http status.
func (h handler) createItemModule(r *http.Request) (data interface{}, status int) {
	var resp itemModuleResponse
	var errMsg string
	var im storage.ItemModule
	err := json.NewDecoder(r.Body).Decode(&im)
	if err != nil {
		errMsg = errWrongFormat.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	if im.ItemID == 0 || im.ModuleID == 0 {
		errMsg = errMissingValue.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	id, err := h.storage.CreateItemModule(im.ItemID, im.ModuleID)
	if err != nil {
		log.Println(err)
		errMsg = errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	im.ID = id
	resp.ItemModule = &im
	return resp, http.StatusCreated
}

// deleteItemModule deletes a item module from the storage and packs the
// deletion information into a response. It returns the response as an empty
// interface and a http status.
func (h handler) deleteItemModule(r *http.Request) (data interface{}, status int) {
	params := mux.Vars(r)
	var resp deleteResponse
	var errMsg string
	id := strings.TrimSpace(params["id"])
	if id == "" {
		errMsg = errMissingValue.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		errMsg = errNaN.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	row, err := h.storage.DeleteItemModule(i)
	if err != nil {
		log.Println(err)
		errMsg = errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	resp.RowsAffected = row
	return resp, http.StatusOK
}

type moduleDependenciesResponse struct {
	ModuleDependencies []*storage.ModuleDependency `json:"module_dependencies"`
	Error              *string                     `json:"error"`
}

// moduleDependencies retrieve all module dependencies from storage and packs
// the retrieval information into a response. It returns the response as an
// empty interface and a http status.
func (h handler) moduleDependencies(r *http.Request) (data interface{}, status int) {
	var resp moduleDependenciesResponse
	moddeps, err := h.storage.GetModuleDependencies()
	if err != nil {
		log.Println(err)
		errMsg := errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	resp.ModuleDependencies = moddeps
	return resp, http.StatusOK
}

// moduleDependenciesByDependentID retrieve all moduledependencies from storage
// by a dependent id and packs retrieval information into a response. It returns
// the response as an empty interface and a http status.
func (h handler) moduleDependenciesByDependentID(r *http.Request) (data interface{}, status int) {
	params := mux.Vars(r)
	var resp moduleDependenciesResponse
	var errMsg string
	id := strings.TrimSpace(params["id"])
	if id == "" {
		errMsg = errMissingValue.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		errMsg = errNaN.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	moddeps, err := h.storage.GetModuleDependenciesByDependentID(i)
	if err != nil {
		log.Println(err)
		errMsg = errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	resp.ModuleDependencies = moddeps
	return resp, http.StatusOK
}

// moduleDependenciesByDependeeID retrieve all module dependencies from storage
// by a dependee id and packs retrieval information into a response. It returns
// the response as an empty interface and a http status.
func (h handler) moduleDependenciesByDependeeID(r *http.Request) (data interface{}, status int) {
	params := mux.Vars(r)
	var resp moduleDependenciesResponse
	var errMsg string
	id := strings.TrimSpace(params["id"])
	if id == "" {
		errMsg = errMissingValue.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		errMsg = errNaN.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	moddeps, err := h.storage.GetModuleDependenciesByDependeeID(i)
	if err != nil {
		log.Println(err)
		errMsg = errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	resp.ModuleDependencies = moddeps
	return resp, http.StatusOK
}

type moduleDependencyResponse struct {
	ModuleDependency *storage.ModuleDependency `json:"module_dependency"`
	Error            *string                   `json:"error"`
}

// createModuleDependency inserts a new module dependency into storage and packs
// creation information into a response. It returns the response as an empty
// interface and a http status.
func (h handler) createModuleDependency(r *http.Request) (data interface{}, status int) {
	var resp moduleDependencyResponse
	var errMsg string
	var md storage.ModuleDependency
	err := json.NewDecoder(r.Body).Decode(&md)
	if err != nil {
		errMsg = errWrongFormat.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	if md.Dependent == 0 || md.Dependee == 0 {
		errMsg = errMissingValue.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	err = h.storage.CreateModuleDependency(md.Dependent, md.Dependee)
	if err != nil {
		log.Println(err)
		errMsg = errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	resp.ModuleDependency = &md
	return resp, http.StatusCreated
}

func (h handler) deleteModuleDependency(r *http.Request) (data interface{}, status int) {
	params := mux.Vars(r)
	var resp deleteResponse
	var errMsg string

	dependentID := strings.TrimSpace(params["dependentID"])
	dependeeID := strings.TrimSpace(params["dependeeID"])
	// routing should prevent this, but might as well guard it
	if dependentID == "" || dependeeID == "" {
		errMsg = errMissingValue.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	i, err := strconv.ParseInt(dependentID, 10, 64)
	// routing should prevent this, but might as well guard it
	if err != nil {
		errMsg = errNaN.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	j, err := strconv.ParseInt(dependeeID, 10, 64)
	// routing should prevent this, but might as well guard it
	if err != nil {
		errMsg = errNaN.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	row, err := h.storage.DeleteModuleDependency(i, j)
	if err != nil {
		log.Println(err)
		errMsg = errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	resp.RowsAffected = row
	return resp, http.StatusOK
}

func (h handler) deleteModuleDependencyByDependentID(r *http.Request) (data interface{}, status int) {
	params := mux.Vars(r)
	var resp deleteResponse
	var errMsg string
	id := strings.TrimSpace(params["id"])
	// routing should prevent this, but might as well guard it
	if id == "" {
		errMsg = errMissingValue.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	i, err := strconv.ParseInt(id, 10, 64)
	// routing should prevent this, but might as well guard it
	if err != nil {
		errMsg = errNaN.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	rows, err := h.storage.DeleteModuleDependencyByDependentID(i)
	if err != nil {
		log.Println(err)
		errMsg = errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	resp.RowsAffected = rows
	return resp, http.StatusOK
}

func (h handler) deleteModuleDependencyByDependeeID(r *http.Request) (data interface{}, status int) {
	params := mux.Vars(r)
	var resp deleteResponse
	var errMsg string
	id := strings.TrimSpace(params["id"])
	// routing should prevent this, but might as well guard it
	if id == "" {
		errMsg = errMissingValue.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	i, err := strconv.ParseInt(id, 10, 64)
	// routing should prevent this, but might as well guard it
	if err != nil {
		errMsg = errNaN.Error()
		resp.Error = &errMsg
		return resp, http.StatusBadRequest
	}

	rows, err := h.storage.DeleteModuleDependencyByDependeeID(i)
	if err != nil {
		log.Println(err)
		errMsg = errInternal.Error()
		resp.Error = &errMsg
		return resp, http.StatusInternalServerError
	}

	resp.RowsAffected = rows
	return resp, http.StatusOK
}
