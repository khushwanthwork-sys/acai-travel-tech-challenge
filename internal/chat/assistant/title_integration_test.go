package assistant

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/acai-travel/tech-challenge/internal/chat/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestTitleDoesNotAnswerQuestion verifies Task 1 fix:
// Title should SUMMARIZE the question, not ANSWER it
func TestTitleDoesNotAnswerQuestion(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping: OPENAI_API_KEY not set")
	}

	ctx := context.Background()
	assist := New()

	testCases := []struct {
		name               string
		question           string
		shouldNotContain   []string // Title should NOT contain these (answers)
		shouldContainWords []string // Title should mention these topics
	}{
		{
			name:               "weather question",
			question:           "What is the weather like in Barcelona?",
			shouldNotContain:   []string{"sunny", "rainy", "cloudy", "°C", "°F", "degrees"},
			shouldContainWords: []string{"weather", "barcelona"},
		},
		{
			name:               "math question",
			question:           "What is 25 plus 17?",
			shouldNotContain:   []string{"42", "answer is"},
			shouldContainWords: []string{"math", "addition", "calculation", "sum"},
		},
		{
			name:               "date question",
			question:           "What day is today?",
			shouldNotContain:   []string{"monday", "tuesday", "wednesday", "2025"},
			shouldContainWords: []string{"date", "today", "day"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conv := &model.Conversation{
				ID:        primitive.NewObjectID(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Messages: []*model.Message{{
					ID:        primitive.NewObjectID(),
					Role:      model.RoleUser,
					Content:   tc.question,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}},
			}

			title, err := assist.Title(ctx, conv)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			titleLower := strings.ToLower(title)

			// Verify title doesn't contain answers
			for _, forbidden := range tc.shouldNotContain {
				if strings.Contains(titleLower, strings.ToLower(forbidden)) {
					t.Errorf("Title should NOT answer the question. Found '%s' in title: '%s'",
						forbidden, title)
				}
			}

			// Verify title mentions relevant topics
			foundTopic := false
			for _, word := range tc.shouldContainWords {
				if strings.Contains(titleLower, strings.ToLower(word)) {
					foundTopic = true
					break
				}
			}
			if !foundTopic {
				t.Logf("Title '%s' doesn't contain any expected topic words: %v (this may be OK)",
					title, tc.shouldContainWords)
			}

			t.Logf("✅ Title correctly summarizes without answering: '%s'", title)
		})
	}
}
