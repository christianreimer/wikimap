export class Camera {
  x = 0.5;
  y = 0.5;
  zoom = 0;

  private static MIN_ZOOM = 0;
  private static MAX_ZOOM = 18;

  pan(dx: number, dy: number, screenW: number, screenH: number): void {
    const scale = Math.pow(2, this.zoom);
    const tileSize = Math.min(screenW, screenH);
    this.x += dx / (tileSize * scale);
    this.y += dy / (tileSize * scale);
  }

  zoomBy(delta: number, screenX: number, screenY: number, screenW: number, screenH: number): void {
    const newZoom = Math.max(Camera.MIN_ZOOM, Math.min(Camera.MAX_ZOOM, this.zoom + delta));
    if (newZoom === this.zoom) return;

    const tileSize = Math.min(screenW, screenH);
    const scale = Math.pow(2, this.zoom);
    const worldX = this.x + (screenX - screenW / 2) / (tileSize * scale);
    const worldY = this.y + (screenY - screenH / 2) / (tileSize * scale);

    const newScale = Math.pow(2, newZoom);
    this.x = worldX - (screenX - screenW / 2) / (tileSize * newScale);
    this.y = worldY - (screenY - screenH / 2) / (tileSize * newScale);
    this.zoom = newZoom;
  }
}

export interface TileCoord {
  z: number;
  x: number;
  y: number;
}

export function visibleTiles(cam: Camera, screenW: number, screenH: number): TileCoord[] {
  const z = Math.floor(cam.zoom);
  const size = 1 << z;
  const tileSize = Math.min(screenW, screenH);
  const scale = Math.pow(2, cam.zoom);
  const worldPerPixel = 1 / (tileSize * scale);

  const halfW = (screenW / 2) * worldPerPixel;
  const halfH = (screenH / 2) * worldPerPixel;

  const minX = Math.max(0, Math.floor((cam.x - halfW) * size));
  const maxX = Math.min(size - 1, Math.floor((cam.x + halfW) * size));
  const minY = Math.max(0, Math.floor((cam.y - halfH) * size));
  const maxY = Math.min(size - 1, Math.floor((cam.y + halfH) * size));

  const tiles: TileCoord[] = [];
  for (let x = minX; x <= maxX; x++) {
    for (let y = minY; y <= maxY; y++) {
      tiles.push({ z, x, y });
    }
  }
  return tiles;
}
