import { AlertTriangle, OctagonAlert } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { BADGE, C, ICON } from "@/lib/design-tokens";
import { formatDate } from "@/lib/format/date";
import type { CertificateApi } from "../api/certificate-api";
import type { CertificateRecovery } from "../hooks/use-certificate-recovery";
import {
  RECOVERY_ACTIVE_UNVERIFIED_TEXT,
  RECOVERY_CLOSED_TEXT,
  RECOVERY_HELD_REASON,
  RECOVERY_INTENT_LABEL,
  RECOVERY_LOADING,
  RECOVERY_LOOKUP_NOTICE,
  RECOVERY_OWN_PENDING_TEXT,
  RECOVERY_TRUNCATED_TEXT,
  recoveryErrorText,
  recoveryHoldText,
} from "../lib/certificate-recovery-messages";
import { CERTIFICATE_RECOVERY_FENCE_AVAILABLE, RECOVERY_FENCE_PAUSED_TEXT } from "../lib/certificate-recovery-gate";
import type { RecoveryRow } from "../lib/certificate-recovery-view";
import { CertificateArtifactPrint } from "./CertificateArtifactPrint";

interface CertificateRecoveryPanelProps {
  api: CertificateApi;
  intent: "issue" | "reprint";
  recovery: CertificateRecovery;
  /** この画面の in-memory フローが持つ未確定の operation key（同じ key は既存フローが所有する）。 */
  ownPendingKey: string | null;
  /** 新しい操作の一時停止の案内を出すか（操作ボタンを持つ場所だけ）。 */
  showPausedNotice?: boolean;
}

const shortKey = (key: string) => key.slice(0, 8);

/**
 * 再 mount / reload 後に発見した、職員本人の issue / reprint 操作の表示（EMR-487 Slice A）。
 * 結果不明は保留（新しい操作を止める）で、再照会は同じ key の既存 lookup だけ。
 * 「新しく始める」「続ける」操作は持たない（世代 fence と新 key での再開は Slice B）。
 */
export function CertificateRecoveryPanel({ api, intent, recovery, ownPendingKey, showPausedNotice = false }: CertificateRecoveryPanelProps) {
  const label = RECOVERY_INTENT_LABEL[intent];
  const { state, view } = recovery;
  const [shown, setShown] = useState<ReadonlySet<string>>(new Set());

  const paused =
    showPausedNotice && !CERTIFICATE_RECOVERY_FENCE_AVAILABLE ? (
      <p role="status" className={`rounded-xs border p-3 text-sm ${BADGE.blue}`}>
        {RECOVERY_FENCE_PAUSED_TEXT}
      </p>
    ) : null;

  if (state.status === "idle") return paused;
  if (state.status === "loading") {
    return (
      <div className="space-y-3">
        {paused}
        <p role="status" aria-busy="true" className={`text-sm ${C.text60}`}>
          {RECOVERY_LOADING}
        </p>
      </div>
    );
  }
  if (state.status === "error" || view === null) {
    return (
      <div className="space-y-3">
        {paused}
      <div role="alert" className={`rounded-xs border p-3 ${BADGE.red}`}>
        <p className="flex items-center gap-1.5 text-sm font-semibold">
          <OctagonAlert className={ICON.xs} aria-hidden="true" />
          未確定の操作を確認できません
        </p>
        <p className={`mt-1 text-sm ${C.text}`}>{recoveryErrorText(state.errorKind, label)}</p>
        <div className="mt-2">
          <Button type="button" variant="outline" size="sm" onClick={recovery.refresh}>
            再確認
          </Button>
        </div>
      </div>
      </div>
    );
  }

  const held = view.rows.filter((row) => row.kind === "held");
  // 再照会で完了を確認した操作は折りたたまず見せる（結果を見落とさせない）。
  const confirmed = view.rows.filter((row) => row.kind === "succeeded" && state.resolved.has(row.item.operation_key));
  const others = view.rows.filter((row) => row.kind !== "held" && !confirmed.includes(row));
  const hold = recoveryHoldText(label);

  function renderHeld(row: RecoveryRow) {
    const key = row.item.operation_key;
    const own = key === ownPendingKey;
    const lookup = state.lookups[key];
    return (
      <li key={key} className={`space-y-1 rounded-xs border p-2 text-sm ${C.borderLight}`}>
        <p>
          <span className="font-semibold">{RECOVERY_HELD_REASON[row.heldReason ?? "unresolved"]}</span>
          <span className={C.text60}>
            {" "}
            ／ 受付 {formatDate(row.item.claimed_at)} ／ 操作 {shortKey(key)}
          </span>
        </p>
        {own ? (
          <p className={C.text60}>{RECOVERY_OWN_PENDING_TEXT}</p>
        ) : (
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={lookup?.checking === true}
            onClick={() => void recovery.lookupItem(key)}
          >
            この操作の結果を再照会
          </Button>
        )}
        {lookup?.notice ? (
          <p role="status" className={C.text60}>
            {RECOVERY_LOOKUP_NOTICE[lookup.notice]}
          </p>
        ) : null}
      </li>
    );
  }

  function renderOther(row: RecoveryRow) {
    const key = row.item.operation_key;
    return (
      <li key={key} className={`space-y-2 rounded-xs border p-2 text-sm ${C.borderLight}`}>
        <p className={C.text60}>
          受付 {formatDate(row.item.claimed_at)} ／ 操作 {shortKey(key)} ／{" "}
          {row.kind === "succeeded" ? (state.resolved.has(key) ? "再照会で完了を確認しました" : "完了を確認済み") : RECOVERY_CLOSED_TEXT}
        </p>
        {key === ownPendingKey ? (
          // 同じ key をこの画面の既存フローも持っている間は、結果の採用・表示を既存フローに一本化する。
          <p className={C.text60}>{RECOVERY_OWN_PENDING_TEXT}</p>
        ) : row.kind === "succeeded" && row.refs !== null ? (
          shown.has(key) ? (
            <CertificateArtifactPrint
              api={api}
              editionId={row.refs.editionId}
              artifactId={row.refs.artifactId}
              heading={`確認できた${label}の結果`}
            />
          ) : (
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setShown((prev) => new Set(prev).add(key))}
            >
              結果を表示
            </Button>
          )
        ) : null}
      </li>
    );
  }

  return (
    <div className="space-y-3">
      {paused}
      {view.hold ? (
        <div role="alert" className={`rounded-xs border p-3 ${BADGE.yellow}`}>
          <p className="flex items-center gap-1.5 text-sm font-semibold">
            <AlertTriangle className={ICON.xs} aria-hidden="true" />
            {hold.title}
          </p>
          <p className={`mt-1 text-sm ${C.text}`}>{hold.detail}</p>
          {view.truncated ? <p className={`mt-1 text-sm ${C.text}`}>{RECOVERY_TRUNCATED_TEXT}</p> : null}
          {view.activeUnverified ? <p className={`mt-1 text-sm ${C.text}`}>{RECOVERY_ACTIVE_UNVERIFIED_TEXT}</p> : null}
          {held.length > 0 ? <ul className="mt-2 space-y-2">{held.map(renderHeld)}</ul> : null}
          <div className="mt-2">
            <Button type="button" variant="outline" size="sm" onClick={recovery.refresh}>
              再確認
            </Button>
          </div>
        </div>
      ) : null}
      {confirmed.length > 0 ? <ul className="space-y-2">{confirmed.map(renderOther)}</ul> : null}
      {others.length > 0 ? (
        <details className="text-sm">
          <summary className={C.text60}>確認できた過去の{label}操作（{others.length}件）</summary>
          <ul className="mt-2 space-y-2">{others.map(renderOther)}</ul>
        </details>
      ) : null}
    </div>
  );
}
