export interface Theme {
  bg: [number, number, number];
  labelFill: string;
  labelStroke: string;
  edgeColor: [number, number, number];
  edgeOpacity: number;
}

export const LIGHT: Theme = {
  bg: [0.97, 0.97, 0.97],
  labelFill: '#333',
  labelStroke: 'rgba(255, 255, 255, 0.9)',
  edgeColor: [0.5, 0.5, 0.5],
  edgeOpacity: 0.3,
};

export const DARK: Theme = {
  bg: [0.11, 0.11, 0.13],
  labelFill: '#ddd',
  labelStroke: 'rgba(0, 0, 0, 0.7)',
  edgeColor: [0.55, 0.55, 0.55],
  edgeOpacity: 0.25,
};
