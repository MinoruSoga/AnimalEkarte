import {
  formatJSTWallDate,
  jstDateStartISOString,
  todayJSTISO,
  toJSTWallDate,
} from "@/lib/jst-date";
import { calculateBillingTotals } from "@/lib/calculations";
import type { Pet, HospitalizationTreatmentPlan } from "@/types";
import type { TreatmentPlanResponse } from "@/types/generated/hospitalization-responses";

import type {
  CreateHospitalizationRequest,
  UpdateHospitalizationRequest,
  BackendHospitalization,
} from "../api/types";
import type { HospitalizationFormData } from "../types";

const DEFAULT_HOSPITALIZATION_DAYS = 7;

function getDefaultHospitalizationEndDate() {
  const endDate = toJSTWallDate(new Date());
  endDate.setDate(endDate.getDate() + DEFAULT_HOSPITALIZATION_DAYS);
  return formatJSTWallDate(endDate);
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
    doctorId: "",
    doctorName: "",
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
  return formData.isInsurance ? formData.insuranceCompanyName || null : null;
}

function toInsuranceNumber(formData: HospitalizationFormData) {
  return formData.isInsurance ? formData.insuranceNumber || null : null;
}

export function buildUpdateHospitalizationRequest(
  formData: HospitalizationFormData,
): UpdateHospitalizationRequest {
  return {
    hospitalization_type: toHospitalizationType(formData.hospitalizationType),
    owner_request: formData.ownerRequest,
    staff_notes: formData.staffNotes,
    memo: formData.memo,
    ...(formData.doctorId ? { doctor_id: formData.doctorId } : {}),
    ...(formData.cageId ? { cage_id: formData.cageId } : {}),
    is_insurance: formData.isInsurance,
    insurance_company_name: toInsuranceCompanyName(formData),
    insurance_number: toInsuranceNumber(formData),
  };
}

export function buildCreateHospitalizationRequest(
  formData: HospitalizationFormData,
  pet: Pet,
  treatmentPlans: readonly HospitalizationTreatmentPlan[] = [],
): CreateHospitalizationRequest {
  const today = todayJSTISO();
  const startISO = jstDateStartISOString(formData.displayDate || today);
  const endISO = jstDateStartISOString(formData.endDate || getDefaultHospitalizationEndDate());
  const nestedPlans = buildPersistableTreatmentPlanRequests(treatmentPlans);

  return {
    pet_id: pet.id,
    owner_id: pet.ownerId || "",
    hospitalization_type: toHospitalizationType(formData.hospitalizationType),
    start_date: startISO,
    end_date: endISO,
    owner_request: formData.ownerRequest,
    staff_notes: formData.staffNotes,
    memo: formData.memo,
    ...(formData.doctorId ? { doctor_id: formData.doctorId } : {}),
    ...(formData.cageId ? { cage_id: formData.cageId } : {}),
    is_insurance: formData.isInsurance,
    insurance_company_name: toInsuranceCompanyName(formData),
    insurance_number: toInsuranceNumber(formData),
    ...(nestedPlans.length > 0 ? { treatment_plans: nestedPlans } : {}),
  };
}

export function buildHospitalizationFormDataFromRecord(
  currentFormData: HospitalizationFormData,
  hospitalization: BackendHospitalization,
): HospitalizationFormData {
  return {
    ...currentFormData,
    hospitalizationType:
      hospitalization.hospitalization_type === "hospitalization" ? "入院" : "ホテル",
    cageId: hospitalization.cage_id ? String(hospitalization.cage_id) : "",
    displayDate: hospitalization.start_date,
    endDate: hospitalization.end_date
      ? hospitalization.end_date.split("T")[0]
      : getDefaultHospitalizationEndDate(),
    memo: hospitalization.memo ?? "",
    ownerRequest: hospitalization.owner_request ?? "",
    staffNotes: hospitalization.staff_notes ?? "",
    doctorId: hospitalization.doctor_id ? String(hospitalization.doctor_id) : "",
    doctorName: hospitalization.doctor?.name ?? "",
    isInsurance: !!(hospitalization.insurance_company_name || hospitalization.insurance_number),
    insuranceCompanyName: hospitalization.insurance_company_name ?? "",
    insuranceNumber: hospitalization.insurance_number ?? "",
  };
}

export function buildSelectedPetFromHospitalization(
  hospitalization: BackendHospitalization,
): Pet | null {
  if (!hospitalization.pet || !hospitalization.owner_id) {
    return null;
  }

  const gender =
    hospitalization.pet.gender === "male"
      ? "雄"
      : hospitalization.pet.gender === "female"
        ? "雌"
        : "不明";
  return {
    id: String(hospitalization.pet_id),
    ownerId: String(hospitalization.owner_id),
    ownerName: hospitalization.owner?.name ?? "",
    name: hospitalization.pet.name,
    species: hospitalization.pet.animal_species?.name ?? "",
    breed: hospitalization.pet.breed,
    status: hospitalization.pet.status === "deceased" ? "死亡" : "生存",
    gender,
    birthDate: hospitalization.pet.birth_date ?? "",
    neuteredDate: hospitalization.pet.neutered_date ?? "",
    weight: hospitalization.pet.weight != null ? String(hospitalization.pet.weight) : "",
  } as Pet;
}

export function mergePetIntoHospitalizationFormData(
  formData: HospitalizationFormData,
  selectedPets: readonly Pet[],
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

/**
 * Map GET /hospitalizations/:id/treatment-plans wire rows to edit-form UI shape.
 * Does NOT read hospitalization.treatment_plans (absent on HospitalizationResponse wire).
 */
export function buildTreatmentPlansFromRecord(
  plans: readonly TreatmentPlanResponse[],
): HospitalizationTreatmentPlan[] {
  return plans.map((plan) => ({
    id: String(plan.id),
    treatmentContent: plan.treatment_content,
    memo: plan.memo,
    is_insurance: plan.is_insurance,
    unitPrice: plan.unit_price,
    quantity: plan.quantity,
    discount: plan.discount_rate,
    discountAmount: plan.discount_amount,
    subtotal: plan.subtotal,
  }));
}

export function createEmptyTreatmentPlan(): HospitalizationTreatmentPlan {
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

/**
 * UI treatment plan → POST /hospitalizations/:id/treatment-plans body.
 * Empty treatmentContent rows are not persistable (BE requires content).
 */
function buildCreateTreatmentPlanRequest(
  plan: HospitalizationTreatmentPlan,
  sortOrder: number,
): {
  treatment_content: string;
  memo: string;
  is_insurance: boolean;
  unit_price: number;
  quantity: number;
  discount_rate: number;
  discount_amount: number;
  sort_order: number;
} | null {
  const content = plan.treatmentContent.trim();
  if (!content) return null;
  const quantity = plan.quantity > 0 ? plan.quantity : 1;
  return {
    treatment_content: content,
    memo: plan.memo ?? "",
    is_insurance: plan.is_insurance,
    unit_price: plan.unitPrice,
    quantity,
    discount_rate: plan.discount,
    discount_amount: plan.discountAmount,
    sort_order: sortOrder,
  };
}

/** Plans with non-empty content, mapped to create wire bodies in display order. */
export function buildPersistableTreatmentPlanRequests(
  plans: readonly HospitalizationTreatmentPlan[],
) {
  const bodies: NonNullable<ReturnType<typeof buildCreateTreatmentPlanRequest>>[] = [];
  let sortOrder = 0;
  for (const plan of plans) {
    const body = buildCreateTreatmentPlanRequest(plan, sortOrder);
    if (body) {
      bodies.push(body);
      sortOrder += 1;
    }
  }
  return bodies;
}

export function updateTreatmentPlanField(
  plan: HospitalizationTreatmentPlan,
  field: keyof HospitalizationTreatmentPlan,
  value: string | number | boolean,
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

export function hospitalizationSubmitFieldErrors(
  pet: Pet | undefined,
  isEdit: boolean,
  cageId: string | undefined,
): Record<string, string> | null {
  if (!pet) {
    return { pet: "ペットを選択してください" };
  }
  if (pet.status === "死亡") {
    return {
      pet: isEdit
        ? "死亡したペットは入院情報を更新できません"
        : "死亡したペットは入院登録できません",
    };
  }
  if (!(cageId?.trim() ?? "")) {
    return { cage_id: "ケージ・個室を選択してください" };
  }
  return null;
}

export function calculateHospitalizationBillingTotals(
  treatmentPlans: readonly HospitalizationTreatmentPlan[],
) {
  const billingItems = treatmentPlans.map((plan) => ({
    ...plan,
    isInsuranceApplicable: plan.is_insurance,
  }));
  const result = calculateBillingTotals(billingItems, 0, 0);
  return {
    subtotalBeforeDiscount: result.subtotal,
    discountAmount: result.globalDiscountAmount,
    subtotalAfterDiscount: result.taxableAmount,
    consumptionTax: result.tax,
    total: result.total,
  };
}
