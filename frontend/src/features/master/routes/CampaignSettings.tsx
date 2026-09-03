import { useCallback } from "react";
import { Tag } from "lucide-react";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { C, ICON } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";
import { MASTER_STATUS_FILTER, MASTER_TABLE_COL } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { CampaignSidePanel } from "../components/CampaignSidePanel";
import type { CampaignFormData } from "../lib/campaign-side-panel-model";
import {
  useGetCampaigns,
  useCreateCampaign,
  useUpdateCampaign,
  useDeleteCampaign,
} from "../api/campaign";
import type { CreateCampaignRequest, Campaign, UpdateCampaignRequest } from "../api/campaign";
import { buildCampaignCreateRequest, buildCampaignUpdateRequest } from "./campaign-settings-model";
import { ResourceAccounting } from "@/types/generated/models";

// ─── Constants ───
const COLUMNS = [
  { header: "名称", className: "flex-1" },
  { header: "期間", className: MASTER_TABLE_COL.w180 },
  { header: "割引", className: MASTER_TABLE_COL.w100, align: "right" as const },
  { header: "ステータス", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "操作", className: MASTER_TABLE_COL.w80, align: "right" as const },
];

// ─── Page ───
export function CampaignSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceAccounting);
  const { data } = useGetCampaigns();
  const createMutation = useCreateCampaign();
  const updateMutation = useUpdateCampaign();
  const deleteMutation = useDeleteCampaign();

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<Campaign>({
    data,
    deleteMutation,
    entityLabel: "キャンペーン",
    dirtyGuard: dirty,
    permissions: { canDelete },
  });
  const handleDirtyChange = useCallback(
    (isDirty: boolean) => (isDirty ? dirty.markDirty() : dirty.markClean()),
    [dirty],
  );

  const { handleSave } = useMasterSave<
    Campaign,
    CampaignFormData,
    CreateCampaignRequest,
    UpdateCampaignRequest
  >({
    crud,
    createMutation,
    updateMutation,
    validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: buildCampaignCreateRequest,
    toUpdateRequest: buildCampaignUpdateRequest,
    permissions: { canCreate, canEdit },
  });

  return (
    <>
      <MasterCRUDPage
        title="割引キャンペーンマスタ"
        icon={<Tag className={`${ICON.page} ${C.text}`} />}
        resource={ResourceAccounting}
        entityLabel="キャンペーン"
        searchPlaceholder="キャンペーン名で検索..."
        emptyMessage="キャンペーンが登録されていません"
        crud={crud}
        handleSave={handleSave}
        columns={COLUMNS}
        filterProperties={[MASTER_STATUS_FILTER]}
        renderRow={(item, onEdit, canEdit) => (
          <DataTableRow key={item.id}>
            <TableCell className={`font-medium ${C.text}`}>
              <DataTableRowButton
                aria-label={`詳細: キャンペーン ${item.name} (ID ${item.id})`}
                onClick={() => onEdit(item)}
              >
                {item.name}
              </DataTableRowButton>
            </TableCell>
            <TableCell className={`text-sm ${C.text50}`}>
              {item.startDate} 〜 {item.endDate}
            </TableCell>
            <TableCell className="text-right text-sm">
              {item.discountType === "rate"
                ? `${item.discountValue}%`
                : formatCurrency(item.discountValue)}
            </TableCell>
            <TableCell className="text-center">
              <StatusPill isActive={item.isActive} />
            </TableCell>
            <TableCell className="text-right">
              {canEdit ? (
                <RowActionButton
                  onClick={() => onEdit(item)}
                  aria-label={`キャンペーン「${item.name}」(ID: ${item.id}) を編集`}
                />
              ) : null}
            </TableCell>
          </DataTableRow>
        )}
        renderSidePanel={(props) => (
          <CampaignSidePanel
            key={props.item?.id ?? "new"}
            {...props}
            onDirtyChange={handleDirtyChange}
          />
        )}
      />
      {dirty.discardDialog}
    </>
  );
}
