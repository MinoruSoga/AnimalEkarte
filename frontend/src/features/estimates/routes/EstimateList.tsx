import { useState, useDeferredValue, useCallback } from 'react';
import { useNavigate } from 'react-router';
import { Plus, FileText, Trash2, ExternalLink } from 'lucide-react';
import { TableCell } from '@/components/ui/table';
import { PageLayout } from '@/components/shared/PageLayout';
import { SearchFilterBar } from '@/components/shared/SearchFilterBar';
import { DataTable, DataTableRow } from '@/components/shared/DataTable';
import { PrimaryButton } from '@/components/shared/Form/PrimaryButton';
import { RowActionDropdown } from '@/components/shared/RowActionDropdown';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';
import { EstimateStatusBadge } from '../components/EstimateStatusBadge/EstimateStatusBadge';
import { useEstimates } from '../api/get-estimates';
import { useDeleteEstimate } from '../api/delete-estimate';
import type { Estimate, EstimateStatus } from '../types';

const STATUS_FILTER_OPTIONS: { value: EstimateStatus | 'all'; label: string }[] = [
  { value: 'all', label: 'すべて' },
  { value: 'draft', label: '下書き' },
  { value: 'sent', label: '送付済み' },
  { value: 'approved', label: '承認済み' },
  { value: 'rejected', label: '却下' },
];

const formatCurrency = (amount: number) =>
  new Intl.NumberFormat('ja-JP', { style: 'currency', currency: 'JPY' }).format(amount);

const COLUMNS = [
  { header: '見積番号', className: 'w-[140px]' },
  { header: 'タイトル' },
  { header: '飼主名', className: 'w-[130px]' },
  { header: '有効期限', className: 'w-[110px]' },
  { header: '合計金額', align: 'right' as const },
  { header: 'ステータス', className: 'w-[110px]' },
  { header: '操作', className: 'w-[60px]', align: 'right' as const },
];

export function EstimateList() {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<EstimateStatus | 'all'>('all');
  const [deleteTargetId, setDeleteTargetId] = useState<string | null>(null);

  const deferredSearch = useDeferredValue(searchTerm);

  const { data: result, isLoading, isError } = useEstimates(
    statusFilter !== 'all' ? { status: statusFilter } : undefined
  );
  const { mutate: deleteEstimate } = useDeleteEstimate();

  const estimates = result?.data ?? [];

  const filtered = estimates.filter(e => {
    if (!deferredSearch) return true;
    const lower = deferredSearch.toLowerCase();
    return (
      e.title.toLowerCase().includes(lower) ||
      (e.ownerName ?? '').toLowerCase().includes(lower) ||
      e.estimateNo.toLowerCase().includes(lower)
    );
  });

  const handleDeleteConfirm = useCallback(() => {
    if (deleteTargetId == null) return;
    deleteEstimate(deleteTargetId);
    setDeleteTargetId(null);
  }, [deleteTargetId, deleteEstimate]);

  const renderRow = (estimate: Estimate) => (
    <DataTableRow key={estimate.id} onClick={() => navigate(`/estimates/${estimate.id}`)}>
      <TableCell className="font-mono text-sm text-[#37352F]/60 py-2">{estimate.estimateNo}</TableCell>
      <TableCell className="text-sm text-[#37352F] py-2 font-medium">{estimate.title}</TableCell>
      <TableCell className="text-sm text-[#37352F] py-2">{estimate.ownerName ?? '-'}</TableCell>
      <TableCell className="text-sm text-[#37352F]/60 py-2">
        {estimate.validUntil ? estimate.validUntil.slice(0, 10) : '-'}
      </TableCell>
      <TableCell className="text-right font-mono font-medium text-sm text-[#37352F] py-2">
        {formatCurrency(estimate.totalAmount)}
      </TableCell>
      <TableCell className="py-2">
        <EstimateStatusBadge status={estimate.status} />
      </TableCell>
      <TableCell className="text-right py-2">
        <RowActionDropdown
          actions={[
            {
              label: '詳細',
              icon: ExternalLink,
              onClick: () => navigate(`/estimates/${estimate.id}`),
            },
            {
              label: '編集',
              icon: FileText,
              onClick: () => navigate(`/estimates/${estimate.id}/edit`),
            },
            {
              label: '削除',
              icon: Trash2,
              variant: 'destructive',
              onClick: () => setDeleteTargetId(estimate.id),
            },
          ]}
        />
      </TableCell>
    </DataTableRow>
  );

  if (isLoading) {
    return (
      <div className="flex justify-center items-center p-8">
        <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[#37352F]" />
      </div>
    );
  }
  if (isError) {
    return <div className="p-4 text-red-600">データの取得に失敗しました</div>;
  }

  return (
    <PageLayout
      title="見積書管理"
      icon={<FileText className="size-4 text-[#37352F]" />}
      headerAction={
        <PrimaryButton onClick={() => navigate('/estimates/new')}>
          <Plus className="mr-1.5 size-4" />
          新規見積書作成
        </PrimaryButton>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        {/* Status filter tabs */}
        <div className="flex gap-1">
          {STATUS_FILTER_OPTIONS.map(opt => (
            <button
              key={opt.value}
              type="button"
              onClick={() => setStatusFilter(opt.value)}
              className={[
                'px-3 h-8 rounded-[6px] text-sm transition-colors',
                statusFilter === opt.value
                  ? 'bg-[#37352F] text-white'
                  : 'text-[#37352F]/60 hover:bg-[rgba(55,53,47,0.06)]',
              ].join(' ')}
            >
              {opt.label}
            </button>
          ))}
        </div>

        <SearchFilterBar
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          placeholder="見積番号、タイトル、飼主名..."
          count={filtered.length}
        />

        <DataTable
          columns={COLUMNS}
          data={filtered}
          emptyMessage="見積書が見つかりません"
          renderRow={renderRow}
        />
      </div>

      <ConfirmDialog
        open={deleteTargetId != null}
        onClose={() => setDeleteTargetId(null)}
        onConfirm={handleDeleteConfirm}
        title="見積書を削除しますか?"
        description="この操作は取り消せません。"
        confirmLabel="削除"
        variant="destructive"
      />
    </PageLayout>
  );
}
