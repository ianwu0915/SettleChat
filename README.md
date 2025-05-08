
### Module 
settlechat/
│
├── cmd/
│   └── server/             # 進入點 main.go
│       └── main.go
│
├── internal/               # 封裝核心邏輯，不對外暴露
│   ├── ws/                 # WebSocket handler、連線升級、進入點
│   ├── chat/               # Chatroom、Client、Hub 等核心邏輯
│   ├── ai/                 # AI 整合（DeepSeek、OpenAI 等）
│   ├── storage/            # 訊息持久化邏輯（目前用 SQLite）
│   ├── command/            # 處理 `/summary` 等指令的邏輯
│   └── utils/              # UUID、時間、字串處理
│
├── web/                    # 靜態資源（HTML/JS/CSS），或 frontend build 出來的內容
│   └── index.html
│
├── go.mod
├── go.sum
└── README.md


## What is in mind when scale up

1. Assigning a Client to a Server => Load Balancing + Consistent Hashing
2. Read/Write Optimization with choices of databases
3. Middleware Like Kafka/Flink to handle high message throughput for a chatserver
4. Caching 
5. Distirubuted ID?
6. Persistent Client connection with Server using Zookeeper and Load-balancer
7. 存新傳的message到資料庫之前，先傳到群組 ，透過flink or Kafka Streaming 
8. Change Current Design Using a single websocket connection for each server
