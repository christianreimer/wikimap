export interface TileNode {
  id: number;
  x: number;
  y: number;
  label: string;
  importance: number;
  clusterId: number;
}

export interface TileEdge {
  fromId: number;
  toId: number;
  weight: number;
  toX: number;
  toY: number;
}

export interface TileData {
  nodes: TileNode[];
  edges: TileEdge[];
}
