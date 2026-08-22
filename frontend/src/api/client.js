const API_BASE = '/api/v1';

/**
 * Returns a stable per-browser-session UUID.
 * Persists across page refreshes via localStorage.
 * Each browser tab that was opened fresh gets the same session as long as
 * localStorage is shared (same origin, same tab or duplicated tab).
 */
function getOrCreateSessionId() {
  let id = localStorage.getItem('cbom_session_id');
  if (!id) {
    id = typeof crypto !== 'undefined' && crypto.randomUUID
      ? crypto.randomUUID()
      : Math.random().toString(36).slice(2) + Date.now().toString(36);
    localStorage.setItem('cbom_session_id', id);
  }
  return id;
}

export async function fetchApi(endpoint, options = {}) {
  const url = `${API_BASE}${endpoint}`;
  const defaultHeaders = {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
    'X-Session-ID': getOrCreateSessionId(),
  };

  const config = {
    ...options,
    headers: {
      ...defaultHeaders,
      ...options.headers,
    },
  };

  try {
    const response = await fetch(url, config);
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
    }
    return await response.json();
  } catch (error) {
    console.error(`[API Client Error] ${endpoint}:`, error.message);
    throw error;
  }
}

export const api = {
  getHealth: () => fetchApi('/health'),
  
  // Scans
  getScans: () => fetchApi('/scans'),
  getScanDetail: (id) => fetchApi(`/scans/${id}`),
  triggerScan: (targetDirectory) => fetchApi('/scans', {
    method: 'POST',
    body: JSON.stringify({ target_directory: targetDirectory }),
  }),
  triggerScanByTarget: (targetId, tlsTargets = []) => fetchApi(`/targets/${targetId}/scan`, {
    method: 'POST',
    body: JSON.stringify(tlsTargets.length > 0 ? { tls_targets: tlsTargets } : {}),
  }),

  // Targets
  getTargets: () => fetchApi('/targets'),
  createTarget: (targetData) => fetchApi('/targets', {
    method: 'POST',
    body: JSON.stringify(targetData),
  }),
  getTargetScans: (targetId) => fetchApi(`/targets/${targetId}/scans`),

  // Upload Folder Ingestion & Scan
  uploadFolderScan: async (formData) => {
    const response = await fetch(`${API_BASE}/scans/upload`, {
      method: 'POST',
      body: formData,
      headers: {
        'Accept': 'application/json',
        'X-Session-ID': getOrCreateSessionId(),
      },
    });
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `Upload scan failed: HTTP ${response.status}`);
    }
    return await response.json();
  },

  // Migration Plan & CycloneDX Export
  getMigrationPlan: (scanId) => fetchApi(`/scans/${scanId}/migration`),
  getCycloneDXJsonUrl: (scanId) => `${API_BASE}/scans/${scanId}/cyclonedx`,
  getCycloneDXXmlUrl: (scanId) => `${API_BASE}/scans/${scanId}/cyclonedx?format=xml`,
};
