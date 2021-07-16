//  Copyright 2015-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included in
//  the file licenses/Couchbase-BSL.txt.  As of the Change Date specified in that
//  file, in accordance with the Business Source License, use of this software will
//  be governed by the Apache License, Version 2.0, included in the file
//  licenses/APL.txt.

package expression

import (
	"github.com/couchbase/query/value"
)

/*
Expression that implements array indexing in CREATE INDEX.
*/
type All struct {
	ExpressionBase
	array        Expression
	distinct     bool
	flatten_keys *FlattenKeys
}

func NewAll(array Expression, distinct bool) *All {
	rv := &All{
		array:    array,
		distinct: distinct,
	}

	rv.expr = rv
	rv.flatten_keys = rv.flattenKeys()
	return rv
}

/*
Visitor pattern.
*/
func (this *All) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitAll(this)
}

func (this *All) Type() value.Type {
	return this.array.Type()
}

func (this *All) Evaluate(item value.Value, context Context) (value.Value, error) {
	return this.array.Evaluate(item, context)
}

func (this *All) EvaluateForIndex(item value.Value, context Context) (value.Value, value.Values, error) {
	val, vals, err := this.array.EvaluateForIndex(item, context)
	if err != nil {
		return nil, nil, err
	}

	if vals != nil {
		return val, vals, nil
	}

	var rv value.Values
	switch val.Type() {
	case value.ARRAY:
		act := val.Actual().([]interface{})
		rv = make(value.Values, len(act))
		for i, a := range act {
			rv[i] = value.NewValue(a)
		}
	case value.NULL:
		rv = _NULL_ARRAY
	case value.MISSING:
		rv = _MISSING_ARRAY
	default:
		// Coerce scalar into array
		rv = value.Values{val}
	}

	return val, rv, nil
}

var _NULL_ARRAY = value.Values{value.NULL_VALUE}

var _MISSING_ARRAY = value.Values{value.MISSING_VALUE}

func (this *All) IsArrayIndexKey() (bool, bool, bool) {
	return true, this.distinct, this.flatten_keys != nil
}

func (this *All) Value() value.Value {
	return this.array.Value()
}

func (this *All) Static() Expression {
	return this.array.Static()
}

func (this *All) Alias() string {
	return this.array.Alias()
}

func (this *All) Indexable() bool {
	return this.array.Indexable()
}

func (this *All) PropagatesMissing() bool {
	return this.array.PropagatesMissing()
}

func (this *All) PropagatesNull() bool {
	return this.array.PropagatesNull()
}

func (this *All) EquivalentTo(other Expression) bool {
	all, ok := other.(*All)
	return ok && (this.distinct == all.distinct) &&
		this.array.EquivalentTo(all.array)
}

func (this *All) DependsOn(other Expression) bool {
	// Unwrap other if possible
	for all, ok := other.(*All); ok && (this.distinct || !all.distinct); all, ok = other.(*All) {
		other = all.array
	}

	return this.array.DependsOn(other)
}

func (this *All) CoveredBy(keyspace string, exprs Expressions, options CoveredOptions) Covered {
	return this.array.CoveredBy(keyspace, exprs, options)
}

func (this *All) Children() Expressions {
	return Expressions{this.array}
}

func (this *All) MapChildren(mapper Mapper) error {
	c, err := mapper.Map(this.array)
	if err == nil && c != this.array {
		this.array = c
	}

	return err
}

func (this *All) Copy() Expression {
	rv := NewAll(this.array.Copy(), this.distinct)
	rv.BaseCopy(this)
	rv.flatten_keys = rv.flattenKeys()
	return rv
}

func (this *All) Array() Expression {
	return this.array
}

func (this *All) Distinct() bool {
	return this.distinct
}

func (this *All) Flatten() bool {
	return this.flatten_keys != nil
}

func (this *All) FlattenSize() int {
	if this.flatten_keys != nil {
		return len(this.flatten_keys.Operands())
	}
	return 0
}

func (this *All) FlattenKeys() *FlattenKeys {
	return this.flatten_keys
}

/*
   DISTINCT ARRAY ( DISTINCT ARRAY flatten_keys(v.c1,v.c2) FOR v IN v1.aa END) FOR v1 IN a1 END
*/

func (this *All) flattenKeys() *FlattenKeys {
	all := this
	for {
		switch array := all.Array().(type) {
		case *Array:
			switch valMapping := array.ValueMapping().(type) {
			case *All:
				all = valMapping
			case *FlattenKeys:
				return valMapping
			default:
				return nil
			}
		default:
			return nil
		}
	}
	return nil
}
