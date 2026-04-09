package api

import (
	"fmt"
	"net/url"
)

const pageSize = 20

type ListResponse[T any] struct {
	Data  []T `json:"data"`
	Total int `json:"total"`
}

// PaginateN fetches up to maxItems results. If maxItems <= 0, fetches all.
func PaginateN[T any](c *Client, path string, params url.Values, maxItems int) ([]T, error) {
	if maxItems <= 0 {
		return Paginate[T](c, path, params)
	}
	if params == nil {
		params = url.Values{}
	}

	var all []T
	offset := 0
	for {
		batchSize := pageSize
		remaining := maxItems - len(all)
		if remaining < batchSize {
			batchSize = remaining
		}

		p := url.Values{}
		for k, v := range params {
			p[k] = v
		}
		p.Set("limit", fmt.Sprintf("%d", batchSize))
		p.Set("offset", fmt.Sprintf("%d", offset))

		var resp ListResponse[T]
		if err := c.Get(path, p, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Data...)
		if len(all) >= maxItems || len(all) >= resp.Total {
			break
		}
		offset += batchSize
	}
	return all, nil
}

func Paginate[T any](c *Client, path string, params url.Values) ([]T, error) {
	if params == nil {
		params = url.Values{}
	}

	var all []T
	offset := 0
	for {
		p := url.Values{}
		for k, v := range params {
			p[k] = v
		}
		p.Set("limit", fmt.Sprintf("%d", pageSize))
		p.Set("offset", fmt.Sprintf("%d", offset))

		var resp ListResponse[T]
		if err := c.Get(path, p, &resp); err != nil {
			return nil, err
		}
		all = append(all, resp.Data...)
		if len(all) >= resp.Total {
			break
		}
		offset += pageSize
	}
	return all, nil
}
