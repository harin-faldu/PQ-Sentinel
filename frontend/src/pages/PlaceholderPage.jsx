import React from 'react';
import { Construction } from 'lucide-react';

export function PlaceholderPage({ title = 'Coming Soon' }) {
  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: 'calc(100vh - 120px)',
      gap: '24px',
      textAlign: 'center',
      padding: '40px',
    }}>
      <div style={{
        width: '80px',
        height: '80px',
        borderRadius: '20px',
        background: 'var(--color-bg-tertiary)',
        border: '1px solid var(--color-border)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}>
        <Construction size={36} style={{ color: 'var(--color-primary)' }} />
      </div>

      <div>
        <h1 style={{
          fontSize: '1.75rem',
          fontWeight: 800,
          color: 'var(--color-text-primary)',
          marginBottom: '8px',
        }}>
          {title}
        </h1>
        <p style={{
          color: 'var(--color-text-secondary)',
          fontSize: '0.95rem',
          maxWidth: '400px',
          lineHeight: '1.6',
          margin: '0 auto',
        }}>
          This section is under development. Check back soon for updates.
        </p>
      </div>

      <div style={{
        padding: '12px 24px',
        borderRadius: '50px',
        background: 'var(--color-bg-tertiary)',
        border: '1px solid var(--color-border)',
        color: 'var(--color-text-muted)',
        fontSize: '0.8rem',
        fontWeight: 600,
        letterSpacing: '0.05em',
        textTransform: 'uppercase',
      }}>
        Placeholder
      </div>
    </div>
  );
}
