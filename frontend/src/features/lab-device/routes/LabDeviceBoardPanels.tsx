import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";

import type {
  LabDeviceJobCard,
  LabDeviceSlot,
  LabDeviceTodayVisit,
  LabDeviceWait,
} from "../api/lab-device";
import {
  LabDeviceJobCardView,
  LabDeviceStatusCard,
  LabDeviceTodayVisitCard,
} from "../components/LabDeviceBoardCards";
import type { LabDeviceBrowserReceiveStatus } from "../lib/lab-device-browser-receiver";
import {
  isLabDeviceAttachPersisted,
  isLabDeviceBoardSlotSupported,
  labDeviceAgentConnectionLabel,
  labDeviceAttachFailureToast,
  labDeviceBoardSlotListenState,
  labDeviceLatestCardForSlot,
  labDeviceLiveReceiveLabel,
  labDeviceReceivedDayLabel,
  resolveLabDeviceReceiveTime,
} from "../lib/lab-device-board-model";

function notifyLabDeviceAttachResult(attached: LabDeviceJobCard): void {
  if (isLabDeviceAttachPersisted(attached)) return;
  toast.error(labDeviceAttachFailureToast(attached));
}

export function LabDeviceWaitAndSlotsPanel({
  wait,
  linkLabel,
  canCreate,
  onClearWait,
  slots,
  receivedCards,
  lastReceives,
  currentClinicId,
  agentStatus,
}: {
  wait: LabDeviceWait | null | undefined;
  linkLabel: ReturnType<typeof labDeviceAgentConnectionLabel>;
  canCreate: boolean;
  onClearWait: () => void;
  slots: LabDeviceSlot[];
  receivedCards: LabDeviceJobCard[];
  lastReceives: Record<string, { clinicId: string; label: string; at: string }>;
  currentClinicId: string | null;
  agentStatus: { connected: boolean; openPorts: number; configuredPorts: number };
}) {
  return (
    <>
      {wait ? (
        <div className="space-y-3">
          <p className="text-heading-1 font-bold">{wait.petName}</p>
          <p className={`text-sm ${C.textInkMuted}`}>待機中 · 接続 {linkLabel}</p>
          {canCreate ? (
            <Button type="button" variant="outline" onClick={onClearWait}>
              待機を解除
            </Button>
          ) : null}
        </div>
      ) : (
        <div className="space-y-3">
          <p className="text-heading-1 font-semibold">受信中</p>
          <p className={`text-sm ${C.textInkMuted}`}>ペット未選択 · 接続 {linkLabel}</p>
        </div>
      )}
      <div className="space-y-2">
        <h2 className="text-xl font-semibold">検査機器</h2>
        <ul className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {slots.map((slot) => {
            const supported = isLabDeviceBoardSlotSupported(slot.sourceType);
            const latestCard = labDeviceLatestCardForSlot(slot, receivedCards);
            const liveCandidate = lastReceives[slot.key];
            const live = liveCandidate?.clinicId === currentClinicId ? liveCandidate : undefined;
            const state = labDeviceBoardSlotListenState({
              supported,
              agentConnected: agentStatus.connected,
              openPorts: agentStatus.openPorts,
              configuredPorts: agentStatus.configuredPorts,
              hasLiveReceive: Boolean(live),
            });
            return (
              <li key={slot.key}>
                <LabDeviceStatusCard
                  slot={slot}
                  state={state}
                  receiveLabel={labDeviceLiveReceiveLabel({
                    liveLabel: live?.label,
                    latestCard,
                  })}
                  receiveTime={resolveLabDeviceReceiveTime(live?.at, latestCard)}
                  latestCard={latestCard}
                />
              </li>
            );
          })}
        </ul>
      </div>
    </>
  );
}

export function LabDeviceTodayVisitsPanel({
  isLoading,
  todayVisits,
  waitPetId,
  canCreate,
  onSelectVisit,
}: {
  isLoading: boolean;
  todayVisits: LabDeviceTodayVisit[];
  waitPetId: number | undefined;
  canCreate: boolean;
  onSelectVisit: (petId: number) => void;
}) {
  return (
    <div className="space-y-2">
      <h2 className="text-xl font-semibold">本日診療中のカルテ</h2>
      {isLoading ? <p>読み込み中</p> : null}
      {todayVisits.length === 0 && !isLoading ? (
        <p className={C.textInkMuted}>本日診療中のカルテはありません</p>
      ) : (
        <ul className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {todayVisits.map((visit) => (
            <li key={visit.recordId}>
              <LabDeviceTodayVisitCard
                visit={visit}
                selected={waitPetId === visit.petId}
                disabled={!canCreate}
                onSelect={() => onSelectVisit(visit.petId)}
              />
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

export function LabDeviceUnlinkedPanel({
  unlinked,
  wait,
  canEdit,
  onAttach,
}: {
  unlinked: LabDeviceJobCard[];
  wait: LabDeviceWait | null | undefined;
  canEdit: boolean;
  onAttach: (jobId: string, petId: number) => Promise<LabDeviceJobCard>;
}) {
  return (
    <section className="space-y-2">
      <h2 className="text-xl font-semibold">未紐付け</h2>
      {unlinked.length === 0 ? (
        <p className={C.textInkMuted}>未紐付けの受信はありません</p>
      ) : (
        unlinked.map((card) => (
          <LabDeviceJobCardView
            key={card.jobId}
            card={card}
            canEdit={canEdit}
            onAttach={
              canEdit && wait
                ? () => {
                    // FE-RC-012: reject 時は useAttachLabDeviceJob.onError が既に通知済み。
                    // ここでは unhandled rejection 化を防ぐためだけに no-op catch を付ける。
                    void onAttach(card.jobId, wait.petId)
                      .then(notifyLabDeviceAttachResult)
                      .catch(() => {});
                  }
                : undefined
            }
          />
        ))
      )}
    </section>
  );
}

export function LabDeviceReceivedPanel({
  receivedGroups,
  today,
  canEdit,
  wait,
  onDetach,
  onAttach,
}: {
  receivedGroups: { day: string; cards: LabDeviceJobCard[] }[];
  today: string;
  canEdit: boolean;
  wait: LabDeviceWait | null | undefined;
  onDetach: (jobId: string) => void;
  onAttach: (jobId: string, petId: number) => Promise<LabDeviceJobCard>;
}) {
  return (
    <section className="space-y-4">
      <h2 className="text-xl font-semibold">受信一覧</h2>
      {receivedGroups.length === 0 ? (
        <p className={C.textInkMuted}>受信した検査はありません</p>
      ) : (
        receivedGroups.map((group) => (
          <div key={group.day} className="space-y-2">
            <h3 className="text-heading-3 font-semibold">
              {labDeviceReceivedDayLabel(group.day, today)}
            </h3>
            {group.cards.map((card) => (
              <LabDeviceJobCardView
                key={card.jobId}
                card={card}
                canEdit={canEdit}
                onDetach={canEdit && card.petId ? () => onDetach(card.jobId) : undefined}
                onAttach={
                  canEdit && !card.petId && wait
                    ? () => {
                        // FE-RC-012: reject 時は useAttachLabDeviceJob.onError が既に通知済み。
                        // ここでは unhandled rejection 化を防ぐためだけに no-op catch を付ける。
                        void onAttach(card.jobId, wait.petId)
                          .then(notifyLabDeviceAttachResult)
                          .catch(() => {});
                      }
                    : undefined
                }
              />
            ))}
          </div>
        ))
      )}
    </section>
  );
}

export function LabDeviceBrowserReceivePanel({
  receiveStatus,
  canCreate,
  onConnect,
  onRetry,
  onStop,
}: {
  receiveStatus: LabDeviceBrowserReceiveStatus;
  canCreate: boolean;
  onConnect: () => void;
  onRetry: () => void;
  onStop: () => void;
}) {
  return (
    <section className="space-y-2" aria-live="polite">
      <h2 className="text-xl font-semibold">ブラウザ直接受信</h2>
      <p className={`text-sm ${C.textInkMuted}`}>
        {receiveStatus.connected ? `受信中 · ${receiveStatus.openPorts}ポート` : "停止中"}
        {` · 未送信 ${receiveStatus.pending}件 · 判定失敗 ${receiveStatus.rejected}件 · 超過 ${receiveStatus.overflow}件`}
      </p>
      {receiveStatus.supported ? null : (
        <p className={`text-sm ${C.textWarning}`}>
          Web SerialにはHTTPSのChromeデスクトップ版が必要です。
        </p>
      )}
      <p className={`text-sm ${C.textInkMuted}`}>
        NX600/AU10V専用ポートを選択してください。PU-4010/IDEXXは選ばないでください。ページを開いている間だけ受信します。停止・ページを閉じると未送信データは失われるため、機器から再送してください。
      </p>
      {receiveStatus.error ? (
        <p className={`text-sm ${C.textWarning}`}>{receiveStatus.error}</p>
      ) : null}
      <div className="flex flex-wrap gap-2">
        <Button
          type="button"
          variant="outline"
          disabled={!canCreate || !receiveStatus.supported}
          onClick={onConnect}
        >
          {receiveStatus.connected ? "ポートを追加" : "接続する"}
        </Button>
        {receiveStatus.connected ||
        receiveStatus.pending > 0 ||
        receiveStatus.rejected > 0 ||
        receiveStatus.authPaused ||
        receiveStatus.error !== null ? (
          <Button type="button" variant="outline" onClick={onStop}>
            停止
          </Button>
        ) : null}
        {receiveStatus.error || receiveStatus.authPaused ? (
          <Button type="button" variant="outline" disabled={!canCreate} onClick={onRetry}>
            再試行
          </Button>
        ) : null}
      </div>
    </section>
  );
}
