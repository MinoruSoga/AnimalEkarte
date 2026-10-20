import {
  CERTIFICATE_INPUT_KEYS,
  emptyCertificateInputs,
  type CertificateBodySize,
  type CertificateInputs,
  type CertificateKind,
  type CertificateLayout,
} from "../api/certificate-schemas";
import type { CertificateContext } from "../api/certificate-schemas";

/**
 * EMR-487: 狂犬病様式 10 入力欄のフォーム表現と契約変換。
 * - 画面ではすべて文字列で持つ（"0001" の先頭 0・空白を数値化・トリムしない）。
 * - 空文字だけを null にする。料金は null（未入力）と明示 0 を区別する。
 * - total は独立入力。料金 3 欄からの自動合計はしない。
 * - 新規料金の初期値は null（会計・マスタから転記しない）。
 */

export interface CertificateInputDraft {
  certificate_number: string;
  injection_tag_number: string;
  dog_registration_tag_number: string;
  body_size: "" | CertificateBodySize;
  distinguishing_features: string;
  injection_amount_text: string;
  registration_fee: string;
  injection_fee: string;
  tag_fee: string;
  total: string;
}

export type CertificateFeeKey = "registration_fee" | "injection_fee" | "tag_fee" | "total";
export const CERTIFICATE_FEE_KEYS: readonly CertificateFeeKey[] = [
  "registration_fee",
  "injection_fee",
  "tag_fee",
  "total",
];

export function emptyDraft(): CertificateInputDraft {
  return {
    certificate_number: "",
    injection_tag_number: "",
    dog_registration_tag_number: "",
    body_size: "",
    distinguishing_features: "",
    injection_amount_text: "",
    registration_fee: "",
    injection_fee: "",
    tag_fee: "",
    total: "",
  };
}

export function draftFromInputs(inputs: CertificateInputs): CertificateInputDraft {
  const text = (v: string | null) => v ?? "";
  const num = (v: number | null) => (v === null ? "" : String(v));
  return {
    certificate_number: text(inputs.certificate_number),
    injection_tag_number: text(inputs.injection_tag_number),
    dog_registration_tag_number: text(inputs.dog_registration_tag_number),
    body_size: inputs.body_size ?? "",
    distinguishing_features: text(inputs.distinguishing_features),
    injection_amount_text: text(inputs.injection_amount_text),
    registration_fee: num(inputs.registration_fee),
    injection_fee: num(inputs.injection_fee),
    tag_fee: num(inputs.tag_fee),
    total: num(inputs.total),
  };
}

const INTEGER_PATTERN = /^-?\d+$/;
/** 正準な整数表記のみ（"007" や "-0" を黙って 7 / 0 に変えない）。 */
const CANONICAL_INTEGER_PATTERN = /^(0|-?[1-9]\d*)$/;

export type DraftErrors = Partial<Record<CertificateFeeKey, string>>;

/** 料金欄: "" → null、整数文字列 → number。JS で安全に表せない値は丸めず拒否する。 */
export function parseFee(raw: string): { value: number | null } | { error: string } {
  if (raw === "") return { value: null };
  if (!INTEGER_PATTERN.test(raw)) return { error: "整数（円）で入力してください" };
  if (!CANONICAL_INTEGER_PATTERN.test(raw)) {
    return { error: "先頭の 0 や「-0」を除いた整数（円）で入力してください" };
  }
  const n = Number(raw);
  if (!Number.isSafeInteger(n)) {
    return { error: "桁数が大きすぎます（9,007,199,254,740,991 以下）" };
  }
  return { value: n };
}

export function draftToInputs(
  draft: CertificateInputDraft,
): { inputs: CertificateInputs; errors: DraftErrors } | { inputs: null; errors: DraftErrors } {
  const errors: DraftErrors = {};
  const inputs = emptyCertificateInputs();
  const orNull = (v: string) => (v === "" ? null : v);
  inputs.certificate_number = orNull(draft.certificate_number);
  inputs.injection_tag_number = orNull(draft.injection_tag_number);
  inputs.dog_registration_tag_number = orNull(draft.dog_registration_tag_number);
  inputs.distinguishing_features = orNull(draft.distinguishing_features);
  inputs.injection_amount_text = orNull(draft.injection_amount_text);
  inputs.body_size = draft.body_size === "" ? null : draft.body_size;
  for (const key of CERTIFICATE_FEE_KEYS) {
    const parsed = parseFee(draft[key]);
    if ("error" in parsed) errors[key] = parsed.error;
    else inputs[key] = parsed.value;
  }
  return Object.keys(errors).length > 0 ? { inputs: null, errors } : { inputs, errors };
}

/**
 * request の inputs。normal は 10 欄すべて null のみ許可（1 欄でも非 null は 422）のため `{}`。
 * rabies は 10 欄を明示 null 含めて送る。
 */
export function requestInputsFor(
  kind: CertificateKind,
  inputs: CertificateInputs,
): Partial<CertificateInputs> {
  if (kind === "normal") return {};
  return Object.fromEntries(CERTIFICATE_INPUT_KEYS.map((k) => [k, inputs[k]])) as Partial<CertificateInputs>;
}

export function layoutsForKind(kind: CertificateKind): readonly CertificateLayout[] {
  return kind === "normal" ? ["portrait", "landscape"] : ["rabies"];
}

/**
 * 発行可能な kind。rabies は eligibility.state === "eligible" かつ allowed_kinds に含むときだけ。
 * false / unknown は normal のみ。名称・価格・種別からの推定はしない。
 */
export function selectableKinds(context: CertificateContext): CertificateKind[] {
  const { state, allowed_kinds } = context.eligibility;
  const kinds: CertificateKind[] = [];
  if (allowed_kinds.includes("normal")) kinds.push("normal");
  if (state === "eligible" && allowed_kinds.includes("rabies")) kinds.push("rabies");
  return kinds;
}

export const LAYOUT_LABEL: Record<CertificateLayout, string> = {
  portrait: "通常（縦）",
  landscape: "通常（横）",
  rabies: "狂犬病予防注射済証",
};

export const KIND_LABEL: Record<CertificateKind, string> = {
  normal: "通常の予防接種証明書",
  rabies: "狂犬病予防注射済証",
};
