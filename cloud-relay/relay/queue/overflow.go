// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package queue

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// OverflowStorage provides S3-compatible overflow storage for queued messages.
// Messages are stored already-encrypted (age encryption happens in queue.go before upload).
type OverflowStorage struct {
	client     *minio.Client
	bucketName string
}

// NewOverflowStorage creates a new S3-compatible overflow storage client.
// It verifies the bucket exists and creates it if needed.
func NewOverflowStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*OverflowStorage, error) {
	// Create MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create S3 client: %w", err)
	}

	// Check if bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket exists: %w", err)
	}

	// Create bucket if it doesn't exist
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket %s: %w", bucket, err)
		}
	}

	return &OverflowStorage{
		client:     client,
		bucketName: bucket,
	}, nil
}

// Upload stores encrypted message data in S3.
// The data is already age-encrypted before calling this function.
// Returns the S3 key where the data was stored.
func (s *OverflowStorage) Upload(ctx context.Context, messageID string, data []byte) (string, error) {
	// Generate safe S3 key from Message-ID
	key := sanitizeMessageIDToKey(messageID)

	// Upload to S3
	_, err := s.client.PutObject(ctx, s.bucketName, key, strings.NewReader(string(data)), int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", fmt.Errorf("upload to S3: %w", err)
	}

	return key, nil
}

// Download retrieves encrypted message data from S3.
func (s *OverflowStorage) Download(ctx context.Context, key string) ([]byte, error) {
	// Get object from S3
	obj, err := s.client.GetObject(ctx, s.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get S3 object: %w", err)
	}
	defer obj.Close()

	// Read all data
	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read S3 object: %w", err)
	}

	return data, nil
}

// Delete removes a message from S3 after successful delivery.
func (s *OverflowStorage) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucketName, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete S3 object: %w", err)
	}
	return nil
}

// List returns all S3 keys with the specified prefix.
// Useful for recovery operations.
func (s *OverflowStorage) List(ctx context.Context, prefix string) ([]string, error) {
	keys := []string{}

	// List objects with prefix
	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for obj := range objectCh {
		if obj.Err != nil {
			return nil, fmt.Errorf("list S3 objects: %w", obj.Err)
		}
		keys = append(keys, obj.Key)
	}

	return keys, nil
}

// sanitizeMessageIDToKey converts a Message-ID to a safe S3 key.
// Format: darkpipe/queue/{sha256-hash-of-message-id}
// This avoids issues with special characters in Message-IDs (angle brackets, @, etc.)
func sanitizeMessageIDToKey(messageID string) string {
	// Hash the Message-ID to get a safe filename
	hash := sha256.Sum256([]byte(messageID))
	hashHex := hex.EncodeToString(hash[:])

	return fmt.Sprintf("darkpipe/queue/%s", hashHex)
}

// sanitizeForFilesystem is a fallback approach that replaces unsafe characters.
// Not used in favor of hash-based approach, but kept for reference.
func sanitizeForFilesystem(s string) string {
	// Remove angle brackets
	s = strings.Trim(s, "<>")

	// Replace unsafe characters with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-._]`)
	s = reg.ReplaceAllString(s, "_")

	// Truncate to reasonable length
	if len(s) > 200 {
		s = s[:200]
	}

	return s
}
