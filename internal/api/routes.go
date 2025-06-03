package api

import "net/http"

var Routes = map[string]func(w http.ResponseWriter, r *http.Request){
	"GET /":        Hello,
	"GET /file":    GetFile,
	"POST /file":   LoadFile,
	"DELETE /file": DeleteFile,
}
