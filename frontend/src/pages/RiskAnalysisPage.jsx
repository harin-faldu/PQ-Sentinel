import React from 'react';
import { ShieldAlert, AlertTriangle, ShieldCheck, FileCode } from 'lucide-react';
import { getRiskBadgeClass } from '../utils/formatters';

export function RiskAnalysisPage({ currentScan }) {
  const assets = currentScan?.assets || [];

  const criticalAssets = assets.filter((a) => a.risk_level === 'CRITICAL');
  const highAssets = assets.filter((a) => a.risk_level === 'HIGH');
  const mediumAssets = assets.filter((a) => a.risk_level === 'MEDIUM');

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      
      {/* Page Header */}
      <div className="glass-card" style={{ padding: '24px' }}>
        <h2 style={{ fontSize: '1.25rem', fontWeight: 800 }}>Cryptographic Vulnerability & PQC Risk Analysis</h2>
        <p style={{ color: '#94a3b8', fontSize: '0.85rem', marginTop: '6px' }}>
          Detailed vulnerability findings scored against NIST SP 800-131A & CNSA 2.0 Post-Quantum Cryptography standards.
        </p>
      </div>

      {/* Critical Section */}
      {criticalAssets.length > 0 && (
        <div className="glass-card" style={{ padding: '24px', borderLeft: '4px solid #ef4444' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '16px' }}>
            <ShieldAlert color="#ef4444" size={22} />
            <h3 style={{ fontSize: '1.1rem', fontWeight: 700, color: '#ef4444' }}>
              Critical Risk Findings ({criticalAssets.length})
            </h3>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            {criticalAssets.map((asset) => (
              <div key={asset.id} style={{ background: 'rgba(0,0,0,0.3)', padding: '16px', borderRadius: '10px', border: '1px solid rgba(239, 68, 68, 0.2)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontWeight: 700, fontSize: '0.95rem' }}>{asset.name}</span>
                  <span className={`badge ${getRiskBadgeClass(asset.risk_level)}`}>{asset.risk_level}</span>
                </div>
                <p style={{ fontSize: '0.825rem', color: '#94a3b8', marginTop: '6px' }}>{asset.details}</p>
                <div style={{ marginTop: '10px', display: 'flex', gap: '16px', fontSize: '0.8rem', color: '#64748b' }}>
                  <span>Location: <code className="code-inline">{asset.location}:{asset.line_number}</code></span>
                  <span>Recommended PQC Action: <strong style={{ color: '#a855f7' }}>{asset.pqc_replacement}</strong></span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* High Risk Section */}
      {highAssets.length > 0 && (
        <div className="glass-card" style={{ padding: '24px', borderLeft: '4px solid #f97316' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '16px' }}>
            <AlertTriangle color="#f97316" size={22} />
            <h3 style={{ fontSize: '1.1rem', fontWeight: 700, color: '#f97316' }}>
              High Risk - Quantum Vulnerable Assets ({highAssets.length})
            </h3>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            {highAssets.map((asset) => (
              <div key={asset.id} style={{ background: 'rgba(0,0,0,0.3)', padding: '16px', borderRadius: '10px', border: '1px solid rgba(249, 115, 22, 0.2)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontWeight: 700, fontSize: '0.95rem' }}>{asset.name}</span>
                  <span className={`badge ${getRiskBadgeClass(asset.risk_level)}`}>{asset.risk_level}</span>
                </div>
                <p style={{ fontSize: '0.825rem', color: '#94a3b8', marginTop: '6px' }}>{asset.details}</p>
                <div style={{ marginTop: '10px', display: 'flex', gap: '16px', fontSize: '0.8rem', color: '#64748b' }}>
                  <span>Location: <code className="code-inline">{asset.location}:{asset.line_number}</code></span>
                  <span>Target PQC Replacement: <strong style={{ color: '#a855f7' }}>{asset.pqc_replacement}</strong></span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

    </div>
  );
}
