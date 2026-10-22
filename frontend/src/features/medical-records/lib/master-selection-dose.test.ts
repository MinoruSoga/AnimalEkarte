import { describe, expect, it } from "vitest";
import { QueryClient } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";

import { server } from "@/testing/mocks/node";
import { queryKeys } from "@/lib/query-keys";

import { resolveMasterSelectionDose } from "./master-selection-dose";
import { DOSE_PARAMS_LOOKUP_FAILED_MESSAGE } from "../api/medicine-dose-lookup";

const qc = () => new QueryClient({ defaultOptions: { queries: { retry: false } } });

describe("resolveMasterSelectionDose", () => {
  it.each(["masters", "vitals", "params"] as const)(
    "malformed %s response blocks with a fixed technical error",
    async (broken) => {
      server.use(
        http.get("*/v1/masters/medicines", () =>
          HttpResponse.json(
            broken === "masters"
              ? {}
              : [
                  {
                    id: 42,
                    name: "薬",
                    price: 100,
                    is_active: true,
                    tax_rate: "0.1",
                    calculation_type: "per_weight",
                    medicine_unit: "per_tablet",
                    strength: "100",
                  },
                ],
          ),
        ),
        http.get("*/v1/medical-records/10/vitals", () =>
          HttpResponse.json(
            broken === "vitals"
              ? {}
              : [{ id: "1", recorded_at: "2026-10-18T00:00:00Z", weight: 10, weight_unit: "Kg" }],
          ),
        ),
        http.get("*/v1/masters/medicines/42/dose-params", () => HttpResponse.json({})),
      );
      await expect(
        resolveMasterSelectionDose({
          item: { id: "42", name: "薬", unitPrice: 100, category: "薬剤", medicineId: "42" },
          medicalRecordId: "10",
          recordClinicId: "3",
          petSpecies: "犬",
          queryClient: qc(),
        }),
      ).resolves.toEqual({
        kind: "blocked",
        message: DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
        retryable: true,
      });
    },
  );
  it("非薬剤（診察/処置）は GET せず ready qty=1 を即返す", async () => {
    let requests = 0;
    server.use(
      http.get("*/v1/masters/medicines", () => {
        requests += 1;
        return HttpResponse.json([]);
      }),
      http.get("*/v1/medical-records/10/vitals", () => {
        requests += 1;
        return HttpResponse.json([]);
      }),
    );
    const result = await resolveMasterSelectionDose({
      item: { id: "7", name: "初診", unitPrice: 3307, category: "診察" },
      medicalRecordId: "10",
      recordClinicId: "3",
      queryClient: qc(),
    });
    expect(result).toEqual({ kind: "ready", quantity: 1 });
    expect(requests).toBe(0);
  });

  it("medicineId 未設定の薬剤は retryable blocked（qty=1 フォールバックしない）", async () => {
    const result = await resolveMasterSelectionDose({
      item: { id: "42", name: "薬A", unitPrice: 100, category: "薬剤" },
      medicalRecordId: "10",
      queryClient: qc(),
    });
    expect(result.kind).toBe("blocked");
    if (result.kind === "blocked") expect(result.retryable).toBe(true);
  });

  it("一致マスタが inactive/不在なら非リトライの再選択 blocks（silent qty1 にしない）", async () => {
    server.use(
      http.get("*/v1/masters/medicines", () =>
        HttpResponse.json([{ id: 42, name: "薬A", price: 100, is_active: false, tax_rate: "0.1" }]),
      ),
      http.get("*/v1/medical-records/10/vitals", () => HttpResponse.json([])),
    );
    const result = await resolveMasterSelectionDose({
      item: { id: "42", name: "薬A", unitPrice: 100, category: "薬剤", medicineId: "42" },
      medicalRecordId: "10",
      recordClinicId: "3",
      queryClient: qc(),
    });
    expect(result.kind).toBe("blocked");
    if (result.kind === "blocked") {
      expect(result.retryable).toBe(false);
      expect(result.message).toContain("再選択");
    }
  });

  it("計算種別 none の薬剤は vitals を引かず、vitals 障害でも ready qty=1", async () => {
    let vitalsRequests = 0;
    server.use(
      http.get("*/v1/masters/medicines", () =>
        HttpResponse.json([
          {
            id: 42,
            name: "薬A",
            price: 100,
            is_active: true,
            tax_rate: "0.1",
            calculation_type: "none",
          },
        ]),
      ),
      http.get("*/v1/medical-records/10/vitals", () => {
        vitalsRequests += 1;
        return HttpResponse.json({}, { status: 500 });
      }),
    );
    const result = await resolveMasterSelectionDose({
      item: { id: "42", name: "薬A", unitPrice: 100, category: "薬剤", medicineId: "42" },
      medicalRecordId: "10",
      recordClinicId: "3",
      queryClient: qc(),
    });
    expect(result).toEqual({ kind: "ready", quantity: 1 });
    expect(vitalsRequests).toBe(0);
  });

  it("per_weight 薬でも species 不明なら dose-params を引かず、障害でも ready qty=1", async () => {
    let doseParamsRequests = 0;
    server.use(
      http.get("*/v1/masters/medicines", () =>
        HttpResponse.json([
          {
            id: 42,
            name: "薬A",
            price: 100,
            is_active: true,
            tax_rate: "0.1",
            calculation_type: "per_weight",
          },
        ]),
      ),
      http.get("*/v1/medical-records/10/vitals", () =>
        HttpResponse.json([
          { id: "1", recorded_at: "2026-10-18T00:00:00Z", weight: 4, weight_unit: "Kg" },
        ]),
      ),
      http.get("*/v1/masters/medicines/42/dose-params", () => {
        doseParamsRequests += 1;
        return HttpResponse.json({}, { status: 500 });
      }),
    );
    const result = await resolveMasterSelectionDose({
      item: { id: "42", name: "薬A", unitPrice: 100, category: "薬剤", medicineId: "42" },
      medicalRecordId: "10",
      recordClinicId: "3",
      petSpecies: "ペンギン",
      queryClient: qc(),
    });
    expect(result).toEqual({ kind: "ready", quantity: 1 });
    expect(doseParamsRequests).toBe(0);
  });

  it("per_weight 薬で体重なしのときも dose-params を引かず ready qty=1", async () => {
    let doseParamsRequests = 0;
    server.use(
      http.get("*/v1/masters/medicines", () =>
        HttpResponse.json([
          {
            id: 42,
            name: "薬A",
            price: 100,
            is_active: true,
            tax_rate: "0.1",
            calculation_type: "per_weight",
          },
        ]),
      ),
      http.get("*/v1/medical-records/10/vitals", () => HttpResponse.json([])),
      http.get("*/v1/masters/medicines/42/dose-params", () => {
        doseParamsRequests += 1;
        return HttpResponse.json({}, { status: 500 });
      }),
    );
    const result = await resolveMasterSelectionDose({
      item: { id: "42", name: "薬A", unitPrice: 100, category: "薬剤", medicineId: "42" },
      medicalRecordId: "10",
      recordClinicId: "3",
      petSpecies: "犬",
      queryClient: qc(),
    });
    expect(result).toEqual({ kind: "ready", quantity: 1 });
    expect(doseParamsRequests).toBe(0);
  });

  it("HTTP 失敗（masters GET 500）は retryable blocked", async () => {
    server.use(
      http.get("*/v1/masters/medicines", () =>
        HttpResponse.json({ message: "x" }, { status: 500 }),
      ),
      http.get("*/v1/medical-records/10/vitals", () => HttpResponse.json([])),
    );
    const result = await resolveMasterSelectionDose({
      item: { id: "42", name: "薬A", unitPrice: 100, category: "薬剤", medicineId: "42" },
      medicalRecordId: "10",
      recordClinicId: "3",
      queryClient: qc(),
    });
    expect(result.kind).toBe("blocked");
    if (result.kind === "blocked") expect(result.retryable).toBe(true);
  });

  it("体重・dose-params 無しでも成功 path は qty=1（missing policy 維持）", async () => {
    server.use(
      http.get("*/v1/masters/medicines", () =>
        HttpResponse.json([{ id: 42, name: "薬A", price: 100, is_active: true, tax_rate: "0.1" }]),
      ),
      http.get("*/v1/medical-records/10/vitals", () => HttpResponse.json([])),
    );
    const result = await resolveMasterSelectionDose({
      item: { id: "42", name: "薬A", unitPrice: 100, category: "薬剤", medicineId: "42" },
      medicalRecordId: "10",
      queryClient: qc(),
    });
    expect(result).toEqual({ kind: "ready", quantity: 1 });
  });

  it("scoped clinicId の medicines キャッシュを使う（グローバル不一致は拾わない）", async () => {
    const queryClient = qc();
    queryClient.setQueryData(queryKeys.masters.category("medicines"), [
      { id: "42", name: "薬A", price: 11, isActive: true },
    ]);
    queryClient.setQueryData(queryKeys.masters.categoryForClinic("medicines", "3"), [
      { id: "42", name: "薬A", price: 3303, isActive: true },
    ]);
    server.use(http.get("*/v1/medical-records/10/vitals", () => HttpResponse.json([])));
    const result = await resolveMasterSelectionDose({
      item: { id: "42", name: "薬A", unitPrice: 100, category: "薬剤", medicineId: "42" },
      medicalRecordId: "10",
      recordClinicId: "3",
      queryClient,
    });
    expect(result).toEqual({ kind: "ready", quantity: 1 });
  });
});
