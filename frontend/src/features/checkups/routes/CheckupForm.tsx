import { useCallback, useMemo } from "react";
import { useNavigate } from "react-router";

import { ClipboardCheck } from "lucide-react";

import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PatientInfoCard, formatPatientPetDetails } from "@/components/shared/PatientInfoCard";
import { PastRecordHistoryPanel } from "@/components/shared/PastRecordHistoryPanel";
import { FormHeaderActions } from "@/components/shared/Form/FormHeaderActions";
import { LoadingFallback } from "@/components/shared/DataStates";
import { paths } from "@/config/paths";
import { formatDate } from "@/lib/format/date";
import { useGetAllCheckupTypes } from "@/hooks/use-treatment-master";
import { useGetStaffs } from "@/hooks/use-staffs";
import { usePermission } from "@/hooks/use-permission";
import { ResourceMedicalRecords } from "@/types/generated/models";

import { useCheckupForm } from "../hooks/use-checkup-form";
import { useGetCheckups } from "../api/get-checkups";
import { toCheckupHistoryItems } from "./checkup-form-model";
import { CheckupFieldsPanel } from "./CheckupFormPanels";

export function CheckupForm() {
  const navigate = useNavigate();
  const { canCreate, canEdit } = usePermission(ResourceMedicalRecords);

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

  // FE-RC-004: 死亡ペットは render 側でも SubmitButton を非表示にする（callback 側の拒否と二重防壁）。
  const isPetDeceased = pet?.status === "死亡";
  const canSubmit = canCreate && canEdit && !isPetDeceased;

  const { data: checkupTypes = [] } = useGetAllCheckupTypes();
  const { data: staffs = [] } = useGetStaffs();
  const { data: checkupsResult, isLoading: isHistoryLoading } = useGetCheckups({
    petId: pet?.id ? String(pet.id) : undefined,
    page: 1,
    limit: 100,
  });
  const doctorName = staffs.find((staff) => staff.id === form.doctorId)?.name ?? "";
  const historyItems = useMemo(() => {
    if (!pet?.id) return [];
    return toCheckupHistoryItems(checkupsResult?.data ?? []);
  }, [checkupsResult?.data, pet?.id]);

  const handleBack = useCallback(() => {
    navigate(paths.checkups.getHref());
  }, [navigate]);

  const guardedFormAction = useCallback(
    (formData: FormData) => {
      if (!canSubmit) return;
      formAction(formData);
    },
    [canSubmit, formAction],
  );

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
            submitLabel={isPetDeceased ? undefined : isPending ? "保存中..." : "保存"}
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
              species: pet.species,
              birthDate: pet.birthDate,
              gender: pet.gender,
              neuteredDate: pet.neuteredDate,
            })}
            insuranceName={pet.insuranceName}
            insuranceDetails={pet.insuranceDetails}
            staffName={doctorName}
            nextVisitDate={form.nextDate ? formatDate(form.nextDate) : undefined}
            status={isPetDeceased ? "deceased" : "alive"}
          />
        ) : null}
        {isPetDeceased ? (
          <div
            role="status"
            aria-label="死亡ペットのため保存不可"
            className={`flex items-center gap-2 px-4 py-2.5 rounded-md border mt-4 ${C.bgWarning50} ${C.borderWarning20} ${C.textWarning}`}
          >
            <span className="text-sm font-medium">
              死亡したペットの定期健診記録は保存できません
            </span>
          </div>
        ) : null}

        <div className="mt-4 grid grid-cols-1 gap-6 lg:grid-cols-5">
          <CheckupFieldsPanel
            canSubmit={canSubmit === true}
            form={form}
            fieldErrors={fieldErrors}
            checkupTypes={checkupTypes}
            staffs={staffs}
            checkupFields={checkupFields}
            fieldValues={fieldValues}
            onDateChange={setDate}
            onCheckupTypeIdChange={setCheckupTypeId}
            onNextScheduleTypeChange={setNextScheduleType}
            onNextDateChange={setNextDate}
            onDoctorIdChange={setDoctorId}
            onResultChange={setResult}
            onFieldValueChange={setFieldValue}
          />
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
