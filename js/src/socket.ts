// Socket - Low-level WebSocket wrapper
export class Socket {
  private url: string;
  private websocket: WebSocket | null = null;
  private callbacks: Map<string, ((data: any) => void)[]> = new Map();
  private heartbeatInterval: number | null = null;
  private reconnectTimer: number | null = null;
  private reconnectAttempts = 0;

  constructor(url: string) {
    this.url = url;
  }

  connect(): void {
    this.websocket = new WebSocket(this.url);

    this.websocket.onopen = () => {
      this.reconnectAttempts = 0;
      this.startHeartbeat();
      this.trigger('open', {});
    };

    this.websocket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      this.trigger('message', data);
    };

    this.websocket.onclose = () => {
      this.stopHeartbeat();
      this.trigger('close', {});
      this.scheduleReconnect();
    };

    this.websocket.onerror = (error) => {
      this.trigger('error', error);
    };
  }

  disconnect(): void {
    this.stopHeartbeat();
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }
    if (this.websocket) {
      this.websocket.close();
    }
  }

  send(data: any): void {
    if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
      this.websocket.send(JSON.stringify(data));
    }
  }

  on(event: string, callback: (data: any) => void): void {
    if (!this.callbacks.has(event)) {
      this.callbacks.set(event, []);
    }
    this.callbacks.get(event)!.push(callback);
  }

  private trigger(event: string, data: any): void {
    const callbacks = this.callbacks.get(event);
    if (callbacks) {
      callbacks.forEach(cb => cb(data));
    }
  }

  private startHeartbeat(): void {
    this.heartbeatInterval = window.setInterval(() => {
      this.send({
        topic: 'phoenix',
        event: 'heartbeat',
        payload: {},
        ref: this.makeRef()
      });
    }, 30000);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  private scheduleReconnect(): void {
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 10000);
    this.reconnectAttempts++;
    
    this.reconnectTimer = window.setTimeout(() => {
      this.connect();
    }, delay);
  }

  private makeRef(): string {
    return Math.random().toString(36).substring(2, 15);
  }
}
