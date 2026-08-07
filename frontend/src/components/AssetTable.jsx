import React, { useState } from 'react';
import { Search, Filter, ShieldCheck, ArrowRightLeft, HelpCircle } from 'lucide-react';
import { getRiskBadgeClass } from '../utils/formatters';

export function AssetTable({ assets = [] }) {
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('ALL');

  const filteredAssets = assets.filter((a) => {
    const matchesSearch = 
      a.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      a.algorithm.toLowerCase().includes(searchTerm.toLowerCase()) ||
      a.location.toLowerCase().includes(searchTerm.toLowerCase());

    const matchesType = filterType === 'ALL' || a.asset_type === filterType;
    return matchesSearch && matchesType;
  });

  return (
    <div className="glass-card" style={{ padding: '24px' }}>
      {/* Search & Filter Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
        <div>
          <h2 style={{ fontSize: '1.1rem', fontWeight: 700 }}>Cryptographic Asset Inventory</h2>
          <p style={{ fontSize: '0.8rem', color: '#64748b' }}>
            Showing {filteredAssets.length} of {assets.length} discovered assets
          </p>
        </div>

        <div style={{ display: 'flex', gap: '12px' }}>
          {/* Search Box */}
          <div style={{ position: 'relative' }}>
            <Search size={16} color="#64748b" style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)' }} />
            <input
              type="text"
              placeholder="Search algorithm, file, location..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              style={{
                background: 'rgba(0,0,0,0.4)',
                border: '1px solid rgba(255,255,255,0.1)',
                padding: '8px 12px 8px 36px',
                borderRadius: '8px',
                color: '#fff',
                fontSize: '0.85rem',
                outline: 'none',
                width: '260px',
              }}
            />
          </div>

          {/* Filter Dropdown */}
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
            style={{
              background: 'rgba(0,0,0,0.4)',
              border: '1px solid rgba(255,255,255,0.1)',
              padding: '8px 14px',
              borderRadius: '8px',
              color: '#fff',
              fontSize: '0.85rem',
              outline: 'none',
              cursor: 'pointer',
            }}
          >
            <option value="ALL">All Types</option>
            <option value="ALGORITHM">Algorithm</option>
            <option value="CERTIFICATE">Certificate</option>
            <option value="PROTOCOL">Protocol</option>
            <option value="KEYSTORE">KeyStore / Secret</option>
            <option value="LIBRARY">Library</option>
          </select>
        </div>
      </div>

      {/* Asset Table */}
      <div style={{ overflowX: 'auto' }}>
        <table className="custom-table">
          <thead>
            <tr>
              <th>Asset Name & Location</th>
              <th>Type</th>
              <th>Algorithm</th>
              <th>Risk Rating</th>
              <th>Risk Rationale (Why)</th>
              <th>PQC Target</th>
            </tr>
          </thead>
          <tbody>
            {filteredAssets.length === 0 ? (
              <tr>
                <td colSpan="6" style={{ textAlign: 'center', padding: '32px', color: '#64748b' }}>
                  No cryptographic assets match the query.
                </td>
              </tr>
            ) : (
              filteredAssets.map((asset) => (
                <tr key={asset.id}>
                  <td>
                    <div style={{ fontWeight: 600, color: '#f8fafc' }}>{asset.name}</div>
                    <div style={{ color: '#94a3b8', fontSize: '0.75rem', fontFamily: 'var(--font-mono)', marginTop: '2px' }}>
                      {asset.location}:{asset.line_number}
                    </div>
                  </td>
                  <td>
                    <span className="code-inline">{asset.asset_type}</span>
                  </td>
                  <td style={{ fontWeight: 600, color: '#06b6d4' }}>
                    {asset.algorithm} {asset.key_size > 0 ? `(${asset.key_size}-bit)` : ''}
                  </td>
                  <td>
                    <span className={`badge ${getRiskBadgeClass(asset.risk_level)}`}>
                      {asset.risk_level}
                    </span>
                  </td>
                  <td style={{ color: '#cbd5e1', fontSize: '0.8rem', maxWidth: '320px', lineHeight: '1.4' }}>
                    {asset.risk_rationale || asset.details}
                  </td>
                  <td style={{ color: '#a855f7', fontWeight: 600, fontSize: '0.825rem' }}>
                    {asset.pqc_replacement}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
