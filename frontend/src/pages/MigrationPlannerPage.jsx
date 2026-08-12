import React, { useState, useEffect } from 'react';
import { ShieldCheck, Cpu, ArrowRight, Zap, CheckCircle2, RefreshCw } from 'lucide-react';
import { api } from '../api/client';
import Card from '../components/Card';
import Button from '../components/Button';

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
      <Card style={{ padding: '40px', textAlign: 'center', color: 'var(--color-text-secondary)' }}>
        Calculating Post-Quantum Migration Playbook...
      </Card>
    );
  }

  const steps = plan?.steps || [];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      
      {/* Plan Summary Card */}
      <Card style={{ padding: '24px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '16px' }}>
          <div>
            <span style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--color-primary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              QUANTUMVAULT PQC REMEDIATION ENGINE
            </span>
            <h2 style={{ fontSize: '1.4rem', fontWeight: 800, marginTop: '6px' }}>
              Automated Post-Quantum Migration Plan
            </h2>
            <p style={{ color: 'var(--color-text-secondary)', fontSize: '0.875rem', marginTop: '6px' }}>
              Step-by-step PQC remediation playbooks mapping vulnerable assets to NIST ML-DSA & ML-KEM algorithms.
            </p>
          </div>

          <div>
            <div style={{ fontSize: '1.5rem', fontWeight: 800, color: 'var(--color-primary)', fontFamily: 'var(--font-family-mono)' }}>
              ~{plan?.estimated_hours || 45} hrs
            </div>
            <span style={{ fontSize: '0.75rem', color: 'var(--color-text-muted)' }}>Estimated Migration Effort</span>
          </div>
        </div>
      </Card>

      {/* Migration Steps Pipeline */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
        {steps.map((step) => {
          const isDone = completedSteps[step.step_number];
          const isRunning = provisioningStep === step.step_number;

          return (
            <Card key={step.step_number} style={{ padding: '20px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '16px' }}>
                <div style={{ display: 'flex', gap: '16px', alignItems: 'flex-start', flex: 1, minWidth: '280px' }}>
                  
                  {/* Step Number Badge */}
                  <div style={{
                    width: '36px',
                    height: '36px',
                    borderRadius: '50%',
                    background: isDone ? 'rgba(0, 208, 132, 0.2)' : 'var(--color-primary)',
                    color: isDone ? 'var(--color-success)' : '#ffffff',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontWeight: 700,
                    fontSize: '0.95rem',
                    flexShrink: 0,
                  }}>
                    {isDone ? <CheckCircle2 size={20} /> : step.step_number}
                  </div>

                  <div>
                    <h3 style={{ fontSize: '1.05rem', fontWeight: 700 }}>
                      {step.title}
                    </h3>
                    <p style={{ color: 'var(--color-text-secondary)', fontSize: '0.85rem', marginTop: '4px', lineHeight: '1.5' }}>
                      {step.description}
                    </p>

                    <div style={{ marginTop: '12px', display: 'flex', gap: '12px', flexWrap: 'wrap', alignItems: 'center', fontSize: '0.8rem' }}>
                      <span style={{ color: 'var(--color-text-muted)' }}>
                        Target: <code className="code-inline">{step.target_asset}</code>
                      </span>
                      <span style={{ color: 'var(--color-text-muted)' }}>
                        PQC Standard: <strong style={{ color: 'var(--color-primary)' }}>{step.pqc_standard}</strong>
                      </span>
                    </div>

                    {/* Remediation Snippet */}
                    {step.suggested_code_diff && (
                      <div style={{ marginTop: '12px', background: 'var(--color-bg-tertiary)', padding: '12px', borderRadius: 'var(--radius-md)', border: '1px solid var(--color-border)', fontFamily: 'var(--font-family-mono)', fontSize: '0.8rem', color: 'var(--color-text-primary)' }}>
                        <div style={{ fontSize: '0.7rem', color: 'var(--color-text-muted)', marginBottom: '6px', fontWeight: 700, textTransform: 'uppercase' }}>
                          Suggested Code Replacement:
                        </div>
                        <pre style={{ margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                          {step.suggested_code_diff}
                        </pre>
                      </div>
                    )}
                  </div>
                </div>

                {/* Provision Action */}
                <div style={{ flexShrink: 0 }}>
                  <Button
                    variant={isDone ? "secondary" : "primary"}
                    size="small"
                    onClick={() => handleSimulateQuantumVaultProvisioning(step.step_number)}
                    disabled={isRunning || isDone}
                  >
                    {isRunning ? (
                      <>
                        <RefreshCw className="animate-spin" size={14} /> Provisioning PQC Key...
                      </>
                    ) : isDone ? (
                      <>
                        <CheckCircle2 size={14} color="var(--color-success)" /> Provisioned in QuantaVault
                      </>
                    ) : (
                      <>
                        <Zap size={14} /> Auto-Provision PQC Key
                      </>
                    )}
                  </Button>
                </div>
              </div>
            </Card>
          );
        })}
      </div>

    </div>
  );
}
