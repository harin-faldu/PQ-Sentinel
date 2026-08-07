import React, { useState, useEffect } from 'react';
import { ShieldCheck, Cpu, ArrowRight, Zap, CheckCircle2, RefreshCw } from 'lucide-react';
import { api } from '../api/client';

export function MigrationPlannerPage({ currentScan }) {
  const [plan, setPlan] = useState(null);
  const [loading, setLoading] = useState(true);
  const [provisioningStep, setProvisioningStep] = useState(null);
  const [completedSteps, setCompletedSteps] = useState({});

  useEffect(() => {
    async function loadPlan() {
      if (!currentScan) return;
      try {
        setLoading(true);
        const data = await api.getMigrationPlan(currentScan.summary.id);
        setPlan(data);
      } catch (err) {
        console.error('Failed to load migration plan:', err);
      } finally {
        setLoading(false);
      }
    }
    loadPlan();
  }, [currentScan]);

  const handleSimulateQuantumVaultProvisioning = (stepNumber) => {
    setProvisioningStep(stepNumber);
    setTimeout(() => {
      setCompletedSteps((prev) => ({ ...prev, [stepNumber]: true }));
      setProvisioningStep(null);
    }, 1500);
  };

  if (loading) {
    return (
      <div className="glass-card" style={{ padding: '40px', textAlign: 'center', color: '#94a3b8' }}>
        Calculating Post-Quantum Migration Playbook...
      </div>
    );
  }

  const steps = plan?.steps || [];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      
      {/* Plan Header */}
      <div className="glass-card" style={{ padding: '28px', background: 'linear-gradient(135deg, rgba(99,102,241,0.1), rgba(168,85,247,0.1))' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <span style={{ fontSize: '0.75rem', fontWeight: 800, color: '#a855f7', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              QUANTUMVAULT PQC REMEDIATION ENGINE
            </span>
            <h2 style={{ fontSize: '1.5rem', fontWeight: 800, marginTop: '6px' }}>
              Automated Post-Quantum Migration Plan
            </h2>
            <p style={{ color: '#94a3b8', fontSize: '0.875rem', marginTop: '6px' }}>
              Step-by-step PQC remediation playbooks mapping vulnerable assets to NIST ML-DSA & ML-KEM algorithms.
            </p>
          </div>

          <div style={{ textAlign: 'right' }}>
            <div style={{ fontSize: '1.5rem', fontWeight: 800, color: '#06b6d4', fontFamily: 'var(--font-mono)' }}>
              ~{plan?.estimated_hours || 45} hrs
            </div>
            <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Estimated Migration Effort</span>
          </div>
        </div>
      </div>

      {/* Migration Steps Pipeline */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
        {steps.map((step) => {
          const isDone = completedSteps[step.step_number];
          const isRunning = provisioningStep === step.step_number;

          return (
            <div key={step.step_number} className="glass-card" style={{ padding: '20px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div style={{ display: 'flex', gap: '16px', alignItems: 'flex-start' }}>
                  
                  {/* Step Number Badge */}
                  <div style={{
                    width: '36px',
                    height: '36px',
                    borderRadius: '10px',
                    background: isDone ? 'rgba(16, 185, 129, 0.2)' : 'linear-gradient(135deg, #6366f1, #a855f7)',
                    color: isDone ? '#10b981' : '#ffffff',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontWeight: 800,
                    fontSize: '1rem',
                  }}>
                    {isDone ? <CheckCircle2 size={20} /> : step.step_number}
                  </div>

                  <div>
                    <h3 style={{ fontSize: '1.05rem', fontWeight: 700 }}>
                      {step.title}
                    </h3>
                    <p style={{ fontSize: '0.85rem', color: '#94a3b8', marginTop: '6px', lineHeight: '1.5' }}>
                      {step.description}
                    </p>

                    <div style={{ marginTop: '12px', display: 'flex', gap: '16px', alignItems: 'center', fontSize: '0.8rem' }}>
                      <span style={{ color: '#64748b' }}>Current: <code className="code-inline">{step.current_algo}</code></span>
                      <ArrowRight size={14} color="#64748b" />
                      <span style={{ color: '#a855f7', fontWeight: 700 }}>Target PQC: {step.new_pqc_algo}</span>
                      <span style={{ marginLeft: '12px', background: 'rgba(255,255,255,0.05)', padding: '2px 8px', borderRadius: '4px', color: '#cbd5e1' }}>
                        Effort: {step.effort}
                      </span>
                    </div>
                  </div>

                </div>

                {/* Provision Action Button */}
                <button
                  className={isDone ? 'btn btn-secondary' : 'btn btn-primary'}
                  onClick={() => handleSimulateQuantumVaultProvisioning(step.step_number)}
                  disabled={isRunning || isDone}
                  style={{ fontSize: '0.8rem' }}
                >
                  {isRunning ? <RefreshCw className="spin" size={14} /> : isDone ? <CheckCircle2 size={14} color="#10b981" /> : <Zap size={14} />}
                  {isRunning ? 'Provisioning in QuantumVault...' : isDone ? 'PQC Key Provisioned' : 'Execute 1-Click Remediation'}
                </button>
              </div>
            </div>
          );
        })}
      </div>

    </div>
  );
}
