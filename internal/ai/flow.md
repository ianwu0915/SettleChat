# AI Command Flow

## 1. WebSocket 層面 (internal/chat/client.go)

- 接收用戶消息
- 檢查是否是AI命令（以 "/" 開頭）
- 如果是AI命令：
  - 創建 `ChatMessage` 實例
  - 通過 EventBus 發布 `AICommandEvent` 到 `ai.command.{roomID}` 主題

## 2. EventBus 層面 (internal/nats_messaging/eventbus.go)

- 接收 `AICommandEvent`
- 將事件轉換為 NATS 消息格式
- 通過 NATS 發布到 `settlechat{env}.ai.command.{roomID}` 主題

## 3. NATS 層面 (internal/nats_messaging/subscriber.go)

- 訂閱 `settlechat{env}.ai.command.{roomID}` 主題
- 收到消息後：
  - 解析主題格式
  - 查找對應的處理器
  - 將消息轉發給 `AICommandHandler`

## 4. AI Command Handler 層面 (internal/event_handlers/ai.go)

- 接收 AI 命令事件
- 創建上下文
- 調用 `AI Manager` 的 `HandleAIMessage` 方法
- 處理結果通過 EventBus 發布回應

## 5. AI Manager 層面 (internal/ai/manager.go)

- 接收 `ChatMessage`
- 解析命令和參數
- 創建 `AIMessage` 實例
- 獲取或創建對應的 `Agent`
- 調用 `Agent` 的 `HandleAIMessage` 方法

## 6. Agent 層面 (internal/ai/agent.go)

- 根據命令類型選擇對應的處理器
- 處理命令（例如：/summary, /prompt, /help 等）
- 生成回應
- 返回處理結果

## 7. 回應流程

### 7.1 從 Agent 返回

- Agent 返回處理結果
- Manager 接收結果並包裝

### 7.2 EventBus 發布回應

- 創建 `ChatMessage` 實例
- 通過 EventBus 發布到 `settlechat{env}.message.chat.{roomID}` 主題

### 7.3 NATS 廣播

- NATS 將消息廣播到所有訂閱者
- 包括 WebSocket 客戶端

### 7.4 WebSocket 發送

- WebSocket 客戶端接收消息
- 將消息發送給用戶

## 關鍵組件

- **EventBus**: 事件處理中心，負責消息路由和分發
- **NATS**: 底層消息傳輸層，提供可靠的消息傳遞
- **Manager**: AI 服務的門面，管理 Agent 生命週期
- **Agent**: 具體命令的處理者，包含業務邏輯
- **Provider**: AI 服務提供者，處理具體的 AI 請求

## 錯誤處理

- 每個層級都有錯誤處理機制
- 錯誤會被記錄並向上傳遞
- 最終用戶會收到適當的錯誤提示

## 監控和日誌

- 每個關鍵步驟都有日誌記錄
- 可以追蹤消息的完整生命週期
- 便於問題診斷和性能優化
