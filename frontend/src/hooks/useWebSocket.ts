import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { WSClient } from '@/lib/websocket';
import type { WSMessage, WSOptions, WSStatus } from '@/types/api';

interface UseWebSocketReturn<T> {
  data: T | null;
  status: WSStatus;
  error: Event | null;
  send: (data: unknown) => void;
  reconnect: () => void;
}

export function useWebSocket<T = unknown>(
  url: string | null,
  options: WSOptions = {},
): UseWebSocketReturn<T> {
  const [data, setData] = useState<T | null>(null);
  const [status, setStatus] = useState<WSStatus>('closed');
  const [error, setError] = useState<Event | null>(null);

  const clientRef = useRef<WSClient | null>(null);
  const optionsRef = useRef(options);
  optionsRef.current = options;

  const stableOptions = useMemo(
    () => ({
      onOpen: optionsRef.current.onOpen,
      onMessage: optionsRef.current.onMessage,
      onError: optionsRef.current.onError,
      onClose: optionsRef.current.onClose,
      onReconnecting: optionsRef.current.onReconnecting,
      protocols: optionsRef.current.protocols,
      retryInterval: optionsRef.current.retryInterval,
      maxRetries: optionsRef.current.maxRetries,
      retryOnClose: optionsRef.current.retryOnClose,
      heartbeatInterval: optionsRef.current.heartbeatInterval,
      heartbeatMessage: optionsRef.current.heartbeatMessage,
      queueMessages: optionsRef.current.queueMessages,
      maxQueueSize: optionsRef.current.maxQueueSize,
    }),
    [],
  );

  useEffect(() => {
    if (!url) {
      setData(null);
      setStatus('closed');
      setError(null);
      return;
    }

    const client = new WSClient(url, {
      ...stableOptions,
      onOpen: (event) => {
        setStatus('open');
        setError(null);
        stableOptions.onOpen?.(event);
      },
      onMessage: (event) => {
        try {
          const parsed = JSON.parse(event.data) as T;
          setData(parsed);
        } catch {
          setData(event.data as unknown as T);
        }
        stableOptions.onMessage?.(event);
      },
      onError: (event) => {
        setError(event);
        stableOptions.onError?.(event);
      },
      onClose: (event) => {
        setStatus('closed');
        stableOptions.onClose?.(event);
      },
      onReconnecting: (attempt) => {
        setStatus('connecting');
        stableOptions.onReconnecting?.(attempt);
      },
    });

    clientRef.current = client;
    client.connect();

    return () => {
      client.close();
      clientRef.current = null;
    };
  }, [url, stableOptions]);

  const send = useCallback((data: unknown) => {
    clientRef.current?.send(data);
  }, []);

  const reconnect = useCallback(() => {
    clientRef.current?.reconnect();
  }, []);

  return { data, status, error, send, reconnect };
}

export function useWebSocketMessage<T = unknown>(
  url: string | null,
  options: WSOptions = {},
): UseWebSocketReturn<WSMessage<T>> {
  return useWebSocket<WSMessage<T>>(url, options);
}
