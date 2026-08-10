import { useState, useEffect, useCallback } from 'react';
import { api } from '../api/client';

export function useCbomScan() {
  const [scans, setScans] = useState([]);
  const [targets, setTargets] = useState([]);
  const [currentScan, setCurrentScan] = useState(null);
  const [loading, setLoading] = useState(true);
  const [scanning, setScanning] = useState(false);
  const [error, setError] = useState(null);
  const [githubRepos, setGithubRepos] = useState([]);
  const [loadingRepos, setLoadingRepos] = useState(false);

  const loadScansAndTargets = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const [scansData, targetsData] = await Promise.all([
        api.getScans().catch(() => []),
        api.getTargets().catch(() => []),
      ]);

      setScans(scansData || []);
      setTargets(targetsData || []);
      
      // Select most recent scan if available
      if (scansData && scansData.length > 0) {
        const detail = await api.getScanDetail(scansData[0].id);
        setCurrentScan(detail);
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  const selectScan = async (scanId) => {
    try {
      setLoading(true);
      const detail = await api.getScanDetail(scanId);
      setCurrentScan(detail);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const startLocalScan = async (targetPath, name = 'Local Codebase') => {
    try {
      setScanning(true);
      setError(null);

      // Create local target first
      const newTarget = await api.createTarget({
        name,
        type: 'REPO',
        location: targetPath,
      });

      // Execute scan by target
      const detail = await api.triggerScanByTarget(newTarget.id);
      setCurrentScan(detail);
      await loadScansAndTargets();
      return detail;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setScanning(false);
    }
  };

  const fetchGitHubRepositories = async (installationId) => {
    try {
      setLoadingRepos(true);
      setError(null);
      const repos = await api.getGitHubRepos(installationId);
      setGithubRepos(repos || []);
      return repos;
    } catch (err) {
      setError(`GitHub App Error: ${err.message}`);
      throw err;
    } finally {
      setLoadingRepos(false);
    }
  };

  const startGitHubScan = async (repo, installationId) => {
    try {
      setScanning(true);
      setError(null);

      // Create GitHub target
      const newTarget = await api.createTarget({
        name: repo.name || repo.full_name,
        type: 'GITHUB_REPO',
        location: repo.full_name,
        installation_id: parseInt(installationId, 10),
        full_name: repo.full_name,
        default_branch: repo.default_branch || 'main',
      });

      // Execute scan by target
      const detail = await api.triggerScanByTarget(newTarget.id);
      setCurrentScan(detail);
      await loadScansAndTargets();
      return detail;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setScanning(false);
    }
  };

  useEffect(() => {
    loadScansAndTargets();
  }, [loadScansAndTargets]);

  return {
    scans,
    targets,
    currentScan,
    loading,
    scanning,
    error,
    githubRepos,
    loadingRepos,
    selectScan,
    startLocalScan,
    fetchGitHubRepositories,
    startGitHubScan,
    reload: loadScansAndTargets,
    setError,
  };
}
