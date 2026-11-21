import { type LiveNotification, parseLiveNotification } from '../../domain/live';

const STREAM_URL = '/api/notifications/stream';
const LOCK = 'waylight-notifications';
const CHANNEL = 'waylight-notifications';

/**
 * Receives the live notifications of the browser channel (#583).
 *
 * Only one tab holds the stream: the tab that gets the Web Lock opens the
 * EventSource and hands every notification to the other tabs over a
 * BroadcastChannel, so many tabs do not use up the browser's connections
 * to the instance. When that tab closes, the lock passes to another one.
 * EventSource reconnects by itself and sends Last-Event-ID, so missed
 * notifications are replayed by the server.
 *
 * `onnotify` runs in every tab; `leader` is true in the tab that holds the
 * stream, which is the one that shows a system notification.
 */
export function startLiveNotifications(onnotify: (n: LiveNotification, leader: boolean) => void): () => void {
  const abort = new AbortController();
  const channel = typeof BroadcastChannel === 'undefined' ? null : new BroadcastChannel(CHANNEL);
  if (channel) channel.onmessage = (e: MessageEvent<LiveNotification>) => onnotify(e.data, false);

  const hold = () =>
    new Promise<void>((resolve) => {
      const source = new EventSource(STREAM_URL);
      source.addEventListener('notification', (e) => {
        const n = parseLiveNotification((e as MessageEvent<string>).data);
        if (!n) return;
        channel?.postMessage(n);
        onnotify(n, true);
      });
      abort.signal.addEventListener('abort', () => {
        source.close();
        resolve();
      });
    });

  if (navigator.locks) {
    navigator.locks.request(LOCK, { signal: abort.signal }, hold).catch(() => undefined);
  } else {
    void hold();
  }
  return () => {
    abort.abort();
    channel?.close();
  };
}
