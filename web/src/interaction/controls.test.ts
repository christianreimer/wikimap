import { describe, it, expect } from 'vitest';
import { PanHandler, ZoomHandler } from './controls';
import { Camera } from '../engine/camera';

describe('PanHandler', () => {
  it('updates camera on drag', () => {
    const cam = new Camera();
    const startX = cam.x;
    const handler = new PanHandler(cam);
    handler.onPointerDown(100, 100);
    handler.onPointerMove(150, 100, 800, 600);
    expect(cam.x).not.toBe(startX);
  });

  it('does nothing without drag start', () => {
    const cam = new Camera();
    const startX = cam.x;
    const handler = new PanHandler(cam);
    handler.onPointerMove(150, 100, 800, 600);
    expect(cam.x).toBe(startX);
  });
});

describe('ZoomHandler', () => {
  it('zooms in on positive wheel delta', () => {
    const cam = new Camera();
    const handler = new ZoomHandler(cam);
    handler.onWheel(-100, 400, 300, 800, 600);
    expect(cam.zoom).toBeGreaterThan(0);
  });

  it('zooms out on negative wheel delta', () => {
    const cam = new Camera();
    cam.zoom = 5;
    const handler = new ZoomHandler(cam);
    handler.onWheel(100, 400, 300, 800, 600);
    expect(cam.zoom).toBeLessThan(5);
  });
});
