import { describe, it, expect } from 'vitest';
import { Camera, visibleTiles } from './camera';

describe('Camera', () => {
  it('starts at default position', () => {
    const cam = new Camera();
    expect(cam.x).toBe(0.5);
    expect(cam.y).toBe(0.5);
    expect(cam.zoom).toBe(0);
  });

  it('pans by screen delta', () => {
    const cam = new Camera();
    cam.pan(100, 0, 800, 600);
    expect(cam.x).toBeGreaterThan(0.5);
  });

  it('zooms in', () => {
    const cam = new Camera();
    cam.zoomBy(1, 400, 300, 800, 600);
    expect(cam.zoom).toBe(1);
  });

  it('clamps zoom to valid range', () => {
    const cam = new Camera();
    cam.zoomBy(-5, 400, 300, 800, 600);
    expect(cam.zoom).toBe(0);
    cam.zoomBy(100, 400, 300, 800, 600);
    expect(cam.zoom).toBeLessThanOrEqual(18);
  });
});

describe('visibleTiles', () => {
  it('returns single tile at zoom 0', () => {
    const cam = new Camera();
    const result = visibleTiles(cam, 800, 600);
    expect(result).toEqual([{ z: 0, x: 0, y: 0 }]);
  });

  it('returns multiple tiles at higher zoom', () => {
    const cam = new Camera();
    cam.zoom = 2;
    const result = visibleTiles(cam, 800, 600);
    expect(result.length).toBeGreaterThan(1);
    result.forEach(t => {
      expect(t.z).toBe(2);
      expect(t.x).toBeGreaterThanOrEqual(0);
      expect(t.y).toBeGreaterThanOrEqual(0);
    });
  });
});
