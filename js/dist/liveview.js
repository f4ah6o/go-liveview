"use strict";
var __create = Object.create;
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __getProtoOf = Object.getPrototypeOf;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toESM = (mod, isNodeMode, target) => (target = mod != null ? __create(__getProtoOf(mod)) : {}, __copyProps(
  // If the importer is in node compatibility mode or this is not an ESM
  // file that has been converted to a CommonJS file using a Babel-
  // compatible transform (i.e. "__esModule" has not been set), then set
  // "default" to the CommonJS "module.exports" for node compatibility.
  isNodeMode || !mod || !mod.__esModule ? __defProp(target, "default", { value: mod, enumerable: true }) : target,
  mod
));
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// src/liveview.ts
var liveview_exports = {};
__export(liveview_exports, {
  Channel: () => Channel,
  LiveSocket: () => LiveSocket
});
module.exports = __toCommonJS(liveview_exports);

// src/renderer.ts
var import_morphdom = __toESM(require("morphdom"));
var Renderer = class {
  constructor(container) {
    this.static = null;
    this.container = container;
  }
  // Apply a patch to the DOM
  apply(patch) {
    if (patch.s) {
      this.static = patch.s;
    }
    if (!this.static) {
      throw new Error("No static template received");
    }
    const html = this.build(this.static, patch.d);
    (0, import_morphdom.default)(this.container, `<div>${html}</div>`, {
      childrenOnly: true,
      onBeforeElUpdated: (fromEl, toEl) => {
        if (fromEl === document.activeElement) {
          return false;
        }
        return true;
      }
    });
  }
  // Build HTML string from static and dynamic parts
  build(staticParts, dynamicParts) {
    let result = "";
    for (let i = 0; i < staticParts.length; i++) {
      result += staticParts[i];
      if (i < dynamicParts.length) {
        result += this.renderDynamic(dynamicParts[i]);
      }
    }
    return result;
  }
  // Render a dynamic value to string
  renderDynamic(value) {
    if (value === null || value === void 0) {
      return "";
    }
    if (typeof value === "string") {
      return this.escapeHtml(value);
    }
    if (typeof value === "number" || typeof value === "boolean") {
      return String(value);
    }
    if (Array.isArray(value)) {
      return value.map((v) => this.renderDynamic(v)).join("");
    }
    if (value.s && value.d) {
      return this.build(value.s, value.d);
    }
    return "";
  }
  // Escape HTML entities
  escapeHtml(text) {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
  }
};
var LiveViewRenderer = class {
  constructor(containerId) {
    const container = document.getElementById(containerId);
    if (!container) {
      throw new Error(`Container #${containerId} not found`);
    }
    this.container = container;
    this.renderer = new Renderer(container);
  }
  render(patch) {
    this.renderer.apply(patch);
  }
  // Set up event delegation for LiveView events
  setupEventDelegation(pushEvent) {
    this.container.addEventListener("click", (e) => {
      const target = e.target;
      const phxClick = target.closest("[phx-click]");
      if (phxClick) {
        e.preventDefault();
        const event = phxClick.getAttribute("phx-click");
        const value = phxClick.getAttribute("phx-value") || {};
        pushEvent(event, { value });
      }
    });
    this.container.addEventListener("input", (e) => {
      const target = e.target;
      const phxChange = target.closest("[phx-change]");
      if (phxChange) {
        const event = phxChange.getAttribute("phx-change");
        pushEvent(event, { value: target.value });
      }
    });
    this.container.addEventListener("submit", (e) => {
      const target = e.target;
      const phxSubmit = target.closest("[phx-submit]");
      if (phxSubmit) {
        e.preventDefault();
        const event = phxSubmit.getAttribute("phx-submit");
        const formData = new FormData(target);
        const data = {};
        formData.forEach((value, key) => {
          data[key] = value.toString();
        });
        pushEvent(event, data);
      }
    });
  }
};

// src/liveview.ts
var LiveSocket = class {
  constructor(url, opts = {}) {
    this.socket = null;
    this.channels = /* @__PURE__ */ new Map();
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 10;
    this.reconnectDelay = 1e3;
    this.url = url;
    this.params = opts.params || {};
  }
  connect() {
    const wsUrl = this.url.startsWith("ws") ? this.url : `ws://${window.location.host}${this.url}`;
    this.socket = new WebSocket(wsUrl);
    this.socket.onopen = () => {
      console.log("LiveSocket connected");
      this.reconnectAttempts = 0;
      this.channels.forEach((channel) => {
        channel.rejoin();
      });
    };
    this.socket.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      this.handleMessage(msg);
    };
    this.socket.onclose = () => {
      console.log("LiveSocket disconnected");
      this.attemptReconnect();
    };
    this.socket.onerror = (error) => {
      console.error("LiveSocket error:", error);
    };
  }
  channel(topic, params = {}) {
    const channel = new Channel(this, topic, params);
    this.channels.set(topic, channel);
    return channel;
  }
  isConnected() {
    return this.socket !== null && this.socket.readyState === WebSocket.OPEN;
  }
  push(topic, event, payload, ref) {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      console.error("Socket not connected");
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
  handleMessage(msg) {
    const channel = this.channels.get(msg.topic);
    if (channel) {
      channel.handleMessage(msg);
    }
  }
  attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error("Max reconnection attempts reached");
      return;
    }
    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    setTimeout(() => {
      console.log(`Reconnecting... (attempt ${this.reconnectAttempts})`);
      this.connect();
    }, delay);
  }
  makeRef() {
    return Math.random().toString(36).substring(2, 15);
  }
};
var Channel = class {
  constructor(socket, topic, params) {
    this.state = "closed";
    this.bindings = /* @__PURE__ */ new Map();
    this.joinRef = null;
    this.socket = socket;
    this.topic = topic;
    this.params = params;
  }
  join() {
    if (this.state === "joined" || this.state === "joining") {
      return;
    }
    if (!this.socket.isConnected()) {
      this.state = "pending";
      return;
    }
    this.state = "joining";
    this.joinRef = this.makeRef();
    this.socket.push(this.topic, "phx_join", {
      params: this.params,
      session: "",
      static: ""
    }, this.joinRef);
  }
  rejoin() {
    this.state = "closed";
    this.join();
  }
  on(event, callback) {
    if (!this.bindings.has(event)) {
      this.bindings.set(event, []);
    }
    this.bindings.get(event).push(callback);
  }
  push(event, payload) {
    this.socket.push(this.topic, "event", {
      type: "click",
      event,
      value: payload
    });
  }
  handleMessage(msg) {
    if (msg.event === "phx_reply" && msg.ref === this.joinRef) {
      const payload2 = typeof msg.payload === "string" ? JSON.parse(msg.payload) : msg.payload;
      if (payload2.status === "ok") {
        this.state = "joined";
        this.trigger("join", payload2.response);
      } else {
        this.state = "errored";
        this.trigger("error", payload2);
      }
      return;
    }
    if (msg.event === "diff") {
      const payload2 = typeof msg.payload === "string" ? JSON.parse(msg.payload) : msg.payload;
      this.trigger("diff", payload2);
      return;
    }
    const payload = typeof msg.payload === "string" ? JSON.parse(msg.payload) : msg.payload;
    this.trigger(msg.event, payload);
  }
  trigger(event, payload) {
    const handlers = this.bindings.get(event);
    if (handlers) {
      handlers.forEach((handler) => handler(payload));
    }
  }
  makeRef() {
    return Math.random().toString(36).substring(2, 15);
  }
};
window.LiveSocket = LiveSocket;
window.LiveViewRenderer = LiveViewRenderer;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  Channel,
  LiveSocket
});
