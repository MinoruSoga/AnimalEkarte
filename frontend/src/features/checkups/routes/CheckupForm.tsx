// React/Framework
import { useCallback, useMemo } from "react";
import { useNavigate } from "react-router";

// External
import { ClipboardCheck } from "lucide-react";

// Internal
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PatientInfoCard, formatPatientPetDetails } from "@/components/shared/PatientInfoCard";
import { PastRecordHistoryPanel } from "@/components/shared/PastRecordHistoryPanel";
import { FormHeaderActions } from "@/components/shared/Form/FormHeaderActions";
import { NextScheduleField } from "@/components/shared/NextScheduleField";
import { DatePicker } from "@/components/shared/DatePicker/DatePicker";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { LoadingFallback } from "@/components/shared/DataStates";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { SearchableSelect } from "@/components/ui/searchable-select";
import { paths } from "@/config/paths";
import { toJSTWallDate } from "@/lib/jst-date";
import { formatDate } from "@/lib/format/date";
import { useGetAllCheckupTypes } from "@/hooks/use-treatment-master";
import { useGetStaffs } from "@/hooks/use-staffs";
import { usePermission } from "@/hooks/use-permission";
import { ResourceMedicalRecords } from "@/types/generated/models";

// Relative
import { useCheckupForm } from "../hooks/use-checkup-form";
import { DynamicCheckupFields } from "../components/DynamicCheckupFields";
import { useGetCheckups } from "../api/get-checkups";

export function CheckupForm() {
  const navigate = useNavigate();
  const { canCreate, canEdit } = usePermission(ResourceMedicalRecords);
  const canSubmit = canCreate && canEdit;

  const {
    pet,
    isPetLoading,
    form,
    formAction,
    isPending,
    fieldErrors,
    checkupFields,
    fieldValues,
    setFieldValue,
    setCheckupTypeId,
    setDate,
    setNextScheduleType,
    setNextDate,
    setDoctorId,
    setResult,
  } = useCheckupForm({ canCreate, canEdit });

  const { data: checkupTypes = [] } = useGetAllCheckupTypes();
  const { data: staffs = [] } = useGetStaffs();
  const { data: checkupsResult, isLoading: isHistoryLoading } = useGetCheckups({
    page: 1,
    limit: 100,
  });
  const doctorName = staffs.find((staff) => staff.id === form.doctorId)?.name ?? "";
  const historyItems = useMemo(() => {
    if (!pet?.id) return [];
    return (checkupsResult?.data ?? [])
      .filter((record) => String(record.petId) === String(pet.id))
      .map((record) => ({
        id: String(record.id),
        date: record.date,
        title: record.checkupTypeName || "健診",
        subtitle: [record.doctorName, record.result].filter(Boolean).join(" / ") || undefined,
      }));
  }, [checkupsResult?.data, pet?.id]);

  const handleBack = useCallback(() => {
    navigate(paths.checkups.getHref());
  }, [navigate]);

  const guardedFormAction = useCallback((formData: FormData) => {
    if (!canSubmit) return;
    formAction(formData);
  }, [canSubmit, formAction]);

  if (isPetLoading) return <LoadingFallback />;

  return (
    <form aria-label="定期健診登録フォーム" action={guardedFormAction}>
      <PageLayout
        title="定期健診登録"
        resource={ResourceMedicalRecords}
        icon={<ClipboardCheck className={`${ICON.page} ${C.text}`} />}
        onBack={handleBack}
        maxWidth={LAYOUT.pageContentMaxWidth.form}
        headerAction={
          <FormHeaderActions
            onCancel={handleBack}
            submitLabel={isPending ? "保存中..." : "保存"}
            submitDisabled={isPending || !canSubmit}
          />
        }
      >
        {pet ? (
          <PatientInfoCard
            ownerName={pet.ownerName ?? ""}
            petName={pet.name}
            petNumber={pet.petNumber ?? ""}
            weight={pet.weight ?? ""}
            petDetails={formatPatientPetDetails({
              birthDate: pet.birthDate,
              gender: pet.gender,
              neuteredDate: pet.neuteredDate,
            })}
            insuranceName={pet.insuranceName}
            insuranceDetails={pet.insuranceDetails}
            staffName={doctorName}
            nextVisitDate={form.nextDate ? formatDate(form.nextDate) : undefined}
            status={pet.status === "死亡" ? "deceased" : "alive"}
          />
        ) : null}

        <div className="mt-4 grid grid-cols-1 gap-6 lg:grid-cols-5">
        <fieldset
          aria-label="定期健診入力"
          disabled={!canSubmit}
          className={`lg:col-span-3 ${C.bgWhite} p-6 rounded-lg border ${C.borderLight} space-y-6 min-w-0`}
        >
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* 実施日 */}
            <div className="space-y-2">
              <Label htmlFor="checkup-date">
                実施日<span className={`${C.textRequired} ml-1`}>*</span>
              </Label>
              <DatePicker
                id="checkup-date"
                value={form.date}
                onChange={setDate}
                disabledDays={{ after: toJSTWallDate(new Date()) }}
              />
              <FormFieldError message={fieldErrors.date} />
            </div>

            {/* 健診種別 */}
            <div className="space-y-2">
              <Label htmlFor="checkup-type-select">
                健診種別<span className={`${C.textRequired} ml-1`}>*</span>
              </Label>
              <SearchableSelect
                id="checkup-type-select"
                value={form.checkupTypeId}
                onValueChange={setCheckupTypeId}
                options={checkupTypes.map((ct) => ({ value: String(ct.id), label: ct.name }))}
                placeholder="選択してください"
                searchPlaceholder="健診種別を検索..."
              />
              <FormFieldError message={fieldErrors.checkupTypeId} />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <NextScheduleField
              typeId="checkup-next-schedule"
              dateId="checkup-next-date"
              scheduleType={form.nextScheduleType}
              nextDate={form.nextDate}
              onScheduleTypeChange={setNextScheduleType}
              onNextDateChange={setNextDate}
            />

            {/* 担当医 */}
            <div className="space-y-2">
              <Label htmlFor="checkup-doctor-select">担当医</Label>
              <SearchableSelect
                id="checkup-doctor-select"
                value={form.doctorId}
                onValueChange={setDoctorId}
                options={staffs
                  .filter((s) => s.isActive)
                  .map((s) => ({ value: s.id, label: s.name }))}
                placeholder="選択してください"
                searchPlaceholder="担当医を検索..."
              />
            </div>
          </div>

          {/* #211 健診パッケージの型付き入力（選択種別にフィールド定義がある場合のみ表示） */}
          {checkupFields.length > 0 ? (
            <div className={`border-t ${C.borderLight} pt-6`}>
              <h2 className={`mb-4 text-sm font-medium ${C.text}`}>健診項目</h2>
              <DynamicCheckupFields
                fields={checkupFields}
                values={fieldValues}
                onChange={setFieldValue}
              />
            </div>
          ) : null}

          {/* 結果・所見 */}
          <div className="space-y-2">
            <Label htmlFor="checkup-result">結果・所見</Label>
            <Textarea
              id="checkup-result"
              value={form.result}
              onChange={(e) => setResult(e.target.value)}
              placeholder="健診結果・所見を入力"
              className="min-h-[120px]"
            />
          </div>
        </fieldset>
        <PastRecordHistoryPanel
          title="過去の健診履歴"
          searchPlaceholder="健診種別・所見で検索..."
          items={historyItems}
          isLoading={isHistoryLoading}
        />
        </div>
      </PageLayout>
    </form>
  );
}
