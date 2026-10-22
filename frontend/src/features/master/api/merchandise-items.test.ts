import type { ReactNode } from "react";
import { createElement } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { useGetAllMerchandiseItems } from "./merchandise-items";
import { merchandiseToFormData } from "../lib/merchandise-side-panel-model";
import { buildMerchandiseUpdateRequest } from "../routes/merchandise-item-settings-model";

vi.mock("@/lib/axios", () => ({
  axios: {
    get: vi.fn(),
  },
}));

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const baseItem = {
  id: 1,
  clinic_id: 1,
  name: "税率検証品",
  category: "goods",
  unit_price: 1000,
  tax_type: "excluded",
  is_active: true,
  sort_order: 1,
  created_at: "2026-10-20T00:00:00Z",
  updated_at: "2026-10-20T00:00:00Z",
};

// EMR-537 B-3: tax_rate の字句が GET transform → form → update request まで文字列のまま伝播する。
describe("useGetAllMerchandiseItems tax_rate exact decimal (EMR-537 B-3)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it.each(["0", "0.10", "0.08333", "0.10000000000000001"])(
    "keeps tax_rate %s as the identical string through GET → form → update request",
    async (taxRate) => {
      vi.mocked(axios.get).mockResolvedValue({
        data: [{ ...baseItem, tax_rate: taxRate }],
      });
      const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });

      const { result } = renderHook(() => useGetAllMerchandiseItems(), {
        wrapper: createWrapper(queryClient),
      });
      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      const data = result.current.data;
      if (!data) throw new Error("query data is not loaded");
      expect(axios.get).toHaveBeenCalledWith("/v1/masters/merchandise-items");
      expect(data).toHaveLength(1);
      expect(data[0].taxRate).toBe(taxRate);

      const form = merchandiseToFormData(data[0]);
      expect(form.taxRate).toBe(taxRate);

      const req = buildMerchandiseUpdateRequest(form);
      expect(req.tax_rate).toBe(taxRate);
    },
  );
});
