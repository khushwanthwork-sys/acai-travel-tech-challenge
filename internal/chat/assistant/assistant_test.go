package assistant

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/acai-travel/tech-challenge/internal/chat/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Skip these tests in CI or when OPENAI_API_KEY is not set
func skipIfNoOpenAI(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}
}

func TestTitle(t *testing.T) {
	skipIfNoOpenAI(t)

	ctx := context.Background()
	assist := New()

	t.Run("generates title from user message", func(t *testing.T) {
		conv := &model.Conversation{
			ID:        primitive.NewObjectID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages: []*model.Message{{
				ID:        primitive.NewObjectID(),
				Role:      model.RoleUser,
				Content:   "What is the weather like in Barcelona today?",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}},
		}

		title, err := assist.Title(ctx, conv)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if title == "" {
			t.Error("title should not be empty")
		}

		if len(title) > 80 {
			t.Errorf("title should not exceed 80 characters, got %d", len(title))
		}

		t.Logf("Generated title: %s", title)
	})

	t.Run("handles empty conversation", func(t *testing.T) {
		conv := &model.Conversation{
			ID:        primitive.NewObjectID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []*model.Message{},
		}

		title, err := assist.Title(ctx, conv)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if title != "An empty conversation" {
			t.Errorf("expected 'An empty conversation', got '%s'", title)
		}
	})

	t.Run("generates concise title without answering", func(t *testing.T) {
		conv := &model.Conversation{
			ID:        primitive.NewObjectID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages: []*model.Message{{
				ID:        primitive.NewObjectID(),
				Role:      model.RoleUser,
				Content:   "What is 25 + 17?",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}},
		}

		title, err := assist.Title(ctx, conv)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Title should NOT contain the answer "42"
		// It should be something like "Addition Question" or "Math Calculation"
		if title == "42" || title == "The answer is 42" {
			t.Errorf("title should not answer the question, got '%s'", title)
		}

		t.Logf("Generated title: %s", title)
	})
}
