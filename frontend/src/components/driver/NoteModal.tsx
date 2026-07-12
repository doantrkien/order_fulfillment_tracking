import React, { useState } from 'react';

interface NoteModalProps {
  currentNote?: string;
  onClose: () => void;
  onSave: (note: string) => Promise<void>;
}

export const NoteModal: React.FC<NoteModalProps> = ({ currentNote = '', onClose, onSave }) => {
  const [note, setNote] = useState(currentNote);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

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
