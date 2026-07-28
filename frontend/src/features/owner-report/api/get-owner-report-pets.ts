import { useQuery } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_GC_TIMES, QUERY_STALE_TIMES } from "@/lib/react-query";

interface OwnerReportPetWire {
  id: number;
  name: string;
  pet_name_kana: string;
  gender: string;
  status: string;
  birth_date?: string | null;
  breed: string;
  color: string;
  blood_type?: string | null;
  microchip_number?: string | null;
  weight?: number | null;
  neutered_date?: string | null;
  acquisition_type?: string | null;
  food: string;
  environment: string;
  last_visit?: string | null;
  remarks: string;
  animal_species?: {
    name: string;
  } | null;
  insurance?: {
    name: string;
    coverage_rate: number;
  } | null;
}

interface OwnerReportPetsResponse {
  data: OwnerReportPetWire[];
}

export interface OwnerReportPet {
  id: string;
  name: string;
  petNameKana: string;
  gender: string;
  status: string;
  birthDate?: string;
  breed: string;
  color: string;
  bloodType?: string;
  microchipNumber?: string;
  weight?: string;
  neuteredDate?: string;
  acquisitionType?: string;
  food: string;
  environment: string;
  lastVisit?: string;
  remarks: string;
  species: string;
  insuranceName?: string;
  insuranceDetails?: string;
}

const GENDER_LABELS: Readonly<Record<string, string>> = {
  male: "雄",
  female: "雌",
  unknown: "不明",
};

const STATUS_LABELS: Readonly<Record<string, string>> = {
  alive: "生存",
  deceased: "死亡",
};

const ACQUISITION_TYPE_LABELS: Readonly<Record<string, string>> = {
  purchased: "購入",
  transferred: "譲渡",
  rescued: "保護",
  other: "その他",
};

function dateOnly(value: string | null | undefined): string | undefined {
  return value?.split("T")[0];
}

function mapOwnerReportPet(pet: OwnerReportPetWire): OwnerReportPet {
  return {
    id: String(pet.id),
    name: pet.name,
    petNameKana: pet.pet_name_kana,
    gender: GENDER_LABELS[pet.gender] ?? pet.gender,
    status: STATUS_LABELS[pet.status] ?? pet.status,
    birthDate: dateOnly(pet.birth_date),
    breed: pet.breed,
    color: pet.color,
    bloodType: pet.blood_type ?? undefined,
    microchipNumber: pet.microchip_number ?? undefined,
    weight: pet.weight?.toString(),
    neuteredDate: dateOnly(pet.neutered_date),
    acquisitionType: pet.acquisition_type
      ? (ACQUISITION_TYPE_LABELS[pet.acquisition_type] ??
        pet.acquisition_type)
      : undefined,
    food: pet.food,
    environment: pet.environment,
    lastVisit: dateOnly(pet.last_visit),
    remarks: pet.remarks,
    species: pet.animal_species?.name ?? "",
    insuranceName: pet.insurance?.name,
    insuranceDetails:
      pet.insurance?.coverage_rate != null
        ? `${pet.insurance.coverage_rate}%補償`
        : undefined,
  };
}

export function useGetOwnerReportPets(ownerId: string) {
  return useQuery({
    queryKey: queryKeys.ownerReportPets(ownerId),
    queryFn: async (): Promise<OwnerReportPet[]> => {
      const { data } = await axios.get<OwnerReportPetsResponse>(
        `/v1/owners/${ownerId}/report/pets`,
      );
      return data.data.map(mapOwnerReportPet);
    },
    enabled: !!ownerId,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
