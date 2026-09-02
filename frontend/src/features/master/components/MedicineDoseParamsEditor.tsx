import { useCallback, useImperativeHandle, useRef, useState, type Ref } from "react";

import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { PropertyInput, PropertyRow } from "@/components/shared/SidePeek";
import { C, STYLE } from "@/lib/design-tokens";
import type { MedicineDoseParam, MedicineDoseSpecies } from "@/types/generated/models";
import {
  MedicineDoseBasisPerAdministration,
  MedicineDoseBasisPerDay,
  MedicineDoseSpeciesCat,
  MedicineDoseSpeciesDog,
  MedicineRoundingModeDown,
  MedicineRoundingModeNearest,
  MedicineRoundingModeUp,
} from "@/types/generated/models";

import {
  useDeleteMedicineDoseParam,
  useGetMedicineDoseParams,
  useUpsertMedicineDoseParam,
} from "../api/medicine-dose-params";
import {
  buildUpsertDoseParamRequest,
  doseParamToFormData,
  findDoseParamBySpecies,
  isDoseParamFormEmpty,
  validateDoseParamForm,
  type DoseParamFormData,
} from "./medicine-dose-params-editor-model";
import { SELECT_TRIGGER_FULL } from "./MedicineSidePanelSections";

const SPECIES_LABEL: Record<MedicineDoseSpecies, string> = {
  [MedicineDoseSpeciesDog]: "犬",
  [MedicineDoseSpeciesCat]: "猫",
};

type MedicineDoseParamDraft = {
  species: MedicineDoseSpecies;
  input: ReturnType<typeof buildUpsertDoseParamRequest>;
};

export type MedicineDoseParamsEditorHandle = {
  saveFilled: () => Promise<boolean>;
  collectFilled: () => Promise<MedicineDoseParamDraft[] | false>;
};

interface MedicineDoseParamsEditorProps {
  medicineId: string;
  ref?: Ref<MedicineDoseParamsEditorHandle>;
}

type DoseParamPanelHandle = {
  saveFilled: () => Promise<boolean>;
  collectFilled: () => Promise<MedicineDoseParamDraft | null | false>;
};

/**
 * #201 種別（犬・猫）投与量パラメータ編集。製品軸(calculation_type)とは別 API で保存する。
 * 親パネルの「更新」からも saveFilled で upsert する（BUG-014）。
 */
export function MedicineDoseParamsEditor({ medicineId, ref }: MedicineDoseParamsEditorProps) {
  const { data: doseParams = [], isLoading } = useGetMedicineDoseParams(medicineId);
  const dogRef = useRef<DoseParamPanelHandle>(null);
  const catRef = useRef<DoseParamPanelHandle>(null);

  useImperativeHandle(ref, () => ({
    saveFilled: async () => {
      const dogOk = (await dogRef.current?.saveFilled()) ?? true;
      const catOk = (await catRef.current?.saveFilled()) ?? true;
      return dogOk && catOk;
    },
    collectFilled: async () => {
      const drafts: MedicineDoseParamDraft[] = [];
      const dog = await dogRef.current?.collectFilled();
      if (dog === false) return false;
      if (dog) drafts.push(dog);
      const cat = await catRef.current?.collectFilled();
      if (cat === false) return false;
      if (cat) drafts.push(cat);
      return drafts;
    },
  }));

  return (
    <>
      <div className={`${STYLE.sectionDivider} mt-3 mb-1`} />
      <div className="py-1">
        <div className="flex items-center gap-1.5 py-2 mb-1">
          <span className={`${STYLE.sectionLabel}`}>
            種別パラメータ（犬・猫）
          </span>
        </div>
        <p className={`text-sm ${C.text40} px-1 pb-2`}>
          過量防止のため、上限(mg/kgまたはmg)のいずれかは必須です。投与量は下限・上限の範囲内で入力してください。
        </p>
        {isLoading ? (
          <p className={`text-sm ${C.text40} px-1`}>読み込み中...</p>
        ) : (
          <>
            <SpeciesDoseParamPanel
              key={`dog-${findDoseParamBySpecies(doseParams, MedicineDoseSpeciesDog)?.id ?? "new"}`}
              medicineId={medicineId}
              species={MedicineDoseSpeciesDog}
              existingParam={findDoseParamBySpecies(doseParams, MedicineDoseSpeciesDog)}
              ref={dogRef}
            />
            <SpeciesDoseParamPanel
              key={`cat-${findDoseParamBySpecies(doseParams, MedicineDoseSpeciesCat)?.id ?? "new"}`}
              medicineId={medicineId}
              species={MedicineDoseSpeciesCat}
              existingParam={findDoseParamBySpecies(doseParams, MedicineDoseSpeciesCat)}
              ref={catRef}
            />
          </>
        )}
      </div>
    </>
  );
}

interface SpeciesDoseParamPanelProps {
  medicineId: string;
  species: MedicineDoseSpecies;
  existingParam: MedicineDoseParam | undefined;
  ref?: Ref<DoseParamPanelHandle>;
}

function SpeciesDoseParamPanel({ medicineId, species, existingParam, ref }: SpeciesDoseParamPanelProps) {
  const [formData, setFormData] = useState<DoseParamFormData>(() => doseParamToFormData(existingParam));
  const [clientErrors, setClientErrors] = useState<string[]>([]);
  const upsertMutation = useUpsertMedicineDoseParam(medicineId);
  const deleteMutation = useDeleteMedicineDoseParam(medicineId);

  const collectFilled = useCallback(async (): Promise<MedicineDoseParamDraft | null | false> => {
    if (isDoseParamFormEmpty(formData) && !existingParam) {
      return null;
    }
    const validation = validateDoseParamForm(formData);
    setClientErrors(validation.errors);
    if (!validation.isValid) return false;
    return { species, input: buildUpsertDoseParamRequest(formData) };
  }, [existingParam, formData, species]);

  const saveFilled = useCallback(async (): Promise<boolean> => {
    const draft = await collectFilled();
    if (draft === false) return false;
    if (draft == null) return true;
    if (!medicineId) return true;
    try {
      await upsertMutation.mutateAsync({ species: draft.species, input: draft.input });
      return true;
    } catch {
      return false;
    }
  }, [collectFilled, medicineId, upsertMutation]);

  useImperativeHandle(ref, () => ({ saveFilled, collectFilled }), [saveFilled, collectFilled]);

  const handleSave = () => {
    void saveFilled();
  };

  const handleDelete = () => {
    deleteMutation.mutate(species, {
      onSuccess: () => {
        setFormData(doseParamToFormData(undefined));
        setClientErrors([]);
      },
    });
  };

  return (
    <div className="py-2 px-1 mb-2 rounded-xxs border border-black/5">
      <div className="flex items-center justify-between px-1 py-1">
        <span className={`text-sm font-medium ${C.text65}`}>{SPECIES_LABEL[species]}</span>
        {existingParam ? (
          <Button
            variant="ghost-danger"
            size="sm"
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
          >
            削除
          </Button>
        ) : null}
      </div>

      <PropertyRow label="投与基準">
        <Select
          value={formData.doseBasis}
          onValueChange={(value) =>
            setFormData((prev) => ({ ...prev, doseBasis: value as DoseParamFormData["doseBasis"] }))
          }
        >
          <SelectTrigger className={SELECT_TRIGGER_FULL}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={MedicineDoseBasisPerAdministration}>1回あたり</SelectItem>
            <SelectItem value={MedicineDoseBasisPerDay}>1日あたり</SelectItem>
          </SelectContent>
        </Select>
      </PropertyRow>

      <PropertyRow label="投与量(mg/kg)">
        <input
          type="number"
          min={0}
          step="any"
          aria-label="投与量(mg/kg)"
          value={formData.dosePerKg}
          onChange={(event) => setFormData((prev) => ({ ...prev, dosePerKg: event.target.value }))}
          placeholder="必須"
          className={`${STYLE.propertyInput} w-28`}
        />
      </PropertyRow>

      <PropertyRow label="下限(mg/kg・任意)">
        <input
          type="number"
          min={0}
          step="any"
          aria-label="下限(mg/kg)"
          value={formData.minMgPerKg}
          onChange={(event) => setFormData((prev) => ({ ...prev, minMgPerKg: event.target.value }))}
          placeholder="未設定"
          className={`${STYLE.propertyInput} w-28`}
        />
      </PropertyRow>

      <PropertyRow label="上限(mg/kg・任意)">
        <input
          type="number"
          min={0}
          step="any"
          aria-label="上限(mg/kg)"
          value={formData.maxMgPerKg}
          onChange={(event) => setFormData((prev) => ({ ...prev, maxMgPerKg: event.target.value }))}
          placeholder="未設定"
          className={`${STYLE.propertyInput} w-28`}
        />
      </PropertyRow>

      <PropertyRow label="絶対上限(mg・任意)">
        <input
          type="number"
          min={0}
          step="any"
          aria-label="絶対上限(mg)"
          value={formData.absoluteMaxDose}
          onChange={(event) => setFormData((prev) => ({ ...prev, absoluteMaxDose: event.target.value }))}
          placeholder="未設定"
          className={`${STYLE.propertyInput} w-28`}
        />
      </PropertyRow>

      <PropertyRow label="丸め幅(任意)">
        <input
          type="number"
          min={0}
          step="any"
          aria-label="丸め幅"
          value={formData.roundingStep}
          onChange={(event) => setFormData((prev) => ({ ...prev, roundingStep: event.target.value }))}
          placeholder="未設定"
          className={`${STYLE.propertyInput} w-28`}
        />
      </PropertyRow>

      <PropertyRow label="丸め方向(任意)">
        <Select
          value={formData.roundingMode || "__none__"}
          onValueChange={(value) =>
            setFormData((prev) => ({
              ...prev,
              roundingMode: value === "__none__" ? "" : (value as DoseParamFormData["roundingMode"]),
            }))
          }
        >
          <SelectTrigger className={SELECT_TRIGGER_FULL}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__none__">未設定</SelectItem>
            <SelectItem value={MedicineRoundingModeUp}>切り上げ</SelectItem>
            <SelectItem value={MedicineRoundingModeDown}>切り捨て</SelectItem>
            <SelectItem value={MedicineRoundingModeNearest}>最近接</SelectItem>
          </SelectContent>
        </Select>
      </PropertyRow>

      <PropertyRow label="備考(任意)">
        <PropertyInput
          value={formData.notes}
          onChange={(value) => setFormData((prev) => ({ ...prev, notes: value }))}
          placeholder="空"
        />
      </PropertyRow>

      {clientErrors.length > 0 ? (
        <ul className={`text-sm ${C.danger} px-1 pt-1 list-disc list-inside`}>
          {/* react-review-201 MEDIUM: 同一文言のエラーが複数返り得るため文字列ではなくindex複合キーにする */}
          {clientErrors.map((error, index) => (
            <li key={`${index}-${error}`}>{error}</li>
          ))}
        </ul>
      ) : null}

      {medicineId ? (
        <div className="flex justify-end px-1 pt-2">
          <Button size="sm" variant="outline" onClick={handleSave} disabled={upsertMutation.isPending}>
            {existingParam ? "更新" : "保存"}
          </Button>
        </div>
      ) : null}
    </div>
  );
}
