package api

import "net/http"

func Hello(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("Welcome to my homepage")); err != nil {
		return
	}
	w.WriteHeader(http.StatusOK)
}
