package main

import (
	"fmt"
	"net/http"
)

// Video handlers temporarily disabled to fix docker build
// Will be re-implemented later

// Placeholder to satisfy build
func (up *UserPortal) disabledVideoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintln(w, `{"error": "Video endpoints temporarily disabled"}`)
}
