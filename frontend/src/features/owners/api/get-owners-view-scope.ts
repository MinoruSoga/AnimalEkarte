import { axios } from "@/lib/axios";
import type { ClinicMembership } from "@/types/auth";

/**
 * EMR-463: 飼主・ペット一覧の現在許可医院（owners:view）を GET /v1/me から取得する owners 専有 adapter。
 *
 * 契約: backend/docs/api.yaml `GetMeResponse.owners_view_clinic_ids`
 * （x-contract-status: selected-not-implemented）と
 * docs/design/domains/owners-pets/technical-spec.md#owners-list-scope-contract-emr463。
 *
 * - loader の各実行・再検証で毎回取得する。全体 `ME_QUERY_KEY` の cached profile や
 *   login/refresh の結果、AuthProvider の所属一覧を許可集合の根拠に流用しない。
 * - `owners_view_clinic_ids` は必須。欠落・NULL・非配列・非文字列・非正規形・重複は
 *   契約未提供／不正な取得結果としてエラーにする（`[]` へ補完しない）。
 * - 医院 ID は正規形 10 進文字列のまま扱い、JS Number へ変換しない（精度損失防止）。
 */

/** 正の正規形 10 進文字列（api.yaml の pattern と同じ）。 */
export const CANONICAL_CLINIC_ID_PATTERN = /^[1-9][0-9]*$/;

export interface OwnersViewScope {
  /** GET /me の `id`（現行 BE の staff.ID 文字列）。保存先 key の所有者に使う。 */
  userId: string;
  /** owners:view が現在許可された医院 ID 集合（正規形文字列・重複なし）。順序に認可上の意味はない。 */
  permittedClinicIds: string[];
  /** 許可医院の表示用メンバーシップ（名称は同じ /me 応答の clinics から引く）。 */
  clinics: ClinicMembership[];
}

/** 契約未提供・不正形の /me 応答。失権の証明ではない（保存選択を消さない）。 */
export class OwnersViewScopeContractError extends Error {
  constructor(reason: string) {
    super(`owners_view_clinic_ids の契約を満たさない応答です: ${reason}`);
    this.name = "OwnersViewScopeContractError";
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function parsePermittedClinicIds(raw: unknown): string[] {
  if (raw === undefined) throw new OwnersViewScopeContractError("field 欠落");
  if (raw === null) throw new OwnersViewScopeContractError("NULL");
  if (!Array.isArray(raw)) throw new OwnersViewScopeContractError("配列ではない");
  const seen = new Set<string>();
  for (const item of raw) {
    if (typeof item !== "string") {
      throw new OwnersViewScopeContractError("文字列でない ID");
    }
    if (!CANONICAL_CLINIC_ID_PATTERN.test(item)) {
      throw new OwnersViewScopeContractError("正規形でない ID");
    }
    if (seen.has(item)) throw new OwnersViewScopeContractError("重複 ID");
    seen.add(item);
  }
  return [...seen];
}

/** 名称参照用。clinics は表示名の材料にすぎず、許可集合へ ID を追加しない。 */
function collectClinicNames(raw: unknown): Map<string, ClinicMembership> {
  const byId = new Map<string, ClinicMembership>();
  if (!Array.isArray(raw)) return byId;
  for (const item of raw) {
    if (!isRecord(item)) continue;
    const clinicId = item.clinic_id;
    const clinicName = item.clinic_name;
    if (typeof clinicId !== "string" || typeof clinicName !== "string") continue;
    if (byId.has(clinicId)) continue;
    byId.set(clinicId, { clinicId, clinicName, isMain: item.is_main === true });
  }
  return byId;
}

export function parseOwnersViewScope(data: unknown): OwnersViewScope {
  if (!isRecord(data)) throw new OwnersViewScopeContractError("応答が object ではない");
  const userId = data.id;
  if (typeof userId !== "string" || userId === "") {
    throw new OwnersViewScopeContractError("id が無い");
  }
  const permittedClinicIds = parsePermittedClinicIds(data.owners_view_clinic_ids);
  const permitted = new Set(permittedClinicIds);
  const names = collectClinicNames(data.clinics);

  // 表示順は従来の所属表示（/me clinics の順）を踏襲し、名称の無い許可 ID は末尾に ID で出す。
  const clinics: ClinicMembership[] = [];
  for (const membership of names.values()) {
    if (permitted.has(membership.clinicId)) clinics.push(membership);
  }
  for (const id of permittedClinicIds) {
    if (!names.has(id)) clinics.push({ clinicId: id, clinicName: `医院ID ${id}`, isMain: false });
  }

  return { userId, permittedClinicIds, clinics };
}

export async function getOwnersViewScope(signal?: AbortSignal): Promise<OwnersViewScope> {
  const { data } = await axios.get<unknown>("/v1/me", { signal });
  return parseOwnersViewScope(data);
}
