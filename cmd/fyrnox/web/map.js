/**
 * FYRNOX Interactive Desktop World Map
 * High-performance vector map with pan/zoom, server nodes, and live origin-to-server connection visualization.
 */

(function (window) {
  'use strict';

  // Geographic coordinates database for known server cities and countries
  const CITY_COORDINATES = {
    // Official Fastly MASQUE POPs (ICAO Codes & Regions)
    US: { lat: 37.0902, lon: -95.7129, city: 'Anycast Edge', country: 'United States', countryCode: 'US' },
    Q30: { lat: 38.8951, lon: -77.0364, city: 'Fastly Anycast', country: 'United States', countryCode: 'US' },
    CA: { lat: 43.6532, lon: -79.3832, city: 'Toronto', country: 'Canada', countryCode: 'CA' },
    MUC: { lat: 48.3538, lon: 11.7861, city: 'Munich', country: 'Germany', countryCode: 'DE' },
    EGLC: { lat: 51.5053, lon: 0.0553, city: 'London', country: 'United Kingdom', countryCode: 'GB' },
    LFPB: { lat: 48.9694, lon: 2.4414, city: 'Paris', country: 'France', countryCode: 'FR' },
    EHRD: { lat: 51.9569, lon: 4.4372, city: 'Rotterdam', country: 'Netherlands', countryCode: 'NL' },
    RJTF: { lat: 35.6717, lon: 139.5281, city: 'Tokyo', country: 'Japan', countryCode: 'JP' },
    WSAT: { lat: 1.4169, lon: 103.8678, city: 'Singapore', country: 'Singapore', countryCode: 'SG' },
    YSBK: { lat: -33.9242, lon: 150.9883, city: 'Sydney', country: 'Australia', countryCode: 'AU' },
    ESSB: { lat: 59.3544, lon: 17.9417, city: 'Stockholm', country: 'Sweden', countryCode: 'SE' },
    HEL: { lat: 60.3172, lon: 24.9633, city: 'Helsinki', country: 'Finland', countryCode: 'FI' },
    OSL: { lat: 60.1976, lon: 11.1004, city: 'Oslo', country: 'Norway', countryCode: 'NO' },
    CPH: { lat: 55.6180, lon: 12.6560, city: 'Copenhagen', country: 'Denmark', countryCode: 'DK' },
    DUB: { lat: 53.4213, lon: -6.2701, city: 'Dublin', country: 'Ireland', countryCode: 'IE' },
    MXP: { lat: 45.6301, lon: 8.7231, city: 'Milan', country: 'Italy', countryCode: 'IT' },
    LETO: { lat: 40.4800, lon: -3.4567, city: 'Madrid', country: 'Spain', countryCode: 'ES' },
    VIE: { lat: 48.1103, lon: 16.5697, city: 'Vienna', country: 'Austria', countryCode: 'AT' },
    BRU: { lat: 50.9014, lon: 4.4844, city: 'Brussels', country: 'Belgium', countryCode: 'BE' },
    LIS: { lat: 38.7742, lon: -9.1342, city: 'Lisbon', country: 'Portugal', countryCode: 'PT' },
    FAOR: { lat: -26.1392, lon: 28.2460, city: 'Johannesburg', country: 'South Africa', countryCode: 'ZA' },
    SPJC: { lat: -12.0219, lon: -77.1143, city: 'Lima', country: 'Peru', countryCode: 'PE' },
    SBSP: { lat: -23.6261, lon: -46.6564, city: 'Sao Paulo', country: 'Brazil', countryCode: 'BR' },
    ICN: { lat: 37.4602, lon: 126.4407, city: 'Seoul', country: 'Korea', countryCode: 'KR' },
    ACC: { lat: 5.6052, lon: -0.1668, city: 'Accra', country: 'Ghana', countryCode: 'GH' },
    EZE: { lat: -34.8222, lon: -58.5358, city: 'Buenos Aires', country: 'Argentina', countryCode: 'AR' },
    MNL: { lat: 14.5086, lon: 121.0194, city: 'Manila', country: 'Philippines', countryCode: 'PH' },
    AKL: { lat: -37.0082, lon: 174.7850, city: 'Auckland', country: 'New Zealand', countryCode: 'NZ' },
    SCL: { lat: -33.3930, lon: -70.7858, city: 'Santiago', country: 'Chile', countryCode: 'CL' },
    KUL: { lat: 2.7456, lon: 101.7099, city: 'Kuala Lumpur', country: 'Malaysia', countryCode: 'MY' },
    MMQT: { lat: 20.6173, lon: -100.3186, city: 'Queretaro', country: 'Mexico', countryCode: 'MX' },
    SKBO: { lat: 4.7016, lon: -74.1469, city: 'Bogota', country: 'Colombia', countryCode: 'CO' },
    BKK: { lat: 13.6900, lon: 100.7501, city: 'Bangkok', country: 'Thailand', countryCode: 'TH' },
    SOF: { lat: 42.6951, lon: 23.4061, city: 'Sofia', country: 'Bulgaria', countryCode: 'BG' },

    // Legacy airport code aliases
    JFK: { lat: 40.7128, lon: -74.0060, city: 'New York', country: 'United States', countryCode: 'US' },
    FRA: { lat: 50.1109, lon: 8.6821, city: 'Frankfurt', country: 'Germany', countryCode: 'DE' },
    AMS: { lat: 52.3676, lon: 4.9041, city: 'Amsterdam', country: 'Netherlands', countryCode: 'NL' },
    LHR: { lat: 51.5074, lon: -0.1278, city: 'London', country: 'United Kingdom', countryCode: 'GB' },
    CDG: { lat: 48.8566, lon: 2.3522, city: 'Paris', country: 'France', countryCode: 'FR' },
    SIN: { lat: 1.3521, lon: 103.8198, city: 'Singapore', country: 'Singapore', countryCode: 'SG' },
    SYD: { lat: -33.8688, lon: 151.2093, city: 'Sydney', country: 'Australia', countryCode: 'AU' }
  };

  // Country Centroid Coordinates (for Origin Detection and Country Mapping)
  const COUNTRY_CENTROIDS = {
    IR: { lat: 32.4279, lon: 53.6880, name: 'Iran' },
    US: { lat: 38.0000, lon: -97.0000, name: 'United States' },
    DE: { lat: 51.1657, lon: 10.4515, name: 'Germany' },
    NL: { lat: 52.1326, lon: 5.2913, name: 'Netherlands' },
    GB: { lat: 54.3781, lon: -2.4360, name: 'United Kingdom' },
    FR: { lat: 46.2276, lon: 2.2137, name: 'France' },
    JP: { lat: 36.2048, lon: 138.2529, name: 'Japan' },
    SG: { lat: 1.3521, lon: 103.8198, name: 'Singapore' },
    CA: { lat: 56.1304, lon: -106.3468, name: 'Canada' },
    CH: { lat: 46.8182, lon: 8.2275, name: 'Switzerland' },
    SE: { lat: 60.1282, lon: 18.6435, name: 'Sweden' },
    FI: { lat: 61.9241, lon: 25.7482, name: 'Finland' },
    NO: { lat: 60.4720, lon: 8.4689, name: 'Norway' },
    DK: { lat: 56.2639, lon: 9.5018, name: 'Denmark' },
    IE: { lat: 53.1424, lon: -7.6921, name: 'Ireland' },
    IT: { lat: 41.8719, lon: 12.5674, name: 'Italy' },
    ES: { lat: 40.4637, lon: -3.7492, name: 'Spain' },
    AT: { lat: 47.5162, lon: 14.5501, name: 'Austria' },
    BE: { lat: 50.5039, lon: 4.4699, name: 'Belgium' },
    PL: { lat: 51.9194, lon: 19.1451, name: 'Poland' },
    PT: { lat: 39.3999, lon: -8.2245, name: 'Portugal' },
    AU: { lat: -25.2744, lon: 133.7751, name: 'Australia' },
    NZ: { lat: -40.9006, lon: 174.8860, name: 'New Zealand' },
    BR: { lat: -14.2350, lon: -51.9253, name: 'Brazil' },
    HK: { lat: 22.3193, lon: 114.1694, name: 'Hong Kong' },
    KR: { lat: 35.9078, lon: 127.7669, name: 'South Korea' },
    TR: { lat: 38.9637, lon: 35.2433, name: 'Turkey' },
    AE: { lat: 23.4241, lon: 53.8478, name: 'United Arab Emirates' },
    IN: { lat: 20.5937, lon: 78.9629, name: 'India' },
    ZA: { lat: -30.5595, lon: 22.9375, name: 'South Africa' },
    RU: { lat: 61.5240, lon: 105.3188, name: 'Russia' },
    CN: { lat: 35.8617, lon: 104.1954, name: 'China' },
    MX: { lat: 23.6345, lon: -102.5528, name: 'Mexico' },
    SA: { lat: 23.8859, lon: 45.0792, name: 'Saudi Arabia' },
    IQ: { lat: 33.2232, lon: 43.6793, name: 'Iraq' },
    AF: { lat: 33.9391, lon: 67.7100, name: 'Afghanistan' },
    PK: { lat: 30.3753, lon: 69.3451, name: 'Pakistan' },
    EG: { lat: 26.8206, lon: 30.8025, name: 'Egypt' }
  };

  class FyrnoxMap {
    constructor(containerId, options = {}) {
      this.container = document.getElementById(containerId);
      if (!this.container) {
        console.warn(`Map container #${containerId} not found`);
        return;
      }

      this.options = Object.assign({
        width: 1000,
        height: 520,
        onServerSelect: null
      }, options);

      this.width = this.options.width;
      this.height = this.options.height;

      // View state for pan & zoom
      this.viewBox = { x: 0, y: 0, width: this.width, height: this.height };
      this.defaultViewBox = { x: 0, y: 0, width: this.width, height: this.height };
      this.isDragging = false;
      this.dragStart = { x: 0, y: 0 };
      this.zoomLevel = 1;

      // Map state
      this.servers = [];
      this.selectedServer = null;
      this.originLocation = null;
      this.connectionState = 'disconnected'; // 'disconnected', 'connecting', 'connected', 'error'
      this.animFrameId = null;
      this.streamParticles = [];

      this.init();
    }

    init() {
      this.container.innerHTML = '';
      this.container.classList.add('fyrnox-map-wrapper');

      // Create SVG Element
      this.svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
      this.svg.setAttribute('viewBox', `0 0 ${this.width} ${this.height}`);
      this.svg.setAttribute('preserveAspectRatio', 'xMidYMid meet');
      this.svg.classList.add('fyrnox-world-svg');

      // SVG Definitions (Filters, Gradients, Markers)
      const defs = document.createElementNS('http://www.w3.org/2000/svg', 'defs');
      defs.innerHTML = `
        <!-- High-tech cyan glow filter -->
        <filter id="fyrnox-glow" x="-30%" y="-30%" width="160%" height="160%">
          <feGaussianBlur stdDeviation="4" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>

        <filter id="fyrnox-intense-glow" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="6" result="blur1" />
          <feGaussianBlur stdDeviation="2" result="blur2" />
          <feMerge>
            <feMergeNode in="blur1" />
            <feMergeNode in="blur2" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>

        <!-- Arc Stream Gradient -->
        <linearGradient id="arc-stream-grad" x1="0%" y1="0%" x2="100%" y2="0%">
          <stop offset="0%" stop-color="#00F2FE" stop-opacity="0.3" />
          <stop offset="50%" stop-color="#00F2FE" stop-opacity="1" />
          <stop offset="100%" stop-color="#00E676" stop-opacity="0.9" />
        </linearGradient>

        <!-- Particle Gradient -->
        <radialGradient id="particle-grad">
          <stop offset="0%" stop-color="#FFFFFF" stop-opacity="1" />
          <stop offset="60%" stop-color="#00F2FE" stop-opacity="0.9" />
          <stop offset="100%" stop-color="#00F2FE" stop-opacity="0" />
        </radialGradient>
      `;
      this.svg.appendChild(defs);

      // Map Layer Groups
      this.layers = {
        grid: document.createElementNS('http://www.w3.org/2000/svg', 'g'),
        land: document.createElementNS('http://www.w3.org/2000/svg', 'g'),
        labels: document.createElementNS('http://www.w3.org/2000/svg', 'g'),
        arcs: document.createElementNS('http://www.w3.org/2000/svg', 'g'),
        servers: document.createElementNS('http://www.w3.org/2000/svg', 'g'),
        nodes: document.createElementNS('http://www.w3.org/2000/svg', 'g'),
        particles: document.createElementNS('http://www.w3.org/2000/svg', 'g')
      };

      this.layers.grid.classList.add('map-grid-layer');
      this.layers.land.classList.add('map-land-layer');
      this.layers.labels.classList.add('map-labels-layer');
      this.layers.arcs.classList.add('map-arcs-layer');
      this.layers.servers.classList.add('map-servers-layer');
      this.layers.nodes.classList.add('map-nodes-layer');
      this.layers.particles.classList.add('map-particles-layer');

      Object.values(this.layers).forEach(layer => this.svg.appendChild(layer));
      this.container.appendChild(this.svg);

      // Render Base World Map
      this.renderGrid();
      this.renderWorldLandmass();
      this.renderGeographicLabels();

      // UI Controls (Zoom in, Zoom out, Recenter)
      this.createMapControls();

      // Tooltip Element
      this.tooltip = document.createElement('div');
      this.tooltip.classList.add('fyrnox-map-tooltip');
      this.tooltip.style.display = 'none';
      this.container.appendChild(this.tooltip);

      // Bind Mouse / Touch Events for Pan & Zoom
      this.bindInteractionEvents();

      // Start Particle Loop
      this.startParticleAnimation();
    }

    // Convert (Latitude, Longitude) to SVG (X, Y) using Miller / Equirectangular Projection
    project(lat, lon) {
      // Clamped lon (-180 to 180)
      const x = ((lon + 180) / 360) * this.width;
      // Latitude to Y: Mercator clamped to avoid infinite poles
      const clampedLat = Math.max(-75, Math.min(82, lat));
      const latRad = (clampedLat * Math.PI) / 180;
      // Mercator formula normalized
      const mercN = Math.log(Math.tan(Math.PI / 4 + latRad / 2));
      const y = this.height / 2 - (this.width * mercN) / (2 * Math.PI) * 0.9 + 25;
      return { x: Math.round(x * 10) / 10, y: Math.round(y * 10) / 10 };
    }

    renderGrid() {
      // Subtle longitude & latitude lines
      let gridSvg = '';
      for (let lon = -150; lon <= 180; lon += 30) {
        const p1 = this.project(80, lon);
        const p2 = this.project(-70, lon);
        gridSvg += `<line x1="${p1.x}" y1="0" x2="${p2.x}" y2="${this.height}" stroke="rgba(255,255,255,0.03)" stroke-width="1" stroke-dasharray="2,6" />`;
      }
      for (let lat = -60; lat <= 80; lat += 20) {
        const p = this.project(lat, 0);
        gridSvg += `<line x1="0" y1="${p.y}" x2="${this.width}" y2="${p.y}" stroke="rgba(255,255,255,0.03)" stroke-width="1" stroke-dasharray="2,6" />`;
      }
      this.layers.grid.innerHTML = gridSvg;
    }

    renderWorldLandmass() {
      // Vector paths of major world landmasses and key countries with distinct borders
      // Coordinates accurately scaled to the 1000x520 canvas
      const countries = [
        {
          id: 'US',
          name: 'United States',
          d: 'M 140 148 L 195 142 L 235 148 L 275 140 L 310 140 L 320 162 L 312 188 L 300 205 L 295 240 L 285 242 L 280 230 L 260 232 L 225 240 L 195 240 L 175 230 L 150 205 L 138 185 L 140 148 Z M 60 70 L 105 60 L 115 88 L 95 105 L 65 95 Z'
        },
        {
          id: 'CA',
          name: 'Canada',
          d: 'M 115 88 L 135 105 L 170 120 L 205 135 L 245 138 L 280 135 L 315 138 L 330 115 L 340 70 L 305 60 L 270 50 L 225 45 L 175 52 L 130 65 L 115 88 Z'
        },
        {
          id: 'MX',
          name: 'Mexico',
          d: 'M 175 230 L 195 240 L 225 240 L 245 270 L 230 300 L 210 295 L 190 270 L 175 230 Z'
        },
        {
          id: 'SA',
          name: 'South America',
          d: 'M 260 295 L 310 280 L 360 320 L 390 355 L 375 410 L 340 460 L 305 480 L 295 440 L 290 380 L 270 340 L 255 315 Z'
        },
        {
          id: 'GB',
          name: 'United Kingdom',
          d: 'M 470 140 L 485 135 L 492 145 L 485 160 L 472 162 L 465 150 Z'
        },
        {
          id: 'IE',
          name: 'Ireland',
          d: 'M 458 145 L 465 142 L 466 152 L 458 154 Z'
        },
        {
          id: 'EU_WEST',
          name: 'Western Europe',
          d: 'M 485 165 L 515 155 L 540 160 L 535 185 L 505 195 L 475 210 L 465 195 L 485 165 Z'
        },
        {
          id: 'DE',
          name: 'Germany',
          d: 'M 510 155 L 530 152 L 535 172 L 518 178 L 508 168 Z'
        },
        {
          id: 'NL',
          name: 'Netherlands',
          d: 'M 504 150 L 512 148 L 514 156 L 506 156 Z'
        },
        {
          id: 'FR',
          name: 'France',
          d: 'M 488 170 L 508 168 L 512 190 L 495 198 L 482 188 Z'
        },
        {
          id: 'ES',
          name: 'Spain',
          d: 'M 465 200 L 495 198 L 490 225 L 460 225 Z'
        },
        {
          id: 'IT',
          name: 'Italy',
          d: 'M 520 185 L 535 188 L 545 215 L 535 225 L 525 205 Z'
        },
        {
          id: 'SCANDINAVIA',
          name: 'Scandinavia',
          d: 'M 515 90 L 545 80 L 565 110 L 545 145 L 525 145 L 520 120 Z'
        },
        {
          id: 'EU_EAST',
          name: 'Eastern Europe',
          d: 'M 535 150 L 590 140 L 610 170 L 585 195 L 535 185 Z'
        },
        {
          id: 'AFRICA',
          name: 'Africa',
          d: 'M 465 240 L 535 235 L 585 270 L 620 310 L 590 390 L 555 450 L 525 450 L 485 360 L 450 300 L 455 260 Z'
        },
        {
          id: 'ZA',
          name: 'South Africa',
          d: 'M 525 410 L 575 410 L 560 455 L 525 450 Z'
        },
        {
          id: 'RUSSIA',
          name: 'Russia',
          d: 'M 590 90 L 680 80 L 780 75 L 860 85 L 910 110 L 890 145 L 820 140 L 740 145 L 670 140 L 600 135 Z'
        },
        {
          id: 'ME',
          name: 'Middle East',
          d: 'M 585 220 L 635 215 L 655 255 L 630 280 L 585 260 Z'
        },
        {
          id: 'IR',
          name: 'Iran',
          d: 'M 625 215 L 658 220 L 665 248 L 640 260 L 620 240 Z'
        },
        {
          id: 'TR',
          name: 'Turkey',
          d: 'M 575 205 L 615 202 L 620 220 L 575 222 Z'
        },
        {
          id: 'ASIA_CENTRAL',
          name: 'Central Asia & India',
          d: 'M 645 170 L 720 170 L 740 250 L 710 300 L 680 270 L 660 230 Z'
        },
        {
          id: 'IN',
          name: 'India',
          d: 'M 685 240 L 730 245 L 720 300 L 690 285 Z'
        },
        {
          id: 'CN',
          name: 'China',
          d: 'M 720 160 L 820 155 L 840 210 L 810 245 L 750 240 L 720 190 Z'
        },
        {
          id: 'JP',
          name: 'Japan',
          d: 'M 865 180 L 885 175 L 890 205 L 870 215 Z'
        },
        {
          id: 'KR',
          name: 'South Korea',
          d: 'M 835 195 L 850 190 L 848 210 L 836 210 Z'
        },
        {
          id: 'SE_ASIA',
          name: 'Southeast Asia',
          d: 'M 760 250 L 800 255 L 820 310 L 775 320 Z'
        },
        {
          id: 'AU',
          name: 'Australia',
          d: 'M 810 370 L 890 360 L 920 420 L 880 460 L 820 440 L 800 400 Z'
        },
        {
          id: 'NZ',
          name: 'New Zealand',
          d: 'M 945 440 L 965 445 L 950 480 L 938 465 Z'
        }
      ];

      let landHtml = '';
      countries.forEach(c => {
        landHtml += `
          <path id="map-country-${c.id}" 
                data-country="${c.id}" 
                class="fyrnox-country-path" 
                d="${c.d}" 
                title="${c.name}" />
        `;
      });

      this.layers.land.innerHTML = landHtml;
    }

    renderGeographicLabels() {
      // Subtle oceanic and continental typography matching the screenshot
      const labels = [
        { text: 'NORTH', x: 260, y: 155 },
        { text: 'AMERICA', x: 260, y: 168 },
        { text: 'SOUTH', x: 335, y: 380 },
        { text: 'AMERICA', x: 335, y: 393 },
        { text: 'EUROPE', x: 575, y: 200 },
        { text: 'AFRICA', x: 535, y: 345 },
        { text: 'ASIA', x: 745, y: 205 },
        { text: 'ATLANTIC', x: 410, y: 325 },
        { text: 'OCEAN', x: 410, y: 338 }
      ];

      let labelsHtml = '';
      labels.forEach(l => {
        labelsHtml += `
          <text x="${l.x}" y="${l.y}" class="fyrnox-geo-label" text-anchor="middle">
            ${l.text}
          </text>
        `;
      });

      this.layers.labels.innerHTML = labelsHtml;
    }

    createMapControls() {
      const controls = document.createElement('div');
      controls.classList.add('fyrnox-map-controls');
      controls.innerHTML = `
        <button class="map-ctrl-btn" id="map-zoom-in" title="Zoom In">+</button>
        <button class="map-ctrl-btn" id="map-zoom-out" title="Zoom Out">−</button>
        <button class="map-ctrl-btn" id="map-recenter" title="Recenter View">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="7" />
            <line x1="12" y1="2" x2="12" y2="6" />
            <line x1="12" y1="18" x2="12" y2="22" />
            <line x1="2" y1="12" x2="6" y2="12" />
            <line x1="18" y1="12" x2="22" y2="12" />
          </svg>
        </button>
      `;
      this.container.appendChild(controls);

      document.getElementById('map-zoom-in').addEventListener('click', () => this.zoom(1.3));
      document.getElementById('map-zoom-out').addEventListener('click', () => this.zoom(0.75));
      document.getElementById('map-recenter').addEventListener('click', () => this.recenter());
    }

    bindInteractionEvents() {
      // Pan with Mouse Drag
      this.svg.addEventListener('mousedown', (e) => {
        if (e.target.closest('.fyrnox-server-dot')) return;
        this.isDragging = true;
        this.dragStart = { x: e.clientX, y: e.clientY };
        this.svg.style.cursor = 'grabbing';
      });

      window.addEventListener('mousemove', (e) => {
        if (!this.isDragging) return;
        const dx = (e.clientX - this.dragStart.x) * (this.viewBox.width / this.container.clientWidth);
        const dy = (e.clientY - this.dragStart.y) * (this.viewBox.height / this.container.clientHeight);

        this.viewBox.x -= dx;
        this.viewBox.y -= dy;
        this.dragStart = { x: e.clientX, y: e.clientY };
        this.updateViewBox();
      });

      window.addEventListener('mouseup', () => {
        if (this.isDragging) {
          this.isDragging = false;
          this.svg.style.cursor = 'grab';
        }
      });

      // Mouse Wheel Zoom
      this.container.addEventListener('wheel', (e) => {
        e.preventDefault();
        const factor = e.deltaY < 0 ? 1.15 : 0.85;
        this.zoom(factor, e.clientX, e.clientY);
      }, { passive: false });
    }

    zoom(factor, clientX, clientY) {
      const rect = this.container.getBoundingClientRect();
      const mouseX = clientX ? clientX - rect.left : rect.width / 2;
      const mouseY = clientY ? clientY - rect.top : rect.height / 2;

      // Mouse position in SVG coordinates
      const svgX = this.viewBox.x + (mouseX / rect.width) * this.viewBox.width;
      const svgY = this.viewBox.y + (mouseY / rect.height) * this.viewBox.height;

      const newWidth = Math.max(300, Math.min(this.width * 2, this.viewBox.width / factor));
      const newHeight = newWidth * (this.height / this.width);

      this.viewBox.x = svgX - (mouseX / rect.width) * newWidth;
      this.viewBox.y = svgY - (mouseY / rect.height) * newHeight;
      this.viewBox.width = newWidth;
      this.viewBox.height = newHeight;

      this.updateViewBox();
    }

    recenter() {
      this.viewBox = Object.assign({}, this.defaultViewBox);
      this.updateViewBox();
    }

    updateViewBox() {
      this.svg.setAttribute('viewBox', `${this.viewBox.x} ${this.viewBox.y} ${this.viewBox.width} ${this.viewBox.height}`);
    }

    // Set real available servers from API
    setServers(serversList) {
      this.servers = serversList || [];
      this.renderServerNodes();
    }

    renderServerNodes() {
      let nodesHtml = '';
      this.servers.forEach(srv => {
        const coords = this.getServerCoords(srv);
        if (!coords) return;
        const pt = this.project(coords.lat, coords.lon);
        const isSelected = this.selectedServer && this.selectedServer.hostname === srv.hostname;

        nodesHtml += `
          <g class="fyrnox-server-node ${isSelected ? 'selected' : ''}" 
             data-hostname="${srv.hostname}" 
             transform="translate(${pt.x}, ${pt.y})">
            <circle class="fyrnox-server-pulse" r="8" />
            <circle class="fyrnox-server-dot" r="4" />
          </g>
        `;
      });

      this.layers.servers.innerHTML = nodesHtml;

      // Bind Click & Hover to Server Nodes
      this.layers.servers.querySelectorAll('.fyrnox-server-node').forEach(nodeEl => {
        const hostname = nodeEl.getAttribute('data-hostname');
        const srv = this.servers.find(s => s.hostname === hostname);

        nodeEl.addEventListener('click', (e) => {
          e.stopPropagation();
          if (this.options.onServerSelect && srv) {
            this.options.onServerSelect(srv);
          }
        });

        nodeEl.addEventListener('mouseenter', (e) => {
          if (!srv) return;
          this.showTooltip(e, `
            <strong>${srv.city_name}, ${srv.country_name}</strong>
            <div style="font-size: 0.75rem; color: #00F2FE; margin-top: 4px;">
              Latency: ${srv.ping_ms > 0 ? srv.ping_ms + ' ms' : 'Click to ping'}
            </div>
          `);
        });

        nodeEl.addEventListener('mouseleave', () => this.hideTooltip());
      });
    }

    getServerCoords(srv) {
      if (!srv) return null;
      if (srv.city_code && CITY_COORDINATES[srv.city_code]) {
        return CITY_COORDINATES[srv.city_code];
      }
      if (srv.country_code && COUNTRY_CENTROIDS[srv.country_code]) {
        return COUNTRY_CENTROIDS[srv.country_code];
      }
      return null;
    }

    getOriginCoords(countryCode) {
      if (!countryCode) return null;
      const code = countryCode.toUpperCase();
      if (COUNTRY_CENTROIDS[code]) {
        return COUNTRY_CENTROIDS[code];
      }
      return null;
    }

    // Update connection state, selected server, and origin IP/location
    updateState(status) {
      if (!status) return;

      this.connectionState = status.state || 'disconnected';
      this.selectedServer = status.server || null;

      // Origin detection from status (Pre-connection IP / Country)
      if (status.current_country) {
        const originInfo = this.getOriginCoords(status.current_country);
        if (originInfo) {
          this.originLocation = {
            countryCode: status.current_country,
            countryName: originInfo.name || status.current_country,
            lat: originInfo.lat,
            lon: originInfo.lon,
            ip: status.current_ip || null
          };
        }
      }

      // Re-render Active Route & Nodes
      this.renderConnection();
      this.highlightCountry();
    }

    highlightCountry() {
      // Remove previous highlights
      this.layers.land.querySelectorAll('.fyrnox-country-path').forEach(p => {
        p.classList.remove('selected-destination', 'selected-origin');
      });

      if (this.selectedServer && this.selectedServer.country_code) {
        const destPath = this.layers.land.querySelector(`[data-country="${this.selectedServer.country_code}"]`);
        if (destPath) {
          destPath.classList.add('selected-destination');
        }
      }

      if (this.originLocation && this.originLocation.countryCode) {
        const origPath = this.layers.land.querySelector(`[data-country="${this.originLocation.countryCode}"]`);
        if (origPath) {
          origPath.classList.add('selected-origin');
        }
      }
    }

    renderConnection() {
      this.layers.nodes.innerHTML = '';
      this.layers.arcs.innerHTML = '';

      const destCoords = this.getServerCoords(this.selectedServer);
      const origCoords = this.originLocation;

      // If no server selected, clean up
      if (!destCoords) {
        this.activeArc = null;
        return;
      }

      const destPt = this.project(destCoords.lat, destCoords.lon);

      // Render Destination Marker (Green/Cyan glowing halo with radar ring)
      const destNodeHtml = `
        <g class="fyrnox-dest-marker" transform="translate(${destPt.x}, ${destPt.y})">
          <circle class="dest-halo" r="16" />
          <circle class="dest-ring" r="9" />
          <circle class="dest-core" r="4.5" />
          <text class="dest-label" y="24" text-anchor="middle">
            ${this.selectedServer.city_name || destCoords.city || destCoords.name}
          </text>
        </g>
      `;

      let origNodeHtml = '';
      let arcsHtml = '';

      if (origCoords) {
        const origPt = this.project(origCoords.lat, origCoords.lon);

        // Render Origin Marker (Pulsating Blue/Cyan Beacon)
        origNodeHtml = `
          <g class="fyrnox-origin-marker" transform="translate(${origPt.x}, ${origPt.y})">
            <circle class="origin-beacon-ring" r="14" />
            <circle class="origin-core" r="4.5" />
            <text class="origin-label" y="24" text-anchor="middle">
              ${origCoords.countryName.toUpperCase()}
            </text>
          </g>
        `;

        // Calculate Great-Circle Curved Arc between Origin and Destination
        const dx = destPt.x - origPt.x;
        const dy = destPt.y - origPt.y;
        const dist = Math.hypot(dx, dy);

        // Arc upward: control point elevated proportionally to distance
        const midX = (origPt.x + destPt.x) / 2;
        const midY = (origPt.y + destPt.y) / 2 - Math.min(dist * 0.38, 130);

        const pathData = `M ${origPt.x} ${origPt.y} Q ${midX} ${midY} ${destPt.x} ${destPt.y}`;

        // 3-Layered luminous connection arcs (Outer glow, Middle stream, Inner core)
        const isConnected = this.connectionState === 'connected';
        const isConnecting = this.connectionState === 'connecting';

        if (isConnected || isConnecting) {
          arcsHtml = `
            <!-- Layer 1: Ambient Glow -->
            <path d="${pathData}" class="arc-glow ${isConnecting ? 'connecting' : ''}" />
            
            <!-- Layer 2: Main Stream Beam -->
            <path id="active-connection-path" d="${pathData}" class="arc-beam ${isConnecting ? 'connecting' : ''}" />
            
            <!-- Layer 3: Sharp High-tech Accent Lines -->
            <path d="${pathData}" class="arc-core ${isConnecting ? 'connecting' : ''}" />
          `;

          // Track path for particle flow
          setTimeout(() => {
            const pathEl = document.getElementById('active-connection-path');
            if (pathEl) {
              this.activeArc = {
                pathEl: pathEl,
                totalLength: pathEl.getTotalLength()
              };
            }
          }, 20);
        } else {
          this.activeArc = null;
        }
      }

      this.layers.nodes.innerHTML = origNodeHtml + destNodeHtml;
      this.layers.arcs.innerHTML = arcsHtml;
    }

    // Animated particles streaming along the connection arc
    startParticleAnimation() {
      // Create pool of 4 animated particles
      this.streamParticles = [
        { progress: 0.0, speed: 0.006 },
        { progress: 0.25, speed: 0.006 },
        { progress: 0.5, speed: 0.006 },
        { progress: 0.75, speed: 0.006 }
      ];

      const render = () => {
        if (this.connectionState === 'connected' && this.activeArc && this.activeArc.pathEl) {
          const len = this.activeArc.totalLength;
          let particlesSvg = '';

          this.streamParticles.forEach(p => {
            p.progress += p.speed;
            if (p.progress > 1) p.progress = 0;

            const pt = this.activeArc.pathEl.getPointAtLength(p.progress * len);
            particlesSvg += `
              <circle cx="${pt.x}" cy="${pt.y}" r="3.5" fill="url(#particle-grad)" filter="url(#fyrnox-intense-glow)" />
              <circle cx="${pt.x}" cy="${pt.y}" r="1.5" fill="#FFFFFF" />
            `;
          });

          this.layers.particles.innerHTML = particlesSvg;
        } else {
          this.layers.particles.innerHTML = '';
        }

        this.animFrameId = requestAnimationFrame(render);
      };

      this.animFrameId = requestAnimationFrame(render);
    }

    showTooltip(e, html) {
      const rect = this.container.getBoundingClientRect();
      const x = e.clientX - rect.left + 14;
      const y = e.clientY - rect.top - 20;

      this.tooltip.innerHTML = html;
      this.tooltip.style.left = `${x}px`;
      this.tooltip.style.top = `${y}px`;
      this.tooltip.style.display = 'block';
    }

    hideTooltip() {
      this.tooltip.style.display = 'none';
    }

    destroy() {
      if (this.animFrameId) {
        cancelAnimationFrame(this.animFrameId);
      }
    }
  }

  // Export to window
  window.FyrnoxMap = FyrnoxMap;

})(window);
