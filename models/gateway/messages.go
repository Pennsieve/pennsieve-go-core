package gateway

import (
	"encoding/json"
	"log"
)

type ErrorMessage struct {
	Code    int
	Message string
}

// CreateErrorMessage parses error message into body for GatewayResponse
func CreateErrorMessage(message string, code int) string {
	m := ErrorMessage{
		Code:    code,
		Message: message,
	}

	jsonBody, err := json.Marshal(m)
	if err != nil {
		log.Fatalln(err)
	}
	return string(jsonBody)

}
