package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var myPort string
var myHost string

func getUrl() string {
	if myPort == "" {
		myPort = getenv("PORT", "8000")
	}
	if myHost == "" {
		myHost = getenv("HOST", "localhost")
	}
	return fmt.Sprintf("%s:%s", myHost, myPort)
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}
	return def
}

var userStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_get_user_status_count",
		Help: "Count of status returned by user",
	},
	[]string{"user", "status"},
)

func init() {
	prometheus.MustRegister(userStatus)
}

type MyRequest struct {
	User string
}

func server(w http.ResponseWriter, r *http.Request) {
	var mr MyRequest
	if err := json.NewDecoder(r.Body).Decode(&mr); err != nil {
		log.Fatalf("Decode failed with %q", err)
	}

	var status string
	if rand.Float32() > 0.8 {
		status = "4xx"
	} else {
		status = "2xx"
	}

	log.Println(mr.User, status)
	if _, err := w.Write([]byte(status)); err != nil {
		log.Fatalf("Write response failed with %q", err)
	}

	user := mr.User
	userStatus.WithLabelValues(user, status).Inc()
}

func producer() {
	userPool := []string{"bob", "alice", "jack"}
	for {
		postBody, err := json.Marshal(MyRequest{
			User: userPool[rand.Intn(len(userPool))],
		})
		if err != nil {
			log.Fatalf("Marshal failed with %q", err)
		}
		requestBody := bytes.NewBuffer(postBody)
		if _, err := http.Post(fmt.Sprintf("http://%s/server", getUrl()), "application/json", requestBody); err != nil {
			log.Fatalf("Post request failed with %q", err)
		}
		time.Sleep(time.Second * 2)
	}
}

func main() {
	go producer()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/server", server)

	if err := http.ListenAndServe(getUrl(), nil); err != nil {
		log.Fatalf("Losten to %s failed with %q", getUrl(), err)
	}
}
