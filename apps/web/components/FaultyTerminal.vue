<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from "vue";

// Procedural terminal glyphs inspired by Vue Bits' Faulty Terminal.
const canvas = ref<HTMLCanvasElement | null>(null);
let dispose = () => {};

onMounted(() => {
  const surface = canvas.value;
  const gl = surface?.getContext("webgl", { alpha: false, antialias: false });
  if (!surface || !gl) return;
  const shaders: WebGLShader[] = [];
  const compile = (type: number, source: string) => {
    const shader = gl.createShader(type)!;
    shaders.push(shader);
    gl.shaderSource(shader, source);
    gl.compileShader(shader);
    return shader;
  };
  const program = gl.createProgram()!;
  gl.attachShader(program, compile(gl.VERTEX_SHADER, `
    attribute vec2 position;
    varying vec2 uv;
    void main() { uv = position * .5 + .5; gl_Position = vec4(position, 0., 1.); }
  `));
  gl.attachShader(program, compile(gl.FRAGMENT_SHADER, `
    precision mediump float;
    varying vec2 uv;
    uniform float time;
    uniform vec2 resolution;
    float hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }
    float noise(vec2 p) {
      return sin(p.x * 10.) * sin(p.y * (3. + sin(time * .09))) + .2;
    }
    float fbm(vec2 p) {
      float f = 0.; float a = .5;
      for (int i = 0; i < 3; i++) { f += a * noise(p); p = p * 2.2 + time * .02; a *= .4545; }
      return f;
    }
    void main() {
      vec2 p = uv;
      float glitch = step(.985, sin(time * 1.7));
      p.x += glitch * sin(p.y * 80. + time * 20.) * .025;
      vec2 grid = vec2(55. * resolution.x / resolution.y, 55.);
      vec2 cell = floor(p * grid);
      vec2 glyph = fract(p * grid) * 1.45;
      vec2 q = vec2(fbm(cell * .04), fbm(cell * .04 + 1.));
      float intensity = fbm(cell * .04 + q + time * .035) * 1.3;
      vec2 dotCell = floor(vec2(glyph.x, 1. - glyph.y) * 5.) - 2.;
      float lit = step(.1, intensity - dot(dotCell, dotCell) * .0625);
      lit *= step(glyph.x, 1.) * step(glyph.y, 1.);
      lit *= .3 + .7 * fract(glyph.y * 5.);
      float scan = .82 + .18 * sin(gl_FragCoord.y * 1.6);
      float pulse = .85 + .15 * sin(time * .7);
      vec3 ink = vec3(.27, .49, 1.) * lit * scan * pulse;
      ink += vec3(.025, .035, .075);
      gl_FragColor = vec4(ink, 1.);
    }
  `));
  gl.linkProgram(program);
  if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
    shaders.forEach(shader => gl.deleteShader(shader));
    gl.deleteProgram(program);
    return;
  }
  gl.useProgram(program);
  const buffer = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
  gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 3, -1, -1, 3]), gl.STATIC_DRAW);
  const position = gl.getAttribLocation(program, "position");
  gl.enableVertexAttribArray(position);
  gl.vertexAttribPointer(position, 2, gl.FLOAT, false, 0, 0);
  const time = gl.getUniformLocation(program, "time");
  const resolution = gl.getUniformLocation(program, "resolution");
  const resize = () => {
    const dpr = Math.min(window.devicePixelRatio || 1, 1.5);
    surface.width = Math.max(1, Math.round(surface.clientWidth * dpr));
    surface.height = Math.max(1, Math.round(surface.clientHeight * dpr));
    gl.viewport(0, 0, surface.width, surface.height);
    gl.uniform2f(resolution, surface.width, surface.height);
  };
  const observer = new ResizeObserver(resize);
  observer.observe(surface);
  resize();
  let frame = 0;
  const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  const render = (timestamp: number) => {
    gl.uniform1f(time, reducedMotion ? 2 : timestamp * .0003);
    gl.drawArrays(gl.TRIANGLES, 0, 3);
    if (!reducedMotion) frame = requestAnimationFrame(render);
  };
  render(0);
  dispose = () => {
    cancelAnimationFrame(frame);
    observer.disconnect();
    gl.deleteBuffer(buffer);
    shaders.forEach(shader => gl.deleteShader(shader));
    gl.deleteProgram(program);
  };
});

onBeforeUnmount(() => dispose());
</script>

<template>
  <canvas ref="canvas" class="faulty-terminal" aria-hidden="true" />
</template>

<style scoped>
.faulty-terminal { position: absolute; inset: 0; width: 100%; height: 100%; background: #080d20; }
</style>
