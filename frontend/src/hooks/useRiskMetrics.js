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

    // Scanner-specific breakdown counts
    let kmsFindings = 0;
    let containerFindings = 0;
    let tlsFindings = 0;
    let secretFindings = 0;

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
          keyStores++;
          break;
        case 'SECRET':
          secretFindings++;
          keyStores++;
          break;
        case 'LIBRARY':
          libraries++;
          break;
      }

      // Detect scanner origin from name/details patterns
      const name = (a.name || '').toLowerCase();
      const details = (a.details || '').toLowerCase();
      if (name.includes('aws kms') || name.includes('cloudformation aws kms') || details.includes('aws kms')) {
        kmsFindings++;
      }
      if (name.includes('container') || name.includes('dockerfile') || details.includes('dockerfile') || details.includes('syft')) {
        containerFindings++;
      }
      if (name.includes('active tls') || name.includes('tls cert') || name.includes('tls endpoint')) {
        tlsFindings++;
      }
      if (a.asset_type === 'SECRET' || name.includes('leaked') || name.includes('secret') || name.includes('exposed')) {
        // Only count if not already counted
        if (a.asset_type !== 'SECRET') secretFindings++;
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
      scannerBreakdown: {
        kms: kmsFindings,
        container: containerFindings,
        tls: tlsFindings,
        secrets: secretFindings,
      },
    };
  }, [assets]);
}

