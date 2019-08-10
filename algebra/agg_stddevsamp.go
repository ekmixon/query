//  Copyright (c) 2018 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package algebra

import (
	"math"

	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/value"
)

/*
This represents the Aggregate function stddev_samp(expr). It returns
the arithmetic sample standard deviation of all the number values in the
group. Type stddev_samp is a struct that inherits from AggregateBase.
*/
type StddevSamp struct {
	AggregateBase
}

/*
The function NewStddevSamp calls NewAggregateBase to
create an aggregate function named StddevSamp with
one expression as input.
*/
func NewStddevSamp(operands expression.Expressions, flags uint32, wTerm *WindowTerm) Aggregate {
	rv := &StddevSamp{
		*NewAggregateBase("stddev_samp", operands, flags, wTerm),
	}

	rv.SetExpr(rv)
	return rv
}

/*
It calls the VisitFunction method by passing in the receiver to
and returns the interface. It is a visitor pattern.
*/
func (this *StddevSamp) Accept(visitor expression.Visitor) (interface{}, error) {
	return visitor.VisitFunction(this)
}

/*
It returns a value of type NUMBER.
*/
func (this *StddevSamp) Type() value.Type {
	return value.NUMBER
}

/*
Calls the evaluate method for aggregate functions and passes in the
receiver, current item and current context.
*/
func (this *StddevSamp) Evaluate(item value.Value, context expression.Context) (value.Value, error) {
	return this.evaluate(this, item, context)
}

/*
The constructor returns a NewStddevSamp with the input operand
cast to a Function as the FunctionConstructor.
*/
func (this *StddevSamp) Constructor() expression.FunctionConstructor {
	return func(operands ...expression.Expression) expression.Function {
		return NewStddevSamp(operands, uint32(0), nil)
	}
}

/*
Copy of the aggregate function
*/

func (this *StddevSamp) Copy() expression.Expression {
	rv := &StddevSamp{
		*NewAggregateBase(this.Name(), expression.CopyExpressions(this.Operands()),
			this.Flags(), CopyWindowTerm(this.WindowTerm())),
	}

	rv.BaseCopy(this)
	rv.SetExpr(rv)
	return rv
}

/*
If no input to the StddevSamp function, then the default value
returned is a null.
*/
func (this *StddevSamp) Default(item value.Value, context Context) (value.Value, error) {
	return value.NULL_VALUE, nil
}

/*
Aggregates input data by evaluating operands.
For all values other than Number, return the input value itself.
Maintain two variables for sum and
list of all the values of type NUMBER.
Call addStddevVariance to compute the intermediate aggregate value and return it.
*/
func (this *StddevSamp) CumulateInitial(item, cumulative value.Value, context Context) (value.Value, error) {
	item, e := this.Operands()[0].Evaluate(item, context)
	if e != nil {
		return nil, e
	}

	if item.Type() != value.NUMBER {
		return cumulative, nil
	}

	return addStddevVariance(item, cumulative, this.Distinct())
}

/*
Aggregates intermediate results and return them.
*/
func (this *StddevSamp) CumulateIntermediate(part, cumulative value.Value, context Context) (value.Value, error) {
	return cumulateStddevVariance(part, cumulative, this.Distinct())
}

/*
Compute the sample standard deviation as the final.
Return NULL if no values of type NUMBER exist.
Return NULL if only one value exists.
Calculate variance according to definition
and return the square root of it as the standard deviation.
*/
func (this *StddevSamp) ComputeFinal(cumulative value.Value, context Context) (value.Value, error) {
	if cumulative == value.NULL_VALUE {
		return cumulative, nil
	}

	variance, e := computeVariance(cumulative, this.Distinct(), true, 1.0)
	if e != nil {
		return nil, e
	}

	if variance == value.NULL_VALUE {
		return value.NULL_VALUE, nil
	}
	return value.NewValue(math.Sqrt(variance.(value.NumberValue).Float64())), nil
}
