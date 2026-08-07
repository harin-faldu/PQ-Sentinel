import React from 'react';
import { AssetTable } from '../components/AssetTable';

export function InventoryPage({ currentScan }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
      <AssetTable assets={currentScan?.assets || []} />
    </div>
  );
}
