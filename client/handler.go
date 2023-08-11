package clientdiscovery

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// here all the handlers will come
var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: os.Args[1] + "_client_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"reqtype"},
	)
	requestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    os.Args[1] + "_client_request_latency",
		Help:    "btree inserts :: btree layer",
		Buckets: []float64{0.0, 20.0, 40.0, 60.0, 80.0, 100.0, 160.0, 180.0, 200.0, 400.0, 800.0, 1600.0},
	}, []string{"reqtype"})
)

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
	fmt.Println("Name is ", val)
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
	t := time.Now()

	key := r.URL.Query().Get("key")
	val := r.URL.Query().Get("value")
	err := ch.service.set(key, val)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {

		w.Write([]byte("OK"))
	}
	requestLatency.WithLabelValues("Set").Observe(float64(time.Since(t)) / 1000)

}
func (ch *clientHandler) AutomateGet(w http.ResponseWriter, r *http.Request) {
	//duration := r.URL.Query().Get("duration")
	// tr := r.URL.Query().Get("requests")

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

func (ch *clientHandler) RequestStatus(w http.ResponseWriter, r *http.Request) {
	rr := r.URL.Query().Get("r")
	d, _ := strconv.Atoi(rr)

	if d != curr {
		w.Write([]byte("Invalid request "))
	} else {
		w.Write([]byte(fmt.Sprintf("errCount ::%d and successCount ::%d", errCount, count)))
	}
}

func init() {
	prometheus.MustRegister(requestsTotal, requestLatency)
}
