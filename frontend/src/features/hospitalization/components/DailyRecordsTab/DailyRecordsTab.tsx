// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { memo, useState, useCallback, useMemo, useTransition } from "react";

// External
import { Loader2, PlusCircle } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { handleApiError } from "@/lib/handle-api-error";

// Relative
import { DailyDateNav } from "@/features/hospitalization/components/DailyRecordsTab/DailyDateNav";
import { DailyVitalsSection } from "@/features/hospitalization/components/DailyRecordsTab/DailyVitalsSection";
import { DailyCareLogsSection } from "@/features/hospitalization/components/DailyRecordsTab/DailyCareLogsSection";
import { DailyStaffNotesSection } from "@/features/hospitalization/components/DailyRecordsTab/DailyStaffNotesSection";
import { useDailyRecord, useCreateDailyRecord, useCreateDailyVital, useCreateCareLog, useCreateStaffNote } from "@/features/hospitalization/api/daily-records";
import { usePermission } from "@/features/auth";

// Types
import type { CreateVitalRecordRequest, CreateCareLogRequest, CreateStaffNoteRequest } from "@/features/hospitalization/api/daily-records-types";

interface DailyRecordsTabProps {
    hospitalizationId: string;
    admissionDate: string; // YYYY-MM-DD
    dischargeDate: string; // YYYY-MM-DD (today if not discharged)
}

function getTodayStr(): string {
    return new Date().toISOString().split("T")[0];
}

function clampDate(date: string, min: string, max: string): string {
    if (date < min) return min;
    if (date > max) return max;
    return date;
}

export const DailyRecordsTab = memo(function DailyRecordsTab({
    hospitalizationId,
    admissionDate,
    dischargeDate,
}: DailyRecordsTabProps) {
    const { canCreate } = usePermission("hospitalization");
    // rerender-simple-expression-in-memo: string primitive は値比較のため useMemo 不要
    const today = getTodayStr();
    const effectiveMax = useMemo(
        () => (dischargeDate && dischargeDate < today ? dischargeDate : today),
        [dischargeDate, today]
    );
    const initialDate = useMemo(
        () => clampDate(today, admissionDate, effectiveMax),
        [today, admissionDate, effectiveMax]
    );

    const [selectedDate, setSelectedDate] = useState(initialDate);

    const handleDateChange = useCallback(
        (date: string) => {
            setSelectedDate(clampDate(date, admissionDate, effectiveMax));
        },
        [admissionDate, effectiveMax]
    );

    const { data: record, isLoading, isError } = useDailyRecord(hospitalizationId, selectedDate);

    const createDailyRecord = useCreateDailyRecord(hospitalizationId);
    const createVital = useCreateDailyVital(hospitalizationId, selectedDate);
    const createCareLog = useCreateCareLog(hospitalizationId, selectedDate);
    const createStaffNote = useCreateStaffNote(hospitalizationId, selectedDate);

    const [isCreateRecordPending, startCreateRecordTransition] = useTransition();
    const [isVitalPending, startVitalTransition] = useTransition();
    const [isCareLogPending, startCareLogTransition] = useTransition();
    const [isStaffNotePending, startStaffNoteTransition] = useTransition();

    const handleCreateDailyRecord = useCallback(() => {
        startCreateRecordTransition(async () => {
            try {
                await createDailyRecord.mutateAsync(selectedDate);
            } catch (error) {
                handleApiError(error, "日次記録の作成");
            }
        });
    }, [createDailyRecord, selectedDate]);

    const handleAddVital = useCallback(
        (payload: CreateVitalRecordRequest) => {
            startVitalTransition(async () => {
                try {
                    await createVital.mutateAsync(payload);
                } catch (error) {
                    handleApiError(error, "バイタル追加");
                }
            });
        },
        [createVital]
    );

    const handleAddCareLog = useCallback(
        (payload: CreateCareLogRequest) => {
            startCareLogTransition(async () => {
                try {
                    await createCareLog.mutateAsync(payload);
                } catch (error) {
                    handleApiError(error, "ケアログ追加");
                }
            });
        },
        [createCareLog]
    );

    const handleAddStaffNote = useCallback(
        (payload: CreateStaffNoteRequest) => {
            startStaffNoteTransition(async () => {
                try {
                    await createStaffNote.mutateAsync(payload);
                } catch (error) {
                    handleApiError(error, "スタッフメモ追加");
                }
            });
        },
        [createStaffNote]
    );

    const vitals = record?.vital_records ?? [];
    const careLogs = record?.care_logs ?? [];
    const staffNotes = record?.staff_notes ?? [];

    return (
        <div className="flex flex-col gap-4">
            <DailyDateNav
                selectedDate={selectedDate}
                admissionDate={admissionDate}
                dischargeDate={effectiveMax}
                onDateChange={handleDateChange}
            />

            {isLoading ? (
                <div className={`flex items-center justify-center py-10 ${C.text40}`}>
                    <Loader2 className={`${ICON.page} animate-spin mr-2`} />
                    <span className="text-sm">読み込み中...</span>
                </div>
            ) : isError ? (
                <div className="flex flex-col items-center justify-center py-10 gap-3">
                    <p className={`text-sm ${C.text50}`}>この日の記録はまだありません</p>
                    {canCreate ? (
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={handleCreateDailyRecord}
                            disabled={isCreateRecordPending}
                            className="gap-1.5"
                        >
                            {isCreateRecordPending ? (
                                <Loader2 className={`${ICON.action} animate-spin`} />
                            ) : (
                                <PlusCircle className={ICON.action} />
                            )}
                            この日の記録を作成
                        </Button>
                    ) : null}
                </div>
            ) : (
                <div className="space-y-4">
                    <DailyVitalsSection
                        vitals={vitals}
                        onAddVital={handleAddVital}
                        isPending={isVitalPending}
                        canCreate={canCreate}
                    />

                    <Separator className="opacity-50" />

                    <DailyCareLogsSection
                        careLogs={careLogs}
                        onAddCareLog={handleAddCareLog}
                        isPending={isCareLogPending}
                        canCreate={canCreate}
                    />

                    <Separator className="opacity-50" />

                    <DailyStaffNotesSection
                        staffNotes={staffNotes}
                        onAddStaffNote={handleAddStaffNote}
                        isPending={isStaffNotePending}
                        canCreate={canCreate}
                    />
                </div>
            )}
        </div>
    );
});
