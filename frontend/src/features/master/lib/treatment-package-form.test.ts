import { describe, expect, it } from "vitest";
import type { TreatmentPackage } from "../api/treatment-packages";
import {
  buildReorderBody,
  canSendReorder,
  draftFromPackage,
  emptyPackageDraft,
  isDraftEqual,
  moveId,
  newComponentDraft,
  reorderBlockReason,
  requireEditingVersion,
  toCreateBody,
  toPatchBody,
  treatmentPackageKindViewResource,
  validateDraft,
} from "./treatment-package-form";

const comp = (key: string, kind: string, id: string, quantity = "1") =>
  ({
    component_key: key,
    item: { kind, [`${kind}_id`]: id },
    quantity,
  }) as TreatmentPackage["components"][number];

const basePkg: TreatmentPackage = {
  id: "201",
  clinic_id: "7",
  version: "3",
  name: "セット",
  description: "",
  is_active: true,
  sort_order: 0,
  created_at: "t",
  updated_at: "t",
  components: [
    comp("a0000000-0000-4000-8000-000000000011", "consultation", "101"),
    comp("a0000000-0000-4000-8000-000000000012", "procedure", "102"),
    comp("a0000000-0000-4000-8000-000000000013", "medicine", "103", "0.25"),
    comp("a0000000-0000-4000-8000-000000000014", "examination_type", "104"),
    comp("a0000000-0000-4000-8000-000000000015", "vaccine", "105"),
    comp("a0000000-0000-4000-8000-000000000016", "checkup_type", "106"),
    comp("a0000000-0000-4000-8000-000000000017", "merchandise_item", "107"),
    {
      component_key: "a0000000-0000-4000-8000-000000000018",
      item: {
        kind: "manual_other",
        manual: {
          name: "手入力項目",
          reason: "理由",
          unit_label: null,
          unit_price: "110",
          tax_type: "included",
          tax_rate: "0.1",
          is_non_insurance: true,
        },
      },
      quantity: "1",
    },
  ],
};

describe("draftFromPackage（7master＋manual混在・型別IDをnumber化しない）", () => {
  it("全8分岐を保持しID/quantity/円/税率を文字列のまま往復する", () => {
    const draft = draftFromPackage(basePkg);
    expect(draft.components).toHaveLength(8);
    expect(draft.components[2]!.masterId).toBe("103");
    expect(draft.components[2]!.quantity).toBe("0.25");
    const manual = draft.components[7]!.manual;
    expect(manual.unitPrice).toBe("110");
    expect(manual.taxRate).toBe("0.1");
    const body = toCreateBody(draft)!;
    expect(body.components[2]!.item).toEqual({ kind: "medicine", medicine_id: "103" });
    expect(body.components[2]!.quantity).toBe("0.25");
  });

  it("重複masterも別component_keyの別行を保持する", () => {
    const draft = draftFromPackage(basePkg);
    draft.components.push({ ...newComponentDraft("consultation"), masterId: "101" });
    const body = toCreateBody(draft)!;
    expect(new Set(body.components.map((c) => c.component_key)).size).toBe(9);
  });
});

describe("toPatchBody（全構成置換・CAS・no-op版不増分）", () => {
  it("no-op は null", () => {
    expect(toPatchBody(draftFromPackage(basePkg), basePkg)).toBeNull();
  });
  it("name のみ変更は components を送らない（省略=維持）", () => {
    const draft = draftFromPackage(basePkg);
    draft.name = "新名称";
    const body = toPatchBody(draft, basePkg)!;
    expect(body).toEqual({ version: "3", name: "新名称" });
  });
  it("構成変更は全構成置換で送る", () => {
    const draft = draftFromPackage(basePkg);
    draft.components = draft.components.slice(0, 1);
    expect(toPatchBody(draft, basePkg)!.components).toHaveLength(1);
  });
  it("is_active 明示変更のみ送り未設定を false 補完しない", () => {
    const draft = draftFromPackage(basePkg);
    draft.isActive = false;
    expect(toPatchBody(draft, basePkg)).toEqual({ version: "3", is_active: false });
  });
});

describe("validateDraft（未入力を補完せずエラー）", () => {
  it("0件構成・不正ID・非canonical数量を検出", () => {
    const draft = emptyPackageDraft();
    draft.components = [];
    expect(validateDraft(draft).valid).toBe(false);
    const bad = emptyPackageDraft();
    bad.components[0]!.masterId = "abc";
    bad.components[0]!.quantity = "1.230";
    const v = validateDraft(bad);
    expect(v.valid).toBe(false);
    expect(v.componentErrors.get(bad.components[0]!.componentKey)).toHaveLength(2);
  });
  it("manual_other は name/reason/単価/税率必須", () => {
    const draft = emptyPackageDraft();
    draft.components[0]!.kind = "manual_other";
    expect(validateDraft(draft).valid).toBe(false);
  });
});

describe("requireEditingVersion / reorder / moveId / isDraftEqual", () => {
  it("canonical 文字列のみ受理", () => {
    expect(() => requireEditingVersion("3")).not.toThrow();
    for (const v of [undefined, "", "0", "01", 3, -1]) {
      expect(() => requireEditingVersion(v)).toThrow("編集開始時");
    }
  });
  it("reorder body は order_version + 全件 ids", () => {
    expect(buildReorderBody("17", ["201", "202"])).toEqual({
      order_version: "17",
      ids: ["201", "202"],
    });
  });
  it("moveId は端で動かず全件保持", () => {
    expect(moveId(["a", "b"], "a", -1)).toEqual(["a", "b"]);
    expect(moveId(["a", "b"], "a", 1)).toEqual(["b", "a"]);
  });
  it("isDraftEqual は未変更で true", () => {
    expect(isDraftEqual(draftFromPackage(basePkg), basePkg)).toBe(true);
  });
});

describe("Q1: reorderBlockReason（raw 件数・pending/error/order_version で事前抑止）", () => {
  const base = { isPending: false, isError: false, orderVersion: "17", rawItemCount: 0 };
  it("9,999件は抑止しない", () => {
    expect(reorderBlockReason({ ...base, rawItemCount: 9999 })).toBeNull();
  });
  it("10,000件・10,001件は抑止する（保守的制約）", () => {
    expect(reorderBlockReason({ ...base, rawItemCount: 10000 })).toBe("count-guard");
    expect(reorderBlockReason({ ...base, rawItemCount: 10001 })).toBe("count-guard");
  });
  it("pending / error / order_version なしは送信不可", () => {
    expect(reorderBlockReason({ ...base, isPending: true })).toBe("pending");
    expect(reorderBlockReason({ ...base, isError: true })).toBe("error");
    expect(reorderBlockReason({ ...base, orderVersion: undefined })).toBe("no-order-version");
  });
});

describe("B2-02: canSendReorder（handler 抑止・ボタン disabled と独立）", () => {
  const base = {
    authorizationLost: false,
    block: null as ReturnType<typeof reorderBlockReason>,
    orderVersion: "17" as string | undefined,
    canEdit: true,
  };
  it("全て通れば送信可", () => {
    expect(canSendReorder(base)).toBe(true);
  });
  it("pending / error / 件数閾値 / 版なし / 失権 / 編集権限なしは送信不可", () => {
    expect(canSendReorder({ ...base, block: "pending" })).toBe(false);
    expect(canSendReorder({ ...base, block: "error" })).toBe(false);
    expect(canSendReorder({ ...base, block: "no-order-version" })).toBe(false);
    expect(canSendReorder({ ...base, block: "count-guard" })).toBe(false);
    expect(canSendReorder({ ...base, orderVersion: undefined })).toBe(false);
    expect(canSendReorder({ ...base, authorizationLost: true })).toBe(false);
    expect(canSendReorder({ ...base, canEdit: false })).toBe(false);
  });
});

describe("Q3: treatmentPackageKindViewResource（backend 型別 view resource と対応）", () => {
  it("checkup_type → checkups / merchandise_item → master-merchandise / 他 → master-medical", () => {
    expect(treatmentPackageKindViewResource("consultation")).toBe("master-medical");
    expect(treatmentPackageKindViewResource("procedure")).toBe("master-medical");
    expect(treatmentPackageKindViewResource("medicine")).toBe("master-medical");
    expect(treatmentPackageKindViewResource("examination_type")).toBe("master-medical");
    expect(treatmentPackageKindViewResource("vaccine")).toBe("master-medical");
    expect(treatmentPackageKindViewResource("checkup_type")).toBe("checkups");
    expect(treatmentPackageKindViewResource("merchandise_item")).toBe("master-merchandise");
  });
});

describe("S08 diff helpers（draftPackageDiff / component summary）", () => {
  it("項目別に差分を返す（構成は全置換比較）", async () => {
    const { draftPackageDiff } = await import("./treatment-package-form");
    const draft = draftFromPackage(basePkg);
    const same = draftPackageDiff(draft, basePkg);
    expect(same).toEqual({ name: false, description: false, isActive: false, components: false });

    draft.name = "別名";
    draft.isActive = !basePkg.is_active;
    draft.components[0]!.masterId = "999";
    const changed = draftPackageDiff(draft, basePkg);
    expect(changed.name).toBe(true);
    expect(changed.isActive).toBe(true);
    expect(changed.components).toBe(true);
    expect(changed.description).toBe(false);
  });

  it("構成サマリは kind ラベル + 識別子 + 数量（number 化しない）", async () => {
    const {
      draftComponentSummary,
      wireComponentSummary,
      draftFromPackage: dfp,
    } = await import("./treatment-package-form");
    const draft = dfp(basePkg);
    expect(draftComponentSummary(draft.components[0]!)).toBe("診察 101 ×1");
    expect(draftComponentSummary(draft.components[7]!)).toContain("手入力（その他）");
    expect(wireComponentSummary(basePkg.components[0]!)).toBe("診察 101 ×1");
    expect(wireComponentSummary(basePkg.components[6]!)).toContain("物販");
  });

  it("A1-02/A2-01: manual の全更新可能フィールドを文字列のまま表示する", async () => {
    const {
      draftComponentSummary,
      wireComponentSummary,
      draftFromPackage: dfp,
    } = await import("./treatment-package-form");
    const draft = dfp(basePkg);
    expect(draftComponentSummary(draft.components[7]!)).toContain("単価:110円");
    expect(draftComponentSummary(draft.components[7]!)).toContain("税込(税率0.1)");
    expect(draftComponentSummary(draft.components[7]!)).toContain("理由:理由");
    expect(draftComponentSummary(draft.components[7]!)).toContain("単位:なし");
    expect(draftComponentSummary(draft.components[7]!)).toContain("保険外");
    expect(wireComponentSummary(basePkg.components[7]!)).toContain("単価:110円");
    expect(wireComponentSummary(basePkg.components[7]!)).toContain("税込(税率0.1)");
    expect(wireComponentSummary(basePkg.components[7]!)).toContain("保険外");
    // 単価だけが違う場合はサマリ文字列が異なり、差分が可視化される。
    const changed = dfp({ ...basePkg, components: [basePkg.components[7]!] });
    changed.components[0]!.manual.unitPrice = "120";
    expect(draftComponentSummary(changed.components[0]!)).not.toBe(
      wireComponentSummary(basePkg.components[7]!),
    );
  });
});
