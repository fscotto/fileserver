package api

import "net/http"

var Routes = map[string]func(w http.ResponseWriter, r *http.Request){
	"GET /":                 Hello,
	"GET /files":            GetFiles,
	"GET /file/{idFile}":    GetFile,
	"POST /file":            LoadFile,
	"DELETE /file/{idFile}": DeleteFile,
}
