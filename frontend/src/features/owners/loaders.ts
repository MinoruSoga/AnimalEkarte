import { isAxiosError } from "axios";
import { getOwner } from "./api/get-owner";
import { getOwnersViewScope, type OwnersViewScope } from "./api/get-owners-view-scope";
import {
  decideOwnersClinicSelection,
  isSameClinicSet,
  parseClinicsParam,
  pruneSavedSelection,
  type OwnersClinicSelectionDecision,
} from "./lib/owners-clinic-scope-model";
import { OWNERS_CLINIC_SCOPE_ACTIVE } from "./lib/owners-clinic-scope-activation";
import {
  readOwnersClinicSelection,
  removeOwnersClinicSelection,
  writeOwnersClinicSelection,
} from "./lib/owners-clinic-selection-storage";
// `axios` は Axios.create() 由来の instance で isAxiosError static を持たないため、
// 型ガードは axios package の named export を使う（features/auth 等と同じ流儀）。
import { axios } from "@/lib/axios";
import {
  PET_GENDER_MAP,
  ACQUISITION_TYPE_MAP,
  DANGER_LEVEL_MAP,
  NEUTERED_STATUS_MAP,
  mapPetStatusLabel,
} from "@/lib/transforms/pet";
import type { Pet } from "@/types";
import type { ClinicMembership } from "@/types/auth";
import type { PetListResponse } from "@/types/generated/pet-responses";
import type { Owner } from "@/types/owner";

// PERF-E5-N1-PETS: GET /v1/pets の list DTO は tygo 生成型 PetListResponse が正本。
// 以前は Go 側が非公開 struct（petListResponse）で生成対象外だったため欠落フィールドを
// ローカル interface で補っていたが、exported 化により手管理の drift 経路を解消した。

interface PetsResponse {
  data: PetListResponse[];
  total: number;
  page: number;
  limit: number;
}

// #266: 飼主・ペット一覧のページサイズ。以前の client-side usePagination のデフォルト(20件)を踏襲。
// ページ粒度はペット行単位（旧: 飼主単位のフラット化）— 1ページの行数は常に一定になる。
const OWNERS_PAGE_SIZE = 20;

// PERF-E5-N1-PETS: 飼主詳細のペット一括取得は単一リクエスト契約のため、
// backend ParsePagination の上限（DefaultMaxPaginationLimit=100）を limit に使う。
const OWNER_PETS_MAX_LIMIT = 100;

/**
 * EMR-463: 一覧の状態。scope 取得失敗・許可ゼロ・選択超過・一覧 0 件（ready + total=0）を区別する。
 * `scopeRechecked` は GET /pets 403 後に現在許可を再取得し、選択を照合し直して取得できたことを示す。
 */
export type OwnersListState =
  | { kind: "ready"; scopeRechecked: boolean }
  | { kind: "scope-error" }
  | { kind: "no-permitted-clinics" }
  | { kind: "selection-over-limit"; selectedCount: number; limit: number };

/** 医院フィルタの表示材料。許可医院だけを持つ（所属全件ではない）。 */
export interface OwnersClinicScopeView {
  /** 選択の保存先 key の所有者（GET /me の id）。 */
  userId: string;
  /** owners:view が現在許可された医院（表示名付き）。 */
  clinics: ClinicMembership[];
  /** この一覧取得に使った（または超過で停止した）有効選択。 */
  selectedClinicIds: string[];
}

export interface OwnersLoaderData {
  pets: Pet[];
  page: number;
  limit: number;
  total: number;
  listState: OwnersListState;
  /**
   * 活性時: scope 取得失敗なら null（所属全件・現在医院・旧 scope へ fallback しない）。
   * 非活性時（OWNERS_CLINIC_SCOPE_ACTIVE=false）: 常に null で listState は ready。
   */
  clinicScope: OwnersClinicScopeView | null;
}

// pets 一覧レスポンス → フロントエンド Pet 型変換。
// transformBackendPetToFrontend (lib/transforms/pet.ts) と同じフィールド構成を維持するが、
// PetListResponse の薄い owner サマリ（address1/2 等を持たない）に合わせて address は常に undefined。
function transformPetListItemToFrontend(p: PetListResponse): Pet {
  return {
    id: String(p.id),
    clinicId: p.clinic_id != null ? String(p.clinic_id) : undefined,
    ownerId: String(p.owner_id),
    ownerNumber: p.owner?.owner_number,
    ownerName: p.owner?.name ?? "",
    ownerNameKana: p.owner?.name_kana ?? undefined,
    // スタッフ向け飼主危険マーク (EMR-173)。detail 経路 (transformBackendPetToFrontend) と
    // 同一契約: true の時だけ optional キーを出す。
    ...(p.owner?.is_dangerous === true && { ownerIsDangerous: true }),
    address: undefined,
    // detail 経路 (transformBackendPetToFrontend) と同じ fallback 契約:
    // owner サマリの phone が空ならペット個体の phone に倒す。
    phone: p.owner?.phone || p.phone || "",
    petNumber: p.pet_number,
    name: p.name ?? "",
    // detail 経路 (transformBackendPetToFrontend) と同じく空文字は undefined に揃える。
    petNameKana: p.pet_name_kana || undefined,
    species: p.animal_species?.name ?? "",
    animalSpeciesId: p.animal_species_id != null ? String(p.animal_species_id) : undefined,
    breed: p.breed,
    color: p.color,
    bloodType: p.blood_type ?? undefined,
    microchipNumber: p.microchip_number ?? undefined,
    gender: p.gender ? (PET_GENDER_MAP[p.gender] ?? p.gender) : undefined,
    status: mapPetStatusLabel(p.status),
    birthDate: p.birth_date ? p.birth_date.split("T")[0] : undefined,
    neuteredDate: p.neutered_date ? p.neutered_date.split("T")[0] : undefined,
    neuteredStatus: p.neutered_status
      ? (NEUTERED_STATUS_MAP[p.neutered_status] ?? p.neutered_status)
      : undefined,
    weight: p.weight?.toString(),
    food: p.food,
    environment: p.environment,
    acquisitionType: p.acquisition_type
      ? (ACQUISITION_TYPE_MAP[p.acquisition_type] ?? p.acquisition_type)
      : undefined,
    dangerLevel: p.danger_level ? (DANGER_LEVEL_MAP[p.danger_level] ?? p.danger_level) : undefined,
    dangerReason: p.danger_reason,
    lastVisit: p.last_visit ? p.last_visit.split("T")[0] : undefined,
    insuranceId: p.insurance_id != null ? String(p.insurance_id) : undefined,
    insuranceName: p.insurance?.name,
    insuranceDetails:
      p.insurance?.coverage_rate != null ? `${p.insurance.coverage_rate}%補償` : undefined,
    remarks: p.remarks,
    // EMR-174: detail 経路 (transformBackendPetToFrontend) と同じ NULL=undefined 契約
    nameOrigin: p.name_origin ?? undefined,
    meetingStory: p.meeting_story ?? undefined,
    // PERF-E5-N1-PETS: PetListResponse は deceased_at / deceased_reason を持つ
    // （ownerLoader の単一 list リクエスト化で detail 経路と情報量を揃えた）。
    deceasedAt: p.deceased_at,
    deceasedReason: p.deceased_reason,
  };
}

function emptyOwnersData(
  listState: OwnersListState,
  clinicScope: OwnersClinicScopeView | null,
): OwnersLoaderData {
  return { pets: [], page: 1, limit: OWNERS_PAGE_SIZE, total: 0, listState, clinicScope };
}

function toScopeView(scope: OwnersViewScope, selectedClinicIds: string[]): OwnersClinicScopeView {
  return { userId: scope.userId, clinics: scope.clinics, selectedClinicIds };
}

/**
 * 現在許可の取得に成功したときだけ、保存選択から許可外 ID を除く
 * （FR-common-ui-list-filter「権限喪失医院は保存済み選択からも除外する」）。
 * 通信障害・契約未提供・中断済み要求では保存値に触れない。
 */
function pruneSavedSelectionAfterScope(
  scope: OwnersViewScope,
  savedClinicIds: string[] | null,
  signal: AbortSignal,
): void {
  if (signal.aborted) return;
  const action = pruneSavedSelection(savedClinicIds, scope.permittedClinicIds);
  if (action.kind === "replace") {
    writeOwnersClinicSelection(scope.userId, action.clinicIds);
  } else if (action.kind === "remove") {
    removeOwnersClinicSelection(scope.userId);
  }
}

type ScopedSelection =
  | { kind: "scope-error" }
  | { kind: "decided"; scope: OwnersViewScope; decision: OwnersClinicSelectionDecision };

/** scope を取得し、URL／保存選択と交差して有効選択を確定する。 */
async function resolveScopedSelection(
  urlClinicIds: string[] | null,
  signal: AbortSignal,
): Promise<ScopedSelection> {
  let scope: OwnersViewScope;
  try {
    scope = await getOwnersViewScope(signal);
  } catch (err) {
    // 中断（user/clinic 切替・再検証）は react-router が結果を捨てるため、そのまま伝播する。
    if (signal.aborted) throw err;
    // 通信障害・5xx・契約未提供は失権の証明ではない。許可ゼロへ縮退せず、保存選択も消さない。
    return { kind: "scope-error" };
  }
  const savedClinicIds = readOwnersClinicSelection(scope.userId);
  const decision = decideOwnersClinicSelection({
    urlClinicIds,
    savedClinicIds,
    permittedClinicIds: scope.permittedClinicIds,
  });
  pruneSavedSelectionAfterScope(scope, savedClinicIds, signal);
  return { kind: "decided", scope, decision };
}

function httpStatusOf(err: unknown): number {
  return isAxiosError(err) ? (err.response?.status ?? 500) : 500;
}

/**
 * EMR-463 非活性時（OWNERS_CLINIC_SCOPE_ACTIVE=false）の一覧取得。PR 前（main 05da02f3）と同じ挙動:
 * URL の `?clinics=` をそのまま `clinic_ids` へ透過し（未指定は現在の医院のみ）、GET /me は呼ばない。
 * 所属検証はサーバ側 resolveListClinicIDs が行う。
 */
async function legacyOwnersLoader(request: Request): Promise<OwnersLoaderData> {
  try {
    const searchParams = new URL(request.url).searchParams;
    const clinics = searchParams.get("clinics") ?? undefined;
    const page = Math.max(1, Number(searchParams.get("page") ?? 1) || 1);
    const search = searchParams.get("search") || undefined;
    const species = searchParams.get("species") || undefined;
    const includeDeceased = searchParams.get("include_deceased") === "true" ? "true" : undefined;
    const checkupHistory = searchParams.get("checkup_history") || undefined;

    const { data: result } = await axios.get<PetsResponse>("/v1/pets", {
      params: {
        ...(clinics ? { clinic_ids: clinics } : {}),
        page,
        limit: OWNERS_PAGE_SIZE,
        ...(search ? { search } : {}),
        ...(species ? { species } : {}),
        ...(includeDeceased ? { include_deceased: includeDeceased } : {}),
        ...(checkupHistory ? { checkup_history: checkupHistory } : {}),
      },
    });

    return {
      pets: result.data.map(transformPetListItemToFrontend),
      page: result.page,
      limit: result.limit,
      total: result.total,
      listState: { kind: "ready", scopeRechecked: false },
      // null = 現在許可 scope を使っていない（OwnersList は既存の useClinicScope 表示に戻る）
      clinicScope: null,
    };
  } catch (err) {
    throw new Response("飼主一覧の取得に失敗しました", { status: httpStatusOf(err) });
  }
}

/**
 * 飼主・ペット一覧ローダー — #266: GET /v1/pets をペット行粒度でサーバサイドページネーション取得する
 * （owners-pets-list-plan.md の PO 決定: owners API+EXISTS 応急案ではなく pets API 拡張を正本とする）。
 * URL の page/search/species/include_deceased/checkup_history をそのまま backend に転送する
 * （species は animal_species_id 数値、include_deceased 未指定 = 生存のみが既定、
 * checkup_history は EMR-197-01 の健診受診履歴 enum で未指定 = フィルタ無し）。
 *
 * EMR-463: 一覧の前に owners 専有 adapter で GET /v1/me の `owners_view_clinic_ids`（現在許可）を
 * 毎回取得し、URL `?clinics=` → 保存選択 → 全許可の順で有効選択を確定してから
 * `clinic_ids`（文字列）を必ず付けて GET /v1/pets を呼ぶ。scope 取得前の先行 list 要求、
 * 取得不能時の所属全件・現在医院・旧 scope への fallback は行わない。
 * GET /v1/pets の最終認可は維持し、403 は scope を 1 回だけ再取得して照合し直す
 * （同じ選択のまま再送しない）。`request.signal` で旧要求を中断する。
 * EMR-463 の経路は OWNERS_CLINIC_SCOPE_ACTIVE が true のときだけ使う（静的 activation。応答を見た fallback ではない）。
 */
export const ownersLoader = async ({
  request,
}: {
  request: Request;
}): Promise<OwnersLoaderData> => {
  if (!OWNERS_CLINIC_SCOPE_ACTIVE) return legacyOwnersLoader(request);
  const { signal } = request;
  const searchParams = new URL(request.url).searchParams;
  const urlClinicIds = parseClinicsParam(searchParams.get("clinics"));
  const page = Math.max(1, Number(searchParams.get("page") ?? 1) || 1);
  const search = searchParams.get("search") || undefined;
  const species = searchParams.get("species") || undefined;
  const includeDeceased = searchParams.get("include_deceased") === "true" ? "true" : undefined;
  // EMR-197-01: checkup_history は列挙値のみ有効 — 検証は backend (400) に委ね、
  // 値はそのまま透過転送する（URL が source of truth）。
  const checkupHistory = searchParams.get("checkup_history") || undefined;

  const fetchPets = async (clinicIds: string[]) => {
    const { data } = await axios.get<PetsResponse>("/v1/pets", {
      params: {
        clinic_ids: clinicIds.join(","),
        page,
        limit: OWNERS_PAGE_SIZE,
        ...(search ? { search } : {}),
        ...(species ? { species } : {}),
        ...(includeDeceased ? { include_deceased: includeDeceased } : {}),
        ...(checkupHistory ? { checkup_history: checkupHistory } : {}),
      },
      signal,
    });
    return data;
  };

  const toLoaderData = (
    result: PetsResponse,
    clinicScope: OwnersClinicScopeView,
    scopeRechecked: boolean,
  ): OwnersLoaderData => ({
    pets: result.data.map(transformPetListItemToFrontend),
    page: result.page,
    limit: result.limit,
    total: result.total,
    listState: { kind: "ready", scopeRechecked },
    clinicScope,
  });

  // 非 ready の決定（許可ゼロ・選択超過）は一覧 API を呼ばずに状態として返す。
  const settleNonReady = (
    scope: OwnersViewScope,
    decision: Exclude<OwnersClinicSelectionDecision, { kind: "ready" }>,
  ): OwnersLoaderData =>
    decision.kind === "no-permitted-clinics"
      ? emptyOwnersData({ kind: "no-permitted-clinics" }, toScopeView(scope, []))
      : emptyOwnersData(
          {
            kind: "selection-over-limit",
            selectedCount: decision.selectedClinicIds.length,
            limit: decision.limit,
          },
          toScopeView(scope, decision.selectedClinicIds),
        );

  const first = await resolveScopedSelection(urlClinicIds, signal);
  if (first.kind === "scope-error") return emptyOwnersData({ kind: "scope-error" }, null);
  if (first.decision.kind !== "ready") return settleNonReady(first.scope, first.decision);
  const firstSelection = first.decision.selectedClinicIds;

  let firstError: unknown;
  try {
    const result = await fetchPets(firstSelection);
    return toLoaderData(result, toScopeView(first.scope, firstSelection), false);
  } catch (err) {
    if (signal.aborted) throw err;
    firstError = err;
  }

  if (httpStatusOf(firstError) === 403) {
    // 最終認可の拒否: 現在許可を 1 回だけ再取得し、選択を照合し直す。
    const recheck = await resolveScopedSelection(urlClinicIds, signal);
    if (recheck.kind === "scope-error") return emptyOwnersData({ kind: "scope-error" }, null);
    if (recheck.decision.kind !== "ready") return settleNonReady(recheck.scope, recheck.decision);
    if (!isSameClinicSet(recheck.decision.selectedClinicIds, firstSelection)) {
      const selection = recheck.decision.selectedClinicIds;
      try {
        const result = await fetchPets(selection);
        return toLoaderData(result, toScopeView(recheck.scope, selection), true);
      } catch (err) {
        if (signal.aborted) throw err;
        throw new Response("飼主一覧の取得に失敗しました", { status: httpStatusOf(err) });
      }
    }
    // 照合し直しても同じ選択なら、拒否された選択で再送しない。
  }

  // 上流のHTTPステータスを保全する。500 へ潰すと 400（リクエスト不正・スキーマ不整合）や
  // 403（権限不足）がすべて「サーバ内部エラー」に見え、原因の切り分けが不可能になる。
  // status を持たない通信エラー（ネットワーク断）と非 axios 例外のみ 500 とする。
  throw new Response("飼主一覧の取得に失敗しました", { status: httpStatusOf(firstError) });
};

export interface OwnerLoaderData {
  owner: Owner;
}

export const ownerLoader = async ({
  params,
}: {
  params: Record<string, string | undefined>;
}): Promise<OwnerLoaderData> => {
  const id = params.id;
  if (!id) {
    throw new Response("Owner ID is required", { status: 400 });
  }
  try {
    const owner = await getOwner(id);
    // PERF-E5-N1-PETS: owner.pets の各 id を個別 GET する 1+N を廃止し、
    // owner_id スコープの一覧リクエストに集約する。include_deceased=true は
    // 旧挙動（owner.pets preload が死亡個体も含む）の保持。
    // 失敗も per-pet ではなく1回だけ表面化する。
    // BUG-010 回帰防止: clinic_ids 未指定だと選択中医院のみに絞られ、他医院の飼主を
    // 開いたときペット一覧が空になる。飼主取得成功 = その医院は認可済みなので、
    // 旧 per-pet detail 経路（全認可医院で取得できた）と同じ結果になるよう
    // owner.clinicId を clinic_ids として渡す。
    const baseParams = {
      owner_id: id,
      include_deceased: "true",
      limit: OWNER_PETS_MAX_LIMIT,
      ...(owner.clinicId ? { clinic_ids: owner.clinicId } : {}),
    };
    const { data: firstPage } = await axios.get<PetsResponse>("/v1/pets", {
      params: { ...baseParams, page: 1 },
    });
    const rows = [...firstPage.data];
    // limit(=100) を超えるペットを持つ飼主は残ページを順次取得して順序を保ったまま結合する
    // （通常は total<=limit で1リクエストのまま）。空ページ返却は total 不整合時の
    // 無限ループ防止のため打ち切る。
    for (let page = 2; rows.length < firstPage.total; page += 1) {
      const { data: next } = await axios.get<PetsResponse>("/v1/pets", {
        params: { ...baseParams, page },
      });
      if (next.data.length === 0) {
        break;
      }
      rows.push(...next.data);
    }
    const pets = rows.map(transformPetListItemToFrontend);
    return { owner: { ...owner, pets } };
  } catch (err) {
    // BUG-010: 他医院作成直後に旧 X-Clinic-ID で GET すると 404。汎用クラッシュではなく明示メッセージにする。
    const status = isAxiosError(err) ? (err.response?.status ?? 500) : 500;
    if (status === 404) {
      throw new Response(
        "飼主が見つかりません。選択中の医院と異なる医院に登録されている可能性があります。医院を切り替えて再度お試しください。",
        { status: 404 },
      );
    }
    throw new Response("飼主情報の取得に失敗しました", { status });
  }
};
