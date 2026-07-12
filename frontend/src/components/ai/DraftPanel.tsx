import React, { useState } from 'react';
import { aiService, type AIDraftResponse } from '../../services/ai.service';
import { toast } from 'react-hot-toast';

interface DraftPanelProps {
  orderId: number;
}

export const DraftPanel: React.FC<DraftPanelProps> = ({ orderId }) => {
  const [channel, setChannel] = useState('email');
  const [tone, setTone] = useState('neutral');
  const [draft, setDraft] = useState<AIDraftResponse | null>(null);
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleGenerate = async () => {
    setIsGenerating(true);
    setError(null);
    try {
      const response = await aiService.generateDraft({ order_id: orderId, channel, tone });
      setDraft(response);
    } catch (err: any) {
      setError(err.message || 'Failed to generate draft');
    } finally {
      setIsGenerating(false);
    }
  };

  const handleCopy = () => {
    if (draft?.draft_message) {
      navigator.clipboard.writeText(draft.draft_message);
      toast.success('Draft copied to clipboard!');
    }
  };

  return (
    <div className="greek-card-static mb-4">
      <h3 className="greek-subheading" style={{ marginBottom: '16px' }}>AI Customer Message Draft</h3>

      <div className="flex gap-4 flex-wrap items-end" style={{ marginBottom: '16px' }}>
        <div>
          <label className="greek-label">Channel</label>
          <select className="greek-input greek-select" style={{ minWidth: '160px' }} value={channel} onChange={e => setChannel(e.target.value)}>
            <option value="email">Email</option>
            <option value="sms">SMS</option>
          </select>
        </div>
        <div>
          <label className="greek-label">Tone</label>
          <select className="greek-input greek-select" style={{ minWidth: '160px' }} value={tone} onChange={e => setTone(e.target.value)}>
            <option value="neutral">Neutral</option>
            <option value="apologetic">Apologetic</option>
            <option value="informative">Informative</option>
            <option value="proactive">Proactive</option>
          </select>
        </div>
        <button className="greek-btn-primary" onClick={handleGenerate} disabled={isGenerating}>
          {isGenerating ? <><span className="greek-spinner" />Generating…</> : 'Generate Draft'}
        </button>
      </div>

      {error && <div className="greek-alert greek-alert-error">{error}</div>}

      {draft && (
        <div style={{
          padding: '16px 20px',
          background: 'rgba(201,168,76,0.04)',
          border: '1px solid rgba(201,168,76,0.18)',
          borderRadius: '2px',
        }}>
          <div className="flex justify-between items-center" style={{ marginBottom: '12px' }}>
            <span className="greek-badge greek-badge-gold">{(draft.channel || channel).toUpperCase()}</span>
            {draft.fallback_used && <span className="greek-badge greek-badge-muted">Rule-Based Fallback</span>}
          </div>
          <p style={{ whiteSpace: 'pre-wrap', lineHeight: 1.7, color: 'rgba(232,213,163,0.80)', fontSize: '13px', fontFamily: "'Inter', sans-serif" }}>
            {draft.draft_message}
          </p>
          <div style={{
            marginTop: '14px',
            paddingTop: '10px',
            borderTop: '1px solid rgba(201,168,76,0.10)',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}>
            <span className="greek-muted">Confidence: {draft.confidence_score.toFixed(2)}</span>
            <button className="greek-btn-secondary" style={{ padding: '6px 14px' }} onClick={handleCopy}>
              Copy Text
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
