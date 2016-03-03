package rest

import (
	"net/http"
	"strconv"

	"github.com/rojters/rottentomatoes"
)

type Client interface {
	Search(query string) ([]Movie, error)
}

type client struct {
	client *rottentomatoes.Client
}

func rottenRatingsToRatings(r rottentomatoes.Ratings) Ratings {
	return Ratings{
		AudienceRating: r.AudienceRating,
		AudienceScore:  r.AudienceScore,
		CriticsRating:  r.CriticsRating,
		CriticsScore:   r.CriticsScore,
	}
}

func (c *client) Search(query string) ([]Movie, error) {
	resp, err := c.client.Search.MovieSearch(query, nil)
	if err != nil {
		return nil, err
	}

	if resp.Total == 0 {
		return nil, nil
	}

	var movies []Movie
	for _, movie := range resp.Movies {
		movies = append(movies, Movie{
			Id:      strconv.FormatInt(int64(movie.Id), 10),
			Title:   movie.Title,
			Ratings: rottenRatingsToRatings(movie.Ratings),
		})
	}

	return movies, nil
}

func NewClientWithHttp(httpClient *http.Client, apiKey string) Client {
	c := &client{rottentomatoes.NewClient(httpClient, apiKey)}
	return c
}

func NewClient(apiKey string) Client {
	return NewClientWithHttp(nil, apiKey)
}
