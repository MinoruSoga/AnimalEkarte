import { Accounting } from "./types";

export const MOCK_ACCOUNTING_LIST: Accounting[] = [
  {
    id: "acc_001",
    medicalRecordId: "mr_001",
    ownerId: "own_001",
    ownerName: "山田 太郎",
    petId: "pet_001",
    petName: "ポチ",
    petSpecies: "犬",
    status: "waiting",
    scheduledDate: "2024-03-20",
    items: [
      {
        id: "item_001",
        category: "examination",
        name: "再診料",
        unitPrice: 800,
        quantity: 1,
        taxRate: 0.1,
        isInsuranceApplicable: true,
        source: "medical_record"
      },
      {
        id: "item_002",
        category: "medicine",
        name: "内服薬A",
        unitPrice: 2000,
        quantity: 1,
        taxRate: 0.1,
        isInsuranceApplicable: true,
        source: "medical_record"
      }
    ]
  },
  {
    id: "acc_002",
    medicalRecordId: "mr_002",
    ownerId: "own_002",
    ownerName: "佐藤 花子",
    petId: "pet_002",
    petName: "ミケ",
    petSpecies: "猫",
    status: "waiting",
    scheduledDate: "2024-03-20",
    items: [
      {
        id: "item_003",
        category: "examination",
        name: "初診料",
        unitPrice: 1500,
        quantity: 1,
        taxRate: 0.1,
        isInsuranceApplicable: true,
        source: "medical_record"
      },
      {
        id: "item_004",
        category: "test",
        name: "血液検査一式",
        unitPrice: 8000,
        quantity: 1,
        taxRate: 0.1,
        isInsuranceApplicable: true,
        source: "medical_record"
      }
    ]
  },
  {
    id: "acc_003",
    medicalRecordId: "mr_003",
    ownerId: "own_003",
    ownerName: "鈴木 一郎",
    petId: "pet_003",
    petName: "タロウ",
    petSpecies: "犬",
    status: "completed",
    scheduledDate: "2024-03-20",
    completedAt: "2024-03-20T10:30:00",
    items: [
      {
        id: "item_005",
        category: "food",
        name: "療法食XYZ",
        unitPrice: 4500,
        quantity: 2,
        taxRate: 0.08,
        isInsuranceApplicable: false,
        source: "manual"
      }
    ],
    payment: {
      subtotal: 9000,
      taxTotal: 720,
      totalAmount: 9720,
      insuranceAmount: 0,
      discountAmount: 0,
      billingAmount: 9720,
      receivedAmount: 10000,
      changeAmount: 280,
      method: "cash"
    }
  },
  {
    id: "acc_004",
    medicalRecordId: "mr_004",
    ownerId: "own_004",
    ownerName: "高橋 由美",
    petId: "pet_004",
    petName: "レオ",
    petSpecies: "猫",
    status: "pending",
    scheduledDate: "2024-03-19",
    items: [
        {
            id: "item_006",
            category: "surgery",
            name: "避妊手術",
            unitPrice: 25000,
            quantity: 1,
            taxRate: 0.1,
            isInsuranceApplicable: true,
            source: "medical_record"
        }
    ],
    memo: "カード忘れのため後日支払い"
  }
];
