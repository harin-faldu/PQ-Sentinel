import React from 'react';
import { PlusCircle, Download, RefreshCw } from 'lucide-react';
import Button from './Button';

export function CbomHeaderActions({ currentScan, scans, onSelectScan, onOpenIngestion, onOpenExport, scanning }) {
  return (
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
            borderRadius: '50px',
            padding: '8px 16px',
            fontSize: 'var(--font-size-xs)',
            fontWeight: 'var(--font-weight-medium)',
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

      {/* Scan Codebase Button */}
      <Button
        variant="primary"
        size="small"
        onClick={onOpenIngestion}
        disabled={scanning}
      >
        {scanning
          ? <RefreshCw className="animate-spin" size={14} />
          : <PlusCircle size={14} />
        }
        {scanning ? 'Scanning...' : 'Scan Codebase'}
      </Button>

      {/* CycloneDX Export Button */}
      {currentScan && (
        <Button
          variant="secondary"
          size="small"
          onClick={onOpenExport}
        >
          <Download size={14} />
          CycloneDX 1.6
        </Button>
      )}
    </div>
  );
}
