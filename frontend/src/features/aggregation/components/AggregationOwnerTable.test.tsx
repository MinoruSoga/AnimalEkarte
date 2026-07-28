import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router';
import { AggregationOwnerTable } from './AggregationOwnerTable';
import type { AggregationOwner } from '../api/get-aggregations';

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
    total_amount: 500000,
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
    total_amount: 200000,
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
  it('keeps header and row selection targets at 44px without expanding table cell padding', () => {
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

    const checkboxes = screen.getAllByRole('checkbox');
    expect(checkboxes).toHaveLength(3);
    for (const checkbox of checkboxes) {
      expect(checkbox).toHaveClass('size-11', '-my-3');
      expect(checkbox.closest('th, td')).not.toHaveClass('py-0');
    }
  });

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

    expect(screen.getByText('年間診療費')).toBeInTheDocument();
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

    // visit タブの主要列「期間内来院回数」が描画される（仕様 §4.2）
    expect(screen.getAllByText('期間内来院回数').length).toBeGreaterThan(0);
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

    expect(screen.getByText('3ヶ月未満')).toBeInTheDocument();
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

    expect(screen.getByText('飼主名')).toBeInTheDocument();
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

    expect(screen.getByText('飼主名')).toBeInTheDocument();
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

    expect(screen.getByText('3ヶ月未満')).toBeInTheDocument();
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

  // 仕様書 §4.3 表示項目: 飼主名 / 最終来院日 / 経過日数 / 分類 / 累計来院回数 / 年間来院回数 / 累計診療費
  it('should render all spec §4.3 columns in last_visit tab', () => {
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

    // ヘッダ列が仕様通り 7 個（チェックボックス除く）揃っている
    expect(screen.getByText('飼主名')).toBeInTheDocument();
    expect(screen.getByText('最終来院日')).toBeInTheDocument();
    expect(screen.getByText('経過日数')).toBeInTheDocument();
    expect(screen.getByText('分類')).toBeInTheDocument();
    expect(screen.getByText('累計来院回数')).toBeInTheDocument();
    expect(screen.getByText('年間来院回数')).toBeInTheDocument();
    expect(screen.getByText('累計診療費')).toBeInTheDocument();

    // 累計診療費は canonical の total_amount を表示
    expect(screen.getByText('¥500,000')).toBeInTheDocument();
    expect(screen.getByText('¥200,000')).toBeInTheDocument();
    // 経過日数は「N日」形式
    expect(screen.getByText('7日')).toBeInTheDocument();
    expect(screen.getByText('107日')).toBeInTheDocument();
  });

  // 仕様 §4.3: 来院なしの飼い主は「来院なし」バッジで明示的に判別できる
  it('should render no_visit owner with dedicated badge and dashes for missing dates', () => {
    const noVisitOwner: AggregationOwner = {
      owner_id: 'owner3',
      owner_name: '来院なし太郎',
      total_visit_count: 0,
      annual_visit_count: 0,
      last_visit_date: null,
      first_visit_date: null,
      days_since_last_visit: null,
      last_visit_bucket: 'no_visit',
      total_amount: 0,
    };

    render(
      <AggregationOwnerTable
        owners={[noVisitOwner]}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="last_visit"
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText('来院なし太郎')).toBeInTheDocument();
    expect(screen.getByText('来院なし')).toBeInTheDocument();
    // last_visit_date / days_since_last_visit が null のときは em-dash で揃う
    const dashes = screen.getAllByText('—');
    expect(dashes.length).toBeGreaterThanOrEqual(2);
  });

  // 互換: BE が canonical の total_amount を返さず total_fee のみのときは total_fee で表示
  it('should fall back to total_fee when total_amount is undefined', () => {
    const legacyOwner: AggregationOwner = {
      owner_id: 'legacy1',
      owner_name: '互換太郎',
      total_visit_count: 5,
      annual_visit_count: 3,
      last_visit_date: '2026-04-01',
      first_visit_date: '2024-01-01',
      days_since_last_visit: 28,
      last_visit_bucket: 'within_3m',
      total_fee: 333000,
      // total_amount を意図的に未指定
    };

    render(
      <AggregationOwnerTable
        owners={[legacyOwner]}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="last_visit"
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText('¥333,000')).toBeInTheDocument();
  });

  it('should render spec §4.1 revenue columns with canonical labels', () => {
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

    expect(screen.getByText('年間診療費')).toBeInTheDocument();
    expect(screen.getByText('会計件数')).toBeInTheDocument();
    expect(screen.getByText('期間内来院回数')).toBeInTheDocument();
    expect(screen.getByText('最終来院日')).toBeInTheDocument();
    expect(screen.getByText('初診日')).toBeInTheDocument();
  });

  it('should render spec §4.2 visit columns with canonical labels', () => {
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

    expect(screen.getByText('期間内来院回数')).toBeInTheDocument();
    expect(screen.getByText('累計来院回数')).toBeInTheDocument();
    expect(screen.getByText('年間来院回数')).toBeInTheDocument();
    expect(screen.getByText('最終来院日')).toBeInTheDocument();
    expect(screen.getByText('初診日')).toBeInTheDocument();
  });

  // 仕様書 §10.2「読み込み中に不整合な操作をしにくいこと」
  it('should disable header checkbox while loading', () => {
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

    const headerCheckbox = screen.getByRole('checkbox', { name: '全選択' });
    expect(headerCheckbox).toBeDisabled();
  });

  it('should disable header checkbox on error', () => {
    render(
      <AggregationOwnerTable
        owners={[]}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="revenue"
        isError={true}
        errorMessage="データの読み込みに失敗しました (HTTP 404)"
      />,
      { wrapper: createWrapper() }
    );

    const headerCheckbox = screen.getByRole('checkbox', { name: '全選択' });
    expect(headerCheckbox).toBeDisabled();
    // エラーメッセージが画面に出ている
    expect(screen.getByText(/読み込みに失敗しました/)).toBeInTheDocument();
  });

  it('should disable header checkbox when owners list is empty', () => {
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

    const headerCheckbox = screen.getByRole('checkbox', { name: '全選択' });
    expect(headerCheckbox).toBeDisabled();
    expect(screen.getByText('データが見つかりません')).toBeInTheDocument();
  });

  it('should enable header checkbox when owners exist and not loading/error', () => {
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

    const headerCheckbox = screen.getByRole('checkbox', { name: '全選択' });
    expect(headerCheckbox).not.toBeDisabled();
  });

  it('should render the CPM segment column with badge on all tabs (ISSUE-180)', () => {
    const ownersWithCpm: AggregationOwner[] = [
      { ...mockOwners[0], cpm_stage: 'cpm_core' },
      { ...mockOwners[1], cpm_stage: 'cpm_dormant' },
    ];

    const { rerender } = render(
      <AggregationOwnerTable
        owners={ownersWithCpm}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="revenue"
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText('CPM')).toBeInTheDocument();
    expect(screen.getByText('Core')).toBeInTheDocument();
    expect(screen.getByText('Dormant')).toBeInTheDocument();

    rerender(
      <AggregationOwnerTable
        owners={ownersWithCpm}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="last_visit"
      />
    );

    expect(screen.getByText('CPM')).toBeInTheDocument();
    expect(screen.getByText('Core')).toBeInTheDocument();
  });

  it('should render a dash for owners without a cpm_stage (ISSUE-180)', () => {
    const ownerNoCpm: AggregationOwner = { ...mockOwners[0], cpm_stage: undefined };

    render(
      <AggregationOwnerTable
        owners={[ownerNoCpm]}
        selectedOwnerIds={new Set()}
        onSelectAll={() => {}}
        onSelectOwner={() => {}}
        isLoading={false}
        activeTab="revenue"
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText('CPM')).toBeInTheDocument();
    expect(screen.getAllByText('—').length).toBeGreaterThanOrEqual(1);
  });
});
