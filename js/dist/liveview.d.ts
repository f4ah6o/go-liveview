interface Patch {
    s?: string[];
    d: (string | Patch | Patch[])[];
}
declare class LiveViewRenderer {
    private container;
    private renderer;
    constructor(containerId: string);
    render(patch: Patch): void;
    setupEventDelegation(pushEvent: (event: string, payload: any) => void): void;
}

declare class LiveSocket {
    private url;
    private params;
    private socket;
    private channels;
    private reconnectAttempts;
    private maxReconnectAttempts;
    private reconnectDelay;
    constructor(url: string, opts?: {
        params?: Record<string, string>;
    });
    connect(): void;
    channel(topic: string, params?: Record<string, any>): Channel;
    isConnected(): boolean;
    push(topic: string, event: string, payload: any, ref?: string): void;
    private handleMessage;
    private attemptReconnect;
    private makeRef;
}
declare class Channel {
    private socket;
    topic: string;
    private params;
    private state;
    private bindings;
    private joinRef;
    constructor(socket: LiveSocket, topic: string, params: Record<string, any>);
    join(): void;
    rejoin(): void;
    on(event: string, callback: (payload: any) => void): void;
    push(event: string, payload: any): void;
    handleMessage(msg: any): void;
    private trigger;
    private makeRef;
}
declare global {
    interface Window {
        LiveSocket: typeof LiveSocket;
        LiveViewRenderer: typeof LiveViewRenderer;
    }
}

export { Channel, LiveSocket };
