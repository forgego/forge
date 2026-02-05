package utils

import (
	"context"
	"net/http"

	"github.com/forgego/forge/server"
)

// MessageLevel represents the level of a message
type MessageLevel string

const (
	MessageSuccess MessageLevel = "success"
	MessageInfo    MessageLevel = "info"
	MessageWarning MessageLevel = "warning"
	MessageError   MessageLevel = "error"
)

// Message represents a flash message
type Message struct {
	Level   MessageLevel
	Message string
}

// MessageUser sends a message to the user (stores in session for flash messages)
func MessageUser(ctx context.Context, r *http.Request, message string, level MessageLevel) error {
	// Get session manager from context or request
	// In a real implementation, this would be injected via middleware
	sessionManager := getSessionManager(r)
	if sessionManager == nil {
		// If no session manager, we can't store messages
		// This is acceptable for development/testing
		return nil
	}

	// Get existing messages
	messages := getMessages(sessionManager, r)
	
	// Add new message
	messages = append(messages, Message{
		Level:   level,
		Message: message,
	})

	// Store back in session
	sessionManager.Put(r, "admin_messages", messages)
	return nil
}

// GetMessages retrieves messages from session and clears them
func GetMessages(sessionManager *server.SessionManager, r *http.Request) []Message {
	messages := getMessages(sessionManager, r)
	// Clear messages after retrieving
	sessionManager.Remove(r, "admin_messages")
	return messages
}

// getMessages retrieves messages without clearing
func getMessages(sessionManager *server.SessionManager, r *http.Request) []Message {
	if sessionManager == nil {
		return []Message{}
	}

	val := sessionManager.Get(r, "admin_messages")
	if val == nil {
		return []Message{}
	}

	messages, ok := val.([]Message)
	if !ok {
		return []Message{}
	}

	return messages
}

// getSessionManager extracts session manager from request context
// In a real implementation, this would be set by middleware
func getSessionManager(r *http.Request) *server.SessionManager {
	// Try to get from context
	if sm, ok := r.Context().Value("session_manager").(*server.SessionManager); ok {
		return sm
	}
	return nil
}

// Success sends a success message
func Success(ctx context.Context, r *http.Request, message string) error {
	return MessageUser(ctx, r, message, MessageSuccess)
}

// Info sends an info message
func Info(ctx context.Context, r *http.Request, message string) error {
	return MessageUser(ctx, r, message, MessageInfo)
}

// Warning sends a warning message
func Warning(ctx context.Context, r *http.Request, message string) error {
	return MessageUser(ctx, r, message, MessageWarning)
}

// Error sends an error message
func Error(ctx context.Context, r *http.Request, message string) error {
	return MessageUser(ctx, r, message, MessageError)
}
