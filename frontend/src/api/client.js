const API_BASE = '/api/v1';

export async function fetchApi(endpoint, options = {}) {
  const url = `${API_BASE}${endpoint}`;
  const defaultHeaders = {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
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

  // GitHub App Integration
  getGitHubRepos: (installationId) => fetchApi(`/github/repos?installation_id=${installationId}`),
  getGitHubInstallUrl: () => fetchApi('/github/install-url'),

  // Migration Plan & CycloneDX Export
  getMigrationPlan: (scanId) => fetchApi(`/scans/${scanId}/migration`),
  getCycloneDXJsonUrl: (scanId) => `${API_BASE}/scans/${scanId}/cyclonedx`,
  getCycloneDXXmlUrl: (scanId) => `${API_BASE}/scans/${scanId}/cyclonedx?format=xml`,
};
