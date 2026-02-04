import morphdom from 'morphdom';

// Patch - Represents a DOM diff patch
export interface Patch {
  s?: string[];           // Static parts
  d: (string | Patch | Patch[])[];  // Dynamic parts
}

// Renderer - Applies patches to the DOM
export class Renderer {
  private container: HTMLElement;
  private static: string[] | null = null;

  constructor(container: HTMLElement) {
    this.container = container;
  }

  // Apply a patch to the DOM
  apply(patch: Patch): void {
    // Update static template if provided
    if (patch.s) {
      this.static = patch.s;
    }

    if (!this.static) {
      throw new Error('No static template received');
    }

    // Build HTML from static and dynamic parts
    const html = this.build(this.static, patch.d);

    // Apply using morphdom for efficient DOM updates
    morphdom(this.container, `<div>${html}</div>`, {
      childrenOnly: true,
      onBeforeElUpdated: (fromEl, toEl) => {
        // Preserve focus and scroll position
        if (fromEl === document.activeElement) {
          return false;
        }
        return true;
      }
    });
  }

  // Build HTML string from static and dynamic parts
  private build(staticParts: string[], dynamicParts: any[]): string {
    let result = '';
    
    for (let i = 0; i < staticParts.length; i++) {
      result += staticParts[i];
      
      if (i < dynamicParts.length) {
        result += this.renderDynamic(dynamicParts[i]);
      }
    }
    
    return result;
  }

  // Render a dynamic value to string
  private renderDynamic(value: any): string {
    if (value === null || value === undefined) {
      return '';
    }

    if (typeof value === 'string') {
      return this.escapeHtml(value);
    }

    if (typeof value === 'number' || typeof value === 'boolean') {
      return String(value);
    }

    if (Array.isArray(value)) {
      return value.map(v => this.renderDynamic(v)).join('');
    }

    if (value.s && value.d) {
      // Nested patch
      return this.build(value.s, value.d);
    }

    return '';
  }

  // Escape HTML entities
  private escapeHtml(text: string): string {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }
}

// LiveViewRenderer - High-level LiveView rendering
export class LiveViewRenderer {
  private container: HTMLElement;
  private renderer: Renderer;

  constructor(containerId: string) {
    const container = document.getElementById(containerId);
    if (!container) {
      throw new Error(`Container #${containerId} not found`);
    }
    this.container = container;
    this.renderer = new Renderer(container);
  }

  render(patch: Patch): void {
    this.renderer.apply(patch);
  }

  // Set up event delegation for LiveView events
  setupEventDelegation(pushEvent: (event: string, payload: any) => void): void {
    this.container.addEventListener('click', (e) => {
      const target = e.target as HTMLElement;
      const phxClick = target.closest('[phx-click]');
      
      if (phxClick) {
        e.preventDefault();
        const event = phxClick.getAttribute('phx-click');
        const value = phxClick.getAttribute('phx-value') || {};
        pushEvent(event!, { value });
      }
    });

    this.container.addEventListener('input', (e) => {
      const target = e.target as HTMLInputElement;
      const phxChange = target.closest('[phx-change]');
      
      if (phxChange) {
        const event = phxChange.getAttribute('phx-change');
        pushEvent(event!, { value: target.value });
      }
    });

    this.container.addEventListener('submit', (e) => {
      const target = e.target as HTMLFormElement;
      const phxSubmit = target.closest('[phx-submit]');
      
      if (phxSubmit) {
        e.preventDefault();
        const event = phxSubmit.getAttribute('phx-submit');
        const formData = new FormData(target);
        const data: Record<string, string> = {};
        formData.forEach((value, key) => {
          data[key] = value.toString();
        });
        pushEvent(event!, data);
      }
    });
  }
}
