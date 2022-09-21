package queryassistor

type queryBox struct {
	req        IReqQuery
	Query      *Query
	Filters    map[string][]FilterItem
	Having     map[string][]FilterItem
	BindValues []interface{}
	parsed     bool
	cached     map[string]string
	totalModel bool
}

func newQueryBox(req IReqQuery) *queryBox {
	q := req.Query()
	q.Id = "t"
	rq := queryBox{
		req:     req,
		Query:   &q,
		Filters: map[string][]FilterItem{},
		Having:  map[string][]FilterItem{},
	}
	rq.Query = parseTag(&rq)
	return &rq
}

func (rq *queryBox) PageSize() int {
	if rq.Query.PageSize <= 0 {
		return 10
	}
	return rq.Query.PageSize
}

func (rq *queryBox) SetPageSize(size int) {
	rq.Query.PageSize = size
}

func (rq *queryBox) Order() string {
	return rq.Query.Order
}
