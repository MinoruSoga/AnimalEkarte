import { memo, useState } from "react";
import { Shield } from "lucide-react";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { PropInput } from "@/components/shared/SidePeek/PropInput";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { C, LAYOUT } from "@/lib/design-tokens";
import { MASTER_INPUT_CLASS, MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";
import { useMasterSave } from "@/features/master/hooks/use-master-save";
import { MasterCRUDPage } from "@/features/master/components/MasterCRUDPage";
import {
  useGetAllInsurances, useCreateInsurance, useUpdateInsurance, useDeleteInsurance,
} from "@/features/master/api/insurances";
import type { Insurance, CreateInsuranceRequest, UpdateInsuranceRequest } from "@/features/master/api/insurances";

const COLUMNS = [
  { header: "名称", className: "flex-1" },
  { header: "補償率", className: "w-[100px]", align: "center" as const },
  { header: "連絡先", className: "w-[140px]" },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

interface InsuranceFormData { name: string; description: string; coverageRate: string; contactPhone: string; isActive: boolean; }

const InsuranceSidePanel = memo(function InsuranceSidePanel({
  item, onClose, onSave, onDeleteRequest,
}: { item: Insurance | null; onClose: () => void; onSave: (d: InsuranceFormData) => void; onDeleteRequest: (i: Insurance) => void; }) {
  const [f, setF] = useState<InsuranceFormData>(() => ({
    name: item?.name ?? "", description: item?.description ?? "",
    coverageRate: item?.coverageRate != null ? String(item.coverageRate) : "0",
    contactPhone: item?.contactPhone ?? "", isActive: item?.isActive ?? true,
  }));
  return (
    <MasterSidePanel isNew={item === null} title={f.name}
      onTitleChange={(v) => setF((p) => ({ ...p, name: v }))} onClose={onClose} onSave={() => onSave(f)}
      onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
      icon={<Shield className={LAYOUT.pageIcon.innerIcon} />}>
      <StatusToggleButton isActive={f.isActive} onToggle={() => setF((p) => ({ ...p, isActive: !p.isActive }))} />
      <PropertyRow label="補償率(%)">
        <input type="number" min={0} max={100} className={MASTER_INPUT_CLASS}
          value={f.coverageRate} onChange={(e) => setF((p) => ({ ...p, coverageRate: e.target.value }))} placeholder="0" />
      </PropertyRow>
      <PropertyRow label="連絡先">
        <input type="tel" className={MASTER_INPUT_CLASS}
          value={f.contactPhone} onChange={(e) => setF((p) => ({ ...p, contactPhone: e.target.value }))} placeholder="電話番号" />
      </PropertyRow>
      <PropertyRow label="備考">
        <PropInput value={f.description} onChange={(v) => setF((p) => ({ ...p, description: v }))} placeholder="補足情報など" />
      </PropertyRow>
    </MasterSidePanel>
  );
});

export function InsuranceSettings() {
  const { data } = useGetAllInsurances();
  const createMutation = useCreateInsurance();
  const updateMutation = useUpdateInsurance();
  const deleteMutation = useDeleteInsurance();
  const crud = useMasterCRUD<Insurance>({ data, deleteMutation, entityLabel: "保険" });
  const { handleSave } = useMasterSave<Insurance, InsuranceFormData, CreateInsuranceRequest, UpdateInsuranceRequest>({
    crud, createMutation, updateMutation,
    validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: (d) => ({
      name: d.name, description: d.description || undefined,
      coverage_rate: d.coverageRate !== "" ? Number(d.coverageRate) : 0,
      contact_phone: d.contactPhone || undefined, is_active: d.isActive,
    }),
    toUpdateRequest: (d) => ({
      name: d.name, description: d.description || undefined,
      coverage_rate: d.coverageRate !== "" ? Number(d.coverageRate) : 0,
      contact_phone: d.contactPhone || undefined, is_active: d.isActive,
    }),
  });

  return (
    <MasterCRUDPage title="保険マスタ" icon={<Shield className="size-5 text-[#37352F]" />}
      entityLabel="保険" searchPlaceholder="保険名で検索..." emptyMessage="保険が登録されていません"
      crud={crud} handleSave={handleSave} columns={COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      renderRow={(item, onEdit) => (
        <DataTableRow key={item.id} onClick={() => onEdit(item)}>
          <TableCell className={`font-medium text-sm ${C.text}`}>{item.name}</TableCell>
          <TableCell className={`text-sm text-center ${C.text}`}>{item.coverageRate > 0 ? `${item.coverageRate}%` : "-"}</TableCell>
          <TableCell className={`text-sm ${C.text70}`}>{item.contactPhone || "-"}</TableCell>
          <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="p-0 text-right"><RowActionButton onClick={() => onEdit(item)} /></TableCell>
        </DataTableRow>
      )}
      renderSidePanel={(props) => <InsuranceSidePanel key={props.item?.id ?? "new"} {...props} />}
    />
  );
}
