//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package algebra

import (
	"github.com/couchbaselabs/query/expression"
	"github.com/couchbaselabs/query/value"
)

type Select struct {
	subresult Subresult             `json:"subresult"`
	order     *Order                `json:"order"`
	offset    expression.Expression `json:"offset"`
	limit     expression.Expression `json:"limit"`
}

func NewSelect(subresult Subresult, order *Order, offset, limit expression.Expression) *Select {
	return &Select{
		subresult: subresult,
		order:     order,
		offset:    offset,
		limit:     limit,
	}
}

func (this *Select) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitSelect(this)
}

func (this *Select) Signature() value.Value {
	return this.subresult.Signature()
}

func (this *Select) MapExpressions(mapper expression.Mapper) (err error) {
	err = this.subresult.MapExpressions(mapper)
	if err != nil {
		return
	}

	if this.order != nil {
		err = this.order.MapExpressions(mapper)
	}

	if this.limit != nil {
		this.limit, err = mapper.Map(this.limit)
		if err != nil {
			return
		}
	}

	if this.offset != nil {
		this.offset, err = mapper.Map(this.offset)
	}

	return
}

func (this *Select) Formalize() (err error) {
	formalizer, err := this.subresult.Formalize()
	if err != nil {
		return err
	}

	if this.order != nil {
		err = this.order.MapExpressions(formalizer)
		if err != nil {
			return
		}
	}

	if this.limit != nil {
		_, err = this.limit.Accept(expression.EMPTY_FORMALIZER)
		if err != nil {
			return
		}
	}

	if this.offset != nil {
		_, err = this.offset.Accept(expression.EMPTY_FORMALIZER)
		if err != nil {
			return
		}
	}

	return
}

func (this *Select) Subresult() Subresult {
	return this.subresult
}

func (this *Select) Order() *Order {
	return this.order
}

func (this *Select) Offset() expression.Expression {
	return this.offset
}

func (this *Select) Limit() expression.Expression {
	return this.limit
}

func (this *Select) SetLimit(limit expression.Expression) {
	this.limit = limit
}

type Order struct {
	terms SortTerms
}

func NewOrder(terms SortTerms) *Order {
	return &Order{
		terms: terms,
	}
}

func (this *Order) MapExpressions(mapper expression.Mapper) error {
	return this.terms.MapExpressions(mapper)
}

func (this *Order) Terms() SortTerms {
	return this.terms
}

type SortTerms []*SortTerm

type SortTerm struct {
	expr       expression.Expression `json:"expr"`
	descending bool                  `json:"desc"`
}

func NewSortTerm(expr expression.Expression, descending bool) *SortTerm {
	return &SortTerm{
		expr:       expr,
		descending: descending,
	}
}

func (this *SortTerm) Expression() expression.Expression {
	return this.expr
}

func (this *SortTerm) Descending() bool {
	return this.descending
}

func (this SortTerms) MapExpressions(mapper expression.Mapper) (err error) {
	for _, term := range this {
		term.expr, err = mapper.Map(term.expr)
		if err != nil {
			return
		}
	}

	return
}

type Subresult interface {
	Projector
	IsCorrelated() bool
	MapExpressions(mapper expression.Mapper) error
	Formalize() (formalizer *Formalizer, err error)
}

type Subselect struct {
	from       FromTerm              `json:"from"`
	let        expression.Bindings   `json:"let"`
	where      expression.Expression `json:"where"`
	group      *Group                `json:"group"`
	projection *Projection           `json:"projection"`
}

func NewSubselect(from FromTerm, let expression.Bindings, where expression.Expression,
	group *Group, projection *Projection) *Subselect {
	return &Subselect{from, let, where, group, projection}
}

func (this *Subselect) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitSubselect(this)
}

func (this *Subselect) Signature() value.Value {
	return this.projection.Signature()
}

func (this *Subselect) IsCorrelated() bool {
	return true // FIXME
}

func (this *Subselect) MapExpressions(mapper expression.Mapper) (err error) {
	if this.from != nil {
		err = this.from.MapExpressions(mapper)
		if err != nil {
			return
		}
	}

	if this.let != nil {
		err = this.let.MapExpressions(mapper)
		if err != nil {
			return
		}
	}

	if this.where != nil {
		this.where, err = mapper.Map(this.where)
		if err != nil {
			return
		}
	}

	if this.group != nil {
		err = this.group.MapExpressions(mapper)
		if err != nil {
			return
		}
	}

	return this.projection.MapExpressions(mapper)
}

func (this *Subselect) Formalize() (f *Formalizer, err error) {
	if this.from == nil {
		f = NewFormalizer()
		return
	}

	f, err = this.from.Formalize()
	if err != nil {
		return
	}

	if this.let != nil {
		err = f.PushBindings(this.let)
		if err != nil {
			return nil, err
		}
	}

	if this.where != nil {
		expr, err := f.Map(this.where)
		if err != nil {
			return nil, err
		}

		this.where = expr
	}

	if this.group != nil {
		f, err = this.group.Formalize(f)
		if err != nil {
			return nil, err
		}
	}

	err = this.projection.MapExpressions(f)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (this *Subselect) From() FromTerm {
	return this.from
}

func (this *Subselect) Let() expression.Bindings {
	return this.let
}

func (this *Subselect) Where() expression.Expression {
	return this.where
}

func (this *Subselect) Group() *Group {
	return this.group
}

func (this *Subselect) Projection() *Projection {
	return this.projection
}

type Group struct {
	by      expression.Expressions `json:by`
	letting expression.Bindings    `json:"letting"`
	having  expression.Expression  `json:"having"`
}

func NewGroup(by expression.Expressions, letting expression.Bindings, having expression.Expression) *Group {
	return &Group{
		by:      by,
		letting: letting,
		having:  having,
	}
}

func (this *Group) MapExpressions(mapper expression.Mapper) (err error) {
	if this.by != nil {
		err = this.by.MapExpressions(mapper)
		if err != nil {
			return
		}
	}

	if this.letting != nil {
		err = this.letting.MapExpressions(mapper)
		if err != nil {
			return
		}
	}

	if this.having != nil {
		this.having, err = mapper.Map(this.having)
	}

	return
}

func (this *Group) Formalize(f *Formalizer) (*Formalizer, error) {
	var err error

	if this.by != nil {
		for i, b := range this.by {
			this.by[i], err = f.Map(b)
			if err != nil {
				return nil, err
			}
		}
	}

	if this.letting != nil {
		err = f.PushBindings(this.letting)
		if err != nil {
			return nil, err
		}
	}

	if this.having != nil {
		this.having, err = f.Map(this.having)
		if err != nil {
			return nil, err
		}
	}

	return f, nil
}

func (this *Group) By() expression.Expressions {
	return this.by
}

func (this *Group) Letting() expression.Bindings {
	return this.letting
}

func (this *Group) Having() expression.Expression {
	return this.having
}

type binarySubresult struct {
	first  Subresult `json:"first"`
	second Subresult `json:"second"`
}

func (this *binarySubresult) Signature() value.Value {
	return this.first.Signature()
}

func (this *binarySubresult) IsCorrelated() bool {
	return this.first.IsCorrelated() || this.second.IsCorrelated()
}

func (this *binarySubresult) MapExpressions(mapper expression.Mapper) (err error) {
	err = this.first.MapExpressions(mapper)
	if err != nil {
		return
	}

	return this.second.MapExpressions(mapper)
}

func (this *binarySubresult) Formalize() (f *Formalizer, err error) {
	var ff, sf *Formalizer
	ff, err = this.first.Formalize()
	if err != nil {
		return nil, err
	}

	sf, err = this.second.Formalize()
	if err != nil {
		return nil, err
	}

	// Intersection
	fa := ff.Allowed.Fields()
	sa := sf.Allowed.Fields()
	for field, _ := range fa {
		_, ok := sa[field]
		if !ok {
			delete(fa, field)
		}
	}

	ff.Allowed = value.NewValue(fa)
	if ff.Keyspace != sf.Keyspace {
		ff.Keyspace = ""
	}

	return ff, nil
}

func (this *binarySubresult) First() Subresult {
	return this.first
}

func (this *binarySubresult) Second() Subresult {
	return this.second
}

type unionSubresult struct {
	binarySubresult
}

func (this *unionSubresult) Signature() value.Value {
	first := this.first.Signature()
	second := this.second.Signature()

	if first.Equals(second) {
		return first
	}

	if first.Type() != value.OBJECT ||
		second.Type() != value.OBJECT {
		return _JSON_SIGNATURE
	}

	rv := first.Copy()
	sa := second.Actual().(map[string]interface{})
	for k, v := range sa {
		cv, ok := rv.Field(k)
		if ok {
			if !value.NewValue(cv).Equals(value.NewValue(v)) {
				rv.SetField(k, _JSON_SIGNATURE)
			}
		} else {
			rv.SetField(k, v)
		}
	}

	return rv
}

var _JSON_SIGNATURE = value.NewValue(value.JSON.String())

type Union struct {
	unionSubresult
}

func NewUnion(first, second Subresult) Subresult {
	return &Union{
		unionSubresult{
			binarySubresult{
				first:  first,
				second: second,
			},
		},
	}
}

func (this *Union) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitUnion(this)
}

type UnionAll struct {
	unionSubresult
}

func NewUnionAll(first, second Subresult) Subresult {
	return &UnionAll{
		unionSubresult{
			binarySubresult{
				first:  first,
				second: second,
			},
		},
	}
}

func (this *UnionAll) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitUnionAll(this)
}

type Intersect struct {
	binarySubresult
}

func NewIntersect(first, second Subresult) Subresult {
	return &Intersect{
		binarySubresult{
			first:  first,
			second: second,
		},
	}
}

func (this *Intersect) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitIntersect(this)
}

type IntersectAll struct {
	binarySubresult
}

func NewIntersectAll(first, second Subresult) Subresult {
	return &IntersectAll{
		binarySubresult{
			first:  first,
			second: second,
		},
	}
}

func (this *IntersectAll) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitIntersectAll(this)
}

type Except struct {
	binarySubresult
}

func NewExcept(first, second Subresult) Subresult {
	return &Except{
		binarySubresult{
			first:  first,
			second: second,
		},
	}
}

func (this *Except) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitExcept(this)
}

type ExceptAll struct {
	binarySubresult
}

func NewExceptAll(first, second Subresult) Subresult {
	return &ExceptAll{
		binarySubresult{
			first:  first,
			second: second,
		},
	}
}

func (this *ExceptAll) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitExceptAll(this)
}
