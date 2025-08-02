package executor

import (
	"bytes"
	"encoding/json"
	"event_service/internal/cfg"
	"event_service/internal/dto"
	"event_service/internal/executor/executor_interface"
	"io"
	"log"
	"net/http"
)

type Executor struct {
}

func NewExecutor() executor_interface.Executor {
	e := Executor{}
	return &e
}

func (e *Executor) Send(dto dto.PipelineDTO) error {
	cfg := cfg.GetConfig()
	c := http.Client{}
	reqAddr := cfg.Scheduler.Host
	reqBody, err := json.Marshal(dto.PipelineTemplate.Query)
	if err != nil {
		return err
	}

	token := "Bearer " + cfg.Token

	req, err := http.NewRequest("POST", reqAddr, bytes.NewReader(reqBody))

	if err != nil {
		return err
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)

	if err != nil {
		return err
	}

	respByte, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}
	log.Println(string(respByte))
	return nil
}
