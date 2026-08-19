import React, { useState } from 'react';
import { Folder, RefreshCw, X, ArrowRight, ShieldCheck, Globe, HardDrive } from 'lucide-react';

export function IngestionModal({ isOpen, onClose, onStartLocal, scanning }) {
  const [localPath, setLocalPath] = useState('sample_target');
  const [targetName, setTargetName] = useState('');
  const [tlsTargets, setTlsTargets] = useState('');

  if (!isOpen) return null;

  const handleLocalSubmit = async (e) => {
    e.preventDefault();
    if (!localPath.trim()) return;
    try {
      const tlsList = tlsTargets.split(',').map(t => t.trim()).filter(Boolean);
      await onStartLocal(localPath.trim(), targetName.trim() || localPath.trim(), tlsList);
      onClose();
    } catch (err) {
      // Error handled in parent hook
    }
  };

  return (
    <div style={{
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor: 'rgba(15, 23, 42, 0.8)',
      backdropFilter: 'blur(8px)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 1000,
      padding: '24px',
    }}>
      <div className="glass-card" style={{
        width: '100%',
        maxWidth: 'min(620px, 95vw)',
        maxHeight: '92vh',
        overflowY: 'auto',
        borderRadius: '16px',
        border: '1px solid rgba(255, 255, 255, 0.15)',
        backgroundColor: '#0f172a',
        color: '#f8fafc',
        boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)',
      }}>
        {/* Header */}
        <div style={{
          padding: '20px 24px',
          borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}>
          <div>
            <h2 style={{ fontSize: '20px', fontWeight: '700', margin: 0, color: '#f8fafc', display: 'flex', alignItems: 'center', gap: '10px' }}>
              <Folder style={{ color: '#38bdf8', width: '22px', height: '22px' }} />
              Scan Local Directory Codebase
            </h2>
            <p style={{ fontSize: '13px', color: '#94a3b8', margin: '4px 0 0 0' }}>
              Analyze source code ASTs, certificates, keystores, and container definitions
            </p>
          </div>
          <button onClick={onClose} style={{
            background: 'none',
            border: 'none',
            color: '#64748b',
            cursor: 'pointer',
            padding: '4px',
            borderRadius: '6px',
          }}>
            <X style={{ width: '20px', height: '20px' }} />
          </button>
        </div>

        {/* Modal Body */}
        <div style={{ padding: '24px' }}>
          <form onSubmit={handleLocalSubmit}>
            {/* Target Path */}
            <div style={{ marginBottom: '18px' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '13px', fontWeight: '600', color: '#cbd5e1', marginBottom: '8px' }}>
                <HardDrive style={{ width: '15px', height: '15px', color: '#38bdf8' }} />
                Target Directory Path <span style={{ color: '#ef4444' }}>*</span>
              </label>
              <input
                type="text"
                value={localPath}
                onChange={(e) => setLocalPath(e.target.value)}
                placeholder="e.g. sample_target or C:\projects\payment-api"
                required
                autoFocus
                style={{
                  width: '100%',
                  padding: '11px 14px',
                  borderRadius: '8px',
                  border: '1px solid rgba(255, 255, 255, 0.15)',
                  backgroundColor: '#1e293b',
                  color: '#f8fafc',
                  fontSize: '14px',
                  outline: 'none',
                  boxSizing: 'border-box',
                }}
              />
              <span style={{ fontSize: '12px', color: '#64748b', marginTop: '5px', display: 'block' }}>
                Scans JS/TS, Python, Go, Java, C/C++, Rust, X.509 certs, keystores, and secrets.
              </span>
            </div>

            {/* Target Display Name */}
            <div style={{ marginBottom: '18px' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '13px', fontWeight: '600', color: '#cbd5e1', marginBottom: '8px' }}>
                <ShieldCheck style={{ width: '15px', height: '15px', color: '#38bdf8' }} />
                Target Display Name (Optional)
              </label>
              <input
                type="text"
                value={targetName}
                onChange={(e) => setTargetName(e.target.value)}
                placeholder="e.g. Payment Gateway Service"
                style={{
                  width: '100%',
                  padding: '11px 14px',
                  borderRadius: '8px',
                  border: '1px solid rgba(255, 255, 255, 0.15)',
                  backgroundColor: '#1e293b',
                  color: '#f8fafc',
                  fontSize: '14px',
                  outline: 'none',
                  boxSizing: 'border-box',
                }}
              />
            </div>

            {/* TLS Targets */}
            <div style={{ marginBottom: '24px' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '13px', fontWeight: '600', color: '#cbd5e1', marginBottom: '8px' }}>
                <Globe style={{ width: '15px', height: '15px', color: '#38bdf8' }} />
                Live TLS / Network Hosts (Optional)
              </label>
              <input
                type="text"
                value={tlsTargets}
                onChange={(e) => setTlsTargets(e.target.value)}
                placeholder="e.g. api.example.com:443, db.example.com:8443"
                style={{
                  width: '100%',
                  padding: '11px 14px',
                  borderRadius: '8px',
                  border: '1px solid rgba(255, 255, 255, 0.15)',
                  backgroundColor: '#1e293b',
                  color: '#f8fafc',
                  fontSize: '14px',
                  outline: 'none',
                  boxSizing: 'border-box',
                }}
              />
              <span style={{ fontSize: '12px', color: '#64748b', marginTop: '5px', display: 'block' }}>
                Optional live host:port endpoints for real-time TLS handshakes and cipher evaluation.
              </span>
            </div>

            {/* Action Buttons */}
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px', paddingTop: '12px', borderTop: '1px solid rgba(255, 255, 255, 0.08)' }}>
              <button
                type="button"
                onClick={onClose}
                style={{
                  padding: '10px 18px',
                  borderRadius: '8px',
                  border: '1px solid rgba(255, 255, 255, 0.1)',
                  backgroundColor: 'transparent',
                  color: '#94a3b8',
                  cursor: 'pointer',
                  fontWeight: '600',
                  fontSize: '14px',
                }}
              >
                Cancel
              </button>

              <button
                type="submit"
                disabled={scanning || !localPath.trim()}
                style={{
                  padding: '10px 22px',
                  borderRadius: '8px',
                  border: 'none',
                  backgroundColor: scanning || !localPath.trim() ? '#334155' : '#0284c7',
                  color: '#ffffff',
                  fontWeight: '600',
                  fontSize: '14px',
                  cursor: scanning || !localPath.trim() ? 'not-allowed' : 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '8px',
                  boxShadow: '0 4px 12px rgba(2, 132, 199, 0.3)',
                }}
              >
                {scanning ? (
                  <>
                    <RefreshCw className="animate-spin" style={{ width: '16px', height: '16px' }} />
                    Scanning Codebase...
                  </>
                ) : (
                  <>
                    Start Scan
                    <ArrowRight style={{ width: '16px', height: '16px' }} />
                  </>
                )}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
