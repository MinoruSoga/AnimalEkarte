import { ICON, C } from "@/lib/design-tokens";
import { useNavigate, useParams } from 'react-router';
import { FileText, Pencil, Trash2, ArrowLeft } from 'lucide-react';
import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { PageLayout } from '@/components/shared/PageLayout/PageLayout';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog/ConfirmDialog';
import { EstimateStatusBadge } from '../components/EstimateStatusBadge/EstimateStatusBadge';
import { EstimateLineItems } from '../components/EstimateLineItems/EstimateLineItems';
import { useGetEstimate } from '../api/get-estimate';
import { useDeleteEstimate } from '../api/delete-estimate';
import { ResourceEstimates } from "@/types/generated/models";

export function EstimateDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);

  const { data: estimate, isLoading, isError } = useGetEstimate(id);
  const { mutate: deleteEstimate, isPending: isDeleting } = useDeleteEstimate();

  const handleDelete = () => {
    if (!id) return;
    deleteEstimate(id, {
      onSuccess: () => navigate('/estimates'),
    });
  };

  if (isLoading) {
    return (
      <div className="flex justify-center items-center p-8">
        <div className={`inline-block animate-spin rounded-full h-8 w-8 border-b-2 ${C.borderPrimary}`} />
      </div>
    );
  }
  if (isError || !estimate) {
    return <div className="p-4 text-red-600">データの取得に失敗しました</div>;
  }

  return (
    <PageLayout
      title={`見積書 ${estimate.estimateNo}`}
      resource={ResourceEstimates}
      icon={<FileText className={`${ICON.page} ${C.text}`} />}
      headerAction={
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate('/estimates')}
            className="h-9 gap-1.5 text-sm"
          >
            <ArrowLeft className={ICON.action} />
            一覧へ
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate(`/estimates/${id}/edit`)}
            className="h-9 gap-1.5 text-sm"
          >
            <Pencil className={ICON.action} />
            編集
          </Button>
          <Button
            variant="ghost-danger"
            size="sm"
            onClick={() => setShowDeleteDialog(true)}
            disabled={isDeleting}
            className="h-9 gap-1.5 text-sm border border-red-200"
          >
            <Trash2 className={ICON.action} />
            削除
          </Button>
        </div>
      }
      maxWidth="max-w-4xl"
    >
      <div className="space-y-6">
        {/* 基本情報 */}
        <div className={`bg-white border ${C.borderLight} rounded-[6px] p-5 space-y-4`}>
          <div className="flex items-center justify-between">
            <h2 className={`text-base font-semibold ${C.text}`}>{estimate.title}</h2>
            <EstimateStatusBadge status={estimate.status} />
          </div>

          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <dt className={`${C.text50} mb-0.5`}>見積番号</dt>
              <dd className={`font-mono ${C.text}`}>{estimate.estimateNo}</dd>
            </div>
            {estimate.ownerName ? (
              <div>
                <dt className={`${C.text50} mb-0.5`}>飼主名</dt>
                <dd className={C.text}>{estimate.ownerName}</dd>
              </div>
            ) : null}
            {estimate.validUntil ? (
              <div>
                <dt className={`${C.text50} mb-0.5`}>有効期限</dt>
                <dd className={C.text}>{estimate.validUntil.slice(0, 10)}</dd>
              </div>
            ) : null}
            <div>
              <dt className={`${C.text50} mb-0.5`}>作成日</dt>
              <dd className={C.text}>{estimate.createdAt.slice(0, 10)}</dd>
            </div>
            <div>
              <dt className={`${C.text50} mb-0.5`}>更新日</dt>
              <dd className={C.text}>{estimate.updatedAt.slice(0, 10)}</dd>
            </div>
          </div>

          {estimate.comment ? (
            <div>
              <dt className={`text-sm ${C.text50} mb-0.5`}>コメント</dt>
              <dd className={`text-sm ${C.text} whitespace-pre-wrap`}>{estimate.comment}</dd>
            </div>
          ) : null}
        </div>

        {/* 見積明細 */}
        <div className={`bg-white border ${C.borderLight} rounded-[6px] p-5`}>
          <h3 className={`text-sm font-medium ${C.text} mb-4`}>見積明細</h3>
          <EstimateLineItems
            items={estimate.items}
            subtotal={estimate.subtotal}
            taxTotal={estimate.taxTotal}
            insuranceAmount={estimate.insuranceAmount}
            discountAmount={estimate.discountAmount}
            totalAmount={estimate.totalAmount}
          />
        </div>

        {/* 備考 */}
        {estimate.notes ? (
          <div className={`bg-white border ${C.borderLight} rounded-[6px] p-5`}>
            <h3 className={`text-sm font-medium ${C.text} mb-2`}>備考</h3>
            <p className={`text-sm ${C.text70} whitespace-pre-wrap`}>{estimate.notes}</p>
          </div>
        ) : null}
      </div>

      <ConfirmDialog
        open={showDeleteDialog}
        onClose={() => setShowDeleteDialog(false)}
        onConfirm={handleDelete}
        title="見積書を削除しますか?"
        description="この操作は取り消せません。"
        confirmLabel="削除"
        variant="destructive"
      />
    </PageLayout>
  );
}
