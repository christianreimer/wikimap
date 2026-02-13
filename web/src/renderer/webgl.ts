import { NODE_VERTEX, NODE_FRAGMENT, EDGE_VERTEX, EDGE_FRAGMENT } from './shaders';
import type { Camera } from '../engine/camera';
import type { TileData } from '../types/tile';
import type { TileCoord } from '../engine/camera';

const CLUSTER_COLORS: [number, number, number][] = [
  [0.90, 0.30, 0.30], [0.30, 0.70, 0.90], [0.40, 0.80, 0.40],
  [0.95, 0.70, 0.20], [0.70, 0.40, 0.90], [0.90, 0.50, 0.70],
  [0.30, 0.80, 0.75], [0.85, 0.85, 0.30], [0.60, 0.60, 0.60],
  [1.00, 0.60, 0.40],
];

export class MapRenderer {
  private gl: WebGLRenderingContext;
  private nodeProgram: WebGLProgram;
  private edgeProgram: WebGLProgram;

  constructor(private canvas: HTMLCanvasElement) {
    const gl = canvas.getContext('webgl', { alpha: false, antialias: true });
    if (!gl) throw new Error('WebGL not supported');
    this.gl = gl;
    gl.enable(gl.BLEND);
    gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

    this.nodeProgram = createProgram(gl, NODE_VERTEX, NODE_FRAGMENT);
    this.edgeProgram = createProgram(gl, EDGE_VERTEX, EDGE_FRAGMENT);
  }

  resize(): void {
    this.canvas.width = this.canvas.clientWidth * devicePixelRatio;
    this.canvas.height = this.canvas.clientHeight * devicePixelRatio;
    this.gl.viewport(0, 0, this.canvas.width, this.canvas.height);
  }

  render(camera: Camera, tileEntries: { coord: TileCoord; data: TileData }[]): void {
    const gl = this.gl;
    gl.clearColor(0.97, 0.97, 0.97, 1.0);
    gl.clear(gl.COLOR_BUFFER_BIT);

    const proj = this.projectionMatrix(camera);

    this.renderEdges(gl, proj, tileEntries);
    this.renderNodes(gl, proj, tileEntries);
  }

  private projectionMatrix(cam: Camera): number[] {
    const scale = Math.pow(2, cam.zoom);
    const tileSize = Math.min(this.canvas.width, this.canvas.height);
    const sx = (tileSize * scale * 2) / this.canvas.width;
    const sy = (tileSize * scale * 2) / this.canvas.height;
    const tx = -cam.x * sx;
    const ty = -cam.y * sy;
    return [sx, 0, 0, 0, -sy, 0, tx, ty, 1];
  }

  private renderNodes(
    gl: WebGLRenderingContext,
    proj: number[],
    tileEntries: { coord: TileCoord; data: TileData }[],
  ): void {
    gl.useProgram(this.nodeProgram);

    const uProj = gl.getUniformLocation(this.nodeProgram, 'u_projection');
    gl.uniformMatrix3fv(uProj, false, proj);

    const positions: number[] = [];
    const sizes: number[] = [];
    const colors: number[] = [];
    const opacities: number[] = [];

    for (const { coord, data } of tileEntries) {
      const tileScale = 1 / (1 << coord.z);
      for (const node of data.nodes) {
        const wx = (coord.x + node.x) * tileScale;
        const wy = (coord.y + node.y) * tileScale;
        positions.push(wx, wy);

        const baseSize = 2 + node.importance * 12;
        sizes.push(baseSize * devicePixelRatio);

        const c = CLUSTER_COLORS[node.clusterId % CLUSTER_COLORS.length];
        colors.push(c[0], c[1], c[2]);
        opacities.push(1.0);
      }
    }

    if (positions.length === 0) return;

    this.bindAttribute(gl, this.nodeProgram, 'a_position', new Float32Array(positions), 2);
    this.bindAttribute(gl, this.nodeProgram, 'a_size', new Float32Array(sizes), 1);
    this.bindAttribute(gl, this.nodeProgram, 'a_color', new Float32Array(colors), 3);
    this.bindAttribute(gl, this.nodeProgram, 'a_opacity', new Float32Array(opacities), 1);

    gl.drawArrays(gl.POINTS, 0, positions.length / 2);
  }

  private renderEdges(
    gl: WebGLRenderingContext,
    proj: number[],
    tileEntries: { coord: TileCoord; data: TileData }[],
  ): void {
    gl.useProgram(this.edgeProgram);

    const uProj = gl.getUniformLocation(this.edgeProgram, 'u_projection');
    gl.uniformMatrix3fv(uProj, false, proj);

    const positions: number[] = [];
    const opacities: number[] = [];

    for (const { coord, data } of tileEntries) {
      const tileScale = 1 / (1 << coord.z);
      const nodePos = new Map<number, [number, number]>();
      for (const node of data.nodes) {
        nodePos.set(node.id, [
          (coord.x + node.x) * tileScale,
          (coord.y + node.y) * tileScale,
        ]);
      }

      for (const edge of data.edges) {
        const from = nodePos.get(edge.fromId);
        if (!from) continue;
        const toX = (coord.x + edge.toX) * tileScale;
        const toY = (coord.y + edge.toY) * tileScale;

        positions.push(from[0], from[1], toX, toY);
        opacities.push(edge.weight, edge.weight);
      }
    }

    if (positions.length === 0) return;

    this.bindAttribute(gl, this.edgeProgram, 'a_position', new Float32Array(positions), 2);
    this.bindAttribute(gl, this.edgeProgram, 'a_opacity', new Float32Array(opacities), 1);

    gl.drawArrays(gl.LINES, 0, positions.length / 2);
  }

  private bindAttribute(
    gl: WebGLRenderingContext,
    program: WebGLProgram,
    name: string,
    data: Float32Array,
    size: number,
  ): void {
    const loc = gl.getAttribLocation(program, name);
    if (loc < 0) return;
    const buf = gl.createBuffer()!;
    gl.bindBuffer(gl.ARRAY_BUFFER, buf);
    gl.bufferData(gl.ARRAY_BUFFER, data, gl.DYNAMIC_DRAW);
    gl.enableVertexAttribArray(loc);
    gl.vertexAttribPointer(loc, size, gl.FLOAT, false, 0, 0);
  }
}

function createProgram(gl: WebGLRenderingContext, vSrc: string, fSrc: string): WebGLProgram {
  const vs = compileShader(gl, gl.VERTEX_SHADER, vSrc);
  const fs = compileShader(gl, gl.FRAGMENT_SHADER, fSrc);
  const program = gl.createProgram()!;
  gl.attachShader(program, vs);
  gl.attachShader(program, fs);
  gl.linkProgram(program);
  if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
    throw new Error('Program link failed: ' + gl.getProgramInfoLog(program));
  }
  return program;
}

function compileShader(gl: WebGLRenderingContext, type: number, src: string): WebGLShader {
  const shader = gl.createShader(type)!;
  gl.shaderSource(shader, src);
  gl.compileShader(shader);
  if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
    throw new Error('Shader compile failed: ' + gl.getShaderInfoLog(shader));
  }
  return shader;
}
