import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AggregationFilterPanel } from '../AggregationFilterPanel';
import type { LtvOwnersParams, LtvTab } from '../api/get-aggregations';

describe('AggregationFilterPanel', () => {
  const mockOnParamsChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render common filters for all tabs', async () => {
    const params: LtvOwnersParams = {
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

    expect(screen.getByPlaceholderText('飼い主名で検索')).toBeInTheDocument();
  });

  it('should render revenue-specific filters', () => {
    const params: LtvOwnersParams = {
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
    expect(screen.getByDisplayValue('gross_total_amount')).toBeInTheDocument();
  });

  it('should render visit-specific filters', () => {
    const params: LtvOwnersParams = {
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

    expect(screen.getByDisplayValue('last_12_months')).toBeInTheDocument();
  });

  it('should render last_visit-specific filters', () => {
    const params: LtvOwnersParams = {
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

    expect(screen.getByDisplayValue('over_3m')).toBeInTheDocument();
  });

  it('should call onParamsChange with updated params on search change', async () => {
    const user = userEvent.setup();
    const params: LtvOwnersParams = {
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

    const searchInput = screen.getByPlaceholderText('飼い主名で検索');
    await user.type(searchInput, '田中');

    await waitFor(() => {
      expect(mockOnParamsChange).toHaveBeenCalled();
    });
  });

  it('should call onParamsChange with page reset on filter change', async () => {
    const user = userEvent.setup();
    const params: LtvOwnersParams = {
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
    const params: LtvOwnersParams = {
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

    expect(screen.queryByDisplayValue('gross_total_amount')).not.toBeInTheDocument();
  });

  it('should not render visit filters for non-visit tabs', () => {
    const params: LtvOwnersParams = {
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

    expect(screen.queryByDisplayValue('last_12_months')).not.toBeInTheDocument();
  });
});
