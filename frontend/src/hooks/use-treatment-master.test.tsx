import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { createTestWrapper } from "@/testing/TestUtils";
import { axios } from "@/lib/axios";
import {
  useGetAllMedicinesMaster,
  useGetAllProcedures,
  useGetAllHospitalizationPlansMaster,
} from "./use-treatment-master";

vi.mock("@/lib/axios", () => ({ axios: { get: vi.fn() } }));

beforeEach(() => {
  vi.mocked(axios.get).mockReset().mockResolvedValue({ data: [] });
});

const masterQueries = [
  [useGetAllMedicinesMaster, "/v1/masters/medicines"],
  [useGetAllProcedures, "/v1/masters/procedures"],
  [useGetAllHospitalizationPlansMaster, "/v1/masters/hospitalization-plans"],
] as const;

describe("care plan master query permission gate", () => {
  it.each(masterQueries)("%s does not request until enabled", async (useMaster, endpoint) => {
    const { result, rerender } = renderHook(({ enabled }) => useMaster({ enabled }), {
      initialProps: { enabled: false },
      wrapper: createTestWrapper(),
    });
    expect(result.current.fetchStatus).toBe("idle");
    expect(axios.get).not.toHaveBeenCalled();
    rerender({ enabled: true });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(axios.get).toHaveBeenCalledExactlyOnceWith(endpoint);
  });

  it.each(masterQueries)(
    "%s keeps fetching by default for existing consumers",
    async (useMaster, endpoint) => {
      const { result } = renderHook(() => useMaster(), { wrapper: createTestWrapper() });
      await waitFor(() => expect(result.current.isSuccess).toBe(true));
      expect(axios.get).toHaveBeenCalledExactlyOnceWith(endpoint);
    },
  );
});
