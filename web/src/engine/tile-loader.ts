import type { TileData } from '../types/tile';

export type TileFetcher = (z: number, x: number, y: number) => Promise<TileData>;

export class TileLoader {
  private cache = new Map<string, TileData>();
  private pending = new Map<string, Promise<TileData>>();
  private accessOrder: string[] = [];

  constructor(
    private fetcher: TileFetcher,
    private maxCached = 256,
  ) {}

  async load(z: number, x: number, y: number): Promise<TileData> {
    const key = `${z}/${x}/${y}`;

    const cached = this.cache.get(key);
    if (cached) {
      this.accessOrder = this.accessOrder.filter(k => k !== key);
      this.accessOrder.push(key);
      return cached;
    }

    let pending = this.pending.get(key);
    if (pending) return pending;

    pending = this.fetcher(z, x, y).then(data => {
      this.pending.delete(key);
      this.cache.set(key, data);
      this.accessOrder.push(key);
      this.evict();
      return data;
    });

    this.pending.set(key, pending);
    return pending;
  }

  private evict(): void {
    while (this.cache.size > this.maxCached && this.accessOrder.length > 0) {
      const oldest = this.accessOrder.shift()!;
      this.cache.delete(oldest);
    }
  }
}

export function createFetcher(baseUrl = ''): TileFetcher {
  return async (z, x, y) => {
    const resp = await fetch(`${baseUrl}/tiles/${z}/${x}/${y}.pbf`);
    if (!resp.ok) return { nodes: [], edges: [] };
    return resp.json();
  };
}
