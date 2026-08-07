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
  getScans: () => fetchApi('/scans'),
  getScanDetail: (id) => fetchApi(`/scans/${id}`),
  triggerScan: (targetDirectory) => fetchApi('/scans', {
    method: 'POST',
    body: JSON.stringify({ target_directory: targetDirectory }),
  }),
  getMigrationPlan: (scanId) => fetchApi(`/scans/${scanId}/migration`),
  getCycloneDXUrl: (scanId) => `${API_BASE}/scans/${scanId}/cyclonedx`,
};
