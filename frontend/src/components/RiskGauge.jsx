import React from 'react';

export function RiskGauge({ score }) {
  const getScoreColor = (s) => {
    if (s >= 80) return 'var(--color-success)'; // Green
    if (s >= 50) return 'var(--color-warning)'; // Yellow
    return 'var(--color-error)'; // Red
  };

  const color = getScoreColor(score);
  const strokeDashoffset = 440 - (440 * score) / 100;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center' }}>
      {/* Circle & Centered Text Container */}
      <div style={{ position: 'relative', width: '180px', height: '180px', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <svg width="180" height="180" viewBox="0 0 160 160">
          <circle
            cx="80"
            cy="80"
            r="70"
            stroke="var(--color-border)"
            strokeWidth="12"
            fill="transparent"
          />
          <circle
            cx="80"
            cy="80"
            r="70"
            stroke={color}
            strokeWidth="12"
            fill="transparent"
            strokeDasharray="440"
            strokeDashoffset={strokeDashoffset}
            strokeLinecap="round"
            style={{
              transition: 'stroke-dashoffset 1s ease-in-out',
              transform: 'rotate(-90deg)',
              transformOrigin: '50% 50%',
            }}
          />
        </svg>

        {/* Perfectly Centered Score & Sublabel */}
        <div style={{
          position: 'absolute',
          top: '50%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          textAlign: 'center',
          lineHeight: 1.1,
        }}>
          <span style={{ fontSize: '2.5rem', fontWeight: 800, color, fontFamily: 'var(--font-family-mono)' }}>
            {score}
          </span>
          <span style={{ fontSize: '0.8rem', color: 'var(--color-text-muted)', fontWeight: 600, marginTop: '2px' }}>
            / 100
          </span>
        </div>
      </div>

      {/* Title Below Gauge */}
      <p style={{ marginTop: '14px', fontSize: '0.875rem', fontWeight: 600, color: 'var(--color-text-secondary)', margin: '14px 0 0 0' }}>
        PQC Readiness Index
      </p>
    </div>
  );
}
