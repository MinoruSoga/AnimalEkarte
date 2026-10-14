// React/Framework
import { memo, useActionState, useId, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

// Internal
import { C } from "@/lib/design-tokens";
import { jstNowISOString } from "@/lib/jst-date";
import { extractApiErrorMessage } from "@/lib/handle-api-error";
import { usePermission } from "@/hooks/use-permission";
import { Button } from "@/components/ui/button";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";

// Relative
import { useCreateVital, useGetVitals } from "../api/vitals";
import {
  EDIT_INPUT_CLASS,
  formatRecordedAt,
  parseVitalsNumber,
  toggleWeightValueAndUnit,
} from "../lib/vitals-tab-table-model";
import type { BodyWeightUnit, InterviewWeightMeasurement } from "../types";

type WeightSaveState = "saved" | "error" | null;

interface InterviewWeightPanelProps {
  /** 対象カルテ id。未確定（新規作成前）では記録 UI を出さない。 */
  medicalRecordId?: string;
  /** record 自身の clinicId（拠点横断時もグローバル選択ではなくこれを使う）。 */
  recordClinicId?: string;
  isPetDeceased?: boolean;
  /**
   * 問診履歴各行に投影された過去測定（行ローカル weightMeasurements を集約したもの）。
   * undefined = 投影未提供（旧 API）、[] = 測定なし — 両者を区別する。
   */
  pastMeasurements?: InterviewWeightMeasurement[];
  /**
   * 集約側が確定した「投影未提供の医療行がある」フラグ（部分履歴）。
   * 省略時は pastMeasurements === undefined を未提供として扱う後方互換。
   */
  pastUnavailable?: boolean;
  /** 履歴取得中（測定なしとの区別のため別状態）。 */
  historyLoading?: boolean;
  /** 履歴取得エラー（部分データがあっても失敗を隠さない）。 */
  historyError?: boolean;
  /** 親カルテの編集ロック/確定。fieldset の視覚 disabled だけでなく明示ゲートへ入れる。 */
  isLocked?: boolean;
  isFinalized?: boolean;
  className?: string;
}

/**
 * EMR-356: 問診画面の体重パネル。
 * - 過去測定は行ローカル投影（props）のみ表示し、他カルテの「最新体重」は代理表示しない。
 * - 今回測定は既存 vital_records SoT（useGetVitals / useCreateVital）をそのまま使う。
 * - 保存は React19 useActionState + 既存 SubmitButton + portal 外側 form 経路。
 *   親カルテ form にネストせず、input/button は form 属性で portal form に関連付ける。
 * - 識別/権限/ロックの「コミット済み最新値」は useLayoutEffect で ref に同期し、
 *   Action は mutation 直前に必ず再検査する（捕捉済み callback の遅延発火対策）。
 */
export const InterviewWeightPanel = memo(function InterviewWeightPanel(
  props: InterviewWeightPanelProps,
) {
  // R1/識別境界: record+clinic の識別で内部エディタに key を付ける。識別切替時は
  // Action・入力・pending 状態ごと別マウントにし、旧 identity の遅延完了応答が
  // 新コンテキストの入力を触れない（アンマウントで閉じ込める）。
  const identity = `${props.medicalRecordId ?? ""}|${props.recordClinicId ?? ""}`;
  return <InterviewWeightEditor key={identity} {...props} />;
});

const InterviewWeightEditor = function InterviewWeightEditor({
  medicalRecordId,
  recordClinicId,
  isPetDeceased = false,
  pastMeasurements,
  pastUnavailable,
  historyLoading = false,
  historyError = false,
  isLocked = false,
  isFinalized = false,
  className,
}: InterviewWeightPanelProps) {
  const { canView, canEdit } = usePermission("medical-records");

  // canView=false では実クエリ自体を止める（DOM を隠すだけでは fetch は走る）。
  const {
    data: currentVitals,
    isLoading: vitalsLoading,
    isError: vitalsError,
  } = useGetVitals(medicalRecordId ?? "", recordClinicId, {
    enabled: !!medicalRecordId && canView,
  });
  const { mutateAsync: createVital, isPending: createPending } = useCreateVital(
    medicalRecordId ?? "",
    recordClinicId,
  );

  // R3: 複数パネル同居でも label/input/error/form の id が一意になるよう useId を使う。
  const uid = useId().replace(/[^a-zA-Z0-9_-]/g, "");
  const inputId = `interview-weight-input-${uid}`;
  const errorId = `interview-weight-error-${uid}`;
  const weightFormId = `interview-weight-form-${uid}`;

  const [weight, setWeight] = useState("");
  const [unit, setUnit] = useState<BodyWeightUnit>("Kg");
  const [error, setError] = useState<string | null>(null);

  // 識別コンテキスト（record/clinic）+ 最新書き込みゲート。
  // R2: isLocked/isFinalized は UI の disabled だけでなく mutation ゲートにも入れる。
  const identity = `${medicalRecordId ?? ""}|${recordClinicId ?? ""}`;
  const canWrite =
    !!medicalRecordId &&
    !!recordClinicId &&
    !isPetDeceased &&
    canEdit &&
    canView &&
    !isLocked &&
    !isFinalized;

  const gateRef = useRef({ identity, canWrite, weight, unit });
  const inFlightRef = useRef(false);
  const mountedRef = useRef(false);
  const weightFormRef = useRef<HTMLFormElement | null>(null);

  // R1: render 中に ref を書かない。コミット済みの最新ゲートだけを
  // レイアウトフェーズで同期する。識別切替時の入力リセットは wrapper の
  // keyed remount が担うため、このエディタ内では identity は不変。
  useLayoutEffect(() => {
    gateRef.current = { identity, canWrite, weight, unit };
  }, [identity, canWrite, weight, unit]);

  // アンマウント済みへの遅延完了書き込みを防ぐ。レイアウトフェーズで立てるため
  // Action が発火し得る時点では必ず反映済み（アンマウント後の queued Action は
  // mutation 直前のこのガードで fail closed する）。
  useLayoutEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  // React19 Action: portal form の submit（click/Enter いずれの経路も同一点）で起動。
  // mutation 直前にコミット済みゲートと bind 済み識別を再検査するため、
  // ロックコミット前に捕捉された発火経路からは記録できない。
  const [saveState, saveAction, savePending] = useActionState<WeightSaveState, FormData>(
    async (_prev, formData) => {
      const gate = gateRef.current;
      // アンマウント済み/queued Action は mutation を起こさない（発火前に検査）。
      if (!mountedRef.current || !gate.canWrite) return _prev;
      const boundIdentity = formData.get("recordIdentity");
      if (
        typeof boundIdentity !== "string" ||
        boundIdentity === "" ||
        boundIdentity !== gate.identity
      ) {
        // submit 時点の識別が欠落・空・または現在のコミット済み識別と違う
        // → 別患者への誤書を fail closed で拒否する。
        return _prev;
      }
      if (inFlightRef.current) return _prev;

      const parsed = parseVitalsNumber(gate.weight);
      if (parsed === null || !Number.isFinite(parsed) || parsed <= 0) {
        setError("体重は正の数値で入力してください");
        return "error";
      }
      setError(null);
      inFlightRef.current = true;
      try {
        await createVital({
          recorded_at: jstNowISOString(),
          temperature: null,
          heart_rate: null,
          respiration_rate: null,
          weight: parsed,
          weight_unit: gate.unit,
          notes: null,
        });
      } catch (err) {
        // 成功/失敗の通知は発火 identity が生きている間だけ書く。
        if (mountedRef.current && gateRef.current.identity === gate.identity) {
          setError(extractApiErrorMessage(err, "体重の記録"));
          return "error";
        }
        return _prev;
      } finally {
        inFlightRef.current = false;
      }
      // 応答が遅れて record/pet/clinic が切り替わっていた場合、
      // 新コンテキストの入力欄を旧測定で上書き/クリアしない。
      if (mountedRef.current && gateRef.current.identity === gate.identity) {
        setWeight("");
        return "saved";
      }
      return _prev;
    },
    null,
  );

  if (!canView) {
    // 読取権限なし: 過去投影の表示も出さず、実クエリも enabled:false で抑止済み。
    return null;
  }

  // Enter 経路のみ明示 submit。click は native type=submit の activation が
  // form owner を submit するため、ここで重複起動しない（1 gesture = 1 Action）。
  const requestWeightSubmit = () => {
    weightFormRef.current?.requestSubmit();
  };

  const handleUnitToggle = () => {
    // 既存契約: 未保存の入力値だけを換算する（表示済み過去測定は変えない）。
    const next = toggleWeightValueAndUnit(weight, unit);
    setWeight(next.weight);
    setUnit(next.weight_unit);
  };

  const currentWeights = (currentVitals ?? []).filter((v) => v.weight != null);
  const pastList = pastMeasurements ?? [];
  const unavailable = pastUnavailable ?? pastMeasurements === undefined;
  const showEmpty =
    !vitalsLoading &&
    !vitalsError &&
    !historyLoading &&
    !historyError &&
    !unavailable &&
    pastList.length === 0 &&
    currentWeights.length === 0;

  // 編集許可のある control はロック/確定でも「表示したまま disabled」にする
  // （捕捉済みイベント経路の回帰テストを実経路として残すため取り下げない）。
  const showWriteUi = !!medicalRecordId && !!recordClinicId && !isPetDeceased && canEdit;
  const controlsDisabled = isLocked || isFinalized || createPending || savePending;

  return (
    <div
      className={`col-span-1 lg:col-span-12 border ${C.borderMedium} ${C.bgWhite} rounded-lg p-3 ${className ?? ""}`}
    >
      <h3 className={`text-sm font-bold ${C.text} mb-2`}>体重</h3>

      {vitalsLoading ? (
        <p role="status" className={`text-sm ${C.text60} mb-2`}>
          体重を読み込み中です
        </p>
      ) : null}
      {vitalsError ? (
        <p role="alert" className={`text-sm ${C.text60} mb-2`}>
          現在の体重を取得できませんでした
        </p>
      ) : null}

      {currentWeights.length > 0 ? (
        <div className="mb-2">
          <span className={`text-xs ${C.text60}`}>今回</span>
          <ul className={`text-sm ${C.text}`}>
            {currentWeights.map((v) => (
              <li key={v.id}>
                {v.weight}
                {v.weight_unit}（{formatRecordedAt(v.recorded_at)}）
              </li>
            ))}
          </ul>
        </div>
      ) : null}

      {historyLoading ? (
        <p role="status" className={`text-sm ${C.text60}`}>
          体重の履歴を読み込み中です
        </p>
      ) : historyError ? (
        <p role="alert" className={`text-sm ${C.text60}`}>
          体重の履歴を取得できませんでした
        </p>
      ) : null}

      {pastList.length > 0 ? (
        <div className="mb-2">
          <span className={`text-xs ${C.text60}`}>過去</span>
          <ul className={`text-sm ${C.text} flex flex-wrap gap-x-3 gap-y-0.5`}>
            {pastList.map((m) => (
              <li key={m.id} className="whitespace-nowrap">
                {m.value}
                {m.unit}（{formatRecordedAt(m.measuredAt)}）
              </li>
            ))}
          </ul>
        </div>
      ) : null}

      {/* 未提供の医療行が混ざる限り「測定なし」とは断定しない（部分履歴の通知） */}
      {!historyLoading && !historyError && unavailable ? (
        <p className={`text-sm ${C.text60} mb-2`}>
          {pastList.length > 0 ? "一部の履歴の体重を確認できません" : "過去の体重を確認できません"}
        </p>
      ) : null}

      {showEmpty ? <p className={`text-sm ${C.text60} mb-2`}>体重の記録はありません</p> : null}

      {showWriteUi ? (
        <>
          {/* 体重記録専用の Action form。親カルテ form にネストさせないため
              portal へ逃がし、input/button は form 属性でこの form を owner にする。
              portal 経由の submit は DOM ではなく React ツリー上を親へバブルするため、
              onSubmit で伝播だけを止める（preventDefault は自身の Action を止める）。 */}
          {createPortal(
            <form
              id={weightFormId}
              ref={weightFormRef}
              action={saveAction}
              onSubmit={(e) => e.stopPropagation()}
            >
              <input type="hidden" name="recordIdentity" value={identity} readOnly />
            </form>,
            document.body,
          )}
          <div className="flex items-end gap-2 flex-wrap">
            <div className="flex flex-col gap-1">
              <label htmlFor={inputId} className={`text-xs ${C.text}`}>
                体重
              </label>
              <div className="flex items-center gap-1">
                <input
                  id={inputId}
                  name="weight"
                  form={weightFormId}
                  type="text"
                  inputMode="decimal"
                  className={`${EDIT_INPUT_CLASS} w-28`}
                  value={weight}
                  disabled={controlsDisabled}
                  aria-invalid={error !== null}
                  aria-describedby={error !== null ? errorId : undefined}
                  onChange={(e) => {
                    setWeight(e.target.value);
                    if (error !== null) setError(null);
                  }}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && !e.nativeEvent.isComposing) {
                      // 親カルテ form の submit ではなく体重記録として処理する
                      e.preventDefault();
                      requestWeightSubmit();
                    }
                  }}
                />
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className={`h-9 text-xs px-2 ${C.borderMedium}`}
                  disabled={controlsDisabled}
                  onClick={handleUnitToggle}
                >
                  {unit}
                </Button>
              </div>
              {error !== null ? (
                <p id={errorId} role="alert" className={`text-xs ${C.danger}`}>
                  {error}
                </p>
              ) : null}
            </div>
            <SubmitButton form={weightFormId} disabled={controlsDisabled} loadingText="記録中...">
              体重を記録
            </SubmitButton>
          </div>
        </>
      ) : null}

      {/* R3: 保存完了は aria-live status で通知する（部分失敗は role=alert 側）。 */}
      {saveState === "saved" ? (
        <p role="status" className={`text-xs ${C.text60} mt-1`}>
          体重を記録しました
        </p>
      ) : null}
    </div>
  );
};
