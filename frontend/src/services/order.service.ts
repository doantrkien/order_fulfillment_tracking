import { api } from './api';

export type OrderStatus = 'created' | 'paid' | 'packed' | 'shipped' | 'delivered' | 'cancelled' | 'refunded';

export interface Order {
  id: number;
  total_amount: number;
  username: string;
  user_phone: string;
  shipping_address: string;
  status: OrderStatus;
  ordered_at: string;
  driver_note?: string;
}

export interface PaginatedOrders {
  data: Order[];
  page: number;
  limit: number;
  total_items: number;
  total_pages: number;
}

// In some standard response wrappers, data might be nested inside 'data' again or 'meta' might be top level.
// Let's assume standard response structure from backend: { data: [...], page, limit, total_items }
export interface ApiResponse<T> {
  message: string;
  data: T;
  page?: number;
  limit?: number;
  total_items?: number;
}

export interface OrderQuery {
  page?: number;
  limit?: number;
  status?: string;
  date?: string;
  customer_name?: string;
}

export interface CreateOrderRequest {
  total_amount: number;
  username: string;
  user_phone: string;
  shipping_address: string;
}

export interface UpdateOrderStatusRequest {
  status: OrderStatus;
  driver_id?: number;
}

export const orderService = {
  getOrders: async (query: OrderQuery = {}): Promise<PaginatedOrders> => {
    const params = new URLSearchParams();
    if (query.page) params.append('page', query.page.toString());
    if (query.limit) params.append('limit', query.limit.toString());
    if (query.status) params.append('status', query.status);
    if (query.date) params.append('date', query.date);
    if (query.customer_name) params.append('customer_name', query.customer_name);

    const queryString = params.toString();
    const url = `/orders${queryString ? `?${queryString}` : ''}`;
    
    const res = await api.get(url) as ApiResponse<Order[]>;
    
    return {
      data: res.data,
      page: res.page || query.page || 1,
      limit: res.limit || query.limit || 10,
      total_items: res.total_items || 0,
      total_pages: Math.ceil((res.total_items || 0) / (res.limit || 10)),
    };
  },

  getOrder: async (id: number): Promise<Order> => {
    const res = await api.get(`/orders/${id}`) as ApiResponse<Order>;
    return res.data;
  },

  createOrder: async (data: CreateOrderRequest): Promise<Order> => {
    const res = await api.post('/orders', { data }) as ApiResponse<Order>;
    return res.data;
  },

  updateStatus: async (id: number, data: UpdateOrderStatusRequest): Promise<void> => {
    await api.patch(`/orders/${id}/status`, { data });
  },
};
