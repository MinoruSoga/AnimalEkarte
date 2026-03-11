import type { AuthUser, Permission } from "../types";

const CLINIC_A_ID = "clinic-001";
const CLINIC_B_ID = "clinic-002";
const CLINIC_C_ID = "clinic-003";

export const MOCK_PASSWORD = "password";

interface MockCredential {
  email: string;
  password: string;
  user: AuthUser;
}

const ADMIN_PERMISSIONS: readonly Permission[] = [
  "account_admin",
  "medical",
  "medical_read",
  "trimming",
  "billing",
  "reception",
  "hospitalization",
  "master_admin",
  "shift_admin",
  "inventory",
];

export const MOCK_USERS: readonly MockCredential[] = [
  {
    email: "admin@example.com",
    password: MOCK_PASSWORD,
    user: {
      id: "user-001",
      email: "admin@example.com",
      displayName: "田中 太郎",
      userType: "clinic_admin",
      jobTitle: "veterinarian",
      avatarUrl: null,
      mainClinicId: CLINIC_A_ID,
      clinics: [
        { clinicId: CLINIC_A_ID, clinicName: "ノア動物病院", branchName: "八王子院", isMain: true },
        { clinicId: CLINIC_B_ID, clinicName: "ノア動物病院", branchName: "城東医院", isMain: false },
        { clinicId: CLINIC_C_ID, clinicName: "ノア動物病院", branchName: "敷島医院", isMain: false },
      ],
      permissions: {
        [CLINIC_A_ID]: ADMIN_PERMISSIONS,
        [CLINIC_B_ID]: ADMIN_PERMISSIONS,
        [CLINIC_C_ID]: ADMIN_PERMISSIONS,
      },
    },
  },
  {
    email: "vet@example.com",
    password: MOCK_PASSWORD,
    user: {
      id: "user-002",
      email: "vet@example.com",
      displayName: "山田 花子",
      userType: "staff",
      jobTitle: "veterinarian",
      avatarUrl: null,
      mainClinicId: CLINIC_A_ID,
      clinics: [
        { clinicId: CLINIC_A_ID, clinicName: "ノア動物病院", branchName: "八王子院", isMain: true },
      ],
      permissions: { [CLINIC_A_ID]: ["medical", "hospitalization"] },
    },
  },
  {
    email: "nurse@example.com",
    password: MOCK_PASSWORD,
    user: {
      id: "user-003",
      email: "nurse@example.com",
      displayName: "佐藤 美咲",
      userType: "staff",
      jobTitle: "nurse",
      avatarUrl: null,
      mainClinicId: CLINIC_A_ID,
      clinics: [
        { clinicId: CLINIC_A_ID, clinicName: "ノア動物病院", branchName: "八王子院", isMain: true },
      ],
      permissions: { [CLINIC_A_ID]: ["medical_read", "hospitalization", "shift_admin", "inventory"] },
    },
  },
  {
    email: "reception@example.com",
    password: MOCK_PASSWORD,
    user: {
      id: "user-004",
      email: "reception@example.com",
      displayName: "鈴木 一郎",
      userType: "staff",
      jobTitle: "reception",
      avatarUrl: null,
      mainClinicId: CLINIC_A_ID,
      clinics: [
        { clinicId: CLINIC_A_ID, clinicName: "ノア動物病院", branchName: "八王子院", isMain: true },
      ],
      permissions: { [CLINIC_A_ID]: ["reception", "billing"] },
    },
  },
  {
    email: "trimmer@example.com",
    password: MOCK_PASSWORD,
    user: {
      id: "user-005",
      email: "trimmer@example.com",
      displayName: "高橋 さくら",
      userType: "staff",
      jobTitle: "trimmer",
      avatarUrl: null,
      mainClinicId: CLINIC_A_ID,
      clinics: [
        { clinicId: CLINIC_A_ID, clinicName: "ノア動物病院", branchName: "八王子院", isMain: true },
      ],
      permissions: { [CLINIC_A_ID]: ["trimming", "reception", "billing"] },
    },
  },
  {
    email: "system@example.com",
    password: MOCK_PASSWORD,
    user: {
      id: "user-006",
      email: "system@example.com",
      displayName: "本部 管理者",
      userType: "system_admin",
      jobTitle: null,
      avatarUrl: null,
      mainClinicId: CLINIC_A_ID,
      clinics: [
        { clinicId: CLINIC_A_ID, clinicName: "ノア動物病院", branchName: "八王子院", isMain: true },
        { clinicId: CLINIC_B_ID, clinicName: "ノア動物病院", branchName: "城東医院", isMain: false },
        { clinicId: CLINIC_C_ID, clinicName: "ノア動物病院", branchName: "敷島医院", isMain: false },
      ],
      permissions: {},
    },
  },
];
