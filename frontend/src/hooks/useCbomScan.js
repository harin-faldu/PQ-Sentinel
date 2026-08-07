import { useState, useEffect, useCallback } from 'react';
import { api } from '../api/client';

export function useCbomScan() {
  const [scans, setScans] = useState([]);
  const [currentScan, setCurrentScan] = useState(null);
  const [loading, setLoading] = useState(true);
  const [scanning, setScanning] = useState(false);
  const [error, setError] = useState(null);

  const loadScans = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.getScans();
      setScans(data || []);
      
      // Select demo scan if available
      if (data && data.length > 0) {
        const detail = await api.getScanDetail(data[0].id);
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

  const startScan = async (targetDirectory) => {
    try {
      setScanning(true);
      setError(null);
      const detail = await api.triggerScan(targetDirectory);
      setCurrentScan(detail);
      await loadScans();
      return detail;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setScanning(false);
    }
  };

  useEffect(() => {
    loadScans();
  }, [loadScans]);

  return {
    scans,
    currentScan,
    loading,
    scanning,
    error,
    selectScan,
    startScan,
    reload: loadScans,
  };
}
