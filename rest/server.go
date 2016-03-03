package rest

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	wq "github.com/plar/movie-service/workerqueue"

	"github.com/gorilla/mux"
)

const (
	totalWorkers = 10
)

type MovieServerContext struct {
	RottenTomatoesAPIKey string
	Logger               *log.Logger
	Client               Client
}

type MovieServer interface {
	Search(w http.ResponseWriter, r *http.Request)
	FullCast(w http.ResponseWriter, r *http.Request)

	Router() *mux.Router

	Start()
	Stop()
}

type movieServer struct {
	client Client

	router *mux.Router
	logger *log.Logger

	workerQueue wq.WorkerQueue
	workers     []wq.Worker
}

func (self *movieServer) newSearch(req Request, query string) {
	worker := <-self.workerQueue
	worker <- func(id int) {
		// Do real job here
		// 1. Make request to rottentomatoes
		movies, err := self.client.Search(query)
		if err != nil {
			resp := NewSearchResponseError(req.RequestId, err)
			var _ = resp
			return
		}

		// 2. Send movies back to exchangeName/routingKey
		resp := NewSearchResponseSuccess(req.RequestId, movies)
		var _ = resp
	}
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
		http.Error(w, "Cannot decode request body", http.StatusBadRequest)
		return
	}

	// create response
	resp := Response{
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
	self.newSearch(req, query)
}

func (self *movieServer) FullCast(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (self *movieServer) Router() *mux.Router {
	return self.router
}

func (self *movieServer) Start() {
	self.workerQueue = make(wq.WorkerQueue, totalWorkers)

	self.workers = make([]wq.Worker, totalWorkers)
	for i := range self.workers {
		self.workers[i] = wq.NewWorker(i, self.workerQueue)
		self.workers[i].Start()
	}
}

func (self *movieServer) Stop() {
	for i := range self.workers {
		self.workers[i].Stop()
	}

	// wait for the workers to quit
	for i := range self.workers {
		self.workers[i].WaitForFinish()
	}
}

func NewMovieServer(ctx MovieServerContext) (MovieServer, error) {
	if len(ctx.RottenTomatoesAPIKey) == 0 {
		return nil, errors.New("RottenTomatoesAPIKey is required")
	}

	client := ctx.Client
	if client == nil {
		client = NewClient(ctx.RottenTomatoesAPIKey)
	}

	logger := ctx.Logger
	if logger == nil {
		logger = log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)
	}

	server := &movieServer{
		client: client,
		logger: logger,
	}

	server.router = mux.NewRouter()
	server.router.HandleFunc("/movies", http.HandlerFunc(server.Search)).Methods("POST").Queries("q", "{q}")
	server.router.HandleFunc("/movie/{id}/full_cast", http.HandlerFunc(server.FullCast)).Methods("POST")

	server.Start()

	return server, nil
}
