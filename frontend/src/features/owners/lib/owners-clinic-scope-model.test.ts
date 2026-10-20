import { describe, expect, it } from "vitest";
import {
  OWNERS_CLINIC_SELECTION_LIMIT,
  decideOwnersClinicSelection,
  intersectWithPermitted,
  isSameClinicSet,
  parseClinicsParam,
  pruneSavedSelection,
  toggleClinicSelection,
} from "./owners-clinic-scope-model";

// EMR-463: technical-spec#owners-list-scope-contract-emr463 の fixture 表（合成 ID）。
const ids = (n: number) => Array.from({ length: n }, (_, i) => String(i + 1));

describe("decideOwnersClinicSelection", () => {
  it("初回（URL・保存なし）は現在の全許可を選ぶ", () => {
    expect(
      decideOwnersClinicSelection({
        urlClinicIds: null,
        savedClinicIds: null,
        permittedClinicIds: ["1", "2"],
      }),
    ).toEqual({ kind: "ready", selectedClinicIds: ["1", "2"], source: "all" });
  });

  it('保存 ["2","3"]・許可 ["1","2"] は交差 ["2"] だけを有効選択にする', () => {
    expect(
      decideOwnersClinicSelection({
        urlClinicIds: null,
        savedClinicIds: ["2", "3"],
        permittedClinicIds: ["1", "2"],
      }),
    ).toEqual({ kind: "ready", selectedClinicIds: ["2"], source: "saved" });
  });

  it("保存の全件が失効したら現在の全許可へ戻す（未検証の 3 や現在医院を補わない）", () => {
    expect(
      decideOwnersClinicSelection({
        urlClinicIds: null,
        savedClinicIds: ["3"],
        permittedClinicIds: ["1", "2"],
      }),
    ).toEqual({ kind: "ready", selectedClinicIds: ["1", "2"], source: "all" });
  });

  it("URL は保存より優先し、許可外 ID を送らない", () => {
    expect(
      decideOwnersClinicSelection({
        urlClinicIds: ["2", "9"],
        savedClinicIds: ["1"],
        permittedClinicIds: ["1", "2"],
      }),
    ).toEqual({ kind: "ready", selectedClinicIds: ["2"], source: "url" });
  });

  it("URL が許可外 ID だけなら空選択にせず保存→全許可の順に決める", () => {
    expect(
      decideOwnersClinicSelection({
        urlClinicIds: ["9"],
        savedClinicIds: null,
        permittedClinicIds: ["1", "2"],
      }),
    ).toEqual({ kind: "ready", selectedClinicIds: ["1", "2"], source: "all" });
  });

  it("許可集合 [] は一覧を呼ばない状態（no-permitted-clinics）", () => {
    expect(
      decideOwnersClinicSelection({
        urlClinicIds: ["1"],
        savedClinicIds: ["1"],
        permittedClinicIds: [],
      }),
    ).toEqual({ kind: "no-permitted-clinics" });
  });

  it("選択 50 件は上限内", () => {
    const decision = decideOwnersClinicSelection({
      urlClinicIds: null,
      savedClinicIds: null,
      permittedClinicIds: ids(50),
    });
    expect(decision.kind).toBe("ready");
    expect(OWNERS_CLINIC_SELECTION_LIMIT).toBe(50);
  });

  it("選択 51 件は超過として停止し、先頭 50 件へ切り捨てない", () => {
    const decision = decideOwnersClinicSelection({
      urlClinicIds: null,
      savedClinicIds: null,
      permittedClinicIds: ids(51),
    });
    expect(decision).toEqual({
      kind: "over-limit",
      selectedClinicIds: ids(51),
      source: "all",
      limit: 50,
    });
  });

  it('"9007199254740993" を文字列のまま選択する', () => {
    expect(
      decideOwnersClinicSelection({
        urlClinicIds: ["9007199254740993"],
        savedClinicIds: null,
        permittedClinicIds: ["9007199254740993", "1"],
      }),
    ).toEqual({ kind: "ready", selectedClinicIds: ["9007199254740993"], source: "url" });
  });
});

describe("parseClinicsParam / intersectWithPermitted / isSameClinicSet", () => {
  it("未指定・空は null、重複と空要素は除く", () => {
    expect(parseClinicsParam(null)).toBeNull();
    expect(parseClinicsParam("")).toBeNull();
    expect(parseClinicsParam(",,")).toBeNull();
    expect(parseClinicsParam("1,2,1,")).toEqual(["1", "2"]);
  });

  it("交差は候補の順序を保つ", () => {
    expect(intersectWithPermitted(["3", "1", "2"], ["1", "2"])).toEqual(["1", "2"]);
  });

  it("集合比較は順序を問わない", () => {
    expect(isSameClinicSet(["1", "2"], ["2", "1"])).toBe(true);
    expect(isSameClinicSet(["1"], ["1", "2"])).toBe(false);
  });
});

describe("toggleClinicSelection", () => {
  it("選択の追加・解除", () => {
    expect(toggleClinicSelection(["1"], "2", ["1", "2"])).toEqual(["1", "2"]);
    expect(toggleClinicSelection(["1", "2"], "1", ["1", "2"])).toEqual(["2"]);
  });

  it("最後の 1 院は解除できない（空選択を作らない）", () => {
    expect(toggleClinicSelection(["1"], "1", ["1", "2"])).toBeNull();
  });

  it("許可外 ID はトグルできず、現在選択の許可外 ID も落とす", () => {
    expect(toggleClinicSelection(["1"], "9", ["1", "2"])).toBeNull();
    expect(toggleClinicSelection(["1", "9"], "2", ["1", "2"])).toEqual(["1", "2"]);
  });
});

describe("pruneSavedSelection（scope 取得成功時だけ使う）", () => {
  it("保存なし・全件許可内は変更しない", () => {
    expect(pruneSavedSelection(null, ["1"])).toEqual({ kind: "unchanged" });
    expect(pruneSavedSelection(["1"], ["1", "2"])).toEqual({ kind: "unchanged" });
  });

  it("許可外を除いた残りで置換する", () => {
    expect(pruneSavedSelection(["2", "3"], ["1", "2"])).toEqual({
      kind: "replace",
      clinicIds: ["2"],
    });
  });

  it("全件が許可外なら削除する", () => {
    expect(pruneSavedSelection(["3"], ["1", "2"])).toEqual({ kind: "remove" });
  });
});
