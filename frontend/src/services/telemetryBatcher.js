import { apiClient } from './apiClient';
import { ENDPOINTS } from '../utils/endpoints';

let eventQueue = [];
let flushTimer = null;

export function recordTelemetry(eventType, entityId, metadata = {}) {
  const event = {
    eventType,
    entityId,
    timestamp: new Date().toISOString(),
    metadata,
  };
  eventQueue.push(event);

  if (eventQueue.length >= 10) {
    flushTelemetry();
  } else if (!flushTimer) {
    flushTimer = setTimeout(flushTelemetry, 15000); // 15s auto-flush
  }
}

export async function flushTelemetry() {
  if (flushTimer) {
    clearTimeout(flushTimer);
    flushTimer = null;
  }

  if (eventQueue.length === 0) return;

  const payload = [...eventQueue];
  eventQueue = [];

  try {
    await apiClient.post(ENDPOINTS.TELEMETRY.EVENTS, { events: payload });
  } catch (err) {
    console.debug('Telemetry flush deferred (will retry next batch)');
  }
}

// Window unload flush
if (typeof window !== 'undefined') {
  window.addEventListener('beforeunload', () => {
    if (eventQueue.length > 0) {
      navigator.sendBeacon?.(
        `${import.meta.env.VITE_API_BASE_URL || '/api/v1'}${ENDPOINTS.TELEMETRY.EVENTS}`,
        JSON.stringify({ events: eventQueue })
      );
    }
  });
}
