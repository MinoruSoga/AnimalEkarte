import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router";
import { FlaskConical } from "lucide-react";
import { toast } from "sonner";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { Input } from "@/components/ui/input";
import { TableCell } from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { C, ICON, LAYOUT, STYLE } from "@/lib/design-tokens";
import { ResourceLabImport } from "@/types/generated/models";

import { useGetAllExaminationTypes } from "../api/exam-types-master";
import {
  useEnsureLabDeviceItemMasters,
  useGetLabDeviceItemMasters,
  useUpdateLabDeviceItemMaster,
  type LabDeviceItemMaster,
} from "../api/lab-device-item-masters";
import { MASTER_TABLE_COL } from "../constants/styles";
import {
  LAB_DEVICE_UNMAPPED_FIELD,
  buildExamFieldOptions,
  buildLabDeviceItemMasterUpdateRequest,
  examFieldOptionsForItem,
  examFieldSelectValue,
  groupLabDeviceItemMasters,
  labDeviceValueShapeLabel,
  parseExamFieldSelectValue,
  validateLabDeviceItemMasterDraft,
} from "./lab-device-item-master-settings-model";

const COLUMNS = [
  { header: "コード", className: MASTER_TABLE_COL.w120 },
  { header: "表示名", className: MASTER_TABLE_COL.w150 },
  { header: "単位", className: MASTER_TABLE_COL.w100 },
  { header: "形", className: MASTER_TABLE_COL.w120 },
  { header: "載せる先", className: "flex-1 min-w-[220px]" },
  { header: "有効", className: MASTER_TABLE_COL.w100, align: "center" as const },
];

export function LabDeviceItemMasterSettings() {
  const navigate = useNavigate();
  const { canEdit } = usePermission(ResourceLabImport);
  const { data: items = [], isLoading } = useGetLabDeviceItemMasters();
  const { data: examTypes = [] } = useGetAllExaminationTypes();
  const updateMutation = useUpdateLabDeviceItemMaster();
  const ensureMutation = useEnsureLabDeviceItemMasters();
  const [nameDrafts, setNameDrafts] = useState<Record<string, string>>({});

  useEffect(() => {
    setNameDrafts((prev) => {
      const next = { ...prev };
      for (const item of items) {
        if (next[item.id] === undefined) {
          next[item.id] = item.displayName;
        }
      }
      return next;
    });
  }, [items]);

  const fieldOptions = useMemo(() => buildExamFieldOptions(examTypes), [examTypes]);
  const groups = useMemo(() => groupLabDeviceItemMasters(items), [items]);

  const persist = useCallback(
    (item: LabDeviceItemMaster, patch: Partial<Pick<LabDeviceItemMaster, "displayName" | "examTypeFieldId" | "isActive">>) => {
      const next = {
        displayName: patch.displayName ?? item.displayName,
        unit: item.unit,
        examTypeFieldId: patch.examTypeFieldId === undefined ? item.examTypeFieldId : patch.examTypeFieldId,
        isActive: patch.isActive ?? item.isActive,
      };
      const error = validateLabDeviceItemMasterDraft(next);
      if (error) {
        toast.error(error);
        return;
      }
      updateMutation.mutate({
        id: item.id,
        req: buildLabDeviceItemMasterUpdateRequest(next),
      });
    },
    [updateMutation],
  );

  return (
    <PageLayout
      title="検査機器マスタ"
      description="NX600 / AU10V / 尿の項目を検査項目へ載せます。日常の送信画面には出しません"
      icon={<FlaskConical className={`${ICON.page} ${C.text}`} />}
      resource={ResourceLabImport}
      onBack={() => navigate(paths.settings.getHref())}
      maxWidth={LAYOUT.pageContentMaxWidth.full}
      headerAction={
        canEdit ? (
          <PrimaryButton
            onClick={() => {
              ensureMutation.mutate(undefined, {
                onSuccess: (result) => {
                  toast.success(
                    result.insertedCount > 0
                      ? `既定項目を ${result.insertedCount} 件用意しました`
                      : "既定項目は揃っています",
                  );
                },
              });
            }}
            disabled={ensureMutation.isPending}
          >
            既定項目を用意
          </PrimaryButton>
        ) : null
      }
    >
      {isLoading ? (
        <p className={`text-sm ${C.text40}`}>読み込み中...</p>
      ) : groups.length === 0 ? (
        <p className={`text-sm ${C.text40}`}>
          まだ項目がありません。既定項目を用意すると 25 件が投入されます
        </p>
      ) : (
        <div className="flex flex-col gap-8">
          {groups.map((group) => (
            <section key={group.sourceType} className="flex flex-col gap-3">
              <h2 className={`text-base font-medium ${C.text}`}>{group.label}</h2>
              <DataTable
                columns={COLUMNS}
                data={group.items}
                emptyMessage="項目がありません"
                renderRow={(item) => (
                  <DataTableRow key={item.id}>
                    <TableCell className={`font-medium ${C.text}`}>{item.deviceItemCode}</TableCell>
                    <TableCell>
                      <Input
                        value={nameDrafts[item.id] ?? item.displayName}
                        disabled={!canEdit}
                        aria-label={`${item.deviceItemCode}の表示名`}
                        onChange={(event) => {
                          const value = event.target.value;
                          setNameDrafts((prev) => ({ ...prev, [item.id]: value }));
                        }}
                        onBlur={() => {
                          const draft = (nameDrafts[item.id] ?? item.displayName).trim();
                          if (draft === item.displayName) {
                            return;
                          }
                          persist(item, { displayName: draft });
                        }}
                      />
                    </TableCell>
                    <TableCell className={C.text70}>{item.unit || "-"}</TableCell>
                    <TableCell className={C.text70}>
                      {labDeviceValueShapeLabel(item.valueShape)}
                    </TableCell>
                    <TableCell>
                      <Select
                        value={examFieldSelectValue(item.examTypeFieldId)}
                        disabled={!canEdit}
                        onValueChange={(value) => {
                          persist(item, { examTypeFieldId: parseExamFieldSelectValue(value) });
                        }}
                      >
                        <SelectTrigger
                          className={STYLE.selectCompact}
                          aria-label={`${item.deviceItemCode}の載せる先`}
                        >
                          <SelectValue placeholder="未設定" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value={LAB_DEVICE_UNMAPPED_FIELD}>未設定</SelectItem>
                          {examFieldOptionsForItem(fieldOptions, item.examTypeFieldId).map((option) => (
                            <SelectItem key={option.id} value={option.id}>
                              {option.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </TableCell>
                    <TableCell className="text-center">
                      {canEdit ? (
                        <button
                          type="button"
                          onClick={() => persist(item, { isActive: !item.isActive })}
                          aria-label={`${item.displayName}の有効を切り替え`}
                          className={`inline-flex items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
                        >
                          <StatusPill isActive={item.isActive} />
                        </button>
                      ) : (
                        <StatusPill isActive={item.isActive} />
                      )}
                    </TableCell>
                  </DataTableRow>
                )}
              />
            </section>
          ))}
        </div>
      )}
    </PageLayout>
  );
}
