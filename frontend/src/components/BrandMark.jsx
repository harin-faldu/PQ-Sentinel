import React from 'react';
import { Shield } from 'lucide-react';

export function BrandMark({ compact = false }) {
  return (
    <div className={`brand-mark${compact ? ' brand-mark--compact' : ''}`}>
      <div className="brand-mark-icon" aria-hidden="true">
        <Shield size={compact ? 16 : 20} strokeWidth={2.25} />
      </div>
      <div className="brand-mark-copy">
        <div className="brand-mark-name">PQ-Sentinel</div>
        {!compact && <div className="brand-mark-sub">Cryptographic Intelligence</div>}
      </div>
    </div>
  );
}
