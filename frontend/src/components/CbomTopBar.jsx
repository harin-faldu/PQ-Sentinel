import React from 'react';
import { PlusCircle, Download, RefreshCw } from 'lucide-react';

export function CbomTopBar({ currentScan, scans, targets, onSelectScan, onOpenIngestion, onOpenExport, scanning }) {
  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: '12px',
      marginBottom: '20px',
      flexWrap: 'wrap',
    }}>
      {/* Left: Title */}
      <div>
        <h1 style={{
          fontSize: '1.5rem',
          fontWeight: 800,
          color: 'var(--color-text-primary)',
          lineHeight: 1.2,
        }}>
          PQ-Sentinel <span style={{ color: 'var(--color-primary)' }}>Intelligence</span>
        </h1>
        <p style={{
          color: 'var(--color-text-muted)',
          fontSize: '0.8rem',
          marginTop: '2px',
          marginBottom: 0,
        }}>
          Post-Quantum Cryptographic Bill of Materials
        </p>
      </div>

      {/* Right: Action Controls */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '10px', flexWrap: 'wrap' }}>
        {/* Scan Selector */}
        {scans && scans.length > 0 && (
          <select
            value={currentScan?.summary?.id || ''}
            onChange={(e) => onSelectScan(e.target.value)}
            style={{
              backgroundColor: 'var(--color-bg-tertiary)',
              color: 'var(--color-text-primary)',
              border: '1px solid var(--color-border)',
              borderRadius: '8px',
              padding: '8px 12px',
              fontSize: '12px',
              fontWeight: '600',
              outline: 'none',
              cursor: 'pointer',
              maxWidth: '200px',
              textOverflow: 'ellipsis',
              fontFamily: 'inherit',
            }}
          >
            {scans.map((s) => (
              <option key={s.id} value={s.id}>
                {s.target_directory || s.id} ({new Date(s.created_at).toLocaleTimeString()})
              </option>
            ))}
          </select>
        )}

        {/* Scan Button */}
        <button
          onClick={onOpenIngestion}
          disabled={scanning}
          className="btn btn-primary"
          style={{
            padding: '8px 16px',
            fontSize: '13px',
          }}
        >
          {scanning
            ? <RefreshCw className="animate-spin" style={{ width: '16px', height: '16px' }} />
            : <PlusCircle style={{ width: '16px', height: '16px' }} />
          }
          {scanning ? 'Scanning...' : 'Scan Codebase'}
        </button>

        {/* Export Button */}
        {currentScan && (
          <button
            onClick={onOpenExport}
            className="btn btn-secondary"
            style={{
              padding: '8px 14px',
              fontSize: '13px',
              color: 'var(--color-primary)',
            }}
          >
            <Download style={{ width: '15px', height: '15px' }} />
            CycloneDX 1.6
          </button>
        )}
      </div>
    </div>
  );
}
