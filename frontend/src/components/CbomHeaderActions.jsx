import React from 'react';
import { PlusCircle, Download, RefreshCw, Folder } from 'lucide-react';
import Button from './Button';

export function CbomHeaderActions({ currentScan, onOpenIngestion, onOpenExport, scanning }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '10px', flexWrap: 'wrap' }}>
      {/* Active Scan Badge */}
      {currentScan?.summary && (
        <div style={{
          backgroundColor: 'var(--color-bg-tertiary)',
          color: 'var(--color-text-primary)',
          border: '1px solid var(--color-border)',
          borderRadius: '50px',
          padding: '7px 14px',
          fontSize: 'var(--font-size-xs)',
          fontWeight: 'var(--font-weight-medium)',
          display: 'flex',
          alignItems: 'center',
          gap: '6px',
          maxWidth: '260px',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
        }}>
          <Folder size={13} style={{ color: '#38bdf8', flexShrink: 0 }} />
          <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {currentScan.summary.target_directory || currentScan.summary.id}
          </span>
        </div>
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
