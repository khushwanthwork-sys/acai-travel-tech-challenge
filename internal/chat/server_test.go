package chat

import (
	"context"
	"testing"

	"github.com/acai-travel/tech-challenge/internal/chat/model"
	. "github.com/acai-travel/tech-challenge/internal/chat/testing"
	"github.com/acai-travel/tech-challenge/internal/pb"
	"github.com/google/go-cmp/cmp"
	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/testing/protocmp"
)

// MockAssistant for testing without calling OpenAI
type MockAssistant struct {
	TitleFunc func(ctx context.Context, conv *model.Conversation) (string, error)
	ReplyFunc func(ctx context.Context, conv *model.Conversation) (string, error)
}

func (m *MockAssistant) Title(ctx context.Context, conv *model.Conversation) (string, error) {
	if m.TitleFunc != nil {
		return m.TitleFunc(ctx, conv)
	}
	return "Test Title", nil
}

func (m *MockAssistant) Reply(ctx context.Context, conv *model.Conversation) (string, error) {
	if m.ReplyFunc != nil {
		return m.ReplyFunc(ctx, conv)
	}
	return "Test Reply", nil
}

func TestServer_StartConversation(t *testing.T) {
	ctx := context.Background()

	t.Run("creates new conversation successfully", WithFixture(func(t *testing.T, f *Fixture) {
		mockAssist := &MockAssistant{
			TitleFunc: func(ctx context.Context, conv *model.Conversation) (string, error) {
				return "Weather Inquiry", nil
			},
			ReplyFunc: func(ctx context.Context, conv *model.Conversation) (string, error) {
				return "The weather is sunny today!", nil
			},
		}

		server := NewServer(f.Repository, mockAssist)

		req := &pb.StartConversationRequest{
			Message: "What's the weather like?",
		}

		resp, err := server.StartConversation(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp == nil {
			t.Fatal("expected response, got nil")
		}

		// Verify response fields
		if resp.ConversationId == "" {
			t.Error("conversation ID should not be empty")
		}
		if resp.Title != "Weather Inquiry" {
			t.Errorf("expected title 'Weather Inquiry', got '%s'", resp.Title)
		}
		if resp.Reply != "The weather is sunny today!" {
			t.Errorf("expected reply 'The weather is sunny today!', got '%s'", resp.Reply)
		}

		// Verify conversation was saved to database
		conv, err := f.Repository.DescribeConversation(ctx, resp.ConversationId)
		if err != nil {
			t.Fatalf("failed to retrieve conversation from database: %v", err)
		}
		if conv == nil {
			t.Fatal("expected conversation in database, got nil")
		}

		// Verify conversation details
		if conv.Title != "Weather Inquiry" {
			t.Errorf("expected conversation title 'Weather Inquiry', got '%s'", conv.Title)
		}
		if len(conv.Messages) != 2 {
			t.Errorf("expected 2 messages (user + assistant), got %d", len(conv.Messages))
		}

		// Verify first message (user)
		if len(conv.Messages) > 0 {
			msg := conv.Messages[0]
			if msg.Role != model.RoleUser {
				t.Errorf("expected first message role to be User, got %s", msg.Role)
			}
			if msg.Content != "What's the weather like?" {
				t.Errorf("expected first message content 'What's the weather like?', got '%s'", msg.Content)
			}
		}

		// Verify second message (assistant)
		if len(conv.Messages) > 1 {
			msg := conv.Messages[1]
			if msg.Role != model.RoleAssistant {
				t.Errorf("expected second message role to be Assistant, got %s", msg.Role)
			}
			if msg.Content != "The weather is sunny today!" {
				t.Errorf("expected second message content 'The weather is sunny today!', got '%s'", msg.Content)
			}
		}

		// Cleanup
		if err := f.Repository.DeleteConversation(ctx, resp.ConversationId); err != nil {
			t.Logf("failed to cleanup conversation: %v", err)
		}
	}))

	t.Run("populates title even if title generation fails", WithFixture(func(t *testing.T, f *Fixture) {
		mockAssist := &MockAssistant{
			TitleFunc: func(ctx context.Context, conv *model.Conversation) (string, error) {
				return "", twirp.InternalError("title generation failed")
			},
			ReplyFunc: func(ctx context.Context, conv *model.Conversation) (string, error) {
				return "Reply works fine", nil
			},
		}

		server := NewServer(f.Repository, mockAssist)

		req := &pb.StartConversationRequest{
			Message: "Hello!",
		}

		resp, err := server.StartConversation(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp == nil {
			t.Fatal("expected response, got nil")
		}

		// Should still return response with default title
		if resp.Title != "Untitled conversation" {
			t.Errorf("expected default title 'Untitled conversation', got '%s'", resp.Title)
		}
		if resp.Reply != "Reply works fine" {
			t.Errorf("expected reply 'Reply works fine', got '%s'", resp.Reply)
		}

		// Cleanup
		if err := f.Repository.DeleteConversation(ctx, resp.ConversationId); err != nil {
			t.Logf("failed to cleanup conversation: %v", err)
		}
	}))

	t.Run("returns error if reply generation fails", WithFixture(func(t *testing.T, f *Fixture) {
		mockAssist := &MockAssistant{
			TitleFunc: func(ctx context.Context, conv *model.Conversation) (string, error) {
				return "Test Title", nil
			},
			ReplyFunc: func(ctx context.Context, conv *model.Conversation) (string, error) {
				return "", twirp.InternalError("reply generation failed")
			},
		}

		server := NewServer(f.Repository, mockAssist)

		req := &pb.StartConversationRequest{
			Message: "Hello!",
		}

		resp, err := server.StartConversation(ctx, req)

		if err == nil {
			t.Fatal("expected error for reply generation failure, got nil")
		}
		if resp != nil {
			t.Errorf("expected nil response on error, got %v", resp)
		}
	}))

	t.Run("requires non-empty message", WithFixture(func(t *testing.T, f *Fixture) {
		server := NewServer(f.Repository, &MockAssistant{})

		req := &pb.StartConversationRequest{
			Message: "   ", // Only whitespace
		}

		resp, err := server.StartConversation(ctx, req)

		if err == nil {
			t.Fatal("expected error for empty message, got nil")
		}
		if resp != nil {
			t.Errorf("expected nil response on error, got %v", resp)
		}

		// Verify it's a RequiredArgument error
		if te, ok := err.(twirp.Error); ok {
			if te.Code() != twirp.InvalidArgument {
				t.Errorf("expected InvalidArgument error code, got %s", te.Code())
			}
			if te.Meta("argument") != "message" {
				t.Errorf("expected error about 'message' argument, got '%s'", te.Meta("argument"))
			}
		} else {
			t.Errorf("expected twirp.Error, got %T", err)
		}
	}))

	t.Run("triggers assistant's response", WithFixture(func(t *testing.T, f *Fixture) {
		var titleCalled, replyCalled bool
		var titleConvMessages, replyConvMessages int

		mockAssist := &MockAssistant{
			TitleFunc: func(ctx context.Context, conv *model.Conversation) (string, error) {
				titleCalled = true
				titleConvMessages = len(conv.Messages)
				if len(conv.Messages) > 0 && conv.Messages[0].Content != "Test message" {
					t.Errorf("expected first message 'Test message', got '%s'", conv.Messages[0].Content)
				}
				return "Generated Title", nil
			},
			ReplyFunc: func(ctx context.Context, conv *model.Conversation) (string, error) {
				replyCalled = true
				replyConvMessages = len(conv.Messages)
				if len(conv.Messages) > 0 && conv.Messages[0].Content != "Test message" {
					t.Errorf("expected first message 'Test message', got '%s'", conv.Messages[0].Content)
				}
				return "Generated Reply", nil
			},
		}

		server := NewServer(f.Repository, mockAssist)

		req := &pb.StartConversationRequest{
			Message: "Test message",
		}

		resp, err := server.StartConversation(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp == nil {
			t.Fatal("expected response, got nil")
		}

		// Verify both methods were called
		if !titleCalled {
			t.Error("Title method should have been called")
		}
		if !replyCalled {
			t.Error("Reply method should have been called")
		}

		// Verify conversation had user message when passed to assistant
		if titleConvMessages != 1 {
			t.Errorf("expected Title to receive conversation with 1 message, got %d", titleConvMessages)
		}
		if replyConvMessages != 1 {
			t.Errorf("expected Reply to receive conversation with 1 message, got %d", replyConvMessages)
		}

		// Cleanup
		if err := f.Repository.DeleteConversation(ctx, resp.ConversationId); err != nil {
			t.Logf("failed to cleanup conversation: %v", err)
		}
	}))
}

// Keep the existing TestServer_DescribeConversation test below...
func TestServer_DescribeConversation(t *testing.T) {
	ctx := context.Background()
	srv := NewServer(model.New(ConnectMongo()), nil)

	t.Run("describe existing conversation", WithFixture(func(t *testing.T, f *Fixture) {
		c := f.CreateConversation()

		out, err := srv.DescribeConversation(ctx, &pb.DescribeConversationRequest{ConversationId: c.ID.Hex()})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, want := out.GetConversation(), c.Proto()
		if !cmp.Equal(got, want, protocmp.Transform()) {
			t.Errorf("DescribeConversation() mismatch (-got +want):\n%s", cmp.Diff(got, want, protocmp.Transform()))
		}
	}))

	t.Run("describe non existing conversation should return 404", WithFixture(func(t *testing.T, f *Fixture) {
		_, err := srv.DescribeConversation(ctx, &pb.DescribeConversationRequest{ConversationId: "08a59244257c872c5943e2a2"})
		if err == nil {
			t.Fatal("expected error for non-existing conversation, got nil")
		}

		if te, ok := err.(twirp.Error); !ok || te.Code() != twirp.NotFound {
			t.Fatalf("expected twirp.NotFound error, got %v", err)
		}
	}))
}
