//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package expression

import (
	"github.com/couchbase/query/value"
)

/*
Represents range predicate Any, that allow testing of a bool
condition over the elements or attributes of a collection or
object. Type Any is a struct that implements collPred.
*/
type Any struct {
	collPredBase
}

/*
This method returns a pointer to the Any struct that has the
bindings and satisfies fields populated by the input args
bindings and expression satisfies.
*/
func NewAny(bindings Bindings, satisfies Expression) Expression {
	rv := &Any{
		collPredBase: collPredBase{
			bindings:  bindings,
			satisfies: satisfies,
		},
	}

	rv.expr = rv
	return rv
}

/*
It calls the VisitAny method by passing in the receiver to
and returns the interface. It is a visitor pattern.
*/
func (this *Any) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitAny(this)
}

/*
It returns a Boolean value.
*/
func (this *Any) Type() value.Type { return value.BOOLEAN }

func (this *Any) Evaluate(item value.Value, context Context) (value.Value, error) {
	missing := false
	null := false

	barr := make([][]interface{}, len(this.bindings))
	for i, b := range this.bindings {
		bv, err := b.Expression().Evaluate(item, context)
		if err != nil {
			return nil, err
		}

		if b.Descend() {
			buffer := make([]interface{}, 0, 256)
			bv = value.NewValue(bv.Descendants(buffer))
		}

		switch bv.Type() {
		case value.ARRAY:
			barr[i] = bv.Actual().([]interface{})
		case value.MISSING:
			missing = true
		default:
			null = true
		}
	}

	if missing {
		return value.MISSING_VALUE, nil
	}

	if null {
		return value.NULL_VALUE, nil
	}

	n := -1
	for _, b := range barr {
		if n < 0 || len(b) < n {
			n = len(b)
		}
	}

	for i := 0; i < n; i++ {
		cv := value.NewScopeValue(make(map[string]interface{}, len(this.bindings)), item)
		for j, b := range this.bindings {
			cv.SetField(b.Variable(), barr[j][i])
		}

		sv, err := this.satisfies.Evaluate(cv, context)
		if err != nil {
			return nil, err
		}

		if sv.Truth() {
			return value.NewValue(true), nil
		}
	}

	return value.NewValue(false), nil
}

func (this *Any) Copy() Expression {
	return NewAny(this.bindings.Copy(), Copy(this.satisfies))
}
