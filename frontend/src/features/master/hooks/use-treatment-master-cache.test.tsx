import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { expect, it, vi } from "vitest";
import { axios } from "@/lib/axios";
import {
  useGetAllCheckupTypes as useReferenceCheckups,
  useGetAllVaccinesMaster as useReferenceVaccines,
} from "@/hooks/use-treatment-master";
import { useGetAllCheckupTypes as useEditorCheckups } from "../api/checkup-types";
import { useGetAllVaccinesMaster as useEditorVaccines } from "../api/vaccines-master";
vi.mock("@/lib/axios", () => ({ axios: { get: vi.fn() } }));

type InsuranceCacheResult = {
  isSuccess: boolean;
  data?: {
    isNonInsurance?: boolean | null;
    insuranceSetting: boolean | null;
    version?: number | null;
  }[];
};

it.each(["reference-first", "editor-first"])(
  "preserves nullable insurance and editing version in either cache order: %s",
  async (order) => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
    vi.mocked(axios.get).mockResolvedValue({
      data: [
        { id: 1, name: "legacy", is_active: true, is_non_insurance: null, version: 7 },
        { id: 2, name: "insured", is_active: true, is_non_insurance: false, version: 3 },
      ],
    });
    const families: [() => InsuranceCacheResult, () => InsuranceCacheResult][] = [
      [useReferenceCheckups, useEditorCheckups],
      [useReferenceVaccines, useEditorVaccines],
    ];
    for (const [read, edit] of families) {
      const first = renderHook(order === "reference-first" ? read : edit, { wrapper });
      await waitFor(() => expect(first.result.current.isSuccess).toBe(true));
      first.unmount();
      const second = renderHook(order === "reference-first" ? edit : read, { wrapper });
      await waitFor(() => expect(second.result.current.isSuccess).toBe(true));
      expect(second.result.current.data?.[0]).toMatchObject({ insuranceSetting: null });
      expect(second.result.current.data?.[1].isNonInsurance).toBe(false);
      if (read === useReferenceCheckups)
        expect(second.result.current.data?.[0]).toMatchObject({
          insuranceSetting: null,
          version: 7,
        });
      second.unmount();
    }
    client.clear();
  },
);
