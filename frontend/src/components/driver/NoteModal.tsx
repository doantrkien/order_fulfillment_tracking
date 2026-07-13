import React, { useState } from 'react';

interface NoteModalProps {
  currentNote?: string;
  onClose: () => void;
  onSave: (note: string) => Promise<void>;
}

const COMMON_ISSUES = [
  { label: '— Select a common issue —', value: '' },
  { label: '🚫 Customer not at address', value: 'Customer is not present at the delivery address.' },
  { label: '📞 Unable to contact customer', value: 'Unable to reach the customer by phone.' },
  { label: '🏠 Incorrect / hard to find address', value: 'The delivery address is incorrect or difficult to locate.' },
  { label: '🌧️ Bad weather conditions', value: 'Delivery delayed/affected due to severe weather conditions.' },
  { label: '🚦 Heavy traffic / traffic jam', value: 'Delayed due to heavy traffic, delivery will be later than scheduled.' },
  { label: '📦 Damaged package in transit', value: 'Package found to be damaged during transport.' },
  { label: '🔒 No access to secure area/building', value: 'Unable to access the building or secure area due to access control.' },
  { label: '⏰ Customer requested to reschedule', value: 'Customer requested to reschedule the delivery to a different time.' },
  { label: '🔄 Customer refused delivery', value: 'Customer refused to accept the package, support contact required.' },
  { label: '💳 Insufficient funds for COD', value: 'Customer has insufficient cash/payment method for Cash on Delivery.' },
  { label: '🚗 Vehicle breakdown / issue', value: 'Delivery vehicle encountered mechanical issues/breakdown.' },
  { label: '📍 Wrong package / order mismatch', value: 'Incorrect package or order mismatch detected, needs verification.' },
];

export const NoteModal: React.FC<NoteModalProps> = ({ currentNote = '', onClose, onSave }) => {
  const [note, setNote] = useState(currentNote);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  const handleIssueSelect = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const selected = e.target.value;
    if (!selected) return;
    setNote(prev => {
      const trimmed = prev.trim();
      return trimmed ? `${trimmed}\n${selected}` : selected;
    });
    // Reset dropdown
    e.target.value = '';
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsSubmitting(true);
    try {
      await onSave(note);
      onClose();
    } catch (err: any) {
      setError(err.message || 'Failed to save note');
      setIsSubmitting(false);
    }
  };

  return (
    <div className="greek-modal-overlay" onClick={onClose}>
      <div className="greek-modal" onClick={e => e.stopPropagation()}>
        <div className="greek-modal-header">
          <h2 className="greek-subheading">Driver Note</h2>
          <button className="greek-close-btn" onClick={onClose}>×</button>
        </div>

        <div className="greek-modal-body">
          <form onSubmit={handleSubmit}>
            {error && <div className="greek-alert greek-alert-error">{error}</div>}

            {/* Common Issues Dropdown */}
            <div style={{ marginBottom: '14px' }}>
              <label className="greek-label" htmlFor="commonIssueSelect">
                Common shipping issues
              </label>
              <div style={{ position: 'relative' }}>
                <select
                  id="commonIssueSelect"
                  className="greek-input"
                  defaultValue=""
                  onChange={handleIssueSelect}
                  style={{
                    appearance: 'none',
                    WebkitAppearance: 'none',
                    paddingRight: '36px',
                    cursor: 'pointer',
                  }}
                >
                  {COMMON_ISSUES.map((issue) => (
                    <option key={issue.value} value={issue.value} disabled={issue.value === '' && issue.label.startsWith('—')}>
                      {issue.label}
                    </option>
                  ))}
                </select>
                <span
                  style={{
                    position: 'absolute',
                    right: '12px',
                    top: '50%',
                    transform: 'translateY(-50%)',
                    pointerEvents: 'none',
                    fontSize: '12px',
                    opacity: 0.6,
                  }}
                >
                  ▼
                </span>
              </div>
              <p style={{ margin: '6px 0 0', fontSize: '12px', opacity: 0.6 }}>
                Select an issue to automatically append it to the notes below.
              </p>
            </div>

            {/* Note Textarea */}
            <div style={{ marginBottom: '16px' }}>
              <label className="greek-label" htmlFor="noteInput">Additional Notes</label>
              <textarea
                id="noteInput"
                className="greek-input"
                rows={5}
                placeholder="E.g., Customer called to delay delivery until 5 PM…"
                value={note}
                onChange={e => setNote(e.target.value)}
                style={{ resize: 'vertical' }}
              />
            </div>

            <div className="greek-modal-footer" style={{ padding: '16px 0 0', border: 'none' }}>
              <button type="button" className="greek-btn-secondary" onClick={onClose} disabled={isSubmitting}>
                Cancel
              </button>
              <button type="submit" className="greek-btn-primary" disabled={isSubmitting}>
                {isSubmitting ? <><span className="greek-spinner" />Saving…</> : 'Save Note'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};
