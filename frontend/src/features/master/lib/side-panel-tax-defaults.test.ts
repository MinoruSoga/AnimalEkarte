import { describe, expect, it } from "vitest";

import type { Medicine } from "@/types";
import type { FrontendMerchandiseItem } from "../api/merchandise-items";
import type { HospitalizationPlan } from "../api/hospitalization-plans";

import { medicineToFormData } from "./medicine-side-panel-model";
import { merchandiseToFormData } from "./merchandise-side-panel-model";
import { hospitalizationToFormData } from "./hospitalization-side-panel-model";

const existingMedicine = (taxType?: "included" | "excluded" | "exempt") =>
  ({
    id: "1",
    name: "既存薬",
    price: 100,
    description: "",
    isActive: true,
    taxType,
    taxRate: 0.1,
    isNonInsurance: false,
  }) as Medicine;

const existingMerchandise = (taxType?: "included" | "excluded" | "exempt") =>
  ({
    id: "1",
    name: "既存商品",
    category: "goods",
    unitPrice: 100,
    taxType,
    taxRate: 0.1,
    isActive: true,
    sortOrder: 1,
  }) as FrontendMerchandiseItem;

const existingHospitalization = (taxType?: "included" | "excluded" | "exempt") =>
  ({
    id: "1",
    name: "既存入院プラン",
    price: 100,
    isActive: true,
    description: "",
    bodySize: null,
    billingUnit: null,
    taxType,
    taxRate: 0.1,
    sortOrder: 1,
  }) as HospitalizationPlan;

describe("master 新規フォームの taxType 既定値 (EMR-401 C7-9)", () => {
  it("medicineToFormData(null) は新規を 内税(included) にする", () => {
    expect(medicineToFormData(null).taxType).toBe("included");
  });

  it("medicineToFormData は既存の excluded/exempt を保持し、legacy 欠落は excluded に落とす", () => {
    expect(medicineToFormData(existingMedicine("excluded")).taxType).toBe("excluded");
    expect(medicineToFormData(existingMedicine("exempt")).taxType).toBe("exempt");
    expect(medicineToFormData(existingMedicine(undefined)).taxType).toBe("excluded");
  });

  it("merchandiseToFormData(null) は新規を included にする", () => {
    expect(merchandiseToFormData(null).taxType).toBe("included");
  });

  it("merchandiseToFormData は既存の excluded/exempt を保持し、legacy 欠落は excluded に落とす", () => {
    expect(merchandiseToFormData(existingMerchandise("excluded")).taxType).toBe("excluded");
    expect(merchandiseToFormData(existingMerchandise("exempt")).taxType).toBe("exempt");
    expect(merchandiseToFormData(existingMerchandise(undefined)).taxType).toBe("excluded");
  });

  it("hospitalizationToFormData(null) は新規を included にする", () => {
    expect(hospitalizationToFormData(null).taxType).toBe("included");
  });

  it("hospitalizationToFormData は既存の excluded/exempt を保持し、legacy 欠落は excluded に落とす", () => {
    expect(hospitalizationToFormData(existingHospitalization("excluded")).taxType).toBe("excluded");
    expect(hospitalizationToFormData(existingHospitalization("exempt")).taxType).toBe("exempt");
    expect(hospitalizationToFormData(existingHospitalization(undefined)).taxType).toBe("excluded");
  });
});
