-- Android devices registered for FCM can no longer be reached: push goes
-- over the relay's Server-Sent Events stream (ADR 0027). The app registers
-- again with the platform "sse".
DELETE FROM notification_app_devices WHERE platform = 'fcm';
