package helpers

import (
	"encoding/json"
	"errors"
	"net/http"
)

// TODO understand what is useful from this class and if I can use it in some better manner

type Envelope map[string]interface{}

func ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxByte := 1048576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxByte))
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != nil {
		return errors.New("body must have only a single JSON object")
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}
	w.Header().Set("Content-Type", "application/json.go")
	w.WriteHeader(status)
	_, err = w.Write(out)

	if err != nil {
		return err
	}

	return nil
}

func ErrorJSON(w http.ResponseWriter, err error, status ...int) {
	statusCode := http.StatusBadRequest
	if len(status) > 0 {
		statusCode = status[0]
	}
	var payload JsonResponse
	payload.Error = true
	payload.Message = err.Error()
	WriteJSON(w, statusCode, payload)
}

type JsonResponse struct {
	Error   bool        `json.go:"error"`
	Message string      `json.go:"message"`
	Data    interface{} `json.go:"data,omitresponse"`
}
