import { describe, expect, it } from "vitest";
import {
  draftFromInputs,
  draftToInputs,
  emptyDraft,
  layoutsForKind,
  parseFee,
  requestInputsFor,
  selectableKinds,
} from "./certificate-inputs";
import type { CertificateContext } from "../api/certificate-schemas";
import { emptyCertificateInputs } from "../api/certificate-schemas";

function contextWith(state: "eligible" | "ineligible" | "unknown", kinds: ("normal" | "rabies")[]) {
  return { eligibility: { state, allowed_kinds: kinds, provenance: "master_snapshot" } } as CertificateContext;
}

describe("証明書の入力 10 欄の変換", () => {
  it("料金は未入力(null)と明示 0 を区別する", () => {
    expect(parseFee("")).toEqual({ value: null });
    expect(parseFee("0")).toEqual({ value: 0 });
    expect(parseFee("3000")).toEqual({ value: 3000 });
  });

  it("整数でない料金・JS で安全に表せない料金は丸めず拒否する", () => {
    expect(parseFee("1.5")).toHaveProperty("error");
    expect(parseFee("1e3")).toHaveProperty("error");
    expect(parseFee("9007199254740993")).toHaveProperty("error");
  });

  it("非正準な整数表記(007 / -0)は黙って変換せず検証エラーにする", () => {
    expect(parseFee("007")).toHaveProperty("error");
    expect(parseFee("00")).toHaveProperty("error");
    expect(parseFee("-0")).toHaveProperty("error");
    expect(parseFee("0")).toEqual({ value: 0 });
    expect(parseFee("-5")).toEqual({ value: -5 });
  });

  it("先頭 0 の番号文字列をトリム・数値化せずそのまま保持する", () => {
    const draft = { ...emptyDraft(), certificate_number: "0001", distinguishing_features: " 白 " };
    const result = draftToInputs(draft);
    expect(result.inputs?.certificate_number).toBe("0001");
    expect(result.inputs?.distinguishing_features).toBe(" 白 ");
  });

  it("合計は独立入力で、3 料金の合計と違っても補正しない", () => {
    const draft = { ...emptyDraft(), injection_fee: "3000", tag_fee: "0", total: "9999" };
    const { inputs } = draftToInputs(draft);
    expect(inputs).toMatchObject({ registration_fee: null, injection_fee: 3000, tag_fee: 0, total: 9999 });
  });

  it("新規の料金初期値は 4 欄とも null（draft は空欄）", () => {
    const draft = draftFromInputs(emptyCertificateInputs());
    expect(draft.registration_fee).toBe("");
    expect(draft.total).toBe("");
    expect(draftToInputs(draft).inputs).toEqual(emptyCertificateInputs());
  });

  it("normal の request inputs は空オブジェクト、rabies は 10 欄を明示 null 含め送る", () => {
    const inputs = { ...emptyCertificateInputs(), tag_fee: 0 };
    expect(requestInputsFor("normal", inputs)).toEqual({});
    const rabies = requestInputsFor("rabies", inputs);
    expect(Object.keys(rabies)).toHaveLength(10);
    expect(rabies.tag_fee).toBe(0);
    expect(rabies.registration_fee).toBeNull();
  });
});

describe("様式の選択可否", () => {
  it("rabies は eligible のときだけ選べる（false / unknown は normal のみ）", () => {
    expect(selectableKinds(contextWith("eligible", ["normal", "rabies"]))).toEqual(["normal", "rabies"]);
    expect(selectableKinds(contextWith("ineligible", ["normal"]))).toEqual(["normal"]);
    expect(selectableKinds(contextWith("unknown", ["normal"]))).toEqual(["normal"]);
    // 契約違反の応答（unknown なのに rabies を許可）でも rabies を出さない。
    expect(selectableKinds(contextWith("unknown", ["normal", "rabies"]))).toEqual(["normal"]);
  });

  it("layout は normal が縦横、rabies は rabies のみ", () => {
    expect(layoutsForKind("normal")).toEqual(["portrait", "landscape"]);
    expect(layoutsForKind("rabies")).toEqual(["rabies"]);
  });
});
