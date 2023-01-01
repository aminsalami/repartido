package httpHandler

import (
	"encoding/json"
	"github.com/aminsalami/repartido/internal/agent/entities"
	"github.com/aminsalami/repartido/internal/agent/ports"
	"go.uber.org/zap"
	"io"
	"net/http"
)

var logger = zap.NewExample().Sugar()

type HttpHandler struct {
	Agent ports.IAgent
	// Address in the form of "host:port", Default "0.0.0.0:6000"
	Addr string
}

func (h *HttpHandler) Run() {
	mux := http.NewServeMux()
	mux.HandleFunc("/data", h.dataHandler)
	logger.Info("Started listening on " + h.Addr)
	err := http.ListenAndServe(h.Addr, mux)
	if err != nil {
		logger.Fatal(err.Error())
	}
}

func (h *HttpHandler) dataHandler(rw http.ResponseWriter, req *http.Request) {
	parsedRequest := entities.ParsedRequest{}
	err := h.parse(req, &parsedRequest)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	// Handle commands based on parsedRequest
	switch parsedRequest.Command {
	case entities.GET:
		res := h.handleGetCommand(parsedRequest)
		rw.Write([]byte(res))
	case entities.SET:
		res := h.handleSetCommand(parsedRequest)
		rw.Write([]byte(res))
	case entities.DEL:
		res := h.handleDelCommand(parsedRequest)
		rw.Write([]byte(res))
	default:
		rw.Write([]byte("FUCK YOU EZEKIEL!"))
	}
}

func (h *HttpHandler) parse(rawRequest *http.Request, parsed *entities.ParsedRequest) error {
	body, err := io.ReadAll(rawRequest.Body)
	if err != nil {
		return err
	}
	bodyStruct := struct {
		Key  string // json: key
		Data string // json: data
	}{}
	err = json.Unmarshal(body, &bodyStruct)
	if err != nil {
		return err
	}
	parsed.Key = bodyStruct.Key
	parsed.Data = bodyStruct.Data
	switch rawRequest.Method {
	case http.MethodGet:
		parsed.Command = entities.GET
	case http.MethodPost:
		parsed.Command = entities.SET
	case http.MethodDelete:
		parsed.Command = entities.DEL
	default:
		parsed.Command = entities.Unknown
	}
	logger.Debugw("parsed requests received", "key", parsed.Key, "data", parsed.Data)
	return nil
}

// -----------------------------------------------------------------

// TODO: Return JSON response

func (h *HttpHandler) handleGetCommand(request entities.ParsedRequest) string {
	result, err := h.Agent.RetrieveData(request)
	if err != nil {
		logger.Error(err.Error())
		return err.Error()
	}
	//return "GET Key was:" + " -- " + request.Key
	return result
}

func (h *HttpHandler) handleSetCommand(request entities.ParsedRequest) string {
	err := h.Agent.StoreData(request)
	if err != nil {
		logger.Error(err.Error())
		return err.Error()
	}
	return "Success. SET Key was:" + " -- " + request.Key
}

func (h *HttpHandler) handleDelCommand(request entities.ParsedRequest) string {
	err := h.Agent.DeleteData(request)
	if err != nil {
		logger.Error(err.Error())
		return err.Error()
	}
	return "Success. Key deleted:" + " -- " + request.Key
}
