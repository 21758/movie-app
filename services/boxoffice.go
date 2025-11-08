package services

import (
	"encoding/json"
	"fmt"
	"io"
	"movie-api/models"
	"net/http"
	"time"
)

type BoxOfficeService struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewBoxOfficeService(baseURL, apiKey string) *BoxOfficeService {
	return &BoxOfficeService{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type BoxOfficeAPIResponse struct {
	Title       string `json:"title"`
	Distributor string `json:"distributor"`
	ReleaseDate string `json:"releaseDate"`
	Budget      int64  `json:"budget"`
	Revenue     struct {
		Worldwide         int64 `json:"worldwide"`
		OpeningWeekendUSA int64 `json:"openingWeekendUSA"`
	} `json:"revenue"`
	MpaRating string `json:"mpaRating"`
}

func (s *BoxOfficeService) GetMovieData(title string) (*models.BoxOffice, error) {
	url := fmt.Sprintf("%s/boxoffice?title=%s", s.baseURL, title)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("movie not found in box office API")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("box office API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse BoxOfficeAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	boxOffice := &models.BoxOffice{}
	boxOffice.Revenue.Worldwide = apiResponse.Revenue.Worldwide
	boxOffice.Revenue.OpeningWeekendUSA = apiResponse.Revenue.OpeningWeekendUSA
	boxOffice.Currency = "USD"
	boxOffice.Source = "BoxOfficeAPI"
	boxOffice.LastUpdated = time.Now().UTC()

	return boxOffice, nil
}
