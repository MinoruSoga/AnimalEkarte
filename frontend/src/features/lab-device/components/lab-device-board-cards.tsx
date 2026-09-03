import { Link } from "react-router";

import { Button } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";
import { formatJSTTime } from "@/lib/jst-date";

import type { LabDeviceJobCard, LabDeviceSlot, LabDeviceTodayVisit } from "../api/lab-device";
import {
  labDeviceCardNeedsReview,
  labDeviceCardTitle,
  labDeviceClockSkewLabel,
  labDeviceHasUnmapped,
  labDeviceListenTone,
  labDeviceNeedsReviewReason,
  labDeviceSlotListenLabel,
  labDeviceUnmappedMasterHref,
  type LabDeviceListenState,
} from "../lib/lab-device-board-model";

export function LabDeviceJobCardView({
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

export function LabDeviceTodayVisitCard({
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

export function LabDeviceStatusCard({
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

// FE-RC-085: ネスト三項を早期returnへ分解。
export function resolveLabDeviceReceiveTime(
  liveAt: string | undefined,
  latestCard: LabDeviceJobCard | undefined,
): string | undefined {
  if (liveAt) return formatJSTTime(liveAt);
  const measuredOrReceived = latestCard?.measuredAt || latestCard?.receivedAt;
  if (measuredOrReceived) return formatJSTTime(measuredOrReceived);
  return undefined;
}
