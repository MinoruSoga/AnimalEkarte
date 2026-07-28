import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AggregationFilterPanel } from './AggregationFilterPanel';
import type { AggregationParams } from '../api/get-aggregations';

describe('AggregationFilterPanel', () => {
  const mockOnParamsChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render common filters for all tabs', async () => {
    const params: AggregationParams = {
      page: 1,
      per_page: 50,
      sort: 'annual_amount',
      order: 'desc',
    };

    render(
      <AggregationFilterPanel
        params={params}
        onParamsChange={mockOnParamsChange}
        activeTab="revenue"
      />
    );

    expect(screen.getByPlaceholderText('飼主名を検索...')).toBeInTheDocument();
  });

  it('should render revenue-specific filters', () => {
    const params: AggregationParams = {
      page: 1,
      per_page: 50,
      year: 2026,
      amount_basis: 'gross_total_amount',
      sort: 'annual_amount',
      order: 'desc',
    };

    render(
      <AggregationFilterPanel
        params={params}
        onParamsChange={mockOnParamsChange}
        activeTab="revenue"
      />
    );

    expect(screen.getByDisplayValue('2026')).toBeInTheDocument();
    expect(screen.getByText('売上総額')).toBeInTheDocument();
  });

  it('売上フィルターの入力とcheckbox hit areaを44px以上に保つ', () => {
    render(
      <AggregationFilterPanel
        params={{ page: 1, per_page: 50, year: 2026, min_amount: 100, max_amount: 200 }}
        onParamsChange={mockOnParamsChange}
        activeTab="revenue"
      />
    );

    for (const input of [
      screen.getByPlaceholderText('年度'),
      screen.getByPlaceholderText('下限'),
      screen.getByPlaceholderText('上限'),
    ]) {
      expect(input).toHaveClass('h-11');
    }
    expect(screen.getByRole('checkbox', { name: '0円を含む' })).toHaveClass('size-11');
  });

  it('should render visit-specific filters', () => {
    const params: AggregationParams = {
      page: 1,
      per_page: 50,
      period_preset: 'last_12_months',
      sort: 'period_visit_count',
      order: 'desc',
    };

    render(
      <AggregationFilterPanel
        params={params}
        onParamsChange={mockOnParamsChange}
        activeTab="visit"
      />
    );

    expect(screen.getByText('直近12ヶ月')).toBeInTheDocument();
  });

  it('should render last_visit-specific filters', () => {
    const params: AggregationParams = {
      page: 1,
      per_page: 50,
      last_visit_bucket: 'over_3m',
      sort: 'last_visit_date',
      order: 'asc',
    };

    render(
      <AggregationFilterPanel
        params={params}
        onParamsChange={mockOnParamsChange}
        activeTab="last_visit"
      />
    );

    expect(screen.getByText('3ヶ月以上')).toBeInTheDocument();
  });

  it('should call onParamsChange with updated params on search change', async () => {
    const user = userEvent.setup();
    const params: AggregationParams = {
      page: 1,
      per_page: 50,
      sort: 'annual_amount',
      order: 'desc',
    };

    render(
      <AggregationFilterPanel
        params={params}
        onParamsChange={mockOnParamsChange}
        activeTab="revenue"
      />
    );

    const searchInput = screen.getByPlaceholderText('飼主名を検索...');
    await user.type(searchInput, '田中');

    await waitFor(() => {
      expect(mockOnParamsChange).toHaveBeenCalled();
    });
  });

  it('should call onParamsChange with page reset on filter change', async () => {
    const user = userEvent.setup();
    const params: AggregationParams = {
      page: 2,
      per_page: 50,
      year: 2026,
      amount_basis: 'gross_total_amount',
      sort: 'annual_amount',
      order: 'desc',
    };

    render(
      <AggregationFilterPanel
        params={params}
        onParamsChange={mockOnParamsChange}
        activeTab="revenue"
      />
    );

    const yearSelect = screen.getByDisplayValue('2026');
    await user.click(yearSelect);
  });

  it('should not render revenue filters for non-revenue tabs', () => {
    const params: AggregationParams = {
      page: 1,
      per_page: 50,
      period_preset: 'last_12_months',
      sort: 'period_visit_count',
      order: 'desc',
    };

    render(
      <AggregationFilterPanel
        params={params}
        onParamsChange={mockOnParamsChange}
        activeTab="visit"
      />
    );

    expect(screen.queryByText('売上総額')).not.toBeInTheDocument();
  });

  it('should not render visit filters for non-visit tabs', () => {
    const params: AggregationParams = {
      page: 1,
      per_page: 50,
      year: 2026,
      amount_basis: 'gross_total_amount',
      sort: 'annual_amount',
      order: 'desc',
    };

    render(
      <AggregationFilterPanel
        params={params}
        onParamsChange={mockOnParamsChange}
        activeTab="revenue"
      />
    );

    expect(screen.queryByText('直近12ヶ月')).not.toBeInTheDocument();
  });

  it('should render the CPM segment filter on all tabs (ISSUE-180)', () => {
    const params: AggregationParams = {
      page: 1,
      per_page: 50,
      sort: 'annual_amount',
      order: 'desc',
    };

    const { rerender } = render(
      <AggregationFilterPanel
        params={params}
        onParamsChange={mockOnParamsChange}
        activeTab="revenue"
      />
    );
    expect(screen.getByText('CPMセグメント')).toBeInTheDocument();

    rerender(
      <AggregationFilterPanel
        params={params}
        onParamsChange={mockOnParamsChange}
        activeTab="visit"
      />
    );
    expect(screen.getByText('CPMセグメント')).toBeInTheDocument();

    rerender(
      <AggregationFilterPanel
        params={params}
        onParamsChange={mockOnParamsChange}
        activeTab="last_visit"
      />
    );
    expect(screen.getByText('CPMセグメント')).toBeInTheDocument();
  });

  it('should reflect the selected CPM segment label in the trigger (ISSUE-180)', () => {
    const params: AggregationParams = {
      page: 1,
      per_page: 50,
      sort: 'annual_amount',
      order: 'desc',
      cpm_stage: 'cpm_core',
    };

    render(
      <AggregationFilterPanel
        params={params}
        onParamsChange={mockOnParamsChange}
        activeTab="revenue"
      />
    );

    expect(screen.getByText('Core（コア顧客）')).toBeInTheDocument();
  });
});
