import { Link } from "react-router";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";
import { formatJSTTime } from "@/lib/jst-date";

import type {
  LabDeviceJobCard,
  LabDeviceSlot,
  LabDeviceTodayVisit,
  LabDeviceWait,
} from "../api/lab-device";
import type { LabDeviceAgentListenStatus } from "../hooks/use-lab-device-agent-listen";
import {
  isLabDeviceAttachPersisted,
  isLabDeviceBoardSlotSupported,
  labDeviceAgentConnectionLabel,
  labDeviceAgentDegradedErrorMessage,
  labDeviceAttachFailureToast,
  labDeviceBoardSlotListenState,
  labDeviceCardNeedsReview,
  labDeviceCardTitle,
  labDeviceClockSkewLabel,
  labDeviceHasUnmapped,
  labDeviceLatestCardForSlot,
  labDeviceListenTone,
  labDeviceLiveReceiveLabel,
  labDeviceNeedsReviewReason,
  labDeviceReceivedDayLabel,
  labDeviceSlotListenLabel,
  labDeviceUnmappedMasterHref,
  type LabDeviceListenState,
} from "../lib/lab-device-board-model";

function LabDeviceJobCardView({
  card,
  duplicate,
  onDetach,
  onAttach,
  canEdit,
}: {
  card: LabDeviceJobCard;
  duplicate?: boolean;
  onDetach?: () => void;
  onAttach?: () => void;
  canEdit: boolean;
}) {
  return (
    <section className={`rounded-lg border ${C.borderLight} ${C.bgWhite} p-4 space-y-2`}>
      <div className="flex items-baseline justify-between gap-3">
        <h3 className="text-heading-3 font-semibold">{labDeviceCardTitle(card)}</h3>
        <p className={`text-sm ${C.textInkMuted}`}>
          {card.petName || "未紐付け"}
          {card.receivedAt || card.measuredAt
            ? ` · ${formatJSTTime(card.measuredAt || card.receivedAt || "")}`
            : ""}
          {duplicate ? " · 再送（取込済み）" : ""}
        </p>
      </div>
      {labDeviceClockSkewLabel(card) ? (
        <p className={`text-sm ${C.textWarning}`}>{labDeviceClockSkewLabel(card)}</p>
      ) : null}
      <ul className="grid grid-cols-2 gap-x-4 gap-y-1 text-sm">
        {card.items.map((item) => (
          <li key={`${card.jobId}-${item.sortOrder}-${item.deviceItemCode}`}>
            <span className="font-medium">{item.deviceItemCode}</span>
            {" "}
            {item.valueRaw}
            {item.unit ? ` ${item.unit}` : ""}
            {item.needsReview ? (
              <>
                {" · "}
                <Link
                  to={labDeviceUnmappedMasterHref(card.sourceType)}
                  className="underline"
                >
                  未対応
                </Link>
              </>
            ) : null}
          </li>
        ))}
      </ul>
      {labDeviceCardNeedsReview(card) ? (
        <p className={`text-sm ${C.textRed700}`}>
          {labDeviceNeedsReviewReason(card)}
        </p>
      ) : null}
      {labDeviceHasUnmapped(card) ? (
        <p className={`text-sm ${C.textRed700}`}>
          <Link to={labDeviceUnmappedMasterHref(card.sourceType)} className="underline">
            未対応項目があります。該当機器のマスタを直してください
          </Link>
        </p>
      ) : null}
      <div className="flex gap-2">
        {onDetach && canEdit ? (
          <Button type="button" variant="outline" size="sm" onClick={onDetach}>取り消す</Button>
        ) : null}
        {onAttach && canEdit ? (
          <Button type="button" variant="outline" size="sm" onClick={onAttach}>この子に付ける</Button>
        ) : null}
      </div>
    </section>
  );
}

function LabDeviceTodayVisitCard({
  visit,
  selected,
  disabled,
  onSelect,
}: {
  visit: LabDeviceTodayVisit;
  selected: boolean;
  disabled: boolean;
  onSelect: () => void;
}) {
  return (
    <button
      type="button"
      aria-pressed={selected}
      disabled={disabled}
      onClick={onSelect}
      className={`min-h-11 w-full rounded-lg border p-4 text-left space-y-1 ${
        selected ? `${C.borderActionPrimary} ${C.bgActionPrimaryLight}` : `${C.borderLight} ${C.bgWhite}`
      }`}
    >
      <p className="text-heading-3 font-semibold">{visit.petName}</p>
      <p className={`text-sm ${C.textInkMuted}`}>
        {visit.ownerName}
        {visit.species ? ` · ${visit.species}` : ""}
        {visit.visitType ? ` · ${visit.visitType}` : ""}
      </p>
      {visit.doctorName ? (
        <p className={`text-sm ${C.textInkMuted}`}>{visit.doctorName}</p>
      ) : null}
      {selected ? <p className="text-sm font-medium">待機中</p> : null}
    </button>
  );
}

function deviceStatusClass(state: LabDeviceListenState): string {
  switch (labDeviceListenTone(state)) {
    case "live":
      return `${C.borderStatusGreen} ${C.bgStatusGreen} ${C.textStatusGreen}`;
    case "idle":
      return `${C.borderLight} ${C.bgStatusAmber} ${C.textStatusAmber}`;
    case "blocked":
      return `${C.borderWarning20} ${C.bgWarning50} ${C.textWarning}`;
    case "unsupported":
      return `${C.borderLight} ${C.bgWhite}`;
  }
}

function LabDeviceStatusCard({
  slot,
  state,
  receiveLabel,
  receiveTime,
  latestCard,
}: {
  slot: LabDeviceSlot;
  state: LabDeviceListenState;
  receiveLabel: string;
  receiveTime?: string;
  latestCard?: LabDeviceJobCard;
}) {
  return (
    <section
      aria-live="polite"
      className={`rounded-lg border p-4 space-y-2 ${deviceStatusClass(state)}`}
    >
      <div className="flex items-baseline justify-between gap-3">
        <h3 className="text-heading-3 font-semibold">{slot.deviceHint}</h3>
        <p className="text-sm font-medium">{labDeviceSlotListenLabel(state)}</p>
      </div>
      <p className={`text-sm ${C.textInkMuted}`}>
        最終受信 {receiveTime ? `${receiveTime} · ${receiveLabel}` : receiveLabel}
        {latestCard?.petName ? ` · ${latestCard.petName}` : ""}
      </p>
      {latestCard ? (
        <p className={`text-sm ${C.textInkMuted}`}>
          {latestCard.itemCount}項目
          {latestCard.unmappedItemCount > 0 ? " · 未対応あり" : ""}
        </p>
      ) : null}
    </section>
  );
}

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
  agentStatus: LabDeviceAgentListenStatus;
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
                  receiveTime={live?.at
                    ? formatJSTTime(live.at)
                    : latestCard?.measuredAt || latestCard?.receivedAt
                      ? formatJSTTime(latestCard.measuredAt || latestCard.receivedAt || "")
                      : undefined}
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
      ) : unlinked.map((card) => (
        <LabDeviceJobCardView
          key={card.jobId}
          card={card}
          canEdit={canEdit}
          onAttach={wait ? () => {
            void onAttach(card.jobId, wait.petId).then(notifyLabDeviceAttachResult);
          } : undefined}
        />
      ))}
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
      ) : receivedGroups.map((group) => (
        <div key={group.day} className="space-y-2">
          <h3 className="text-heading-3 font-semibold">
            {labDeviceReceivedDayLabel(group.day, today)}
          </h3>
          {group.cards.map((card) => (
            <LabDeviceJobCardView
              key={card.jobId}
              card={card}
              canEdit={canEdit}
              onDetach={card.petId ? () => onDetach(card.jobId) : undefined}
              onAttach={!card.petId && wait ? () => {
                void onAttach(card.jobId, wait.petId).then(notifyLabDeviceAttachResult);
              } : undefined}
            />
          ))}
        </div>
      ))}
    </section>
  );
}

export function LabDeviceAgentPanel({
  agentStatus,
}: {
  agentStatus: LabDeviceAgentListenStatus;
}) {
  const degradedMessage = agentStatus.degraded
    ? labDeviceAgentDegradedErrorMessage(agentStatus.lastErrorCategory)
    : null;
  return (
    <section className="space-y-2" aria-live="polite">
      <h2 className="text-xl font-semibold">ローカル受信機</h2>
      <p className={`text-sm ${C.textInkMuted}`}>
        {agentStatus.connected
          ? `稼働中 · ${agentStatus.openPorts}/${agentStatus.configuredPorts}ポート監視`
          : "停止中 · Macのローカル受信機を起動してください"}
      </p>
      {agentStatus.degraded ? (
        <p className={`text-sm ${C.textWarning}`}>
          未処理 {agentStatus.pending}件 · 判定失敗 {agentStatus.rejected}件 · 受付超過 {agentStatus.overflow + agentStatus.inputOverflow}件
        </p>
      ) : null}
      {degradedMessage ? (
        <p className={`text-sm ${C.textWarning}`}>{degradedMessage}</p>
      ) : null}
      {agentStatus.connected && agentStatus.openPorts < agentStatus.configuredPorts
        && agentStatus.lastErrorCategory === "none" ? (
          <p className={`text-sm ${C.textWarning}`}>USB接続を確認してください。</p>
        ) : null}
      <p className={`text-sm ${C.textInkMuted}`}>
        USBを選ぶ操作はありません。現在はNX600とAU10Vを自動判定します。PU-4010とIDEXXは通信条件確認後に対応します。
      </p>
    </section>
  );
}
