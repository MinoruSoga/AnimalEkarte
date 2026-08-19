import { useCallback, useMemo, useState } from "react";
import { toast } from "sonner";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useGetPets } from "@/hooks/use-pet";
import { usePermission } from "@/hooks/use-permission";
import { C, LAYOUT } from "@/lib/design-tokens";
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
} from "../api/lab-device";
import {
  isWebSerialSupported,
  labDeviceCardTitle,
  labDeviceHasUnmapped,
} from "../lib/lab-device-board-model";
import { bytesToBase64, readLabDeviceSlot, requestLabDevicePort } from "../lib/lab-device-serial";

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
          {duplicate ? " · 再送（取込済み）" : ""}
        </p>
      </div>
      <ul className="grid grid-cols-2 gap-x-4 gap-y-1 text-sm">
        {card.items.map((item) => (
          <li key={`${card.jobId}-${item.sortOrder}-${item.deviceItemCode}`}>
            <span className="font-medium">{item.deviceItemCode}</span>
            {" "}
            {item.valueRaw}
            {item.unit ? ` ${item.unit}` : ""}
            {item.needsReview ? " · 未対応" : ""}
          </li>
        ))}
      </ul>
      {labDeviceHasUnmapped(card) ? (
        <p className={`text-sm ${C.textRed700}`}>未対応項目があります。マスタの該当行を直してから付け直してください。</p>
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

export function LabDeviceBoard() {
  const { canCreate, canEdit } = usePermission(ResourceLabImport);
  const { data: board, isLoading } = useGetLabDeviceBoard(canCreate);
  const putWait = usePutLabDeviceWait();
  const clearWait = useClearLabDeviceWait();
  const receive = useReceiveLabDeviceFrames();
  const attach = useAttachLabDeviceJob();
  const detach = useDetachLabDeviceJob();
  const [search, setSearch] = useState("");
  const [lastReceive, setLastReceive] = useState("未受信");
  const { data: pets = [] } = useGetPets(undefined, { search, limit: 8 }, { enabled: search.length > 0 });
  const slots = useMemo(() => parseLabDeviceSlots(board?.station.slotsJson ?? "[]"), [board?.station.slotsJson]);
  const serialOk = isWebSerialSupported();

  const listen = useCallback(async (hint: string, slotKey: string, baud: number) => {
    try {
      const bytes = await readLabDeviceSlot(slotKey, baud);
      if (!bytes || bytes.length === 0) {
        toast.error("受信できませんでした。口の許可を確認してください。");
        return;
      }
      const results = await receive.mutateAsync({ payloadBase64: bytesToBase64(bytes), deviceHint: hint });
      const first = results[0];
      setLastReceive(first?.duplicate ? "再送（取込済み）" : "受信");
    } catch {
      toast.error("電文を読めませんでした");
    }
  }, [receive]);

  return (
    <PageLayout
      title="検査受信"
      resource={ResourceLabImport}
      maxWidth={LAYOUT.pageContentMaxWidth.form}
      align="left"
    >
      <div className="space-y-6">
        <section className="space-y-3">
          {board?.wait ? (
            <div className="space-y-3">
              <p className="text-heading-1 font-bold">{board.wait.petName}</p>
              <p className={`text-sm ${C.textInkMuted}`}>待機中 · 最終受信 {lastReceive}</p>
              {canCreate ? (
                <Button type="button" variant="outline" onClick={() => void clearWait.mutateAsync()}>
                  待機を解除
                </Button>
              ) : null}
            </div>
          ) : (
            <div className="space-y-3">
              <p className="text-heading-1 font-semibold">受信中</p>
              <p className={`text-sm ${C.textInkMuted}`}>ペット未選択 · 最終受信 {lastReceive}</p>
              <Input
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder="ペット名で検索"
                aria-label="待機するペットを検索"
              />
              <ul className="space-y-1">
                {pets.map((pet) => (
                  <li key={pet.id}>
                    <button
                      type="button"
                      className={`text-left w-full rounded px-2 py-1 ${C.bgActive}`}
                      onClick={() => void putWait.mutateAsync(Number(pet.id))}
                    >
                      {pet.name}
                    </button>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </section>

        <section className="space-y-2">
          <h2 className="text-xl font-semibold">未紐付け</h2>
          {isLoading ? <p>読み込み中</p> : null}
          {(board?.unlinked ?? []).length === 0 ? (
            <p className={C.textInkMuted}>未紐付けの受信はありません</p>
          ) : (board?.unlinked ?? []).map((card) => (
            <JobCardView
              key={card.jobId}
              card={card}
              canEdit={canEdit}
              onAttach={board?.wait ? () => void attach.mutateAsync({ jobId: card.jobId, petId: board.wait!.petId }) : undefined}
            />
          ))}
        </section>

        <section className="space-y-2">
          <h2 className="text-xl font-semibold">保存</h2>
          {(board?.saved ?? []).map((card) => (
            <JobCardView
              key={card.jobId}
              card={card}
              canEdit={canEdit}
              onDetach={() => void detach.mutateAsync(card.jobId)}
            />
          ))}
        </section>

        <section className="space-y-2">
          <h2 className="text-xl font-semibold">口</h2>
          {!serialOk ? (
            <p className={C.textInkMuted}>このブラウザは有線シリアルに対応していません。掲示板と後付けは使えます。</p>
          ) : null}
          <div className="flex flex-wrap gap-2">
            {slots.map((slot) => (
              <div key={slot.key} className="flex gap-2">
                {serialOk ? (
                  <PrimaryButton type="button" onClick={() => void requestLabDevicePort(slot.key)}>
                    {slot.deviceHint} を許可
                  </PrimaryButton>
                ) : null}
                {serialOk && canCreate ? (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => void listen(slot.deviceHint, slot.key, slot.baud)}
                  >
                    {slot.deviceHint} を読む
                  </Button>
                ) : null}
              </div>
            ))}
          </div>
        </section>
      </div>
    </PageLayout>
  );
}
