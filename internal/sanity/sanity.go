package sanity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type SocialLink struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Icon  struct {
		Type string `json:"_type"`
		Name string `json:"name"`
	} `json:"icon"`
}

type Card struct {
	Title       string        `json:"title"`
	Description []interface{} `json:"description"`
}

type Homepage struct {
	Title         string        `json:"title"`
	Tagline       string        `json:"tagline"`
	AboutTitle    string        `json:"aboutTitle"`
	AboutContent  []interface{} `json:"aboutContent"`
	GoalsTitle    string        `json:"goalsTitle"`
	Goals         []Card        `json:"goals"`
	ProjectsTitle string        `json:"projectsTitle"`
	Projects      []Card        `json:"projects"`
	SocialLinks   []SocialLink  `json:"socialLinks"`
}

type Client struct {
	ProjectID  string
	Dataset    string
	HTTPClient *http.Client
}

func NewClient(projectID, dataset string) *Client {
	return &Client{
		ProjectID:  projectID,
		Dataset:    dataset,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) FetchHomepage() (Homepage, error) {
	baseURL := fmt.Sprintf("https://%s.api.sanity.io/v2021-06-07/data/query/%s", c.ProjectID, c.Dataset)

	query := `*[_type == "homepage"][0]{
		title,
		tagline,
		aboutTitle,
		"aboutContent": aboutContent,
		goalsTitle,
		"goals": goals[]{
			title,
			"description": description
		},
		projectsTitle,
		"projects": projects[]{
			title,
			"description": description
		},
		"socialLinks": socialLinks[]{
			title,
			url,
			"icon": icon
		}
	}`

	values := url.Values{}
	values.Add("query", query)
	reqURL := baseURL + "?" + values.Encode()

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return Homepage{}, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Homepage{}, err
	}
	defer resp.Body.Close()

	var result struct {
		Result Homepage `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Homepage{}, err
	}

	return result.Result, nil
}
