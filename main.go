package main

import (
    "crypto/sha1"
    "encoding/base64"
    "fmt"
    "io"
    "net/http"
    "strings"
)

// Функція для генерації WebSocket accept key
func generateAcceptKey(key string) string {
    magicString := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
    hash := sha1.New()
    io.WriteString(hash, key+magicString)
    return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

// Функція для обробки WebSocket з'єднання
func handleConnections(w http.ResponseWriter, r *http.Request) {
    if r.Header.Get("Upgrade") != "websocket" {
        http.Error(w, "Not a websocket handshake", http.StatusBadRequest)
        return
    }

    key := r.Header.Get("Sec-WebSocket-Key")
    acceptKey := generateAcceptKey(key)

    headers := http.Header{}
    headers.Add("Upgrade", "websocket")
    headers.Add("Connection", "Upgrade")
    headers.Add("Sec-WebSocket-Accept", acceptKey)

    // Відправляємо відповідь на запит WebSocket
    for k, v := range headers {
        w.Header().Set(k, strings.Join(v, ""))
    }

    w.WriteHeader(http.StatusSwitchingProtocols)

    conn, _, err := w.(http.Hijacker).Hijack()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer conn.Close()

    // Читаємо повідомлення від клієнта
    for {
        message := make([]byte, 512)
        n, err := conn.Read(message)
        if err != nil {
            fmt.Println("Error reading message:", err)
            break
        }
        fmt.Printf("Received: %s\n", message[:n])

        // Відправляємо повідомлення назад клієнту
        _, err = conn.Write(message[:n])
        if err != nil {
            fmt.Println("Error writing message:", err)
            break
        }
    }
}

func main() {
    http.HandleFunc("/ws", handleConnections)
    fmt.Println("Server started on :8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println("Server failed:", err)
    }
}