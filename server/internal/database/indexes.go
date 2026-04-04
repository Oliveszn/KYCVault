package database

import (
	"errors"
	"fmt"
)

func CreateIndexes() error {
	if DB == nil {
		return errors.New("database is not initialized")
	}

	//KYCSESSION
	// Only one active (non-terminal) session per user at a time.
	err := DB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_kyc_sessions_one_active_per_user
		ON kyc_sessions (user_id)
		WHERE status NOT IN ('approved', 'rejected');
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create partial index: %w", err)
	}

	// Fast lookups by vendor reference
	err = DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_kyc_sessions_vendor_session_id
		ON kyc_sessions (vendor_session_id)
		WHERE vendor_session_id IS NOT NULL;
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create vendor index: %w", err)
	}

	// Status filtering for admin dashboard queue.
	err = DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_kyc_sessions_status
		ON kyc_sessions (status);
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create status index: %w", err)
	}

	//KYCDOCUMENT
	//A session can have at most one accepted image per side.
	err = DB.Exec(`
	CREATE UNIQUE INDEX IF NOT EXISTS idx_kyc_documents_one_accepted_per_side
	ON kyc_documents (session_id, side)
	WHERE status = 'accepted';
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create session index: %w", err)
	}

	//Integrity check lookups.
	err = DB.Exec(`
	CREATE INDEX IF NOT EXISTS idx_kyc_documents_checksum ON kyc_documents (checksum);
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create integrity check index: %w", err)
	}

	//WEBHOOK ENDPOINT
	// Active endpoints only, for dispatcher lookups.
	err = DB.Exec(`
	CREATE INDEX IF NOT EXISTS idx_webhook_endpoints_active ON webhook_endpoints (user_id)
	WHERE is_active = true AND deleted_at IS NULL;
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create active endpoints only index: %w", err)
	}

	//WEBHOOK DELIVERIES
	// Retry worker polls this to find jobs ready to fire.
	err = DB.Exec(`
	CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_retry_queue
	ON webhook_deliveries (next_retry_at ASC)
	WHERE status = 'pending' AND next_retry_at IS NOT NULL;
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create retry worker index: %w", err)
	}

	//all deliveries for a session. for admin
	err = DB.Exec(`
	CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_session ON webhook_deliveries (session_id, created_at DESC);
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create admin deliveries index: %w", err)
	}

	// Idempotency dedup check.
	err = DB.Exec(`
	CREATE UNIQUE INDEX IF NOT EXISTS idx_webhook_deliveries_idempotency ON webhook_deliveries (idempotency_key)
	WHERE idempotency_key IS NOT NULL;
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create idempotency check index: %w", err)
	}

	//AUDIT
	// Timeline for a specific session (most common admin query).
	err = DB.Exec(`
	CREATE INDEX IF NOT EXISTS idx_audit_events_session_timeline
	ON audit_events (session_id, created_at DESC)
	WHERE session_id IS NOT NULL;
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create specific session admin index: %w", err)
	}

	//Timeline for a specific user (compliance export)
	err = DB.Exec(`
	CREATE INDEX IF NOT EXISTS idx_audit_events_user_timeline
	ON audit_events (user_id, created_at DESC)
	WHERE user_id IS NOT NULL;
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create specific user admin index: %w", err)
	}

	// Event type filtering for dashboards.
	err = DB.Exec(`
	CREATE INDEX IF NOT EXISTS idx_audit_events_type_time ON audit_events (event_type, created_at DESC);
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create event type filtering index: %w", err)
	}

	fmt.Println("Indexes created successfully")
	return nil
}
