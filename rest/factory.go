package rest

import wq "github.com/plar/movie-service/workerqueue"

type JobFactory interface {
	NewSearch(req Request, query string)
}

type jobFactory struct {
	messageQueue MessageQueue
	client       Client
	workerQueue  wq.WorkerQueue
}

type testJobFactory struct {
}

func (self *jobFactory) NewSearch(req Request, query string) {
	worker := <-self.workerQueue
	worker <- func(id int) {
		var resp *SearchResponse
		movies, err := self.client.Search(query)
		if err == nil {
			resp = NewSearchResponseSuccess(req.RequestId, movies)
		} else {
			resp = NewSearchResponseError(req.RequestId, err)
		}
		self.messageQueue.PublishSearchResponse(&req, resp)
	}
}

func (self *testJobFactory) NewSearch(req Request, query string) {
}

func NewJobFactory(mq MessageQueue, client Client, workerQueue wq.WorkerQueue) JobFactory {
	return &jobFactory{mq, client, workerQueue}
}

func NewTestJobFactory() JobFactory {
	return &testJobFactory{}
}
