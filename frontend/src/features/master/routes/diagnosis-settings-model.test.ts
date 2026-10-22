import { describe, expect, it } from "vitest";

import type {
  DiagnosisNameFormData,
  DiagnosisTypeFormData,
} from "../lib/diagnosis-side-panel-model";
import {
  buildDiagnosisNameCreateRequest,
  buildDiagnosisNameUpdateRequest,
  buildDiagnosisTypeCreateRequest,
  buildDiagnosisTypeUpdateRequest,
} from "./diagnosis-settings-model";

const TYPE_FORM: DiagnosisTypeFormData = {
  name: "皮膚疾患",
  description: "",
  isActive: false,
};

const NAME_FORM: DiagnosisNameFormData = {
  name: "膿皮症",
  diagnosisTypeId: "12",
  description: "",
  isActive: true,
};

describe("EMR-529 診断カテゴリの備考", () => {
  it("Create は空の備考を送らない（既定を維持）", () => {
    expect(buildDiagnosisTypeCreateRequest(TYPE_FORM)).toEqual({
      name: "皮膚疾患",
      description: undefined,
      is_active: true,
    });
  });

  it("Update は空の備考を明示的に送り、既存備考を空へ変更できる", () => {
    expect(buildDiagnosisTypeUpdateRequest(TYPE_FORM)).toEqual({
      name: "皮膚疾患",
      description: "",
      is_active: false,
    });
  });

  it("Update は非空の備考をそのまま送る", () => {
    expect(buildDiagnosisTypeUpdateRequest({ ...TYPE_FORM, description: "追記" })).toMatchObject({
      description: "追記",
    });
  });
});

describe("EMR-529 診断病名の備考", () => {
  it("Create は空の備考を送らない（既定を維持）", () => {
    expect(buildDiagnosisNameCreateRequest(NAME_FORM)).toEqual({
      name: "膿皮症",
      diagnosis_type_id: 12,
      description: undefined,
      is_active: true,
    });
  });

  it("Update は空の備考を明示的に送り、所属カテゴリと有効状態を保つ", () => {
    expect(buildDiagnosisNameUpdateRequest({ ...NAME_FORM, isActive: false })).toEqual({
      name: "膿皮症",
      diagnosis_type_id: 12,
      description: "",
      is_active: false,
    });
  });

  it("Update は非空の備考をそのまま送る", () => {
    expect(buildDiagnosisNameUpdateRequest({ ...NAME_FORM, description: "再発例" })).toMatchObject({
      description: "再発例",
    });
  });
});
