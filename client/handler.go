package clientdiscovery

import (
	"fmt"
	"net/http"
)

// here all the handlers will come

type clientHandler struct {
	service *service
}

func NewClientHandler() *clientHandler {
	return &clientHandler{
		service: NewService(),
	}
}

func (ch *clientHandler) AddServer(w http.ResponseWriter, r *http.Request) {
	val := r.URL.Query().Get("name")
	fmt.Println("Name is ", val)
	if val != "" {

		err := ch.service.addServer(val)
		if err != nil {
			w.Write([]byte("Not OK"))
			return
		}
		w.Write([]byte("OK"))
		return
	}
	w.Write([]byte("Not OK"))

}

func (ch *clientHandler) Get(w http.ResponseWriter, r *http.Request) {
	// key := r.URL.Query().Get("key")
	// err := ch.service.get(key, val)
}

func (ch *clientHandler) Set(w http.ResponseWriter, r *http.Request) {
	// to be done only to the leader
	key := r.URL.Query().Get("key")
	val := r.URL.Query().Get("value")
	err := ch.service.set(key, val)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	w.Write([]byte("OK"))
}
func (ch *clientHandler) AutomateGet(w http.ResponseWriter, r *http.Request) {

}

func (ch *clientHandler) AutomateSet(w http.ResponseWriter, r *http.Request) {

}
