package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	url := "http://127.0.0.1:18443/"
	
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

func showBlockchainInfo() error {
    var info struct {
        Chain      string  `json:"chain"`
        Blocks     int     `json:"blocks"`
        Difficulty float64 `json:"difficulty"`
    }
    if err := rpc("getblockchaininfo", nil, "", &info); err != nil {
        return err
    }
    fmt.Println("=== Blockchain Info ===")
    fmt.Printf("Chain:      %s\n", info.Chain)
    fmt.Printf("Blocks:     %d\n", info.Blocks)
    fmt.Printf("Difficulty: %v\n", info.Difficulty)
    return nil
}

func main() {
	err := showBlockchainInfo()
	if err != nil {
		fmt.Printf("error showing blockchain info: %s", err)
		os.Exit(1)
	}
}