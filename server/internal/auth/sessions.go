package auth

// Idle timeout: compare now() with last_seen_at.
// Absolute timeout: compare now() with absolute_expires_at.
// Renewal timeout: compare now() with created_at.
