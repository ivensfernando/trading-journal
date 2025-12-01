package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// COLOCA TUAS CREDENCIAIS AQUI (NÃO COMMITA ISSO NO GIT)
const (
	// a que você digitou ao criar a API, NÃO o trading password
	apiBaseURL = "https://api.kucoin.com"
)

// KC-API-PASSPHRASE = base64( HMAC_SHA256(apiSecret, apiPassphrase) )
func signPassphrase(secret, passphrase string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(passphrase))
	hash := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(hash)
}

// KC-API-SIGN = base64( HMAC_SHA256(apiSecret, timestamp+method+endpoint+body) )
func signRequest(secret, timestamp, method, endpoint, body string) string {
	prehash := timestamp + method + endpoint + body
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(prehash))
	hash := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(hash)
}

func main() {
	apiKey := os.Getenv("KUCOIN_API_KEY")
	apiSecret := os.Getenv("KUCOIN_API_SECRET")
	apiPassphrase := os.Getenv("KUCOIN_API_PASSPHRASE")
	keyVersion := os.Getenv("KUCOIN_API_KEY_VERSION")

	log.Println(apiKey)
	log.Println(apiSecret)
	log.Println(apiPassphrase)

	if apiKey == "" || apiSecret == "" || apiPassphrase == "" {
		log.Fatal("environment variables KUCOIN_API_KEY, KUCOIN_API_SECRET and KUCOIN_API_PASSPHRASE are required")
	}

	// Exemplo: GET /api/v1/accounts (sem body)
	method := http.MethodGet
	endpoint := "/api/v1/accounts"
	bodyStr := "" // GET sem body

	// timestamp em ms (string)
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))

	// assinatura da request
	signature := signRequest(apiSecret, timestamp, method, endpoint, bodyStr)

	// passphrase criptografada (V3)
	encryptedPassphrase := signPassphrase(apiSecret, apiPassphrase)

	// monta a request HTTP
	req, err := http.NewRequest(method, apiBaseURL+endpoint, bytes.NewBuffer([]byte(bodyStr)))
	if err != nil {
		panic(err)
	}

	// headers exigidos pela KuCoin
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("KC-API-KEY", apiKey)
	req.Header.Set("KC-API-SIGN", signature)
	req.Header.Set("KC-API-TIMESTAMP", timestamp)
	req.Header.Set("KC-API-PASSPHRASE", encryptedPassphrase)
	req.Header.Set("KC-API-KEY-VERSION", keyVersion) // sua chave é V3

	// executa a requisição
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	fmt.Println("Status:", resp.Status)
	fmt.Println("Body:", string(respBody))

}
