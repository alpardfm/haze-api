CREATE TABLE admins (
	id bigserial PRIMARY KEY,
	name text NOT NULL,
	email text NOT NULL UNIQUE,
	phone text NOT NULL,
	password_hash text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE appointments (
	id bigserial PRIMARY KEY,
	client_name text NOT NULL,
	address text NOT NULL,
	notes text,
	meeting_date date NOT NULL,
	meeting_time time without time zone NOT NULL,
	duration_minutes integer NOT NULL DEFAULT 120,
	start_at timestamptz NOT NULL,
	end_at timestamptz NOT NULL,
	status text NOT NULL DEFAULT 'scheduled',
	is_reminder_enabled boolean NOT NULL DEFAULT false,
	reminder_start_at timestamptz,
	reminder_interval_hours integer,
	created_by_admin_id bigint NOT NULL REFERENCES admins(id) ON DELETE RESTRICT,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	cancelled_at timestamptz,
	CONSTRAINT appointments_duration_v1_check CHECK (duration_minutes = 120),
	CONSTRAINT appointments_time_range_check CHECK (end_at > start_at),
	CONSTRAINT appointments_status_check CHECK (status IN ('scheduled', 'on_going', 'done', 'cancelled')),
	CONSTRAINT appointments_reminder_interval_check CHECK (reminder_interval_hours IS NULL OR reminder_interval_hours > 0),
	CONSTRAINT appointments_reminder_config_check CHECK (
		(is_reminder_enabled = false)
		OR (reminder_start_at IS NOT NULL AND reminder_interval_hours IS NOT NULL)
	),
	CONSTRAINT appointments_cancelled_at_check CHECK (
		(status = 'cancelled' AND cancelled_at IS NOT NULL)
		OR (status <> 'cancelled')
	)
);

CREATE INDEX idx_appointments_meeting_date_start_at ON appointments (meeting_date, start_at);
CREATE INDEX idx_appointments_status_start_at ON appointments (status, start_at);
CREATE INDEX idx_appointments_created_by_admin_id ON appointments (created_by_admin_id);
CREATE INDEX idx_appointments_active_time_range ON appointments (start_at, end_at)
WHERE status IN ('scheduled', 'on_going');

CREATE TABLE notification_logs (
	id bigserial PRIMARY KEY,
	appointment_id bigint NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
	notification_type text NOT NULL,
	scheduled_for timestamptz NOT NULL,
	sent_at timestamptz,
	recipient text NOT NULL,
	status text NOT NULL,
	message text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_logs_appointment_id ON notification_logs (appointment_id);
CREATE INDEX idx_notification_logs_scheduled_for ON notification_logs (scheduled_for);
CREATE INDEX idx_notification_logs_appointment_type_schedule ON notification_logs (
	appointment_id,
	notification_type,
	scheduled_for
);

