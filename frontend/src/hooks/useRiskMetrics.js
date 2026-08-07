import { useMemo } from 'react';

export function useRiskMetrics(assets = []) {
  return useMemo(() => {
    let critical = 0;
    let high = 0;
    let medium = 0;
    let safe = 0;

    let algorithms = 0;
    let certificates = 0;
    let protocols = 0;
    let keyStores = 0;
    let libraries = 0;

    assets.forEach((a) => {
      switch (a.risk_level) {
        case 'CRITICAL':
          critical++;
          break;
        case 'HIGH':
          high++;
          break;
        case 'MEDIUM':
          medium++;
          break;
        case 'SAFE':
        case 'LOW':
          safe++;
          break;
      }

      switch (a.asset_type) {
        case 'ALGORITHM':
          algorithms++;
          break;
        case 'CERTIFICATE':
          certificates++;
          break;
        case 'PROTOCOL':
          protocols++;
          break;
        case 'KEYSTORE':
        case 'SECRET':
          keyStores++;
          break;
        case 'LIBRARY':
          libraries++;
          break;
      }
    });

    const total = assets.length;
    const score = total > 0 
      ? Math.max(0, Math.round(100 - ((critical * 35 + high * 20 + medium * 10) / total) * 2.5))
      : 100;

    return {
      total,
      score,
      critical,
      high,
      medium,
      safe,
      typeBreakdown: {
        algorithms,
        certificates,
        protocols,
        keyStores,
        libraries,
      },
    };
  }, [assets]);
}
