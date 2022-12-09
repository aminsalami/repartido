package agent

import (
	"encoding/json"
	"io"
	"net/http"
)

type ParsedRequest struct {
	Command Command
	Key     string
	Data    string
}

// -----------------------------------------------------------------

// RestParser implements RequestParser interface to convert REST body to a ParsedRequest.
// Rest Examples:
//
//	http GET /data --json {"key": "myUsername"}
//	http POST /data --json {"key": "myUsername", "data": "jasonMoMoa3344"}
//	http DELETE /data --json {"key": "myUsername"}
//	http GET /data/nodes --json {}
type RestParser struct{}

func (parser RestParser) Parse(rawRequest any, parsed *ParsedRequest) error {
	req := rawRequest.(*http.Request)
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	bodyStruct := struct {
		key  string // json: key
		data string // json: data
	}{}
	err = json.Unmarshal(body, &bodyStruct)
	if err != nil {
		return err
	}
	parsed.Key = bodyStruct.key
	parsed.Data = bodyStruct.data
	switch req.Method {
	case http.MethodGet:
		parsed.Command = GET
	case http.MethodPost:
		parsed.Command = SET
	case http.MethodDelete:
		parsed.Command = DEL
	default:
		parsed.Command = UNKNOWN
	}
	return nil
}

func NewRestParser() RequestParser {
	return RestParser{}
}
