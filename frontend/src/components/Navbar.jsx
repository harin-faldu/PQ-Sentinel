import React from 'react';
import { Shield, LayoutDashboard, Database, AlertTriangle, GitPullRequest, PlusCircle, Download, RefreshCw } from 'lucide-react';

export function Navbar({ activeTab, setActiveTab, currentScan, scans, targets, onSelectScan, onOpenIngestion, onOpenExport, scanning }) {
  return (
    <header style={{
      backgroundColor: '#0f172a',
      borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
      position: 'sticky',
      top: 0,
      zIndex: 100,
      boxShadow: '0 4px 20px -2px rgba(0, 0, 0, 0.4)',
    }}>
      <div className="navbar-container">
        {/* Brand Logo & Name */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <div style={{
            width: '38px',
            height: '38px',
            borderRadius: '10px',
            background: 'linear-gradient(135deg, #0284c7 0%, #0d9488 100%)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
            boxShadow: '0 0 15px rgba(2, 132, 199, 0.4)',
          }}>
            <Shield style={{ color: '#ffffff', width: '22px', height: '22px' }} />
          </div>
          <div>
            <div style={{ fontSize: '18px', fontWeight: '800', color: '#f8fafc', letterSpacing: '-0.5px' }}>
              PQ-Sentinel <span style={{ color: '#38bdf8' }}>CBOM</span>
            </div>
            <div style={{ fontSize: '11px', color: '#64748b', fontWeight: '500' }}>
              Post-Quantum Cryptographic Intelligence Platform
            </div>
          </div>
        </div>

        {/* Navigation Tabs - Responsive Scroll Container */}
        <div className="nav-tabs-wrapper">
          <nav style={{
            display: 'flex',
            gap: '4px',
            backgroundColor: 'rgba(30, 41, 59, 0.5)',
            padding: '4px',
            borderRadius: '10px',
            border: '1px solid rgba(255, 255, 255, 0.05)',
            width: 'max-content',
          }}>
            {[
              { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
              { id: 'inventory', label: 'Asset Inventory', icon: Database },
              { id: 'risk', label: 'Risk Analysis', icon: AlertTriangle },
              { id: 'migration', label: 'Migration Strategy', icon: GitPullRequest },
            ].map((item) => {
              const Icon = item.icon;
              const isActive = activeTab === item.id;
              return (
                <button
                  key={item.id}
                  onClick={() => setActiveTab(item.id)}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '8px',
                    padding: '8px 14px',
                    borderRadius: '7px',
                    border: 'none',
                    backgroundColor: isActive ? '#0284c7' : 'transparent',
                    color: isActive ? '#ffffff' : '#94a3b8',
                    fontSize: '13px',
                    fontWeight: '600',
                    cursor: 'pointer',
                    whiteSpace: 'nowrap',
                    transition: 'all 0.15s ease',
                  }}
                >
                  <Icon style={{ width: '16px', height: '16px' }} />
                  {item.label}
                </button>
              );
            })}
          </nav>
        </div>

        {/* Action Controls & Selectors */}
        <div className="nav-actions-wrapper" style={{ display: 'flex', alignItems: 'center', gap: '10px', flexWrap: 'wrap' }}>
          {/* Scan Selector */}
          {scans && scans.length > 0 && (
            <select
              value={currentScan?.summary?.id || ''}
              onChange={(e) => onSelectScan(e.target.value)}
              style={{
                backgroundColor: '#1e293b',
                color: '#f8fafc',
                border: '1px solid rgba(255, 255, 255, 0.15)',
                borderRadius: '8px',
                padding: '8px 12px',
                fontSize: '12px',
                fontWeight: '600',
                outline: 'none',
                cursor: 'pointer',
                maxWidth: '180px',
                textOverflow: 'ellipsis',
              }}
            >
              {scans.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.target_directory || s.id} ({new Date(s.created_at).toLocaleTimeString()})
                </option>
              ))}
            </select>
          )}

          {/* Ingest Codebase Button */}
          <button
            onClick={onOpenIngestion}
            disabled={scanning}
            className="btn btn-primary"
            style={{
              padding: '8px 14px',
              fontSize: '13px',
              backgroundColor: scanning ? '#334155' : '#0284c7',
            }}
          >
            {scanning ? <RefreshCw className="animate-spin" style={{ width: '16px', height: '16px' }} /> : <PlusCircle style={{ width: '16px', height: '16px' }} />}
            {scanning ? 'Scanning...' : 'Scan Codebase'}
          </button>

          {/* CycloneDX Export Button */}
          {currentScan && (
            <button
              onClick={onOpenExport}
              className="btn btn-secondary"
              style={{
                padding: '8px 12px',
                fontSize: '13px',
                color: '#38bdf8',
              }}
            >
              <Download style={{ width: '15px', height: '15px' }} />
              CycloneDX 1.6
            </button>
          )}
        </div>
      </div>
    </header>
  );
}
