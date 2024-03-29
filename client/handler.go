package clientdiscovery

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// here all the handlers will come

type clientHandler struct {
	service *service
	l       *zap.SugaredLogger
}

func NewClientHandler(l *zap.SugaredLogger) *clientHandler {
	return &clientHandler{
		service: NewService(),
		l:       l,
	}
}

func (ch *clientHandler) AddServer(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	ch.l.Info("AddServer")
	requestsTotal.WithLabelValues("AddServer").Inc()
	val := r.URL.Query().Get("name")
	if val != "" {

		err := ch.service.addServer(val)
		if err != nil {
			w.Write([]byte("Not OK"))
			ch.l.Error(err.Error())
			zap.Fields(zap.String("api", "add-server"))
			return
		}
		w.Write([]byte("OK"))

	} else {
		w.Write([]byte("Not OK"))
	}
	requestLatency.WithLabelValues("AddServer").Observe(float64(time.Since(t).Milliseconds()))
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
	requestLatency.WithLabelValues("Get").Observe(float64(time.Since(t).Milliseconds()))

}

func (ch *clientHandler) Leader(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(ch.service.leader))

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

	requestLatency.WithLabelValues("GetRandom").Observe(float64(time.Since(t).Milliseconds()))
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
	linear := r.URL.Query().Get("linear")
	// we need to send data as well
	val, err := ch.service.automateGet(duration, tr, sleep, linear)
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
