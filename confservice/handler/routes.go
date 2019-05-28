package handler

import (
	"encoding/json"
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
	r.HandleFunc("/items", h.items).Methods(http.MethodGet)
	r.HandleFunc("/items/{id:[0-9]+}", h.item).Methods(http.MethodGet)
	r.HandleFunc("/items", h.createItem).Methods(http.MethodPost)
	r.HandleFunc("/items/{id:[0-9]+}", h.deleteItem).Methods(http.MethodDelete)
	r.HandleFunc("/modules", h.modules).Methods(http.MethodGet)
	r.HandleFunc("/modules/{id:[0-9]+}", h.module).Methods(http.MethodGet)
	r.HandleFunc("/modules", h.createModule).Methods(http.MethodPost)
	r.HandleFunc("/modules/{id:[0-9]+}", h.deleteModule).Methods(http.MethodDelete)
	r.HandleFunc("/itemmodules", h.itemModules).Methods(http.MethodGet)
	r.HandleFunc("/itemmodules/{id:[0-9]+}", h.itemModule).Methods(http.MethodGet)
	r.HandleFunc("/itemmodules", h.createItemModule).Methods(http.MethodPost)
	r.HandleFunc("/itemmodules/{id:[0-9]+}", h.deleteItemModule).Methods(http.MethodDelete)
	r.HandleFunc("/moduledependencies", h.moduleDependencies).Methods(http.MethodGet)
	r.HandleFunc(
		"/moduledependencies/dependent/{id:[0-9]+}",
		h.moduleDependenciesByDependentID,
	).Methods(http.MethodGet)
	r.HandleFunc(
		"/moduledependencies/dependee/{id:[0-9]+}",
		h.moduleDependenciesByDependeeID,
	).Methods(http.MethodGet)
	r.HandleFunc("/moduledependencies", h.createModuleDependency).Methods(http.MethodPost)
	r.HandleFunc(
		"/moduledependencies/dependent/{dependentID:[0-9]+}/dependee/{dependeeID:[0-9]+}",
		h.deleteModuleDependency,
	).Methods(http.MethodDelete)
	r.HandleFunc(
		"/moduledependencies/dependent/{id:[0-9]+}",
		h.deleteModuleDependencyByDependentID,
	).Methods(http.MethodDelete)
	r.HandleFunc(
		"/moduledependencies/dependee/{id:[0-9]+}",
		h.deleteModuleDependencyByDependeeID,
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
)

func (h handler) item(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := strings.TrimSpace(params["id"])

	// routing should prevent this, but might as well guard it
	if id == "" {
		http.Error(w, "missing value", http.StatusBadRequest)
		return
	}

	i, err := strconv.ParseInt(id, 10, 64)
	// routing should prevent this, but might as well guard it
	if err != nil {
		http.Error(w, "not a number", http.StatusBadRequest)
		return
	}

	item, err := h.storage.GetItem(i)
	if err != nil {
		http.Error(w, "Ups something went wrong", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", contentType["json"])
	json.NewEncoder(w).Encode(item)
}

func (h handler) items(w http.ResponseWriter, r *http.Request) {
	i, err := h.storage.GetItems()
	if err != nil {
		http.Error(w, "Ups something went wrong", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", contentType["json"])
	json.NewEncoder(w).Encode(i)
}

func (h handler) createItem(w http.ResponseWriter, r *http.Request) {
	item := &storage.Item{}
	err := json.NewDecoder(r.Body).Decode(item)
	if err != nil {
		http.Error(w, "wrong input format", http.StatusBadRequest)
		return
	}

	if item.Value == "" || item.Type == "" || item.Version == "" {
		http.Error(w, "missing values", http.StatusBadRequest)
		return
	}

	i, err := h.storage.CreateItem(item.Value, item.Type, item.Version)
	if err != nil {
		http.Error(
			w,
			"Ups something happend, thus the item is not created",
			http.StatusInternalServerError,
		)
		log.Println(err)
		return
	}

	// TODO: perhaps a better response besides the item?
	item.ID = i
	w.Header().Set("Content-Type", contentType["json"])
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (h handler) deleteItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := strings.TrimSpace(params["id"])
	// routing should prevent this, but might as well guard it
	if id == "" {
		http.Error(w, "missing value", http.StatusBadRequest)
		return
	}

	i, err := strconv.ParseInt(id, 10, 64)
	// routing should prevent this, but might as well guard it
	if err != nil {
		http.Error(w, "not a number", http.StatusBadRequest)
		return
	}

	rows, err := h.storage.DeleteItem(i)
	if err != nil {
		http.Error(w, "could not delete item", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", contentType["json"])
	json.NewEncoder(w).Encode(map[string]interface{}{"RowsAffected": rows})
}

func (h handler) module(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		http.Error(w, "missing value", http.StatusBadRequest)
		return
	}

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "not a number", http.StatusBadRequest)
		return
	}

	module, err := h.storage.GetModule(i)
	if err != nil {
		http.Error(w, "Ups something went wrong", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if module == nil {
		// TODO: do something about it
	}

	w.Header().Set("Content-Type", contentType["json"])
	json.NewEncoder(w).Encode(module)
}

func (h handler) modules(w http.ResponseWriter, r *http.Request) {
	modules, err := h.storage.GetModules()
	if err != nil {
		http.Error(w, "Ups something went wrong", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", contentType["json"])
	json.NewEncoder(w).Encode(modules)
}

func (h handler) createModule(w http.ResponseWriter, r *http.Request) {
	module := &storage.Module{}
	err := json.NewDecoder(r.Body).Decode(module)
	if err != nil {
		http.Error(w, "wrong input format", http.StatusBadRequest)
		return
	}

	if module.Value == "" || module.Version == "" {
		http.Error(w, "missing values", http.StatusBadRequest)
		return
	}

	i, err := h.storage.CreateModule(module.Value, module.Version)
	if err != nil {
		http.Error(w, "Ups something went wrong", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	module.ID = i
	w.Header().Set("Content-Type", contentType["json"])
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(module)
}

func (h handler) deleteModule(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := strings.TrimSpace(params["id"])
	if id == "" {
		http.Error(w, "missing value", http.StatusBadRequest)
		return
	}

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "not a number", http.StatusBadRequest)
		return
	}

	row, err := h.storage.DeleteModule(i)
	if err != nil {
		http.Error(w, "Ups something went wrong", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", contentType["json"])
	json.NewEncoder(w).Encode(map[string]interface{}{"RowsAffected": row})
}

func (h handler) itemModule(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := strings.TrimSpace(params["id"])
	if id == "" {
		http.Error(w, "missing value", http.StatusBadRequest)
		return
	}

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "not a number", http.StatusBadRequest)
		return
	}

	im, err := h.storage.GetItemModule(i)
	if err != nil {
		http.Error(w, "Ups something went wrong", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", contentType["json"])
	json.NewEncoder(w).Encode(im)
}

func (h handler) itemModules(w http.ResponseWriter, r *http.Request) {
	ims, err := h.storage.GetItemModules()
	if err != nil {
		http.Error(w, "Ups something went wrong", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", contentType["json"])
	json.NewEncoder(w).Encode(ims)
}

func (h handler) createItemModule(w http.ResponseWriter, r *http.Request) {
	im := &storage.ItemModule{}
	err := json.NewDecoder(r.Body).Decode(im)
	if err != nil {
		http.Error(w, "wrong input format", http.StatusBadRequest)
		return
	}

	if im.ItemID == 0 || im.ModuleID == 0 {
		http.Error(w, "missing values", http.StatusBadRequest)
		return
	}

	id, err := h.storage.CreateItemModule(im.ItemID, im.ModuleID)
	if err != nil {
		http.Error(w, "Ups something went wrong", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	im.ID = id
	w.Header().Set("Content-Type", contentType["json"])
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(im)
}

func (h handler) deleteItemModule(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := strings.TrimSpace(params["id"])
	if id == "" {
		http.Error(w, "missing value", http.StatusBadRequest)
		return
	}

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "not a number", http.StatusBadRequest)
		return
	}

	row, err := h.storage.DeleteItemModule(i)
	if err != nil {
		http.Error(w, "Ups something went wrong", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", contentType["json"])
	json.NewEncoder(w).Encode(map[string]interface{}{"RowsAffected": row})
}

func (h handler) moduleDependencies(w http.ResponseWriter, r *http.Request) {
	moddeps, err := h.storage.GetModuleDependencies()
	if err != nil {
		http.Error(w, "Ups something went wrong", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", contentType["json"])
	json.NewEncoder(w).Encode(moddeps)
}

func (h handler) moduleDependenciesByDependentID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := strings.TrimSpace(params["id"])
	if id == "" {
		http.Error(w, "missing value", http.StatusBadRequest)
		return
	}

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "not a number", http.StatusBadRequest)
		return
	}

	moddeps, err := h.storage.GetModuleDependenciesByDependentID(i)
	if err != nil {
		http.Error(w, "Ups something went wrong", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", contentType["json"])
	json.NewEncoder(w).Encode(moddeps)
}

func (h handler) moduleDependenciesByDependeeID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := strings.TrimSpace(params["id"])
	if id == "" {
		http.Error(w, "missing value", http.StatusBadRequest)
		return
	}

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "not a number", http.StatusBadRequest)
		return
	}

	moddeps, err := h.storage.GetModuleDependenciesByDependeeID(i)
	if err != nil {
		http.Error(w, "Ups something went wrong", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", contentType["json"])
	json.NewEncoder(w).Encode(moddeps)
}

func (h handler) createModuleDependency(w http.ResponseWriter, r *http.Request) {
	md := &storage.ModuleDependency{}
	err := json.NewDecoder(r.Body).Decode(md)
	if err != nil {
		http.Error(w, "wrong input format", http.StatusBadRequest)
		return
	}

	if md.Dependent == 0 || md.Dependee == 0 {
		http.Error(w, "missing values", http.StatusBadRequest)
		return
	}

	err = h.storage.CreateModuleDependency(md.Dependent, md.Dependee)
	if err != nil {
		http.Error(w, "Ups something went wrong", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", contentType["json"])
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(md)
}

func (h handler) deleteModuleDependency(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	dependentID := strings.TrimSpace(params["dependentID"])
	dependeeID := strings.TrimSpace(params["dependeeID"])
	// routing should prevent this, but might as well guard it
	if dependentID == "" || dependeeID == "" {
		http.Error(w, "missing value", http.StatusBadRequest)
		return
	}

	i, err := strconv.ParseInt(dependentID, 10, 64)
	// routing should prevent this, but might as well guard it
	if err != nil {
		http.Error(w, "not a number", http.StatusBadRequest)
		return
	}

	j, err := strconv.ParseInt(dependeeID, 10, 64)
	// routing should prevent this, but might as well guard it
	if err != nil {
		http.Error(w, "not a number", http.StatusBadRequest)
		return
	}

	rows, err := h.storage.DeleteModuleDependency(i, j)
	if err != nil {
		http.Error(w, "could not delete item", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", contentType["json"])
	json.NewEncoder(w).Encode(map[string]interface{}{"RowsAffected": rows})
}

func (h handler) deleteModuleDependencyByDependentID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := strings.TrimSpace(params["id"])
	// routing should prevent this, but might as well guard it
	if id == "" {
		http.Error(w, "missing value", http.StatusBadRequest)
		return
	}

	i, err := strconv.ParseInt(id, 10, 64)
	// routing should prevent this, but might as well guard it
	if err != nil {
		http.Error(w, "not a number", http.StatusBadRequest)
		return
	}

	rows, err := h.storage.DeleteModuleDependencyByDependentID(i)
	if err != nil {
		http.Error(w, "could not delete item", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", contentType["json"])
	json.NewEncoder(w).Encode(map[string]interface{}{"RowsAffected": rows})
}

func (h handler) deleteModuleDependencyByDependeeID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := strings.TrimSpace(params["id"])
	// routing should prevent this, but might as well guard it
	if id == "" {
		http.Error(w, "missing value", http.StatusBadRequest)
		return
	}

	i, err := strconv.ParseInt(id, 10, 64)
	// routing should prevent this, but might as well guard it
	if err != nil {
		http.Error(w, "not a number", http.StatusBadRequest)
		return
	}

	rows, err := h.storage.DeleteModuleDependencyByDependeeID(i)
	if err != nil {
		http.Error(w, "could not delete item", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", contentType["json"])
	json.NewEncoder(w).Encode(map[string]interface{}{"RowsAffected": rows})
}
