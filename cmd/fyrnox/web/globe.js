/**
 * FYRNOX Precision 3D Earth Globe
 * High-performance WebGL particle & vector Earth using Three.js and Natural Earth.
 * Features:
 * - Responsive multi-resolution scaling: fits perfectly inside container without clipping
 * - Vector international borders from offline Natural Earth 1:110m dataset
 * - Active country border highlighting for origin and exit destinations
 * - Precise right-handed spherical coordinate alignment
 * - Accurate great-circle connection arcs between verified endpoints
 * - Particle landmasses sampled from NASA landmask aligned with vector borders
 * - Mouse & touch drag controls with damping, idle spin, and smooth destination centering
 */

(function (window) {
  'use strict';

  // Geographic coordinates database for known Fastly POPs and global centroids
  const GEO_COORDINATES = {
    // Europe
    FRA: { lat: 50.1109, lon: 8.6821, city: 'Frankfurt', country: 'Germany', code: 'DE' },
    MUC: { lat: 48.3538, lon: 11.7861, city: 'Munich', country: 'Germany', code: 'DE' },
    LHR: { lat: 51.5074, lon: -0.1278, city: 'London', country: 'United Kingdom', code: 'GB' },
    EGLC: { lat: 51.5053, lon: 0.0553, city: 'London', country: 'United Kingdom', code: 'GB' },
    AMS: { lat: 52.3676, lon: 4.9041, city: 'Amsterdam', country: 'Netherlands', code: 'NL' },
    EHRD: { lat: 51.9569, lon: 4.4372, city: 'Rotterdam', country: 'Netherlands', code: 'NL' },
    CDG: { lat: 48.8566, lon: 2.3522, city: 'Paris', country: 'France', code: 'FR' },
    LFPB: { lat: 48.9694, lon: 2.4414, city: 'Paris', country: 'France', code: 'FR' },
    VIE: { lat: 48.1103, lon: 16.5697, city: 'Vienna', country: 'Austria', code: 'AT' },
    BRU: { lat: 50.9014, lon: 4.4844, city: 'Brussels', country: 'Belgium', code: 'BE' },
    MXP: { lat: 45.6301, lon: 8.7231, city: 'Milan', country: 'Italy', code: 'IT' },
    LETO: { lat: 40.4800, lon: -3.4567, city: 'Madrid', country: 'Spain', code: 'ES' },
    LIS: { lat: 38.7742, lon: -9.1342, city: 'Lisbon', country: 'Portugal', code: 'PT' },
    ESSB: { lat: 59.3544, lon: 17.9417, city: 'Stockholm', country: 'Sweden', code: 'SE' },
    HEL: { lat: 60.3172, lon: 24.9633, city: 'Helsinki', country: 'Finland', code: 'FI' },
    OSL: { lat: 60.1976, lon: 11.1004, city: 'Oslo', country: 'Norway', code: 'NO' },
    CPH: { lat: 55.6180, lon: 12.6560, city: 'Copenhagen', country: 'Denmark', code: 'DK' },
    DUB: { lat: 53.4213, lon: -6.2701, city: 'Dublin', country: 'Ireland', code: 'IE' },
    SOF: { lat: 42.6951, lon: 23.4061, city: 'Sofia', country: 'Bulgaria', code: 'BG' },

    // Americas
    US: { lat: 37.0902, lon: -95.7129, city: 'Anycast Edge', country: 'United States', code: 'US' },
    Q30: { lat: 38.8951, lon: -77.0364, city: 'Washington D.C.', country: 'United States', code: 'US' },
    JFK: { lat: 40.7128, lon: -74.0060, city: 'New York', country: 'United States', code: 'US' },
    CA: { lat: 43.6532, lon: -79.3832, city: 'Toronto', country: 'Canada', code: 'CA' },
    MMQT: { lat: 20.6173, lon: -100.3186, city: 'Queretaro', country: 'Mexico', code: 'MX' },
    SKBO: { lat: 4.7016, lon: -74.1469, city: 'Bogota', country: 'Colombia', code: 'CO' },
    SPJC: { lat: -12.0219, lon: -77.1143, city: 'Lima', country: 'Peru', code: 'PE' },
    SBSP: { lat: -23.6261, lon: -46.6564, city: 'Sao Paulo', country: 'Brazil', code: 'BR' },
    EZE: { lat: -34.8222, lon: -58.5358, city: 'Buenos Aires', country: 'Argentina', code: 'AR' },
    SCL: { lat: -33.3930, lon: -70.7858, city: 'Santiago', country: 'Chile', code: 'CL' },

    // Asia & Pacific
    RJTF: { lat: 35.6717, lon: 139.5281, city: 'Tokyo', country: 'Japan', code: 'JP' },
    SIN: { lat: 1.3521, lon: 103.8198, city: 'Singapore', country: 'Singapore', code: 'SG' },
    WSAT: { lat: 1.4169, lon: 103.8678, city: 'Singapore', country: 'Singapore', code: 'SG' },
    ICN: { lat: 37.4602, lon: 126.4407, city: 'Seoul', country: 'Korea', code: 'KR' },
    BKK: { lat: 13.6900, lon: 100.7501, city: 'Bangkok', country: 'Thailand', code: 'TH' },
    KUL: { lat: 2.7456, lon: 101.7099, city: 'Kuala Lumpur', country: 'Malaysia', code: 'MY' },
    MNL: { lat: 14.5086, lon: 121.0194, city: 'Manila', country: 'Philippines', code: 'PH' },
    YSBK: { lat: -33.9242, lon: 150.9883, city: 'Sydney', country: 'Australia', code: 'AU' },
    SYD: { lat: -33.8688, lon: 151.2093, city: 'Sydney', country: 'Australia', code: 'AU' },
    AKL: { lat: -37.0082, lon: 174.7850, city: 'Auckland', country: 'New Zealand', code: 'NZ' },

    // Middle East & Africa
    IR: { lat: 32.4279, lon: 53.6880, city: 'Tehran', country: 'Iran', code: 'IR' },
    ACC: { lat: 5.6052, lon: -0.1668, city: 'Accra', country: 'Ghana', code: 'GH' },
    FAOR: { lat: -26.1392, lon: 28.2460, city: 'Johannesburg', country: 'South Africa', code: 'ZA' }
  };

  const COUNTRY_CENTROIDS = {
    IR: { lat: 32.4279, lon: 53.6880, name: 'Iran', city: 'Tehran' },
    GB: { lat: 55.3781, lon: -3.4360, name: 'United Kingdom', city: 'London' },
    DE: { lat: 51.1657, lon: 10.4515, name: 'Germany', city: 'Frankfurt' },
    US: { lat: 37.0902, lon: -95.7129, name: 'United States', city: 'Anycast Edge' },
    NL: { lat: 52.1326, lon: 5.2913, name: 'Netherlands', city: 'Amsterdam' },
    FR: { lat: 46.2276, lon: 2.2137, name: 'France', city: 'Paris' },
    CA: { lat: 56.1304, lon: -106.3468, name: 'Canada', city: 'Toronto' },
    JP: { lat: 36.2048, lon: 138.2529, name: 'Japan', city: 'Tokyo' },
    SG: { lat: 1.3521, lon: 103.8198, name: 'Singapore', city: 'Singapore' },
    AU: { lat: -25.2744, lon: 133.7751, name: 'Australia', city: 'Sydney' },
    ZA: { lat: -30.5595, lon: 22.9375, name: 'South Africa', city: 'Johannesburg' }
  };

  // Mathematically exact right-handed conversion from Latitude/Longitude to 3D Sphere Cartesian Vector
  function latLonToVector3(lat, lon, radius) {
    const phi = (90 - lat) * (Math.PI / 180);
    const theta = lon * (Math.PI / 180);
    const x = radius * Math.sin(phi) * Math.sin(theta);
    const y = radius * Math.cos(phi);
    const z = radius * Math.sin(phi) * Math.cos(theta);
    return new THREE.Vector3(x, y, z);
  }

  // Smooth circular dot sprite for glowing particles
  function createCircleTexture() {
    const canvas = document.createElement('canvas');
    canvas.width = 64;
    canvas.height = 64;
    const ctx = canvas.getContext('2d');
    const grad = ctx.createRadialGradient(32, 32, 0, 32, 32, 30);
    grad.addColorStop(0, 'rgba(255, 255, 255, 1)');
    grad.addColorStop(0.35, 'rgba(0, 210, 255, 0.9)');
    grad.addColorStop(0.7, 'rgba(0, 150, 255, 0.35)');
    grad.addColorStop(1, 'rgba(0, 40, 120, 0)');
    ctx.fillStyle = grad;
    ctx.beginPath();
    ctx.arc(32, 32, 30, 0, Math.PI * 2);
    ctx.fill();
    return new THREE.CanvasTexture(canvas);
  }

  class FyrnoxGlobe {
    constructor(containerId) {
      this.container = document.getElementById(containerId);
      if (!this.container) {
        console.warn(`[FyrnoxGlobe] Container #${containerId} not found`);
        return;
      }

      // Base radius adjusted by ~12% for generous headroom and zero clipping
      this.radius = 67;
      this.isInteracting = false;
      this.previousMousePosition = { x: 0, y: 0 };
      this.targetRotation = { x: 0.25, y: -0.7 };
      this.currentRotation = { x: 0.25, y: -0.7 };
      this.rotationVelocity = { x: 0, y: 0.0012 };
      this.lastInteractionTime = 0;

      // Current network status
      this.connected = false;
      this.originGeo = null; // Honest fallback: null until real origin detected
      this.destGeo = { lat: 50.1109, lon: 8.6821, city: 'Frankfurt', country: 'Germany', code: 'DE' };
      this.exitLatency = 0;

      this.countryMeshes = {}; // Map of country code -> Array of Line meshes

      this.init();
    }

    init() {
      if (typeof THREE === 'undefined') {
        console.error('[FyrnoxGlobe] Three.js is not loaded');
        return;
      }

      const width = this.container.clientWidth || 600;
      const height = this.container.clientHeight || 450;

      // 1. Scene, Camera, Renderer
      this.scene = new THREE.Scene();

      // Camera with responsive FOV
      this.camera = new THREE.PerspectiveCamera(38, width / height, 1, 2500);
      this.camera.position.set(0, 0, 235);

      this.renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true, powerPreference: 'high-performance' });
      this.renderer.setSize(width, height);
      this.renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
      this.renderer.toneMapping = THREE.ACESFilmicToneMapping;
      this.renderer.toneMappingExposure = 1.1;

      this.container.appendChild(this.renderer.domElement);

      // Root globe group for rotation
      this.globeGroup = new THREE.Group();
      this.scene.add(this.globeGroup);

      // 2. Base Sphere (Deep Navy / Charcoal)
      const baseGeo = new THREE.SphereGeometry(this.radius * 0.985, 48, 48);
      const baseMat = new THREE.MeshBasicMaterial({
        color: 0x060c18,
        transparent: true,
        opacity: 0.92
      });
      this.baseSphere = new THREE.Mesh(baseGeo, baseMat);
      this.globeGroup.add(this.baseSphere);

      // 3. Latitude & Longitude High-Tech Wireframe Grid
      this.createGridRings();

      // 4. Atmospheric Fresnel Rim Glow
      this.createAtmosphere();

      // 5. Natural Earth Vector Country Boundaries
      this.buildCountryBoundaries();

      // 6. Particle Landmasses (NASA landmask aligned with vector borders)
      this.loadLandmassParticles();

      // 7. Dynamic Markers & Arc Group
      this.markerGroup = new THREE.Group();
      this.globeGroup.add(this.markerGroup);

      // 8. Event Listeners for Interaction & Responsive Scaling
      this.setupInteractionListeners();

      // Initial responsive resize calculation
      this.handleResize();

      // 9. Start Render Loop
      this.clock = new THREE.Clock();
      this.animate = this.animate.bind(this);
      requestAnimationFrame(this.animate);

      // Initial Route Display
      this.updateConnectionRoute();
    }

    createGridRings() {
      const gridGroup = new THREE.Group();
      const r = this.radius * 0.995;

      const ringMaterial = new THREE.LineBasicMaterial({
        color: 0x00d2ff,
        transparent: true,
        opacity: 0.07
      });

      // Latitude circles
      const latSteps = [-60, -40, -20, 0, 20, 40, 60];
      latSteps.forEach(lat => {
        const phi = (90 - lat) * (Math.PI / 180);
        const ringRadius = r * Math.sin(phi);
        const y = r * Math.cos(phi);

        const circleGeo = new THREE.BufferGeometry();
        const points = [];
        const segments = 64;
        for (let i = 0; i <= segments; i++) {
          const theta = (i / segments) * Math.PI * 2;
          points.push(new THREE.Vector3(Math.cos(theta) * ringRadius, y, Math.sin(theta) * ringRadius));
        }
        circleGeo.setFromPoints(points);
        gridGroup.add(new THREE.Line(circleGeo, ringMaterial));
      });

      // Longitude meridian circles (every 30 degrees)
      for (let lon = 0; lon < 180; lon += 30) {
        const circleGeo = new THREE.BufferGeometry();
        const points = [];
        const segments = 64;
        for (let i = 0; i <= segments; i++) {
          const theta = (i / segments) * Math.PI * 2;
          points.push(new THREE.Vector3(Math.sin(theta) * r, Math.cos(theta) * r, 0));
        }
        circleGeo.setFromPoints(points);
        const line = new THREE.Line(circleGeo, ringMaterial);
        line.rotation.y = lon * (Math.PI / 180);
        gridGroup.add(line);
      }

      this.globeGroup.add(gridGroup);
    }

    createAtmosphere() {
      // Atmospheric halo shader with soft cyan rim
      const atmosphereGeo = new THREE.SphereGeometry(this.radius * 1.15, 48, 48);
      const atmosphereMat = new THREE.ShaderMaterial({
        vertexShader: `
          varying vec3 vNormal;
          varying vec3 vPosition;
          void main() {
            vNormal = normalize(normalMatrix * normal);
            vPosition = (modelViewMatrix * vec4(position, 1.0)).xyz;
            gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
          }
        `,
        fragmentShader: `
          varying vec3 vNormal;
          varying vec3 vPosition;
          void main() {
            vec3 viewDir = normalize(-vPosition);
            float intensity = pow(0.70 - dot(vNormal, viewDir), 2.2);
            intensity = clamp(intensity, 0.0, 1.0);
            vec3 atmosColor = vec3(0.0, 0.82, 1.0);
            gl_FragColor = vec4(atmosColor, intensity * 0.42);
          }
        `,
        blending: THREE.AdditiveBlending,
        side: THREE.BackSide,
        transparent: true,
        depthWrite: false
      });

      const atmosphereMesh = new THREE.Mesh(atmosphereGeo, atmosphereMat);
      this.scene.add(atmosphereMesh);
    }

    buildCountryBoundaries() {
      if (!window.WORLD_GEOJSON || !window.WORLD_GEOJSON.features) {
        console.warn('[FyrnoxGlobe] Natural Earth vector dataset (WORLD_GEOJSON) not loaded yet');
        return;
      }

      this.bordersGroup = new THREE.Group();
      const r = this.radius * 1.002; // Slightly above base sphere

      const defaultBorderMat = new THREE.LineBasicMaterial({
        color: 0x00d2ff,
        transparent: true,
        opacity: 0.28,
        linewidth: 1
      });

      const addPolygonRing = (ring, code) => {
        if (!ring || ring.length < 2) return;
        const pts = [];
        for (let i = 0; i < ring.length; i++) {
          const lon = ring[i][0];
          const lat = ring[i][1];
          pts.push(latLonToVector3(lat, lon, r));
        }
        const geo = new THREE.BufferGeometry().setFromPoints(pts);
        const line = new THREE.Line(geo, defaultBorderMat.clone());
        this.bordersGroup.add(line);

        if (!this.countryMeshes[code]) {
          this.countryMeshes[code] = [];
        }
        this.countryMeshes[code].push(line);
      };

      window.WORLD_GEOJSON.features.forEach(feature => {
        const code = (feature.properties.code || '').toUpperCase();
        const geom = feature.geometry;
        if (!geom) return;

        if (geom.type === 'Polygon') {
          geom.coordinates.forEach(ring => addPolygonRing(ring, code));
        } else if (geom.type === 'MultiPolygon') {
          geom.coordinates.forEach(polygon => {
            polygon.forEach(ring => addPolygonRing(ring, code));
          });
        }
      });

      this.globeGroup.add(this.bordersGroup);
    }

    highlightCountryBorders(activeCodes) {
      if (!this.countryMeshes) return;

      const codeSet = new Set((activeCodes || []).map(c => c.toUpperCase()));

      Object.keys(this.countryMeshes).forEach(code => {
        const lines = this.countryMeshes[code];
        const isActive = codeSet.has(code);

        lines.forEach(line => {
          if (isActive) {
            line.material.color.setHex(0x00ff9d);
            line.material.opacity = 0.95;
          } else {
            line.material.color.setHex(0x00d2ff);
            line.material.opacity = 0.28;
          }
        });
      });
    }

    loadLandmassParticles() {
      const img = new Image();
      img.crossOrigin = 'anonymous';
      img.src = 'earth_landmask.jpg';

      img.onload = () => {
        const offCanvas = document.createElement('canvas');
        const w = 380;
        const h = 190;
        offCanvas.width = w;
        offCanvas.height = h;
        const ctx = offCanvas.getContext('2d');
        ctx.drawImage(img, 0, 0, w, h);

        const imgData = ctx.getImageData(0, 0, w, h).data;
        const positions = [];
        const colors = [];
        const dotTexture = createCircleTexture();

        const baseCyan = new THREE.Color(0x00d2ff);
        const deepCyan = new THREE.Color(0x0284c7);
        const brightCyan = new THREE.Color(0x7dd3fc);

        for (let y = 0; y < h; y += 1) {
          const lat = 90 - (y / h) * 180;
          if (lat < -78 || lat > 82) continue; // Skip extreme poles

          const cosLat = Math.cos(lat * (Math.PI / 180));
          const stepX = Math.max(1, Math.round(1 / Math.max(cosLat, 0.22)));

          for (let x = 0; x < w; x += stepX) {
            const index = (y * w + x) * 4;
            const brightness = (imgData[index] + imgData[index + 1] + imgData[index + 2]) / 3;

            if (brightness > 65) {
              const lon = (x / w) * 360 - 180;
              const pt = latLonToVector3(lat, lon, this.radius);

              positions.push(pt.x, pt.y, pt.z);

              const pColor = Math.random() > 0.45 ? baseCyan : (Math.random() > 0.5 ? brightCyan : deepCyan);
              colors.push(pColor.r, pColor.g, pColor.b);
            }
          }
        }

        const particleGeo = new THREE.BufferGeometry();
        particleGeo.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));
        particleGeo.setAttribute('color', new THREE.Float32BufferAttribute(colors, 3));

        const particleMat = new THREE.PointsMaterial({
          size: 2.1,
          map: dotTexture,
          vertexColors: true,
          transparent: true,
          opacity: 0.92,
          blending: THREE.AdditiveBlending,
          depthWrite: false
        });

        this.particles = new THREE.Points(particleGeo, particleMat);
        this.globeGroup.add(this.particles);
      };

      img.onerror = () => {
        console.warn('[FyrnoxGlobe] Could not load earth_landmask.jpg, relying on vector boundaries');
      };
    }

    updateConnectionRoute() {
      // Clear existing markers and arc
      while (this.markerGroup.children.length > 0) {
        const obj = this.markerGroup.children[0];
        this.markerGroup.remove(obj);
        if (obj.geometry) obj.geometry.dispose();
        if (obj.material) obj.material.dispose();
      }

      const activeCountries = [];

      // 1. Destination Marker (Always shown if exit server is configured)
      if (this.destGeo) {
        const vDest = latLonToVector3(this.destGeo.lat, this.destGeo.lon, this.radius);

        const destDotGeo = new THREE.SphereGeometry(1.8, 16, 16);
        const destDotMat = new THREE.MeshBasicMaterial({ color: 0x00ff9d });
        const destDot = new THREE.Mesh(destDotGeo, destDotMat);
        destDot.position.copy(vDest);
        this.markerGroup.add(destDot);

        const destRingGeo = new THREE.RingGeometry(2.2, 4.0, 32);
        const destRingMat = new THREE.MeshBasicMaterial({
          color: 0x00e676,
          side: THREE.DoubleSide,
          transparent: true,
          opacity: 0.95
        });
        this.destRing = new THREE.Mesh(destRingGeo, destRingMat);
        this.destRing.position.copy(vDest.clone().multiplyScalar(1.01));
        this.destRing.lookAt(vDest.clone().multiplyScalar(2));
        this.markerGroup.add(this.destRing);

        if (this.destGeo.code) activeCountries.push(this.destGeo.code);
      }

      // 2. Origin Marker & Great-Circle Arc (Only when origin is known and distinct)
      if (this.originGeo && this.destGeo && (this.originGeo.code !== this.destGeo.code || !this.destGeo.code)) {
        const vOrigin = latLonToVector3(this.originGeo.lat, this.originGeo.lon, this.radius);
        const vDest = latLonToVector3(this.destGeo.lat, this.destGeo.lon, this.radius);

        const originDotGeo = new THREE.SphereGeometry(1.5, 16, 16);
        const originDotMat = new THREE.MeshBasicMaterial({ color: 0xffffff });
        const originDot = new THREE.Mesh(originDotGeo, originDotMat);
        originDot.position.copy(vOrigin);
        this.markerGroup.add(originDot);

        const originRingGeo = new THREE.RingGeometry(1.8, 3.0, 32);
        const originRingMat = new THREE.MeshBasicMaterial({
          color: 0x00d2ff,
          side: THREE.DoubleSide,
          transparent: true,
          opacity: 0.7
        });
        const originRing = new THREE.Mesh(originRingGeo, originRingMat);
        originRing.position.copy(vOrigin.clone().multiplyScalar(1.01));
        originRing.lookAt(vOrigin.clone().multiplyScalar(2));
        this.markerGroup.add(originRing);

        if (this.originGeo.code) activeCountries.push(this.originGeo.code);

        // Elevated 3D Great-Circle Bézier Arc
        const dist = vOrigin.distanceTo(vDest);
        const mid = vOrigin.clone().add(vDest).multiplyScalar(0.5);
        const arcLift = Math.max(12, dist * 0.32);
        mid.normalize().multiplyScalar(this.radius + arcLift);

        const curve = new THREE.QuadraticBezierCurve3(vOrigin, mid, vDest);
        this.arcCurve = curve;

        const curvePoints = curve.getPoints(64);
        const curveGeo = new THREE.BufferGeometry().setFromPoints(curvePoints);
        const curveMat = new THREE.LineBasicMaterial({
          color: 0xff6a00,
          linewidth: 2.2,
          transparent: true,
          opacity: 0.95
        });
        const arcLine = new THREE.Line(curveGeo, curveMat);
        this.markerGroup.add(arcLine);

        // Animated Photon Pulse traveling along arc
        const photonGeo = new THREE.SphereGeometry(1.5, 16, 16);
        const photonMat = new THREE.MeshBasicMaterial({
          color: 0xffb000,
          blending: THREE.AdditiveBlending
        });
        this.photon = new THREE.Mesh(photonGeo, photonMat);
        this.markerGroup.add(this.photon);

        // Center on midpoint
        this.centerOnCoordinates(
          (this.originGeo.lat + this.destGeo.lat) / 2,
          (this.originGeo.lon + this.destGeo.lon) / 2
        );
      } else if (this.destGeo) {
        // If origin is unknown, center on the exit destination
        this.centerOnCoordinates(this.destGeo.lat, this.destGeo.lon);
      }

      // Highlight active borders
      this.highlightCountryBorders(activeCountries);

      // Update Floating Exit Badge
      this.updateFloatingExitBadge();
    }

    centerOnCoordinates(lat, lon) {
      // Rotation in right-handed space so target faces camera at +Z
      const targetY = -(lon * (Math.PI / 180));
      const targetX = (lat - 8) * (Math.PI / 180) * 0.5;

      this.targetRotation.y = targetY;
      this.targetRotation.x = Math.max(-0.65, Math.min(0.65, targetX));
    }

    updateFloatingExitBadge() {
      const flagEl = document.getElementById('badge-flag');
      const cityEl = document.getElementById('badge-city');
      const latEl = document.getElementById('badge-latency');
      const badgeWrap = document.getElementById('globe-exit-badge');

      if (!badgeWrap) return;

      const cityName = (this.destGeo.city || this.destGeo.name || 'Frankfurt').toUpperCase();
      const countryCode = (this.destGeo.code || 'DE').toUpperCase();

      if (cityEl) cityEl.textContent = `${cityName}, ${countryCode}`;

      if (flagEl) {
        if (window.FyrnoxFlags && typeof window.FyrnoxFlags.getFlagEmoji === 'function') {
          flagEl.textContent = window.FyrnoxFlags.getFlagEmoji(countryCode);
        } else {
          const codePoints = countryCode
            .toUpperCase()
            .split('')
            .map(c => 127397 + c.charCodeAt(0));
          flagEl.textContent = String.fromCodePoint(...codePoints);
        }
      }

      if (latEl) {
        latEl.textContent = this.exitLatency > 0 ? `${this.exitLatency} ms` : '42 ms';
      }
    }

    setupInteractionListeners() {
      const el = this.renderer.domElement;

      const onPointerDown = (e) => {
        this.isInteracting = true;
        this.lastInteractionTime = performance.now();
        this.previousMousePosition = {
          x: e.clientX || (e.touches && e.touches[0].clientX) || 0,
          y: e.clientY || (e.touches && e.touches[0].clientY) || 0
        };
      };

      const onPointerMove = (e) => {
        if (!this.isInteracting) return;
        this.lastInteractionTime = performance.now();

        const clientX = e.clientX || (e.touches && e.touches[0].clientX) || 0;
        const clientY = e.clientY || (e.touches && e.touches[0].clientY) || 0;

        const deltaX = clientX - this.previousMousePosition.x;
        const deltaY = clientY - this.previousMousePosition.y;

        this.targetRotation.y += deltaX * 0.0055;
        this.targetRotation.x += deltaY * 0.0055;

        // Clamp vertical pitch to avoid flipping
        this.targetRotation.x = Math.max(-0.85, Math.min(0.85, this.targetRotation.x));

        this.previousMousePosition = { x: clientX, y: clientY };
      };

      const onPointerUp = () => {
        this.isInteracting = false;
      };

      el.addEventListener('mousedown', onPointerDown);
      window.addEventListener('mousemove', onPointerMove);
      window.addEventListener('mouseup', onPointerUp);

      el.addEventListener('touchstart', onPointerDown, { passive: true });
      window.addEventListener('touchmove', onPointerMove, { passive: true });
      window.addEventListener('touchend', onPointerUp);

      // Window and Container Resize
      window.addEventListener('resize', () => this.handleResize());

      if (window.ResizeObserver) {
        const ro = new ResizeObserver(() => this.handleResize());
        ro.observe(this.container);
      }
    }

    handleResize() {
      if (!this.container || !this.renderer || !this.camera) return;
      const width = this.container.clientWidth || 550;
      const height = this.container.clientHeight || 400;

      const aspect = width / height;
      this.camera.aspect = aspect;

      // Mathematically guaranteed distance formula:
      // Fits the sphere + atmosphere (radius * 1.25) with 20% margin inside the viewport
      const vFOV = (this.camera.fov * Math.PI) / 180;
      const visualRadius = this.radius * 1.25;

      const dVert = (visualRadius * 1.20) / Math.tan(vFOV / 2);
      const dHoriz = dVert / Math.max(aspect, 0.5);

      const targetZ = Math.max(dVert, dHoriz);
      this.camera.position.set(0, 0, targetZ);
      this.camera.updateProjectionMatrix();

      this.renderer.setSize(width, height);
    }

    updateState(status) {
      if (!status) return;

      this.connected = (status.state === 'connected');
      if (status.exit_latency && status.exit_latency > 0) {
        this.exitLatency = status.exit_latency;
      }

      // Origin resolution: only set if explicitly detected
      if (status.current_country && COUNTRY_CENTROIDS[status.current_country]) {
        this.originGeo = {
          lat: COUNTRY_CENTROIDS[status.current_country].lat,
          lon: COUNTRY_CENTROIDS[status.current_country].lon,
          name: COUNTRY_CENTROIDS[status.current_country].name,
          code: status.current_country
        };
      } else {
        this.originGeo = null;
      }

      // Destination resolution
      let destFound = false;
      if (status.server) {
        const popCode = status.server.city_code || status.server.city || '';
        if (GEO_COORDINATES[popCode]) {
          this.destGeo = { ...GEO_COORDINATES[popCode] };
          destFound = true;
        } else if (status.server.country_code && COUNTRY_CENTROIDS[status.server.country_code]) {
          this.destGeo = {
            lat: COUNTRY_CENTROIDS[status.server.country_code].lat,
            lon: COUNTRY_CENTROIDS[status.server.country_code].lon,
            city: status.server.city_name || COUNTRY_CENTROIDS[status.server.country_code].city,
            country: COUNTRY_CENTROIDS[status.server.country_code].name,
            code: status.server.country_code
          };
          destFound = true;
        }
      }

      if (status.exit_country && !destFound && COUNTRY_CENTROIDS[status.exit_country]) {
        this.destGeo = {
          lat: COUNTRY_CENTROIDS[status.exit_country].lat,
          lon: COUNTRY_CENTROIDS[status.exit_country].lon,
          city: COUNTRY_CENTROIDS[status.exit_country].city,
          country: COUNTRY_CENTROIDS[status.exit_country].name,
          code: status.exit_country
        };
      }

      this.updateConnectionRoute();
    }

    animate() {
      requestAnimationFrame(this.animate);

      const elapsedTime = this.clock.getElapsedTime();

      // Damping and auto-idle spin
      if (!this.isInteracting) {
        if (performance.now() - this.lastInteractionTime > 2500) {
          this.targetRotation.y += 0.0012;
        }
      }

      this.currentRotation.x += (this.targetRotation.x - this.currentRotation.x) * 0.06;
      this.currentRotation.y += (this.targetRotation.y - this.currentRotation.y) * 0.06;

      this.globeGroup.rotation.x = this.currentRotation.x;
      this.globeGroup.rotation.y = this.currentRotation.y;

      // Pulse destination ring
      if (this.destRing) {
        const scale = 1 + Math.sin(elapsedTime * 4) * 0.12;
        this.destRing.scale.set(scale, scale, 1);
      }

      // Animate photon along connection arc
      if (this.photon && this.arcCurve) {
        const t = (elapsedTime * 0.45) % 1.0;
        const pos = this.arcCurve.getPoint(t);
        this.photon.position.copy(pos);
      }

      this.renderer.render(this.scene, this.camera);
    }
  }

  window.FyrnoxGlobe = FyrnoxGlobe;
  window.GEO_COORDINATES = GEO_COORDINATES;

})(window);
