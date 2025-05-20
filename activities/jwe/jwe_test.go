package jwe

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/go-jose/go-jose/v3"
	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/logger/zap"
	"github.com/stretchr/testify/require"
)

func generateRSAKeyPair(t *testing.T) (string, *rsa.PrivateKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM), privateKey
}

func TestJWEEncrypt(t *testing.T) {
	activity := New("JWEEncrypt", zap.NewLogger(logger.DebugLevel))
	publicKeyPEM, privateKey := generateRSAKeyPair(t)

	tests := []struct {
		name        string
		args        map[string]any
		validate    func(*testing.T, string)
		expectError bool
	}{
		{
			name: "Successful Encryption with RSA-OAEP",
			args: map[string]any{
				"payload":                    "sensitive data",
				"publicKey":                  publicKeyPEM,
				"contentEncryptionAlgorithm": "A256GCM",
				"keyManagementAlgorithm":     "RSA-OAEP",
			},
			validate: func(t *testing.T, result string) {
				// Parse the JWE
				object, err := jose.ParseEncrypted(result)
				require.NoError(t, err)

				// Decrypt the payload
				decrypted, err := object.Decrypt(privateKey)
				require.NoError(t, err)
				require.Equal(t, "sensitive data", string(decrypted))
			},
		},
		{
			name: "Successful Encryption with RSA-OAEP-256",
			args: map[string]any{
				"payload":                    "sensitive data",
				"publicKey":                  publicKeyPEM,
				"contentEncryptionAlgorithm": "A128GCM",
				"keyManagementAlgorithm":     "RSA-OAEP-256",
			},
			validate: func(t *testing.T, result string) {
				object, err := jose.ParseEncrypted(result)
				require.NoError(t, err)

				decrypted, err := object.Decrypt(privateKey)
				require.NoError(t, err)
				require.Equal(t, "sensitive data", string(decrypted))
			},
		},
		{
			name: "Invalid Public Key",
			args: map[string]any{
				"payload":                    "sensitive data",
				"publicKey":                  "invalid key",
				"contentEncryptionAlgorithm": "A256GCM",
				"keyManagementAlgorithm":     "RSA-OAEP",
			},
			expectError: true,
		},
		{
			name: "Invalid Content Encryption Algorithm",
			args: map[string]any{
				"payload":                    "sensitive data",
				"publicKey":                  publicKeyPEM,
				"contentEncryptionAlgorithm": "INVALID-ALG",
				"keyManagementAlgorithm":     "RSA-OAEP",
			},
			expectError: true,
		},
		{
			name: "Missing Required Arguments",
			args: map[string]any{
				"payload": "sensitive data",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := activity.Execute(context.Background(), tt.args)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			resultStr, ok := result.(string)
			require.True(t, ok, "Result should be a string")
			tt.validate(t, resultStr)
		})
	}
}
