import React from 'react';
import { ShieldAlert, Cpu, Download, RefreshCw, FolderSearch } from 'lucide-react';
import { api } from '../api/client';

export function Navbar({ activeTab, setActiveTab, currentScan, onScanTrigger, scanning }) {
  const handleDownloadCycloneDX = () => {
    if (!currentScan) return;
    window.open(api.getCycloneDXUrl(currentScan.summary.id), '_blank');
  };

  return (
    <header className="glass-card" style={{ margin: '20px 24px 0 24px', padding: '16px 24px' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        
        {/* Brand Identity */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '14px' }}>
          <div style={{
            background: 'linear-gradient(135deg, #6366f1, #a855f7)',
            padding: '10px',
            borderRadius: '12px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: '0 4px 12px rgba(99, 102, 241, 0.4)',
          }}>
            <Cpu size={24} color="#ffffff" />
          </div>
          <div>
            <h1 style={{ fontSize: '1.25rem', fontWeight: 800, letterSpacing: '-0.02em', background: 'linear-gradient(90deg, #ffffff, #94a3b8)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent' }}>
              QuantumVault <span style={{ color: '#6366f1', WebkitTextFillColor: 'initial' }}>CBOM</span>
            </h1>
            <p style={{ fontSize: '0.75rem', color: '#64748b', fontWeight: 500 }}>
              Cryptographic Inventory & Post-Quantum Intelligence
            </p>
          </div>
        </div>

        {/* Navigation Tabs */}
        <nav style={{ display: 'flex', gap: '6px', background: 'rgba(0,0,0,0.3)', padding: '4px', borderRadius: '12px', border: '1px solid rgba(255,255,255,0.05)' }}>
          {[
            { id: 'dashboard', label: 'Executive Dashboard' },
            { id: 'inventory', label: 'Crypto Inventory' },
            { id: 'risk', label: 'Risk Analysis' },
            { id: 'migration', label: 'PQC Migration Plan' },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              style={{
                padding: '8px 16px',
                borderRadius: '8px',
                border: 'none',
                background: activeTab === tab.id ? 'linear-gradient(135deg, #6366f1, #4f46e5)' : 'transparent',
                color: activeTab === tab.id ? '#ffffff' : '#94a3b8',
                fontWeight: activeTab === tab.id ? 700 : 500,
                fontSize: '0.85rem',
                cursor: 'pointer',
                transition: 'all 0.2s ease',
              }}
            >
              {tab.label}
            </button>
          ))}
        </nav>

        {/* Action Controls */}
        <div style={{ display: 'flex', gap: '12px' }}>
          <button 
            className="btn btn-secondary"
            onClick={onScanTrigger}
            disabled={scanning}
            style={{ fontSize: '0.825rem' }}
          >
            {scanning ? <RefreshCw className="spin" size={16} /> : <FolderSearch size={16} />}
            {scanning ? 'Scanning...' : 'Trigger Scan'}
          </button>

          <button 
            className="btn btn-primary" 
            onClick={handleDownloadCycloneDX}
            disabled={!currentScan}
            style={{ fontSize: '0.825rem' }}
          >
            <Download size={16} /> Export CycloneDX 1.6
          </button>
        </div>

      </div>
    </header>
  );
}
