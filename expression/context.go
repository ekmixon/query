//  Copyright 2014-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included in
//  the file licenses/Couchbase-BSL.txt.  As of the Change Date specified in that
//  file, in accordance with the Business Source License, use of this software will
//  be governed by the Apache License, Version 2.0, included in the file
//  licenses/APL.txt.

package expression

import (
	"regexp"
	"time"

	"github.com/couchbase/query/auth"
	"github.com/couchbase/query/value"
)

/*
It imports the time package that provides the functionality
to measure and display the time. The type Context is an
interface that has a method Now that returns the Time that
returns the instant it time with a nanosecond precision.
*/
type Context interface {
	Now() time.Time
	GetTimeout() time.Duration
	AuthenticatedUsers() []string
	Credentials() *auth.Credentials
	DatastoreVersion() string
	NewQueryContext(queryContext string, readonly bool) interface{}
	Readonly() bool
	SetAdvisor()
	StoreValue(key string, val interface{})
	RetrieveValue(key string) interface{}
	ReleaseValue(key string)
	EvaluateStatement(statement string, namedArgs map[string]value.Value, positionalArgs value.Values, subquery, readonly bool) (value.Value, uint64, error)
	OpenStatement(statement string, namedArgs map[string]value.Value, positionalArgs value.Values, subquery, readonly bool) (interface {
		NextDocument() (value.Value, error)
		Cancel()
	}, error)
	Parse(s string) (interface{}, error)
}

type ExecutionHandle interface {
	NextDocument() (value.Value, error)
}

type CurlContext interface {
	Context
	GetWhitelist() map[string]interface{}
	UrlCredentials(urlS string) *auth.Credentials
	DatastoreURL() string
}

type InlistContext interface {
	Context
	GetInlistHash(in *In) *InlistHash
	EnableInlistHash(in *In)
	RemoveInlistHash(in *In)
}

type LikeContext interface {
	Context
	GetLikeRegex(in *Like, s string) *regexp.Regexp
	CacheLikeRegex(in *Like, s string, re *regexp.Regexp)
}
