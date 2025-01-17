//  Copyright 2021-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included in
//  the file licenses/Couchbase-BSL.txt.  As of the Change Date specified in that
//  file, in accordance with the Business Source License, use of this software will
//  be governed by the Apache License, Version 2.0, included in the file
//  licenses/APL.txt.

// package couchbase provides low level access to the KV store and the orchestrator
package couchbase

import (
	"testing"
	"unsafe"
)

func testBucket() Bucket {
	b := Bucket{vBucketServerMap: unsafe.Pointer(&VBucketServerMap{
		VBucketMap: make([][]int, 256),
	})}
	return b
}

/*
key: k0 master: 10.1.7.1:11210 vBucketId: 9 couchApiBase: http://10.1.7.1:8092/default replicas: 10.1.7.2:11210
key: k1 master: 10.1.7.1:11210 vBucketId: 14 couchApiBase: http://10.1.7.1:8092/default replicas: 10.1.7.3:11210
key: k2 master: 10.1.7.1:11210 vBucketId: 7 couchApiBase: http://10.1.7.1:8092/default replicas: 10.1.7.2:11210
key: k3 master: 10.1.7.1:11210 vBucketId: 0 couchApiBase: http://10.1.7.1:8092/default replicas: 10.1.7.2:11210
key: k4 master: 10.1.7.2:11210 vBucketId: 100 couchApiBase: http://10.1.7.2:8092/default replicas: 10.1.7.5:11210
key: k5 master: 10.1.7.2:11210 vBucketId: 99 couchApiBase: http://10.1.7.2:8092/default replicas: 10.1.7.5:11210
*/

func TestVBHash(t *testing.T) {
	b := testBucket()
	m := map[string]uint32{
		"k0": 9,
		"k1": 14,
		"k2": 7,
		"k3": 0,
		"k4": 100,
		"k5": 99,
	}

	for k, v := range m {
		assert(t, k, b.VBHash(k), v)
	}
}
