package jwe

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/go-jose/go-jose/v3"
	"github.com/kshitiz1403/jsonjuggler/activities"
	"github.com/kshitiz1403/jsonjuggler/logger"
	"github.com/kshitiz1403/jsonjuggler/utils"
)

type EncryptArgs struct {
	Payload              string `arg:"payload" required:"true"`
	PublicKey            string `arg:"publicKey" required:"true"`
	ContentEncryptionAlg string `arg:"contentEncryptionAlgorithm" required:"true" validate:"oneof=A128GCM A256GCM"`
	KeyManagementAlg     string `arg:"keyManagementAlgorithm" required:"true" validate:"oneof=RSA-OAEP RSA-OAEP-256"`
}

type EncryptActivity struct {
	*activities.BaseActivity
}

func New(activityName string, logger logger.Logger) *EncryptActivity {
	return &EncryptActivity{
		BaseActivity: &activities.BaseActivity{
			ActivityName: activityName,
			Logger:       logger,
		},
	}
}

func (a *EncryptActivity) Execute(ctx context.Context, arguments map[string]any) (interface{}, error) {
	var encryptArgs EncryptArgs
	if err := utils.ParseAndValidateArgs(ctx, arguments, &encryptArgs); err != nil {
		a.GetLogger().ErrorContextf(ctx, "Invalid JWE encryption arguments: %v", err)
		return nil, activities.NewActivityError(
			activities.ErrInvalidArguments,
			"Invalid JWE encryption arguments",
			"JWEEncrypt",
		).WithArguments(arguments).WithCause(err)
	}

	a.GetLogger().DebugContext(ctx, "Starting JWE encryption")
	a.GetLogger().DebugContextf(ctx, "Using content encryption algorithm: %s, key management algorithm: %s",
		encryptArgs.ContentEncryptionAlg, encryptArgs.KeyManagementAlg)

	// Parse public key
	block, _ := pem.Decode([]byte(encryptArgs.PublicKey))
	if block == nil {
		a.GetLogger().ErrorContext(ctx, "Failed to parse PEM block containing public key")
		return nil, activities.NewActivityError(
			activities.ErrJWEEncryptError,
			"Failed to parse PEM block containing public key",
			"JWEEncrypt",
		).WithArguments(map[string]interface{}{
			"publicKey": "(redacted)",
		})
	}
	a.GetLogger().DebugContext(ctx, "Successfully decoded PEM block")

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		a.GetLogger().ErrorContextf(ctx, "Failed to parse public key: %v", err)
		return nil, activities.NewActivityError(
			activities.ErrJWEEncryptError,
			"Failed to parse public key",
			"JWEEncrypt",
		).WithArguments(map[string]interface{}{
			"publicKey": "(redacted)",
		}).WithCause(err)
	}
	a.GetLogger().DebugContext(ctx, "Successfully parsed public key")

	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		a.GetLogger().ErrorContextf(ctx, "Expected RSA public key, got: %T", pub)
		return nil, activities.NewActivityError(
			activities.ErrJWEEncryptError,
			"Not an RSA public key",
			"JWEEncrypt",
		).WithArguments(map[string]interface{}{
			"publicKey": "(redacted)",
			"keyType":   fmt.Sprintf("%T", pub),
		})
	}
	a.GetLogger().DebugContext(ctx, "Successfully validated RSA public key")

	// Create encrypter
	recipient := jose.Recipient{
		Algorithm: jose.KeyAlgorithm(encryptArgs.KeyManagementAlg),
		Key:       rsaKey,
	}

	enc, err := jose.NewEncrypter(
		jose.ContentEncryption(encryptArgs.ContentEncryptionAlg),
		recipient,
		nil,
	)
	if err != nil {
		a.GetLogger().ErrorContextf(ctx, "Failed to create encrypter: %v", err)
		return nil, activities.NewActivityError(
			activities.ErrJWEEncryptError,
			"Failed to create encrypter",
			"JWEEncrypt",
		).WithArguments(map[string]interface{}{
			"keyManagementAlg":     encryptArgs.KeyManagementAlg,
			"contentEncryptionAlg": encryptArgs.ContentEncryptionAlg,
		}).WithCause(err)
	}
	a.GetLogger().DebugContext(ctx, "Successfully created encrypter")

	// Encrypt payload
	obj, err := enc.Encrypt([]byte(encryptArgs.Payload))
	if err != nil {
		a.GetLogger().ErrorContextf(ctx, "Failed to encrypt payload: %v", err)
		return nil, activities.NewActivityError(
			activities.ErrJWEEncryptError,
			"Failed to encrypt payload",
			"JWEEncrypt",
		).WithCause(err)
	}
	a.GetLogger().DebugContext(ctx, "Successfully encrypted payload")

	serialized, err := obj.CompactSerialize()
	if err != nil {
		a.GetLogger().ErrorContextf(ctx, "Failed to serialize encrypted payload: %v", err)
		return nil, activities.NewActivityError(
			activities.ErrJWEEncryptError,
			"Failed to serialize encrypted payload",
			"JWEEncrypt",
		).WithCause(err)
	}
	a.GetLogger().DebugContext(ctx, "Successfully serialized encrypted payload")

	return serialized, nil
}
