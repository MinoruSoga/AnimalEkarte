import type {
  ApiDailyRecord,
  BackendHospitalization,
  CarePlanItem,
} from "@/features/hospitalization";
import type { BackendMedicalRecord } from "@/features/medical-records";
import type { BackendVaccination } from "@/features/vaccinations";
import type {
  MeClinicInfo,
  MeResponse,
  ResourcePermission,
} from "../../src/types/generated/auth-responses";
import type {
  AnimalSpecies,
  Cage,
  Owner,
  Pet,
  Reservation,
  ReservationType,
  Staff,
  TreatmentPlan,
  TrimmingCourse,
  TrimmingCourseType,
  TrimmingOption,
  Vaccine,
} from "../../src/types/generated/models";
import type { BackendTrimming } from "../../src/types/trimming";
import type { SyntheticEndpoint } from "../helpers/synthetic-api";
import type { SyntheticRenderedAssertion } from "../helpers/ui-design-audit";
import { SYNTHETIC_CREATED_AT as CREATED_AT, SYNTHETIC_IDS } from "./ui-design-clinical-constants";
import {
  validateMedicalRecordCreate,
  validateReservationCreate,
} from "./ui-design-clinical-request-contracts";

export { SYNTHETIC_IDS } from "./ui-design-clinical-constants";

const SYNTHETIC_ME_CLINIC = {
  id: String(SYNTHETIC_IDS.clinic),
  name: "合成監査医院",
  postal_code: "",
  address: "",
  phone_number: "",
  fax_number: "",
  registration_number: "",
  director_name: "合成監査院長",
  email: "ui-audit-clinic@example.invalid",
  website: "https://example.invalid",
  standard_tax_rate: 0.1,
  reduced_tax_rate: 0.08,
  accounting_document_show_logo: false,
  accounting_document_show_registration_warning: false,
  accounting_document_show_item_category: true,
  accounting_document_footer_note: "synthetic fixture only",
  accounting_document_show_clinic_header: true,
  accounting_document_show_owner_pet_info: true,
  accounting_document_show_items_table: true,
  accounting_document_show_payment_summary: true,
  accounting_document_section_order: ["clinic", "owner-pet", "items", "payment"],
} satisfies MeClinicInfo;

export function createSyntheticMeResponse(
  permissions: Readonly<Record<string, ResourcePermission>>,
  isSystemAdmin = false,
): MeResponse {
  return {
    id: "synthetic-staff-990005",
    email: "ui-audit-staff@example.invalid",
    display_name: "合成監査スタッフ",
    is_system_admin: isSystemAdmin,
    occupation: "合成監査職種",
    main_clinic_id: String(SYNTHETIC_IDS.clinic),
    clinic: SYNTHETIC_ME_CLINIC,
    clinics: [
      {
        clinic_id: String(SYNTHETIC_IDS.clinic),
        clinic_name: SYNTHETIC_ME_CLINIC.name,
        is_main: true,
      },
    ],
    permissions: { ...permissions },
  };
}

const SYNTHETIC_ADMIN_ME = createSyntheticMeResponse({}, true);

const SYNTHETIC_CARE_PLAN_ITEM = {
  id: String(SYNTHETIC_IDS.carePlanItem),
  hospitalization_id: String(SYNTHETIC_IDS.hospitalization),
  type: "treatment",
  name: "合成監査処置",
  description: "synthetic fixture only",
  timing: ["morning"],
  status: "active",
  notes: "synthetic fixture only",
  procedure_id: String(SYNTHETIC_IDS.carePlanItem),
  unit_price: 4_321,
  category: "synthetic",
  sort_order: 1,
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
} satisfies CarePlanItem;

const SYNTHETIC_TREATMENT_PLAN = {
  id: SYNTHETIC_IDS.treatmentPlan,
  clinic_id: SYNTHETIC_IDS.clinic,
  hospitalization_id: SYNTHETIC_IDS.hospitalization,
  treatment_content: "合成監査輸液",
  memo: "既存record由来",
  is_insurance: true,
  unit_price: 3_210,
  quantity: 2,
  discount_rate: 10,
  discount_amount: 642,
  subtotal: 5_778,
  sort_order: 1,
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
} satisfies TreatmentPlan;

const SYNTHETIC_OWNER = {
  id: SYNTHETIC_IDS.owner,
  clinic_id: SYNTHETIC_IDS.clinic,
  name: "合成監査飼主",
  name_kana: "ゴウセイカンサイカイヌシ",
  company: "",
  postal_code: "",
  address1: "",
  address2: "",
  home_postal_code: "",
  home_address1: "",
  home_address2: "",
  phone: "",
  company_phone: "",
  email: "ui-audit-owner@example.invalid",
  remarks: "synthetic fixture only",
  is_dangerous: false,
  discount_rate: 0,
  membership_type: "non_member",
  lstep_opt_out: false,
  delivery_excluded: false,
  delivery_caution: false,
  is_transferred: false,
  dm_preference: undefined,
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
} satisfies Owner;

const SYNTHETIC_SPECIES = {
  id: SYNTHETIC_IDS.species,
  name: "合成動物種",
  is_active: true,
  sort_order: 1,
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
} satisfies AnimalSpecies;

export const SYNTHETIC_PET = {
  id: SYNTHETIC_IDS.pet,
  clinic_id: SYNTHETIC_IDS.clinic,
  owner_id: SYNTHETIC_IDS.owner,
  animal_species_id: SYNTHETIC_IDS.species,
  pet_number: "SYN-PET-990003",
  name: "合成監査ペット",
  name_kana: "ゴウセイカンサイペット",
  gender: "unknown",
  status: "alive",
  breed: "合成品種",
  color: "",
  weight: 4.2,
  danger_level: "low",
  food: "",
  environment: "",
  phone: "",
  remarks: "synthetic fixture only",
  version: 1,
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
  owner: SYNTHETIC_OWNER,
  animal_species: SYNTHETIC_SPECIES,
} satisfies Pet;

const SYNTHETIC_STAFF = {
  id: SYNTHETIC_IDS.staff,
  clinic_id: SYNTHETIC_IDS.clinic,
  name: "合成監査スタッフ",
  is_active: true,
  license_number: "SYNTHETIC",
  sort_order: 1,
  staff_type: "doctor",
  reservation_display_name: "合成監査スタッフ",
  reservation_visible: true,
  reservation_comment: "",
  reservation_image_url: "",
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
} satisfies Staff;

const SYNTHETIC_CAGE = {
  id: SYNTHETIC_IDS.cage,
  clinic_id: SYNTHETIC_IDS.clinic,
  name: "合成監査ケージ",
  price: 1000,
  is_active: true,
  description: "synthetic fixture only",
  cage_type: "general",
  cage_size: "medium",
  sort_order: 1,
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
} satisfies Cage;

const SYNTHETIC_RESERVATION_TYPE = {
  id: SYNTHETIC_IDS.reservationType,
  clinic_id: SYNTHETIC_IDS.clinic,
  name: "合成一般診療",
  is_active: true,
  description: "synthetic fixture only",
  color: "#2563eb",
  sort_order: 1,
  reservation_display_name: "合成一般診療",
  duration_minutes: 30,
  short_name: "合成診療",
  show_short_name: true,
  reservation_visible: true,
  reservation_comment: "",
  reservation_image_url: "",
  reservation_day_option: "anyday",
  is_internal: false,
  category: "general",
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
} satisfies ReservationType;

const SYNTHETIC_RESERVATION = {
  id: SYNTHETIC_IDS.reservation,
  clinic_id: SYNTHETIC_IDS.clinic,
  start_time: CREATED_AT,
  end_time: "2026-07-23T00:30:00+09:00",
  owner_id: SYNTHETIC_IDS.owner,
  pet_id: SYNTHETIC_IDS.pet,
  visit_type: "revisit",
  reservation_type_id: SYNTHETIC_IDS.reservationType,
  is_designated: false,
  status: "in_consultation",
  notes: "",
  reservation_route: "record_shortcut",
  source: "manual",
  is_staff_delegated: false,
  customer_fields: {},
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
  owner: SYNTHETIC_OWNER,
  pet: SYNTHETIC_PET,
  reservation_type: SYNTHETIC_RESERVATION_TYPE,
} satisfies Reservation;

const SYNTHETIC_HOSPITALIZATION = {
  id: SYNTHETIC_IDS.hospitalization,
  clinic_id: SYNTHETIC_IDS.clinic,
  owner_id: SYNTHETIC_IDS.owner,
  pet_id: SYNTHETIC_IDS.pet,
  hospitalization_type: "hospitalization",
  start_date: "2026-01-01T00:00:00+09:00",
  end_date: "2099-12-31T00:00:00+09:00",
  status: "admitted",
  cage_id: SYNTHETIC_IDS.cage,
  doctor_id: SYNTHETIC_IDS.staff,
  memo: "合成監査メモ",
  owner_request: "合成監査リクエスト",
  staff_notes: "合成監査スタッフメモ",
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
  owner: SYNTHETIC_OWNER,
  pet: SYNTHETIC_PET,
  doctor: SYNTHETIC_STAFF,
} satisfies BackendHospitalization;

const SYNTHETIC_DAILY_RECORD = {
  id: "syn-daily-990007",
  hospitalization_id: String(SYNTHETIC_IDS.hospitalization),
  date: CREATED_AT,
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
  vital_records: [],
  care_logs: [],
  staff_notes: [],
} satisfies ApiDailyRecord;

const SYNTHETIC_TRIMMING_COURSE_TYPE = {
  id: SYNTHETIC_IDS.trimmingCourseType,
  clinic_id: SYNTHETIC_IDS.clinic,
  name: "合成コース種別",
  sort_order: 1,
  is_active: true,
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
} satisfies TrimmingCourseType;

const SYNTHETIC_TRIMMING_COURSE = {
  id: SYNTHETIC_IDS.trimmingCourse,
  clinic_id: SYNTHETIC_IDS.clinic,
  name: "合成監査コース",
  price: 5000,
  is_active: true,
  description: "synthetic fixture only",
  course_type_id: SYNTHETIC_IDS.trimmingCourseType,
  sort_order: 1,
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
} satisfies TrimmingCourse;

const SYNTHETIC_TRIMMING_OPTION = {
  id: SYNTHETIC_IDS.trimmingOption,
  clinic_id: SYNTHETIC_IDS.clinic,
  name: "合成監査オプション",
  price: 500,
  is_active: true,
  description: "synthetic fixture only",
  is_combinable: true,
  sort_order: 1,
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
} satisfies TrimmingOption;

const SYNTHETIC_TRIMMING = {
  id: SYNTHETIC_IDS.trimming,
  clinic_id: SYNTHETIC_IDS.clinic,
  reservation_type_id: SYNTHETIC_IDS.reservationType,
  start_time: CREATED_AT,
  end_time: "2026-07-23T01:00:00+09:00",
  pet_id: SYNTHETIC_IDS.pet,
  staff_id: SYNTHETIC_IDS.staff,
  status: "in_consultation",
  source: "manual",
  has_detail: true,
  course_id: SYNTHETIC_IDS.trimmingCourse,
  style_request: "合成監査スタイル",
  bw: 4.2,
  bw_unit: "Kg",
  bt: 38.1,
  used_shampoo: "合成シャンプー",
  used_ribbon: "合成リボン",
  remarks: "合成監査備考",
  style_image: "",
  completed_image: "",
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
  course: {
    id: SYNTHETIC_IDS.trimmingCourse,
    name: SYNTHETIC_TRIMMING_COURSE.name,
    price: SYNTHETIC_TRIMMING_COURSE.price,
  },
  options: [{ id: SYNTHETIC_IDS.trimmingOption, name: SYNTHETIC_TRIMMING_OPTION.name }],
  pet: {
    id: SYNTHETIC_IDS.pet,
    name: SYNTHETIC_PET.name,
    pet_number: SYNTHETIC_PET.pet_number,
    weight: SYNTHETIC_PET.weight,
    breed: SYNTHETIC_PET.breed,
    animal_species: { id: SYNTHETIC_IDS.species, name: SYNTHETIC_SPECIES.name },
    owner: { id: SYNTHETIC_IDS.owner, name: SYNTHETIC_OWNER.name },
  },
  staff: { id: SYNTHETIC_IDS.staff, name: SYNTHETIC_STAFF.name },
} satisfies BackendTrimming;

const SYNTHETIC_VACCINE = {
  id: SYNTHETIC_IDS.vaccine,
  clinic_id: SYNTHETIC_IDS.clinic,
  name: "合成監査ワクチン",
  price: 4000,
  is_active: true,
  description: "synthetic fixture only",
  species: "both",
  interval: "1年",
  sort_order: 1,
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
} satisfies Vaccine;

const SYNTHETIC_VACCINATION = {
  id: SYNTHETIC_IDS.vaccination,
  clinic_id: SYNTHETIC_IDS.clinic,
  medical_record_id: SYNTHETIC_IDS.medicalRecord,
  pet_id: SYNTHETIC_IDS.pet,
  vaccine_id: SYNTHETIC_IDS.vaccine,
  date: CREATED_AT,
  doctor_id: SYNTHETIC_IDS.staff,
  next_date: "2027-07-23T00:00:00+09:00",
  next_schedule_type: "1year",
  supplemental: "合成補足",
  lot1: "SYN-LOT-1",
  lot2: "",
  lot3: "",
  lot4: "",
  remarks: "合成監査備考",
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
  pet: SYNTHETIC_PET,
  vaccine: SYNTHETIC_VACCINE,
  doctor: SYNTHETIC_STAFF,
} satisfies BackendVaccination;

const SYNTHETIC_MEDICAL_RECORD = {
  id: SYNTHETIC_IDS.medicalRecord,
  clinic_id: SYNTHETIC_IDS.clinic,
  record_no: "SYN-MR-990013",
  date: CREATED_AT,
  owner_id: SYNTHETIC_IDS.owner,
  pet_id: SYNTHETIC_IDS.pet,
  doctor_id: SYNTHETIC_IDS.staff,
  appointment_id: SYNTHETIC_IDS.reservation,
  status: "draft",
  version: 1,
  visit_type: "revisit",
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
  visit_count: 1,
  owner: SYNTHETIC_OWNER,
  pet: SYNTHETIC_PET,
  doctor: SYNTHETIC_STAFF,
} satisfies BackendMedicalRecord;

const {
  name: syntheticOwnerName,
  name_kana: syntheticOwnerNameKana,
  dm_preference: _syntheticDmPreference,
  ...syntheticOwnerWithoutApiRenames
} = SYNTHETIC_OWNER;

const SYNTHETIC_OWNER_API_RESPONSE = {
  ...syntheticOwnerWithoutApiRenames,
  owner_name: syntheticOwnerName,
  owner_name_kana: syntheticOwnerNameKana,
  dm_preference: null,
} satisfies Omit<Owner, "name" | "name_kana" | "dm_preference"> & {
  owner_name: string;
  owner_name_kana?: string;
  dm_preference?: boolean | null;
};

const SYNTHETIC_CLINICAL_PLAN = {
  id: "syn-plan-990013",
  medical_record_id: String(SYNTHETIC_IDS.medicalRecord),
  physical_exam: "",
  diagnosis_type_id: undefined,
  diagnosis_name_id: undefined,
  diagnosis_2_type_id: undefined,
  diagnosis_2_name_id: undefined,
  diagnosis_details: "",
  treatment_policy: "",
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
  diagnosis_type: null,
  diagnosis_name: null,
  diagnosis_2_type: null,
  diagnosis_2_name: null,
  version: 1,
} satisfies {
  id: string;
  medical_record_id: string;
  physical_exam: string;
  diagnosis_type_id?: string | null;
  diagnosis_name_id?: string | null;
  diagnosis_2_type_id?: string | null;
  diagnosis_2_name_id?: string | null;
  diagnosis_details: string;
  treatment_policy: string;
  created_at: string;
  updated_at: string;
  diagnosis_type?: { id: string; name: string } | null;
  diagnosis_name?: { id: string; name: string } | null;
  diagnosis_2_type?: { id: string; name: string } | null;
  diagnosis_2_name?: { id: string; name: string } | null;
  version: number;
};

const SYNTHETIC_LINE_STATUS = {
  line_user_id: null,
  is_linked: false,
  lstep_opt_out: false,
  tags: [],
  fetched_at: CREATED_AT,
} satisfies {
  line_user_id: string | null;
  is_linked: boolean;
  lstep_opt_out: boolean;
  tags: string[];
  fetched_at: string;
};

export type SyntheticClinicalKey =
  | "medicalRecordCreate"
  | "hospitalizationDetail"
  | "hospitalizationEdit"
  | "trimmingEdit"
  | "vaccinationEdit";

export interface SyntheticClinicalScenario {
  entryPath: string;
  renderedPath: string;
  currentClinicId: string;
  fixedTime: string;
  renderedAssertions: readonly SyntheticRenderedAssertion[];
  endpoints: readonly SyntheticEndpoint[];
  expectedLocalBusinessNonGet: readonly string[];
}

const HOSPITALIZATION_PATH = `/api/v1/hospitalizations/${SYNTHETIC_IDS.hospitalization}`;
const MEDICAL_RECORD_PATH = `/api/v1/medical-records/${SYNTHETIC_IDS.medicalRecord}`;
const SYNTHETIC_ME_ENDPOINT = {
  method: "GET",
  pathname: "/api/v1/me",
  response: SYNTHETIC_ADMIN_ME,
} satisfies SyntheticEndpoint;

function defineEndpoints(endpoints: readonly SyntheticEndpoint[]): readonly SyntheticEndpoint[] {
  return endpoints;
}

export const SYNTHETIC_CLINICAL_SCENARIOS = {
  medicalRecordCreate: {
    entryPath: `/medical-records/new?petId=${SYNTHETIC_IDS.pet}`,
    renderedPath: `/medical-records/${SYNTHETIC_IDS.medicalRecord}`,
    currentClinicId: String(SYNTHETIC_IDS.clinic),
    fixedTime: CREATED_AT,
    renderedAssertions: [
      { kind: "visibleText", value: SYNTHETIC_PET.name },
      { kind: "visibleText", value: "再診" },
      {
        kind: "containsText",
        selector: "button[aria-label^='担当医:']",
        value: SYNTHETIC_STAFF.name,
      },
      { kind: "value", selector: "input[aria-label='診察日']", value: CREATED_AT.slice(0, 10) },
    ],
    expectedLocalBusinessNonGet: ["POST:/api/v1/reservations", "POST:/api/v1/medical-records"],
    endpoints: defineEndpoints([
      SYNTHETIC_ME_ENDPOINT,
      { method: "GET", pathname: `/api/v1/pets/${SYNTHETIC_IDS.pet}`, response: SYNTHETIC_PET },
      {
        method: "GET",
        pathname: `/api/v1/owners/${SYNTHETIC_IDS.owner}`,
        response: SYNTHETIC_OWNER_API_RESPONSE,
      },
      {
        method: "GET",
        pathname: "/api/v1/masters/reservation-types",
        response: [SYNTHETIC_RESERVATION_TYPE],
      },
      {
        method: "GET",
        pathname: "/api/v1/reservations",
        query: {
          page: "1",
          limit: "100",
          date: CREATED_AT.slice(0, 10),
          pet_id: String(SYNTHETIC_IDS.pet),
        } satisfies Readonly<Record<string, string>>,
        response: { data: [], total: 0, page: 1, limit: 100 },
      },
      {
        method: "POST",
        pathname: "/api/v1/reservations",
        validateBody: validateReservationCreate,
        response: SYNTHETIC_RESERVATION,
      },
      {
        method: "POST",
        pathname: "/api/v1/medical-records",
        validateBody: validateMedicalRecordCreate,
        response: SYNTHETIC_MEDICAL_RECORD,
      },
      { method: "GET", pathname: MEDICAL_RECORD_PATH, response: SYNTHETIC_MEDICAL_RECORD },
      {
        method: "GET",
        pathname: "/api/v1/medical-records",
        query: {
          limit: "50",
          page: "1",
          pet_id: String(SYNTHETIC_IDS.pet),
        } satisfies Readonly<Record<string, string>>,
        response: { data: [], total: 0, page: 1, limit: 50 },
      },
      {
        method: "GET",
        pathname: `${MEDICAL_RECORD_PATH}/clinical-plan`,
        response: SYNTHETIC_CLINICAL_PLAN,
      },
      { method: "GET", pathname: `${MEDICAL_RECORD_PATH}/treatments`, response: [] },
      { method: "GET", pathname: `${MEDICAL_RECORD_PATH}/vitals`, response: [] },
      { method: "GET", pathname: `${MEDICAL_RECORD_PATH}/addenda`, response: [] },
      { method: "GET", pathname: "/api/v1/masters/chief-complaint-types", response: [] },
      { method: "GET", pathname: "/api/v1/masters/diagnosis-types", response: [] },
      { method: "GET", pathname: "/api/v1/masters/diagnosis-names", response: [] },
      { method: "GET", pathname: "/api/v1/masters/staffs", response: [SYNTHETIC_STAFF] },
      {
        method: "GET",
        pathname: `/api/v1/clinics/${SYNTHETIC_IDS.clinic}/owners/${SYNTHETIC_IDS.owner}/lstep/tags`,
        response: SYNTHETIC_LINE_STATUS,
      },
    ]),
  },
  hospitalizationDetail: {
    entryPath: `/hospitalization/${SYNTHETIC_IDS.hospitalization}`,
    renderedPath: `/hospitalization/${SYNTHETIC_IDS.hospitalization}`,
    currentClinicId: String(SYNTHETIC_IDS.clinic),
    fixedTime: CREATED_AT,
    renderedAssertions: [
      { kind: "visibleText", value: SYNTHETIC_PET.name },
      { kind: "visibleText", value: SYNTHETIC_OWNER.name },
      { kind: "visibleText", value: SYNTHETIC_STAFF.name },
      { kind: "visibleText", value: "退院処理" },
      { kind: "visibleText", value: "デイリーカルテ" },
      { kind: "clickTabIfVisible", value: "プラン管理・詳細" },
      { kind: "visibleText", value: SYNTHETIC_CARE_PLAN_ITEM.name },
      { kind: "visibleText", value: "単価 ￥4,321" },
    ],
    expectedLocalBusinessNonGet: [],
    endpoints: defineEndpoints([
      SYNTHETIC_ME_ENDPOINT,
      { method: "GET", pathname: HOSPITALIZATION_PATH, response: SYNTHETIC_HOSPITALIZATION },
      {
        method: "GET",
        pathname: `${HOSPITALIZATION_PATH}/care-plan-items`,
        response: [SYNTHETIC_CARE_PLAN_ITEM],
      },
      {
        method: "GET",
        pathname: new RegExp(`^${HOSPITALIZATION_PATH}/daily-records/\\d{4}-\\d{2}-\\d{2}$`),
        response: SYNTHETIC_DAILY_RECORD,
      },
    ]),
  },
  hospitalizationEdit: {
    entryPath: `/hospitalization/${SYNTHETIC_IDS.hospitalization}/edit`,
    renderedPath: `/hospitalization/${SYNTHETIC_IDS.hospitalization}/edit`,
    currentClinicId: String(SYNTHETIC_IDS.clinic),
    fixedTime: CREATED_AT,
    renderedAssertions: [
      { kind: "visibleText", value: SYNTHETIC_PET.name },
      { kind: "checked", selector: "#type-hospitalization" },
      { kind: "containsText", selector: "#cage_id", value: SYNTHETIC_CAGE.name },
      { kind: "value", selector: "#memo", value: SYNTHETIC_HOSPITALIZATION.memo },
      { kind: "value", selector: "#owner_request", value: SYNTHETIC_HOSPITALIZATION.owner_request },
      { kind: "value", selector: "#staff_notes", value: SYNTHETIC_HOSPITALIZATION.staff_notes },
      { kind: "visibleText", value: "診療費 小計" },
      {
        kind: "value",
        selector: "input[placeholder='治療内容を入力...']",
        value: SYNTHETIC_TREATMENT_PLAN.treatment_content,
      },
      { kind: "disabled", selector: "button:has-text('追加')" },
      { kind: "disabled", selector: "input[placeholder='治療内容を入力...']" },
      { kind: "disabled", selector: "#hospitalization-global-discount" },
      { kind: "visibleText", value: "会計時に確定します" },
      { kind: "visibleText", value: "￥6,355" },
    ],
    expectedLocalBusinessNonGet: [],
    endpoints: defineEndpoints([
      SYNTHETIC_ME_ENDPOINT,
      { method: "GET", pathname: HOSPITALIZATION_PATH, response: SYNTHETIC_HOSPITALIZATION },
      { method: "GET", pathname: "/api/v1/masters/cages", response: [SYNTHETIC_CAGE] },
    ]),
  },
  trimmingEdit: {
    entryPath: `/trimming/${SYNTHETIC_IDS.trimming}`,
    renderedPath: `/trimming/${SYNTHETIC_IDS.trimming}`,
    currentClinicId: String(SYNTHETIC_IDS.clinic),
    fixedTime: CREATED_AT,
    renderedAssertions: [
      { kind: "visibleText", value: SYNTHETIC_PET.name },
      { kind: "containsText", selector: "#staffId", value: SYNTHETIC_STAFF.name },
      { kind: "containsText", selector: "#courseId", value: SYNTHETIC_TRIMMING_COURSE.name },
      { kind: "checked", selector: `#option-${SYNTHETIC_IDS.trimmingOption}` },
      {
        kind: "value",
        selector: "textarea[placeholder='スタイルの希望を入力...']",
        value: SYNTHETIC_TRIMMING.style_request,
      },
      {
        kind: "value",
        selector: "input[placeholder='体重']",
        value: String(SYNTHETIC_TRIMMING.bw),
      },
      {
        kind: "value",
        selector: "input[placeholder='体温']",
        value: String(SYNTHETIC_TRIMMING.bt),
      },
      {
        kind: "value",
        selector: "input[placeholder='シャンプー名']",
        value: SYNTHETIC_TRIMMING.used_shampoo,
      },
      {
        kind: "value",
        selector: "input[placeholder='リボン']",
        value: SYNTHETIC_TRIMMING.used_ribbon,
      },
    ],
    expectedLocalBusinessNonGet: [],
    endpoints: defineEndpoints([
      SYNTHETIC_ME_ENDPOINT,
      {
        method: "GET",
        pathname: `/api/v1/trimmings/${SYNTHETIC_IDS.trimming}`,
        response: SYNTHETIC_TRIMMING,
      },
      {
        method: "GET",
        pathname: "/api/v1/trimmings",
        query: {
          pet_id: String(SYNTHETIC_IDS.pet),
          page: "1",
          limit: "100",
        } satisfies Readonly<Record<string, string>>,
        response: { data: [SYNTHETIC_TRIMMING], total: 1, page: 1, limit: 100 },
      },
      {
        method: "GET",
        pathname: "/api/v1/masters/reservation-types",
        response: [SYNTHETIC_RESERVATION_TYPE],
      },
      {
        method: "GET",
        pathname: "/api/v1/masters/trimming-courses",
        response: [SYNTHETIC_TRIMMING_COURSE],
      },
      {
        method: "GET",
        pathname: "/api/v1/masters/trimming-options",
        response: [SYNTHETIC_TRIMMING_OPTION],
      },
      { method: "GET", pathname: "/api/v1/masters/staffs", response: [SYNTHETIC_STAFF] },
      {
        method: "GET",
        pathname: "/api/v1/masters/trimming-course-types",
        response: [SYNTHETIC_TRIMMING_COURSE_TYPE],
      },
    ]),
  },
  vaccinationEdit: {
    entryPath: `/vaccinations/${SYNTHETIC_IDS.vaccination}`,
    renderedPath: `/vaccinations/${SYNTHETIC_IDS.vaccination}`,
    currentClinicId: String(SYNTHETIC_IDS.clinic),
    fixedTime: CREATED_AT,
    renderedAssertions: [
      { kind: "visibleText", value: SYNTHETIC_PET.name },
      { kind: "visibleText", value: SYNTHETIC_STAFF.name },
      { kind: "valueContains", selector: "#vaccination-date", value: "2026年7月23日" },
      { kind: "selectedOption", selector: "#vaccine-select", value: SYNTHETIC_VACCINE.name },
      {
        kind: "value",
        selector: "input[placeholder='補助説明を入力']",
        value: SYNTHETIC_VACCINATION.supplemental,
      },
      {
        kind: "value",
        selector: "input[placeholder='LOT 1番号']",
        value: SYNTHETIC_VACCINATION.lot1,
      },
      { kind: "valueContains", selector: "#vaccination-next-date", value: "2027年7月23日" },
      {
        kind: "value",
        selector: "textarea[placeholder='備考を入力']",
        value: SYNTHETIC_VACCINATION.remarks,
      },
    ],
    expectedLocalBusinessNonGet: [],
    endpoints: defineEndpoints([
      SYNTHETIC_ME_ENDPOINT,
      {
        method: "GET",
        pathname: `/api/v1/vaccinations/${SYNTHETIC_IDS.vaccination}`,
        response: SYNTHETIC_VACCINATION,
      },
      {
        method: "GET",
        pathname: "/api/v1/vaccinations",
        response: { data: [SYNTHETIC_VACCINATION] },
      },
      { method: "GET", pathname: `/api/v1/pets/${SYNTHETIC_IDS.pet}`, response: SYNTHETIC_PET },
      { method: "GET", pathname: "/api/v1/masters/vaccines", response: [SYNTHETIC_VACCINE] },
    ]),
  },
} satisfies Record<SyntheticClinicalKey, SyntheticClinicalScenario>;
