import { Camera } from '../engine/camera';

export class PanHandler {
  private dragging = false;
  private lastX = 0;
  private lastY = 0;

  constructor(private camera: Camera) {}

  onPointerDown(x: number, y: number): void {
    this.dragging = true;
    this.lastX = x;
    this.lastY = y;
  }

  onPointerMove(x: number, y: number, screenW: number, screenH: number): void {
    if (!this.dragging) return;
    const dx = x - this.lastX;
    const dy = y - this.lastY;
    this.camera.pan(dx, dy, screenW, screenH);
    this.lastX = x;
    this.lastY = y;
  }

  onPointerUp(): void {
    this.dragging = false;
  }
}

export class ZoomHandler {
  private targetZoom: number | null = null;
  private animating = false;
  private lastScreenX = 0;
  private lastScreenY = 0;
  private lastScreenW = 0;
  private lastScreenH = 0;
  private onRender: (() => void) | null = null;

  constructor(private camera: Camera) {}

  setRenderCallback(cb: () => void): void {
    this.onRender = cb;
  }

  onWheel(deltaY: number, screenX: number, screenY: number, screenW: number, screenH: number): void {
    // Use proportional delta for smooth trackpad zoom, clamped for mouse wheel
    const rawDelta = -deltaY * 0.002;
    const zoomDelta = Math.max(-0.15, Math.min(0.15, rawDelta));
    this.camera.zoomBy(zoomDelta, screenX, screenY, screenW, screenH);
  }

  zoomTo(target: number, screenX: number, screenY: number, screenW: number, screenH: number): void {
    this.targetZoom = target;
    this.lastScreenX = screenX;
    this.lastScreenY = screenY;
    this.lastScreenW = screenW;
    this.lastScreenH = screenH;
    if (!this.animating) this.animate();
  }

  private animate = (): void => {
    if (this.targetZoom === null) {
      this.animating = false;
      return;
    }
    this.animating = true;
    const diff = this.targetZoom - this.camera.zoom;
    if (Math.abs(diff) < 0.005) {
      this.camera.zoomBy(diff, this.lastScreenX, this.lastScreenY, this.lastScreenW, this.lastScreenH);
      this.targetZoom = null;
      this.animating = false;
      this.onRender?.();
      return;
    }
    const step = diff * 0.15;
    this.camera.zoomBy(step, this.lastScreenX, this.lastScreenY, this.lastScreenW, this.lastScreenH);
    this.onRender?.();
    requestAnimationFrame(this.animate);
  };
}
