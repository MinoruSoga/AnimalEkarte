import { boundSucceededRefs } from "../api/certificate-binding";
import type { CertificateRecoveryItem } from "../api/certificate-recovery-schemas";
import type { CertificateRecoveryDiscovery } from "./certificate-recovery-discovery";

/**
 * EMR-487 Slice A: discovery の結果から、画面が表示する行と「保留」を導く純関数。
 *
 * - hold = 結果が確定していない操作がある / 一覧を最後まで確認できない / active key を確認できない。
 *   hold の間は新しい発行・再印刷を開始しない。
 * - hold でないこと = 「既知の阻害要因がない」だけ。安全の証明ではなく、0 件・complete でも新 key を作る根拠にしない。
 * - 保留の解除（再 lookup で成功が確認できた key）は resolved で渡す。成功の確認だけが保留を外せる。
 *   「成功なし」「対象なし」「通信失敗」は保留のまま（遅れて到着する旧送信が成功しうるため。fence は Slice B）。
 */

/** 再照会（同じ key の既存 lookup）の結果の種類。成功を確認できた以外は保留のまま。 */
export type RecoveryLookupNotice = "not-succeeded" | "unavailable" | "mismatch";

export interface RecoveryRefs {
  readonly editionId: string;
  readonly artifactId: string;
}

export type RecoveryRowKind = "succeeded" | "held" | "closed";

export interface RecoveryRow {
  readonly item: CertificateRecoveryItem;
  readonly kind: RecoveryRowKind;
  /** held のときの理由。unavailable は lookup 不調/metadata 不明。 */
  readonly heldReason: "unresolved" | "unavailable" | null;
  /** succeeded の表示用参照（結果が指す edition / artifact）。 */
  readonly refs: RecoveryRefs | null;
}

export interface RecoveryView {
  readonly rows: readonly RecoveryRow[];
  readonly hold: boolean;
  readonly truncated: boolean;
  /** active key が一覧に見えない（確認できない進行中の操作がありうる）。 */
  readonly activeUnverified: boolean;
}

function expectedEditionId(discovery: CertificateRecoveryDiscovery): string | undefined {
  return discovery.target.intent === "reprint" ? String(discovery.target.editionId) : undefined;
}

function rowFor(
  item: CertificateRecoveryItem,
  discovery: CertificateRecoveryDiscovery,
  resolved: ReadonlyMap<string, RecoveryRefs>,
): RecoveryRow {
  if (item.execution_state === "succeeded") {
    const refs = boundSucceededRefs(item.result, {
      operationKey: item.operation_key,
      vaccinationId: discovery.target.vaccinationId,
      editionId: expectedEditionId(discovery),
    });
    // 束縛できない成功は採用せず保留に倒す。
    return refs === null
      ? { item, kind: "held", heldReason: "unavailable", refs: null }
      : { item, kind: "succeeded", heldReason: null, refs };
  }
  const refs = resolved.get(item.operation_key) ?? null;
  if (refs !== null) return { item, kind: "succeeded", heldReason: null, refs };
  if (item.execution_state === "closed_without_effect") {
    return { item, kind: "closed", heldReason: null, refs: null };
  }
  return {
    item,
    kind: "held",
    heldReason: item.discovery_state === "unavailable" ? "unavailable" : "unresolved",
    refs: null,
  };
}

export function deriveRecoveryView(
  discovery: CertificateRecoveryDiscovery,
  resolved: ReadonlyMap<string, RecoveryRefs>,
): RecoveryView {
  const rows = discovery.items.map((item) => rowFor(item, discovery, resolved));
  const active = discovery.activeOperationKey;
  const activeUnverified = active !== null && !discovery.items.some((item) => item.operation_key === active);
  const hold = rows.some((row) => row.kind === "held") || discovery.truncated || activeUnverified;
  return { rows, hold, truncated: discovery.truncated, activeUnverified };
}
