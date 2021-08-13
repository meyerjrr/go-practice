package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Coaster struct {
	Name         string
	Manufacturer string
	ID           string
	InPark       string
	Height       int
}

type coasterHandlers struct {
	sync.Mutex
	store map[string]Coaster
}

func handleError(writer http.ResponseWriter) {

	writer.WriteHeader(http.StatusInternalServerError)
	writer.Write([]byte("An error occured"))
	return
}

func (h *coasterHandlers) coasters(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		h.get(writer, request)
		return
	case "POST":
		h.post(writer, request)
		return
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
		writer.Write([]byte("method not allowed"))
		return
	}

}

func (h *coasterHandlers) getCoaster(writer http.ResponseWriter, request *http.Request) {

	parts := strings.Split(request.URL.String(), "/")

	if len(parts) != 3 {
		handleError(writer)
	}

	fmt.Println(parts[2])

	h.Lock()

	coaster, ok := h.store[parts[2]]

	h.Unlock()

	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	jsonBytes, err := json.Marshal(coaster)
	if err != nil {

		writer.Write([]byte(err.Error()))

	}
	writer.Header().Add("content-type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(jsonBytes)
}

func (h *coasterHandlers) get(writer http.ResponseWriter, request *http.Request) {

	coasters := make([]Coaster, len(h.store))

	h.Lock()
	i := 0
	for _, coaster := range h.store {
		coasters[i] = coaster
		i++
	}
	h.Unlock()
	jsonBytes, err := json.Marshal(coasters)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))

	}
	writer.Header().Add("content-type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(jsonBytes)
}

func (h *coasterHandlers) post(writer http.ResponseWriter, request *http.Request) {
	bodyBytes, err := ioutil.ReadAll(request.Body)

	defer request.Body.Close()

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	contentType := request.Header.Get("content-type")
	if contentType != "application/json" {
		writer.WriteHeader(http.StatusUnsupportedMediaType)
		writer.Write([]byte(fmt.Sprintf("nee dcontent type 'application/json. but got '%s'", contentType)))

	}
	var coaster Coaster

	err = json.Unmarshal(bodyBytes, &coaster)

	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
	}

	coaster.ID = fmt.Sprintf("%d", time.Now().UnixNano())

	h.Lock()

	h.store[coaster.ID] = coaster

	writer.WriteHeader(http.StatusCreated)
	writer.Write([]byte("Successfully added"))
	defer h.Unlock()

}

func newCoasterHandlers() *coasterHandlers {
	return &coasterHandlers{
		store: map[string]Coaster{},
	}
}

type adminPortal struct {
	password string
}

func newAdminPortal() *adminPortal {
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		panic("env not set")
	}

	return &adminPortal{password: password}
}

func (a adminPortal) handler(writer http.ResponseWriter, request *http.Request) {
	user, pass, ok := request.BasicAuth()

	if !ok || user != "admin" || pass != a.password {
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte("401 - unauthorized user"))
		return
	}

	writer.Write([]byte("<html><h1>Super secret admin portal</h1></html>"))
}

func main() {
	fmt.Println("Server started")
	coasterHandlers := newCoasterHandlers()

	admin := newAdminPortal()
	http.HandleFunc("/coasters", coasterHandlers.coasters)
	http.HandleFunc("/coasters/", coasterHandlers.getCoaster)
	http.HandleFunc("/admin", admin.handler)
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		panic(err)
	}
}
