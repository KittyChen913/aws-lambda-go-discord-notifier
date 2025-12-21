package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Discord embed 裡的欄位
type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

// Discord 訊息中的 embed 部分
type DiscordEmbed struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	URL         string         `json:"url"`
	Color       int            `json:"color"`
	Fields      []DiscordField `json:"fields"`
	Timestamp   string         `json:"timestamp"`
}

// 發送到 Discord Webhook 的 JSON 結構
type DiscordPayload struct {
	Embeds []DiscordEmbed `json:"embeds"`
}

// S3 上傳事件的相關資訊
type S3NotificationEvent struct {
	Bucket    string
	Key       string
	Region    string
	EventTime string
}

// DiscordClient 提供 Discord Webhook 的客戶端，負責構建和發送通知訊息
type DiscordClient struct {
	webhookURL string
	httpClient *http.Client
}

// 建立一個新的 Discord 客戶端
func NewDiscordClient(webhookURL string) *DiscordClient {
	return &DiscordClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// 構建 S3 上傳通知的 embed
func (dc *DiscordClient) buildS3UploadEmbed(event S3NotificationEvent) (DiscordEmbed, error) {
	// S3 的 object key 可能會被 URL encode 過，先解碼回來
	decodedKey, err := url.QueryUnescape(event.Key)
	if err != nil {
		log.Printf("無法對 object key 進行 URL 解碼 '%s': %v", event.Key, err)
		decodedKey = event.Key
	}

	// 建立 S3 物件的 URL
	objectURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", event.Bucket, event.Region, url.PathEscape(decodedKey))

	return DiscordEmbed{
		Title:       "📁 S3 物件上傳通知",
		Description: "有一個新的檔案被上傳到 S3 Bucket 了！",
		URL:         objectURL,
		Color:       3447003,
		Fields: []DiscordField{
			{
				Name:   "Bucket 名稱",
				Value:  event.Bucket,
				Inline: true,
			},
			{
				Name:   "Region",
				Value:  event.Region,
				Inline: true,
			},
			{
				Name:   "檔案路徑 (Object Key)",
				Value:  "`" + decodedKey + "`",
				Inline: false,
			},
		},
		Timestamp: event.EventTime,
	}, nil
}

// 發送 S3 上傳通知到 Discord
func (dc *DiscordClient) SendS3Notification(event S3NotificationEvent) error {
	embed, err := dc.buildS3UploadEmbed(event)
	if err != nil {
		return fmt.Errorf("無法構建 Discord embed: %w", err)
	}

	payload := DiscordPayload{
		Embeds: []DiscordEmbed{embed},
	}

	return dc.send(payload, event.Bucket, event.Key)
}

// 發送 payload 到 Discord Webhook
func (dc *DiscordClient) send(payload DiscordPayload, bucket, key string) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("無法將 Discord payload 編碼為 JSON: %w", err)
	}

	req, err := http.NewRequest("POST", dc.webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("無法建立 HTTP 請求: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := dc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("發送請求到 Discord 失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body := new(bytes.Buffer)
		body.ReadFrom(resp.Body)
		return fmt.Errorf("Discord 回傳非預期的狀態碼 %d: %s", resp.StatusCode, body.String())
	}

	log.Printf("成功發送通知到 Discord: %s/%s", bucket, key)
	return nil
}
