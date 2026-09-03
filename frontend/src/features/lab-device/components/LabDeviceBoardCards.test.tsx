import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { LabDeviceJobCard } from "../api/lab-device";
import { LabDeviceJobCardView } from "./LabDeviceBoardCards";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const { mockToast } = vi.hoisted(() => ({
  mockToast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock("sonner", () => ({
  toast: mockToast,
}));

const card = (patch: Partial<LabDeviceJobCard> = {}): LabDeviceJobCard => ({
  jobId: "j1",
  sourceType: "fuji_au10v",
  deviceHint: "AU10V",
  status: "received",
  specimenIdRaw: "TEST1",
  itemCount: 1,
  unmappedItemCount: 0,
  clockSkew: false,
  items: [],
  petId: 42,
  petName: "ポチ",
  ...patch,
});

describe("LabDeviceJobCardView permissions", () => {
  beforeEach(() => {
    mockToast.error.mockReset();
  });

  it("canEdit=true なら取り消す click で onDetach する", async () => {
    const onDetach = vi.fn();
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <LabDeviceJobCardView card={card()} canEdit={true} onDetach={onDetach} />
      </MemoryRouter>,
    );

    await user.click(screen.getByRole("button", { name: "取り消す" }));

    expect(onDetach).toHaveBeenCalledTimes(1);
    expect(mockToast.error).not.toHaveBeenCalled();
  });

  it("canEdit=true ならこの子に付ける click で onAttach する", async () => {
    const onAttach = vi.fn();
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <LabDeviceJobCardView
          card={card({ petId: undefined })}
          canEdit={true}
          onAttach={onAttach}
        />
      </MemoryRouter>,
    );

    await user.click(screen.getByRole("button", { name: "この子に付ける" }));

    expect(onAttach).toHaveBeenCalledTimes(1);
    expect(mockToast.error).not.toHaveBeenCalled();
  });

  it("canEdit=false なら取り消す click で onDetach せず toast する", async () => {
    const onDetach = vi.fn();
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <LabDeviceJobCardView card={card()} canEdit={false} onDetach={onDetach} />
      </MemoryRouter>,
    );

    await user.click(screen.getByRole("button", { name: "取り消す" }));

    expect(onDetach).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("canEdit=false ならこの子に付ける click で onAttach せず toast する", async () => {
    const onAttach = vi.fn();
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <LabDeviceJobCardView
          card={card({ petId: undefined })}
          canEdit={false}
          onAttach={onAttach}
        />
      </MemoryRouter>,
    );

    await user.click(screen.getByRole("button", { name: "この子に付ける" }));

    expect(onAttach).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });
});
