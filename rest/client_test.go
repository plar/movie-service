package rest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func httpTestClient(httpCode int, body []byte) (*httptest.Server, *http.Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpCode)
		w.Write(body)
	}))

	tr := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}
	client := &http.Client{Transport: tr}

	return server, client
}

func TestCreateClient(t *testing.T) {
	client := NewClientWithHttp(nil, "APIKEY")
	assert.NotNil(t, client)
}

func TestClientSearchMartian(t *testing.T) {

	// api.rottentomatoes.com returns wrong 'total' number for the martian movie 'total' != len(movies) in the movies-martian.json file
	fixture, err := ioutil.ReadFile("../fixtures/movies-martian.json")
	if err != nil {
		t.Error("Cannot read fixtures \"../fixtures/movies-martian.json\"")
		return
	}

	server, httpClient := httpTestClient(http.StatusOK, fixture)
	defer server.Close()

	client := NewClientWithHttp(httpClient, "APIKEY")
	movies, err := client.Search("Martian")
	assert.NoError(t, err)
	assert.Equal(t, 21, len(movies))

	assert.ObjectsAreEqual(movies[0], Movie{
		Id:    "771380589",
		Title: "The Martian",
		Ratings: Ratings{
			CriticsRating:  "Certified Fresh",
			CriticsScore:   92,
			AudienceRating: "Upright",
			AudienceScore:  92,
		},
	})
}

func TestClientSearchEmpty(t *testing.T) {

	fixture, err := ioutil.ReadFile("../fixtures/movies-empty.json")
	if err != nil {
		t.Error("Cannot read fixtures \"../fixtures/movies-empty.json\"")
		return
	}

	server, httpClient := httpTestClient(http.StatusOK, fixture)
	defer server.Close()

	client := NewClientWithHttp(httpClient, "APIKEY")
	movies, err := client.Search("xxxxxxxxxxxxxxxxxxxx")
	assert.Nil(t, movies)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(movies))
}

func TestClientSearchError(t *testing.T) {

	server, httpClient := httpTestClient(http.StatusNotFound, []byte("non-json-body"))
	defer server.Close()

	client := NewClientWithHttp(httpClient, "APIKEY")
	movies, err := client.Search("Martian")
	assert.Nil(t, movies)
	assert.Error(t, err)
	assert.EqualError(t, err, "api error, response code: 404")
}
