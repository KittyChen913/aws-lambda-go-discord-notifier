# AWS Lambda Go Discord Notifier

這是一個用 Go 編寫的 AWS Lambda function，當有新物件上傳到 S3 Bucket 時，自動發送 Discord 通知，支援自訂 Webhook 和完整的物件資訊。

## 功能特色

- 有新物件上傳到 S3 Bucket 時會自動觸發
- 透過 Discord Webhook 發送通知
- 訊息包含 Bucket、檔案路徑、Region、物件 URL
- 環境變數方式管理 Webhook URL

## 環境需求

- **Go**
- **AWS CLI:** 已設定 AWS 帳號存取憑證（`aws configure`）
- **Discord Webhook URL:** 從 Discord 伺服器取得（整合 > Webhook > 新增 Webhook）

## 快速開始

1. 下載相依套件
```bash
go mod tidy
```

2. 編譯 + 打包部署檔案（Windows PowerShell）
```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bootstrap; Compress-Archive -Path .\bootstrap -DestinationPath function.zip -Force
```

3. 上傳至 AWS Lambda 並設定環境變數 DISCORD_WEBHOOK_URL

---

## 📁 專案結構

```
aws-lambda-go-discord-notifier/
├── main.go                 # Lambda 進入點，處理 S3 事件
├── discord.go              # Discord 客戶端，構建和發送通知
├── go.mod                  # Go module 設定檔
├── go.sum                  # Go module 依賴雜湊
├── bootstrap               # Go 的可執行檔
├── function.zip            # Lambda 部署 zip 檔案，直接上傳至 Lambda 用的
├── README.md               # 專案說明文件
└── .gitignore              # Git 忽略檔案清單
```

---

## 部署到 AWS

### 1. 建立 Lambda Function

1. 前往 AWS Lambda 主控台 → 函式 → 建立函式
2. 選擇 **從頭開始撰寫**
3. **函式名稱**：例如 `s3-discord-notifier`
4. **Runtime**：選擇 `Amazon Linux 2023`
5. **架構**：選擇 `x86_64`
6. 點擊 **建立函式**

### 2. 上傳程式碼與設定環境變數

1. **上傳程式碼：**
   - 點擊 **上傳來源** → **.zip 檔案** → **上傳**
   - 選擇之前打包的 `function.zip`

2. **設定環境變數：**
   - 前往 **Configuration** → **Environment variables** → **Edit**
   - 新增環境變數：
     - **金鑰**：`DISCORD_WEBHOOK_URL`
     - **值**：貼上你的 Discord Webhook URL
   - 儲存

### 3. 新增 S3 觸發器

1. 點擊 **Add trigger**
2. **Trigger configuration**：選擇 **S3**
3. **Bucket**：選擇要監控的 S3 Bucket
4. **Event types**：選擇 **All object create events**
5. **（可選）Prefix / Suffix：**
   - Prefix：監控特定資料夾（例如 `uploads/`）
   - Suffix：監控特定副檔名（例如 `.jpg`）
6. 勾選確認，點擊 **Add**

---

## ✅ 測試

上傳檔案到設定的 S3 Bucket，幾秒鐘後會在 Discord 頻道看到通知

### 查看執行日誌

若無收到通知，可查看日誌確認問題：

1. 前往 Lambda function 頁面
2. **Monitor** → **Logs** → **View in CloudWatch**
3. 查看最新的 Log Stream
