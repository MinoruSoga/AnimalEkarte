import { useCallback, useMemo, useState } from "react";
import Axios from "axios";
import { Link } from "react-router";
import { toast } from "sonner";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { Button } from "@/components/ui/button";
import { usePermission } from "@/hooks/use-permission";
import { C, LAYOUT } from "@/lib/design-tokens";
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
import { useLabDeviceListen } from "../hooks/use-lab-device-listen";
import {
  findSlotByHint,
  groupLabDeviceCardsByDay,
  isLabDeviceAttachPersisted,
  isWebSerialSupported,
  labDeviceBoardLinkLabel,
  labDeviceCardNeedsReview,
  labDeviceCardTitle,
  labDeviceClockSkewLabel,
  labDeviceHasUnmapped,
  labDeviceLatestCardForSlot,
  labDeviceListenTone,
  labDeviceLiveReceiveLabel,
  labDeviceReceiveFailure,
  labDeviceReceivedCards,
  labDeviceReceivedDayLabel,
  labDeviceSelectableTodayVisits,
  labDeviceSlotListenLabel,
  labDeviceUnmappedMasterHref,
  type LabDeviceListenState,
} from "../lib/lab-device-board-model";
import { bytesToBase64, requestLabDevicePort } from "../lib/lab-device-serial";

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
  const { canCreate, canEdit } = usePermission(ResourceLabImport);
  const { data: board, isLoading } = useGetLabDeviceBoard(canCreate);
  const putWait = usePutLabDeviceWait();
  const clearWait = useClearLabDeviceWait();
  const receive = useReceiveLabDeviceFrames();
  const attach = useAttachLabDeviceJob();
  const detach = useDetachLabDeviceJob();
  const [lastReceives, setLastReceives] = useState<Record<string, { label: string; at: string }>>({});
  const [listenEpoch, setListenEpoch] = useState(0);
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
  const serialOk = isWebSerialSupported();
  const onFrame = useCallback(async (hint: string, bytes: Uint8Array) => {
    const slot = findSlotByHint(slots, hint);
    try {
      const results = await receive.mutateAsync({ payloadBase64: bytesToBase64(bytes), deviceHint: hint });
      const first = results[0];
      if (slot) {
        setLastReceives((current) => ({
          ...current,
          [slot.key]: {
            label: first?.duplicate ? "再送（取込済み）" : "受信",
            at: new Date().toISOString(),
          },
        }));
      }
    } catch (error) {
      const status = Axios.isAxiosError(error) ? error.response?.status : undefined;
      const failure = labDeviceReceiveFailure(status);
      if (slot) {
        setLastReceives((current) => ({
          ...current,
          [slot.key]: { label: failure.label, at: new Date().toISOString() },
        }));
      }
      toast.error(failure.message);
    }
  }, [receive, slots]);
  const listenStates = useLabDeviceListen({
    slots,
    enabled: serialOk && canCreate,
    listenEpoch,
    onFrame,
  });
  const linkLabel = labDeviceBoardLinkLabel(Object.values(listenStates));

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
              <p className={`text-sm ${C.textInkMuted}`}>待機中 · 接続 {serialOk ? linkLabel : "非対応"}</p>
              {canCreate ? (
                <Button type="button" variant="outline" onClick={() => void clearWait.mutateAsync()}>
                  待機を解除
                </Button>
              ) : null}
            </div>
          ) : (
            <div className="space-y-3">
              <p className="text-heading-1 font-semibold">受信中</p>
              <p className={`text-sm ${C.textInkMuted}`}>ペット未選択 · 接続 {serialOk ? linkLabel : "非対応"}</p>
            </div>
          )}
          <div className="space-y-2">
            <h2 className="text-xl font-semibold">検査機器</h2>
            <ul className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {slots.map((slot) => {
                const state = listenStates[slot.key] ?? (serialOk ? "needs_permission" : "unsupported");
                const latestCard = labDeviceLatestCardForSlot(slot, receivedCards);
                const live = lastReceives[slot.key];
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

        <section className="space-y-2">
          <h2 className="text-xl font-semibold">医院セットアップ</h2>
          <p className={`text-sm ${C.textInkMuted}`}>パソコンと検査機器をつなぐ許可は、最初の1回だけ必要です。許可のあとは、この画面を開いたまま検査結果を受け取ります。</p>
          {!serialOk ? (
            <p className={C.textInkMuted}>このブラウザでは検査機器を直接つなぐことができません。掲示板と後からの紐付けは使えます。</p>
          ) : null}
          <ul className="space-y-2">
            {slots.map((slot) => (
              <li key={slot.key} className="flex flex-wrap items-center gap-2">
                <span className="font-medium">{slot.deviceHint}</span>
                {serialOk ? (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      void requestLabDevicePort(slot.key).then(() => {
                        setListenEpoch((current) => current + 1);
                      });
                    }}
                  >
                    {slot.deviceHint} の接続を許可
                  </Button>
                ) : null}
              </li>
            ))}
          </ul>
        </section>
      </div>
    </PageLayout>
  );
}
