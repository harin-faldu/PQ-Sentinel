import React, { useState } from 'react';
import { Search, ChevronDown, ChevronUp } from 'lucide-react';
import { SeverityBadge, TypeBadge } from './Badge';

export function AssetTable({ assets = [] }) {
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('ALL');
  const [filterRisk, setFilterRisk] = useState('ALL');
  const [expandedRows, setExpandedRows] = useState({});

  const toggleRow = (id) => {
    setExpandedRows((prev) => ({
      ...prev,
      [id]: !prev[id],
    }));
  };

  const filteredAssets = assets.filter((a) => {
    const matchesSearch = 
      a.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      a.algorithm.toLowerCase().includes(searchTerm.toLowerCase()) ||
      a.location.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (a.details && a.details.toLowerCase().includes(searchTerm.toLowerCase()));

    const matchesType = filterType === 'ALL' || (a.asset_type || a.type) === filterType;
    const matchesRisk = filterRisk === 'ALL' || a.risk_level === filterRisk;
    return matchesSearch && matchesType && matchesRisk;
  });

  return (
    <div>
      {/* Search & Filter Controls (Outside Card, matching QuantaVault screenshot) */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '16px', flexWrap: 'wrap' }}>
        {/* Search Box */}
        <div style={{ position: 'relative', flex: '0 0 240px' }}>
          <Search size={14} style={{ position: 'absolute', left: 12, top: '50%', transform: 'translateY(-50%)', color: 'var(--color-text-muted)' }} />
          <input
            style={{
              width: '100%',
              background: 'var(--color-bg-tertiary)',
              border: '1px solid var(--color-border)',
              borderRadius: '50px',
              padding: '8px 12px 8px 34px',
              color: 'var(--color-text-primary)',
              fontSize: '13px',
              outline: 'none',
              fontFamily: 'inherit',
            }}
            placeholder="Search assets..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>

        {/* Type Filter */}
        <select
          style={{
            background: 'var(--color-bg-tertiary)',
            border: '1px solid var(--color-border)',
            borderRadius: '50px',
            padding: '8px 16px',
            color: 'var(--color-text-primary)',
            fontSize: '13px',
            outline: 'none',
            cursor: 'pointer',
            fontFamily: 'inherit',
          }}
          value={filterType}
          onChange={(e) => setFilterType(e.target.value)}
        >
          <option value="ALL">All Types</option>
          <option value="ALGORITHM">Algorithm</option>
          <option value="CERTIFICATE">Certificate</option>
          <option value="PROTOCOL">Protocol</option>
          <option value="KEYSTORE">KeyStore / Secret</option>
          <option value="LIBRARY">Library</option>
          <option value="SECRET">Secret</option>
          <option value="KEY_MATERIAL">Key Material</option>
        </select>

        {/* Risk Filter */}
        <select
          style={{
            background: 'var(--color-bg-tertiary)',
            border: '1px solid var(--color-border)',
            borderRadius: '50px',
            padding: '8px 16px',
            color: 'var(--color-text-primary)',
            fontSize: '13px',
            outline: 'none',
            cursor: 'pointer',
            fontFamily: 'inherit',
          }}
          value={filterRisk}
          onChange={(e) => setFilterRisk(e.target.value)}
        >
          <option value="ALL">All Severities</option>
          <option value="CRITICAL">Critical</option>
          <option value="HIGH">High</option>
          <option value="MEDIUM">Medium</option>
          <option value="SAFE">Safe</option>
        </select>
      </div>

      {/* Table Card Container */}
      <div className="table-card">
        {/* Table Card Header Title */}
        <div className="table-card-header">
          <h3 className="table-card-title">
            Cryptographic Assets ({filteredAssets.length})
          </h3>
        </div>

        {/* Table */}
        <div className="table-responsive">
          <table className="data-table">
            <thead>
              <tr>
                <th>Asset Name & Location</th>
                <th>Type</th>
                <th>Severity</th>
                <th>Symbol Graph Call-Site Context</th>
                <th style={{ width: '40px', textAlign: 'center' }}></th>
              </tr>
            </thead>
            <tbody>
              {filteredAssets.length === 0 ? (
                <tr>
                  <td colSpan="5" style={{ textAlign: 'center', padding: '40px', color: 'var(--color-text-muted)' }}>
                    No cryptographic assets match the selected search criteria.
                  </td>
                </tr>
              ) : (
                filteredAssets.map((asset) => {
                  const isExpanded = !!expandedRows[asset.id];
                  return (
                    <React.Fragment key={asset.id}>
                      {/* Main Row */}
                      <tr
                        onClick={() => toggleRow(asset.id)}
                        style={{
                          cursor: 'pointer',
                          transition: 'background 0.15s ease',
                          background: isExpanded ? 'var(--color-bg-tertiary)' : 'transparent',
                        }}
                      >
                        {/* Asset Name & Location */}
                        <td>
                          <div style={{ fontWeight: 600, color: 'var(--color-text-primary)', fontSize: '0.9rem' }}>
                            {asset.name}
                          </div>
                          <div style={{ color: 'var(--color-primary)', fontSize: '0.75rem', fontFamily: 'var(--font-family-mono)', marginTop: '3px' }}>
                            {asset.location}:{asset.line_number}
                          </div>
                        </td>

                        {/* Type Badge (Rectangular Box) */}
                        <td>
                          <TypeBadge type={asset.asset_type || asset.type || 'ALGORITHM'} />
                        </td>

                        {/* Severity Badge (Oval Pill) */}
                        <td>
                          <SeverityBadge level={asset.risk_level} />
                        </td>

                        {/* Symbol Graph Call-Site Context */}
                        <td style={{ color: 'var(--color-text-secondary)', fontSize: '0.85rem', maxWidth: '380px', lineHeight: '1.4' }}>
                          {asset.risk_rationale || asset.details || 'Discovered via static analysis'}
                        </td>

                        {/* Expand Chevron Icon */}
                        <td style={{ textAlign: 'center', color: 'var(--color-text-muted)' }}>
                          {isExpanded ? <ChevronUp size={18} /> : <ChevronDown size={18} />}
                        </td>
                      </tr>

                      {/* Expanded Detail Panel */}
                      {isExpanded && (
                        <tr key={`${asset.id}-expanded`}>
                          <td colSpan="5" style={{ padding: 0, borderBottom: '1px solid var(--color-border)' }}>
                            <div style={{
                              padding: '20px 24px',
                              background: 'var(--color-bg-tertiary)',
                              borderTop: '1px dashed var(--color-border)',
                              display: 'flex',
                              flexDirection: 'column',
                              gap: '16px',
                            }}>
                              {/* Key Asset Metrics Header */}
                              <div style={{ display: 'flex', gap: '24px', alignItems: 'center', flexWrap: 'wrap' }}>
                                <div>
                                  <span style={{ fontSize: '0.7rem', color: 'var(--color-text-muted)', display: 'block', marginBottom: '4px', textTransform: 'uppercase', fontWeight: 700, letterSpacing: '0.05em' }}>
                                    Algorithm & Key Size
                                  </span>
                                  <span style={{ fontWeight: 700, color: 'var(--color-primary)', fontSize: '0.9rem', fontFamily: 'var(--font-family-mono)' }}>
                                    {asset.algorithm} {asset.key_size > 0 ? `(${asset.key_size}-bit)` : ''}
                                  </span>
                                </div>

                                <div style={{ height: '24px', width: '1px', background: 'var(--color-border)' }} />

                                <div>
                                  <span style={{ fontSize: '0.7rem', color: 'var(--color-text-muted)', display: 'block', marginBottom: '4px', textTransform: 'uppercase', fontWeight: 700, letterSpacing: '0.05em' }}>
                                    PQC Target Replacement
                                  </span>
                                  <span style={{ fontWeight: 700, color: '#a855f7', fontSize: '0.9rem' }}>
                                    {asset.pqc_replacement || 'NIST ML-DSA-65 (FIPS 204)'}
                                  </span>
                                </div>

                                <div style={{ height: '24px', width: '1px', background: 'var(--color-border)' }} />

                                <div>
                                  <span style={{ fontSize: '0.7rem', color: 'var(--color-text-muted)', display: 'block', marginBottom: '4px', textTransform: 'uppercase', fontWeight: 700, letterSpacing: '0.05em' }}>
                                    Compliance & Standards
                                  </span>
                                  <span style={{ fontWeight: 600, color: 'var(--color-success)', fontSize: '0.85rem' }}>
                                    NIST SP 800-131A & CNSA 2.0 PQC Standard
                                  </span>
                                </div>
                              </div>

                              {/* Full Rationale & Symbol Graph Details */}
                              <div>
                                <span style={{ fontSize: '0.7rem', color: 'var(--color-text-muted)', display: 'block', marginBottom: '4px', textTransform: 'uppercase', fontWeight: 700, letterSpacing: '0.05em' }}>
                                  Symbol Graph Context & Risk Rationale
                                </span>
                                <p style={{ margin: 0, fontSize: '0.85rem', color: 'var(--color-text-secondary)', lineHeight: '1.5' }}>
                                  {asset.risk_rationale || 'Component evaluated for post-quantum vulnerability against Shor\'s and Grover\'s quantum cryptanalysis algorithms.'}
                                </p>
                              </div>

                              {/* Matched Consumer Code Snippet */}
                              {asset.details && (
                                <div>
                                  <span style={{ fontSize: '0.7rem', color: 'var(--color-text-muted)', display: 'block', marginBottom: '4px', textTransform: 'uppercase', fontWeight: 700, letterSpacing: '0.05em' }}>
                                    Consumer Call Site & Source Code Details
                                  </span>
                                  <div style={{
                                    background: 'var(--color-bg-primary)',
                                    border: '1px solid var(--color-border)',
                                    borderRadius: '8px',
                                    padding: '12px 16px',
                                    fontFamily: 'var(--font-family-mono)',
                                    fontSize: '0.8rem',
                                    color: 'var(--color-primary)',
                                    whiteSpace: 'pre-wrap',
                                    wordBreak: 'break-all',
                                  }}>
                                    {asset.location}:{asset.line_number} → {asset.details}
                                  </div>
                                </div>
                              )}
                            </div>
                          </td>
                        </tr>
                      )}
                    </React.Fragment>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
