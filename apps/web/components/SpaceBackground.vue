<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();
const canvasRef = ref<HTMLCanvasElement | null>(null);

let width = 0;
let height = 0;
let time = 0;
let scrollY = 0;
let animationFrameId = 0;

const bayer = [
  [0, 8, 2, 10],
  [12, 4, 14, 6],
  [3, 11, 1, 9],
  [15, 7, 13, 5]
];

const colorBlack = "#121212";
const colorWhite = "#F8F8F8";
const colorGreen = "#42ffd1";

function hexToRgba(hex: string, alpha = 1) {
  const r = Number.parseInt(hex.slice(1, 3), 16);
  const g = Number.parseInt(hex.slice(3, 5), 16);
  const b = Number.parseInt(hex.slice(5, 7), 16);

  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

function resizeCanvas() {
  const canvas = canvasRef.value;
  if (!canvas) return;

  width = window.innerWidth;
  height = window.innerHeight;
  canvas.width = width;
  canvas.height = height;
}

function handleMouseMove(event: MouseEvent) {
  const canvas = canvasRef.value;
  if (!canvas) return;

  const moveX = (event.clientX - width / 2) * 0.01;
  const moveY = (event.clientY - height / 2) * 0.01;
  canvas.style.transform = `translate(${moveX}px, ${moveY}px)`;
}

function drawScene() {
  const canvas = canvasRef.value;
  const ctx = canvas?.getContext("2d", { alpha: false });
  if (!canvas || !ctx) return;

  time += 0.01;
  // 直接在这里获取 scrollY，避免路由切换时事件未触发导致的不同步
  scrollY = window.scrollY;

  const parallaxBgY = scrollY * 0.2;
  const parallaxPlanetY = scrollY * 0.5;
  const isGlitch = Math.random() > 0.98;
  const glitchOffset = isGlitch ? (Math.random() - 0.5) * 10 : 0;

  ctx.fillStyle = colorBlack;
  ctx.fillRect(0, 0, width, height);

  ctx.fillStyle = hexToRgba(colorWhite, 0.25);
  for (let i = 0; i < 800; i += 1) {
    const seed = Math.sin(i * 123.456) * 10000;
    const x = width * 0.7 + Math.sin(seed) * width * 0.8;
    let y = height * 0.3 + Math.cos(seed) * height * 0.6 - parallaxBgY;

    if (y < -100) y = height + 100;
    if (y > height + 100) y = -100;

    const dist = Math.sqrt((x - width * 0.7) ** 2 + (y - height * 0.3) ** 2);
    if (Math.abs(Math.sin(seed * 2)) > dist / (width * 0.5)) {
      ctx.fillRect(Math.floor(x / 6) * 6 + (isGlitch ? glitchOffset : 0), Math.floor(y / 6) * 6, 4, 4);
    }
  }

  for (let i = 0; i < 150; i += 1) {
    const seed = Math.cos(i * 321.654) * 10000;
    const x = Math.abs(Math.sin(seed)) * width;
    let y = Math.abs(Math.cos(seed)) * height - parallaxBgY;

    if (y < 0) y += height;
    if (y > height) y -= height;

    const twinkle = (Math.sin(time * 5 + i) + 1) / 2;
    ctx.fillStyle = hexToRgba(colorWhite, twinkle * 0.8);
    ctx.fillRect(Math.floor(x / 3) * 3, Math.floor(y / 3) * 3, 3, 3);
  }

  const pixelSize = 6;
  const cx = width * 0.2 + glitchOffset;
  const cy = height * 0.9 - parallaxPlanetY;
  const radius = Math.min(width, height) * 0.55;

  // 仅在首页渲染 3D 星球
  if (route.path === '/') {
    const lightDir = { x: -0.5, y: -0.8, z: 0.6 };
    const len = Math.sqrt(lightDir.x ** 2 + lightDir.y ** 2 + lightDir.z ** 2);
    lightDir.x /= len;
    lightDir.y /= len;
    lightDir.z /= len;

    for (let y = cy - radius; y <= cy + radius; y += pixelSize) {
      if (y < 0 || y > height) continue;

      for (let x = cx - radius; x <= cx + radius; x += pixelSize) {
        const dx = x - cx;
        const dy = y - cy;
        const dz2 = radius * radius - dx * dx - dy * dy;

        if (dz2 >= 0) {
          const dz = Math.sqrt(dz2);
          const nx = dx / radius;
          const ny = dy / radius;
          const nz = dz / radius;

          let intensity = Math.max(0, nx * lightDir.x + ny * lightDir.y + nz * lightDir.z);
          const noise =
            Math.sin(nx * 8 + time) * Math.cos(ny * 8 - time * 0.5) * 0.15 +
            Math.sin(nx * 4 + ny * 4 + time * 2) * 0.1;
          intensity += noise;

          const rimLight = (1 - nz) ** 3 * 0.3;
          intensity += rimLight;

          const matrixX = Math.abs(Math.floor(x / pixelSize)) % 4;
          const matrixY = Math.abs(Math.floor(y / pixelSize)) % 4;
          const threshold = (bayer[matrixY][matrixX] + 0.5) / 16;

          // 使用嵌套三角函数 (Domain Warping) 制造“卷曲、流体”的曲线纹理
          const warpX = Math.sin(ny * 6 + time * 0.8);
          const warpY = Math.cos(nx * 6 - time * 0.6);
          const curveNoise1 = Math.sin(nx * 5 + warpX * 2.5 + time * 0.5);
          const curveNoise2 = Math.cos(ny * 5 + warpY * 2.5 - time * 0.7);
          const organicPattern = (curveNoise1 + curveNoise2) * 0.5;

          // 曲线扩散波：计算点到一个“动态移动中心”的距离，产生弧形涟漪
          const cxPulse = Math.sin(time * 0.4) * 0.5;
          const cyPulse = Math.cos(time * 0.5) * 0.5;
          const distToCenter = Math.sqrt((nx - cxPulse)**2 + (ny - cyPulse)**2 + (nz - 0.5)**2);
          const curvedPulse = Math.sin(distToCenter * 12 - time * 1.5);

          // 综合判断：流体曲线斑块 + 弧形涟漪 + 随机数字火花
          const isGreenFlash = 
            (organicPattern > 0.65 && Math.random() > 0.35) || 
            (curvedPulse > 0.92 && Math.random() > 0.3) || 
            (Math.random() > 0.995);

          const renderColor = isGreenFlash ? colorGreen : colorWhite;

          if (intensity > threshold * 1.5) {
            ctx.fillStyle = hexToRgba(renderColor, 0.9);
            ctx.fillRect(x, y, pixelSize, pixelSize);
          } else if (intensity > threshold) {
            ctx.fillStyle = hexToRgba(renderColor, 0.4);
            ctx.fillRect(x, y, pixelSize, pixelSize);
          } else if (intensity > threshold * 0.4) {
            ctx.fillStyle = hexToRgba(renderColor, 0.15);
            ctx.fillRect(x, y, pixelSize, pixelSize);
          }
        }
      }
    }
  }

  animationFrameId = window.requestAnimationFrame(drawScene);
}

onMounted(() => {
  resizeCanvas();
  animationFrameId = window.requestAnimationFrame(drawScene);
  window.addEventListener("resize", resizeCanvas);
  document.addEventListener("mousemove", handleMouseMove);
});

onBeforeUnmount(() => {
  window.removeEventListener("resize", resizeCanvas);
  document.removeEventListener("mousemove", handleMouseMove);
  window.cancelAnimationFrame(animationFrameId);
});
</script>

<template>
  <canvas id="space-canvas" ref="canvasRef"></canvas>
</template>

<style scoped>
#space-canvas {
  position: fixed;
  inset: 0;
  width: 100%;
  height: 100%;
  z-index: 0;
  pointer-events: none;
}
</style>
