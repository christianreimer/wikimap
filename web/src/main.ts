import { Camera, visibleTiles, type TileCoord } from './engine/camera';
import { TileLoader, createFetcher } from './engine/tile-loader';
import { MapRenderer } from './renderer/webgl';
import { LabelRenderer } from './renderer/labels';
import { PanHandler, ZoomHandler } from './interaction/controls';
import { LIGHT, DARK, type Theme } from './theme';
import type { TileData } from './types/tile';

const canvas = document.getElementById('map') as HTMLCanvasElement;
const labelCanvas = document.getElementById('labels') as HTMLCanvasElement;

const camera = new Camera();
camera.zoom = 2; // Start zoomed in where nodes are visible
const loader = new TileLoader(createFetcher());
const renderer = new MapRenderer(canvas);
const labelRenderer = new LabelRenderer(labelCanvas);
const panHandler = new PanHandler(camera);
const zoomHandler = new ZoomHandler(camera);

let renderPending = false;
let theme: Theme = LIGHT;
const loadedTiles = new Map<string, { coord: TileCoord; data: TileData }>();

function resize() {
  canvas.style.width = window.innerWidth + 'px';
  canvas.style.height = window.innerHeight + 'px';
  labelCanvas.style.width = window.innerWidth + 'px';
  labelCanvas.style.height = window.innerHeight + 'px';
  renderer.resize();
  labelRenderer.resize();
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

  renderer.render(camera, entries, theme);
  labelRenderer.render(camera, entries, theme);
}

// Wire up animated zoom for button clicks
zoomHandler.setRenderCallback(requestRender);

document.getElementById('zoom-in')!.addEventListener('click', () => {
  const cx = canvas.clientWidth / 2;
  const cy = canvas.clientHeight / 2;
  const target = Math.min(4, Math.ceil(camera.zoom + 0.5));
  zoomHandler.zoomTo(target, cx, cy, canvas.clientWidth, canvas.clientHeight);
});

document.getElementById('zoom-out')!.addEventListener('click', () => {
  const cx = canvas.clientWidth / 2;
  const cy = canvas.clientHeight / 2;
  const target = Math.max(0, Math.floor(camera.zoom - 0.5));
  zoomHandler.zoomTo(target, cx, cy, canvas.clientWidth, canvas.clientHeight);
});

// Theme toggle
document.getElementById('theme-toggle')!.addEventListener('click', () => {
  theme = theme === LIGHT ? DARK : LIGHT;
  document.body.classList.toggle('dark', theme === DARK);
  requestRender();
});

requestRender();
