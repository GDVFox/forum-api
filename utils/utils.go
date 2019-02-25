package utils

import (
	"net/http"

	"github.com/mailru/easyjson"
)

// WriteEasyjson принимает структуру для easyjson, формирует и отправляет json ответ,
// если перед отправкой что-то ломается, отправляет 500
func WriteEasyjson(v easyjson.Marshaler, w http.ResponseWriter) {
	started, _, err := easyjson.MarshalToHTTPResponseWriter(v, w)
	if err != nil {
		return
	}

	if !started {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
