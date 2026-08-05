<template>
  <div class="relative w-full h-full overflow-hidden" ref="container" :style="{ backgroundColor: bgColor }">
    <canvas ref="canvas" class="absolute inset-0 w-full h-full pointer-events-auto cursor-crosshair"></canvas>
    
    <!-- Cursor Glow -->
    <div 
      class="pointer-events-none absolute rounded-full blur-[80px]"
      :style="{
        width: `${glowRadius * 2}px`,
        height: `${glowRadius * 2}px`,
        transform: `translate(${mouseX - glowRadius}px, ${mouseY - glowRadius}px)`,
        background: glowColor,
        opacity: isHovering ? 1 : 0,
        transition: 'opacity 0.3s ease'
      }"
    ></div>
    
    <div class="relative z-10 w-full h-full pointer-events-none">
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue';

const props = defineProps({
  dotRadius: { type: Number, default: 1.5 },
  dotSpacing: { type: Number, default: 16 },
  cursorRadius: { type: Number, default: 150 },
  cursorForce: { type: Number, default: 0.2 },
  bulgeOnly: { type: Boolean, default: true },
  bulgeStrength: { type: Number, default: 40 },
  glowRadius: { type: Number, default: 120 },
  sparkle: { type: Boolean, default: true },
  waveAmplitude: { type: Number, default: 0 },
  gradientFrom: { type: String, default: 'rgba(66, 255, 209, 0.4)' },
  gradientTo: { type: String, default: 'rgba(255, 255, 255, 0.05)' },
  glowColor: { type: String, default: 'rgba(66, 255, 209, 0.15)' },
  bgColor: { type: String, default: 'transparent' }
});

const container = ref<HTMLElement | null>(null);
const canvas = ref<HTMLCanvasElement | null>(null);
const mouseX = ref(-1000);
const mouseY = ref(-1000);
const isHovering = ref(false);

let ctx: CanvasRenderingContext2D | null = null;
let animationFrameId: number | null = null;
let width = 0;
let height = 0;
let time = 0;
let dots: Dot[] = [];

class Dot {
  baseX: number;
  baseY: number;
  x: number;
  y: number;
  vx: number;
  vy: number;
  isSparkle: boolean;
  sparkleOffset: number;

  constructor(x: number, y: number) {
    this.baseX = x;
    this.baseY = y;
    this.x = x;
    this.y = y;
    this.vx = 0;
    this.vy = 0;
    this.isSparkle = props.sparkle && Math.random() < 0.05;
    this.sparkleOffset = Math.random() * Math.PI * 2;
  }

  update(mx: number, my: number, isHovering: boolean, t: number) {
    let waveY = 0;
    if (props.waveAmplitude > 0) {
      waveY = Math.sin(this.baseX * 0.01 + t) * props.waveAmplitude;
    }

    if (!isHovering) {
      this.x += (this.baseX - this.x) * 0.1;
      this.y += ((this.baseY + waveY) - this.y) * 0.1;
      return;
    }

    const dx = mx - this.baseX;
    const dy = my - this.baseY;
    const dist = Math.sqrt(dx * dx + dy * dy);

    if (dist < props.cursorRadius) {
      if (props.bulgeOnly) {
        // Bulge effect
        const bulgeVal = Math.sin((dist / props.cursorRadius) * Math.PI);
        const bulgeScale = bulgeVal * props.bulgeStrength * (1 - dist / props.cursorRadius);
        
        const angle = Math.atan2(dy, dx);
        const targetX = this.baseX - Math.cos(angle) * bulgeScale;
        const targetY = this.baseY - Math.sin(angle) * bulgeScale + waveY;

        this.x += (targetX - this.x) * 0.2;
        this.y += (targetY - this.y) * 0.2;
      } else {
        // Physics push effect
        const force = (props.cursorRadius - dist) * props.cursorForce;
        const angle = Math.atan2(dy, dx);
        this.vx -= Math.cos(angle) * force;
        this.vy -= Math.sin(angle) * force;
        
        this.x += this.vx;
        this.y += this.vy + waveY;
        
        this.vx *= 0.8;
        this.vy *= 0.8;
        
        this.x += (this.baseX - this.x) * 0.1;
        this.y += (this.baseY - this.y) * 0.1;
      }
    } else {
      this.x += (this.baseX - this.x) * 0.1;
      this.y += ((this.baseY + waveY) - this.y) * 0.1;
    }
  }

  draw(ctx: CanvasRenderingContext2D, t: number) {
    let r = props.dotRadius;
    if (this.isSparkle) {
      r += Math.sin(t * 3 + this.sparkleOffset) * 0.8;
      r = Math.max(0.5, r);
    }

    ctx.beginPath();
    ctx.arc(this.x, this.y, r, 0, Math.PI * 2);
    ctx.fill();
  }
}

const initDots = () => {
  dots = [];
  for (let x = props.dotSpacing; x < width; x += props.dotSpacing) {
    for (let y = props.dotSpacing; y < height; y += props.dotSpacing) {
      dots.push(new Dot(x, y));
    }
  }
};

const handleMouseMove = (e: MouseEvent) => {
  if (!canvas.value) return;
  const rect = canvas.value.getBoundingClientRect();
  mouseX.value = e.clientX - rect.left;
  mouseY.value = e.clientY - rect.top;
};

const handleMouseEnter = () => {
  isHovering.value = true;
};

const handleMouseLeave = () => {
  isHovering.value = false;
};

const handleResize = () => {
  if (!container.value || !canvas.value) return;
  const rect = container.value.getBoundingClientRect();
  width = rect.width;
  height = rect.height;
  canvas.value.width = width;
  canvas.value.height = height;
  initDots();
};

const animate = () => {
  if (!ctx) return;
  ctx.clearRect(0, 0, width, height);
  time += 0.05;

  const gradient = ctx.createLinearGradient(0, 0, width, height);
  gradient.addColorStop(0, props.gradientFrom);
  gradient.addColorStop(1, props.gradientTo);
  ctx.fillStyle = gradient;

  for (const dot of dots) {
    dot.update(mouseX.value, mouseY.value, isHovering.value, time);
    dot.draw(ctx, time);
  }

  animationFrameId = requestAnimationFrame(animate);
};

onMounted(() => {
  if (canvas.value) {
    ctx = canvas.value.getContext('2d');
    handleResize();
    window.addEventListener('resize', handleResize);
    
    canvas.value.addEventListener('mousemove', handleMouseMove);
    canvas.value.addEventListener('mouseenter', handleMouseEnter);
    canvas.value.addEventListener('mouseleave', handleMouseLeave);

    animate();
  }
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize);
  if (canvas.value) {
    canvas.value.removeEventListener('mousemove', handleMouseMove);
    canvas.value.removeEventListener('mouseenter', handleMouseEnter);
    canvas.value.removeEventListener('mouseleave', handleMouseLeave);
  }
  if (animationFrameId !== null) {
    cancelAnimationFrame(animationFrameId);
  }
});
</script>
