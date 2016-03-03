package rest

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	wq "github.com/plar/movie-service/workerqueue"
)

type testmqAndClientImpl struct {
	simulateSearchError error

	query string
	req   *Request
	resp  *SearchResponse
}

func (self *testmqAndClientImpl) Search(query string) ([]Movie, error) {
	self.query = query

	if self.simulateSearchError != nil {
		return nil, self.simulateSearchError
	} else {
		movies := []Movie{
			Movie{Id: "Id", Title: "Title", Ratings: Ratings{
				CriticsRating:  "CriticsRating",
				CriticsScore:   111,
				AudienceRating: "AudienceRating",
				AudienceScore:  222,
			}}}
		return movies, nil
	}
}

func (self *testmqAndClientImpl) PublishSearchResponse(req *Request, resp *SearchResponse) error {
	self.req = req
	self.resp = resp
	return nil
}

func TestNewJobFactory(t *testing.T) {
	mqAndClient := &testmqAndClientImpl{}
	workerQueue := make(wq.WorkerQueue, 1)

	factory := NewJobFactory(mqAndClient, mqAndClient, workerQueue)
	assert.NotNil(t, factory)

	factoryImpl, ok := factory.(*jobFactory)
	assert.True(t, ok)
	assert.Equal(t, mqAndClient, factoryImpl.messageQueue)
	assert.Equal(t, mqAndClient, factoryImpl.client)
	assert.Equal(t, workerQueue, factoryImpl.workerQueue)
}

func TestNewSearch(t *testing.T) {

	mqAndClient := &testmqAndClientImpl{}
	workerQueue := make(wq.WorkerQueue, 1)
	worker, _ := wq.NewWorker(1, workerQueue)
	worker.Start()

	factory := NewJobFactory(mqAndClient, mqAndClient, workerQueue)

	req := Request{RequestId: "RequestId", ExchangeName: "ExchangeName", RoutingKey: "RoutingKey"}
	factory.NewSearch(req, "test-query")

	// wait for finish
FINISH:
	for {
		select {
		case <-time.After(10 * time.Second):
			assert.Fail(t, "Cannot stop worker")
			return
		default:
			worker.Stop()
			worker.WaitForFinish()
			break FINISH
		}
	}

	assert.Equal(t, mqAndClient.query, "test-query")
	assert.Equal(t, 1, len(mqAndClient.resp.Data.Movies))
	assert.Equal(t, req, *mqAndClient.req)
	assert.Equal(t, mqAndClient.resp.Meta.RequestId, req.RequestId)
	assert.Equal(t, mqAndClient.resp.Meta.Error, "")
	assert.Equal(t, mqAndClient.resp.Meta.Status, SUCCESS)
	assert.Equal(t, mqAndClient.resp.Data.Movies[0],
		Movie{Id: "Id", Title: "Title", Ratings: Ratings{
			CriticsRating:  "CriticsRating",
			CriticsScore:   111,
			AudienceRating: "AudienceRating",
			AudienceScore:  222,
		}})
}

func TestNewSearchWithError(t *testing.T) {
	mqAndClient := &testmqAndClientImpl{simulateSearchError: errors.New("API is not available")}
	workerQueue := make(wq.WorkerQueue, 1)
	worker, _ := wq.NewWorker(1, workerQueue)
	worker.Start()

	factory := NewJobFactory(mqAndClient, mqAndClient, workerQueue)

	req := Request{RequestId: "RequestId", ExchangeName: "ExchangeName", RoutingKey: "RoutingKey"}
	factory.NewSearch(req, "test-query")

	// wait for finish
FINISH:
	for {
		select {
		case <-time.After(10 * time.Second):
			assert.Fail(t, "Cannot stop worker")
			return
		default:
			worker.Stop()
			worker.WaitForFinish()
			break FINISH
		}
	}

	assert.Equal(t, mqAndClient.query, "test-query")
	assert.Equal(t, 0, len(mqAndClient.resp.Data.Movies))
	assert.Equal(t, req, *mqAndClient.req)
	assert.Equal(t, mqAndClient.resp.Meta.RequestId, req.RequestId)
	assert.Equal(t, mqAndClient.resp.Meta.Error, "API is not available")
	assert.Equal(t, mqAndClient.resp.Meta.Status, ERROR)

}
