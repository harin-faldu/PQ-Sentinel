import React from 'react';
import { ShieldAlert, AlertTriangle, ShieldCheck, Layers } from 'lucide-react';
import Card from './Card';

export function OverviewCards({ metrics }) {
  const cards = [
    {
      title: 'CRITICAL VULNERABILITIES',
      value: metrics.critical,
      subtitle: 'SHA-1 / MD5 / Deprecated TLS',
      icon: ShieldAlert,
    },
    {
      title: 'HIGH RISK (PQC VULNERABLE)',
      value: metrics.high,
      subtitle: 'RSA-2048 / ECDSA P-256',
      icon: AlertTriangle,
    },
    {
      title: 'QUANTUM SAFE ASSETS',
      value: metrics.safe,
      subtitle: 'AES-256 / SHA-384 / PQC',
      icon: ShieldCheck,
    },
    {
      title: 'TOTAL CRYPTO ASSETS',
      value: metrics.total,
      subtitle: 'Discovered in Inventory',
      icon: Layers,
    },
  ];

  return (
    <div className="metrics-grid">
      {cards.map((card, i) => {
        const Icon = card.icon;
        return (
          <Card key={i} className="metric-card">
            <div className="metric-header">
              <span className="metric-label">{card.title}</span>
              <Icon size={18} className="metric-icon" />
            </div>
            <div className="metric-value">{card.value}</div>
            <div className="metric-sublabel">{card.subtitle}</div>
          </Card>
        );
      })}
    </div>
  );
}
