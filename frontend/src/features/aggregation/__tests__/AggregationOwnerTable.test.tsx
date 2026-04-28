import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router';
import { AggregationOwnerTable } from '../AggregationOwnerTable';
import type { AggregationOwner } from '../api/get-aggregations';
import type { AggregationTab } from '../AggregationOwnerTable';

const mockOwners: AggregationOwner[] = [
  {
    owner_id: 'owner1',
    owner_name: '田中太郎',
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
];

const createWrapper = () => {
  return ({ children }: { children: React.ReactNode }) => (
    <BrowserRouter>
      {children}
    </BrowserRouter>
  );
};

describe('AggregationOwnerTable', () => {
  it('should render owners in revenue tab', () => {
    render(
      <AggregationOwnerTable
        owners={mockOwners}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="revenue"
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText('田中太郎')).toBeInTheDocument();
    expect(screen.getByText('鈴木花子')).toBeInTheDocument();
  });

  it('should render annual_amount column in revenue tab', () => {
    render(
      <AggregationOwnerTable
        owners={mockOwners}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="revenue"
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText('期間診療費')).toBeInTheDocument();
  });

  it('should render period_visit_count in visit tab', () => {
    render(
      <AggregationOwnerTable
        owners={mockOwners}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="visit"
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText('来院回数(期間)')).toBeInTheDocument();
  });

  it('should render last_visit_bucket badge in last_visit tab', () => {
    render(
      <AggregationOwnerTable
        owners={mockOwners}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="last_visit"
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText('3ヶ月以内')).toBeInTheDocument();
    expect(screen.getByText('3ヶ月以上')).toBeInTheDocument();
  });

  it('should display error message when isError is true', () => {
    render(
      <AggregationOwnerTable
        owners={[]}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="revenue"
        isError={true}
        errorMessage="データの読み込みに失敗しました"
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText('データの読み込みに失敗しました')).toBeInTheDocument();
  });

  it('should display loading state', () => {
    render(
      <AggregationOwnerTable
        owners={[]}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={true}
        activeTab="revenue"
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText('読み込み中...')).toBeInTheDocument();
  });

  it('should display empty state when no owners', () => {
    render(
      <AggregationOwnerTable
        owners={[]}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="revenue"
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText('データが見つかりません')).toBeInTheDocument();
  });

  it('should display common columns for all tabs', () => {
    const { rerender } = render(
      <AggregationOwnerTable
        owners={mockOwners}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="revenue"
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText('飼い主名')).toBeInTheDocument();
    expect(screen.getByText('最終来院日')).toBeInTheDocument();

    rerender(
      <AggregationOwnerTable
        owners={mockOwners}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="visit"
      />
    );

    expect(screen.getByText('飼い主名')).toBeInTheDocument();
  });

  it('should render correct badge colors for different last_visit_bucket values', () => {
    const ownersWithVariousBuckets: AggregationOwner[] = [
      {
        owner_id: 'owner1',
        owner_name: 'Owner 1',
        total_visit_count: 10,
        annual_visit_count: 5,
        last_visit_date: '2026-04-20',
        first_visit_date: '2024-01-15',
        annual_amount: 50000,
        billing_count: 4,
        period_visit_count: 3,
        days_since_last_visit: 2,
        last_visit_bucket: 'within_3m',
      },
      {
        owner_id: 'owner2',
        owner_name: 'Owner 2',
        total_visit_count: 5,
        annual_visit_count: 2,
        last_visit_date: '2025-10-20',
        first_visit_date: '2025-06-01',
        annual_amount: 25000,
        billing_count: 2,
        period_visit_count: 1,
        days_since_last_visit: 160,
        last_visit_bucket: 'over_6m',
      },
    ];

    render(
      <AggregationOwnerTable
        owners={ownersWithVariousBuckets}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="last_visit"
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText('3ヶ月以内')).toBeInTheDocument();
    expect(screen.getByText('6ヶ月以上')).toBeInTheDocument();
  });

  it('should format fees correctly', () => {
    render(
      <AggregationOwnerTable
        owners={mockOwners}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="revenue"
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText('¥150,000')).toBeInTheDocument();
    expect(screen.getByText('¥80,000')).toBeInTheDocument();
  });

  it('should format dates correctly', () => {
    render(
      <AggregationOwnerTable
        owners={mockOwners}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="visit"
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText('2026-04-20')).toBeInTheDocument();
  });
});
