package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	wsURL     = "wss://testnet-api.phemex.com/ws" //"wss://ws.phemex.com/ws" // Produção
	apiKey    = "b9eb2ef0-0a6d-43de-9cee-aeb783d14fff"
	apiSecret = "_Tmd_6WPf7oF5yUTSkq1WshVICDwrMJq78kvOQsZ98NmODhlOTA2Zi0yZTA1LTQ5YjMtYTAyOS02MDc2MmZlNWRlNDk"
)

func main() {
	// 1. Conectar ao WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("Erro ao conectar:", err)
	}
	defer conn.Close()

	fmt.Println("Conectado ao WebSocket!")

	// 2. Gerar expiração e assinatura
	expiry := time.Now().Add(1 * time.Minute).Unix() // max 2 min à frente

	// signature = HMAC_SHA256(apiKey + expiry)
	message := fmt.Sprintf("%s%d", apiKey, expiry)

	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(message))
	signature := hex.EncodeToString(mac.Sum(nil))

	// 3. Enviar mensagem de autenticação
	authMsg := fmt.Sprintf(
		`{"id":1, "method":"user.auth", "params":["API","%s","%s",%d]}`,
		apiKey, signature, expiry,
	)

	fmt.Println("Enviando autenticação...")
	err = conn.WriteMessage(websocket.TextMessage, []byte(authMsg))
	if err != nil {
		log.Fatal("Erro ao enviar auth:", err)
	}

	// 4. Ler resposta da autenticação
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Fatal("Erro ao ler resposta auth:", err)
	}
	fmt.Println("Resposta:", string(msg))

	// 5. Enviar um Ping simples
	ping := `{"id":2, "method":"server.ping", "params":[]}`
	fmt.Println("Enviando ping...")
	err = conn.WriteMessage(websocket.TextMessage, []byte(ping))
	if err != nil {
		log.Fatal("Erro ao enviar ping:", err)
	}

	// 6. Ler resposta do ping
	_, pongMsg, err := conn.ReadMessage()
	if err != nil {
		log.Fatal("Erro ao ler pong:", err)
	}

	fmt.Println("Pong recebido:", string(pongMsg))
}
