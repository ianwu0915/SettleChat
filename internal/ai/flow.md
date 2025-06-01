## 傳送AI command的flow

1. **WebSocket 層面**：
   - 接收用戶消息
   - 檢查是否是命令（以 "/" 開頭）
   - 如果是命令：
     - 創建 `AIMessage` 實例
     - 設置 `CommandType` 和初始 `Status`
     - 通過 EventBus 發布 `AIMessage` 到 `ai.command` 事件

2. **EventBus 層面**：
   - 接收 `ai.command` 事件
   - 處理事件路由
   - 通過 NATS 發布消息到 `ai.command` 主題

3. **NATS 層面**：
   - 訂閱 `ai.command` 主題
   - 收到消息後通過 EventBus 轉發給 AI Manager

4. **AI Manager 層面**：
   - 從 EventBus 接收 `AIMessage`
   - 根據 `CommandType` 分發給相應的處理邏輯
   - 更新消息 `Status`
   - 調用相應的 Agent 處理命令

5. **Agent 層面**：
   - 處理具體的命令
   - 生成回應
   - 更新處理狀態
   - 通過 EventBus 發布處理結果

6. **回應流程**：
   - EventBus 接收處理結果
   - 通過 NATS 發布回應
   - 最終到達 WebSocket 層面
   - WebSocket 發送回應給用戶

- EventBus 作為主要的事件處理中心
- NATS 作為底層的消息傳輸層
- 所有消息都通過 EventBus 進行路由和分發



