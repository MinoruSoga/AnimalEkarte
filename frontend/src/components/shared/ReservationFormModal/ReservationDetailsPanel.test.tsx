import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import type { Reservation } from "@/types";

import { ReservationDetailsPanel } from "./ReservationFormModalPanels";

/**
 * Codex PR606 #12: record_intent は BE provenance（カルテ由来 tentative の昇格経路）で
 * ユーザー選択肢ではない。編集モーダルで ReservationRouteSelect（PATCH 可能な
 * セレクタ）に出すと immutable な経路を上書きできてしまうため、読み取り専用
 * ラベルで表示する分岐を固定する。
 */

// 経路セレクタは PATCH を内蔵するため、このテストでは分岐の有無だけを見る。
const routeSelectSpy = vi.hoisted(() => vi.fn(() => <div data-testid="route-select" />));
vi.mock("@/components/shared/ReservationRouteSelect", () => ({
  ReservationRouteSelect: (props: Record<string, unknown>) => {
    routeSelectSpy(props);
    return <div data-testid="route-select" />;
  },
}));

// フォーム本体はこのテストの対象外（record_intent 分岐のみ検証）。
vi.mock("./ReservationFormFields", () => ({
  ReservationFormFields: () => <div data-testid="form-fields" />,
}));

const noop = () => {};

function renderPanel(reservationRoute: string | undefined) {
  return render(
    <ReservationDetailsPanel
      mobilePanel="form"
      selectedPets={[]}
      lstepStatus={undefined}
      formData={{ reservationRoute } as Partial<Reservation>}
      validationErrors={{}}
      holidayDates={new Set()}
      isEditMode={true}
      reservationId="42"
      canEdit={true}
      onSelectedPetsChange={noop}
      onFormChange={noop}
      onClearError={noop}
      onMonthChange={noop}
    />,
  );
}

describe("ReservationDetailsPanel — record_intent 経路の読み取り専用表示（Codex #12）", () => {
  it("reservationRoute=record_intent はセレクタではなく読み取り専用ラベルを出す", () => {
    renderPanel("record_intent");

    expect(screen.getByTestId("reservation-route-readonly")).toHaveTextContent(
      "次回予定（カルテ由来）",
    );
    expect(screen.queryByTestId("route-select")).not.toBeInTheDocument();
    expect(routeSelectSpy).not.toHaveBeenCalled();
  });

  it("通常の経路値（walk_in 等）は従来どおりセレクタを出す", () => {
    renderPanel("walk_in");

    expect(screen.getByTestId("route-select")).toBeInTheDocument();
    expect(screen.queryByTestId("reservation-route-readonly")).not.toBeInTheDocument();
  });
});
