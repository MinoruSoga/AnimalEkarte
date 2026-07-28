// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { memo, useState, useCallback, useLayoutEffect, useRef, useTransition } from "react";

// External
import { Loader2, PlusCircle } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { EmptyState } from "@/components/shared/DataStates";
import { handleApiError } from "@/lib/handle-api-error";
import { todayJSTISO } from "@/lib/jst-date";

// Relative
import { DailyDateNav } from "../../components/DailyRecordsTab/DailyDateNav";
import { DailyVitalsSection } from "../../components/DailyRecordsTab/DailyVitalsSection";
import { DailyCareLogsSection } from "../../components/DailyRecordsTab/DailyCareLogsSection";
import { DailyStaffNotesSection } from "../../components/DailyRecordsTab/DailyStaffNotesSection";
import { useGetDailyRecord, useCreateDailyRecord, useCreateDailyVital, useCreateCareLog, useCreateStaffNote } from "../../api/daily-records";
import { useAuth } from "@/hooks/use-auth";
import { usePermission } from "@/hooks/use-permission";

// Types
import type { CreateVitalRecordRequest, CreateCareLogRequest, CreateStaffNoteRequest } from "../../api/daily-records-types";

interface DailyRecordsTabProps {
    hospitalizationId: string;
    admissionDate: string; // YYYY-MM-DD
    dischargeDate: string; // YYYY-MM-DD (today if not discharged)
    petIsDeceased: boolean;
}

function getTodayStr(): string {
    return todayJSTISO();
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
    petIsDeceased,
}: DailyRecordsTabProps) {
    const { canCreate } = usePermission("hospitalization");
    const canCreateRef = useRef(canCreate);
    const petIsDeceasedRef = useRef(petIsDeceased);
    useLayoutEffect(() => {
        canCreateRef.current = canCreate;
        petIsDeceasedRef.current = petIsDeceased;
    }, [canCreate, petIsDeceased]);
    const isMutationAllowed = useCallback(
        () => canCreateRef.current === true && petIsDeceasedRef.current !== true,
        [],
    );
    const { user } = useAuth();
    const currentUserId = Number(user?.id ?? 0);
    // rerender-simple-expression-in-memo: string primitive は値比較のため useMemo 不要
    const today = getTodayStr();
    const effectiveMax = dischargeDate && dischargeDate < today ? dischargeDate : today;
    const initialDate = clampDate(today, admissionDate, effectiveMax);

    const [selectedDate, setSelectedDate] = useState(initialDate);

    const handleDateChange = useCallback(
        (date: string) => {
            setSelectedDate(clampDate(date, admissionDate, effectiveMax));
        },
        [admissionDate, effectiveMax]
    );

    const { data: record, isLoading, isError } = useGetDailyRecord(hospitalizationId, selectedDate);

    const createDailyRecord = useCreateDailyRecord(hospitalizationId);
    const createVital = useCreateDailyVital(hospitalizationId, selectedDate);
    const createCareLog = useCreateCareLog(hospitalizationId, selectedDate);
    const createStaffNote = useCreateStaffNote(hospitalizationId, selectedDate);

    const [isCreateRecordPending, startCreateRecordTransition] = useTransition();
    const [isVitalPending, startVitalTransition] = useTransition();
    const [isCareLogPending, startCareLogTransition] = useTransition();
    const [isStaffNotePending, startStaffNoteTransition] = useTransition();

    const handleCreateDailyRecord = useCallback(() => {
        if (!isMutationAllowed()) return;
        startCreateRecordTransition(async () => {
            try {
                await createDailyRecord.mutateAsync(selectedDate);
            } catch (error) {
                handleApiError(error, "日次記録の作成");
            }
        });
    }, [createDailyRecord, isMutationAllowed, selectedDate]);

    const handleAddVital = useCallback(
        (payload: CreateVitalRecordRequest) => {
            if (!isMutationAllowed()) return;
            startVitalTransition(async () => {
                try {
                    await createVital.mutateAsync({ ...payload, staff_id: currentUserId });
                } catch (error) {
                    handleApiError(error, "バイタル追加");
                }
            });
        },
        [createVital, currentUserId, isMutationAllowed]
    );

    const handleAddCareLog = useCallback(
        (payload: CreateCareLogRequest) => {
            if (!isMutationAllowed()) return;
            startCareLogTransition(async () => {
                try {
                    await createCareLog.mutateAsync({ ...payload, staff_id: currentUserId });
                } catch (error) {
                    handleApiError(error, "ケアログ追加");
                }
            });
        },
        [createCareLog, currentUserId, isMutationAllowed]
    );

    const handleAddStaffNote = useCallback(
        (payload: CreateStaffNoteRequest) => {
            if (!isMutationAllowed()) return;
            startStaffNoteTransition(async () => {
                try {
                    await createStaffNote.mutateAsync({ ...payload, staff_id: currentUserId });
                } catch (error) {
                    handleApiError(error, "スタッフメモ追加");
                }
            });
        },
        [createStaffNote, currentUserId, isMutationAllowed]
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
                <EmptyState message="この日の記録はまだありません">
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
                </EmptyState>
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
