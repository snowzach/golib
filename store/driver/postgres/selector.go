package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/snowzach/queryp"
	"github.com/snowzach/queryp/qppg"

	"github.com/snowzach/golib/store"
)

type Scanner[T any] interface {
	Scan() *T
}

// Selector is a tool for fetching slices of records based on any query.
type Selector[T any] struct {
	// The query to use
	Query string
	// OmitCount will not run the count query
	OmitCount bool
	// CountQuery allows you to override the query used to obtain counts with the same filter parameters.
	// If not specified it will wrap the Query provided in select count(*)
	CountQuery string

	// The fields you want to allow filtering on and the types
	FilterFieldTypes queryp.FilterFieldTypes
	// The fields you want to allow sorting on
	SortFields queryp.SortFields
	// If no sort is provided in the QueryParameters, the default sort to use.
	DefaultSort queryp.Sort

	// A callback to be used on every record to provide any transformations.
	PostProcessRecord func(*T) error
	// A callback to be used on the slice of records being returned before they are returned.
	PostProcessRecords func([]*T) error
}

func (s *Selector[T]) Select(ctx context.Context, db DB, qp *queryp.QueryParameters) ([]*T, *int64, error) {
	var query strings.Builder
	var queryParams []any

	query.WriteString(s.Query)

	if len(qp.Sort) == 0 && len(s.DefaultSort) > 0 {
		qp.Sort = s.DefaultSort
	}

	if len(qp.Filter) > 0 {
		query.WriteString(" WHERE ")
	}

	if err := qppg.FilterQuery(s.FilterFieldTypes, qp.Filter, &query, &queryParams); err != nil {
		return nil, nil, &store.Error{Type: store.ErrorTypeQuery, Err: err}
	}
	var count *int64
	if !s.OmitCount {
		count = new(int64)
		var countQuery strings.Builder
		countQuery.WriteString(`SELECT COUNT(*) AS count FROM (`)
		if s.CountQuery != "" {
			countQuery.WriteString(s.CountQuery)
		} else {
			countQuery.WriteString(query.String())
		}
		countQuery.WriteString(`) _count_query`)

		if err := db.GetContext(ctx, count, countQuery.String(), queryParams...); err != nil {
			return nil, nil, WrapError(err)
		}
	}
	if err := qppg.SortQuery(s.SortFields, qp.Sort, &query, &queryParams); err != nil {
		return nil, nil, &store.Error{Type: store.ErrorTypeQuery, Err: err}
	}
	if qp.Limit > 0 {
		query.WriteString(" LIMIT " + strconv.FormatInt(qp.Limit, 10))
	}
	if qp.Offset > 0 {
		query.WriteString(" OFFSET " + strconv.FormatInt(qp.Offset, 10))
	}

	var records = make([]*T, 0)
	err := db.SelectContext(ctx, &records, query.String(), queryParams...)
	if err != nil {
		return records, nil, WrapError(err)
	}
	if s.PostProcessRecord != nil {
		for _, record := range records {
			if err := s.PostProcessRecord(record); err != nil {
				return nil, nil, fmt.Errorf("post proccess record error: %w", err)
			}
		}
	}
	if s.PostProcessRecords != nil {
		if err := s.PostProcessRecords(records); err != nil {
			return nil, nil, fmt.Errorf("post proccess records error: %w", err)
		}
	}

	return records, count, nil

}

func (s *Selector[T]) SelectFirst(ctx context.Context, db DB, qp *queryp.QueryParameters) (*T, error) {

	var query strings.Builder
	var queryParams []any

	query.WriteString(s.Query)
	query.WriteString(" WHERE ")

	if err := qppg.FilterQuery(s.FilterFieldTypes, qp.Filter, &query, &queryParams); err != nil {
		return nil, &store.Error{Type: store.ErrorTypeQuery, Err: err}
	}
	if err := qppg.SortQuery(s.SortFields, qp.Sort, &query, &queryParams); err != nil {
		return nil, &store.Error{Type: store.ErrorTypeQuery, Err: err}
	}

	query.WriteString(" LIMIT 1")

	var record = new(T)
	err := db.GetContext(ctx, record, query.String(), queryParams...)
	if err != nil {
		return record, WrapError(err)
	}
	if s.PostProcessRecord != nil {
		if err := s.PostProcessRecord(record); err != nil {
			return nil, fmt.Errorf("post proccess record error: %w", err)
		}
	}

	return record, nil

}
