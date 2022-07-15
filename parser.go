package fpbun

import (
	"fmt"

	"github.com/cbrand/go-filterparams"
	"github.com/cbrand/go-filterparams/definition"
	"github.com/uptrace/bun"
)

type Parser interface {
	AppendTo(sq *bun.SelectQuery, data *filterparams.QueryData) *bun.SelectQuery
}

type defaultParser struct {
	nameConverter NameConverter
}

func NewParser(nameConverter NameConverter) Parser {
	return &defaultParser{nameConverter: nameConverter}
}

func (p defaultParser) AppendTo(sq *bun.SelectQuery, data *filterparams.QueryData) *bun.SelectQuery {
	sq = p.appendFilter(sq, data.GetFilter(), false)
	sq = p.appendOrder(sq, data.GetOrders())

	return sq
}

func (p defaultParser) appendFilter(sq *bun.SelectQuery, filter interface{}, or bool) *bun.SelectQuery {
	switch f := filter.(type) {
	case *definition.And:
		sq = p.appendAnd(sq, f, or)
	case *definition.Or:
		sq = p.appendOr(sq, f, or)
	case *definition.Negate:
		sq = p.appendNegate(sq, f, or)
	case *definition.Parameter:
		sq = p.appendParameter(sq, f, or)
	}

	return sq
}

func (p defaultParser) appendAnd(sq *bun.SelectQuery, def *definition.And, or bool) *bun.SelectQuery {
	groupFunc := func(sq *bun.SelectQuery) *bun.SelectQuery {
		sq = p.appendFilter(sq, def.Left, false)
		sq = p.appendFilter(sq, def.Right, false)

		return sq
	}

	return p.appendGroup(sq, groupFunc, or)
}

func (p defaultParser) appendOr(sq *bun.SelectQuery, def *definition.Or, or bool) *bun.SelectQuery {
	groupFunc := func(sq *bun.SelectQuery) *bun.SelectQuery {
		sq = p.appendFilter(sq, def.Left, true)
		sq = p.appendFilter(sq, def.Right, true)

		return sq
	}

	return p.appendGroup(sq, groupFunc, or)
}

func (p defaultParser) appendNegate(sq *bun.SelectQuery, def *definition.Negate, or bool) *bun.SelectQuery {
	groupFunc := func(sq *bun.SelectQuery) *bun.SelectQuery {
		sq = p.appendFilter(sq, def.Negated, false)

		return sq
	}

	return p.appendGroup(sq, groupFunc, or)
}

func (p defaultParser) appendGroup(sq *bun.SelectQuery, groupFunc func(sq *bun.SelectQuery) *bun.SelectQuery, or bool) *bun.SelectQuery {
	if or {
		sq = sq.WhereGroup(" OR ", groupFunc)
	} else {
		sq = sq.WhereGroup(" AND ", groupFunc)
	}

	return sq
}

func (p defaultParser) appendParameter(sq *bun.SelectQuery, def *definition.Parameter, or bool) *bun.SelectQuery {
	op := "="

	switch def.Filter.Identification {
	case definition.FilterLike.Identification:
		op = "LIKE"
	case definition.FilterILike.Identification:
		op = "ILIKE"
	case definition.FilterGt.Identification:
		op = ">"
	case definition.FilterGte.Identification:
		op = ">="
	case definition.FilterLt.Identification:
		op = "<"
	case definition.FilterLte.Identification:
		op = "<="
	}

	cond := fmt.Sprintf("? %s ?", op)
	ident := bun.Ident(p.nameConverter.Convert(def.Name))

	if or {
		sq = sq.WhereOr(cond, ident, def.Value)
	} else {
		sq = sq.Where(cond, ident, def.Value)
	}

	return sq
}

func (p defaultParser) appendOrder(sq *bun.SelectQuery, orders []*definition.Order) *bun.SelectQuery {
	for _, order := range orders {
		ident := bun.Ident(p.nameConverter.Convert(order.GetOrderBy()))

		if order.OrderDesc() {
			sq = sq.OrderExpr("? DESC", ident)
		} else {
			sq = sq.OrderExpr("? ASC", ident)
		}
	}

	return sq
}
