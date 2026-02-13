export const NODE_VERTEX = `
  attribute vec2 a_position;
  attribute float a_size;
  attribute vec3 a_color;
  attribute float a_opacity;

  uniform mat3 u_projection;

  varying vec3 v_color;
  varying float v_opacity;

  void main() {
    vec3 pos = u_projection * vec3(a_position, 1.0);
    gl_Position = vec4(pos.xy, 0.0, 1.0);
    gl_PointSize = a_size;
    v_color = a_color;
    v_opacity = a_opacity;
  }
`;

export const NODE_FRAGMENT = `
  precision mediump float;
  varying vec3 v_color;
  varying float v_opacity;

  void main() {
    vec2 coord = gl_PointCoord - vec2(0.5);
    float dist = length(coord);
    if (dist > 0.5) discard;

    float edge = smoothstep(0.45, 0.5, dist);
    gl_FragColor = vec4(v_color, v_opacity * (1.0 - edge));
  }
`;

export const EDGE_VERTEX = `
  attribute vec2 a_position;
  attribute float a_opacity;

  uniform mat3 u_projection;

  varying float v_opacity;

  void main() {
    vec3 pos = u_projection * vec3(a_position, 1.0);
    gl_Position = vec4(pos.xy, 0.0, 1.0);
    v_opacity = a_opacity;
  }
`;

export const EDGE_FRAGMENT = `
  precision mediump float;
  uniform vec3 u_color;
  uniform float u_baseOpacity;
  varying float v_opacity;

  void main() {
    gl_FragColor = vec4(u_color, v_opacity * u_baseOpacity);
  }
`;
