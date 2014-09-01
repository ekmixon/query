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
	"strconv"

	"github.com/couchbaselabs/query/expression"
	"github.com/couchbaselabs/query/value"
)

type Projector interface {
	Node
	Signature() value.Value
}

type Projection struct {
	distinct bool        `json:"distinct"`
	raw      bool        `json:"raw"`
	terms    ResultTerms `json:terms`
}

func NewProjection(distinct bool, terms ResultTerms) *Projection {
	rv := &Projection{
		distinct: distinct,
		raw:      false,
		terms:    terms,
	}

	rv.setAliases()
	return rv
}

func NewRawProjection(distinct bool, expr expression.Expression) *Projection {
	rv := &Projection{
		distinct: distinct,
		raw:      true,
		terms:    ResultTerms{NewResultTerm(expr, false, "")},
	}

	rv.setAliases()
	return rv
}

func (this *Projection) Signature() value.Value {
	if this.raw {
		return value.NewValue(this.terms[0].Expression().Type().String())
	}

	rv := value.NewValue(make(map[string]interface{}, len(this.terms)))
	for _, term := range this.terms {
		alias := term.Alias()
		if alias == "" {
			rv.SetField("*", "*")
		} else {
			rv.SetField(alias, term.Expression().Type().String())
		}
	}

	return rv
}

func (this *Projection) MapExpressions(mapper expression.Mapper) (err error) {
	for _, term := range this.terms {
		err = term.MapExpression(mapper)
		if err != nil {
			return
		}
	}

	return
}

func (this *Projection) Formalize(f *Formalizer) (err error) {
	for _, term := range this.terms {
		if term.expr != nil {
			term.expr, err = f.Map(term.expr)
			if err != nil {
				return
			}
		}
	}

	return
}

func (this *Projection) Distinct() bool {
	return this.distinct
}

func (this *Projection) Raw() bool {
	return this.raw
}

func (this *Projection) Terms() ResultTerms {
	return this.terms
}

func (this *Projection) setAliases() {
	a := 1
	for _, term := range this.terms {
		a = term.setAlias(a)
	}
}

type ResultTerms []*ResultTerm

type ResultTerm struct {
	expr  expression.Expression `json:"expr"`
	star  bool                  `json:"star"`
	as    string                `json:"as"`
	alias string                `json:"_"`
}

func NewResultTerm(expr expression.Expression, star bool, as string) *ResultTerm {
	return &ResultTerm{
		expr: expr,
		star: star,
		as:   as,
	}
}

func (this *ResultTerm) MapExpression(mapper expression.Mapper) (err error) {
	if this.expr != nil {
		this.expr, err = mapper.Map(this.expr)
	}

	return
}

func (this *ResultTerm) Expression() expression.Expression {
	return this.expr
}

func (this *ResultTerm) Star() bool {
	return this.star
}

func (this *ResultTerm) As() string {
	return this.as
}

func (this *ResultTerm) Alias() string {
	return this.alias
}

func (this *ResultTerm) setAlias(a int) int {
	if this.star {
		return a
	}

	if this.as != "" {
		this.alias = this.as
	} else {
		this.alias = this.expr.Alias()
	}

	if this.alias == "" {
		this.alias = "$" + strconv.Itoa(a)
		a++
	}

	return a
}
