ALTER TABLE ai_customer_update_drafts
ADD COLUMN channel VARCHAR(20) NOT NULL DEFAULT 'email';
