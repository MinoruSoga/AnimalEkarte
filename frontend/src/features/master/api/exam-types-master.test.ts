import { beforeEach, describe, expect, it, vi } from "vitest";
import { QueryClient } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import {
  createExaminationTypeField,
  deleteExaminationTypeField,
  invalidateExaminationTypeFieldQueries,
  reorderExaminationTypeFields,
  replaceExamTypeFieldReferenceRanges,
  transformExaminationTypeResponse,
  updateExaminationTypeField,
} from "./exam-types-master";

vi.mock("@/lib/axios", () => ({
  axios: {
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    put: vi.fn(),
  },
}));

describe("examination type field master API", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("preserves items and reference ranges from the examination type response", () => {
    const result = transformExaminationTypeResponse({
      id: 3,
      clinic_id: 9,
      name: "血液検査",
      price: 1200,
      is_active: true,
      description: "",
      sort_order: 1,
      is_non_insurance: false,
      items: [
        {
          id: 31,
          exam_type_id: 3,
          name: "白血球",
          inspection_value: "",
          normal_value: "",
          unit: "/μL",
          sort_order: 2,
          created_at: "2026-07-27T00:00:00Z",
          updated_at: "2026-07-27T00:00:00Z",
          reference_ranges: [
            {
              id: 41,
              exam_type_field_id: 31,
              animal_species_id: 2,
              ref_min: 5,
              ref_max: 10,
            },
          ],
        },
      ],
      created_at: "2026-07-27T00:00:00Z",
      updated_at: "2026-07-27T00:00:00Z",
    });

    expect(result.items).toEqual([
      expect.objectContaining({
        id: "31",
        name: "白血球",
        unit: "/μL",
        referenceRanges: [
          expect.objectContaining({
            id: "41",
            animalSpeciesId: "2",
            refMin: 5,
            refMax: 10,
          }),
        ],
      }),
    ]);
  });

  it("uses the fixed CRUD, reorder, replace, and explicit clear endpoints", async () => {
    vi.mocked(axios.post).mockResolvedValue({ data: { id: 1 } });
    vi.mocked(axios.patch).mockResolvedValue({ data: { id: 1 } });
    vi.mocked(axios.put).mockResolvedValue({ data: { id: 1 } });
    vi.mocked(axios.delete).mockResolvedValue({ data: undefined });

    await createExaminationTypeField("3", { name: "白血球", sort_order: 1 });
    await updateExaminationTypeField("3", "31", { unit: "/μL" });
    await deleteExaminationTypeField("3", "31");
    await reorderExaminationTypeFields("3", [31, 32]);
    await replaceExamTypeFieldReferenceRanges("3", "31", [
      {
        animal_species_id: 2,
        ref_min: 5,
        ref_max: 10,
      },
    ]);
    await replaceExamTypeFieldReferenceRanges("3", "31", []);

    expect(axios.post).toHaveBeenCalledWith("/v1/masters/examination-types/3/fields", {
      name: "白血球",
      sort_order: 1,
    });
    expect(axios.patch).toHaveBeenCalledWith("/v1/masters/examination-types/3/fields/31", {
      unit: "/μL",
    });
    expect(axios.delete).toHaveBeenCalledWith("/v1/masters/examination-types/3/fields/31");
    expect(axios.patch).toHaveBeenCalledWith("/v1/masters/examination-types/3/fields/reorder", {
      ids: [31, 32],
    });
    expect(axios.put).toHaveBeenNthCalledWith(
      1,
      "/v1/masters/examination-types/3/fields/31/reference-ranges",
      { ranges: [{ animal_species_id: 2, ref_min: 5, ref_max: 10 }] },
    );
    expect(axios.put).toHaveBeenNthCalledWith(
      2,
      "/v1/masters/examination-types/3/fields/31/reference-ranges",
      { ranges: [] },
    );
  });

  it("invalidates both master and examination field query keys", async () => {
    const queryClient = new QueryClient();
    const invalidate = vi.spyOn(queryClient, "invalidateQueries").mockResolvedValue(undefined);

    await invalidateExaminationTypeFieldQueries(queryClient, "3");

    expect(invalidate).toHaveBeenNthCalledWith(1, {
      queryKey: queryKeys.masters.category("examination-types"),
    });
    expect(invalidate).toHaveBeenNthCalledWith(2, {
      queryKey: queryKeys.examinations.typeFields("3"),
    });
  });
});
