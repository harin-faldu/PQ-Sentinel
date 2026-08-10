import React from 'react';
import { OverviewCards } from '../components/OverviewCards';
import { RiskGauge } from '../components/RiskGauge';
import { AssetTable } from '../components/AssetTable';
import { useRiskMetrics } from '../hooks/useRiskMetrics';
import { Cpu, ArrowRight } from 'lucide-react';

export function DashboardPage({ currentScan, onNavigateToMigration }) {
  const assets = currentScan?.assets || [];
  const metrics = useRiskMetrics(assets);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      
      {/* Top Banner & Risk Gauge - Responsive Grid */}
      <div className="dashboard-grid">
        
        {/* Banner */}
        <div className="glass-card" style={{ padding: '28px', display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
          <div>
            <div style={{ display: 'inline-flex', alignItems: 'center', gap: '6px', background: 'rgba(99, 102, 241, 0.15)', color: '#6366f1', padding: '4px 12px', borderRadius: '20px', fontSize: '0.75rem', fontWeight: 700 }}>
              <Cpu size={14} /> ACTIVE TARGET: {currentScan?.summary?.target_directory || 'PaymentGatewayService'}
            </div>
            
            <h2 style={{ fontSize: '1.65rem', fontWeight: 800, marginTop: '12px', lineHeight: '1.25', color: '#f8fafc' }}>
              Cryptographic Health & Post-Quantum Risk Assessment
            </h2>

            <p style={{ color: '#94a3b8', marginTop: '10px', fontSize: '0.9rem', lineHeight: '1.5' }}>
              Discovered {metrics.total} cryptographic components across source code, certificates, key stores, AWS KMS, and Docker containers. {metrics.critical + metrics.high} assets require immediate PQC migration.
            </p>
          </div>

          <div style={{ marginTop: '24px', display: 'flex', gap: '14px', flexWrap: 'wrap' }}>
            <button className="btn btn-primary" onClick={onNavigateToMigration}>
              Launch PQC Migration Planner <ArrowRight size={16} />
            </button>
          </div>
        </div>

        {/* Risk Gauge */}
        <div className="glass-card" style={{ padding: '24px', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <RiskGauge score={metrics.score} />
        </div>

      </div>

      {/* Overview Cards */}
      <OverviewCards metrics={metrics} />

      {/* Asset Inventory Table */}
      <AssetTable assets={assets} />

    </div>
  );
}
