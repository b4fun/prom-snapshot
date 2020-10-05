package promclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type httpClient struct {
	baseURL *url.URL
	client  *http.Client
	logger  logrus.FieldLogger
}

var _ Client = (*httpClient)(nil)

// example: {"status":"success","data":{"name":"20201005T145700Z-380704bb7b4d7c03"}}
type promCreateSnapshotResponse struct {
	Status string `json:"status"`
	Data   struct {
		Name string `json:"name"`
	} `json:"data"`
}

func (c *httpClient) CreateSnapshot(ctx context.Context) (*CreateSnapshotResponse, error) {
	logger := c.logger.WithField("action", "CreateSnapshot")

	u, err := url.Parse(c.baseURL.String())
	if err != nil {
		return nil, fmt.Errorf("build url: %w", err)
	}
	u.Path = path.Join(u.Path, "/api/v1/admin/tsdb/snapshot")
	c.logger.Debugf("request url: %s", u.String())

	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request snapshot api: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		// TODO: dump response
		logger.Errorf("request failed: statusCode=%d", resp.StatusCode)
		return nil, fmt.Errorf("request failed: statusCode=%d", resp.StatusCode)
	}

	var respBody promCreateSnapshotResponse
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&respBody); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	logger.Debugf("response body: %q", respBody)

	if !strings.EqualFold(respBody.Status, "success") {
		logger.Warnf("response status != success: %q", respBody)
		return nil, fmt.Errorf("response failed: status=%s", respBody.Status)
	}

	return &CreateSnapshotResponse{
		SnapshotName: respBody.Data.Name,
	}, nil
}

// ClientOption sets the option for client.
type ClientOption struct {
	// BaseURL sets the base url for the prometheus api.
	BaseURL string

	// HTTPClient sets the HTTP client to use.
	HTTPClient *http.Client

	// Logger sets the logger to use.
	Logger logrus.FieldLogger
}

// Create creates the client instance.
func (o ClientOption) Create() (Client, error) {
	baseURL, err := url.Parse(o.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %s: %w", o.BaseURL, err)
	}

	rv := &httpClient{
		baseURL: baseURL,
		client:  o.HTTPClient,
		logger:  o.Logger,
	}

	if rv.client == nil {
		rv.client = &http.Client{
			Timeout: 3 * time.Minute,
		}
	}

	if rv.logger == nil {
		logger := logrus.New()
		rv.logger = logger
	}

	return rv, nil
}
