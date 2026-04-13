import { ICON, C } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { LoadingFallback } from "@/components/shared/DataStates";
import { memo, useCallback, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router';
import { FileText } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
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
import { NotionDatePicker } from '@/components/shared/NotionDatePicker/NotionDatePicker';
import { NavigationBlocker } from '@/components/shared/NavigationBlocker';
import { NumberInput } from '@/components/shared/NumberInput/NumberInput';
import { FormFieldError } from '@/components/shared/FormFieldError';
import { useUnsavedChanges } from '@/hooks/use-unsaved-changes';
import { useGetEstimate } from '../api/get-estimate';
import { useEstimateForm } from '../hooks/use-estimate-form';
import { usePermission } from '@/features/auth';
import type { EstimateStatus } from '../types';
import { ResourceEstimates } from "@/types/generated/models";

// rendering-hoist-jsx: ステータス選択肢は静的なのでモジュール定数に巻き上げ
const STATUS_OPTIONS: { value: EstimateStatus; label: string }[] = [
  { value: 'draft', label: '下書き' },
  { value: 'sent', label: '送付済み' },
  { value: 'approved', label: '承認済み' },
  { value: 'rejected', label: '却下' },
];

// rendering-hoist-jsx: ステータス SelectItem リストは静的なのでモジュール定数に巻き上げ
const STATUS_SELECT_ITEMS = STATUS_OPTIONS.map(opt => (
  <SelectItem key={opt.value} value={opt.value}>
    {opt.label}
  </SelectItem>
));

// rerender-memo: 基本情報セクション（タイトル/ステータス/有効期限）を独立 memo 化
// 金額・テキストフィールドの変更では再レンダーしない
interface BasicInfoSectionProps {
  title: string;
  status: EstimateStatus;
  validUntil: string;
  onChange: (key: string, value: unknown) => void;
  titleError?: string;
}

const BasicInfoSection = memo(function BasicInfoSection({
  title,
  status,
  validUntil,
  onChange,
  titleError,
}: BasicInfoSectionProps) {
  return (
    <>
      {/* タイトル */}
      <div className="space-y-1.5">
        <Label htmlFor="title" className={`text-sm font-medium ${C.text}`}>
          タイトル <span className={C.textRequired}>*</span>
        </Label>
        <Input
          id="title"
          value={title}
          onChange={e => onChange('title', e.target.value)}
          placeholder="見積書タイトルを入力"
          className="h-9 text-sm"
        />
        <FormFieldError message={titleError} />
      </div>

      {/* ステータス */}
      <div className="space-y-1.5">
        <Label htmlFor="status" className={`text-sm font-medium ${C.text}`}>
          ステータス
        </Label>
        <Select
          value={status}
          onValueChange={v => onChange('status', v as EstimateStatus)}
        >
          <SelectTrigger id="status" className="h-9 text-sm w-[200px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {STATUS_SELECT_ITEMS}
          </SelectContent>
        </Select>
      </div>

      {/* 有効期限 */}
      <div className="space-y-1.5">
        <Label htmlFor="validUntil" className={`text-sm font-medium ${C.text}`}>
          有効期限
        </Label>
        <NotionDatePicker
          id="validUntil"
          value={validUntil ? validUntil.slice(0, 10) : ''}
          onChange={(v) => onChange('validUntil', v)}
          placeholder="有効期限を選択…"
          className="w-[220px]"
        />
      </div>
    </>
  );
});

// rerender-memo: 金額サマリセクションを独立 memo 化
// タイトル/ステータス/テキストフィールドの変更では再レンダーしない
interface AmountSectionProps {
  subtotal: number;
  taxTotal: number;
  insuranceAmount: number;
  discountAmount: number;
  totalAmount: number;
  onChange: (key: string, value: unknown) => void;
}

const AmountSection = memo(function AmountSection({
  subtotal,
  taxTotal,
  insuranceAmount,
  discountAmount,
  totalAmount,
  onChange,
}: AmountSectionProps) {
  return (
    <div className="grid grid-cols-2 gap-4">
      <div className="space-y-1.5">
        <Label htmlFor="subtotal" className={`text-sm font-medium ${C.text}`}>
          小計（税抜）
        </Label>
        <NumberInput
          id="subtotal"
          min={0}
          value={subtotal}
          onChange={v => onChange('subtotal', Number(v))}
          suffix="円"
          className="h-9 text-sm"
        />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="taxTotal" className={`text-sm font-medium ${C.text}`}>
          消費税
        </Label>
        <NumberInput
          id="taxTotal"
          min={0}
          value={taxTotal}
          onChange={v => onChange('taxTotal', Number(v))}
          suffix="円"
          className="h-9 text-sm"
        />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="insuranceAmount" className={`text-sm font-medium ${C.text}`}>
          保険適用額
        </Label>
        <NumberInput
          id="insuranceAmount"
          min={0}
          value={insuranceAmount}
          onChange={v => onChange('insuranceAmount', Number(v))}
          suffix="円"
          className="h-9 text-sm"
        />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="discountAmount" className={`text-sm font-medium ${C.text}`}>
          割引額
        </Label>
        <NumberInput
          id="discountAmount"
          min={0}
          value={discountAmount}
          onChange={v => onChange('discountAmount', Number(v))}
          suffix="円"
          className="h-9 text-sm"
        />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="totalAmount" className={`text-sm font-medium ${C.text}`}>
          合計金額
        </Label>
        <NumberInput
          id="totalAmount"
          min={0}
          value={totalAmount}
          onChange={v => onChange('totalAmount', Number(v))}
          suffix="円"
          className="h-9 text-sm"
        />
      </div>
    </div>
  );
});

// rerender-memo: テキストセクション（コメント/備考）を独立 memo 化
// 金額・基本情報フィールドの変更では再レンダーしない
interface TextSectionProps {
  comment: string;
  notes: string;
  onChange: (key: string, value: unknown) => void;
}

const TextSection = memo(function TextSection({
  comment,
  notes,
  onChange,
}: TextSectionProps) {
  return (
    <>
      {/* コメント */}
      <div className="space-y-1.5">
        <Label htmlFor="comment" className={`text-sm font-medium ${C.text}`}>
          コメント
        </Label>
        <Textarea
          id="comment"
          value={comment}
          onChange={e => onChange('comment', e.target.value)}
          placeholder="飼主向けコメントを入力"
          className="text-sm min-h-[80px] resize-none"
        />
      </div>

      {/* 備考 */}
      <div className="space-y-1.5">
        <Label htmlFor="notes" className={`text-sm font-medium ${C.text}`}>
          備考（社内メモ）
        </Label>
        <Textarea
          id="notes"
          value={notes}
          onChange={e => onChange('notes', e.target.value)}
          placeholder="社内メモを入力"
          className="text-sm min-h-[80px] resize-none"
        />
      </div>
    </>
  );
});

function EstimateFormContent({ id }: { id?: string }) {
  const navigate = useNavigate();
  const { data: estimate, isLoading } = useGetEstimate(id);
  const { form, handleChange, formAction, formState, handleCancel, isPending } = useEstimateForm(
    id ? estimate : undefined
  );

  const isEdit = !!id;
  const { canEdit, canCreate } = usePermission("estimates");
  const canSubmit = isEdit ? canEdit : canCreate;

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  // React 19 Action の成功を検知して遷移
  // rerender-dependencies: estimate（オブジェクト）の代わりに estimate?.id（primitive）を deps に使用
  const estimateId = estimate?.id;
  useEffect(() => {
    if (formState.success) {
      markClean();
      if (isEdit && estimateId != null) {
        navigate(paths.estimates.detail.getHref(estimateId));
      } else {
        // Since we don't have the new ID easily here, we might need to handle it in hook
        // but for now redirect to list
        navigate(paths.estimates.getHref());
      }
    }
  }, [formState.success, formState.timestamp, navigate, markClean, isEdit, estimateId]);

  // rerender-memo: memo'd セクションに渡すハンドラを useCallback で安定化
  const handleChangeWithDirty = useCallback(
    (key: string, value: unknown) => {
      markDirty();
      (handleChange as (k: string, v: unknown) => void)(key, value);
    },
    [markDirty, handleChange]
  );

  if (isEdit && isLoading) {
    return <LoadingFallback />;
  }

  return (
    <form action={formAction}>
    <PageLayout
      title={isEdit ? '見積書編集' : '新規見積書作成'}
      resource={ResourceEstimates}
      icon={<FileText className={`${ICON.page} ${C.text}`} />}
      headerAction={
        <div className="flex gap-2">
          <Button variant="outline" type="button" size="sm" onClick={handleCancel} className="h-9 text-sm">
            キャンセル
          </Button>
          {canSubmit ? (
            <SubmitButton
              size="sm"
              disabled={!form.title.trim()}
              className={`h-9 ${C.bgAccent} ${C.bgAccentHover} ${C.textWhite} text-sm`}
            >
              {isEdit ? '更新' : '作成'}
            </SubmitButton>
          ) : null}
        </div>
      }
      maxWidth="max-w-2xl"
    >
      <NavigationBlocker when={isDirty && !isPending} />
      <div className={`bg-white border ${C.borderLight} rounded-[6px] p-5 space-y-5`}>
        {/* rerender-memo: BasicInfoSection — 金額/テキスト変更では再レンダーしない */}
        <BasicInfoSection
          title={form.title}
          status={form.status}
          validUntil={form.validUntil}
          onChange={handleChangeWithDirty}
          titleError={formState.fieldErrors?.title}
        />

        {/* rerender-memo: AmountSection — 基本情報/テキスト変更では再レンダーしない */}
        <AmountSection
          subtotal={form.subtotal}
          taxTotal={form.taxTotal}
          insuranceAmount={form.insuranceAmount}
          discountAmount={form.discountAmount}
          totalAmount={form.totalAmount}
          onChange={handleChangeWithDirty}
        />

        {/* rerender-memo: TextSection — 基本情報/金額変更では再レンダーしない */}
        <TextSection
          comment={form.comment}
          notes={form.notes}
          onChange={handleChangeWithDirty}
        />
      </div>
    </PageLayout>
    </form>
  );
}

export function EstimateForm() {
  const { id } = useParams<{ id: string }>();
  return <EstimateFormContent id={id} />;
}
