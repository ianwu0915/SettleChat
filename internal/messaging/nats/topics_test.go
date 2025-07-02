package nats

import "testing"



func setupNewTopicFormatter() *TopicFormatter {
	return NewTopicFormatter("")
}
func TestGetPresenceTopic(t *testing.T) {
	formatter := setupNewTopicFormatter()

	got := formatter.GetPresenceTopic("123") 
	want := "settlechat.user.presence.123"

	assertCorrect(t, got, want)
}

func TestUserLeftJoinTopic(t *testing.T) {
	formatter := setupNewTopicFormatter()
	t.Run("getting userleft Topic", func(t *testing.T) {
		got := formatter.GetUserLeftTopic("123")
		want := "settlechat.user.left.123"

		assertCorrect(t, got, want)
	})

	t.Run("getting userJoin Topic", func(t *testing.T) {
		got := formatter.GetUserJoinedTopic("123")
		want := "settlechat.user.joined.123"

		assertCorrect(t, got, want)
	})
}


func assertCorrect(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}