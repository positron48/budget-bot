package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyAuthCodeResult_NewResult(t *testing.T) {
	result := &VerifyAuthCodeResult{
		SessionID: "session123",
	}

	assert.NotNil(t, result)
	assert.Equal(t, "session123", result.SessionID)
}
