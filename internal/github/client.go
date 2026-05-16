package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/cli/go-gh/v2/pkg/api"
)

// NewRESTClient returns an authenticated GitHub REST API client using
// the active gh credentials.
func NewRESTClient() (*api.RESTClient, error) {
	return api.DefaultRESTClient()
}

func jsonBody(v interface{}) (io.Reader, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}
	return bytes.NewReader(data), nil
}
