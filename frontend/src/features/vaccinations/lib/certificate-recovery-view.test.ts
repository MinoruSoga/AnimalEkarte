import { describe, expect, it } from "vitest";
import {
  RECOVERY_KEY_A,
  RECOVERY_KEY_B,
  RECOVERY_KEY_C,
  closedItem,
  succeededItem,
  unresolvedItem,
} from "../mock/certificate-recovery-fixtures";
import type { CertificateRecoveryDiscovery } from "./certificate-recovery-discovery";
import { deriveRecoveryView } from "./certificate-recovery-view";

function discovery(
  items: CertificateRecoveryDiscovery["items"],
  extra: Partial<CertificateRecoveryDiscovery> = {},
): CertificateRecoveryDiscovery {
  return {
    target: { intent: "issue", vaccinationId: 101 },
    items,
    recoveryGeneration: "1",
    activeOperationKey: null,
    scanEnded: true,
    truncated: false,
    ...extra,
  };
}

describe("保留（hold）の導出", () => {
  it("unresolved / unavailable は保留。succeeded / closed だけなら保留なし（安全の証明ではない）", () => {
    expect(deriveRecoveryView(discovery([unresolvedItem(RECOVERY_KEY_A)]), new Map()).hold).toBe(true);
    expect(deriveRecoveryView(discovery([unresolvedItem(RECOVERY_KEY_A, { unavailable: true })]), new Map()).hold).toBe(true);
    const view = deriveRecoveryView(discovery([succeededItem(RECOVERY_KEY_A), closedItem(RECOVERY_KEY_B)]), new Map());
    expect(view.hold).toBe(false);
    expect(view.rows.map((r) => r.kind)).toEqual(["succeeded", "closed"]);
  });

  it("0 件でも既知の阻害要因が無いだけ（rows は空、許可の語彙は返さない）", () => {
    const view = deriveRecoveryView(discovery([]), new Map());
    expect(view.rows).toEqual([]);
    expect(view.hold).toBe(false);
  });

  it("打ち切り・active key が一覧に見えない場合は保留", () => {
    expect(deriveRecoveryView(discovery([], { truncated: true, scanEnded: false }), new Map()).hold).toBe(true);
    const view = deriveRecoveryView(discovery([succeededItem(RECOVERY_KEY_A)], { activeOperationKey: RECOVERY_KEY_B }), new Map());
    expect(view.activeUnverified).toBe(true);
    expect(view.hold).toBe(true);
    expect(deriveRecoveryView(discovery([succeededItem(RECOVERY_KEY_A)], { activeOperationKey: RECOVERY_KEY_A }), new Map()).hold).toBe(false);
  });

  it("再照会で成功を確認した key だけが保留を外れ、他の未確定は残る", () => {
    const items = [unresolvedItem(RECOVERY_KEY_A), unresolvedItem(RECOVERY_KEY_B)];
    const resolved = new Map([[RECOVERY_KEY_A, { editionId: "301", artifactId: "401" }]]);
    const view = deriveRecoveryView(discovery(items), resolved);
    expect(view.rows.map((r) => r.kind)).toEqual(["succeeded", "held"]);
    expect(view.hold).toBe(true);
    const all = new Map([...resolved, [RECOVERY_KEY_B, { editionId: "302", artifactId: "402" }]]);
    expect(deriveRecoveryView(discovery(items), all).hold).toBe(false);
  });

  it("対象に束縛できない成功は採用せず保留に倒す（reprint で別の版を指す成功）", () => {
    const item = succeededItem(RECOVERY_KEY_C, { intent: "reprint", refs: { editionId: 302 } });
    const view = deriveRecoveryView(
      discovery([item], { target: { intent: "reprint", vaccinationId: 101, editionId: 301 } }),
      new Map(),
    );
    expect(view.rows[0]).toMatchObject({ kind: "held", heldReason: "unavailable" });
    expect(view.hold).toBe(true);
  });
});
