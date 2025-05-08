package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ianwu0915/SettleChat/internal/storage"
	"github.com/joho/godotenv"
)

var store *storage.PostgresStore

func TestMain(m *testing.M) {
	_ = godotenv.Load("../../../.env")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("DATABASE_URL not set")
	}
	fmt.Println(dsn)
	var err error
	store, err = storage.NewPostgresStore(dsn)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestUpsertUser(t *testing.T) {
	user := storage.User{
		ID:         "test_user_1",
		UserName:   "TestUser",
		LastActive: time.Now().UTC(),
		CreatedAt:  time.Now().UTC(),
	}

	err := store.UpsertUser(context.Background(), user)
	if err != nil {
		t.Fatalf("UpsertUser Failed: %v", err)
	}
}

func TestSaveAndQueryMessages(t *testing.T) {
	roomID := "testroom"
	now := time.Now().UTC()

	msg := storage.ChatMessage{
		RoomID: roomID,
		SenderID: "test_user_1",
		Sender: "TestUser",
		Content: "Hello from test!", 
		Timestamp: now,
	}

	err := store.SaveMessage(context.Background(), msg)
	if err != nil {
		t.Fatalf("Save Message failed: %v", err)
	}

	msgs, err := store.GetRecentMessages(context.Background(), roomID, 5)
	if err != nil {
		t.Fatalf("GetRecentMessages failed: %v", err)
	}

	if len(msgs) == 0 {
		t.Fatal("No messages returned")
	}

	found := false
	for _, m := range msgs {
		fmt.Printf("cmp: content(%q == %q) → %v\n", m.Content, msg.Content, m.Content == msg.Content)
		fmt.Printf("cmp: sender_id(%q == %q) → %v\n", m.SenderID, msg.SenderID, m.SenderID == msg.SenderID)
		
		if m.Content == msg.Content && m.SenderID == msg.SenderID {
			found = true
			assertCorrectMessage(t, m.RoomID, msg.RoomID)
			assertCorrectMessage(t, m.Sender, msg.Sender)
			assertCorrectMessage(t, m.Content, msg.Content)
			break
		}
	}
	if !found {
		t.Errorf("Saved message not found in query result. Expected content: %q", msg.Content)
	}

}

func assertCorrectMessage(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}