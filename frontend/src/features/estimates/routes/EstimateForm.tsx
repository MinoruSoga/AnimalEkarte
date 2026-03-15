import { useCallback } from 'react';
import { useParams } from 'react-router';
import { FileText } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { PageLayout } from '@/components/shared/PageLayout/PageLayout';
import { NavigationBlocker } from '@/components/shared/NavigationBlocker';
import { useUnsavedChanges } from '@/hooks/useUnsavedChanges';
import { useGetEstimate } from '../api/get-estimate';
import { useEstimateForm } from '../hooks/useEstimateForm';
import type { EstimateStatus } from '../types';

const STATUS_OPTIONS: { value: EstimateStatus; label: string }[] = [
  { value: 'draft', label: '下書き' },
  { value: 'sent', label: '送付済み' },
  { value: 'approved', label: '承認済み' },
  { value: 'rejected', label: '却下' },
];

function EstimateFormContent({ id }: { id?: string }) {
  const { data: estimate, isLoading } = useGetEstimate(id);
  const { form, handleChange, handleSubmit, handleCancel, isPending } = useEstimateForm(
    id ? estimate : undefined
  );

  const isEdit = !!id;

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  const handleChangeWithDirty = useCallback(
    (key: string, value: unknown) => {
      markDirty();
      (handleChange as (k: string, v: unknown) => void)(key, value);
    },
    [markDirty, handleChange]
  );

  const handleSubmitClick = useCallback(() => {
    markClean();
    handleSubmit();
  }, [markClean, handleSubmit]);

  if (isEdit && isLoading) {
    return (
      <div className="flex justify-center items-center p-8">
        <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[#37352F]" />
      </div>
    );
  }

  return (
    <PageLayout
      title={isEdit ? '見積書編集' : '新規見積書作成'}
      icon={<FileText className="size-4 text-[#37352F]" />}
      headerAction={
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={handleCancel} className="h-9 text-sm">
            キャンセル
          </Button>
          <Button
            size="sm"
            onClick={handleSubmitClick}
            disabled={isPending || !form.title.trim()}
            className="h-9 bg-[#37352F] hover:bg-[#37352F]/90 text-white text-sm"
          >
            {isPending ? '保存中...' : isEdit ? '更新' : '作成'}
          </Button>
        </div>
      }
      maxWidth="max-w-2xl"
    >
      <NavigationBlocker when={isDirty && !isPending} />
      <div className="bg-white border border-[rgba(55,53,47,0.09)] rounded-[6px] p-5 space-y-5">
        {/* タイトル */}
        <div className="space-y-1.5">
          <Label htmlFor="title" className="text-sm font-medium text-[#37352F]">
            タイトル <span className="text-red-500">*</span>
          </Label>
          <Input
            id="title"
            value={form.title}
            onChange={e => handleChangeWithDirty('title', e.target.value)}
            placeholder="見積書タイトルを入力"
            className="h-9 text-sm"
          />
        </div>

        {/* ステータス */}
        <div className="space-y-1.5">
          <Label htmlFor="status" className="text-sm font-medium text-[#37352F]">
            ステータス
          </Label>
          <Select
            value={form.status}
            onValueChange={v => handleChangeWithDirty('status', v as EstimateStatus)}
          >
            <SelectTrigger id="status" className="h-9 text-sm w-[200px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {STATUS_OPTIONS.map(opt => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* 有効期限 */}
        <div className="space-y-1.5">
          <Label htmlFor="validUntil" className="text-sm font-medium text-[#37352F]">
            有効期限
          </Label>
          <Input
            id="validUntil"
            type="date"
            value={form.validUntil ? form.validUntil.slice(0, 10) : ''}
            onChange={e => handleChangeWithDirty('validUntil', e.target.value)}
            className="h-9 text-sm w-[180px]"
          />
        </div>

        {/* 金額サマリ */}
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <Label htmlFor="subtotal" className="text-sm font-medium text-[#37352F]">
              小計（税抜）
            </Label>
            <Input
              id="subtotal"
              type="number"
              min={0}
              value={form.subtotal}
              onChange={e => handleChangeWithDirty('subtotal', Number(e.target.value))}
              className="h-9 text-sm"
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="taxTotal" className="text-sm font-medium text-[#37352F]">
              消費税
            </Label>
            <Input
              id="taxTotal"
              type="number"
              min={0}
              value={form.taxTotal}
              onChange={e => handleChangeWithDirty('taxTotal', Number(e.target.value))}
              className="h-9 text-sm"
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="insuranceAmount" className="text-sm font-medium text-[#37352F]">
              保険適用額
            </Label>
            <Input
              id="insuranceAmount"
              type="number"
              min={0}
              value={form.insuranceAmount}
              onChange={e => handleChangeWithDirty('insuranceAmount', Number(e.target.value))}
              className="h-9 text-sm"
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="discountAmount" className="text-sm font-medium text-[#37352F]">
              割引額
            </Label>
            <Input
              id="discountAmount"
              type="number"
              min={0}
              value={form.discountAmount}
              onChange={e => handleChangeWithDirty('discountAmount', Number(e.target.value))}
              className="h-9 text-sm"
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="totalAmount" className="text-sm font-medium text-[#37352F]">
              合計金額
            </Label>
            <Input
              id="totalAmount"
              type="number"
              min={0}
              value={form.totalAmount}
              onChange={e => handleChangeWithDirty('totalAmount', Number(e.target.value))}
              className="h-9 text-sm"
            />
          </div>
        </div>

        {/* コメント */}
        <div className="space-y-1.5">
          <Label htmlFor="comment" className="text-sm font-medium text-[#37352F]">
            コメント
          </Label>
          <Textarea
            id="comment"
            value={form.comment}
            onChange={e => handleChangeWithDirty('comment', e.target.value)}
            placeholder="飼主向けコメントを入力"
            className="text-sm min-h-[80px] resize-none"
          />
        </div>

        {/* 備考 */}
        <div className="space-y-1.5">
          <Label htmlFor="notes" className="text-sm font-medium text-[#37352F]">
            備考（社内メモ）
          </Label>
          <Textarea
            id="notes"
            value={form.notes}
            onChange={e => handleChangeWithDirty('notes', e.target.value)}
            placeholder="社内メモを入力"
            className="text-sm min-h-[80px] resize-none"
          />
        </div>
      </div>
    </PageLayout>
  );
}

export function EstimateForm() {
  const { id } = useParams<{ id: string }>();
  return <EstimateFormContent id={id} />;
}
