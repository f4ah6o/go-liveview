import { LiveViewRenderer } from './renderer';

// LiveSocket - Main entry point for the LiveView client
export class LiveSocket {
  private url: string;
  private params: Record<string, string>;
  private socket: WebSocket | null = null;
  private channels: Map<string, Channel> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000;

  constructor(url: string, opts: { params?: Record<string, string> } = {}) {
    this.url = url;
    this.params = opts.params || {};
  }

  connect(): void {
    const wsUrl = this.url.startsWith('ws') ? this.url : `ws://${window.location.host}${this.url}`;
    this.socket = new WebSocket(wsUrl);
    
    this.socket.onopen = () => {
      console.log('LiveSocket connected');
      this.reconnectAttempts = 0;
      
      // Join all pending channels
      this.channels.forEach(channel => {
        channel.rejoin();
      });
    };

    this.socket.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      this.handleMessage(msg);
    };

    this.socket.onclose = () => {
      console.log('LiveSocket disconnected');
      this.attemptReconnect();
    };

    this.socket.onerror = (error) => {
      console.error('LiveSocket error:', error);
    };
  }

  channel(topic: string, params: Record<string, any> = {}): Channel {
    const channel = new Channel(this, topic, params);
    this.channels.set(topic, channel);
    return channel;
  }

  isConnected(): boolean {
    return this.socket !== null && this.socket.readyState === WebSocket.OPEN;
  }

  push(topic: string, event: string, payload: any, ref?: string): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      console.error('Socket not connected');
      return;
    }

    const msg = {
      topic,
      event,
      payload,
      ref: ref || this.makeRef()
    };

    this.socket.send(JSON.stringify(msg));
  }

  private handleMessage(msg: any): void {
    const channel = this.channels.get(msg.topic);
    if (channel) {
      channel.handleMessage(msg);
    }
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    
    setTimeout(() => {
      console.log(`Reconnecting... (attempt ${this.reconnectAttempts})`);
      this.connect();
    }, delay);
  }

  private makeRef(): string {
    return Math.random().toString(36).substring(2, 15);
  }
}

// Channel - Represents a LiveView channel
type ChannelState = 'closed' | 'errored' | 'joined' | 'joining' | 'pending';

export class Channel {
  private socket: LiveSocket;
  topic: string;
  private params: Record<string, any>;
  private state: ChannelState = 'closed';
  private bindings: Map<string, ((payload: any) => void)[]> = new Map();
  private joinRef: string | null = null;

  constructor(socket: LiveSocket, topic: string, params: Record<string, any>) {
    this.socket = socket;
    this.topic = topic;
    this.params = params;
  }

  join(): void {
    if (this.state === 'joined' || this.state === 'joining') {
      return;
    }

    // If socket is not connected yet, mark as pending and wait for rejoin on connect
    if (!this.socket.isConnected()) {
      this.state = 'pending';
      return;
    }

    this.state = 'joining';
    this.joinRef = this.makeRef();
    
    this.socket.push(this.topic, 'phx_join', {
      params: this.params,
      session: '',
      static: ''
    }, this.joinRef);
  }

  rejoin(): void {
    this.state = 'closed';
    this.join();
  }

  on(event: string, callback: (payload: any) => void): void {
    if (!this.bindings.has(event)) {
      this.bindings.set(event, []);
    }
    this.bindings.get(event)!.push(callback);
  }

  push(event: string, payload: any): void {
    this.socket.push(this.topic, 'event', {
      type: 'click',
      event: event,
      value: payload
    });
  }

  handleMessage(msg: any): void {
    // Handle join reply
    if (msg.event === 'phx_reply' && msg.ref === this.joinRef) {
      const payload = typeof msg.payload === 'string' ? JSON.parse(msg.payload) : msg.payload;
      if (payload.status === 'ok') {
        this.state = 'joined';
        this.trigger('join', payload.response);
      } else {
        this.state = 'errored';
        this.trigger('error', payload);
      }
      return;
    }

    // Handle diff
    if (msg.event === 'diff') {
      const payload = typeof msg.payload === 'string' ? JSON.parse(msg.payload) : msg.payload;
      this.trigger('diff', payload);
      return;
    }

    // Handle other events
    const payload = typeof msg.payload === 'string' ? JSON.parse(msg.payload) : msg.payload;
    this.trigger(msg.event, payload);
  }

  private trigger(event: string, payload: any): void {
    const handlers = this.bindings.get(event);
    if (handlers) {
      handlers.forEach(handler => handler(payload));
    }
  }

  private makeRef(): string {
    return Math.random().toString(36).substring(2, 15);
  }
}

// Export for global access
declare global {
  interface Window {
    LiveSocket: typeof LiveSocket;
    LiveViewRenderer: typeof LiveViewRenderer;
  }
}

window.LiveSocket = LiveSocket;
window.LiveViewRenderer = LiveViewRenderer;
