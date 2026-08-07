import React from 'react';
import { ShieldAlert, AlertTriangle, ShieldCheck, KeyRound, Layers, FileCode } from 'lucide-react';

export function OverviewCards({ metrics }) {
  const cards = [
    {
      title: 'Critical Vulnerabilities',
      value: metrics.critical,
      subtitle: 'SHA-1 / MD5 / Deprecated TLS',
      icon: ShieldAlert,
      color: '#ef4444',
      bg: 'rgba(239, 68, 68, 0.12)',
    },
    {
      title: 'High Risk (PQC Vulnerable)',
      value: metrics.high,
      subtitle: 'RSA-2048 / ECDSA P-256',
      icon: AlertTriangle,
      color: '#f97316',
      bg: 'rgba(249, 115, 22, 0.12)',
    },
    {
      title: 'Quantum Safe Assets',
      value: metrics.safe,
      subtitle: 'AES-256 / SHA-384 / PQC',
      icon: ShieldCheck,
      color: '#10b981',
      bg: 'rgba(16, 185, 129, 0.12)',
    },
    {
      title: 'Total Cryptographic Assets',
      value: metrics.total,
      subtitle: 'Discovered in Inventory',
      icon: Layers,
      color: '#06b6d4',
      bg: 'rgba(6, 182, 212, 0.12)',
    },
  ];

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '20px' }}>
      {cards.map((card, i) => {
        const Icon = card.icon;
        return (
          <div key={i} className="glass-card" style={{ padding: '20px' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <span style={{ fontSize: '0.85rem', color: '#94a3b8', fontWeight: 600 }}>
                {card.title}
              </span>
              <div style={{ background: card.bg, padding: '8px', borderRadius: '10px', display: 'flex' }}>
                <Icon size={20} color={card.color} />
              </div>
            </div>

            <div style={{ marginTop: '14px', fontSize: '2rem', fontWeight: 800, color: card.color, fontFamily: 'var(--font-mono)' }}>
              {card.value}
            </div>

            <p style={{ marginTop: '4px', fontSize: '0.75rem', color: '#64748b' }}>
              {card.subtitle}
            </p>
          </div>
        );
      })}
    </div>
  );
}
