// This code is copypasted from the tonutils-storage-provider package and modified.
// Original package:
// https://github.com/xssnick/tonutils-storage-provider
// According to the license, this code is licensed under the Apache License 2.0
// See the LICENSE file in the original package for more details.
package tonstorage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ErrNotFound = errors.New("not found")

type Client interface {
	GetBag(ctx context.Context, bagId string) (*BagDetailed, error)
}

type client struct {
	base        string
	client      http.Client
	credentials *Credentials
}

func (c *client) GetBag(ctx context.Context, bagId string) (*BagDetailed, error) {
	var res BagDetailed
	if err := c.doRequest(ctx, "GET", "/api/v1/details?bag_id="+bagId, nil, &res); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	if res.InfoLoaded && res.MerkleHash == "" {
		return nil, fmt.Errorf("too old tonutils-storage version, please update")
	}
	return &res, nil
}

func (c *client) doRequest(ctx context.Context, method, url string, req, resp any) error {
	buf := &bytes.Buffer{}
	if req != nil {
		if err := json.NewEncoder(buf).Encode(req); err != nil {
			return fmt.Errorf("failed to encode request data: %w", err)
		}
	}

	r, err := http.NewRequestWithContext(ctx, method, c.base+url, buf)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	if c.credentials != nil {
		r.SetBasicAuth(c.credentials.Login, c.credentials.Password)
	}

	res, err := c.client.Do(r)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return ErrNotFound
	}

	if res.StatusCode != 200 {
		var e Result
		if err = json.NewDecoder(res.Body).Decode(&e); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
		return fmt.Errorf("status code is %d, error: %s", res.StatusCode, e.Error)
	}

	if err = json.NewDecoder(res.Body).Decode(resp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

func NewClient(base string, credentials *Credentials) Client {
	return &client{
		base: base,
		client: http.Client{
			Timeout: 15 * time.Second,
		},
		credentials: credentials,
	}
}
