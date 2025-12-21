package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// DiscordWebhookPayload 是發送到 Discord Webhook 的 JSON 結構
type DiscordWebhookPayload struct {
	Embeds []DiscordEmbed `json:"embeds"`
}

// DiscordEmbed 是 Discord 訊息中的 embed 部分
type DiscordEmbed struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	URL         string         `json:"url"`
	Color       int            `json:"color"`
	Fields      []DiscordField `json:"fields"`
	Timestamp   string         `json:"timestamp"`
}

// DiscordField 是 Discord embed 裡的欄位
type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

// sendDiscordNotification 負責把 S3 上傳事件通知發送到 Discord
func sendDiscordNotification(webhookURL, bucket, key, region, eventTime string) error {
	// S3 的 object key 可能會被 URL encode 過，先解碼回來
	decodedKey, err := url.QueryUnescape(key)
	if err != nil {
		log.Printf("無法對 object key 進行 URL 解碼 '%s': %v", key, err)
		decodedKey = key
	}

	// 建立 S3 物件的 URL
	objectURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, url.PathEscape(decodedKey))

	// 準備要發送到 Discord 的 payload
	payload := DiscordWebhookPayload{
		Embeds: []DiscordEmbed{
			{
				Title:       "📁 S3 物件上傳通知",
				Description: "有一個新的檔案被上傳到 S3 Bucket 了！",
				URL:         objectURL,
				Color:       3447003,
				Fields: []DiscordField{
					{
						Name:   "Bucket 名稱",
						Value:  bucket,
						Inline: true,
					},
					{
						Name:   "Region",
						Value:  region,
						Inline: true,
					},
					{
						Name:   "檔案路徑 (Object Key)",
						Value:  "`" + decodedKey + "`",
						Inline: false,
					},
				},
				Timestamp: eventTime,
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("無法將 Discord payload 編碼為 JSON: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("無法建立 HTTP 請求: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("發送請求到 Discord 失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body := new(bytes.Buffer)
		body.ReadFrom(resp.Body)
		return fmt.Errorf("Discord 回傳非預期的狀態碼 %d: %s", resp.StatusCode, body.String())
	}

	log.Printf("成功發送通知到 Discord: %s/%s", bucket, decodedKey)
	return nil
}

// handleRequest 是 AWS Lambda 的進入點
// 這裡指定傳入 S3Event，當 S3 有指定動作時，它會把事件內容傳進來
func handleRequest(ctx context.Context, s3Event events.S3Event) error {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("錯誤：環境變數 DISCORD_WEBHOOK_URL 未設定")
	}

	// 逐筆處理 S3 事件記錄
	for _, record := range s3Event.Records {
		s3 := record.S3
		bucket := s3.Bucket.Name
		key := s3.Object.Key
		region := record.AWSRegion
		eventTime := record.EventTime.Format(time.RFC3339)

		log.Printf("偵測到事件: 在 %s 上傳了 %s", bucket, key)

		if err := sendDiscordNotification(webhookURL, bucket, key, region, eventTime); err != nil {
			log.Printf("發送 Discord 通知失敗: %v", err)
		}
	}
	return nil
}

// main 函式啟動 Lambda 處理器
func main() {
	lambda.Start(handleRequest)
}
