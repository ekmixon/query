//  Copyright 2016-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included in
//  the file licenses/Couchbase-BSL.txt.  As of the Change Date specified in that
//  file, in accordance with the Business Source License, use of this software will
//  be governed by the Apache License, Version 2.0, included in the file
//  licenses/APL.txt.

package plannerbase

import (
	"github.com/couchbase/query/expression"
)

func (this *subset) VisitWithin(expr *expression.Within) (interface{}, error) {
	return this.visitDefault(expr)
}