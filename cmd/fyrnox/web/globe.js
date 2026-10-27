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
      this.originGeo = { lat: 32.4279, lon: 53.6880, city: 'Tehran', country: 'Iran', code: 'IR' };
      this.destGeo = { lat: 50.1109, lon: 8.6821, city: 'Frankfurt', country: 'Germany', code: 'DE' };
      this.exitLatency = 0;

      this.cachedOriginKey = '';
      this.cachedDestKey = '';
      this.cachedConnState = '';
      this.routeBuilt = false;

      this.countryMeshes = {};

      this.init();
    }

    init() {
      const width = this.container.clientWidth || 550;
      const height = this.container.clientHeight || 400;

      // 1. Scene, Camera, Renderer
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

      // 2. Base Sphere (Deep Obsidian Navy)
      const baseGeo = new THREE.SphereGeometry(this.radius * 0.985, 48, 48);
      const baseMat = new THREE.MeshBasicMaterial({
        color: 0x060c18,
        transparent: true,
        opacity: 0.92
      });
      this.baseSphere = new THREE.Mesh(baseGeo, baseMat);
      this.globeGroup.add(this.baseSphere);

      // 3. Latitude & Longitude Wireframe Grid
      this.createGridRings();

      // 4. Atmospheric Fresnel Rim Glow
      this.createAtmosphere();

      // 5. Natural Earth Vector Country Boundaries
      this.buildCountryBoundaries();

      // 6. Particle Landmasses (NASA landmask)
      this.loadLandmassParticles();

      // 7. Dynamic Markers & Route Group
      this.markerGroup = new THREE.Group();
      this.globeGroup.add(this.markerGroup);

      // 8. Event Listeners for Interaction & Responsive Scaling
      this.setupInteractionListeners();

      // Initial responsive resize calculation
      this.handleResize();

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
      const r = this.radius * 1.002;

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
        const code = (feature.properties && (feature.properties.ISO_A2 || feature.properties.ADM0_A3 || '')) || '';
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
          if (lat < -78 || lat > 82) continue;

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

        const particles = new THREE.Points(particleGeo, particleMat);
        this.globeGroup.add(particles);
      };

      img.onerror = () => {
        console.warn('[FyrnoxGlobe] Could not load earth_landmask.jpg, relying on vector boundaries');
      };
    }

    /**
     * Rebuilds connection route geometry using true Spherical Great-Circle Interpolation (Slerp)
     * and a continuous 3D Tube mesh to guarantee zero gaps, zero clipping, and exact marker anchoring.
     */
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
        const destDotMat = new THREE.MeshBasicMaterial({ color: 0x00ff9d });
        const destDot = new THREE.Mesh(destDotGeo, destDotMat);
        destDot.position.copy(vDest);
        this.markerGroup.add(destDot);

        // Destination Pulse Ring
        const destRingGeo = new THREE.RingGeometry(2.0, 3.8, 32);
        const destRingMat = new THREE.MeshBasicMaterial({
          color: 0x00e676,
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
          color: 0x00d2ff,
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
            color: 0xffb000, // Vibrant warm gold
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
            color: 0xffe066,
            blending: THREE.AdditiveBlending
          });
          this.photon = new THREE.Mesh(photonGeo, photonMat);
          this.photon.position.copy(arcPoints[0]);
          this.markerGroup.add(this.photon);
        }
      }

      // Highlight active borders
      this.highlightCountryBorders(activeCountries);

      // Update Floating Exit Badge
      this.updateFloatingExitBadge();
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
        newDestGeo = { lat: 50.1109, lon: 8.6821, city: 'Frankfurt', country: 'Germany', code: 'DE' };
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
      const cityEl = document.getElementById('badge-city');
      const latEl = document.getElementById('badge-latency');
      const badgeWrap = document.getElementById('globe-exit-badge');

      if (!badgeWrap || !this.destGeo) return;

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
      const width = this.container.clientWidth || 550;
      const height = this.container.clientHeight || 400;

      const aspect = width / height;
      this.camera.aspect = aspect;

      const vFOV = (this.camera.fov * Math.PI) / 180;
      const visualRadius = this.radius * 1.25;

      const dVert = (visualRadius * 1.22) / Math.tan(vFOV / 2);
      const dHoriz = dVert / Math.max(aspect, 0.5);

      const targetZ = Math.max(dVert, dHoriz);
      this.camera.position.set(0, 0, targetZ);
      this.camera.updateProjectionMatrix();

      this.renderer.setSize(width, height);
    }

    animate() {
      requestAnimationFrame(this.animate);

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
