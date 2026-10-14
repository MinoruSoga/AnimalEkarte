import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { createTestWrapper } from "@/testing/TestUtils";
import { CampaignSettings } from "./CampaignSettings";

const { query } = vi.hoisted(() => ({
  query: {
    current: {
      data: undefined as
        | undefined
        | {
            id: string;
            name: string;
            startDate: string;
            endDate: string;
            discountType: "rate";
            discountValue: number;
            isActive: boolean;
          }[],
      isPending: true,
      isError: false,
    },
  },
}));
vi.mock("@/hooks/use-auth", () => ({ useAuth: () => ({ hasPermission: () => true }) }));
vi.mock("@/hooks/use-side-peek-dirty", () => ({
  useSidePeekDirty: () => ({
    discardDialog: null,
    markDirty: vi.fn(),
    markClean: vi.fn(),
    runWithDiscardCheck: (action: () => void) => action(),
  }),
}));
vi.mock("../api/campaign", () => ({
  useGetCampaigns: () => query.current,
  useCreateCampaign: () => ({ mutateAsync: vi.fn() }),
  useUpdateCampaign: () => ({ mutateAsync: vi.fn() }),
  useDeleteCampaign: () => ({ mutateAsync: vi.fn() }),
}));

describe("CampaignSettings count states", () => {
  it("loading is not zero; loaded actual count and genuine zero remain visible", () => {
    query.current = { data: undefined, isPending: true, isError: false };
    const view = render(<CampaignSettings />, { wrapper: createTestWrapper({ router: true }) });
    expect(screen.queryByText("0 件")).not.toBeInTheDocument();
    query.current = {
      data: [
        {
          id: "1",
          name: "合成割引",
          startDate: "2026-10-01",
          endDate: "2026-10-31",
          discountType: "rate",
          discountValue: 10,
          isActive: true,
        },
      ],
      isPending: false,
      isError: false,
    };
    view.rerender(<CampaignSettings />);
    expect(screen.getByText("1 件")).toBeInTheDocument();
    query.current = { data: [], isPending: false, isError: false };
    view.rerender(<CampaignSettings />);
    expect(screen.getByText("0 件")).toBeInTheDocument();
  });
});
