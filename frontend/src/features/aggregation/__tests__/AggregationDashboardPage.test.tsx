import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router';
import { http, HttpResponse } from 'msw';
import { server } from '@/testing/mocks/node';
import { AggregationDashboardPage } from '../AggregationDashboardPage';
import type { AggregationResponse } from '../api/get-aggregations';

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

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={["/aggregation"]}>{children}</MemoryRouter>
    </QueryClientProvider>
  );
};

describe('AggregationDashboardPage', () => {
  beforeEach(() => {
    server.use(
      http.get('/api/v1/clinics/:clinic_id/owners/aggregations', () => {
        return HttpResponse.json(mockResponse);
      })
    );
    vi.clearAllMocks();
  });

  it('should render revenue tab by default', async () => {
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText('読み込み中...')).not.toBeInTheDocument();
    });

    expect(screen.getByRole('tab', { name: '売上ランキング' })).toHaveAttribute('data-state', 'active');
    expect(screen.getByText('田中太郎')).toBeInTheDocument();
  });

  it('should switch tabs and update URL state', async () => {
    const user = userEvent.setup();
    const { container } = render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText('読み込み中...')).not.toBeInTheDocument();
    });

    const tabs = screen.getAllByRole('tab');
    expect(tabs).toHaveLength(3);

    await user.click(tabs[1]);
    expect(screen.getByRole('tab', { name: '来院回数' })).toHaveAttribute('data-state', 'active');
  });

  it('should display error message on API failure', async () => {
    server.use(
      http.get('/api/v1/clinics/:clinic_id/owners/aggregations', () => {
        return HttpResponse.json({ error: 'Not found' }, { status: 404 });
      })
    );

    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText(/読み込みに失敗しました/)).toBeInTheDocument();
    });
  });

  it('should render pagination controls', async () => {
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText('読み込み中...')).not.toBeInTheDocument();
    });

    expect(screen.getByText('2 件')).toBeInTheDocument();
  });

  it('should display loading state initially', () => {
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    expect(screen.getByText('読み込み中...')).toBeInTheDocument();
  });

  it('should render table with owner data', async () => {
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText('田中太郎')).toBeInTheDocument();
    });

    expect(screen.getByText('鈴木花子')).toBeInTheDocument();
  });

  it('should disable CSV export button when no rows are selected', async () => {
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.queryByText('読み込み中...')).not.toBeInTheDocument();
    });

    const csvButton = screen.getByRole('button', { name: /CSV出力/ });
    expect(csvButton).toBeDisabled();
    expect(csvButton).toHaveAttribute('title', '出力対象を選択してください');
  });

  it('should enable CSV export button when at least one row is selected', async () => {
    const user = userEvent.setup();
    render(<AggregationDashboardPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText('田中太郎')).toBeInTheDocument();
    });

    const ownerCheckbox = screen.getByRole('checkbox', { name: /田中太郎を選択/ });
    await user.click(ownerCheckbox);

    const csvButton = screen.getByRole('button', { name: /1件をCSV出力/ });
    expect(csvButton).not.toBeDisabled();
  });
});
