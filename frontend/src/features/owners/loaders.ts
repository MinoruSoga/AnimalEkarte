import { isAxiosError } from "axios";
import { getOwner } from "./api/get-owner";
// `axios` は Axios.create() 由来の instance で isAxiosError static を持たないため、
// 型ガードは axios package の named export を使う（features/auth 等と同じ流儀）。
import { axios } from "@/lib/axios";
import {
  PET_GENDER_MAP,
  ACQUISITION_TYPE_MAP,
  DANGER_LEVEL_MAP,
  mapPetStatusLabel,
  transformBackendPetToFrontend,
} from "@/lib/transforms/pet";
import type { Pet } from "@/types";
import type { PetResponse } from "@/types/generated/pet-responses";
import type { Owner } from "@/types/owner";

// #266: GET /v1/pets 専用の list DTO は PetResponse（詳細）より薄い。
// PetOwnerNested 等は pet-responses と一致するが、list 専用の欠落フィールドがあるため
// ローカル型を維持する。
interface PetListOwnerNested {
  id: number;
  owner_number: number;
  name: string;
  name_kana: string;
  phone: string;
  is_dangerous: boolean;
}

interface PetListAnimalSpeciesNested {
  id: number;
  name: string;
  sort_order: number;
}

interface PetListInsuranceNested {
  id: number;
  name: string;
  coverage_rate: number;
  contact_phone: string;
}

interface PetListApiItem {
  id: number;
  clinic_id: number;
  owner_id: number;
  animal_species_id: number;
  pet_number: string;
  name: string;
  pet_name_kana: string;
  gender: string;
  status: string;
  birth_date?: string;
  breed: string;
  color: string;
  blood_type?: string;
  microchip_number?: string;
  weight?: number;
  neutered_date?: string;
  acquisition_type?: string;
  danger_level: string;
  danger_reason?: string;
  food: string;
  environment: string;
  last_visit?: string;
  insurance_id?: number;
  remarks: string;
  owner?: PetListOwnerNested;
  animal_species?: PetListAnimalSpeciesNested;
  insurance?: PetListInsuranceNested;
}

interface PetsResponse {
  data: PetListApiItem[];
  total: number;
  page: number;
  limit: number;
}

// #266: 飼主・ペット一覧のページサイズ。以前の client-side usePagination のデフォルト(20件)を踏襲。
// ページ粒度はペット行単位（旧: 飼主単位のフラット化）— 1ページの行数は常に一定になる。
const OWNERS_PAGE_SIZE = 20;

export interface OwnersLoaderData {
  pets: Pet[];
  page: number;
  limit: number;
  total: number;
}

// pets 一覧レスポンス → フロントエンド Pet 型変換。
// transformBackendPetToFrontend (lib/transforms/pet.ts) と同じフィールド構成を維持するが、
// petListResponse の薄い owner サマリ（address1/2 等を持たない）に合わせて address は常に undefined。
function transformPetListItemToFrontend(p: PetListApiItem): Pet {
  return {
    id: String(p.id),
    clinicId: p.clinic_id != null ? String(p.clinic_id) : undefined,
    ownerId: String(p.owner_id),
    ownerNumber: p.owner?.owner_number,
    ownerName: p.owner?.name ?? "",
    ownerNameKana: p.owner?.name_kana ?? undefined,
    address: undefined,
    phone: p.owner?.phone ?? "",
    petNumber: p.pet_number,
    name: p.name ?? "",
    petNameKana: p.pet_name_kana ?? undefined,
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
    weight: p.weight?.toString(),
    food: p.food,
    environment: p.environment,
    acquisitionType: p.acquisition_type ? (ACQUISITION_TYPE_MAP[p.acquisition_type] ?? p.acquisition_type) : undefined,
    dangerLevel: p.danger_level ? (DANGER_LEVEL_MAP[p.danger_level] ?? p.danger_level) : undefined,
    dangerReason: p.danger_reason,
    lastVisit: p.last_visit ? p.last_visit.split("T")[0] : undefined,
    insuranceId: p.insurance_id != null ? String(p.insurance_id) : undefined,
    insuranceName: p.insurance?.name,
    insuranceDetails: p.insurance?.coverage_rate != null ? `${p.insurance.coverage_rate}%補償` : undefined,
    remarks: p.remarks,
    // petListResponse は deceased_at を含まない（status で十分 — StatusBadge は pet.status を使用）。
    deceasedAt: undefined,
  };
}

/**
 * 飼主・ペット一覧ローダー — #266: GET /v1/pets をペット行粒度でサーバサイドページネーション取得する
 * （owners-pets-list-plan.md の PO 決定: owners API+EXISTS 応急案ではなく pets API 拡張を正本とする）。
 * URL の page/search/species/include_deceased をそのまま backend に転送する
 * （species は animal_species_id 数値、include_deceased 未指定 = 生存のみが既定）。
 * #86: URL の ?clinics=1,2 を API の clinic_ids に引き渡し拠点横断取得する
 * （所属検証はサーバ側 resolveListClinicIDs が行う。未指定は現在の医院のみ）。
 */
export const ownersLoader = async ({ request }: { request: Request }): Promise<OwnersLoaderData> => {
  try {
    const searchParams = new URL(request.url).searchParams;
    const clinics = searchParams.get("clinics") ?? undefined;
    const page = Math.max(1, Number(searchParams.get("page") ?? 1) || 1);
    const search = searchParams.get("search") || undefined;
    const species = searchParams.get("species") || undefined;
    const includeDeceased = searchParams.get("include_deceased") === "true" ? "true" : undefined;

    const { data: result } = await axios.get<PetsResponse>("/v1/pets", {
      params: {
        ...(clinics ? { clinic_ids: clinics } : {}),
        page,
        limit: OWNERS_PAGE_SIZE,
        ...(search ? { search } : {}),
        ...(species ? { species } : {}),
        ...(includeDeceased ? { include_deceased: includeDeceased } : {}),
      },
    });

    return {
      pets: result.data.map(transformPetListItemToFrontend),
      page: result.page,
      limit: result.limit,
      total: result.total,
    };
  } catch (err) {
    // 上流のHTTPステータスを保全する。500 へ潰すと 400（リクエスト不正・スキーマ不整合）や
    // 403（権限不足）がすべて「サーバ内部エラー」に見え、原因の切り分けが不可能になる。
    // status を持たない通信エラー（ネットワーク断）と非 axios 例外のみ 500 とする。
    const status = isAxiosError(err) ? (err.response?.status ?? 500) : 500;
    throw new Response("飼主一覧の取得に失敗しました", { status });
  }
};

export interface OwnerLoaderData {
  owner: Owner;
}

export const ownerLoader = async ({ params }: { params: Record<string, string | undefined> }): Promise<OwnerLoaderData> => {
  const id = params.id;
  if (!id) {
    throw new Response("Owner ID is required", { status: 400 });
  }
  try {
    const owner = await getOwner(id);
    const pets = await Promise.all(
      (owner.pets ?? []).map(async (pet) => {
        const { data } = await axios.get<PetResponse>(`/v1/pets/${pet.id}`);
        return transformBackendPetToFrontend(data);
      }),
    );
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
