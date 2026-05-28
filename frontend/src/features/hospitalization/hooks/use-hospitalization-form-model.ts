import type { Pet, TreatmentPlan } from "@/types";

import type {
  CreateHospitalizationRequest,
  UpdateHospitalizationRequest,
  BackendHospitalization,
} from "../api/types";
import type { HospitalizationFormData } from "../types";

const MS_PER_DAY = 86_400_000;
export const DEFAULT_HOSPITALIZATION_DAYS = 7;

export const DEFAULT_TREATMENT_PLANS: TreatmentPlan[] = [
  {
    id: "1",
    treatmentContent: "adm rate",
    memo: "入院料1日分",
    is_insurance: true,
    unitPrice: 990,
    quantity: 1,
    discount: 0,
    discountAmount: 0,
    subtotal: 990,
  },
  {
    id: "2",
    treatmentContent: "PCG/SC ~15kg",
    memo: "",
    is_insurance: false,
    unitPrice: 990,
    quantity: 1,
    discount: 0,
    discountAmount: 0,
    subtotal: 990,
  },
];

export function getDefaultHospitalizationEndDate() {
  return new Date(Date.now() + DEFAULT_HOSPITALIZATION_DAYS * MS_PER_DAY).toISOString().split("T")[0];
}

export function createInitialHospitalizationFormData(): HospitalizationFormData {
  return {
    hospitalizationType: "入院",
    ownerName: "",
    species: "",
    petName: "",
    petNumber: "",
    petInsurance: "",
    petDetails: "",
    visit: "診療",
    nextVisit: "",
    weight: "",
    displayDate: "",
    endDate: getDefaultHospitalizationEndDate(),
    memo: "",
    ownerRequest: "",
    staffNotes: "",
    cageId: "",
    isInsurance: false,
    insuranceCompanyName: "",
    insuranceNumber: "",
  };
}

function toHospitalizationType(value: string) {
  return value === "入院" ? "hospitalization" : "hotel";
}

function toInsuranceCompanyName(formData: HospitalizationFormData) {
  return formData.isInsurance ? (formData.insuranceCompanyName || null) : null;
}

function toInsuranceNumber(formData: HospitalizationFormData) {
  return formData.isInsurance ? (formData.insuranceNumber || null) : null;
}

export function buildUpdateHospitalizationRequest(
  formData: HospitalizationFormData
): UpdateHospitalizationRequest {
  return {
    hospitalization_type: toHospitalizationType(formData.hospitalizationType),
    owner_request: formData.ownerRequest,
    staff_notes: formData.staffNotes,
    memo: formData.memo,
    cage_id: formData.cageId || undefined,
    is_insurance: formData.isInsurance,
    insurance_company_name: toInsuranceCompanyName(formData),
    insurance_number: toInsuranceNumber(formData),
  };
}

export function buildCreateHospitalizationRequest(
  formData: HospitalizationFormData,
  pet: Pet
): CreateHospitalizationRequest {
  const today = new Date().toISOString().split("T")[0];
  const startISO = `${formData.displayDate || today}T00:00:00Z`;
  const endISO = `${formData.endDate || getDefaultHospitalizationEndDate()}T00:00:00Z`;

  return {
    pet_id: pet.id,
    owner_id: pet.ownerId || "",
    hospitalization_type: toHospitalizationType(formData.hospitalizationType),
    start_date: startISO,
    end_date: endISO,
    owner_request: formData.ownerRequest,
    staff_notes: formData.staffNotes,
    memo: formData.memo,
    cage_id: formData.cageId || undefined,
    is_insurance: formData.isInsurance,
    insurance_company_name: toInsuranceCompanyName(formData),
    insurance_number: toInsuranceNumber(formData),
  };
}

export function buildHospitalizationFormDataFromRecord(
  currentFormData: HospitalizationFormData,
  hospitalization: BackendHospitalization
): HospitalizationFormData {
  return {
    ...currentFormData,
    hospitalizationType: hospitalization.hospitalization_type === "hospitalization" ? "入院" : "ホテル",
    cageId: hospitalization.cage_id ? String(hospitalization.cage_id) : "",
    displayDate: hospitalization.start_date,
    endDate: hospitalization.end_date ? hospitalization.end_date.split("T")[0] : getDefaultHospitalizationEndDate(),
    memo: hospitalization.memo ?? "",
    ownerRequest: hospitalization.owner_request ?? "",
    staffNotes: hospitalization.staff_notes ?? "",
    isInsurance: !!(hospitalization.insurance_company_name || hospitalization.insurance_number),
    insuranceCompanyName: hospitalization.insurance_company_name ?? "",
    insuranceNumber: hospitalization.insurance_number ?? "",
  };
}

export function buildSelectedPetFromHospitalization(hospitalization: BackendHospitalization): Pet | null {
  if (!hospitalization.pet || !hospitalization.owner_id) {
    return null;
  }

  return {
    id: String(hospitalization.pet_id),
    ownerId: String(hospitalization.owner_id),
    ownerName: hospitalization.owner?.name ?? "",
    name: hospitalization.pet.name,
    species: hospitalization.pet.animal_species?.name ?? "",
    breed: hospitalization.pet.breed,
    gender: hospitalization.pet.gender,
  } as Pet;
}

export function mergePetIntoHospitalizationFormData(
  formData: HospitalizationFormData,
  selectedPets: readonly Pet[]
) {
  if (selectedPets.length === 0) {
    return formData;
  }

  const selectedPet = selectedPets[0];
  return {
    ...formData,
    ownerName: selectedPet.ownerName,
    petName: selectedPet.name,
    petNumber: selectedPet.id,
    species: selectedPet.species,
    weight: selectedPet.weight ? `${selectedPet.weight}kg` : "",
  };
}

export function createEmptyTreatmentPlan(): TreatmentPlan {
  return {
    id: crypto.randomUUID(),
    treatmentContent: "",
    memo: "",
    is_insurance: false,
    unitPrice: 0,
    quantity: 1,
    discount: 0,
    discountAmount: 0,
    subtotal: 0,
  };
}

export function updateTreatmentPlanField(
  plan: TreatmentPlan,
  field: keyof TreatmentPlan,
  value: string | number | boolean
) {
  const updated = { ...plan, [field]: value };
  if (field === "unitPrice" || field === "quantity" || field === "discount") {
    const unitPrice = (field === "unitPrice" ? value : plan.unitPrice) as number;
    const quantity = (field === "quantity" ? value : plan.quantity) as number;
    const discount = (field === "discount" ? value : plan.discount) as number;
    const baseAmount = unitPrice * quantity;
    updated.discountAmount = Math.floor(baseAmount * (discount / 100));
    updated.subtotal = baseAmount - updated.discountAmount;
  }
  return updated;
}
