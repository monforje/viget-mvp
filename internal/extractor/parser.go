package extractor

import "encoding/json"

func ParseJSONResponse(response string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(response), &result)
	return result, err
}
