//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package planner

import (
	"regexp"

	"github.com/couchbaselabs/query/expression"
)

type sargableLike struct {
	predicate
}

func newSargableLike(expr expression.BinaryFunction, re *regexp.Regexp) *sargableLike {
	rv := &sargableLike{}
	rv.test = func(expr2 expression.Expression) (bool, error) {
		if !expr.First().EquivalentTo(expr2) {
			return false, nil
		}

		if re == nil {
			return false, nil
		}

		prefix, complete := re.LiteralPrefix()
		return complete || prefix != "", nil
	}

	return rv
}
