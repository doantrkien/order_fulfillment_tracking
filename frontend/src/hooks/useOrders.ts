import { useState, useEffect, useCallback } from 'react';
import { orderService, type OrderQuery, type PaginatedOrders } from '../services/order.service';

export const useOrders = (initialQuery: OrderQuery = { page: 1, limit: 10 }) => {
  const [query, setQuery] = useState<OrderQuery>(initialQuery);
  const [data, setData] = useState<PaginatedOrders>({
    data: [],
    page: 1,
    limit: 10,
    total_items: 0,
    total_pages: 0,
  });
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchOrders = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const result = await orderService.getOrders(query);
      setData(result);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch orders');
    } finally {
      setIsLoading(false);
    }
  }, [query]);

  // Refetch when query changes
  useEffect(() => {
    fetchOrders();
  }, [fetchOrders]);

  // Convenience methods to update query parameters
  const setPage = (page: number) => {
    setQuery(prev => ({ ...prev, page }));
  };

  const setFilter = (filters: Partial<OrderQuery>) => {
    // When changing filters, reset to page 1
    setQuery(prev => ({ ...prev, ...filters, page: 1 }));
  };

  const clearFilters = () => {
    setQuery({ page: 1, limit: query.limit || 10 });
  };

  return {
    orders: data.data,
    pagination: {
      page: data.page,
      limit: data.limit,
      totalItems: data.total_items,
      totalPages: data.total_pages,
    },
    isLoading,
    error,
    setPage,
    setFilter,
    clearFilters,
    refetch: fetchOrders,
  };
};
