//  Copyright 2015-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included in
//  the file licenses/Couchbase-BSL.txt.  As of the Change Date specified in that
//  file, in accordance with the Business Source License, use of this software will
//  be governed by the Apache License, Version 2.0, included in the file
//  licenses/APL.txt.

package command

import (
	"io"
	"strconv"
	"strings"

	"github.com/couchbase/godbc/n1ql"
	"github.com/couchbase/query/errors"
)

/* Unset Command */
type Unset struct {
	ShellCommand
}

func (this *Unset) Name() string {
	return "UNSET"
}

func (this *Unset) CommandCompletion() bool {
	return false
}

func (this *Unset) MinArgs() int {
	return ONE_ARG
}

func (this *Unset) MaxArgs() int {
	return ONE_ARG
}

func (this *Unset) ExecCommand(args []string) (int, string) {
	/* Command to Unset the value of the given parameter.
	 */

	if len(args) > this.MaxArgs() {
		return errors.TOO_MANY_ARGS, ""

	} else if len(args) < this.MinArgs() {
		return errors.TOO_FEW_ARGS, ""

	} else {
		//Check what kind of parameter needs to be Unset.
		// For query parameters
		if strings.HasPrefix(args[0], "-$") {
			// For Named Parameters
			vble := args[0]
			vble = vble[2:]

			err_code, err_str := PopValue_Helper(true, NamedParam, vble)
			if err_code != 0 {
				return err_code, err_str
			}
			name := "$" + vble
			n1ql.UnsetQueryParams(name)

		} else if strings.HasPrefix(args[0], "-") {
			// For query parameters
			vble := args[0]
			vble = vble[1:]

			err_code, err_str := PopValue_Helper(true, QueryParam, vble)
			if err_code != 0 {
				return err_code, err_str
			}
			n1ql.UnsetQueryParams(vble)

		} else if strings.HasPrefix(args[0], "$") {
			// For User defined session variables
			vble := args[0]
			vble = vble[1:]

			err_code, err_str := PopValue_Helper(true, UserDefSV, vble)
			if err_code != 0 {
				return err_code, err_str
			}

		} else {
			// For Predefined session variables
			vble := args[0]

			err_code, err_str := PopValue_Helper(true, PreDefSV, vble)
			if err_code != 0 {
				return err_code, err_str
			}

			if vble == "histfile" {
				err_code, err_str = PushValue_Helper(false, PreDefSV, "histfile", "\".cbq_history\"")
				if err_code != 0 {
					return err_code, err_str

				}
				HISTFILE = ".cbq_history"

			}

			if vble == "batch" {
				err_code, err_str = PushValue_Helper(false, PreDefSV, "batch", "off")
				if err_code != 0 {
					return err_code, err_str

				}
				BATCH = "off"

			}

			if vble == "quiet" {
				err_code, err_str = PushValue_Helper(false, PreDefSV, "quiet", strconv.FormatBool(false))
				if err_code != 0 {
					return err_code, err_str

				}
				QUIET = false
			}

			//Print the path to histfile
			err_code, err_str = printPath(HISTFILE)
			if err_code != 0 {
				return err_code, err_str
			}

		}
	}
	return 0, ""
}

func (this *Unset) PrintHelp(desc bool) (int, string) {
	_, werr := io.WriteString(W, HUNSET)
	if desc {
		err_code, err_str := printDesc(this.Name())
		if err_code != 0 {
			return err_code, err_str
		}
	}
	_, werr = io.WriteString(W, "\n")
	if werr != nil {
		return errors.WRITER_OUTPUT, werr.Error()
	}
	return 0, ""
}
