package bot

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"valid email", "user@example.com", true},
		{"valid email with subdomain", "user@sub.example.com", true},
		{"valid email with numbers", "user123@example.com", true},
		{"valid email with dots", "user.name@example.com", true},
		{"valid email with plus", "user+tag@example.com", true},
		{"empty email", "", false},
		{"no @ symbol", "userexample.com", false},
		{"multiple @ symbols", "user@example@com", false},
		{"no local part", "@example.com", false},
		{"no domain part", "user@", false},
		{"no domain dot", "user@example", false},
		{"short domain", "user@a.b", false},
		{"just @", "@", false},
		{"just dot", ".", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEmail(tt.email)
			if result != tt.expected {
				t.Errorf("isValidEmail(%q) = %v, want %v", tt.email, result, tt.expected)
			}
		})
	}
}

func TestGetUserFriendlyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "invalid email format",
			err:      status.Error(codes.InvalidArgument, "invalid email format"),
			expected: "Неверный формат email. Пожалуйста, введите корректный email адрес.",
		},
		{
			name:     "invalid verification code",
			err:      status.Error(codes.InvalidArgument, "invalid verification code"),
			expected: "Неверный код подтверждения. Пожалуйста, проверьте код и попробуйте снова.",
		},
		{
			name:     "not found",
			err:      status.Error(codes.NotFound, "user not found"),
			expected: "Данные не найдены. Попробуйте снова.",
		},
		{
			name:     "permission denied",
			err:      status.Error(codes.PermissionDenied, "access denied"),
			expected: "Доступ запрещен. Проверьте права доступа.",
		},
		{
			name:     "unavailable",
			err:      status.Error(codes.Unavailable, "service unavailable"),
			expected: "Сервис временно недоступен. Попробуйте позже.",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUserFriendlyError(tt.err)
			if result != tt.expected {
				t.Errorf("GetUserFriendlyError() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "unavailable error",
			err:      status.Error(codes.Unavailable, "service unavailable"),
			expected: true,
		},
		{
			name:     "resource exhausted",
			err:      status.Error(codes.ResourceExhausted, "rate limit exceeded"),
			expected: true,
		},
		{
			name:     "deadline exceeded",
			err:      status.Error(codes.DeadlineExceeded, "timeout"),
			expected: true,
		},
		{
			name:     "invalid argument",
			err:      status.Error(codes.InvalidArgument, "invalid email"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryableError() = %v, want %v", result, tt.expected)
			}
		})
	}
}
