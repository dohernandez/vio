// Package must panics on errors.
package must

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bool64/ctxd"
)

// NotFail panics on error.
func NotFail(err error) {
	if err != nil {
		msg := err.Error()

		var sErr ctxd.StructuredError
		if errors.As(err, &sErr) {
			j, err := json.MarshalIndent(sErr.Fields(), "", " ")
			if err == nil {
				msg += "\nerror context: " + string(j)
			} else {
				msg += fmt.Sprintf("\nfailed to marshal error context %+v: %s", sErr.Fields(), err.Error())
			}
		}

		panic(msg)
	}
}
