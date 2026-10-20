import type {
  CertificateInputs,
  CertificateKind,
  CertificateLayout,
  CertificateSnapshot,
} from "../api/certificate-schemas";
import { emptyCertificateInputs } from "../api/certificate-schemas";

/**
 * EMR-487: 契約モック用の合成 fixture（実患者・実医院・実 hash・実 PDF ではない）。
 * 値の出典: docs/work/emr/emr-486-certificate-api-contract-20261020.md の合成例。
 * すべて「合成」と分かる名称にし、実運用値に見える文字列を使わない。
 */

export const MOCK_TODAY = "2026-10-20";
export const MOCK_VACCINATION_ID = 101;
export const MOCK_SOURCE_VERSION = 7;

/** 64 桁 hex の決定的ダミー（shape 用。SHA-256 の実装ではなく、照合合格の証拠にもならない）。 */
export function mockHash(input: string): string {
  let out = "";
  for (let round = 0; round < 8; round += 1) {
    let h = (0x811c9dc5 ^ (round * 0x9e3779b1)) >>> 0;
    for (let i = 0; i < input.length; i += 1) {
      h = Math.imul(h ^ input.charCodeAt(i), 0x01000193) >>> 0;
    }
    out += h.toString(16).padStart(8, "0");
  }
  return out;
}

const MOCK_PDF_TEXT = "%PDF-1.1\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj\n3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 200 100]>>endobj\ntrailer<</Root 1 0 R>>\n%%EOF\n";

/** 合成 PDF（空白 1 ページ・mock 用）。様式の描画は server renderer の責務でありここでは再現しない。 */
export function mockPdfBase64(): string {
  return btoa(MOCK_PDF_TEXT);
}

export function mockPdfBlob(): Blob {
  return new Blob([MOCK_PDF_TEXT], { type: "application/pdf" });
}

export function buildSnapshot(args: {
  kind: CertificateKind;
  layout: CertificateLayout;
  sourceVersion: number;
  rabiesEligible: boolean | null;
  inputs: CertificateInputs;
  displayIssueDate?: string;
  /** 省略時は MOCK_VACCINATION_ID。 */
  vaccinationId?: number;
}): CertificateSnapshot {
  return {
    schema_version: 1,
    kind: args.kind,
    source: {
      vaccination_id: args.vaccinationId ?? MOCK_VACCINATION_ID,
      version: args.sourceVersion,
      vaccine_id: 11,
      vaccine_name: "合成ワクチン",
      administered_on: MOCK_TODAY,
      next_scheduled_on: null,
      lots: [null, null, null, null],
      doctor_name: null,
    },
    owner: { name: "合成 飼主", name_kana: null, postal_code: null, address: null, phone: null },
    pet: {
      name: "合成ペット",
      species: null,
      breed: null,
      sex: null,
      coat_color: null,
      birth_date: null,
      age_text: null,
    },
    clinic: {
      name: "合成動物病院",
      director_name: null,
      postal_code: null,
      address: null,
      phone: null,
      fax: null,
      email: null,
      url: null,
      veterinarian_association: null,
      designated_veterinarian: null,
      veterinarian_name: null,
    },
    display_issue_date: args.displayIssueDate ?? MOCK_TODAY,
    inputs: args.inputs,
    eligibility: {
      rabies_eligible: args.rabiesEligible,
      vaccine_id: 11,
      vaccine_version: 3,
      provenance: args.rabiesEligible === null ? "legacy_unresolved" : "master_snapshot",
    },
    render_manifest: {
      renderer_version: "mock-renderer-v1",
      font_asset_version: "mock-font-v1",
      templates:
        args.kind === "normal"
          ? { portrait: "mock-portrait-v1", landscape: "mock-landscape-v1" }
          : { rabies: "mock-rabies-v1" },
    },
  };
}

/** 契約例の rabies 入力（total が 3 欄合計と一致しなくても補正しない確認用に tag_fee=0 を含む）。 */
export function sampleRabiesInputs(): CertificateInputs {
  return {
    ...emptyCertificateInputs(),
    certificate_number: "0001",
    registration_fee: null,
    injection_fee: 3000,
    tag_fee: 0,
    total: 3000,
  };
}
