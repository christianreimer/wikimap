import { describe, it, expect, vi } from 'vitest';
import { TileLoader } from './tile-loader';
import type { TileData } from '../types/tile';

describe('TileLoader', () => {
  it('fetches and caches tiles', async () => {
    const mockData: TileData = {
      nodes: [{ id: 1, x: 0.5, y: 0.5, label: 'Test', importance: 1, clusterId: 0 }],
      edges: [],
    };

    const fetcher = vi.fn().mockResolvedValue(mockData);
    const loader = new TileLoader(fetcher);

    const tile = await loader.load(0, 0, 0);
    expect(tile).toEqual(mockData);
    expect(fetcher).toHaveBeenCalledOnce();

    const tile2 = await loader.load(0, 0, 0);
    expect(tile2).toEqual(mockData);
    expect(fetcher).toHaveBeenCalledOnce();
  });

  it('evicts tiles beyond cache limit', async () => {
    const fetcher = vi.fn().mockResolvedValue({ nodes: [], edges: [] });
    const loader = new TileLoader(fetcher, 2);

    await loader.load(0, 0, 0);
    await loader.load(1, 0, 0);
    await loader.load(1, 1, 0);

    expect(fetcher).toHaveBeenCalledTimes(3);

    await loader.load(0, 0, 0);
    expect(fetcher).toHaveBeenCalledTimes(4);
  });
});
