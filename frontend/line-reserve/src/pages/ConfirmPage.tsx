import { useActionState } from "react";
import axios from "axios";
import liff from "@line/liff";
import { z } from "zod";
import type { ReservationFlow } from "../types/models";
import { liffApi } from "../api/liff-api";
import { LIFF_MOCK } from "../lib/liff-config";
import { ProgressDots } from "../components/ProgressDots";
import { PrimaryButton } from "../components/PrimaryButton";
import { BackButton } from "../components/BackButton";
import { formatJapaneseDate, formatTimeHHMM } from "@/shared-liff/jst-date";
import { getStepProgress } from "../lib/step-progress";
import { FINAL_CONFIRM_CTA_LABEL } from "../lib/advance-policy";

// EMR-519: 201 契約違反で予約が登録済みか不明なときの案内。
const RESULT_UNKNOWN_MESSAGE =
  "予約結果を確認できませんでした。予約が登録されている可能性があります。「予約確認・キャンセル」からご確認ください。";

const slotTakenResponseSchema = z.object({
  error: z.string().optional(),
  code: z.string().optional(),
  redirect_step: z.number().optional(),
});

interface ConfirmPageProps {
  clinicId: string;
  idToken: string;
  flow: ReservationFlow;
  reservationNotice?: string;
  cancelNotice?: string;
  privacyPolicy?: string;
  onConfirm: (reservationId: string, notes: string) => void;
  onSlotTaken: (message: string, redirectStep: number) => void;
  onBack: () => void;
  /** EMR-519: 201 契約違反（予約が登録済みか不明）のときだけ使う予約一覧への導線。 */
  onMyReservations?: () => void;
  /** EMR-519: 201 契約違反を flow state へ報告する（遷移を挟んでも再 POST しないため）。 */
  onResultUnknown?: () => void;
  /** EMR-519: flow 側に保持された結果不明フラグ。true の間は確定ボタンを無効化する。 */
  createResultUnknown?: boolean;
}

interface ConfirmFormState {
  error: string | null;
  /** 201 契約違反で結果不明。以後この画面内では確定ボタンを無効化し自動再送しない。 */
  resultUnknown?: boolean;
}

/** 予約完了後に LINE トーク画面へメッセージを送信する */
async function sendLiffMessage(flow: ReservationFlow, notes: string): Promise<void> {
  if (LIFF_MOCK) return; // ローカル開発時はスキップ
  if (!liff.isInClient()) return; // LINE アプリ外では送信不可

  const confirmNum = notes.match(/R-\d{8}-\d{4}/)?.[0] ?? "";
  const time = `${formatTimeHHMM(flow.startTime)}〜${formatTimeHHMM(flow.endTime)}`;

  const petList = flow.customerInfo.pets
    .map((p) => (p.type ? `${p.name}(${p.type})` : p.name))
    .join("、");

  const text = [
    "ご予約を承りました。",
    "",
    `■ 予約番号: ${confirmNum}`,
    `■ 日時: ${formatJapaneseDate(flow.date, true)} ${time}`,
    `■ コース: ${flow.courseName}`,
    flow.staffId > 0 ? `■ 担当: ${flow.staffName}` : "",
    petList ? `■ ペット: ${petList}` : "",
    "",
    "キャンセルはLINEメニューの",
    "「予約確認・キャンセル」から行えます。",
  ]
    .filter(Boolean)
    .join("\n");

  try {
    await liff.sendMessages([{ type: "text", text }]);
  } catch {
    // メッセージ送信失敗は予約自体には影響しないので無視
  }
}

export function ConfirmPage({
  clinicId,
  idToken,
  flow,
  reservationNotice,
  cancelNotice,
  privacyPolicy,
  onConfirm,
  onSlotTaken,
  onBack,
  onMyReservations,
  onResultUnknown,
  createResultUnknown,
}: ConfirmPageProps) {
  // SD-16: トリミング分岐で挿入される2ステップ分、以降のフロー全体の total を一貫させる
  const { current, total } = getStepProgress("confirm", flow.courseCategory === "trimming");

  async function confirmAction(
    previousState: ConfirmFormState,
    _formData: FormData,
  ): Promise<ConfirmFormState> {
    if (!flow.courseId) return previousState;

    try {
      const reservation = await liffApi.createReservation(
        clinicId,
        {
          course_id: flow.courseId,
          staff_id: flow.staffId,
          date: flow.date,
          start_time: flow.startTime,
          end_time: flow.endTime,
          customer_fields: {
            customer_name: flow.customerInfo.name,
            phone: flow.customerInfo.phone,
            owner_name: flow.customerInfo.ownerName,
            pets: flow.customerInfo.pets.map((p) => ({
              name: p.name,
              type: p.type,
              is_new: p.isNew,
            })),
          },
          request_text: flow.requestText,
          ...(flow.courseCategory === "trimming" && flow.trimmingCourseId !== null
            ? {
                trimming_course_id: flow.trimmingCourseId,
                trimming_option_ids: flow.trimmingOptionIds,
              }
            : {}),
        },
        idToken,
      );
      await sendLiffMessage(flow, reservation.notes);
      onConfirm(reservation.id, reservation.notes);
      return { error: null };
    } catch (err: unknown) {
      if (axios.isAxiosError(err) && err.response?.status === 409) {
        const parsed = slotTakenResponseSchema.safeParse(err.response.data);
        const message =
          (parsed.success && parsed.data.error) || "選択された時間枠は既に予約が入っています。";
        const redirectStep = (parsed.success && parsed.data.redirect_step) || 4;
        onSlotTaken(message, redirectStep);
        return { error: null };
      }
      if (axios.isAxiosError(err)) {
        return { error: "予約の確定に失敗しました。もう一度お試しください。" };
      }
      // EMR-519: 201 応答の契約違反（parse 失敗）は POST 自体が成功している可能性を
      // 消せない。自動再送（二重予約になり得る）は行わず、この画面内では確定ボタンを
      // 無効化したまま予約一覧での確認を案内する。flow 側にも結果不明を報告し、
      // Back → 再入場でも確定ボタンが再有効化されないようにする。
      onResultUnknown?.();
      return { error: RESULT_UNKNOWN_MESSAGE, resultUnknown: true };
    }
  }

  const [formState, formAction, isPending] = useActionState(confirmAction, { error: null });
  // EMR-519: 結果不明は画面内 state（新規の契約違反）と flow state（再入場）の和。
  const resultUnknown = Boolean(formState.resultUnknown) || Boolean(createResultUnknown);

  const petDisplay =
    flow.customerInfo.pets.length > 0
      ? flow.customerInfo.pets.map((p) => (p.type ? `${p.name}（${p.type}）` : p.name)).join("、")
      : "—";

  const rows: Array<{ label: string; value: string }> = [
    { label: "お名前", value: flow.customerInfo.name },
    { label: "電話番号", value: flow.customerInfo.phone },
    { label: "飼い主名", value: flow.customerInfo.ownerName || "—" },
    { label: "ペット", value: petDisplay },
    { label: "コース", value: flow.courseName },
    { label: "スタッフ", value: flow.staffName || "指名なし" },
    { label: "日付", value: formatJapaneseDate(flow.date) },
    {
      label: "時間",
      value: `${formatTimeHHMM(flow.startTime)} 〜 ${formatTimeHHMM(flow.endTime)}`,
    },
    { label: "ご要望", value: flow.requestText || "—" },
  ];

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={current} total={total} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-noah-teal-dark mb-4">予約内容の確認</h2>
        </div>

        <div className="flex-1 px-4 space-y-4">
          {/* 警告メッセージ（EMR-519: 結果不明時は「未完了」と断定しない） */}
          {resultUnknown ? null : (
            <div className="bg-noah-warning-box-bg border border-noah-warning-border rounded-xl px-4 py-3">
              <p className="text-sm text-noah-warning-text font-medium">
                まだ予約は完了していません
              </p>
              <p className="text-xs text-noah-warning-text-sub mt-1">
                「予約を確定する」ボタンを押してください
              </p>
            </div>
          )}

          {/* 予約内容テーブル */}
          <div className="bg-white rounded-xl border border-noah-border overflow-hidden">
            {rows.map((row) => (
              <div
                key={row.label}
                className="flex border-b border-noah-border-light last:border-b-0"
              >
                <div className="w-28 flex-shrink-0 px-4 py-3 bg-noah-teal-light text-sm text-noah-text-sub font-medium">
                  {row.label}
                </div>
                <div className="flex-1 px-4 py-3 text-sm text-noah-text break-words">
                  {row.value}
                </div>
              </div>
            ))}
          </div>

          {reservationNotice ? (
            <div className="bg-noah-info-box-bg border border-noah-info-border rounded-xl px-4 py-3">
              <p className="text-xs font-medium text-noah-info-text-sub mb-1">ご予約にあたって</p>
              <p className="text-sm text-noah-info-text whitespace-pre-wrap">{reservationNotice}</p>
            </div>
          ) : null}

          {cancelNotice ? (
            <div className="bg-noah-surface-muted border border-noah-border rounded-xl px-4 py-3">
              <p className="text-xs font-medium text-noah-text-muted mb-1">キャンセルポリシー</p>
              <p className="text-sm text-noah-text-body whitespace-pre-wrap">{cancelNotice}</p>
            </div>
          ) : null}

          {formState.error || resultUnknown ? (
            <div className="bg-noah-danger-bg border border-noah-danger-border rounded-lg px-4 py-3">
              <p className="text-sm text-noah-danger-text" role="alert">
                {formState.error ?? RESULT_UNKNOWN_MESSAGE}
              </p>
            </div>
          ) : null}
        </div>

        <form action={formAction} className="px-4 py-6 space-y-3">
          {privacyPolicy ? (
            <p className="text-xs text-center text-noah-text-muted whitespace-pre-wrap">
              {privacyPolicy}
            </p>
          ) : null}
          <PrimaryButton type="submit" disabled={isPending || resultUnknown}>
            {isPending ? "送信中..." : FINAL_CONFIRM_CTA_LABEL}
          </PrimaryButton>
          {resultUnknown && onMyReservations ? (
            <button
              type="button"
              onClick={onMyReservations}
              className="w-full py-3 px-4 border border-noah-teal rounded-xl text-noah-teal-dark font-semibold hover:bg-noah-teal-light"
            >
              予約確認・キャンセル
            </button>
          ) : null}
        </form>
      </div>
    </div>
  );
}
