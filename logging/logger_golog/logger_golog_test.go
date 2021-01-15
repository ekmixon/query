//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package logger_golog

import (
	"fmt"
	"os"
	"testing"

	"github.com/couchbase/query/logging"
)

func logMessages(logger *goLogger) {
	logger.Debugf("This is a message from %s", "Debugf")
	logger.Tracef("This is a message from %s", "Tracef")
	logger.Requestf(logging.WARN, "This is a message from %s", "Requestf")
	logger.Infof("This is a message from %s", "Infof")
	logger.Warnf("This is a message from %s", "Warnf")
	logger.Errorf("This is a message from %s", "Errorf")
	logger.Severef("This is a message from %s", "Severef")
	logger.Fatalf("This is a message from %s", "Fatalf")

	logging.Debugf("This is a message from %s", "Debugf")
	logging.Tracef("This is a message from %s", "Tracef")
	logging.Requestf(logging.WARN, "This is a message from %s", "Requestf")
	logging.Infof("This is a message from %s", "Infof")
	logging.Warnf("This is a message from %s", "Warnf")
	logging.Errorf("This is a message from %s", "Errorf")
	logging.Severef("This is a message from %s", "Severef")
	logging.Fatalf("This is a message from %s", "Fatalf")

	logger.Debuga(func() string { return "This is a message from Debuga" })
	logger.Tracea(func() string { return "This is a message from Tracea" })
	logger.Requesta(logging.WARN, func() string { return "This is a message from Requesta" })
	logger.Infoa(func() string { return "This is a message from Infoa" })
	logger.Warna(func() string { return "This is a message from Warna" })
	logger.Errora(func() string { return "This is a message from Errora" })
	logger.Severea(func() string { return "This is a message from Severea" })
	logger.Fatala(func() string { return "This is a message from Fatala" })

	logging.Debuga(func() string { return "This is a message from Debuga" })
	logging.Tracea(func() string { return "This is a message from Tracea" })
	logging.Requesta(logging.WARN, func() string { return "This is a message from Requesta" })
	logging.Infoa(func() string { return "This is a message from Infoa" })
	logging.Warna(func() string { return "This is a message from Warna" })
	logging.Errora(func() string { return "This is a message from Errora" })
	logging.Severea(func() string { return "This is a message from Severea" })
	logging.Fatala(func() string { return "This is a message from Fatala" })
}

func TestStub(t *testing.T) {
	logger := NewLogger(os.Stdout, logging.DEBUG, false)
	logging.SetLogger(logger)

	logMessages(logger)

	logger.SetLevel(logging.WARN)
	fmt.Printf("Log level is %s\n", logger.Level())

	logMessages(logger)

	fmt.Printf("Changing to json formatter\n")
	logger.entryFormatter = &jsonFormatter{}
	logger.SetLevel(logging.DEBUG)

	logMessages(logger)
}
