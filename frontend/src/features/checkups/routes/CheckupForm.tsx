// React/Framework
import { useCallback } from "react";
import { useNavigate } from "react-router";

// External
import { ClipboardCheck } from "lucide-react";

// Internal
import { C, ICON } from "@/lib/design-tokens";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker/NotionDatePicker";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { LoadingFallback } from "@/components/shared/DataStates";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { paths } from "@/config/paths";
import { useGetAllCheckupTypes } from "@/hooks/use-treatment-master";
import { useGetStaffs } from "@/hooks/use-staffs";
import { ResourceCheckups } from "@/types/generated/models";

// Relative
import { useCheckupForm } from "../hooks/use-checkup-form";

export function CheckupForm() {
  const navigate = useNavigate();

  const {
    pet,
    isPetLoading,
    form,
    formAction,
    isPending,
    fieldErrors,
    setCheckupTypeId,
    setDate,
    setNextDate,
    setDoctorId,
    setResult,
  } = useCheckupForm();

  const { data: checkupTypes = [] } = useGetAllCheckupTypes();
  const { data: staffs = [] } = useGetStaffs();

  const handleBack = useCallback(() => {
    navigate(paths.checkups.getHref());
  }, [navigate]);

  if (isPetLoading) return <LoadingFallback />;

  return (
    <form action={formAction}>
      <PageLayout
        title="定期健診登録"
        resource={ResourceCheckups}
        icon={<ClipboardCheck className={`${ICON.page} ${C.text}`} />}
        onBack={handleBack}
        maxWidth="max-w-[900px]"
        headerAction={
          <SubmitButton
            className={`${C.bgAccent} ${C.bgAccentHover} ${C.textWhite} shadow-sm px-6 h-10 text-sm`}
            disabled={isPending}
          >
            {isPending ? "保存中..." : "保存"}
          </SubmitButton>
        }
      >
        {pet ? (
          <PatientInfoCard
            ownerName={pet.ownerName ?? ""}
            petName={pet.name}
            petNumber={pet.petNumber ?? ""}
            weight={pet.weight ?? ""}
          />
        ) : null}

        <div className={`${C.bgWhite} p-6 rounded-lg border ${C.borderLight} space-y-6 mt-4`}>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* 実施日 */}
            <div className="space-y-2">
              <Label htmlFor="checkup-date">
                実施日<span className={`${C.textRequired} ml-1`}>*</span>
              </Label>
              <NotionDatePicker
                id="checkup-date"
                value={form.date}
                onChange={setDate}
                disabledDays={{ after: new Date() }}
              />
              <FormFieldError message={fieldErrors.date} />
            </div>

            {/* 健診種別 */}
            <div className="space-y-2">
              <Label htmlFor="checkup-type-select">
                健診種別<span className={`${C.textRequired} ml-1`}>*</span>
              </Label>
              <Select value={form.checkupTypeId} onValueChange={setCheckupTypeId}>
                <SelectTrigger id="checkup-type-select">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  {checkupTypes.map((ct) => (
                    <SelectItem key={ct.id} value={String(ct.id)}>
                      {ct.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormFieldError message={fieldErrors.checkupTypeId} />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* 次回予定日 */}
            <div className="space-y-2">
              <Label htmlFor="checkup-next-date">次回予定日</Label>
              <NotionDatePicker
                id="checkup-next-date"
                value={form.nextDate}
                onChange={setNextDate}
              />
            </div>

            {/* 担当医 */}
            <div className="space-y-2">
              <Label htmlFor="checkup-doctor-select">担当医</Label>
              <Select value={form.doctorId} onValueChange={setDoctorId}>
                <SelectTrigger id="checkup-doctor-select">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  {staffs
                    .filter((s) => s.isActive)
                    .map((s) => (
                      <SelectItem key={s.id} value={s.id}>
                        {s.name}
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
            </div>
          </div>

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
        </div>
      </PageLayout>
    </form>
  );
}
