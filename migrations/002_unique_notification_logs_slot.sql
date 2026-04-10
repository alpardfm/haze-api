CREATE UNIQUE INDEX idx_notification_logs_unique_slot ON notification_logs (
	appointment_id,
	notification_type,
	scheduled_for
);

