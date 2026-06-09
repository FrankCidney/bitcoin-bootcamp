package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// RPC Helper
type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func rpc(method string, params []any, wallet string, out any) error {
	url := "127.0.0.1"
	
	if wallet != "" { url += "wallet/" + wallet }
	body, _ := json.Marshal(rpcRequest{
		JSONRPC: "1.0",
		ID: "explorer",
		Method: method,
		Params: params,
	})

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.SetBasicAuth("bootcamp", "bootcamp123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var parsed rpcResponse
	json.NewDecoder(resp.Body).Decode(&parsed)
	if parsed.Error != nil {
		return fmt.Errorf("RPC error: %s", parsed.Error.Message)
	}

	return json.Unmarshal(parsed.Result, out)
}
