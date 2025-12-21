package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// handleRequest 是 AWS Lambda 的進入點
// 這裡指定傳入 S3Event，當 S3 有指定動作時，它會把事件內容傳進來
func handleRequest(ctx context.Context, s3Event events.S3Event) error {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("錯誤：環境變數 DISCORD_WEBHOOK_URL 未設定")
	}

	// 初始化 Discord 客戶端
	discordClient := NewDiscordClient(webhookURL)

	// 逐筆處理 S3 事件記錄
	for _, record := range s3Event.Records {
		s3 := record.S3

		// 構建事件資訊
		event := S3NotificationEvent{
			Bucket:    s3.Bucket.Name,
			Key:       s3.Object.Key,
			Region:    record.AWSRegion,
			EventTime: record.EventTime.Format(time.RFC3339),
		}

		log.Printf("偵測到事件: 在 %s 上傳了 %s", event.Bucket, event.Key)

		// 發送通知到 Discord
		if err := discordClient.SendS3Notification(event); err != nil {
			log.Printf("發送 Discord 通知失敗: %v", err)
		}
	}
	return nil
}

// main 函式啟動 Lambda 處理器
func main() {
	lambda.Start(handleRequest)
}
