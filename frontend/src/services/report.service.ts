import { api } from './api';
import type { ApiResponse } from './order.service';

export interface DailyReportResponse {
  date: string;
  total_orders: number;
  total_new: number;
  total_delivered: number;
  total_cancelled: number;
  total_refunded: number;
  total_income: number;
  avg_deliver_time: number;
}

export const reportService = {
  getDailyReport: async (date: string): Promise<DailyReportResponse> => {
    // API expects format: YYYY-MM-DD
    const res = await api.get(`/reports/daily?date=${date}`) as ApiResponse<DailyReportResponse>;
    return res.data;
  }
};
