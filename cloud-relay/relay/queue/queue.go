package queue

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"filippo.io/age"
)

var (
	// ErrQueueFull is returned when the queue has reached its RAM limit.
	ErrQueueFull = errors.New("queue: RAM limit exceeded")

	// ErrNotFound is returned when a message ID is not in the queue.
	ErrNotFound = errors.New("queue: message not found")
)

// QueuedMessage represents a message stored in the encrypted queue.
type QueuedMessage struct {
	ID            string    // Message-ID header or fallback hash
	From          string    // Envelope sender
	To            []string  // Envelope recipients
	EnqueuedAt    time.Time // When the message was queued
	EncryptedData []byte    // Age-encrypted message content
	Checksum      uint32    // CRC32 checksum of encrypted data
	Size          int64     // Size of encrypted data in bytes
	InOverflow    bool      // Whether message is in S3 overflow (Phase 05-02)
	OverflowKey   string    // S3 key if in overflow
}

// QueueConfig holds configuration for the message queue.
type QueueConfig struct {
	KeyPath       string // Path to age identity file
	MaxRAMBytes   int64  // Maximum RAM usage (default 200MB)
	MaxMessages   int    // Maximum number of messages (default 10000)
	TTLHours      int    // Time-to-live in hours (default 168 = 7 days)
	SnapshotPath  string // Path for queue metadata snapshots
}

// MessageQueue manages an encrypted in-memory message queue.
type MessageQueue struct {
	mu              sync.RWMutex
	messages        map[string]*QueuedMessage // Keyed by Message-ID
	order           []string                  // FIFO ordering of message IDs
	recipient       age.Recipient             // For encryption
	identity        age.Identity              // For decryption
	maxRAMBytes     int64
	currentRAMBytes int64
	snapshotPath    string
	maxMessages     int
	ttlHours        int
	overflow        *OverflowStorage // S3-compatible overflow storage (nil if disabled)
}

// NewMessageQueue creates a new encrypted message queue.
func NewMessageQueue(cfg QueueConfig) (*MessageQueue, error) {
	// Load or create age identity
	identity, err := LoadOrCreateIdentity(cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("load queue identity: %w", err)
	}

	// Set defaults
	maxRAMBytes := cfg.MaxRAMBytes
	if maxRAMBytes == 0 {
		maxRAMBytes = 200 * 1024 * 1024 // 200MB default
	}

	maxMessages := cfg.MaxMessages
	if maxMessages == 0 {
		maxMessages = 10000
	}

	ttlHours := cfg.TTLHours
	if ttlHours == 0 {
		ttlHours = 168 // 7 days default
	}

	q := &MessageQueue{
		messages:        make(map[string]*QueuedMessage),
		order:           make([]string, 0),
		recipient:       identity.Recipient(),
		identity:        identity,
		maxRAMBytes:     maxRAMBytes,
		currentRAMBytes: 0,
		snapshotPath:    cfg.SnapshotPath,
		maxMessages:     maxMessages,
		ttlHours:        ttlHours,
	}

	return q, nil
}

// SetOverflow injects overflow storage into the queue.
// Must be called before Enqueue if overflow is needed.
func (q *MessageQueue) SetOverflow(overflow *OverflowStorage) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.overflow = overflow
}

// Enqueue adds a message to the queue with encryption.
// Returns ErrQueueFull if the queue has reached its RAM limit.
func (q *MessageQueue) Enqueue(ctx context.Context, from string, to []string, data io.Reader) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Read message data
	msgData, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("read message data: %w", err)
	}

	// Extract Message-ID from headers
	msgID, err := extractMessageID(msgData)
	if err != nil {
		// Generate fallback ID using SHA-256 hash
		hash := sha256.Sum256(msgData)
		msgID = fmt.Sprintf("<fallback-%s@darkpipe.local>", hex.EncodeToString(hash[:16]))
	}

	// Check for duplicate
	if _, exists := q.messages[msgID]; exists {
		// Message already queued, treat as success (dedup)
		return nil
	}

	// Check message count limit
	if len(q.messages) >= q.maxMessages {
		return ErrQueueFull
	}

	// Encrypt message
	encrypted, checksum, err := Encrypt(msgData, q.recipient)
	if err != nil {
		return fmt.Errorf("encrypt message: %w", err)
	}

	msgSize := int64(len(encrypted))

	// Check RAM limit
	if q.currentRAMBytes+msgSize > q.maxRAMBytes {
		// RAM limit exceeded - try overflow if available
		if q.overflow == nil {
			return ErrQueueFull
		}

		// Spill oldest messages to S3 to free up RAM
		needed := (q.currentRAMBytes + msgSize) - q.maxRAMBytes
		if err := q.overflowOldestMessages(ctx, needed); err != nil {
			return fmt.Errorf("overflow messages: %w", err)
		}
	}

	// Add to queue
	msg := &QueuedMessage{
		ID:            msgID,
		From:          from,
		To:            to,
		EnqueuedAt:    time.Now(),
		EncryptedData: encrypted,
		Checksum:      checksum,
		Size:          msgSize,
		InOverflow:    false,
		OverflowKey:   "",
	}

	q.messages[msgID] = msg
	q.order = append(q.order, msgID)
	q.currentRAMBytes += msgSize

	return nil
}

// Dequeue removes a message from the queue, decrypts it, and returns the plaintext.
// If the message is in overflow storage, it retrieves it from S3 first.
func (q *MessageQueue) Dequeue(id string) (*QueuedMessage, []byte, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	msg, exists := q.messages[id]
	if !exists {
		return nil, nil, ErrNotFound
	}

	// Retrieve encrypted data (from RAM or S3)
	var encryptedData []byte
	if msg.InOverflow {
		// Download from S3 overflow
		if q.overflow == nil {
			return nil, nil, fmt.Errorf("message in overflow but overflow storage not available")
		}

		var err error
		encryptedData, err = q.overflow.Download(context.Background(), msg.OverflowKey)
		if err != nil {
			return nil, nil, fmt.Errorf("download from overflow: %w", err)
		}
	} else {
		encryptedData = msg.EncryptedData
	}

	// Decrypt message
	plaintext, err := Decrypt(encryptedData, msg.Checksum, q.identity)
	if err != nil {
		return nil, nil, fmt.Errorf("decrypt message: %w", err)
	}

	// Remove from queue
	delete(q.messages, id)
	if !msg.InOverflow {
		q.currentRAMBytes -= msg.Size
	}

	// Remove from order slice
	for i, oid := range q.order {
		if oid == id {
			q.order = append(q.order[:i], q.order[i+1:]...)
			break
		}
	}

	return msg, plaintext, nil
}

// Len returns the number of messages in the queue.
func (q *MessageQueue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.messages)
}

// RAMUsage returns the current RAM usage in bytes.
func (q *MessageQueue) RAMUsage() int64 {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.currentRAMBytes
}

// PurgeExpired removes messages older than the TTL.
// Returns the number of messages purged.
func (q *MessageQueue) PurgeExpired() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	expireTime := time.Now().Add(-time.Duration(q.ttlHours) * time.Hour)
	purged := 0

	// Build new order slice without expired messages
	newOrder := make([]string, 0, len(q.order))
	for _, id := range q.order {
		msg := q.messages[id]
		if msg.EnqueuedAt.Before(expireTime) {
			// Expired - remove
			delete(q.messages, id)
			q.currentRAMBytes -= msg.Size
			purged++
		} else {
			newOrder = append(newOrder, id)
		}
	}
	q.order = newOrder

	return purged
}

// Snapshot writes queue metadata to disk atomically.
// Only metadata is persisted (encrypted data stays in memory).
func (q *MessageQueue) Snapshot() error {
	q.mu.RLock()
	defer q.mu.RUnlock()

	// Build metadata snapshot
	type snapshotMessage struct {
		ID          string    `json:"id"`
		From        string    `json:"from"`
		To          []string  `json:"to"`
		EnqueuedAt  time.Time `json:"enqueued_at"`
		Size        int64     `json:"size"`
		InOverflow  bool      `json:"in_overflow"`
		OverflowKey string    `json:"overflow_key,omitempty"`
	}

	type snapshot struct {
		CreatedAt time.Time         `json:"created_at"`
		Messages  []snapshotMessage `json:"messages"`
	}

	snap := snapshot{
		CreatedAt: time.Now(),
		Messages:  make([]snapshotMessage, 0, len(q.order)),
	}

	for _, id := range q.order {
		msg := q.messages[id]
		snap.Messages = append(snap.Messages, snapshotMessage{
			ID:          msg.ID,
			From:        msg.From,
			To:          msg.To,
			EnqueuedAt:  msg.EnqueuedAt,
			Size:        msg.Size,
			InOverflow:  msg.InOverflow,
			OverflowKey: msg.OverflowKey,
		})
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	// Atomic write: temp file + fsync + rename
	dir := filepath.Dir(q.snapshotPath)
	tmpFile, err := os.CreateTemp(dir, "snapshot-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp snapshot file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // Clean up on error

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write snapshot: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("sync snapshot: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close snapshot: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, q.snapshotPath); err != nil {
		return fmt.Errorf("rename snapshot: %w", err)
	}

	return nil
}

// Close writes a final snapshot and cleans up resources.
func (q *MessageQueue) Close() error {
	if q.snapshotPath != "" {
		return q.Snapshot()
	}
	return nil
}

// extractMessageID extracts the Message-ID header from an RFC 5322 message.
func extractMessageID(data []byte) (string, error) {
	// Fallback: scan for Message-ID header manually (simple but works)
	lines := string(data)
	prefix := "Message-ID:"
	start := 0
	for {
		idx := findHeaderLine(lines[start:], prefix)
		if idx == -1 {
			break
		}
		start += idx
		end := start + len(prefix)

		// Find end of line
		lineEnd := end
		for lineEnd < len(lines) && lines[lineEnd] != '\n' {
			lineEnd++
		}

		// Extract value
		value := lines[end:lineEnd]
		// Trim whitespace
		value = trimSpace(value)
		if value != "" {
			return value, nil
		}

		start = lineEnd + 1
		if start >= len(lines) {
			break
		}
	}

	return "", errors.New("Message-ID header not found")
}

// findHeaderLine finds a header line (case-insensitive).
func findHeaderLine(text, header string) int {
	lower := toLower(text)
	lowerHeader := toLower(header)

	idx := 0
	for {
		pos := indexOf(lower[idx:], lowerHeader)
		if pos == -1 {
			return -1
		}
		// Check if it's at start of line
		if idx+pos == 0 || text[idx+pos-1] == '\n' {
			return idx + pos
		}
		idx += pos + 1
	}
}

// Simple helpers to avoid external dependencies
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		b[i] = c
	}
	return string(b)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r' || s[end-1] == '\n') {
		end--
	}

	return s[start:end]
}

// overflowOldestMessages spills oldest messages to S3 overflow storage
// until at least 'needed' bytes of RAM are freed.
// Messages already in overflow are skipped.
func (q *MessageQueue) overflowOldestMessages(ctx context.Context, needed int64) error {
	if q.overflow == nil {
		return fmt.Errorf("overflow storage not configured")
	}

	freed := int64(0)

	// Iterate oldest messages first (FIFO order)
	for _, id := range q.order {
		if freed >= needed {
			break // Enough space freed
		}

		msg := q.messages[id]

		// Skip if already in overflow
		if msg.InOverflow {
			continue
		}

		// Upload encrypted data to S3
		key, err := q.overflow.Upload(ctx, msg.ID, msg.EncryptedData)
		if err != nil {
			return fmt.Errorf("upload message %s to overflow: %w", msg.ID, err)
		}

		// Mark as in overflow and free RAM
		msg.InOverflow = true
		msg.OverflowKey = key
		freed += msg.Size
		q.currentRAMBytes -= msg.Size

		// Nil out encrypted data to release memory
		msg.EncryptedData = nil
	}

	return nil
}
