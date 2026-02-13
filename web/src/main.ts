import { Camera, visibleTiles, type TileCoord } from './engine/camera';
import { TileLoader, createFetcher } from './engine/tile-loader';
import { MapRenderer } from './renderer/webgl';
import { PanHandler, ZoomHandler } from './interaction/controls';
import type { TileData } from './types/tile';

const canvas = document.getElementById('map') as HTMLCanvasElement;

const camera = new Camera();
const loader = new TileLoader(createFetcher());
const renderer = new MapRenderer(canvas);
const panHandler = new PanHandler(camera);
const zoomHandler = new ZoomHandler(camera);

function resize() {
  canvas.style.width = window.innerWidth + 'px';
  canvas.style.height = window.innerHeight + 'px';
  renderer.resize();
  requestRender();
}
window.addEventListener('resize', resize);
resize();

canvas.addEventListener('pointerdown', e => {
  panHandler.onPointerDown(e.clientX, e.clientY);
  canvas.setPointerCapture(e.pointerId);
});

canvas.addEventListener('pointermove', e => {
  panHandler.onPointerMove(e.clientX, e.clientY, canvas.clientWidth, canvas.clientHeight);
  requestRender();
});

canvas.addEventListener('pointerup', () => {
  panHandler.onPointerUp();
});

canvas.addEventListener('wheel', e => {
  e.preventDefault();
  zoomHandler.onWheel(e.deltaY, e.clientX, e.clientY, canvas.clientWidth, canvas.clientHeight);
  requestRender();
}, { passive: false });

let renderPending = false;
const loadedTiles = new Map<string, { coord: TileCoord; data: TileData }>();

function requestRender() {
  if (renderPending) return;
  renderPending = true;
  requestAnimationFrame(renderFrame);
}

function renderFrame() {
  renderPending = false;

  const tiles = visibleTiles(camera, canvas.clientWidth, canvas.clientHeight);

  for (const coord of tiles) {
    const key = `${coord.z}/${coord.x}/${coord.y}`;
    if (!loadedTiles.has(key)) {
      loader.load(coord.z, coord.x, coord.y).then(data => {
        loadedTiles.set(key, { coord, data });
        requestRender();
      });
    }
  }

  const entries: { coord: TileCoord; data: TileData }[] = [];
  for (const coord of tiles) {
    const key = `${coord.z}/${coord.x}/${coord.y}`;
    const entry = loadedTiles.get(key);
    if (entry) entries.push(entry);
  }

  renderer.render(camera, entries);
}

requestRender();
