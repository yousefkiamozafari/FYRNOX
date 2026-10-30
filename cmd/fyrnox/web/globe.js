/**
 * FYRNOX Precision 3D Earth Globe
 * High-performance WebGL particle & vector Earth using Three.js and Natural Earth.
 * Features:
 * - Truly stationary idle state with zero continuous drift, random jitter, or micro-oscillations
 * - Finite, intentional transitions for meaningful events (server selection, connection, disconnection)
 * - Geometrically continuous elevated Great-Circle spherical Slerp route with zero clipping
 * - 3D Tube geometry with vibrant glow and exact marker anchoring
 * - Offline vector country boundaries and responsive multi-resolution scaling
 */

(function (window) {
  'use strict';

  // Geographic coordinates database for known Fastly POPs and regional centroids
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
    ZRH: { lat: 47.4582, lon: 8.5555, city: 'Zurich', country: 'Switzerland', code: 'CH' },
    WAW: { lat: 52.1672, lon: 20.9679, city: 'Warsaw', country: 'Poland', code: 'PL' },
    PRG: { lat: 50.1008, lon: 14.2600, city: 'Prague', country: 'Czech Republic', code: 'CZ' },

    // Americas
    US: { lat: 37.0902, lon: -95.7129, city: 'Anycast Edge', country: 'United States', code: 'US' },
    Q30: { lat: 38.8951, lon: -77.0364, city: 'Washington D.C.', country: 'United States', code: 'US' },
    JFK: { lat: 40.7128, lon: -74.0060, city: 'New York', country: 'United States', code: 'US' },
    ORD: { lat: 41.9742, lon: -87.9073, city: 'Chicago', country: 'United States', code: 'US' },
    SFO: { lat: 37.6213, lon: -122.3790, city: 'San Francisco', country: 'United States', code: 'US' },
    LAX: { lat: 33.9416, lon: -118.4085, city: 'Los Angeles', country: 'United States', code: 'US' },
    MIA: { lat: 25.7959, lon: -80.2870, city: 'Miami', country: 'United States', code: 'US' },
    CA: { lat: 43.6532, lon: -79.3832, city: 'Toronto', country: 'Canada', code: 'CA' },
    YVR: { lat: 49.1967, lon: -123.1815, city: 'Vancouver', country: 'Canada', code: 'CA' },
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
    THR: { lat: 35.6892, lon: 51.3890, city: 'Tehran', country: 'Iran', code: 'IR' },
    DXB: { lat: 25.2048, lon: 55.2708, city: 'Dubai', country: 'United Arab Emirates', code: 'AE' },
    IST: { lat: 41.0082, lon: 28.9784, city: 'Istanbul', country: 'Turkey', code: 'TR' },
    ACC: { lat: 5.6052, lon: -0.1668, city: 'Accra', country: 'Ghana', code: 'GH' },
    FAOR: { lat: -26.1392, lon: 28.2460, city: 'Johannesburg', country: 'South Africa', code: 'ZA' }
  };

  const COUNTRY_CENTROIDS = {
    IR: { lat: 32.4279, lon: 53.6880, name: 'Iran', city: 'Tehran' },
    DE: { lat: 51.1657, lon: 10.4515, name: 'Germany', city: 'Frankfurt' },
    GB: { lat: 55.3781, lon: -3.4360, name: 'United Kingdom', city: 'London' },
    US: { lat: 37.0902, lon: -95.7129, name: 'United States', city: 'Anycast Edge' },
    NL: { lat: 52.1326, lon: 5.2913, name: 'Netherlands', city: 'Amsterdam' },
    FR: { lat: 46.2276, lon: 2.2137, name: 'France', city: 'Paris' },
    CA: { lat: 56.1304, lon: -106.3468, name: 'Canada', city: 'Toronto' },
    SE: { lat: 60.1282, lon: 18.6435, name: 'Sweden', city: 'Stockholm' },
    FI: { lat: 61.9241, lon: 25.7482, name: 'Finland', city: 'Helsinki' },
    NO: { lat: 60.4720, lon: 8.4689, name: 'Norway', city: 'Oslo' },
    DK: { lat: 56.2639, lon: 9.5018, name: 'Denmark', city: 'Copenhagen' },
    AT: { lat: 47.5162, lon: 14.5501, name: 'Austria', city: 'Vienna' },
    CH: { lat: 46.8182, lon: 8.2275, name: 'Switzerland', city: 'Zurich' },
    BE: { lat: 50.5039, lon: 4.4699, name: 'Belgium', city: 'Brussels' },
    IT: { lat: 41.8719, lon: 12.5674, name: 'Italy', city: 'Milan' },
    ES: { lat: 40.4637, lon: -3.7492, name: 'Spain', city: 'Madrid' },
    PT: { lat: 39.3999, lon: -8.2245, name: 'Portugal', city: 'Lisbon' },
    IE: { lat: 53.1424, lon: -7.6921, name: 'Ireland', city: 'Dublin' },
    PL: { lat: 51.9194, lon: 19.1451, name: 'Poland', city: 'Warsaw' },
    CZ: { lat: 49.8175, lon: 15.4730, name: 'Czech Republic', city: 'Prague' },
    RO: { lat: 45.9432, lon: 24.9668, name: 'Romania', city: 'Bucharest' },
    BG: { lat: 42.7339, lon: 25.4858, name: 'Bulgaria', city: 'Sofia' },
    GR: { lat: 39.0742, lon: 21.8243, name: 'Greece', city: 'Athens' },
    TR: { lat: 38.9637, lon: 35.2433, name: 'Turkey', city: 'Istanbul' },
    AE: { lat: 23.4241, lon: 53.8478, name: 'United Arab Emirates', city: 'Dubai' },
    JP: { lat: 36.2048, lon: 138.2529, name: 'Japan', city: 'Tokyo' },
    SG: { lat: 1.3521, lon: 103.8198, name: 'Singapore', city: 'Singapore' },
    KR: { lat: 35.9078, lon: 127.7669, name: 'Korea', city: 'Seoul' },
    AU: { lat: -25.2744, lon: 133.7751, name: 'Australia', city: 'Sydney' },
    NZ: { lat: -40.9006, lon: 174.8860, name: 'New Zealand', city: 'Auckland' },
    BR: { lat: -14.2350, lon: -51.9253, name: 'Brazil', city: 'Sao Paulo' },
    AR: { lat: -38.4161, lon: -63.6167, name: 'Argentina', city: 'Buenos Aires' },
    CL: { lat: -35.6751, lon: -71.5430, name: 'Chile', city: 'Santiago' },
    CO: { lat: 4.5709, lon: -74.2973, name: 'Colombia', city: 'Bogota' },
    MX: { lat: 23.6345, lon: -102.5528, name: 'Mexico', city: 'Queretaro' },
    ZA: { lat: -30.5595, lon: 22.9375, name: 'South Africa', city: 'Johannesburg' },
    IN: { lat: 20.5937, lon: 78.9629, name: 'India', city: 'Mumbai' }
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
    const grad = ctx.createRadialGradient(32, 32, 0, 32, 32, 28);
    grad.addColorStop(0, 'rgba(255, 255, 255, 1.0)');
    grad.addColorStop(0.35, 'rgba(255, 255, 255, 0.75)');
    grad.addColorStop(0.70, 'rgba(255, 255, 255, 0.20)');
    grad.addColorStop(1.0, 'rgba(255, 255, 255, 0.0)');
    ctx.fillStyle = grad;
    ctx.beginPath();
    ctx.arc(32, 32, 28, 0, Math.PI * 2);
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

      this.radius = 60;
      this.isInteracting = false;
      this.previousMousePosition = { x: 0, y: 0 };

      // Stationary State Machine: strictly finite transitions on events, zero idle drift
      this.isTransitioning = false;
      this.transitionStartTime = 0;
      this.transitionDuration = 900; // ms
      this.startRotation = { x: 0.25, y: -0.7 };
      this.targetRotation = { x: 0.25, y: -0.7 };
      this.currentRotation = { x: 0.25, y: -0.7 };

      // Current network status cache to prevent redundant rebuilds
      this.connected = false;
      this.isConnecting = false;
      this.connectionState = 'disconnected';
      this.originGeo = { lat: 50.1109, lon: 8.6821, city: 'Frankfurt', country: 'Germany', code: 'DE' };
      this.destGeo = { lat: -26.1392, lon: 28.2460, city: 'Johannesburg', country: 'South Africa', code: 'ZA' };
      this.exitLatency = 0;

      this.cachedOriginKey = '';
      this.cachedDestKey = '';
      this.cachedConnState = '';
      this.routeBuilt = false;

      this.countryMeshes = {};

      this.init();
    }

    init() {
      const width = this.container.clientWidth || 600;
      const height = this.container.clientHeight || 500;

      // 1. Scene, Camera, Renderer (Strictly synchronized with container aspect ratio)
      this.scene = new THREE.Scene();

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
      this.globeGroup.rotation.x = this.currentRotation.x;
      this.globeGroup.rotation.y = this.currentRotation.y;
      this.scene.add(this.globeGroup);

      // 2. Atmospheric Fresnel Rim Glow (Glassy 3D depth, crystal clear center, zero background disc)
      this.createAtmosphere();

      // 2b. Luminous Orbital Rings matching Reference Image
      this.createGridRings();

      // 3. Dense Cybernetic Holographic Particle Matrix (Continents strictly formed by glowing dots)
      this.loadLandmassParticles();

      // 4. Offline Vector Country Boundaries from GeoJSON
      this.createCountryBoundaries();

      // 7. Dynamic Markers & Route Group
      this.markerGroup = new THREE.Group();
      this.globeGroup.add(this.markerGroup);

      // 8. Event Listeners for Interaction & Responsive Scaling
      this.setupInteractionListeners();

      // Initial responsive resize calculation
      this.handleResize();
      setTimeout(() => this.handleResize(), 50);
      setTimeout(() => this.handleResize(), 150);
      setTimeout(() => this.handleResize(), 300);
      setTimeout(() => this.handleResize(), 600);

      // 9. Initial Route Display
      this.updateConnectionRoute();
      this.routeBuilt = true;

      // Center initial view to midpoint
      this.centerOnCoordinates(
        (this.originGeo.lat + this.destGeo.lat) / 2,
        (this.originGeo.lon + this.destGeo.lon) / 2,
        false
      );

      // 10. Start Render Loop
      this.clock = new THREE.Clock();
      this.animate = this.animate.bind(this);
      requestAnimationFrame(this.animate);
    }

    createGridRings() {
      const gridGroup = new THREE.Group();
      const r = this.radius;

      // Subtle fine meridian circles
      const gridMat = new THREE.LineBasicMaterial({
        color: 0x83AFCB,
        transparent: true,
        opacity: 0.08
      });

      for (let lon = 0; lon < 180; lon += 45) {
        const circleGeo = new THREE.BufferGeometry();
        const points = [];
        const segments = 64;
        for (let i = 0; i <= segments; i++) {
          const theta = (i / segments) * Math.PI * 2;
          points.push(new THREE.Vector3(Math.sin(theta) * r, Math.cos(theta) * r, 0));
        }
        circleGeo.setFromPoints(points);
        const line = new THREE.Line(circleGeo, gridMat);
        line.rotation.y = lon * (Math.PI / 180);
        gridGroup.add(line);
      }

      // 3 Luminous Tilted Orbital Rings matching Reference Image
      const orbitalConfigs = [
        { radius: r * 1.15, rotX: 0.65, rotY: 0.35, color: 0x83AFCB, opacity: 0.38 },
        { radius: r * 1.22, rotX: -0.55, rotY: 0.82, color: 0xD9954B, opacity: 0.35 },
        { radius: r * 1.30, rotX: 0.28, rotY: -0.65, color: 0x5fa3c9, opacity: 0.30 }
      ];

      orbitalConfigs.forEach(cfg => {
        const ringGeo = new THREE.BufferGeometry();
        const pts = [];
        const segs = 96;
        for (let i = 0; i <= segs; i++) {
          const theta = (i / segs) * Math.PI * 2;
          pts.push(new THREE.Vector3(Math.cos(theta) * cfg.radius, 0, Math.sin(theta) * cfg.radius));
        }
        ringGeo.setFromPoints(pts);

        const ringMat = new THREE.LineBasicMaterial({
          color: cfg.color,
          transparent: true,
          opacity: cfg.opacity
        });
        const orbitLine = new THREE.Line(ringGeo, ringMat);
        orbitLine.rotation.set(cfg.rotX, cfg.rotY, 0);
        gridGroup.add(orbitLine);

        // Orbital particle nodes on the ring
        const nodeGeo = new THREE.SphereGeometry(0.8, 8, 8);
        const nodeMat = new THREE.MeshBasicMaterial({ color: cfg.color, transparent: true, opacity: 0.85 });
        for (let n = 0; n < 3; n++) {
          const nodeMesh = new THREE.Mesh(nodeGeo, nodeMat);
          const angle = (n * (Math.PI * 2 / 3)) + cfg.rotX;
          nodeMesh.position.set(Math.cos(angle) * cfg.radius, 0, Math.sin(angle) * cfg.radius);
          orbitLine.add(nodeMesh);
        }
      });

      this.globeGroup.add(gridGroup);
    }

    createAtmosphere() {
      // Atmospheric Fresnel Rim Light: Razor-sharp glassy 3D spherical depth, 100% transparent center
      const atmosphereGeo = new THREE.SphereGeometry(this.radius * 1.025, 64, 64);
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
            float dotNV = dot(vNormal, viewDir);
            float rim = 1.0 - max(0.0, dotNV);
            // Sharp falloff: 0.0 at center, glowing only at outermost rim
            float intensity = pow(rim, 3.8);

            // Dual-tone blend: Warm Amber (#D9954B) to Deep Cyber Blue (#83AFCB)
            float blend = clamp(vNormal.x * 0.5 + 0.5, 0.0, 1.0);
            vec3 amber = vec3(0.851, 0.584, 0.294);     // #D9954B
            vec3 steelBlue = vec3(0.514, 0.686, 0.796); // #83AFCB
            vec3 rimColor = mix(steelBlue, amber, blend);

            gl_FragColor = vec4(rimColor, intensity * 0.90);
          }
        `,
        blending: THREE.AdditiveBlending,
        side: THREE.FrontSide,
        transparent: true,
        depthWrite: false
      });

      const atmosphereMesh = new THREE.Mesh(atmosphereGeo, atmosphereMat);
      this.scene.add(atmosphereMesh);
    }

    loadLandmassParticles() {
      // True Spherical Fibonacci Sampling + High-Resolution Polygon Rasterization
      // Uniform surface area density worldwide - zero equatorial or polar distortion!
      const w = 1440;
      const h = 720;
      const offCanvas = document.createElement('canvas');
      offCanvas.width = w;
      offCanvas.height = h;
      const ctx = offCanvas.getContext('2d');
      ctx.fillStyle = '#ffffff';

      const drawRing = (ring) => {
        if (!ring || ring.length < 3) return;
        ctx.beginPath();
        for (let i = 0; i < ring.length; i++) {
          const lon = ring[i][0];
          const lat = ring[i][1];
          const x = ((lon + 180) / 360) * w;
          const y = ((90 - lat) / 180) * h;
          if (i === 0) ctx.moveTo(x, y);
          else ctx.lineTo(x, y);
        }
        ctx.closePath();
        ctx.fill();
      };

      if (window.WORLD_GEOJSON && Array.isArray(window.WORLD_GEOJSON.features)) {
        window.WORLD_GEOJSON.features.forEach(feat => {
          const geom = feat.geometry;
          if (!geom) return;
          if (geom.type === 'Polygon') {
            geom.coordinates.forEach(ring => drawRing(ring));
          } else if (geom.type === 'MultiPolygon') {
            geom.coordinates.forEach(poly => poly.forEach(ring => drawRing(ring)));
          }
        });
      }

      const imgData = ctx.getImageData(0, 0, w, h).data;
      const positions = [];
      const colors = [];
      const dotTexture = createCircleTexture();

      // Authentic Night Lights & Atmosphere Palette matching Reference Image
      const cAmberBase   = new THREE.Color(0xEEA050); // Luminous Amber City Light
      const cAmberWarm   = new THREE.Color(0xD97732); // Brand Amber Core
      const cGold        = new THREE.Color(0xFFCC66); // Dense Urban Cluster
      const cBlueBase    = new THREE.Color(0x4B6E88); // Subtle Coastal Rim
      const cBlueDeep    = new THREE.Color(0x243545); // Deep Continent Interior

      const r = this.radius;

      // Spherical Fibonacci Lattice: 26,000 points uniformly distributed over sphere area
      const N = 26000;
      const goldenRatio = (1 + Math.sqrt(5)) / 2;

      for (let i = 0; i < N; i++) {
        const yNorm = 1 - (i / (N - 1)) * 2;
        const radiusAtY = Math.sqrt(Math.max(0, 1 - yNorm * yNorm));
        const phi = 2 * Math.PI * i / goldenRatio;
        const xNorm = radiusAtY * Math.sin(phi);
        const zNorm = radiusAtY * Math.cos(phi);

        // Convert 3D spherical direction to lat/lon for landmass polygon lookup
        const phiAngle = Math.acos(Math.max(-1, Math.min(1, yNorm)));
        const lat = 90 - (phiAngle * 180 / Math.PI);
        if (lat < -72 || lat > 82) continue; // Skip polar extremities

        const lon = Math.atan2(xNorm, zNorm) * (180 / Math.PI);
        const px = Math.min(w - 1, Math.max(0, Math.floor(((lon + 180) / 360) * w)));
        const py = Math.min(h - 1, Math.max(0, Math.floor(((90 - lat) / 180) * h)));
        const index = (py * w + px) * 4;

        // If point falls inside continent polygon, place dot on sphere
        if (imgData[index + 3] > 120) {
          positions.push(xNorm * r, yNorm * r, zNorm * r);

          const rand = Math.random();
          let pColor;
          if (rand < 0.22) {
            pColor = cGold;
          } else if (rand < 0.50) {
            pColor = cAmberBase;
          } else if (rand < 0.70) {
            pColor = cAmberWarm;
          } else if (rand < 0.88) {
            pColor = cBlueBase;
          } else {
            pColor = cBlueDeep;
          }
          colors.push(pColor.r, pColor.g, pColor.b);
        }
      }

      const particleGeo = new THREE.BufferGeometry();
      particleGeo.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));
      particleGeo.setAttribute('color', new THREE.Float32BufferAttribute(colors, 3));

      const particleMat = new THREE.PointsMaterial({
        size: 1.85,
        map: dotTexture,
        vertexColors: true,
        transparent: true,
        opacity: 0.90,
        blending: THREE.NormalBlending, // NormalBlending prevents washed-out white blowouts!
        depthWrite: false
      });

      const particles = new THREE.Points(particleGeo, particleMat);
      this.globeGroup.add(particles);
    }

    createCountryBoundaries() {
      if (!window.WORLD_GEOJSON || !Array.isArray(window.WORLD_GEOJSON.features)) {
        return;
      }

      this.countryBordersGroup = new THREE.Group();
      this.defaultBorderMat = new THREE.LineBasicMaterial({
        color: 0x3d6482,
        transparent: true,
        opacity: 0.38,
        depthWrite: false,
        depthTest: true
      });

      this.activeBorderMat = new THREE.LineBasicMaterial({
        color: 0xD97732,
        transparent: true,
        opacity: 0.92,
        depthWrite: false
      });

      const r = this.radius * 1.002;

      window.WORLD_GEOJSON.features.forEach(feat => {
        const code = (feat.properties && feat.properties.code || '').toUpperCase();
        const geom = feat.geometry;
        if (!geom) return;

        const rings = [];
        if (geom.type === 'Polygon') {
          rings.push(...geom.coordinates);
        } else if (geom.type === 'MultiPolygon') {
          geom.coordinates.forEach(poly => rings.push(...poly));
        }

        const pts = [];
        rings.forEach(ring => {
          if (!ring || ring.length < 2) return;
          for (let i = 0; i < ring.length - 1; i++) {
            const p1 = latLonToVector3(ring[i][1], ring[i][0], r);
            const p2 = latLonToVector3(ring[i + 1][1], ring[i + 1][0], r);
            pts.push(p1, p2);
          }
        });

        if (pts.length > 0) {
          const borderGeo = new THREE.BufferGeometry().setFromPoints(pts);
          const lineMesh = new THREE.LineSegments(borderGeo, this.defaultBorderMat);
          this.countryBordersGroup.add(lineMesh);
          if (code) {
            this.countryMeshes[code] = lineMesh;
          }
        }
      });

      this.globeGroup.add(this.countryBordersGroup);
      if (this.destGeo && this.destGeo.code) {
        this.highlightCountryBorders([this.destGeo.code]);
      }
    }

    updateConnectionRoute() {
      // Clear existing markers and arc
      while (this.markerGroup.children.length > 0) {
        const obj = this.markerGroup.children[0];
        this.markerGroup.remove(obj);
        if (obj.geometry) obj.geometry.dispose();
        if (obj.material) {
          if (Array.isArray(obj.material)) obj.material.forEach(m => m.dispose());
          else obj.material.dispose();
        }
      }

      this.destRing = null;
      this.photon = null;
      this.arcCurve = null;

      const activeCountries = [];
      const surfaceR = this.radius;
      const h0 = 1.0; // Baseline surface elevation: prevents any z-fighting with country borders

      // 1. Destination Marker (Always rendered if destination is configured)
      if (this.destGeo) {
        const uDest = latLonToVector3(this.destGeo.lat, this.destGeo.lon, 1.0).normalize();
        const vDest = uDest.clone().multiplyScalar(surfaceR + h0);

        // Destination Sphere Pin
        const destDotGeo = new THREE.SphereGeometry(1.6, 16, 16);
        const destDotMat = new THREE.MeshBasicMaterial({ color: 0xD9954B });
        const destDot = new THREE.Mesh(destDotGeo, destDotMat);
        destDot.position.copy(vDest);
        this.markerGroup.add(destDot);

        // Destination Pulse Ring
        const destRingGeo = new THREE.RingGeometry(2.0, 3.8, 32);
        const destRingMat = new THREE.MeshBasicMaterial({
          color: 0xD9954B,
          side: THREE.DoubleSide,
          transparent: true,
          opacity: 0.95
        });
        this.destRing = new THREE.Mesh(destRingGeo, destRingMat);
        this.destRing.position.copy(uDest.clone().multiplyScalar(surfaceR + h0 + 0.15));
        this.destRing.lookAt(uDest.clone().multiplyScalar(surfaceR * 2));
        this.markerGroup.add(this.destRing);

        if (this.destGeo.code) activeCountries.push(this.destGeo.code);
      }

      // 2. Origin Marker & Genuine Spherical Great-Circle Slerp Route
      if (this.originGeo && this.destGeo) {
        const uOrigin = latLonToVector3(this.originGeo.lat, this.originGeo.lon, 1.0).normalize();
        const uDest = latLonToVector3(this.destGeo.lat, this.destGeo.lon, 1.0).normalize();

        const vOrigin = uOrigin.clone().multiplyScalar(surfaceR + h0);
        const vDest = uDest.clone().multiplyScalar(surfaceR + h0);

        // Origin Sphere Pin
        const originDotGeo = new THREE.SphereGeometry(1.4, 16, 16);
        const originDotMat = new THREE.MeshBasicMaterial({ color: 0xffffff });
        const originDot = new THREE.Mesh(originDotGeo, originDotMat);
        originDot.position.copy(vOrigin);
        this.markerGroup.add(originDot);

        // Origin Halo Ring
        const originRingGeo = new THREE.RingGeometry(1.6, 2.8, 32);
        const originRingMat = new THREE.MeshBasicMaterial({
          color: 0x83AFCB,
          side: THREE.DoubleSide,
          transparent: true,
          opacity: 0.8
        });
        const originRing = new THREE.Mesh(originRingGeo, originRingMat);
        originRing.position.copy(uOrigin.clone().multiplyScalar(surfaceR + h0 + 0.15));
        originRing.lookAt(uOrigin.clone().multiplyScalar(surfaceR * 2));
        this.markerGroup.add(originRing);

        if (this.originGeo.code) activeCountries.push(this.originGeo.code);

        // Angular distance Omega between origin and destination
        let dot = uOrigin.dot(uDest);
        dot = Math.max(-1.0, Math.min(1.0, dot));
        const omega = Math.acos(dot);

        // Only draw connection route if endpoints are distinct
        if (omega > 0.02) {
          const sinOmega = Math.sin(omega);
          const numSteps = 80;
          const arcPoints = [];

          // Smooth apex elevation based on angular distance (minimum 6, maximum 22 units above surface)
          const peakAltitude = Math.min(22, Math.max(6, omega * 13));

          for (let i = 0; i <= numSteps; i++) {
            const t = i / numSteps;
            let u_t;
            if (sinOmega < 1e-5) {
              u_t = uOrigin.clone().lerp(uDest, t).normalize();
            } else {
              // Spherical Linear Interpolation (Slerp) along great-circle
              const wOrigin = Math.sin((1.0 - t) * omega) / sinOmega;
              const wDest = Math.sin(t * omega) / sinOmega;
              u_t = uOrigin.clone().multiplyScalar(wOrigin).add(uDest.clone().multiplyScalar(wDest)).normalize();
            }

            // Smooth sine elevation curve: starts at h0, peaks at midpoint, returns cleanly to h0
            const alt = h0 + peakAltitude * Math.sin(Math.PI * t);
            const pt = u_t.multiplyScalar(surfaceR + alt);
            arcPoints.push(pt);
          }

          // Smooth 3D Spline Curve
          const curve = new THREE.CatmullRomCurve3(arcPoints);
          this.arcCurve = curve;

          // Continuous 3D Tube Geometry (radius 0.45, 80 tubular segments)
          // Guarantees solid, continuous visibility on all WebGL/DirectX implementations without 1px line aliasing
          const tubeGeo = new THREE.TubeGeometry(curve, 80, 0.45, 8, false);
          const tubeMat = new THREE.MeshBasicMaterial({
            color: 0xD9954B, // Smoked Glass Brand Amber
            transparent: true,
            opacity: 0.95,
            depthTest: true
          });
          const tubeMesh = new THREE.Mesh(tubeGeo, tubeMat);
          this.markerGroup.add(tubeMesh);

          // Inner luminous core line for crisp edge definition
          const coreLineGeo = new THREE.BufferGeometry().setFromPoints(arcPoints);
          const coreLineMat = new THREE.LineBasicMaterial({
            color: 0xffffff,
            transparent: true,
            opacity: 0.85
          });
          const coreLine = new THREE.Line(coreLineGeo, coreLineMat);
          this.markerGroup.add(coreLine);

          // Animated Photon traveling along the curve
          const photonGeo = new THREE.SphereGeometry(1.3, 16, 16);
          const photonMat = new THREE.MeshBasicMaterial({
            color: 0xF2BA7E,
            blending: THREE.AdditiveBlending
          });
          this.photon = new THREE.Mesh(photonGeo, photonMat);
          this.photon.position.copy(arcPoints[0]);
          this.markerGroup.add(this.photon);
        }
      }

      // Highlight active borders (clean particle-only aesthetic)
      if (typeof this.highlightCountryBorders === 'function') {
        this.highlightCountryBorders(activeCountries);
      }

      // Update Floating Exit Badge
      this.updateFloatingExitBadge();
    }

    highlightCountryBorders(activeCountries) {
      if (!this.countryMeshes || !this.activeBorderMat || !this.defaultBorderMat) return;
      const activeSet = new Set((activeCountries || []).map(c => String(c).toUpperCase()));

      Object.keys(this.countryMeshes).forEach(code => {
        const mesh = this.countryMeshes[code];
        if (!mesh) return;
        if (activeSet.has(code)) {
          mesh.material = this.activeBorderMat;
        } else {
          mesh.material = this.defaultBorderMat;
        }
      });
    }

    centerOnCoordinates(lat, lon, animate = true) {
      const targetY = -(lon * (Math.PI / 180));
      const targetX = Math.max(-0.65, Math.min(0.65, (lat - 8) * (Math.PI / 180) * 0.5));

      if (animate) {
        // If already approximately centered, avoid tiny jittery transitions
        const diffY = Math.abs(targetY - this.currentRotation.y);
        const diffX = Math.abs(targetX - this.currentRotation.x);
        if (diffY < 0.005 && diffX < 0.005) {
          return;
        }

        this.startRotation = { x: this.currentRotation.x, y: this.currentRotation.y };
        this.targetRotation = { x: targetX, y: targetY };
        this.transitionStartTime = performance.now();
        this.transitionDuration = 900; // ms
        this.isTransitioning = true;
      } else {
        this.isTransitioning = false;
        this.targetRotation.x = targetX;
        this.targetRotation.y = targetY;
        this.currentRotation.x = targetX;
        this.currentRotation.y = targetY;
        this.globeGroup.rotation.x = targetX;
        this.globeGroup.rotation.y = targetY;
      }
    }

    updateState(status) {
      if (!status) return;

      const newConnState = status.state || 'disconnected';
      const stateChanged = (this.connectionState !== newConnState);
      this.connectionState = newConnState;
      this.connected = (newConnState === 'connected');
      this.isConnecting = (newConnState === 'connecting');

      if (status.exit_latency && status.exit_latency > 0) {
        this.exitLatency = status.exit_latency;
      }

      // 1. Resolve Origin Coordinates
      let originCode = (status.current_country || '').toUpperCase();
      if (!originCode) originCode = 'IR'; // Baseline fallback

      let newOriginGeo = null;
      if (COUNTRY_CENTROIDS[originCode]) {
        const c = COUNTRY_CENTROIDS[originCode];
        newOriginGeo = {
          lat: c.lat,
          lon: c.lon,
          name: c.name,
          city: c.city,
          code: originCode
        };
      }

      // 2. Resolve Destination Coordinates
      let newDestGeo = null;
      if (status.server) {
        const popCode = (status.server.city_code || status.server.city || '').toUpperCase();
        if (GEO_COORDINATES[popCode]) {
          newDestGeo = { ...GEO_COORDINATES[popCode] };
        } else if (status.server.country_code && COUNTRY_CENTROIDS[status.server.country_code.toUpperCase()]) {
          const c = COUNTRY_CENTROIDS[status.server.country_code.toUpperCase()];
          newDestGeo = {
            lat: c.lat,
            lon: c.lon,
            city: status.server.city_name || c.city,
            country: c.name,
            code: status.server.country_code.toUpperCase()
          };
        }
      }

      if (!newDestGeo && status.exit_country && COUNTRY_CENTROIDS[status.exit_country.toUpperCase()]) {
        const c = COUNTRY_CENTROIDS[status.exit_country.toUpperCase()];
        newDestGeo = {
          lat: c.lat,
          lon: c.lon,
          city: c.city,
          country: c.name,
          code: status.exit_country.toUpperCase()
        };
      }

      // Default destination if not set
      if (!newDestGeo) {
        newDestGeo = { lat: -26.1392, lon: 28.2460, city: 'Johannesburg', country: 'South Africa', code: 'ZA' };
      }

      const originKey = newOriginGeo ? newOriginGeo.code : '';
      const destKey = newDestGeo ? `${newDestGeo.code}-${newDestGeo.city}` : '';

      const originChanged = (this.cachedOriginKey !== originKey);
      const destChanged = (this.cachedDestKey !== destKey);

      this.originGeo = newOriginGeo;
      this.destGeo = newDestGeo;

      // CRITICAL FOR BUG 1:
      // Only rebuild meshes and transition camera if meaningful event happened!
      // Telemetry updates (ping, bandwidth, quota) do NOT rebuild or reset the globe!
      if (originChanged || destChanged || stateChanged || !this.routeBuilt) {
        this.cachedOriginKey = originKey;
        this.cachedDestKey = destKey;
        this.cachedConnState = newConnState;

        this.updateConnectionRoute();
        this.routeBuilt = true;

        // Smooth transition to target when server changed or connection initiated
        if (destChanged || stateChanged) {
          if (this.originGeo && this.destGeo) {
            this.centerOnCoordinates(
              (this.originGeo.lat + this.destGeo.lat) / 2,
              (this.originGeo.lon + this.destGeo.lon) / 2,
              true
            );
          } else if (this.destGeo) {
            this.centerOnCoordinates(this.destGeo.lat, this.destGeo.lon, true);
          }
        }
      }

      this.updateFloatingExitBadge();
    }

    updateFloatingExitBadge() {
      const flagEl = document.getElementById('badge-flag');
      const codeEl = document.getElementById('badge-code');
      const cityEl = document.getElementById('badge-city');
      const latEl = document.getElementById('badge-latency');
      const badgeWrap = document.getElementById('globe-exit-badge');

      if (!badgeWrap || !this.destGeo) return;

      const cityName = (this.destGeo.city || this.destGeo.name || 'JOHANNESBURG').toUpperCase();
      const countryCode = (this.destGeo.code || 'ZA').toUpperCase();

      if (cityEl) cityEl.textContent = `${cityName}, ${countryCode}`;
      if (codeEl) codeEl.textContent = countryCode;

      if (flagEl) {
        if (typeof getFlag === 'function') {
          flagEl.innerHTML = getFlag(countryCode);
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
        this.isTransitioning = false; // User interrupt cancels automated transition immediately
        this.previousMousePosition = {
          x: e.clientX || (e.touches && e.touches[0].clientX) || 0,
          y: e.clientY || (e.touches && e.touches[0].clientY) || 0
        };
      };

      const onPointerMove = (e) => {
        if (!this.isInteracting) return;

        const clientX = e.clientX || (e.touches && e.touches[0].clientX) || 0;
        const clientY = e.clientY || (e.touches && e.touches[0].clientY) || 0;

        const deltaX = clientX - this.previousMousePosition.x;
        const deltaY = clientY - this.previousMousePosition.y;

        this.targetRotation.y += deltaX * 0.0055;
        this.targetRotation.x += deltaY * 0.0055;

        // Clamp vertical pitch to avoid flipping
        this.targetRotation.x = Math.max(-0.85, Math.min(0.85, this.targetRotation.x));

        this.currentRotation.x = this.targetRotation.x;
        this.currentRotation.y = this.targetRotation.y;
        this.globeGroup.rotation.x = this.currentRotation.x;
        this.globeGroup.rotation.y = this.currentRotation.y;

        this.previousMousePosition = { x: clientX, y: clientY };
      };

      const onPointerUp = () => {
        this.isInteracting = false;
        // CRITICAL FOR BUG 1:
        // Do NOT restart auto-spin when user releases! Globe remains in exact position left by user.
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
      const width = this.container.clientWidth;
      const height = this.container.clientHeight;
      if (width <= 0 || height <= 0) return;

      const aspect = width / height;
      this.camera.aspect = aspect;

      const vFOV = (this.camera.fov * Math.PI) / 180;
      const visualRadius = this.radius * 1.25;

      const dVert = (visualRadius * 1.22) / Math.tan(vFOV / 2);
      const dHoriz = dVert / Math.max(aspect, 0.5);

      const targetZ = Math.max(dVert, dHoriz) * 0.86;
      this.camera.position.set(0, 0, targetZ);
      this.camera.updateProjectionMatrix();

      this.renderer.setSize(width, height);
    }

    animate() {
      requestAnimationFrame(this.animate);

      // Performance: completely pause WebGL draw passes when globe is not visible
      if (this.container && this.container.offsetParent === null) {
        return;
      }

      const now = performance.now();
      const elapsedTime = this.clock.getElapsedTime();

      // 1. Event-Driven Smooth Finite Transitions
      if (this.isTransitioning) {
        const elapsed = now - this.transitionStartTime;
        const progress = Math.min(1.0, elapsed / this.transitionDuration);
        const ease = 1 - Math.pow(1 - progress, 3); // Smooth cubic ease-out

        this.currentRotation.x = this.startRotation.x + (this.targetRotation.x - this.startRotation.x) * ease;
        this.currentRotation.y = this.startRotation.y + (this.targetRotation.y - this.startRotation.y) * ease;

        if (progress >= 1.0) {
          this.isTransitioning = false;
          this.currentRotation.x = this.targetRotation.x;
          this.currentRotation.y = this.targetRotation.y;
        }

        this.globeGroup.rotation.x = this.currentRotation.x;
        this.globeGroup.rotation.y = this.currentRotation.y;
      }
      // When idle (not transitioning and not dragging): ZERO rotation change! Pure stillness.

      // 2. Pulse destination ring only when connecting or connected
      if (this.destRing && (this.connected || this.isConnecting)) {
        const scale = 1 + Math.sin(elapsedTime * 3.5) * 0.08;
        this.destRing.scale.set(scale, scale, 1);
      }

      // 3. Animate photon along connection arc only when connected or connecting
      if (this.photon && this.arcCurve && (this.connected || this.isConnecting)) {
        this.photon.visible = true;
        const t = (elapsedTime * 0.35) % 1.0;
        const pos = this.arcCurve.getPoint(t);
        this.photon.position.copy(pos);
      } else if (this.photon) {
        this.photon.visible = false;
      }

      this.renderer.render(this.scene, this.camera);
    }
  }

  window.FyrnoxGlobe = FyrnoxGlobe;
  window.GEO_COORDINATES = GEO_COORDINATES;
  window.COUNTRY_CENTROIDS = COUNTRY_CENTROIDS;

})(window);
