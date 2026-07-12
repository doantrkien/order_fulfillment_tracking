import { api } from './api';

export interface AIExceptionAnalysis {
  order_id: string;
  exception_type: string;
  severity: string;
  likely_reason: string;
  internal_next_action: string;
  confidence_score: number;
  fallback_used: boolean;
  fallback_reason?: string;
  prompt_template_version: string;
  evaluated_at: string;
}

export interface AIDraftRequest {
  order_id: number;
  channel: string;
  tone: string;
}

export interface AIDraftResponse {
  order_id: number;
  draft_message: string;
  channel: string;
  tone: string;
  confidence_score: number;
  fallback_used: boolean;
  prompt_template_version: string;
  generated_at: string;
}

export const aiService = {
  analyzeException: async (orderId: number): Promise<AIExceptionAnalysis> => {
    const res = await api.post(`/ai/orders/${orderId}/exception-analysis`);
    return res.data;
  },

  getLatestInsights: async (orderId: number): Promise<AIExceptionAnalysis> => {
    const res = await api.get(`/ai/orders/${orderId}/insights/latest`);
    return res.data;
  },

  generateDraft: async (data: AIDraftRequest): Promise<AIDraftResponse> => {
    const res = await api.post('/ai/orders/customer-update-draft', { data });
    return res.data;
  }
};
