import { memo, useState, useCallback } from "react";
import { Bed } from "lucide-react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { TableCell } from "@/components/ui/table";
import { TaxTypeSelector } from "@/components/shared/TaxTypeSelector/TaxTypeSelector";
import { TaxRateSelector } from "@/components/shared/TaxRateSelector/TaxRateSelector";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { MoneyInput } from "@/components/shared/SidePeek/MoneyInput";
import { PropertyInput } from "@/components/shared/SidePeek/PropertyInput";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { C, STYLE, LAYOUT, ICON } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";
import { useMasterSave } from "@/features/master/hooks/use-master-save";
import { MasterCRUDPage } from "@/features/master/components/MasterCRUDPage";
import {
  useGetAllHospitalizationPlans, useCreateHospitalizationPlan, useUpdateHospitalizationPlan, useDeleteHospitalizationPlan,
  BODY_SIZE_OPTIONS, BODY_SIZE_LABELS, BILLING_UNIT_OPTIONS, BILLING_UNIT_LABELS,
} from "@/features/master/api/hospitalization-plans";
import type { HospitalizationPlan, CreateHospitalizationPlanRequest, UpdateHospitalizationPlanRequest } from "@/features/master/api/hospitalization-plans";
import type { BodySize, BillingUnit, TaxType } from "@/types/generated/models";

const COLUMNS = [
  { header: "名称" }, { header: "対象体格", className: "w-[100px]" },
  { header: "料金単位", className: "w-[120px]" }, { header: "単価(税込)", className: "w-[120px]", align: "right" as const },
  { header: "ステータス", className: "w-[100px]", align: "center" as const }, { header: "操作", className: "w-[80px]", align: "right" as const },
];

const BODY_SIZE_SELECT_ITEMS = BODY_SIZE_OPTIONS.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>);
const BILLING_UNIT_SELECT_ITEMS = BILLING_UNIT_OPTIONS.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>);

interface HospitalizationFormData { name: string; price: number; description: string; isActive: boolean; bodySize: BodySize | ""; billingUnit: BillingUnit | ""; taxType: TaxType; taxRate: number; }

const HospitalizationSidePanel = memo(function HospitalizationSidePanel({
  item, onClose, onSave, onDeleteRequest,
}: { item: HospitalizationPlan | null; onClose: () => void; onSave: (d: HospitalizationFormData) => void; onDeleteRequest: (i: HospitalizationPlan) => void; }) {
  const [f, setF] = useState<HospitalizationFormData>(() => ({
    name: item?.name ?? "", price: item?.price ?? 0, description: item?.description ?? "",
    isActive: item?.isActive ?? true, bodySize: item?.bodySize ?? "", billingUnit: item?.billingUnit ?? "",
    taxType: item?.taxType ?? "excluded", taxRate: item?.taxRate ?? 0.1,
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  const handleTitleChange = useCallback((v: string) => {
    setF((p) => ({ ...p, name: v }));
    setIsDirty(true);
    if (v.trim()) setNameError("");
  }, []);

  const handleBodySizeChange = useCallback((v: string) => {
    setF((p) => ({ ...p, bodySize: v as BodySize }));
    setIsDirty(true);
  }, []);

  const handleBillingUnitChange = useCallback((v: string) => {
    setF((p) => ({ ...p, billingUnit: v as BillingUnit }));
    setIsDirty(true);
  }, []);

  const handlePriceChange = useCallback((v: number) => {
    setF((p) => ({ ...p, price: v }));
    setIsDirty(true);
  }, []);

  const handleTaxTypeChange = useCallback((v: TaxType) => {
    setF((p) => ({ ...p, taxType: v }));
    setIsDirty(true);
  }, []);

  const handleTaxRateChange = useCallback((v: number) => {
    setF((p) => ({ ...p, taxRate: v }));
    setIsDirty(true);
  }, []);

  const handleDescriptionChange = useCallback((v: string) => {
    setF((p) => ({ ...p, description: v }));
    setIsDirty(true);
  }, []);

  const handleToggleActive = useCallback(() => {
    setF((p) => ({ ...p, isActive: !p.isActive }));
    setIsDirty(true);
  }, []);

  const handleAction = useCallback(() => {
    if (!f.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    onSave(f);
    setIsDirty(false);
  }, [f, onSave]);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel isNew={item === null} title={f.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose} action={handleAction} onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
      icon={<Bed className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}>
      <StatusToggleButton isActive={f.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="対象体格">
        <Select value={f.bodySize} onValueChange={handleBodySizeChange}>
          <SelectTrigger className={STYLE.selectCompact}><SelectValue placeholder="選択" /></SelectTrigger>
          <SelectContent>{BODY_SIZE_SELECT_ITEMS}</SelectContent>
        </Select>
      </PropertyRow>
      <PropertyRow label="料金単位">
        <Select value={f.billingUnit} onValueChange={handleBillingUnitChange}>
          <SelectTrigger className={STYLE.selectCompact}><SelectValue placeholder="選択" /></SelectTrigger>
          <SelectContent>{BILLING_UNIT_SELECT_ITEMS}</SelectContent>
        </Select>
      </PropertyRow>
      <MoneyInput value={f.price} onChange={handlePriceChange} />
      <PropertyRow label="課税区分">
        <TaxTypeSelector
          value={f.taxType}
          onChange={handleTaxTypeChange}
        />
      </PropertyRow>
      <PropertyRow label="税率">
        <TaxRateSelector
          value={f.taxRate}
          onChange={handleTaxRateChange}
        />
      </PropertyRow>
      <PropertyRow label="備考">
        <PropertyInput value={f.description} onChange={handleDescriptionChange} placeholder="補足情報など" />
      </PropertyRow>
    </MasterSidePanel>
  );
});

export function HospitalizationSettings() {
  const { data } = useGetAllHospitalizationPlans();
  const createMutation = useCreateHospitalizationPlan();
  const updateMutation = useUpdateHospitalizationPlan();
  const deleteMutation = useDeleteHospitalizationPlan();
  const crud = useMasterCRUD<HospitalizationPlan>({ data, deleteMutation, entityLabel: "入院プラン" });
  const { handleSave } = useMasterSave<HospitalizationPlan, HospitalizationFormData, CreateHospitalizationPlanRequest, UpdateHospitalizationPlanRequest>({
    crud, createMutation, updateMutation,
    validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: (d) => ({
      name: d.name, price: d.price || undefined, description: d.description || undefined, is_active: d.isActive,
      body_size: d.bodySize !== "" ? d.bodySize : undefined, billing_unit: d.billingUnit !== "" ? d.billingUnit : undefined,
      tax_type: d.taxType, tax_rate: d.taxRate,
    }),
    toUpdateRequest: (d) => ({
      name: d.name, price: d.price, description: d.description || undefined, is_active: d.isActive,
      body_size: d.bodySize !== "" ? d.bodySize : null, billing_unit: d.billingUnit !== "" ? d.billingUnit : null,
      tax_type: d.taxType, tax_rate: d.taxRate,
    }),
  });

  return (
    <MasterCRUDPage title="入院マスタ" icon={<Bed className={`${ICON.page} text-[#37352F]`} />}
      entityLabel="入院プラン" searchPlaceholder="名称で検索..." emptyMessage="入院プランが登録されていません"
      crud={crud} handleSave={handleSave} columns={COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      renderRow={(item, onEdit) => (
        <DataTableRow key={item.id} onClick={() => onEdit(item)}>
          <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
          <TableCell className={`text-base ${C.text70}`}>{item.bodySize ? (BODY_SIZE_LABELS[item.bodySize] ?? item.bodySize) : "-"}</TableCell>
          <TableCell className={`text-base ${C.text70}`}>{item.billingUnit ? (BILLING_UNIT_LABELS[item.billingUnit] ?? item.billingUnit) : "-"}</TableCell>
          <TableCell className={`text-right font-mono text-base ${C.text}`}>{item.price > 0 ? `¥${item.price.toLocaleString()}` : "-"}</TableCell>
          <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="p-0 text-right"><RowActionButton onClick={() => onEdit(item)} /></TableCell>
        </DataTableRow>
      )}
      renderSidePanel={(props) => <HospitalizationSidePanel key={props.item?.id ?? "new"} {...props} />}
    />
  );
}
