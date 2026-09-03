import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";

import type { LabDeviceJobCard } from "@/hooks/use-lab-device-unlinked";

import { LabDeviceUnlinkedBanner } from "./LabDeviceUnlinkedBanner";

const { attachMutate, detachMutate, unlinkedCards, permission } = vi.hoisted(() => ({
  attachMutate: vi.fn(),
  detachMutate: vi.fn(),
  unlinkedCards: { current: [] as LabDeviceJobCard[] },
  permission: {
    canView: true,
    canCreate: false,
    canEdit: true,
    canDelete: false,
  },
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => permission,
}));

vi.mock("@/hooks/use-lab-device-unlinked", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/hooks/use-lab-device-unlinked")>();
  return {
    ...actual,
    useGetLabDeviceUnlinked: () => ({ data: unlinkedCards.current }),
    useAttachLabDeviceJob: () => ({ mutate: attachMutate }),
    useDetachLabDeviceJob: () => ({ mutate: detachMutate }),
  };
});

function makeCard(overrides: Partial<LabDeviceJobCard> = {}): LabDeviceJobCard {
  return {
    jobId: "job-1",
    sourceType: "fuji_nx600",
    deviceHint: "NX600-A",
    status: "received",
    specimenIdRaw: "S-1",
    itemCount: 1,
    unmappedItemCount: 0,
    clockSkew: false,
    items: [],
    ...overrides,
  };
}

function persistedCard(overrides: Partial<LabDeviceJobCard> = {}): LabDeviceJobCard {
  return makeCard({
    status: "persisted",
    petId: 42,
    ...overrides,
  });
}

describe("LabDeviceUnlinkedBanner (FE-RC-107)", () => {
  beforeEach(() => {
    attachMutate.mockReset();
    detachMutate.mockReset();
    unlinkedCards.current = [makeCard()];
    permission.canView = true;
    permission.canEdit = true;
  });

  it("付け失敗時は attachError を出し justAttached を残さない", () => {
    attachMutate.mockImplementation(
      (_vars: { jobId: string; petId: number }, options?: { onError?: (error: Error) => void }) => {
        options?.onError?.(new Error("attach failed"));
      },
    );

    render(<LabDeviceUnlinkedBanner petId="42" />);
    fireEvent.click(screen.getByRole("button", { name: "付ける" }));

    expect(screen.getByText("保存できませんでした。未紐付けのままです")).toBeInTheDocument();
    expect(screen.queryByText(/付けました/)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "付ける" })).toBeInTheDocument();
  });

  it("付け成功後の取り消し失敗では justAttached を残す", () => {
    attachMutate.mockImplementation(
      (
        _vars: { jobId: string; petId: number },
        options?: { onSuccess?: (card: LabDeviceJobCard) => void },
      ) => {
        options?.onSuccess?.(persistedCard());
      },
    );
    detachMutate.mockImplementation(
      (_jobId: string, options?: { onError?: (error: Error) => void }) => {
        options?.onError?.(new Error("detach failed"));
      },
    );

    render(<LabDeviceUnlinkedBanner petId="42" />);
    fireEvent.click(screen.getByRole("button", { name: "付ける" }));
    expect(screen.getByText(/付けました/)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "取り消す" }));
    expect(screen.getByText(/付けました/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "取り消す" })).toBeInTheDocument();
  });

  it("付け成功後の別ジョブ付け失敗では justAttached を消し attachError を出す", () => {
    unlinkedCards.current = [makeCard(), makeCard({ jobId: "job-2", deviceHint: "AU10V-B" })];
    attachMutate
      .mockImplementationOnce(
        (
          _vars: { jobId: string; petId: number },
          options?: { onSuccess?: (card: LabDeviceJobCard) => void },
        ) => {
          options?.onSuccess?.(persistedCard());
        },
      )
      .mockImplementationOnce(
        (
          _vars: { jobId: string; petId: number },
          options?: { onError?: (error: Error) => void },
        ) => {
          options?.onError?.(new Error("attach failed"));
        },
      );

    render(<LabDeviceUnlinkedBanner petId="42" />);
    fireEvent.click(screen.getAllByRole("button", { name: "付ける" })[0]);
    expect(screen.getByText(/付けました/)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "付ける" }));
    expect(screen.queryByText(/付けました/)).not.toBeInTheDocument();
    expect(screen.getByText("保存できませんでした。未紐付けのままです")).toBeInTheDocument();
  });
});
