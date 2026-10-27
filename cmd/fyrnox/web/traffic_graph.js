/**
 * FYRNOX Real-Time Dual-Series Throughput Waveform
 * High-performance DPI-aware HTML5 Canvas renderer.
 * Features:
 * - Rolling 60-second real-time download and upload curves
 * - Smooth Catmull-Rom / cubic Bezier spline interpolation
 * - Glowing orange download curve (#FF6A00) with vertical area gradient
 * - Glowing cyan upload curve (#00D2FF) with vertical area gradient
 * - Dynamic auto-scaling Y-axis with smooth damping
 * - Subtle ambient breathing baseline wave when idle/disconnected
 */

(function (window) {
  'use strict';

  function formatBytesSpeed(bytesPerSec) {
    if (!bytesPerSec || bytesPerSec <= 0) return '0 B/s';
    if (bytesPerSec < 1024) return bytesPerSec.toFixed(0) + ' B/s';
    if (bytesPerSec < 1024 * 1024) return (bytesPerSec / 1024).toFixed(1) + ' KB/s';
    if (bytesPerSec < 1024 * 1024 * 1024) return (bytesPerSec / (1024 * 1024)).toFixed(1) + ' MB/s';
    return (bytesPerSec / (1024 * 1024 * 1024)).toFixed(2) + ' GB/s';
  }

  class FyrnoxTrafficGraph {
    constructor(canvasId) {
      this.canvas = document.getElementById(canvasId);
      if (!this.canvas) {
        console.warn(`[FyrnoxTrafficGraph] Canvas #${canvasId} not found`);
        return;
      }

      this.ctx = this.canvas.getContext('2d');
      this.pointsCount = 45; // 45 points along the width
      this.historyDown = new Array(this.pointsCount).fill(0);
      this.historyUp = new Array(this.pointsCount).fill(0);

      this.currentMaxY = 1024 * 1024; // 1 MB/s initial scale
      this.targetMaxY = 1024 * 1024;

      this.latestDown = 0;
      this.latestUp = 0;
      this.isConnected = false;
      this.phase = 0;

      this.init();
    }

    init() {
      this.handleResize();
      window.addEventListener('resize', () => this.handleResize());

      if (window.ResizeObserver) {
        const ro = new ResizeObserver(() => this.handleResize());
        ro.observe(this.canvas.parentElement || this.canvas);
      }

      this.render = this.render.bind(this);
      requestAnimationFrame(this.render);
    }

    handleResize() {
      if (!this.canvas) return;
      const rect = this.canvas.getBoundingClientRect();
      const dpr = Math.min(window.devicePixelRatio || 1, 2);

      this.width = rect.width || 340;
      this.height = rect.height || 64;

      this.canvas.width = Math.round(this.width * dpr);
      this.canvas.height = Math.round(this.height * dpr);

      this.ctx.resetTransform?.();
      this.ctx.scale(dpr, dpr);
    }

    update(speedDown, speedUp, isConnected) {
      this.latestDown = speedDown || 0;
      this.latestUp = speedUp || 0;
      this.isConnected = !!isConnected;

      // Shift history and push latest measurement
      this.historyDown.shift();
      this.historyDown.push(this.latestDown);

      this.historyUp.shift();
      this.historyUp.push(this.latestUp);

      // Determine max value for auto-scaling
      let peak = Math.max(...this.historyDown, ...this.historyUp, 50 * 1024); // Minimum floor: 50 KB/s
      this.targetMaxY = peak * 1.35; // Headroom
    }

    drawSpline(data, strokeColor, fillColor, lineWidth, isIdleWave) {
      const w = this.width;
      const h = this.height;
      const paddingBottom = 6;
      const paddingTop = 8;
      const graphHeight = h - paddingBottom - paddingTop;
      const maxY = Math.max(this.currentMaxY, 100);

      const pts = [];
      const step = w / (data.length - 1);

      for (let i = 0; i < data.length; i++) {
        let val = data[i];
        if (isIdleWave) {
          // Ambient breathing sine wave
          const wave = Math.sin(this.phase * 2 + i * 0.45) * 4 + Math.sin(this.phase * 1.2 + i * 0.3) * 3;
          val = Math.max(0, 15 + wave);
        }

        const normalized = Math.min(1.0, Math.max(0, val / maxY));
        const x = i * step;
        const y = h - paddingBottom - normalized * graphHeight;
        pts.push({ x, y });
      }

      if (pts.length < 2) return;

      this.ctx.beginPath();
      this.ctx.moveTo(pts[0].x, pts[0].y);

      for (let i = 0; i < pts.length - 1; i++) {
        const p0 = pts[i === 0 ? 0 : i - 1];
        const p1 = pts[i];
        const p2 = pts[i + 1];
        const p3 = pts[i + 2 < pts.length ? i + 2 : i + 1];

        // Catmull-Rom to Cubic Bezier conversion
        const cp1x = p1.x + (p2.x - p0.x) / 6;
        const cp1y = p1.y + (p2.y - p0.y) / 6;
        const cp2x = p2.x - (p3.x - p1.x) / 6;
        const cp2y = p2.y - (p3.y - p1.y) / 6;

        this.ctx.bezierCurveTo(cp1x, cp1y, cp2x, cp2y, p2.x, p2.y);
      }

      // Stroke Line
      this.ctx.save();
      this.ctx.strokeStyle = strokeColor;
      this.ctx.lineWidth = lineWidth;
      this.ctx.lineCap = 'round';
      this.ctx.lineJoin = 'round';
      this.ctx.shadowColor = strokeColor;
      this.ctx.shadowBlur = 8;
      this.ctx.stroke();

      // Area Fill
      if (fillColor) {
        this.ctx.lineTo(w, h);
        this.ctx.lineTo(0, h);
        this.ctx.closePath();
        this.ctx.fillStyle = fillColor;
        this.ctx.shadowBlur = 0;
        this.ctx.fill();
      }
      this.ctx.restore();
    }

    render() {
      requestAnimationFrame(this.render);

      this.phase += 0.035;

      // Smooth MaxY interpolation
      this.currentMaxY += (this.targetMaxY - this.currentMaxY) * 0.08;

      const w = this.width;
      const h = this.height;
      if (!w || !h) return;

      this.ctx.clearRect(0, 0, w, h);

      const hasTraffic = this.isConnected && (this.latestDown > 50 || this.latestUp > 50);

      // Gradients for area fills
      const downGrad = this.ctx.createLinearGradient(0, 0, 0, h);
      downGrad.addColorStop(0, 'rgba(255, 106, 0, 0.22)');
      downGrad.addColorStop(0.8, 'rgba(255, 106, 0, 0.04)');
      downGrad.addColorStop(1, 'rgba(255, 106, 0, 0.0)');

      const upGrad = this.ctx.createLinearGradient(0, 0, 0, h);
      upGrad.addColorStop(0, 'rgba(59, 130, 246, 0.20)');
      upGrad.addColorStop(0.8, 'rgba(59, 130, 246, 0.03)');
      upGrad.addColorStop(1, 'rgba(0, 210, 255, 0.0)');

      if (hasTraffic) {
        // Real-time active curves
        this.drawSpline(this.historyUp, '#3B82F6', upGrad, 1.8, false);
        this.drawSpline(this.historyDown, '#ff6a00', downGrad, 2.2, false);
      } else {
        // Ambient elegant wave when idle
        const idleDown = [10, 15, 25, 38, 28, 20, 32, 45, 35, 22, 18, 30, 42, 32, 18, 24, 38, 48, 35, 20];
        const idleUp = [8, 12, 18, 24, 20, 14, 22, 30, 25, 16, 12, 20, 28, 22, 14, 18, 26, 32, 24, 15];
        this.drawSpline(idleUp, 'rgba(59, 130, 246, 0.75)', upGrad, 1.5, true);
        this.drawSpline(idleDown, 'rgba(255, 106, 0, 0.85)', downGrad, 1.8, true);
      }

      // Update graph legend texts
      const legDown = document.getElementById('graph-legend-down');
      const legUp = document.getElementById('graph-legend-up');
      if (legDown) {
        legDown.textContent = this.isConnected ? formatBytesSpeed(this.latestDown) : '—';
      }
      if (legUp) {
        legUp.textContent = this.isConnected ? formatBytesSpeed(this.latestUp) : '—';
      }
    }
  }

  window.FyrnoxTrafficGraph = FyrnoxTrafficGraph;

})(window);
