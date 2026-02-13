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
  constructor(private camera: Camera) {}

  onWheel(deltaY: number, screenX: number, screenY: number, screenW: number, screenH: number): void {
    const zoomDelta = deltaY > 0 ? -0.5 : 0.5;
    this.camera.zoomBy(zoomDelta, screenX, screenY, screenW, screenH);
  }
}
