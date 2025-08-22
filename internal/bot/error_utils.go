package bot

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetUserFriendlyError возвращает понятное пользователю сообщение об ошибке
func GetUserFriendlyError(err error) string {
	if err == nil {
		return ""
	}

	// Проверяем, является ли это gRPC ошибкой
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			// Извлекаем сообщение из деталей ошибки
			msg := st.Message()
			if strings.Contains(msg, "invalid email format") {
				return "Неверный формат email. Пожалуйста, введите корректный email адрес."
			}
			if strings.Contains(msg, "invalid verification code") {
				return "Неверный код подтверждения. Пожалуйста, проверьте код и попробуйте снова."
			}
			if strings.Contains(msg, "email") {
				return "Ошибка с email: " + msg
			}
			return "Неверные данные: " + msg
		case codes.NotFound:
			return "Данные не найдены. Попробуйте снова."
		case codes.PermissionDenied:
			return "Доступ запрещен. Проверьте права доступа."
		case codes.Unauthenticated:
			return "Требуется авторизация. Выполните вход снова."
		case codes.ResourceExhausted:
			return "Превышен лимит запросов. Попробуйте позже."
		case codes.Unavailable:
			return "Сервис временно недоступен. Попробуйте позже."
		case codes.DeadlineExceeded:
			return "Превышено время ожидания. Попробуйте снова."
		case codes.Internal:
			return "Внутренняя ошибка сервера. Попробуйте позже."
		default:
			return "Произошла ошибка: " + st.Message()
		}
	}

	// Если это не gRPC ошибка, проверяем содержимое сообщения
	errMsg := err.Error()
	if strings.Contains(errMsg, "invalid email format") {
		return "Неверный формат email. Пожалуйста, введите корректный email адрес."
	}
	if strings.Contains(errMsg, "invalid verification code") {
		return "Неверный код подтверждения. Пожалуйста, проверьте код и попробуйте снова."
	}
	if strings.Contains(errMsg, "email") {
		return "Ошибка с email: " + errMsg
	}
	if strings.Contains(errMsg, "network") || strings.Contains(errMsg, "connection") {
		return "Ошибка соединения. Проверьте интернет и попробуйте снова."
	}

	// Общая ошибка
	return "Произошла ошибка: " + errMsg
}

// IsRetryableError проверяет, можно ли повторить запрос при этой ошибке
func IsRetryableError(err error) bool {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Unavailable, codes.ResourceExhausted, codes.DeadlineExceeded:
			return true
		}
	}
	return false
}

// isValidEmail выполняет простую валидацию email
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	
	// Проверяем наличие @ символа
	if !strings.Contains(email, "@") {
		return false
	}
	
	// Проверяем, что есть часть до и после @
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	
	localPart := parts[0]
	domainPart := parts[1]
	
	// Проверяем, что части не пустые
	if localPart == "" || domainPart == "" {
		return false
	}
	
	// Проверяем, что домен содержит точку
	if !strings.Contains(domainPart, ".") {
		return false
	}
	
	// Проверяем минимальную длину
	if len(localPart) < 1 || len(domainPart) < 4 {
		return false
	}
	
	return true
}
