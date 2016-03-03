package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"gopkg.in/tylerb/graceful.v1"

	log "github.com/cihub/seelog"
	"github.com/streadway/amqp"

	wq "github.com/plar/movie-service/workerqueue"

	"github.com/gorilla/mux"
)

const (
	totalWorkers = 10
)

type MovieServerContext struct {
	MessageQueueURI      string
	ServiceURI           string
	RottenTomatoesAPIKey string
	Client               Client
	JobFactory           JobFactory
}

type MessageQueue interface {
	PublishSearchResponse(req *Request, resp *SearchResponse) error
}

type MovieServer interface {
	MessageQueue

	Search(w http.ResponseWriter, r *http.Request)
	FullCast(w http.ResponseWriter, r *http.Request)

	Router() *mux.Router

	Start()

	Quit()
}

type movieServer struct {
	messageQueueURI string
	serviceURI      string
	client          Client
	router          *mux.Router
	jobFactory      JobFactory
	workerQueue     wq.WorkerQueue
	workers         []wq.Worker
	quit            chan bool
}

func (self *movieServer) PublishSearchResponse(req *Request, resp *SearchResponse) error {
	// TBD: Connect during service start...
	conn, err := amqp.Dial(self.messageQueueURI)
	if err != nil {
		log.Errorf("Cannot connect to MessageQueue, uri=%s, error=%s", self.messageQueueURI, err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("Cannot open channel, error=%s", err)
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		req.ExchangeName,
		"topic", // type
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Errorf("Cannot declare exchange, error=%s", err)
		return err
	}

	body, err := json.Marshal(*resp)
	if err != nil {
		log.Errorf("Cannot encode response body, error=%s", err)
		return err
	}

	err = ch.Publish(
		req.ExchangeName,
		req.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		log.Errorf("Cannot publish message, error=%s", err)
		return err
	}

	return nil
}

func (self *movieServer) Search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	query, exists := vars["q"]
	if !exists || len(query) == 0 {
		http.Error(w, "Query cannot be empty", http.StatusBadRequest)
		return
	}

	// read POST Body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Cannot read request body", http.StatusInternalServerError)
		return
	}

	var req Request
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot decode request body: %v", err), http.StatusBadRequest)
		return
	}

	// create response
	resp := Response{
		RequestId:    req.RequestId,
		Method:       "movies",
		Query:        query,
		ExchangeName: req.ExchangeName,
		RoutingKey:   req.RoutingKey,
	}

	body, err = json.Marshal(resp)
	if err != nil {
		http.Error(w, "Cannot encode response body", http.StatusInternalServerError)
		return
	}
	w.Write(body)

	// send query to the workerpool
	self.jobFactory.NewSearch(req, query)
}

func (self *movieServer) FullCast(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (self *movieServer) Router() *mux.Router {
	return self.router
}

func (self *movieServer) Start() {
	log.Infof("Welcome to Movie Service!")
	log.Infof("Listen on %s\nPress Ctrl-C to quit...\n", self.serviceURI)
	graceful.Run(self.serviceURI, 10*time.Second, self.router)
	self.waitForQuit()
}

func (self *movieServer) Quit() {
	log.Infof("Stop workers...")
	for i := range self.workers {
		self.workers[i].Stop()
	}

	log.Infof("Wait for the workers to quit...")
	for i := range self.workers {
		self.workers[i].WaitForFinish()
	}

	log.Infof("Stop service (10s timeout)...")
	self.quit <- true
}

func (self *movieServer) waitForQuit() {
	for {
		select {
		case <-self.quit:
			log.Infof("Bye-bye")
			return
		}
	}
}

func NewMovieServer(ctx MovieServerContext) (MovieServer, error) {
	if len(ctx.RottenTomatoesAPIKey) == 0 {
		return nil, errors.New("RottenTomatoesAPIKey is required")
	}

	serviceURI := ctx.ServiceURI
	if len(serviceURI) == 0 {
		serviceURI = "127.0.0.1:12345"
	}

	client := ctx.Client
	if client == nil {
		client = NewClient(ctx.RottenTomatoesAPIKey)
	}

	server := &movieServer{
		messageQueueURI: ctx.MessageQueueURI,
		serviceURI:      serviceURI,
		client:          client,
		workerQueue:     make(wq.WorkerQueue, totalWorkers),
		quit:            make(chan bool, 1),
	}

	server.workers = make([]wq.Worker, totalWorkers)
	for i := range server.workers {
		server.workers[i] = wq.NewWorker(i, server.workerQueue)
		server.workers[i].Start()
	}

	jobFactory := ctx.JobFactory
	if jobFactory == nil {
		jobFactory = NewJobFactory(server, server.client, server.workerQueue)
	}
	server.jobFactory = jobFactory

	server.router = mux.NewRouter()
	server.router.HandleFunc("/movies", http.HandlerFunc(server.Search)).Methods("POST").Queries("q", "{q}")
	server.router.HandleFunc("/movie/{id}/full_cast", http.HandlerFunc(server.FullCast)).Methods("POST")

	return server, nil
}
