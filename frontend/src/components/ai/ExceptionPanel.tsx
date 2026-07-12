import type { AIExceptionAnalysis } from '../../services/ai.service';
import { useTheme } from '../../hooks/useTheme';

interface ExceptionPanelProps {
  analysis: AIExceptionAnalysis | null;
  isLoading: boolean;
  onTriggerAnalysis: () => void;
}

const getSeverityClass = (severity: string) => {
  switch (severity.toLowerCase()) {
    case 'high':   return 'greek-badge greek-badge-red';
    case 'medium': return 'greek-badge greek-badge-gold';
    case 'low':    return 'greek-badge greek-badge-green';
    default:       return 'greek-badge greek-badge-muted';
  }
};

const FIELD_LABEL: React.CSSProperties = {
  fontFamily: "'Cinzel', serif",
  fontSize: '9px',
  fontWeight: 600,
  letterSpacing: '0.22em',
  textTransform: 'uppercase',
  marginBottom: '4px',
};
const FIELD_VALUE: React.CSSProperties = {
  fontFamily: "'Inter', sans-serif",
  fontSize: '13px',
  lineHeight: 1.5,
};

export const ExceptionPanel: React.FC<ExceptionPanelProps> = ({ analysis, isLoading, onTriggerAnalysis }) => {
  const { isLight } = useTheme();

  if (isLoading) {
    return (
      <div className="greek-card-static mb-4" style={{ display: 'flex', alignItems: 'center', gap: '12px', padding: '20px 24px' }}>
        <span className="greek-spinner" />
        <span className="greek-muted">Analyzing order exceptions…</span>
      </div>
    );
  }

  if (!analysis) {
    return (
      <div className="greek-card-static mb-4" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '20px 24px' }}>
        <span className="greek-muted">No AI insights available for this order.</span>
        <button className="greek-btn-primary" onClick={onTriggerAnalysis}>
          Run AI Analysis
        </button>
      </div>
    );
  }

  return (
    <div className="greek-card-static mb-4" style={{ borderLeft: `3px solid ${isLight ? 'rgba(160,120,20,0.45)' : 'rgba(201,168,76,0.45)'}` }}>
      <div className="flex justify-between items-center" style={{ marginBottom: '16px' }}>
        <h3 className="greek-subheading">AI Exception Insights</h3>
        <div className="flex items-center gap-2">
          {analysis.fallback_used && (
            <span className="greek-badge greek-badge-muted" title={analysis.fallback_reason}>
              Rule-Based Fallback {analysis.fallback_reason ? `(${analysis.fallback_reason})` : ''}
            </span>
          )}
          <button
            className="greek-btn-secondary"
            style={{ padding: '4px 12px', fontSize: '11px' }}
            onClick={onTriggerAnalysis}
          >
            Re-run Analysis
          </button>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
        <div>
          <p className="greek-label" style={FIELD_LABEL}>Exception Type</p>
          <p className="greek-text" style={FIELD_VALUE}>{analysis.exception_type}</p>
        </div>
        <div>
          <p className="greek-label" style={FIELD_LABEL}>Severity</p>
          <span className={getSeverityClass(analysis.severity)}>{analysis.severity.toUpperCase()}</span>
        </div>
        <div style={{ gridColumn: '1 / -1' }}>
          <p className="greek-label" style={FIELD_LABEL}>Likely Reason</p>
          <p className="greek-text" style={FIELD_VALUE}>{analysis.likely_reason}</p>
        </div>
        <div style={{ gridColumn: '1 / -1' }}>
          <p className="greek-label" style={FIELD_LABEL}>Recommended Next Action</p>
          <p className="greek-text" style={FIELD_VALUE}>{analysis.internal_next_action}</p>
        </div>
      </div>

      <div style={{
        marginTop: '16px',
        paddingTop: '12px',
        borderTop: `1px solid ${isLight ? 'rgba(160,120,20,0.10)' : 'rgba(201,168,76,0.10)'}`,
        display: 'flex',
        justifyContent: 'space-between',
      }}>
        <span className="greek-muted">Confidence: {analysis.confidence_score.toFixed(2)}</span>
        <span className="greek-muted">Evaluated: {new Date(analysis.evaluated_at).toLocaleString('vi-VN')}</span>
      </div>
    </div>
  );
};
