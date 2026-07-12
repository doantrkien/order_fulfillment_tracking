import { api } from './api';
import type { OrderStatus } from './order.service';

export interface ImportOrderEventRequest {
  order_id: number;
  status: OrderStatus;
  driver_id?: number;
  driver_note?: string;
  event_at?: string;
  updated_by?: string;
}

export interface ImportOrderEventsResponse {
  accepted_count: number;
  rejected_count: number;
  duplicate_count: number;
  errors?: any[];
}

export const eventService = {
  importEvents: async (events: ImportOrderEventRequest[]): Promise<ImportOrderEventsResponse> => {
    // Fill in defaults for event_at if not provided
    const payload = events.map(e => ({
      ...e,
      event_at: e.event_at || new Date().toISOString()
    }));

    const res = await api.post('/order-events/import', { data: payload });
    return res.data;
  },

  // Dedicated endpoint for drivers: only updates the note, does not create new events
  addDriverNote: async (orderId: number, note: string) => {
    const res = await api.patch(`/orders/${orderId}/driver-note`, { data: { note } });
    return res.data;
  },

  // Driver-only: update status, only allowed when current status is 'packed' -> 'shipped'
  driverUpdateStatus: async (orderId: number, status: string) => {
    const res = await api.patch(`/orders/${orderId}/status`, { data: { status } });
    return res.data;
  }
};
