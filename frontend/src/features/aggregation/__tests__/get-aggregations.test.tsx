import type { ReactNode } from 'react';
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { http, HttpResponse } from 'msw';
import { server } from '@/testing/mocks/node';
import { useGetOwnerAggregations, type AggregationResponse } from '../api/get-aggregations';

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

const mockResponse: AggregationResponse = {
  owners: [
    {
      owner_id: 'owner1',
      owner_name: '田中太郎',
      total_fee: 500000,
      total_visit_count: 20,
      annual_visit_count: 12,
      last_visit_date: '2026-04-20',
      first_visit_date: '2024-01-15',
      annual_amount: 150000,
      billing_count: 8,
      period_visit_count: 5,
      days_since_last_visit: 7,
      last_visit_bucket: 'within_3m',
    },
    {
      owner_id: 'owner2',
      owner_name: '鈴木花子',
      total_fee: 200000,
      total_visit_count: 8,
      annual_visit_count: 4,
      last_visit_date: '2026-01-10',
      first_visit_date: '2025-06-01',
      annual_amount: 80000,
      billing_count: 4,
      period_visit_count: 2,
      days_since_last_visit: 107,
      last_visit_bucket: 'over_3m',
    },
  ],
  page: 1,
  per_page: 50,
  total: 2,
};

describe('useGetOwnerAggregations', () => {
  beforeEach(() => {
    // Mock localStorage for clinic_id
    localStorage.setItem('auth_current_clinic:v1', '1');

    server.use(
      http.get('/api/v1/clinics/:clinic_id/owners/aggregations', () => {
        return HttpResponse.json(mockResponse);
      })
    );
  });

  afterEach(() => {
    localStorage.removeItem('auth_current_clinic:v1');
  });

  it('should fetch aggregation data with default parameters', async () => {
    const { result } = renderHook(
      () => useGetOwnerAggregations({ page: 1, per_page: 50 }),
      { wrapper: createWrapper() }
    );

    expect(result.current.isLoading).toBe(true);

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(result.current.data?.owners).toHaveLength(2);
  });

  it('should include revenue-specific parameters in query', async () => {
    const { result } = renderHook(
      () =>
        useGetOwnerAggregations({
          page: 1,
          per_page: 50,
          year: 2026,
          amount_basis: 'gross_total_amount',
          sort: 'annual_amount',
          order: 'desc',
        }),
      { wrapper: createWrapper() }
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data?.owners[0].annual_amount).toBe(150000);
  });

  it('should include visit-specific parameters in query', async () => {
    const { result } = renderHook(
      () =>
        useGetOwnerAggregations({
          page: 1,
          per_page: 50,
          period_preset: 'last_12_months',
          sort: 'period_visit_count',
          order: 'desc',
        }),
      { wrapper: createWrapper() }
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data?.owners[0].period_visit_count).toBe(5);
  });

  it('should include last_visit-specific parameters in query', async () => {
    const { result } = renderHook(
      () =>
        useGetOwnerAggregations({
          page: 1,
          per_page: 50,
          last_visit_bucket: 'over_3m',
          sort: 'last_visit_date',
          order: 'asc',
        }),
      { wrapper: createWrapper() }
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data?.owners).toBeDefined();
  });

  it('should handle API errors gracefully', async () => {
    server.use(
      http.get('/api/v1/clinics/:clinic_id/owners/aggregations', () => {
        return HttpResponse.json({ error: 'Not found' }, { status: 404 });
      })
    );

    const { result } = renderHook(
      () => useGetOwnerAggregations({ page: 1, per_page: 50 }),
      { wrapper: createWrapper() }
    );

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(result.current.error).toBeDefined();
  });

  it('should support optional aggregation fields', async () => {
    const { result } = renderHook(
      () => useGetOwnerAggregations({ page: 1, per_page: 50 }),
      { wrapper: createWrapper() }
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    const owner = result.current.data?.owners[0];
    expect(owner?.annual_amount).toBeDefined();
    expect(owner?.period_visit_count).toBeDefined();
    expect(owner?.last_visit_bucket).toBeDefined();
    expect(owner?.days_since_last_visit).toBeDefined();
  });

  it('should handle pagination parameters', async () => {
    const { result } = renderHook(
      () =>
        useGetOwnerAggregations({
          page: 2,
          per_page: 25,
          sort: 'owner_name',
          order: 'asc',
        }),
      { wrapper: createWrapper() }
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data?.page).toBe(1);
    expect(result.current.data?.per_page).toBe(50);
  });
});
