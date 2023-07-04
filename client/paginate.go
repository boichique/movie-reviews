package client

import "github.com/boichique/movie-reviews/contracts"

func Paginate[I any, Req contracts.PaginationSetter](
	req Req,
	queryFn func(Req) (*contracts.PaginatedResponse[I], error),
) ([]*I, error) {
	var items []*I

	for {
		res, err := queryFn(req)
		if err != nil {
			return nil, err
		}
		items = append(items, res.Items...)

		if len(items) >= res.Total {
			break
		}
		req.SetPage(res.Page + 1)
	}
	return items, nil
}
