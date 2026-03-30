// React/Framework
import { memo, useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router";

// External
import { Trash2 } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
import { Button } from "@/components/ui/button";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker/NotionDatePicker";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

// Relative
import { useVaccinationForm } from "../hooks/use-vaccination-form";
import { usePermission } from "@/features/auth/hooks/use-permission";

export const VaccinationForm = memo(function VaccinationForm() {
  const navigate = useNavigate();
  const { id } = useParams();
  const { canEdit, canDelete } = usePermission("vaccinations");
  
  const {
      isEdit,
      petSelection,
      form,
      formAction,
      formState,
      isSaving,
      handleDelete,
      isDeleting
  } = useVaccinationForm(id);

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  // React 19 Action の成功を検知して遷移
  useEffect(() => {
    if (formState.success) {
      markClean();
      navigate(paths.vaccinations.getHref());
    }
  }, [formState.success, formState.timestamp, navigate, markClean]);

  const { selectedPets } = petSelection;
  const selectedPet = selectedPets[0];

  const {
    date, setDate,
    vaccineId, setVaccineId,
    nextScheduleType, setNextScheduleType,
    nextDate, setNextDate,
    lot1, setLot1,
    remarks, setRemarks,
  } = form;

  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

  const handleBack = useCallback(() => {
    navigate(paths.vaccinations.getHref());
  }, [navigate]);

  if (!selectedPet && !isEdit) return null;

  return (
    <form action={formAction}>
    <PageLayout
      title={isEdit ? "予防接種詳細・編集" : "新規予防接種登録"}
      onBack={handleBack}
      maxWidth="max-w-[800px]"
      headerAction={
        <div className="flex gap-2">
            {canDelete && isEdit ? (
                <Button
                    variant="ghost"
                    type="button"
                    className={`${STYLE.btnDangerGhost} px-4 h-10 text-sm`}
                    onClick={() => setDeleteConfirmOpen(true)}
                    disabled={isDeleting}
                >
                    <Trash2 className={`mr-1.5 ${ICON.action}`} />
                    {isDeleting ? "削除中..." : "削除"}
                </Button>
            ) : null}
            {canEdit ? (
              <SubmitButton
                  className={`${C.bgAccent} ${C.bgAccentHover} text-white shadow-sm px-6 h-10 text-sm`}
              >
                  保存
              </SubmitButton>
            ) : null}
        </div>
      }
    >
        <NavigationBlocker when={isDirty && !isSaving} />
        <div className="flex flex-col gap-6">
            {selectedPet ? (
                <PatientInfoCard
                    ownerName={selectedPet.ownerName}
                    petName={selectedPet.name}
                    petNumber={selectedPet.petNumber}
                    weight={selectedPet.weight}
                />
            ) : null}

            <div className="bg-white p-6 rounded-lg border border-[rgba(55,53,47,0.09)] space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                        <Label>接種日</Label>
                        <NotionDatePicker
                            value={date}
                            onChange={(v) => { markDirty(); setDate(v); }}
                        />
                    </div>
                    <div className="space-y-2">
                        <Label>ワクチン</Label>
                        <Select value={vaccineId} onValueChange={(v) => { markDirty(); setVaccineId(v); }}>
                            <SelectTrigger>
                                <SelectValue placeholder="選択してください" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="1">混合ワクチン</SelectItem>
                                <SelectItem value="2">狂犬病ワクチン</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                        <Label>ロット番号</Label>
                        <Input
                            value={lot1}
                            onChange={(e) => { markDirty(); setLot1(e.target.value); }}
                            placeholder="ロット番号を入力"
                        />
                    </div>
                    <div className="space-y-2">
                        <Label>次回の予定</Label>
                        <div className="flex gap-4 items-center h-10">
                            <Select value={nextScheduleType} onValueChange={(v) => { markDirty(); setNextScheduleType(v); }}>
                                <SelectTrigger className="w-[120px]">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="1year">1年後</SelectItem>
                                    <SelectItem value="custom">指定日</SelectItem>
                                </SelectContent>
                            </Select>
                            <NotionDatePicker
                                value={nextDate}
                                onChange={(v) => { markDirty(); setNextDate(v); }}
                                className="flex-1"
                            />
                        </div>
                    </div>
                </div>

                <div className="space-y-2">
                    <Label>備考</Label>
                    <Textarea
                        value={remarks}
                        onChange={(e) => { markDirty(); setRemarks(e.target.value); }}
                        placeholder="備考を入力"
                        className="min-h-[100px]"
                    />
                </div>
            </div>
        </div>

        <ConfirmDialog
          open={deleteConfirmOpen}
          onClose={() => setDeleteConfirmOpen(false)}
          title="削除確認"
          description="この予防接種情報を削除してもよろしいですか？"
          confirmLabel="削除"
          variant="destructive"
          onConfirm={() => {
            handleDelete(() => {
              markClean();
              navigate(paths.vaccinations.getHref());
            });
          }}
        />
    </PageLayout>
    </form>
  );
});
