package service

import (
	"encoding/json"
	"net/http"
)

func (*RouteHandler) response(w http.ResponseWriter, status int, message string, data interface{}) {
	w.WriteHeader(status)
	var jsonData []byte
	if data != nil {
		raw, err := json.Marshal(data)
		if err != nil {
			return
		}
		jsonData = raw
	}
	json.NewEncoder(w).Encode(Response{
		Code:    status,
		Message: message,
		Data:    string(jsonData),
	})
}

func (*RouteHandler) setDefaultHeaders(handler ResponseFunction) ResponseFunction {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS,PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(Response{
				Code:    200,
				Message: "",
				Data:    "",
			})
			return
		}

		handler(w, r)
	}
}
