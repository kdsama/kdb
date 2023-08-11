package clientdiscovery

import (
	"fmt"
	"net/http"
	"time"
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
	t := time.Now()
	requestsTotal.WithLabelValues("AddServer").Inc()
	val := r.URL.Query().Get("name")

	if val != "" {

		err := ch.service.addServer(val)
		if err != nil {
			w.Write([]byte("Not OK"))
			return
		}
		w.Write([]byte("OK"))

	} else {
		w.Write([]byte("Not OK"))
	}
	requestLatency.WithLabelValues("AddServer").Observe(float64(time.Since(t)) / 1000)
}

func (ch *clientHandler) Get(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	requestsTotal.WithLabelValues("Get").Inc()
	key := r.URL.Query().Get("key")
	val, err := ch.service.get(key)
	if err != nil {
		w.Write([]byte("Not OK"))
		return
	} else {
		w.Write([]byte(fmt.Sprintf("Key :: %s, value :: %s ::", key, val)))
	}
	requestLatency.WithLabelValues("Get").Observe(float64(time.Since(t)) / 1000)

}

func (ch *clientHandler) GetRandom(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	requestsTotal.WithLabelValues("GetRandom").Inc()
	key := r.URL.Query().Get("key")
	val, err := ch.service.getRandom(key)
	if err != nil {
		w.Write([]byte("Not OK"))
		return
	} else {
		w.Write([]byte(fmt.Sprintf("Key :: %s, value :: %s ::", key, val)))
	}

	requestLatency.WithLabelValues("GetRandom").Observe(float64(time.Since(t)) / 1000)
}

func (ch *clientHandler) Set(w http.ResponseWriter, r *http.Request) {
	// to be done only to the leader

	key := r.URL.Query().Get("key")
	val := r.URL.Query().Get("value")
	err := ch.service.set(key, val)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {

		w.Write([]byte("OK"))
	}

}
func (ch *clientHandler) AutomateGet(w http.ResponseWriter, r *http.Request) {
	//duration := r.URL.Query().Get("duration")
	// tr := r.URL.Query().Get("requests")
	duration := r.URL.Query().Get("duration")
	tr := r.URL.Query().Get("requests")
	sleep := r.URL.Query().Get("sleep")
	// we need to send data as well
	val, err := ch.service.automateGet(duration, tr, sleep)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte(fmt.Sprint(val)))
	}

}

func (ch *clientHandler) AutomateSet(w http.ResponseWriter, r *http.Request) {
	duration := r.URL.Query().Get("duration")
	tr := r.URL.Query().Get("requests")
	sleep := r.URL.Query().Get("sleep")
	// we need to send data as well
	val, err := ch.service.automateSet(duration, tr, sleep)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte(fmt.Sprint(val)))
	}
}
