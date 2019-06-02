package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// helper structs since for code clearness. The struct are not used in the
// templates but only for marshalling into json to the api gateway. Using
// map[string]interface{} for templates - seems easier to do.. for now.

type item struct {
	ID      int64  `json:"id"`
	Value   string `json:"value"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

type module struct {
	ID      int64  `json:"id"`
	Value   string `json:"value"`
	Version string `json:"version"`
}

type itemModule struct {
	ItemID   int64 `json:"item_id"`
	ModuleID int64 `json:"module_id"`
}

type moduleDependency struct {
	Dependent int64 `json:"dependent"`
	Dependee  int64 `json:"dependee"`
}

// lets define a new type which can be used as an enum.
// This will help with the templating. (there might be a better way to do this,
// have not made many web apps.)
type conmansys int

// get the name representing the value of our conmansys enum
func (c conmansys) String() string {
	return [...]string{"unknown", "items", "modules", "itemmodules", "moduledependencies"}[c]
}

// create the enum.
const (
	unknown conmansys = iota
	itemType
	moduleType
	itemModuleType
	moduleDependencyType
)

func save(c conmansys, r *http.Request) ([]byte, int, error) {
	val := strings.TrimSpace(r.FormValue("value"))
	t := strings.TrimSpace(r.FormValue("type"))
	ver := strings.TrimSpace(r.FormValue("version"))
	itemID := strings.TrimSpace(r.FormValue("item_id"))
	moduleID := strings.TrimSpace(r.FormValue("module_id"))
	dependentID := strings.TrimSpace(r.FormValue("dependent_id"))
	dependeeID := strings.TrimSpace(r.FormValue("dependee_id"))

	var data interface{}
	switch c {
	case itemType:
		if len(val) == 0 || len(t) == 0 || len(ver) == 0 {
			return nil, http.StatusBadRequest, fmt.Errorf(
				"May not have empty input: Value: %q, Type: %q, Version: %q",
				val, t, ver,
			)
		}

		data = item{Value: val, Type: t, Version: ver}
	case moduleType:
		if len(val) == 0 || len(ver) == 0 {
			return nil, http.StatusBadRequest, fmt.Errorf(
				"May not have empty input: Value: %q, Version: %q", val, ver)
		}

		data = module{Value: val, Version: ver}
	case itemModuleType:
		if len(itemID) == 0 || len(moduleID) == 0 {
			return nil, http.StatusBadRequest, fmt.Errorf(
				"May not have empty input: ItemID: %q, ModuleID: %q",
				itemID, moduleID,
			)
		}
		// TODO: properly check for parse error
		iID, _ := strconv.ParseInt(itemID, 10, 64)
		mID, _ := strconv.ParseInt(moduleID, 10, 64)

		data = itemModule{ItemID: iID, ModuleID: mID}
	case moduleDependencyType:
		if len(dependentID) == 0 || len(dependeeID) == 0 {
			return nil, http.StatusBadRequest, fmt.Errorf(
				"May not have empty input: Dependent: %q, Dependee: %q",
				dependentID, dependeeID,
			)
		}

		d, _ := strconv.ParseInt(dependentID, 10, 64)
		dd, _ := strconv.ParseInt(dependeeID, 10, 64)

		data = moduleDependency{Dependent: d, Dependee: dd}
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("could not marshal data: %v, error: %v", data, err)
		return nil, http.StatusInternalServerError, err
	}

	return b, http.StatusOK, nil
}

func saveHandler(c conmansys, target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, status, err := save(c, r)
		if err != nil {
			http.Error(w, err.Error(), status)
			return
		}

		var m map[string]interface{}
		err = json.Unmarshal(b, &m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		req, err := http.NewRequest(http.MethodPost, target, bytes.NewReader(b))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var mm map[string]interface{}
		err = json.Unmarshal(b, &mm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if v, ok := mm["id"]; ok {
			m["id"] = int64(v.(float64))
		}

		m["string"] = c.String()

		renderTemplate(w, "save.html", m)
	}
}

func createHandler(c conmansys) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := map[string]interface{}{"string": c.String()}
		renderTemplate(w, "create.html", m)
	}
}

func deleteModDepHandler(c conmansys, target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		dependentID := params["dependentID"]
		dependeeID := params["dependeeID"]

		var url string

		if dependentID != "" && dependeeID != "" {
			url = fmt.Sprintf(
				"%v/dependent/%v/dependee/%v",
				target, dependentID, dependeeID,
			)
		} else if dependentID != "" {
			url = fmt.Sprintf(
				"%v/dependent/%v",
				target, dependentID,
			)
		} else if dependeeID != "" {
			url = fmt.Sprintf(
				"%v/dependee/%v",
				target, dependeeID,
			)
		} else {
			http.Error(w, "Missing Dependent or Dependee or both", http.StatusBadRequest)
			return
		}

		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			log.Printf("could not create DELETE request: %v", err)
			http.NotFound(w, r)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("could not create DELETE request: %v", err)
			http.NotFound(w, r)
			return
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err.Error())
			return
		}

		log.Println(string(b))

		m := make(map[string]interface{})
		m["id"] = fmt.Sprintf("%v %v", dependentID, dependeeID)
		m["string"] = c.String()
		//err = json.NewDecoder(resp.Body).Decode(&m)
		err = json.Unmarshal(b, &m)
		if err != nil {
			log.Printf("could not decode response body: %v", err)
			http.NotFound(w, r)
			return
		}

		renderTemplate(w, "delete.html", m)
	}
}

func deleteHandler(c conmansys, target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]

		url := fmt.Sprintf("%v/%v", target, id)
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			log.Printf("could not create DELETE request: %v", err)
			http.NotFound(w, r)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("could not create DELETE request: %v", err)
			http.NotFound(w, r)
			return
		}
		defer resp.Body.Close()

		m := make(map[string]interface{})
		m["id"] = id
		m["string"] = c.String()
		err = json.NewDecoder(resp.Body).Decode(&m)
		if err != nil {
			log.Printf("could not decode response body: %v", err)
			http.NotFound(w, r)
			return
		}

		renderTemplate(w, "delete.html", m)
	}
}

func viewHandler(c conmansys, target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(target)
		if err != nil {
			log.Printf("could not receive data: %v", err)
			http.NotFound(w, r)
			return
		}
		defer resp.Body.Close()

		var m []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&m)
		if err != nil {
			log.Printf("could not decode data: %v", err)
			http.NotFound(w, r)
			return
		}

		mm := map[string]interface{}{
			"string": c.String(),
			"list":   m,
		}

		renderTemplate(w, "view.html", mm)
	}
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin.html", nil)
}

func insfileHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "insfile.html", nil)
}

func convertCommaToJSON(modules string) []byte {
	modules = strings.TrimSpace(modules)
	modules = strings.TrimSuffix(modules, ",")
	if modules == "" {
		return []byte("[]")
	}

	s := strings.Split(modules, ",")

	var builder strings.Builder
	builder.WriteString("[")

	for i, ss := range s {
		builder.WriteString("{\"id\": ")
		builder.WriteString(ss)
		builder.WriteString("}")

		if len(s)-1 != i {
			builder.WriteString(",")
		}
	}

	builder.WriteString("]")

	return []byte(builder.String())
}

func insfileCreationHandler(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mods := r.FormValue("modules")
		j := r.FormValue("json")
		t := r.FormValue("text")
		jTraverse := r.FormValue("json_with_traverse")
		tTraverse := r.FormValue("text_with_traverse")

		data := convertCommaToJSON(mods)

		if !json.Valid(data) {
			log.Printf("not valid json: %v", data)
		}

		url := target

		if t != "" {
			url = url + "/text"
		} else if jTraverse != "" {
			url = url + "/traverse"
		} else if tTraverse != "" {
			url = url + "/traverse/text"
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// if they are not empty, the value read from body must be in json format
		if j != "" || jTraverse != "" {
			if !json.Valid(b) {
				log.Printf("given data: %v, was not valid json", string(b))
				http.Error(w, "Ups something went wrong", http.StatusInternalServerError)
				return
			}
			var d map[string]interface{}
			json.Unmarshal(b, &d)

			if d["error"] != nil && d["error"].(string) != "" {
				log.Printf("There was an error: %v", d["error"])
			}

			d["string"] = "insfile"

			renderTemplate(w, "save.html", d)
		} else {
			// TODO: maybe write text to a text area?
			w.Header().Set("Content-Type", "text/plain")
			w.Write(b)
		}
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, data)

	if err != nil {
		log.Printf("could not render template %v: %v", tmpl, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var templates *template.Template

func main() {

	apigateway, ok := os.LookupEnv("APIGATEWAY_HOST")
	if !ok {
		log.Fatal("Missing enviroment variable APIGATEWAY_HOST")
	}

	// endpoints to the api gateway
	var (
		apigatewayURL         = fmt.Sprintf("http://%v/api", apigateway)
		itemsURL              = apigatewayURL + "/items"
		modulesURL            = apigatewayURL + "/modules"
		itemModulesURL        = apigatewayURL + "/itemmodules"
		moduleDependenciesURL = apigatewayURL + "/moduledependencies"
		insfileURL            = apigatewayURL + "/insfile"
	)
	funcs := template.FuncMap{
		"title": strings.Title,
	}
	templates = template.Must(
		template.New("admin.html").Funcs(funcs).
			ParseGlob("templates/*.html"),
	)
	r := mux.NewRouter()

	// create fileserver to serve static content.
	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	r.HandleFunc("/", adminHandler)
	item := r.PathPrefix("/items").Subrouter()
	item.HandleFunc("", viewHandler(itemType, itemsURL))
	item.HandleFunc("/create", createHandler(itemType))
	item.HandleFunc("/save", saveHandler(itemType, itemsURL))
	item.HandleFunc("/delete/{id:[0-9]+}", deleteHandler(itemType, itemsURL))
	module := r.PathPrefix("/modules").Subrouter()
	module.HandleFunc("", viewHandler(moduleType, modulesURL))
	module.HandleFunc("/create", createHandler(moduleType))
	module.HandleFunc("/save", saveHandler(moduleType, modulesURL))
	module.HandleFunc("/delete/{id:[0-9]+}", deleteHandler(moduleType, modulesURL))
	itemModule := r.PathPrefix("/itemmodules").Subrouter()
	itemModule.HandleFunc("", viewHandler(itemModuleType, itemModulesURL))
	itemModule.HandleFunc("/create", createHandler(itemModuleType))
	itemModule.HandleFunc("/save", saveHandler(itemModuleType, itemModulesURL))
	itemModule.HandleFunc("/delete/{id:[0-9]+}", deleteHandler(itemModuleType, itemModulesURL))
	moduleDependency := r.PathPrefix("/moduledependencies").Subrouter()
	moduleDependency.HandleFunc("", viewHandler(moduleDependencyType, moduleDependenciesURL))
	moduleDependency.HandleFunc("/create", createHandler(moduleDependencyType))
	moduleDependency.HandleFunc("/save", saveHandler(moduleDependencyType, moduleDependenciesURL))
	moduleDependency.HandleFunc(
		"/delete/dependent/{dependentID:[0-9]+}/dependee/{dependeeID:[0-9]+}",
		deleteModDepHandler(moduleDependencyType, moduleDependenciesURL),
	)
	moduleDependency.HandleFunc(
		"/delete/dependent/{dependentID:[0-9]+}",
		deleteModDepHandler(moduleDependencyType, moduleDependenciesURL),
	)
	moduleDependency.HandleFunc(
		"/delete/dependee/{dependeeID:[0-9]+}",
		deleteModDepHandler(moduleDependencyType, moduleDependenciesURL),
	)
	insfile := r.PathPrefix("/insfile").Subrouter()
	insfile.HandleFunc("", insfileHandler)
	insfile.HandleFunc("/created", insfileCreationHandler(insfileURL))

	srv := http.Server{
		Addr:         "",
		Handler:      r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
