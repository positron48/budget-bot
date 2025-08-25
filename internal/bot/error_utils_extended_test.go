package bot

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetUserFriendlyError_GRPCInvalidArgument_Email(t *testing.T) {
	err := status.Error(codes.InvalidArgument, "invalid email format")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Неверный формат email. Пожалуйста, введите корректный email адрес.", result)
}

func TestGetUserFriendlyError_GRPCInvalidArgument_VerificationCode(t *testing.T) {
	err := status.Error(codes.InvalidArgument, "invalid verification code")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Неверный код подтверждения. Пожалуйста, проверьте код и попробуйте снова.", result)
}

func TestGetUserFriendlyError_GRPCInvalidArgument_EmailGeneric(t *testing.T) {
	err := status.Error(codes.InvalidArgument, "email already exists")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Ошибка с email: email already exists", result)
}

func TestGetUserFriendlyError_GRPCInvalidArgument_Generic(t *testing.T) {
	err := status.Error(codes.InvalidArgument, "invalid data")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Неверные данные: invalid data", result)
}

func TestGetUserFriendlyError_GRPCNotFound(t *testing.T) {
	err := status.Error(codes.NotFound, "user not found")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Данные не найдены. Попробуйте снова.", result)
}

func TestGetUserFriendlyError_GRPCPermissionDenied(t *testing.T) {
	err := status.Error(codes.PermissionDenied, "access denied")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Доступ запрещен. Проверьте права доступа.", result)
}

func TestGetUserFriendlyError_GRPCUnauthenticated(t *testing.T) {
	err := status.Error(codes.Unauthenticated, "token expired")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Требуется авторизация. Выполните вход снова.", result)
}

func TestGetUserFriendlyError_GRPCResourceExhausted(t *testing.T) {
	err := status.Error(codes.ResourceExhausted, "rate limit exceeded")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Превышен лимит запросов. Попробуйте позже.", result)
}

func TestGetUserFriendlyError_GRPCUnavailable(t *testing.T) {
	err := status.Error(codes.Unavailable, "service down")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Сервис временно недоступен. Попробуйте позже.", result)
}

func TestGetUserFriendlyError_GRPCDeadlineExceeded(t *testing.T) {
	err := status.Error(codes.DeadlineExceeded, "timeout")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Превышено время ожидания. Попробуйте снова.", result)
}

func TestGetUserFriendlyError_GRPCInternal(t *testing.T) {
	err := status.Error(codes.Internal, "server error")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Внутренняя ошибка сервера. Попробуйте позже.", result)
}

func TestGetUserFriendlyError_GRPCUnknown(t *testing.T) {
	err := status.Error(codes.Unknown, "unknown error")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Произошла ошибка: unknown error", result)
}

func TestGetUserFriendlyError_RegularError_EmailFormat(t *testing.T) {
	err := errors.New("invalid email format")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Неверный формат email. Пожалуйста, введите корректный email адрес.", result)
}

func TestGetUserFriendlyError_RegularError_VerificationCode(t *testing.T) {
	err := errors.New("invalid verification code")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Неверный код подтверждения. Пожалуйста, проверьте код и попробуйте снова.", result)
}

func TestGetUserFriendlyError_RegularError_EmailGeneric(t *testing.T) {
	err := errors.New("email validation failed")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Ошибка с email: email validation failed", result)
}

func TestGetUserFriendlyError_RegularError_Network(t *testing.T) {
	err := errors.New("network connection failed")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Ошибка соединения. Проверьте интернет и попробуйте снова.", result)
}

func TestGetUserFriendlyError_RegularError_Connection(t *testing.T) {
	err := errors.New("connection refused")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Ошибка соединения. Проверьте интернет и попробуйте снова.", result)
}

func TestGetUserFriendlyError_RegularError_Generic(t *testing.T) {
	err := errors.New("some other error")
	result := GetUserFriendlyError(err)
	assert.Equal(t, "Произошла ошибка: some other error", result)
}
