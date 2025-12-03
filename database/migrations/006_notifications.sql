-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, -- NULL for broadcast notifications
    type VARCHAR(100) NOT NULL,
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal', -- low, normal, high, critical
    status VARCHAR(20) NOT NULL DEFAULT 'unread', -- unread, read, archived
    data JSONB, -- Additional data as JSON
    action_url TEXT,
    action_text VARCHAR(100),
    read_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP
);

-- Create indexes for notifications
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_priority ON notifications(priority);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX idx_notifications_expires_at ON notifications(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_notifications_data ON notifications USING GIN (data); -- For JSON queries

-- Create notification_preferences table
CREATE TABLE IF NOT EXISTS notification_preferences (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email_enabled BOOLEAN NOT NULL DEFAULT true,
    sms_enabled BOOLEAN NOT NULL DEFAULT false,
    push_enabled BOOLEAN NOT NULL DEFAULT true,
    websocket_enabled BOOLEAN NOT NULL DEFAULT true,
    enabled_types TEXT[], -- Array of enabled notification types (empty = all types)
    min_priority VARCHAR(20) NOT NULL DEFAULT 'normal', -- Minimum priority to receive
    quiet_hours_start TIMESTAMP,
    quiet_hours_end TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

-- Create index for notification_preferences
CREATE INDEX idx_notification_preferences_user_id ON notification_preferences(user_id);

-- Create notification_delivery_logs table
CREATE TABLE IF NOT EXISTS notification_delivery_logs (
    id UUID PRIMARY KEY,
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel VARCHAR(50) NOT NULL, -- email, sms, push, websocket
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, sent, failed, delivered
    error TEXT,
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for notification_delivery_logs
CREATE INDEX idx_notification_delivery_logs_notification_id ON notification_delivery_logs(notification_id);
CREATE INDEX idx_notification_delivery_logs_user_id ON notification_delivery_logs(user_id);
CREATE INDEX idx_notification_delivery_logs_channel ON notification_delivery_logs(channel);
CREATE INDEX idx_notification_delivery_logs_status ON notification_delivery_logs(status);
CREATE INDEX idx_notification_delivery_logs_created_at ON notification_delivery_logs(created_at DESC);

-- Add comments for documentation
COMMENT ON TABLE notifications IS 'Stores user notifications for real-time delivery via WebSocket, email, SMS, or push';
COMMENT ON COLUMN notifications.user_id IS 'NULL for broadcast notifications to all users';
COMMENT ON COLUMN notifications.data IS 'Additional notification data as JSON (e.g., resource IDs, metadata)';
COMMENT ON COLUMN notifications.expires_at IS 'Notification expiration timestamp (NULL = never expires)';

COMMENT ON TABLE notification_preferences IS 'User notification delivery preferences and settings';
COMMENT ON COLUMN notification_preferences.enabled_types IS 'Array of enabled notification types (empty array = all types enabled)';
COMMENT ON COLUMN notification_preferences.quiet_hours_start IS 'Start time for quiet hours (only critical notifications during this period)';

COMMENT ON TABLE notification_delivery_logs IS 'Tracks notification delivery across different channels';
COMMENT ON COLUMN notification_delivery_logs.channel IS 'Delivery channel: email, sms, push, or websocket';
COMMENT ON COLUMN notification_delivery_logs.status IS 'Delivery status: pending, sent, failed, or delivered';
