import React from 'react';
import { ShieldAlert, AlertTriangle } from 'lucide-react';
import Card from '../components/Card';
import { SeverityBadge, TypeBadge } from '../components/Badge';

export function RiskAnalysisPage({ currentScan }) {
  const assets = currentScan?.assets || [];

  const criticalAssets = assets.filter((a) => a.risk_level === 'CRITICAL');
  const highAssets = assets.filter((a) => a.risk_level === 'HIGH');

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      
      {/* Critical Section */}
      {criticalAssets.length > 0 && (
        <Card style={{ padding: '24px', borderLeft: '4px solid var(--color-error)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '16px' }}>
            <ShieldAlert color="var(--color-error)" size={22} />
            <h3 style={{ fontSize: '1.1rem', fontWeight: 700, color: 'var(--color-error)' }}>
              Critical Risk Findings ({criticalAssets.length})
            </h3>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            {criticalAssets.map((asset) => (
              <div key={asset.id} style={{ background: 'var(--color-bg-tertiary)', padding: '16px', borderRadius: 'var(--radius-md)', border: '1px solid var(--color-border)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                    <span style={{ fontWeight: 700, fontSize: '0.95rem' }}>{asset.name}</span>
                    <TypeBadge type={asset.asset_type || asset.type || 'ALGORITHM'} />
                  </div>
                  <SeverityBadge level={asset.risk_level} />
                </div>
                <p style={{ fontSize: '0.825rem', color: 'var(--color-text-secondary)', marginTop: '6px' }}>{asset.details}</p>
                <div style={{ marginTop: '10px', display: 'flex', gap: '16px', fontSize: '0.8rem', color: 'var(--color-text-muted)', flexWrap: 'wrap' }}>
                  <span>Location: <code className="code-inline">{asset.location}:{asset.line_number}</code></span>
                  <span>Recommended PQC Action: <strong style={{ color: 'var(--color-primary)' }}>{asset.pqc_replacement}</strong></span>
                </div>
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* High Risk Section */}
      {highAssets.length > 0 && (
        <Card style={{ padding: '24px', borderLeft: '4px solid var(--color-warning)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '16px' }}>
            <AlertTriangle color="var(--color-warning)" size={22} />
            <h3 style={{ fontSize: '1.1rem', fontWeight: 700, color: 'var(--color-warning)' }}>
              High Risk - Quantum Vulnerable Assets ({highAssets.length})
            </h3>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            {highAssets.map((asset) => (
              <div key={asset.id} style={{ background: 'var(--color-bg-tertiary)', padding: '16px', borderRadius: 'var(--radius-md)', border: '1px solid var(--color-border)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                    <span style={{ fontWeight: 700, fontSize: '0.95rem' }}>{asset.name}</span>
                    <TypeBadge type={asset.asset_type || asset.type || 'ALGORITHM'} />
                  </div>
                  <SeverityBadge level={asset.risk_level} />
                </div>
                <p style={{ fontSize: '0.825rem', color: 'var(--color-text-secondary)', marginTop: '6px' }}>{asset.details}</p>
                <div style={{ marginTop: '10px', display: 'flex', gap: '16px', fontSize: '0.8rem', color: 'var(--color-text-muted)', flexWrap: 'wrap' }}>
                  <span>Location: <code className="code-inline">{asset.location}:{asset.line_number}</code></span>
                  <span>Target PQC Replacement: <strong style={{ color: 'var(--color-primary)' }}>{asset.pqc_replacement}</strong></span>
                </div>
              </div>
            ))}
          </div>
        </Card>
      )}

    </div>
  );
}
