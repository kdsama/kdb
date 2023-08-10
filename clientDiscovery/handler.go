package clientdiscovery

import "net/http"

// here all the handlers will come

type clientHandler struct {
}

func NewClientHandler() *clientHandler {
	return &clientHandler{}
}

func (ch *clientHandler) AddServer(w http.ResponseWriter, r *http.Request) {

}

func (ch *clientHandler) Get(w http.ResponseWriter, r *http.Request) {

}

func (ch *clientHandler) Set(w http.ResponseWriter, r *http.Request) {

}
func (ch *clientHandler) AutomateGet(w http.ResponseWriter, r *http.Request) {

}

func (ch *clientHandler) AutomateSet(w http.ResponseWriter, r *http.Request) {

}
