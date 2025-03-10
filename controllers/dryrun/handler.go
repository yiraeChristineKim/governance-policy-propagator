package dryrun

import (
	"fmt"
	"net/http"
)

func PostHandler(w http.ResponseWriter, r *http.Request) {
	// Retrie dve context from request
	ctx := r.Context()

	// Check for a deadline
	deadline, ok := ctx.Deadline()
	if ok {
		fmt.Fprintf(w, "Context has a deadline: %v\n", deadline)
	} else {
		fmt.Fprintln(w, "No deadline set in context")
	}
}
