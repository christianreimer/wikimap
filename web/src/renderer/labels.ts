import type { Camera } from '../engine/camera';
import type { TileCoord } from '../engine/camera';
import type { TileData } from '../types/tile';
import type { Theme } from '../theme';

export class LabelRenderer {
  private ctx: CanvasRenderingContext2D;

  constructor(private canvas: HTMLCanvasElement) {
    const ctx = canvas.getContext('2d');
    if (!ctx) throw new Error('2D context not supported');
    this.ctx = ctx;
  }

  resize(): void {
    this.canvas.width = this.canvas.clientWidth * devicePixelRatio;
    this.canvas.height = this.canvas.clientHeight * devicePixelRatio;
  }

  render(camera: Camera, tileEntries: { coord: TileCoord; data: TileData }[], theme: Theme): void {
    const ctx = this.ctx;
    const w = this.canvas.clientWidth;
    const h = this.canvas.clientHeight;
    const dpr = devicePixelRatio;

    ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
    ctx.save();
    ctx.scale(dpr, dpr);

    const scale = Math.pow(2, camera.zoom);
    const tileSize = Math.min(w, h);

    // Collect all visible nodes with screen positions
    const labels: { screenX: number; screenY: number; label: string; importance: number }[] = [];

    for (const { coord, data } of tileEntries) {
      const ts = 1 / (1 << coord.z);
      for (const node of data.nodes) {
        if (!node.label) continue;
        const wx = (coord.x + node.x) * ts;
        const wy = (coord.y + node.y) * ts;
        const sx = (wx - camera.x) * tileSize * scale + w / 2;
        const sy = (wy - camera.y) * tileSize * scale + h / 2;

        // Skip off-screen labels (with padding)
        if (sx < -100 || sx > w + 100 || sy < -20 || sy > h + 20) continue;

        labels.push({ screenX: sx, screenY: sy, label: node.label, importance: node.importance });
      }
    }

    // Sort by importance (most important first) so they get priority
    labels.sort((a, b) => b.importance - a.importance);

    // Simple overlap rejection: track occupied regions
    const occupied: { x: number; y: number; w: number; h: number }[] = [];
    const fontSize = 11;
    ctx.font = `${fontSize}px -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`;
    ctx.textAlign = 'center';
    ctx.textBaseline = 'top';

    for (const label of labels) {
      const textW = ctx.measureText(label.label).width;
      const labelX = label.screenX - textW / 2;
      const labelY = label.screenY + 6; // offset below node
      const labelH = fontSize + 2;

      // Check overlap with already placed labels
      const rect = { x: labelX - 2, y: labelY - 1, w: textW + 4, h: labelH + 2 };
      let overlaps = false;
      for (const o of occupied) {
        if (rect.x < o.x + o.w && rect.x + rect.w > o.x &&
            rect.y < o.y + o.h && rect.y + rect.h > o.y) {
          overlaps = true;
          break;
        }
      }
      if (overlaps) continue;

      occupied.push(rect);

      ctx.strokeStyle = theme.labelStroke;
      ctx.lineWidth = 3;
      ctx.lineJoin = 'round';
      ctx.strokeText(label.label, label.screenX, labelY);

      ctx.fillStyle = theme.labelFill;
      ctx.fillText(label.label, label.screenX, labelY);
    }

    ctx.restore();
  }
}
