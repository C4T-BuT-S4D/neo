package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func NewFarmClient(cfg FarmConfig) *FarmClient {
	return &FarmClient{
		cfg.Url,
		cfg.Password,
		http.Client{
			Timeout: time.Second * 3,
		},
	}
}

type FarmClient struct {
	url      string
	password string
	client   http.Client
}

func (fc *FarmClient) FillConfig(cfg *FarmConfig) error {
	url := fmt.Sprintf("%s/api/get_config", fc.url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Authorization", fc.password)
	resp, err := fc.client.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Errorf("Error closing farm client response body: %v", err)
		}
	}()
	if err := cfg.ParseJson(resp.Body); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return nil
}
