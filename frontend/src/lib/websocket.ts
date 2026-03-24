import type { WSOptions, WSStatus } from '@/types/api';

type MessageHandler = (data: unknown) => void;

export class WSClient {
  private url: string;
  private options: WSOptions;
  private ws: WebSocket | null = null;
  private status: WSStatus = 'closed';
  private retryCount = 0;
  private messageHandlers: Set<MessageHandler> = new Set();
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
  private heartbeatInterval: ReturnType<typeof setInterval> | null = null;
  private messageQueue: unknown[] = [];

  constructor(url: string, options: WSOptions = {}) {
    this.url = url;
    this.options = {
      retryInterval: 1000,
      maxRetries: 10,
      retryOnClose: true,
      heartbeatInterval: 30000,
      queueMessages: true,
      maxQueueSize: 100,
      ...options,
    };
  }

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    this.setStatus('connecting');
    this.createWebSocket();
  }

  private createWebSocket(): void {
    const { protocols } = this.options;

    this.ws = protocols
      ? new WebSocket(this.url, protocols)
      : new WebSocket(this.url);

    this.setupEventListeners();
  }

  private setupEventListeners(): void {
    if (!this.ws) return;

    this.ws.onopen = (event) => {
      this.setStatus('open');
      this.retryCount = 0;
      this.flushMessageQueue();
      this.startHeartbeat();
      this.options.onOpen?.(event);
    };

    this.ws.onmessage = (event) => {
      let data: unknown = event.data;

      try {
        data = JSON.parse(event.data);
      } catch {
        // Keep raw data if not JSON
      }

      this.options.onMessage?.(event);
      this.notifyMessageHandlers(data);
    };

    this.ws.onerror = (event) => {
      this.options.onError?.(event);
    };

    this.ws.onclose = (event) => {
      this.setStatus('closed');
      this.stopHeartbeat();
      this.options.onClose?.(event);

      if (!event.wasClean && this.shouldReconnect()) {
        this.scheduleReconnect();
      }
    };
  }

  private shouldReconnect(): boolean {
    const { maxRetries, retryOnClose } = this.options;

    if (!retryOnClose) return false;
    if (maxRetries !== undefined && this.retryCount >= maxRetries) {
      return false;
    }

    return true;
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
    }

    this.retryCount++;

    const { retryInterval } = this.options;
    const delay =
      typeof retryInterval === 'number'
        ? retryInterval * 1.5 ** (this.retryCount - 1)
        : 1000;

    this.options.onReconnecting?.(this.retryCount);

    this.reconnectTimeout = setTimeout(
      () => {
        this.close(false);
        this.connect();
      },
      Math.min(delay, 30000),
    );
  }

  private setStatus(status: WSStatus): void {
    this.status = status;
  }

  private notifyMessageHandlers(data: unknown): void {
    this.messageHandlers.forEach((handler) => {
      try {
        handler(data);
      } catch (error) {
        console.error('WebSocket message handler error:', error);
      }
    });
  }

  private startHeartbeat(): void {
    const { heartbeatInterval, heartbeatMessage } = this.options;
    if (!heartbeatInterval) return;

    this.heartbeatInterval = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        const msg = heartbeatMessage ?? { type: 'ping', timestamp: Date.now() };
        this.send(msg);
      }
    }, heartbeatInterval);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  private flushMessageQueue(): void {
    while (
      this.messageQueue.length > 0 &&
      this.ws?.readyState === WebSocket.OPEN
    ) {
      const message = this.messageQueue.shift();
      if (message !== undefined) {
        this.doSend(message);
      }
    }
  }

  private doSend(data: unknown): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;

    const message = typeof data === 'string' ? data : JSON.stringify(data);
    this.ws.send(message);
  }

  send(data: unknown): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.doSend(data);
      return;
    }

    if (this.options.queueMessages) {
      const { maxQueueSize } = this.options;
      if (maxQueueSize && this.messageQueue.length >= maxQueueSize) {
        this.messageQueue.shift();
      }
      this.messageQueue.push(data);
    }
  }

  onMessage(handler: MessageHandler): () => void {
    this.messageHandlers.add(handler);
    return () => this.messageHandlers.delete(handler);
  }

  getStatus(): WSStatus {
    return this.status;
  }

  close(callOnClose = true): void {
    this.stopHeartbeat();

    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    if (this.ws) {
      this.ws.close(1000, 'Client closed connection');
      this.ws = null;
    }

    this.setStatus('closed');
    this.retryCount = 0;
    this.messageQueue = [];

    if (callOnClose) {
      this.options.onClose?.({} as CloseEvent);
    }
  }

  reconnect(): void {
    this.close(false);
    this.connect();
  }
}

export function createWS(url: string, options?: WSOptions): WSClient {
  return new WSClient(url, options);
}
