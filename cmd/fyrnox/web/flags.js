// FYRNOX Vector SVG Country Flags
// Pixel-perfect vector rendering for all 35 verified countries, completely bypassing
// Windows font/Segoe UI Emoji limitations (which fail to render flag emojis).

const SVG_FLAGS = {
  US: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="24" fill="#B22234"/><path d="M0,3.7h32M0,7.4h32M0,11h32M0,14.7h32M0,18.4h32M0,22.1h32" stroke="#FFF" stroke-width="1.85"/><rect width="13" height="13" fill="#3C3B6E"/><g fill="#FFF"><circle cx="2.5" cy="2.5" r="0.7"/><circle cx="6.5" cy="2.5" r="0.7"/><circle cx="10.5" cy="2.5" r="0.7"/><circle cx="4.5" cy="4.5" r="0.7"/><circle cx="8.5" cy="4.5" r="0.7"/><circle cx="2.5" cy="6.5" r="0.7"/><circle cx="6.5" cy="6.5" r="0.7"/><circle cx="10.5" cy="6.5" r="0.7"/><circle cx="4.5" cy="8.5" r="0.7"/><circle cx="8.5" cy="8.5" r="0.7"/><circle cx="2.5" cy="10.5" r="0.7"/><circle cx="6.5" cy="10.5" r="0.7"/><circle cx="10.5" cy="10.5" r="0.7"/></g></svg>`,
  
  GB: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="24" fill="#012169"/><path d="M0,0 L32,24 M32,0 L0,24" stroke="#FFF" stroke-width="4"/><path d="M0,0 L32,24 M32,0 L0,24" stroke="#C8102E" stroke-width="2"/><path d="M16,0 V24 M0,12 H32" stroke="#FFF" stroke-width="6"/><path d="M16,0 V24 M0,12 H32" stroke="#C8102E" stroke-width="3.6"/></svg>`,
  
  DE: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="8" fill="#000"/><rect y="8" width="32" height="8" fill="#DD0000"/><rect y="16" width="32" height="8" fill="#FFCE00"/></svg>`,
  
  FR: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="10.7" height="24" fill="#002654"/><rect x="10.7" width="10.6" height="24" fill="#FFF"/><rect x="21.3" width="10.7" height="24" fill="#CE1126"/></svg>`,
  
  NL: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="8" fill="#AE1C28"/><rect y="8" width="32" height="8" fill="#FFF"/><rect y="16" width="32" height="8" fill="#21468B"/></svg>`,
  
  CA: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="8" height="24" fill="#FF0000"/><rect x="8" width="16" height="24" fill="#FFF"/><rect x="24" width="8" height="24" fill="#FF0000"/><path d="M16,5 L17.5,9.5 L19.5,8 L19,10.5 L21.5,11.5 L20,13 L21,15 L18,14.5 L17.5,17 L16.5,15.5 L16.5,19 L15.5,19 L15.5,15.5 L14.5,17 L14,14.5 L11,15 L12,13 L10.5,11.5 L13,10.5 L12.5,8 L14.5,9.5 Z" fill="#FF0000"/></svg>`,
  
  JP: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="24" fill="#FFF"/><circle cx="16" cy="12" r="7" fill="#BC002D"/></svg>`,
  
  SE: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="24" fill="#006AA7"/><rect x="9" width="4.5" height="24" fill="#FECC00"/><rect y="10" width="32" height="4.5" fill="#FECC00"/></svg>`,
  
  NO: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="24" fill="#BA0C2F"/><path d="M9,0 H15 V24 H9 Z M0,9 H32 V15 H0 Z" fill="#FFF"/><path d="M10.5,0 H13.5 V24 H10.5 Z M0,10.5 H32 V13.5 H0 Z" fill="#00205B"/></svg>`,
  
  FI: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="24" fill="#FFF"/><rect x="9" width="5" height="24" fill="#003580"/><rect y="9.5" width="32" height="5" fill="#003580"/></svg>`,
  
  DK: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="24" fill="#C8102E"/><rect x="9" width="4" height="24" fill="#FFF"/><rect y="10" width="32" height="4" fill="#FFF"/></svg>`,
  
  IE: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="10.7" height="24" fill="#169B62"/><rect x="10.7" width="10.6" height="24" fill="#FFF"/><rect x="21.3" width="10.7" height="24" fill="#FF883E"/></svg>`,
  
  IT: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="10.7" height="24" fill="#009246"/><rect x="10.7" width="10.6" height="24" fill="#FFF"/><rect x="21.3" width="10.7" height="24" fill="#CE2B37"/></svg>`,
  
  ES: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="6" fill="#AA151B"/><rect y="6" width="32" height="12" fill="#F1BF00"/><rect y="18" width="32" height="6" fill="#AA151B"/><circle cx="9" cy="12" r="2.5" fill="#AA151B"/></svg>`,
  
  AT: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="8" fill="#ED2939"/><rect y="8" width="32" height="8" fill="#FFF"/><rect y="16" width="32" height="8" fill="#ED2939"/></svg>`,
  
  BE: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="10.7" height="24" fill="#000"/><rect x="10.7" width="10.6" height="24" fill="#FDDA24"/><rect x="21.3" width="10.7" height="24" fill="#EF3340"/></svg>`,
  
  PT: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="12" height="24" fill="#046A38"/><rect x="12" width="20" height="24" fill="#DA291C"/><circle cx="12" cy="12" r="3.5" fill="#FED100"/></svg>`,
  
  AU: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="24" fill="#00008B"/><path d="M0,0 L14,10.5 M14,0 L0,10.5" stroke="#FFF" stroke-width="2"/><path d="M7,0 V10.5 M0,5.25 H14" stroke="#FFF" stroke-width="3"/><path d="M7,0 V10.5 M0,5.25 H14" stroke="#C8102E" stroke-width="1.8"/><g fill="#FFF"><circle cx="24" cy="4" r="0.8"/><circle cx="26" cy="8" r="0.8"/><circle cx="24" cy="14" r="0.8"/><circle cx="21" cy="9" r="0.8"/><circle cx="23" cy="10.5" r="0.5"/><circle cx="7" cy="17" r="1.8"/></g></svg>`,
  
  NZ: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="24" fill="#00247D"/><path d="M0,0 L14,10.5 M14,0 L0,10.5" stroke="#FFF" stroke-width="2"/><path d="M7,0 V10.5 M0,5.25 H14" stroke="#FFF" stroke-width="3"/><path d="M7,0 V10.5 M0,5.25 H14" stroke="#CC142B" stroke-width="1.8"/><g fill="#CC142B" stroke="#FFF" stroke-width="0.3"><circle cx="24" cy="4" r="0.9"/><circle cx="26.5" cy="8.5" r="0.9"/><circle cx="24" cy="14" r="0.9"/><circle cx="21" cy="9.5" r="0.7"/></g></svg>`,
  
  SG: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="12" fill="#ED2939"/><rect y="12" width="32" height="12" fill="#FFF"/><circle cx="6" cy="6" r="3.5" fill="#FFF"/><circle cx="7" cy="6" r="3.2" fill="#ED2939"/><g fill="#FFF"><circle cx="7.5" cy="4" r="0.5"/><circle cx="9" cy="5" r="0.5"/><circle cx="8.5" cy="7" r="0.5"/><circle cx="6.5" cy="7" r="0.5"/><circle cx="6" cy="5" r="0.5"/></g></svg>`,
  
  KR: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="24" fill="#FFF"/><circle cx="16" cy="12" r="5" fill="#CD2E3A"/><path d="M16,12 A2.5,2.5 0 0,1 18.5,12 A2.5,2.5 0 0,0 13.5,12 A5,5 0 0,0 16,17 A5,5 0 0,0 21,12 Z" fill="#0047A0"/><g stroke="#000" stroke-width="0.8"><line x1="8" y1="6" x2="10" y2="8"/><line x1="22" y1="16" x2="24" y2="18"/></g></svg>`,
  
  BR: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="24" fill="#009739"/><polygon points="16,3 29,12 16,21 3,12" fill="#FEDD00"/><circle cx="16" cy="12" r="4.5" fill="#012169"/><path d="M12,13 A5,5 0 0,1 20,11" stroke="#FFF" stroke-width="0.6" fill="none"/></svg>`,
  
  MX: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="10.7" height="24" fill="#006847"/><rect x="10.7" width="10.6" height="24" fill="#FFF"/><rect x="21.3" width="10.7" height="24" fill="#CE1126"/><circle cx="16" cy="12" r="2.2" fill="#8B4513"/></svg>`,
  
  AR: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="8" fill="#74ACDF"/><rect y="8" width="32" height="8" fill="#FFF"/><rect y="16" width="32" height="8" fill="#74ACDF"/><circle cx="16" cy="12" r="2" fill="#F6B40E"/></svg>`,
  
  CL: `<svg viewBox="0 0 32 24" class="flag-svg"><rect y="12" width="32" height="12" fill="#D52B1E"/><rect width="32" height="12" fill="#FFF"/><rect width="12" height="12" fill="#0039A6"/><polygon points="6,3.5 7,6 9.5,6 7.5,7.5 8.2,10 6,8.5 3.8,10 4.5,7.5 2.5,6 5,6" fill="#FFF"/></svg>`,
  
  CO: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="12" fill="#FCD116"/><rect y="12" width="32" height="6" fill="#003893"/><rect y="18" width="32" height="6" fill="#CE1126"/></svg>`,
  
  PE: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="10.7" height="24" fill="#D91023"/><rect x="10.7" width="10.6" height="24" fill="#FFF"/><rect x="21.3" width="10.7" height="24" fill="#D91023"/></svg>`,
  
  ZA: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="12" fill="#E03C31"/><rect y="12" width="32" height="12" fill="#001489"/><polygon points="0,0 12,12 0,24" fill="#000"/><polygon points="0,0 15,12 0,24" fill="none" stroke="#FFB81C" stroke-width="2"/><polygon points="0,0 18,12 0,24" fill="none" stroke="#007749" stroke-width="3"/><rect y="9.5" width="32" height="5" fill="#007749"/><line x1="0" y1="12" x2="32" y2="12" stroke="#FFF" stroke-width="1"/></svg>`,
  
  GH: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="8" fill="#CE1126"/><rect y="8" width="32" height="8" fill="#FCD116"/><rect y="16" width="32" height="8" fill="#006B3F"/><polygon points="16,8.5 17,11.5 20,11.5 17.5,13 18.5,16 16,14.2 13.5,16 14.5,13 12,11.5 15,11.5" fill="#000"/></svg>`,
  
  MY: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="24" fill="#CC0000"/><path d="M0,2.4h32M0,4.8h32M0,7.2h32M0,9.6h32M0,12h32M0,14.4h32M0,16.8h32M0,19.2h32M0,21.6h32" stroke="#FFF" stroke-width="1.2"/><rect width="16" height="13" fill="#000066"/><circle cx="7" cy="6.5" r="4" fill="#FFCC00"/><circle cx="8" cy="6.5" r="3.4" fill="#000066"/><circle cx="11" cy="6.5" r="1.8" fill="#FFCC00"/></svg>`,
  
  PH: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="12" fill="#0038A8"/><rect y="12" width="32" height="12" fill="#CE1126"/><polygon points="0,0 14,12 0,24" fill="#FFF"/><circle cx="4.5" cy="12" r="2" fill="#FCD116"/><circle cx="2" cy="3" r="0.6" fill="#FCD116"/><circle cx="2" cy="21" r="0.6" fill="#FCD116"/><circle cx="10" cy="12" r="0.6" fill="#FCD116"/></svg>`,
  
  TH: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="4" fill="#A51931"/><rect y="4" width="32" height="4" fill="#F4F5F8"/><rect y="8" width="32" height="8" fill="#2D2A4A"/><rect y="16" width="32" height="4" fill="#F4F5F8"/><rect y="20" width="32" height="4" fill="#A51931"/></svg>`,
  
  BG: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="8" fill="#FFF"/><rect y="8" width="32" height="8" fill="#00966E"/><rect y="16" width="32" height="8" fill="#D62612"/></svg>`,
  
  IR: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="8" fill="#239F40"/><rect y="8" width="32" height="8" fill="#FFF"/><rect y="16" width="32" height="8" fill="#DA0000"/><circle cx="16" cy="12" r="2.2" fill="#DA0000"/></svg>`,

  // Global Anycast / Recommended Edge / Unknown fallback
  GLOBAL: `<svg viewBox="0 0 32 24" class="flag-svg"><rect width="32" height="24" fill="#1E293B"/><circle cx="16" cy="12" r="8" fill="none" stroke="#FF9400" stroke-width="1.5"/><ellipse cx="16" cy="12" rx="4" ry="8" fill="none" stroke="#FE6C31" stroke-width="1.2"/><line x1="8" y1="12" x2="24" y2="12" stroke="#FF9400" stroke-width="1.2"/></svg>`
};

function getFlagSVG(code) {
  if (!code) return SVG_FLAGS.GLOBAL;
  const c = String(code).toUpperCase().trim();
  const svg = SVG_FLAGS[c] || SVG_FLAGS.GLOBAL;
  return `<span class="flag-icon-svg" title="${c}">${svg}</span>`;
}

// Export for browser environment
if (typeof window !== 'undefined') {
  window.SVG_FLAGS = SVG_FLAGS;
  window.getFlagSVG = getFlagSVG;
}
