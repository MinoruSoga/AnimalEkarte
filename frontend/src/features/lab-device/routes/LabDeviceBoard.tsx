import { useCallback, useMemo, useState } from "react";
import Axios from "axios";
import { Link } from "react-router";
import { toast } from "sonner";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/hooks/use-auth";
import { usePermission } from "@/hooks/use-permission";
import { C, LAYOUT } from "@/lib/design-tokens";
import { handleApiError } from "@/lib/handle-api-error";
import { formatJSTTime, todayJSTISO } from "@/lib/jst-date";
import { ResourceLabImport } from "@/types/generated/models";

import {
  parseLabDeviceSlots,
  useAttachLabDeviceJob,
  useClearLabDeviceWait,
  useDetachLabDeviceJob,
  useGetLabDeviceBoard,
  usePutLabDeviceWait,
  useReceiveLabDeviceFrames,
  type LabDeviceJobCard,
  type LabDeviceSlot,
  type LabDeviceTodayVisit,
} from "../api/lab-device";
import { useLabDeviceAgentListen } from "../hooks/use-lab-device-agent-listen";
import {
  groupLabDeviceCardsByDay,
  isLabDeviceAttachPersisted,
  labDeviceCardNeedsReview,
  labDeviceCardTitle,
  labDeviceClockSkewLabel,
  labDeviceHasUnmapped,
  labDeviceLatestCardForSlot,
  labDeviceListenTone,
  labDeviceLiveReceiveLabel,
  labDeviceReceiveFailure,
  requireLabDeviceReceiveResult,
  labDeviceReceivedCards,
  labDeviceReceivedDayLabel,
  labDeviceSelectableTodayVisits,
  labDeviceSlotListenLabel,
  labDeviceUnmappedMasterHref,
  type LabDeviceListenState,
} from "../lib/lab-device-board-model";

const LAB_DEVICE_AGENT_RECEIVE_TOAST_ID = "lab-device-agent-receive";

function JobCardView({
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
          保存できませんでした（検査種別が複数）
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

function TodayVisitCard({
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

function DeviceStatusCard({
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
        toast.dismiss(LAB_DEVICE_AGENT_RECEIVE_TOAST_ID);
        handleApiError(error, "検査機器電文の受信");
      } else {
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
  const linkLabel = !agentStatus.connected
    ? "切断"
    : agentStatus.configuredPorts === 0 || agentStatus.openPorts < agentStatus.configuredPorts
      ? "要確認"
      : "監視中";

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
          {board?.wait ? (
            <div className="space-y-3">
              <p className="text-heading-1 font-bold">{board.wait.petName}</p>
              <p className={`text-sm ${C.textInkMuted}`}>待機中 · 接続 {linkLabel}</p>
              {canCreate ? (
                <Button type="button" variant="outline" onClick={() => void clearWait.mutateAsync()}>
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
                const supported = slot.sourceType === "fuji_nx600" || slot.sourceType === "fuji_au10v";
                const latestCard = labDeviceLatestCardForSlot(slot, receivedCards);
                const liveCandidate = lastReceives[slot.key];
                const live = liveCandidate?.clinicId === currentClinicId ? liveCandidate : undefined;
                const state: LabDeviceListenState = supported
                  ? agentStatus.connected && agentStatus.openPorts > 0
                    && agentStatus.openPorts === agentStatus.configuredPorts
                    ? live ? "listening" : "monitoring"
                    : "disconnected"
                  : "unsupported";
                return (
                  <li key={slot.key}>
                    <DeviceStatusCard
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
          <div className="space-y-2">
            <h2 className="text-xl font-semibold">本日診療中のカルテ</h2>
            {isLoading ? <p>読み込み中</p> : null}
            {todayVisits.length === 0 && !isLoading ? (
              <p className={C.textInkMuted}>本日診療中のカルテはありません</p>
            ) : (
              <ul className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {todayVisits.map((visit) => (
                  <li key={visit.recordId}>
                    <TodayVisitCard
                      visit={visit}
                      selected={board?.wait?.petId === visit.petId}
                      disabled={!canCreate}
                      onSelect={() => void putWait.mutateAsync(visit.petId)}
                    />
                  </li>
                ))}
              </ul>
            )}
          </div>
        </section>

        <section className="space-y-2">
          <h2 className="text-xl font-semibold">未紐付け</h2>
          {(board?.unlinked ?? []).length === 0 ? (
            <p className={C.textInkMuted}>未紐付けの受信はありません</p>
          ) : (board?.unlinked ?? []).map((card) => (
            <JobCardView
              key={card.jobId}
              card={card}
              canEdit={canEdit}
              onAttach={board?.wait ? () => {
                void attach.mutateAsync({ jobId: card.jobId, petId: board.wait!.petId }).then((attached) => {
                  if (!isLabDeviceAttachPersisted(attached)) {
                    toast.error(
                      labDeviceCardNeedsReview(attached)
                        ? "保存できませんでした（検査種別が複数）"
                        : "保存できませんでした。未紐付けのままです",
                    );
                  }
                });
              } : undefined}
            />
          ))}
        </section>

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
                <JobCardView
                  key={card.jobId}
                  card={card}
                  canEdit={canEdit}
                  onDetach={card.petId ? () => void detach.mutateAsync(card.jobId) : undefined}
                  onAttach={!card.petId && board?.wait ? () => {
                    void attach.mutateAsync({ jobId: card.jobId, petId: board.wait!.petId }).then((attached) => {
                      if (!isLabDeviceAttachPersisted(attached)) {
                        toast.error(
                          labDeviceCardNeedsReview(attached)
                            ? "保存できませんでした（検査種別が複数）"
                            : "保存できませんでした。未紐付けのままです",
                        );
                      }
                    });
                  } : undefined}
                />
              ))}
            </div>
          ))}
        </section>

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
          {agentStatus.degraded && agentStatus.lastErrorCategory !== "none" ? (
            <p className={`text-sm ${C.textWarning}`}>
              {agentStatus.lastErrorCategory === "discovery_failed"
                ? "USB接続の確認に失敗しました。Macのローカル受信機を再起動してください。"
                : agentStatus.lastErrorCategory === "queue_write_failed"
                  ? "受信結果を保持できません。追加送信を止めてサポート担当へ連絡してください。"
                  : agentStatus.lastErrorCategory === "port_close_failed"
                    ? "USBポートの終了に失敗しました。Macのローカル受信機を再起動してください。"
                    : agentStatus.lastErrorCategory === "response_write_failed"
                      ? "ローカル受信機との通信に失敗しました。画面を再読み込みしてください。"
                      : "USBポートを開けません。接続とアクセス権を確認してください。"}
            </p>
          ) : null}
          {agentStatus.connected && agentStatus.openPorts < agentStatus.configuredPorts
            && agentStatus.lastErrorCategory === "none" ? (
              <p className={`text-sm ${C.textWarning}`}>USB接続を確認してください。</p>
            ) : null}
          <p className={`text-sm ${C.textInkMuted}`}>
            USBを選ぶ操作はありません。現在はNX600とAU10Vを自動判定します。PU-4010とIDEXXは通信条件確認後に対応します。
          </p>
        </section>
      </div>
    </PageLayout>
  );
}
