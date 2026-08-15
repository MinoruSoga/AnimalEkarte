import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import type { Hospitalization } from "@/types";
import { HospitalizationListView } from "./HospitalizationListView";

function makeHospitalization(overrides: Partial<Hospitalization> = {}): Hospitalization {
  return {
    id: "h-1",
    hospitalizationNo: "H-001",
    ownerName: "山田太郎",
    petName: "ポチ",
    species: "犬",
    hospitalizationType: "入院",
    startDate: "2026-07-13",
    endDate: "",
    status: "入院中",
    cageId: "cage-1",
    petId: "pet-1",
    doctorId: undefined,
    doctorName: undefined,
    petIsDeceased: false,
    memo: undefined,
    ownerRequest: undefined,
    staffNotes: undefined,
    ...overrides,
  };
}

function renderList(hospitalization: Hospitalization, canEdit = false) {
  const onNavigate = vi.fn();
  render(
    <MemoryRouter initialEntries={["/hospitalization"]}>
      <HospitalizationListView
        hospitalizations={[hospitalization]}
        onNavigate={onNavigate}
        canEdit={canEdit}
      />
    </MemoryRouter>,
  );

  return { onNavigate };
}

describe("HospitalizationListView row navigation accessibility", () => {
  it("編集権限がなくても入院No・ペット名・ID付き44px native detail linkを表示する", () => {
    renderList(makeHospitalization(), false);

    const detailLink = screen.getByRole("link", { name: /H-001/ });
    expect(detailLink).toHaveAttribute("href", "/hospitalization/h-1");
    expect(detailLink).toHaveAccessibleName(/ポチ/);
    expect(detailLink).toHaveAccessibleName(/h-1/);
    expect(detailLink).toHaveClass("min-h-11", "min-w-11");
  });

  it("detail link以外のセルclickでは編集callbackを発火しない", () => {
    const { onNavigate } = renderList(makeHospitalization(), true);

    fireEvent.click(screen.getByText("山田太郎"));

    expect(onNavigate).not.toHaveBeenCalled();
  });

  it("死亡ペット行はdetail linkと編集操作を表示しない", () => {
    const { onNavigate } = renderList(makeHospitalization({ petIsDeceased: true }), true);

    expect(screen.queryByRole("link")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /H-001/ })).not.toBeInTheDocument();
    expect(screen.getByText("死亡")).toBeInTheDocument();

    fireEvent.click(screen.getByText("山田太郎"));
    expect(onNavigate).not.toHaveBeenCalled();
  });
});
