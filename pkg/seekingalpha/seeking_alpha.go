package seekingalpha

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"github.com/nthskyradiated/openpricegap-analysis/internal/news"
)


var urlPath = "/news/v2/list-by-symbol"
var apiKeyHeader = os.Getenv("APIKEYHEADER")
var pageSize = os.Getenv("PAGESIZE")

type client struct {
	baseUrl string
	apiKey string
}

func (c *client) Fetch(ticker string) ([]news.Article, error) {

	url, err := c.buildURL(ticker)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(apiKeyHeader, c.apiKey)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil, fmt.Errorf("unsuccessful status code %d received", response.StatusCode)
	}

	return c.parse(response)	 
}

func (c *client) parse(resp *http.Response) ([]news.Article, error) {
	res := 	&SeekingAlphaResponse{}
	err := json.NewDecoder(resp.Body).Decode(res)
	if err != nil {
		return nil, err
	}

	var articles []news.Article
	for _, item := range res.Data {
		art := news.Article {
			PublishOn: item.Attributes.PublishOn,
			Headline: item.Attributes.Title,
		}
		articles = append(articles, art)

	}
	return articles, nil
}

func (c *client) buildURL(ticker string) (string, error) {

	parsedURL, err := url.Parse(c.baseUrl)
	if err != nil {
		return "", err
	}

	parsedURL.Path += urlPath

	params:= url.Values{}
	params.Add("size", fmt.Sprint(pageSize))
	params.Add("id", ticker)
	parsedURL.RawQuery = params.Encode()

	return parsedURL.String(), nil
}

func NewClient(baseURL, apiKey string) news.Fetcher {
	return &client{baseUrl: baseURL, apiKey: apiKey}
}

func init() {
	// Load the environment variables at the beginning
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Now fetch the environment variables after loading
	apiKeyHeader = os.Getenv("APIKEYHEADER")
	pageSize = os.Getenv("PAGESIZE")
}