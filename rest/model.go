package rest

const (
	SUCCESS = "success"
	ERROR   = "error"
)

// Movie Service Request objects

type Request struct {
	RequestId    string `json:"request_id,omitempty"`
	ExchangeName string `json:"exchange_name,omitempty"`
	RoutingKey   string `json:"routing_key,omitempty"`
}

type Response struct {
	RequestId    string `json:"request_id,omitempty"`
	Method       string `json:"method,omitempty"`
	Query        string `json:"query,omitempty"`
	ExchangeName string `json:"exchange_name,omitempty"`
	RoutingKey   string `json:"routing_key,omitempty"`
}

// Movie MQ Response objects
type Meta struct {
	RequestId string `json:"request_id,omitempty"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

type SearchData struct {
	Movies []Movie `json:"movies"`
}

type SearchResponse struct {
	Meta Meta       `json:"meta"`
	Data SearchData `json:"data"`
}

// Movie Service objects, (DTO)

type Ratings struct {
	CriticsRating  string
	CriticsScore   int
	AudienceRating string
	AudienceScore  int
}

type Movie struct {
	Id      string
	Title   string
	Ratings Ratings
}

func NewSearchResponseSuccess(requestId string, movies []Movie) *SearchResponse {
	return &SearchResponse{Meta: Meta{RequestId: requestId, Status: SUCCESS}, Data: SearchData{movies}}
}

func NewSearchResponseError(requestId string, err error) *SearchResponse {
	return &SearchResponse{Meta: Meta{RequestId: requestId, Status: ERROR, Error: err.Error()}}
}
