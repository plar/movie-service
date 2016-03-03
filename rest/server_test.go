package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type errorReader struct{}

func (self *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("Cannot read from the buffer")
}

func (self *errorReader) Close() error {
	return nil
}

func NewTestMovieServerContext() MovieServerContext {
	return MovieServerContext{
		JobFactory:           NewTestJobFactory(),
		RottenTomatoesAPIKey: "APIKEY",
	}
}

func TestCreateMovieServerApiKeyRequired(t *testing.T) {
	ctx := MovieServerContext{}
	server, err := NewMovieServer(ctx)
	assert.Nil(t, server)
	assert.EqualError(t, err, "RottenTomatoesAPIKey is required")
}

func TestCreateMovieServer(t *testing.T) {
	ctx := NewTestMovieServerContext()
	server, err := NewMovieServer(ctx)
	assert.NotNil(t, server)
	assert.NoError(t, err)
}

func TestMovieServerSearch(t *testing.T) {
	ctx := NewTestMovieServerContext()
	server, _ := NewMovieServer(ctx)
	recorder := httptest.NewRecorder()

	reqBody, err := json.Marshal(Request{
		RequestId:    "unique-request-id",
		ExchangeName: "ExchangeName",
		RoutingKey:   "RoutingKey",
	})

	req, err := http.NewRequest("POST", "http://movie-search.devel/movies", bytes.NewReader(reqBody))
	query := req.URL.Query()
	query.Add("q", "martian")
	req.URL.RawQuery = query.Encode()

	server.Router().ServeHTTP(recorder, req)

	resp := Response{}

	err = json.Unmarshal(recorder.Body.Bytes(), &resp)
	assert.NoError(t, err)

	expected := Response{
		RequestId:    "unique-request-id",
		Method:       "movies",
		Query:        "martian",
		ExchangeName: "ExchangeName",
		RoutingKey:   "RoutingKey",
	}
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, expected, resp)
}

func TestMovieServerSearchEmptyQuery(t *testing.T) {
	ctx := NewTestMovieServerContext()
	server, _ := NewMovieServer(ctx)
	recorder := httptest.NewRecorder()

	reqBody, err := json.Marshal(Request{ExchangeName: "ExchangeName", RoutingKey: "RoutingKey"})
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "http://movie-search.devel/movies", bytes.NewReader(reqBody))
	assert.NoError(t, err)

	query := req.URL.Query()
	query.Add("q", "")
	req.URL.RawQuery = query.Encode()

	server.Router().ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	body, _ := ioutil.ReadAll(recorder.Body)
	assert.Equal(t, "Query cannot be empty\n", string(body))
}

func TestMovieServerSearchEmptyBody(t *testing.T) {
	ctx := NewTestMovieServerContext()
	server, _ := NewMovieServer(ctx)
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "http://movie-search.devel/movies", strings.NewReader(""))
	req.Body = &errorReader{}
	assert.NoError(t, err)

	query := req.URL.Query()
	query.Add("q", "query")
	req.URL.RawQuery = query.Encode()

	server.Router().ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	body, _ := ioutil.ReadAll(recorder.Body)
	assert.Equal(t, "Cannot read request body\n", string(body))
}

func TestMovieServerSearchWrongBody(t *testing.T) {
	ctx := NewTestMovieServerContext()
	server, _ := NewMovieServer(ctx)
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "http://movie-search.devel/movies", strings.NewReader(""))
	assert.NoError(t, err)

	query := req.URL.Query()
	query.Add("q", "query")
	req.URL.RawQuery = query.Encode()

	server.Router().ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	body, _ := ioutil.ReadAll(recorder.Body)
	assert.Equal(t, "Cannot decode request body: unexpected end of JSON input\n", string(body))
}
