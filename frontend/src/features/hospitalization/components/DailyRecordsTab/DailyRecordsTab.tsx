// React/Framework
import { memo, useState, useCallback, useLayoutEffect, useRef, useTransition } from "react";

// External
import { Loader2, PlusCircle } from "lucide-react";

// Internal
import { C, ICON } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { EmptyState } from "@/components/shared/DataStates";
import { todayJSTISO } from "@/lib/jst-date";
import { useAuth } from "@/hooks/use-auth";
import { usePermission } from "@/hooks/use-permission";

// Relative
import { DailyDateNav } from "../../components/DailyRecordsTab/DailyDateNav";
import { DailyVitalsSection } from "../../components/DailyRecordsTab/DailyVitalsSection";
import { DailyCareLogsSection } from "../../components/DailyRecordsTab/DailyCareLogsSection";
import { DailyStaffNotesSection } from "../../components/DailyRecordsTab/DailyStaffNotesSection";
import { useGetDailyRecord, useCreateDailyRecord, useCreateDailyVital, useCreateCareLog, useCreateStaffNote } from "../../api/daily-records";
import { HOSPITALIZATION_DECEASED_BLOCK_MESSAGE } from "../../constants";

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

    // rerender-dependencies: useMutation の戻り値オブジェクト全体でなく、安定参照の関数のみを deps に置く。
    const { mutateAsync: createDailyRecordAsync } = useCreateDailyRecord(hospitalizationId);
    const { mutateAsync: createVitalAsync } = useCreateDailyVital(hospitalizationId, selectedDate);
    const { mutateAsync: createCareLogAsync } = useCreateCareLog(hospitalizationId, selectedDate);
    const { mutateAsync: createStaffNoteAsync } = useCreateStaffNote(hospitalizationId, selectedDate);

    const [isCreateRecordPending, startCreateRecordTransition] = useTransition();
    const [isCareLogPending, startCareLogTransition] = useTransition();
    const [isStaffNotePending, startStaffNoteTransition] = useTransition();

    const handleCreateDailyRecord = useCallback(() => {
        if (!isMutationAllowed()) return;
        startCreateRecordTransition(async () => {
            try {
                await createDailyRecordAsync(selectedDate);
            } catch {
                // useCreateDailyRecord.onError → handleApiError 済み
            }
        });
    }, [createDailyRecordAsync, isMutationAllowed, selectedDate]);

    // FE-RC-022: DailyVitalsSection は <form action> + SubmitButton で自身の pending を管理するため、
    // ここでは useTransition を介さず mutateAsync を直接 await する。
    const handleAddVital = useCallback(
        async (payload: CreateVitalRecordRequest) => {
            if (!isMutationAllowed()) return;
            try {
                await createVitalAsync({ ...payload, staff_id: currentUserId });
            } catch {
                // useCreateDailyVital.onError → handleApiError 済み
            }
        },
        [createVitalAsync, currentUserId, isMutationAllowed]
    );

    const handleAddCareLog = useCallback(
        (payload: CreateCareLogRequest) => {
            if (!isMutationAllowed()) return;
            startCareLogTransition(async () => {
                try {
                    await createCareLogAsync({ ...payload, staff_id: currentUserId });
                } catch {
                    // useCreateCareLog.onError → handleApiError 済み
                }
            });
        },
        [createCareLogAsync, currentUserId, isMutationAllowed]
    );

    const handleAddStaffNote = useCallback(
        (payload: CreateStaffNoteRequest) => {
            if (!isMutationAllowed()) return;
            startStaffNoteTransition(async () => {
                try {
                    await createStaffNoteAsync({ ...payload, staff_id: currentUserId });
                } catch {
                    // useCreateStaffNote.onError → handleApiError 済み
                }
            });
        },
        [createStaffNoteAsync, currentUserId, isMutationAllowed]
    );

    const vitals = record?.vital_records ?? [];
    const careLogs = record?.care_logs ?? [];
    const staffNotes = record?.staff_notes ?? [];
    // 臨床安全境界1: 死亡ペットは render 側でも追加操作を出さない（callback 側は isMutationAllowed で維持）。
    const canCreateNow = canCreate && !petIsDeceased;
    const showDeceasedBlockNotice = canCreate && petIsDeceased;

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
                    {canCreateNow ? (
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
                    ) : showDeceasedBlockNotice ? (
                        <p role="status" className={`text-xs ${C.text50}`}>
                            {HOSPITALIZATION_DECEASED_BLOCK_MESSAGE.DAILY_RECORD}
                        </p>
                    ) : null}
                </EmptyState>
            ) : (
                <div className="space-y-4">
                    {showDeceasedBlockNotice ? (
                        <p role="status" className={`text-xs ${C.text50}`}>
                            {HOSPITALIZATION_DECEASED_BLOCK_MESSAGE.DAILY_RECORD}
                        </p>
                    ) : null}

                    <DailyVitalsSection
                        vitals={vitals}
                        onAddVital={handleAddVital}
                        canCreate={canCreateNow}
                    />

                    <Separator className="opacity-50" />

                    <DailyCareLogsSection
                        careLogs={careLogs}
                        onAddCareLog={handleAddCareLog}
                        isPending={isCareLogPending}
                        canCreate={canCreateNow}
                    />

                    <Separator className="opacity-50" />

                    <DailyStaffNotesSection
                        staffNotes={staffNotes}
                        onAddStaffNote={handleAddStaffNote}
                        isPending={isStaffNotePending}
                        canCreate={canCreateNow}
                    />
                </div>
            )}
        </div>
    );
});
