package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"go-weather-cli/internal/models"
)

const (
	baseURL = "https://api.openweathermap.org/data/2.5/weather"

	dialTimeout         = 5 * time.Second
	tlsHandshakeTimeout = 5 * time.Second
	responseTimeout     = 10 * time.Second
	keepAlive           = 30 * time.Second
	maxIdleConns        = 10
	idleConnTimeout     = 90 * time.Second
)

type WeatherClient struct {
	httpClient *http.Client
	apiKey     string
}

func New(apiKey string) *WeatherClient {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   dialTimeout,
			KeepAlive: keepAlive,
		}).DialContext,
		TLSHandshakeTimeout: tlsHandshakeTimeout,
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConns,
		IdleConnTimeout:     idleConnTimeout,
		DisableCompression:  true,
		ForceAttemptHTTP2:   true,
	}

	return &WeatherClient{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   responseTimeout,
		},
		apiKey: apiKey,
	}
}

func (wc *WeatherClient) FetchByCity(ctx context.Context, cityName, units string) (*models.WeatherResponse, error) {
	reqURL, err := wc.buildURL(cityName, units)
	if err != nil {
		return nil, fmt.Errorf("building request URL: %w", err)
	}

	return wc.doRequest(ctx, reqURL)
}

func (wc *WeatherClient) FetchByCoords(ctx context.Context, lat, lon float64, units string) (*models.WeatherResponse, error) {
	reqURL, err := wc.buildCoordsURL(lat, lon, units)
	if err != nil {
		return nil, fmt.Errorf("building coords URL: %w", err)
	}

	return wc.doRequest(ctx, reqURL)
}

func (wc *WeatherClient) doRequest(ctx context.Context, reqURL string) (*models.WeatherResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "AeroStat/1.0")

	resp, err := wc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, wc.handleAPIError(resp)
	}

	var weather models.WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return nil, fmt.Errorf("decoding API response: %w", err)
	}

	return &weather, nil
}

func (wc *WeatherClient) buildURL(city, units string) (string, error) {
	params := url.Values{}
	params.Set("q", city)
	params.Set("appid", wc.apiKey)
	params.Set("units", units)
	return baseURL + "?" + params.Encode(), nil
}

func (wc *WeatherClient) buildCoordsURL(lat, lon float64, units string) (string, error) {
	params := url.Values{}
	params.Set("lat", fmt.Sprintf("%.4f", lat))
	params.Set("lon", fmt.Sprintf("%.4f", lon))
	params.Set("appid", wc.apiKey)
	params.Set("units", units)
	return baseURL + "?" + params.Encode(), nil
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

type apiErrorBody struct {
	Message string `json:"message"`
}

func (wc *WeatherClient) handleAPIError(resp *http.Response) error {
	var body apiErrorBody
	_ = json.NewDecoder(resp.Body).Decode(&body)

	msg := body.Message
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}
	return &APIError{StatusCode: resp.StatusCode, Message: msg}
}
