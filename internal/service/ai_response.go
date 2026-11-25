package service

import (
	"strings"
)

func decodeAIJSON(response string, target any) error {

	response = strings.TrimSpace(response)

	if strings.HasPrefix(response, "```") {
	}
	return nil
}
