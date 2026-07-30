import { describe, it, expect, vi, beforeEach } from "vitest";

import { axios } from "@/lib/axios";

import {
  MEDICAL_RECORD_IMAGE_UPLOAD_CONCURRENCY,
  mapWithConcurrency,
  uploadMedicalRecordImages,
} from "./medical-record-images";

vi.mock("@/lib/axios", () => ({
  axios: {
    post: vi.fn(),
    delete: vi.fn(),
  },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

describe("mapWithConcurrency — SEC-CS-F08 bounded pool", () => {
  it("同時実行数を concurrency 以下に抑える", async () => {
    let inFlight = 0;
    let maxInFlight = 0;
    const items = Array.from({ length: 8 }, (_, i) => i);

    const results = await mapWithConcurrency(items, 3, async (item) => {
      inFlight += 1;
      maxInFlight = Math.max(maxInFlight, inFlight);
      await new Promise((resolve) => setTimeout(resolve, 20));
      inFlight -= 1;
      return item * 2;
    });

    expect(results).toEqual([0, 2, 4, 6, 8, 10, 12, 14]);
    expect(maxInFlight).toBeLessThanOrEqual(3);
    expect(maxInFlight).toBe(3);
  });

  it("空配列は空結果を返す", async () => {
    const mapper = vi.fn();
    await expect(mapWithConcurrency([], 3, mapper)).resolves.toEqual([]);
    expect(mapper).not.toHaveBeenCalled();
  });
});

describe("uploadMedicalRecordImages — SEC-CS-F08", () => {
  beforeEach(() => {
    vi.mocked(axios.post).mockReset();
  });

  it("既定 concurrency で並列アップロードし全結果を返す", async () => {
    let inFlight = 0;
    let maxInFlight = 0;
    vi.mocked(axios.post).mockImplementation(async () => {
      inFlight += 1;
      maxInFlight = Math.max(maxInFlight, inFlight);
      await new Promise((resolve) => setTimeout(resolve, 15));
      inFlight -= 1;
      return { data: { id: maxInFlight } };
    });

    const files = Array.from({ length: 6 }, (_, i) => new File([`x${i}`], `f${i}.jpg`, { type: "image/jpeg" }));
    const results = await uploadMedicalRecordImages("99", files);

    expect(results).toHaveLength(6);
    expect(axios.post).toHaveBeenCalledTimes(6);
    expect(maxInFlight).toBeLessThanOrEqual(MEDICAL_RECORD_IMAGE_UPLOAD_CONCURRENCY);
    expect(maxInFlight).toBe(MEDICAL_RECORD_IMAGE_UPLOAD_CONCURRENCY);
  });
});
