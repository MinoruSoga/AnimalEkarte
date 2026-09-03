import { useCallback, useMemo, useState } from "react";
import Axios from "axios";
import { toast } from "sonner";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { useAuth } from "@/hooks/use-auth";
import { usePermission } from "@/hooks/use-permission";
import { C, LAYOUT } from "@/lib/design-tokens";
import { todayJSTISO } from "@/lib/jst-date";
import { ResourceLabImport } from "@/types/generated/models";

import {
  parseLabDeviceSlots,
  useAttachLabDeviceJob,
  useClearLabDeviceWait,
  useDetachLabDeviceJob,
  useGetLabDeviceBoard,
  usePutLabDeviceWait,
  useReceiveLabDeviceFrames,
} from "../api/lab-device";
import { useLabDeviceAgentListen } from "../hooks/use-lab-device-agent-listen";
import {
  groupLabDeviceCardsByDay,
  labDeviceAgentConnectionLabel,
  labDeviceReceiveFailure,
  requireLabDeviceReceiveResult,
  labDeviceReceivedCards,
  labDeviceSelectableTodayVisits,
} from "../lib/lab-device-board-model";
import {
  LabDeviceAgentPanel,
  LabDeviceReceivedPanel,
  LabDeviceTodayVisitsPanel,
  LabDeviceUnlinkedPanel,
  LabDeviceWaitAndSlotsPanel,
} from "./lab-device-board-panels";

const LAB_DEVICE_AGENT_RECEIVE_TOAST_ID = "lab-device-agent-receive";

export function LabDeviceBoard() {
  const { currentClinicId } = useAuth();
  const { canCreate, canEdit } = usePermission(ResourceLabImport);
  const { data: board, isLoading } = useGetLabDeviceBoard(canCreate);
  const putWait = usePutLabDeviceWait();
  const clearWait = useClearLabDeviceWait();
  const receive = useReceiveLabDeviceFrames();
  const attach = useAttachLabDeviceJob();
  const detach = useDetachLabDeviceJob();
  const [lastReceives, setLastReceives] = useState<Record<string, {
    clinicId: string;
    label: string;
    at: string;
  }>>({});
  const todayVisits = useMemo(
    () => labDeviceSelectableTodayVisits(board?.todayVisits ?? []),
    [board?.todayVisits],
  );
  const receivedCards = useMemo(
    () => labDeviceReceivedCards({
      received: board?.received ?? [],
      unlinked: board?.unlinked ?? [],
      saved: board?.saved ?? [],
    }),
    [board?.received, board?.saved, board?.unlinked],
  );
  const receivedGroups = useMemo(
    () => groupLabDeviceCardsByDay(receivedCards),
    [receivedCards],
  );
  const today = todayJSTISO();
  const slots = useMemo(() => parseLabDeviceSlots(board?.station.slotsJson ?? "[]"), [board?.station.slotsJson]);
  const onFrame = useCallback(async (frame: { payloadBase64: string; deviceHint: "auto" }) => {
    try {
      const results = await receive.mutateAsync(frame);
      toast.dismiss(LAB_DEVICE_AGENT_RECEIVE_TOAST_ID);
      const first = requireLabDeviceReceiveResult(results);
      if (currentClinicId === null) {
        throw new Error("clinic is not selected");
      }
      const slot = slots.find((candidate) => (
        candidate.deviceHint === first.job.deviceHint || candidate.sourceType === first.job.sourceType
      ));
      if (slot) {
        setLastReceives((current) => ({
          ...current,
          [slot.key]: {
            clinicId: currentClinicId,
            label: first.duplicate ? "再送（取込済み）" : "受信",
            at: new Date().toISOString(),
          },
        }));
      }
    } catch (error) {
      const status = Axios.isAxiosError(error) ? error.response?.status : undefined;
      const failure = labDeviceReceiveFailure(status);
      if (status === 400) {
        // 400 は mutation 側の onError（handleApiError）が通知済み。ここでは再通知しない（FE-RC-005）。
        toast.dismiss(LAB_DEVICE_AGENT_RECEIVE_TOAST_ID);
      } else {
        // mutation 側 onError の汎用トーストを消し、機器の再送要否を含む具体的な案内に差し替える。
        toast.dismiss();
        toast.error(failure.message, { id: LAB_DEVICE_AGENT_RECEIVE_TOAST_ID });
      }
      throw Object.assign(new Error("lab device receive failed", { cause: error }), { status });
    }
  }, [currentClinicId, receive, slots]);
  const agentStatus = useLabDeviceAgentListen({
    enabled: canCreate && currentClinicId !== null,
    clinicId: currentClinicId,
    onFrame,
  });
  const linkLabel = labDeviceAgentConnectionLabel(agentStatus);

  return (
    <PageLayout
      title="検査受信"
      resource={ResourceLabImport}
      maxWidth={LAYOUT.pageContentMaxWidth.form}
      align="left"
    >
      <div className="space-y-6">
        {!canCreate ? (
          <p className={`text-sm ${C.textWarning}`}>
            受信権限（lab-import の作成）が無いため、このページでは受信できません。権限グループ設定で付与してください
          </p>
        ) : null}
        <section className="space-y-3">
          <LabDeviceWaitAndSlotsPanel
            wait={board?.wait}
            linkLabel={linkLabel}
            canCreate={canCreate}
            onClearWait={() => clearWait.mutate()}
            slots={slots}
            receivedCards={receivedCards}
            lastReceives={lastReceives}
            currentClinicId={currentClinicId}
            agentStatus={agentStatus}
          />
          <LabDeviceTodayVisitsPanel
            isLoading={isLoading}
            todayVisits={todayVisits}
            waitPetId={board?.wait?.petId}
            canCreate={canCreate}
            onSelectVisit={(petId) => putWait.mutate(petId)}
          />
        </section>
        <LabDeviceUnlinkedPanel
          unlinked={board?.unlinked ?? []}
          wait={board?.wait}
          canEdit={canEdit}
          onAttach={(jobId, petId) => attach.mutateAsync({ jobId, petId })}
        />
        <LabDeviceReceivedPanel
          receivedGroups={receivedGroups}
          today={today}
          canEdit={canEdit}
          wait={board?.wait}
          onDetach={(jobId) => detach.mutate(jobId)}
          onAttach={(jobId, petId) => attach.mutateAsync({ jobId, petId })}
        />
        <LabDeviceAgentPanel agentStatus={agentStatus} />
      </div>
    </PageLayout>
  );
}
