//  Copyright (c) 2017 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package plan

import (
	"encoding/json"

	"github.com/couchbase/query/algebra"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/expression/parser"
)

type NLNest struct {
	readonly
	outer       bool
	alias       string
	onclause    expression.Expression
	hintError   string
	cost        float64
	cardinality float64
	child       Operator
}

func NewNLNest(nest *algebra.AnsiNest, child Operator, cost, cardinality float64) *NLNest {
	rv := &NLNest{
		outer:       nest.Outer(),
		alias:       nest.Alias(),
		onclause:    nest.Onclause(),
		hintError:   nest.HintError(),
		child:       child,
		cost:        cost,
		cardinality: cardinality,
	}

	return rv
}

func (this *NLNest) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitNLNest(this)
}

func (this *NLNest) New() Operator {
	return &NLNest{}
}

func (this *NLNest) Outer() bool {
	return this.outer
}

func (this *NLNest) Alias() string {
	return this.alias
}

func (this *NLNest) Onclause() expression.Expression {
	return this.onclause
}

func (this *NLNest) HintError() string {
	return this.hintError
}

func (this *NLNest) Child() Operator {
	return this.child
}

func (this *NLNest) Cost() float64 {
	return this.cost
}

func (this *NLNest) Cardinality() float64 {
	return this.cardinality
}

func (this *NLNest) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.MarshalBase(nil))
}

func (this *NLNest) MarshalBase(f func(map[string]interface{})) map[string]interface{} {
	r := map[string]interface{}{"#operator": "NestedLoopNest"}
	r["alias"] = this.alias
	r["on_clause"] = expression.NewStringer().Visit(this.onclause)

	if this.outer {
		r["outer"] = this.outer
	}

	if this.hintError != "" {
		r["hint_not_followed"] = this.hintError
	}

	if this.cost > 0.0 {
		r["cost"] = this.cost
	}

	if this.cardinality > 0.0 {
		r["cardinality"] = this.cardinality
	}

	if f != nil {
		f(r)
	} else {
		r["~child"] = this.child
	}
	return r
}

func (this *NLNest) UnmarshalJSON(body []byte) error {
	var _unmarshalled struct {
		_           string          `json:"#operator"`
		Onclause    string          `json:"on_clause"`
		Outer       bool            `json:"outer"`
		Alias       string          `json:"alias"`
		HintError   string          `json:"hint_not_followed"`
		Cost        float64         `json:"cost"`
		Cardinality float64         `json:"cardinality"`
		Child       json.RawMessage `json:"~child"`
	}

	err := json.Unmarshal(body, &_unmarshalled)
	if err != nil {
		return err
	}

	if _unmarshalled.Onclause != "" {
		this.onclause, err = parser.Parse(_unmarshalled.Onclause)
		if err != nil {
			return err
		}
	}

	this.outer = _unmarshalled.Outer
	this.alias = _unmarshalled.Alias
	this.hintError = _unmarshalled.HintError

	this.cost = getCost(_unmarshalled.Cost)
	this.cardinality = getCardinality(_unmarshalled.Cardinality)

	raw_child := _unmarshalled.Child
	var child_type struct {
		Op_name string `json:"#operator"`
	}

	err = json.Unmarshal(raw_child, &child_type)
	if err != nil {
		return err
	}

	this.child, err = MakeOperator(child_type.Op_name, raw_child)
	if err != nil {
		return err
	}

	return nil
}

func (this *NLNest) verify(prepared *Prepared) bool {
	return this.child.verify(prepared)
}
