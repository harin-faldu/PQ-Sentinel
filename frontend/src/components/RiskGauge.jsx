import React from 'react';

export function RiskGauge({ score }) {
  const getScoreColor = (s) => {
    if (s >= 80) return '#10b981'; // Green
    if (s >= 50) return '#eab308'; // Yellow
    return '#ef4444'; // Red
  };

  const color = getScoreColor(score);
  const strokeDashoffset = 440 - (440 * score) / 100;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', position: 'relative' }}>
      <svg width="180" height="180" viewBox="0 0 160 160">
        <circle
          cx="80"
          cy="80"
          r="70"
          stroke="rgba(255,255,255,0.06)"
          strokeWidth="14"
          fill="transparent"
        />
        <circle
          cx="80"
          cy="80"
          r="70"
          stroke={color}
          strokeWidth="14"
          fill="transparent"
          strokeDasharray="440"
          strokeDashoffset={strokeDashoffset}
          strokeLinecap="round"
          style={{ transition: 'stroke-dashoffset 1s ease-in-out', transform: 'rotate(-90deg)', transformOrigin: '50% 50%' }}
        />
      </svg>
      <div style={{ position: 'absolute', textAlign: 'center' }}>
        <span style={{ fontSize: '2.5rem', fontWeight: 800, color, fontFamily: 'var(--font-mono)' }}>
          {score}
        </span>
        <span style={{ fontSize: '0.8rem', color: '#64748b', display: 'block', fontWeight: 600 }}>
          / 100
        </span>
      </div>
      <p style={{ marginTop: '12px', fontSize: '0.85rem', fontWeight: 700, color: '#94a3b8' }}>
        PQC Readiness Index
      </p>
    </div>
  );
}
