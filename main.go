package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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

func showWalletBalance(wallet string) error {
    // Wallet may already be loaded -> ignore the error
    _ = rpc("loadwallet", []any{wallet}, "", nil)

    var balance float64
    if err := rpc("getbalance", nil, wallet, &balance); err != nil {
        return err
    }
    fmt.Printf("=== Wallet: %s ===\n", wallet)
    fmt.Printf("Balance: %v BTC\n", balance)
    return nil
}

func listTransactions(wallet string, count int) error {
    _ = rpc("loadwallet", []any{wallet}, "", nil)

    var txs []struct {
        Category      string  `json:"category"`
        Amount        float64 `json:"amount"`
        TxID          string  `json:"txid"`
        Confirmations int     `json:"confirmations"`
    }
    if err := rpc("listtransactions", []any{"*", count}, wallet, &txs); err != nil {
        return err
    }
    for _, tx := range txs {
        dir := "OUT"
        switch tx.Category {
        case "receive", "generate", "immature":
            dir = "IN "
        }
        fmt.Printf("%s %+.8f BTC | %d confs\n", dir, tx.Amount, tx.Confirmations)
        fmt.Printf("     TXID: %s\n", tx.TxID)
    }
    return nil
}

func decodeTransaction(txid string) error {
    var tx struct {
        Vin []struct {
            Coinbase string `json:"coinbase"`
            TxID     string `json:"txid"`
            Vout     int    `json:"vout"`
        } `json:"vin"`
        Vout []struct {
            Value        float64 `json:"value"`
            ScriptPubKey struct {
                Address string `json:"address"`
            } `json:"scriptPubKey"`
        } `json:"vout"`
    }
    rpc("getrawtransaction", []any{txid, true}, "", &tx)
    for _, vin := range tx.Vin {
        if vin.Coinbase != "" {
            fmt.Println("  COINBASE (mining reward)")
        } else {
            fmt.Printf("  From: %s...\n", vin.TxID[:20])
        }
    }
    for _, vout := range tx.Vout {
        fmt.Printf("  %.8f BTC -> %s\n", vout.Value, vout.ScriptPubKey.Address)
    }
    return nil
}

func showBlock(blockhash string) error {
    if blockhash == "" {
        rpc("getbestblockhash", nil, "", &blockhash)
    }
    var block struct {
        Height int      `json:"height"`
        Hash   string   `json:"hash"`
        Time   int64    `json:"time"`
        NTx    int      `json:"nTx"`
        Tx     []string `json:"tx"`
    }
    if err := rpc("getblock", []any{blockhash, 1}, "", &block); err != nil {
        return err
    }
    fmt.Printf("=== Block #%d ===\n", block.Height)
    fmt.Printf("Hash: %s...\n", block.Hash[:32])
    fmt.Printf("Time: %d\n", block.Time)
    fmt.Printf("Transactions: %d\n", block.NTx)
    return nil
}

func main() {	
	// if err := showBlockchainInfo(); err != nil {
	// 	log.Fatal("error showing blockchain info: ", err)
	// }

	// if err := showWalletBalance("alice"); err != nil {
	// 	log.Fatal("error showing wallet balance: ", err)
	// }	

	// if err := listTransactions("alice", 5); err != nil {
	// 	log.Fatal("error listing transactions: ", err)
	// }

	// if err := decodeTransaction("41f410bc6f389ff39557f2d270a27cae1e41ffa64f759f8fd73914f82585cc01"); err != nil {
	// 	log.Fatal("error decoding transaction: ", err)
	// }

	if err := showBlock(""); err != nil {
		log.Fatal("error showing block: ", err)
	}
	// if err != nil {
	// 	fmt.Printf("error showing blockchain info: %s", err)
	// 	os.Exit(1)
	// }
}