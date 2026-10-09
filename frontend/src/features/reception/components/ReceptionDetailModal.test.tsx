import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { paths } from "@/config/paths";

import { ReceptionDetailModal } from "./ReceptionDetailModal";
import type { ReceptionAppointment } from "../api/types";

const { navigateMock } = vi.hoisted(() => ({
  navigateMock: vi.fn(),
}));

vi.mock("react-router", () => ({
  useNavigate: () => navigateMock,
}));

afterEach(() => {
  navigateMock.mockClear();
});

const baseAppointment: ReceptionAppointment = {
  id: "101",
  time: "09:45",
  visitDate: "2026-05-29",
  end: new Date(2026, 4, 29, 10, 15, 0),
  ownerName: "山田",
  petType: "犬",
  petName: "ポチ",
  visitType: "再診",
  reservationType: "一般診察",
  reservationTypeId: "1",
  reservationCategory: "general",
  isDesignated: false,
  doctor: "担当者A",
  doctorId: "33",
  petId: "",
  ownerId: "20",
  status: "checked_in",
  notes: undefined,
  source: "manual",
};

interface RenderModalOptions {
  appointment?: ReceptionAppointment;
  currentStatus?: string;
  onClose?: () => void;
  canViewVaccinations?: boolean;
  canViewOwners?: boolean;
}

function renderModal({
  appointment = baseAppointment,
  currentStatus = "受付済",
  onClose = vi.fn(),
  canViewVaccinations = false,
  canViewOwners = true,
}: RenderModalOptions = {}) {
  return render(
    <ReceptionDetailModal
      isOpen={true}
      onClose={onClose}
      appointment={appointment}
      currentStatus={currentStatus}
      canCreateMedicalRecord={true}
      canCreateAccounting={true}
      canCreateHospitalization={true}
      canViewVaccinations={canViewVaccinations}
      canViewOwners={canViewOwners}
    />,
  );
}

describe("ReceptionDetailModal", () => {
  it("petId 未確定の通常予約では appointmentId を保持してペット選択へ遷移する", () => {
    renderModal();

    fireEvent.click(screen.getByRole("button", { name: /カルテ作成/ }));

    expect(navigateMock).toHaveBeenCalledWith(
      "/medical-records/select-pet?appointmentId=101&visitDate=2026-05-29",
      { state: { from: "/?date=2026-05-29", appointmentId: "101", visitDate: "2026-05-29" } },
    );
  });

  it("petId 未確定のトリミング予約では appointmentId を保持してペット選択へ遷移する", () => {
    renderModal({
      appointment: {
        ...baseAppointment,
        id: "202",
        reservationType: "シャンプーコース",
        reservationCategory: "trimming",
      },
    });

    fireEvent.click(screen.getByRole("button", { name: /トリミングカルテ作成/ }));

    expect(navigateMock).toHaveBeenCalledWith(
      "/trimming/select-pet?appointmentId=202&visitDate=2026-05-29",
      { state: { from: "/?date=2026-05-29", appointmentId: "202", visitDate: "2026-05-29" } },
    );
  });

  it("受付済の通常予約では関連ページのカルテ導線を表示しない", () => {
    renderModal();

    expect(screen.queryByRole("button", { name: /^カルテ$/ })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /カルテ作成/ })).toBeInTheDocument();
  });

  it("診療中の通常予約では関連ページのカルテ導線から appointmentId と visitDate を保持して遷移する", () => {
    renderModal({
      currentStatus: "診療中",
      appointment: {
        ...baseAppointment,
        petId: "10",
      },
    });

    fireEvent.click(screen.getByRole("button", { name: /^カルテ$/ }));

    expect(navigateMock).toHaveBeenCalledWith(
      "/medical-records/new?petId=10&appointmentId=101&visitDate=2026-05-29",
      { state: { from: "/?date=2026-05-29", appointmentId: "101", visitDate: "2026-05-29" } },
    );
  });

  it("診療中のホテル予約では関連ページのカルテ導線を表示しない", () => {
    renderModal({
      currentStatus: "診療中",
      appointment: {
        ...baseAppointment,
        id: "404",
        reservationType: "ホテル",
        reservationCategory: "general",
        petId: "10",
      },
    });

    expect(screen.queryByRole("button", { name: /^カルテ$/ })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^入院$/ })).toBeInTheDocument();
  });

  it("ownerId がある予約は飼主詳細へ遷移してモーダルを閉じる", () => {
    const onClose = vi.fn();
    renderModal({ onClose });

    fireEvent.click(screen.getByRole("button", { name: /飼主詳細/ }));

    expect(navigateMock).toHaveBeenCalledWith("/owners/20", {
      state: { from: "/?date=2026-05-29" },
    });
    expect(onClose).toHaveBeenCalled();
  });

  it("ownerId 不在でも ownerName があれば飼主一覧の検索へ遷移する", () => {
    renderModal({
      appointment: { ...baseAppointment, ownerId: "" },
    });

    fireEvent.click(screen.getByRole("button", { name: /飼主詳細/ }));

    expect(navigateMock).toHaveBeenCalledWith("/owners?search=%E5%B1%B1%E7%94%B0", {
      state: { from: "/?date=2026-05-29" },
    });
  });

  it("ownerId・ownerName ともに不在なら飼主一覧へ遷移する", () => {
    renderModal({
      appointment: { ...baseAppointment, ownerId: "", ownerName: "" },
    });

    fireEvent.click(screen.getByRole("button", { name: /飼主詳細/ }));

    expect(navigateMock).toHaveBeenCalledWith("/owners", {
      state: { from: "/?date=2026-05-29" },
    });
  });

  it("ownerId と petId があるペット名はペット詳細へ遷移し受付日の from を残す", () => {
    const onClose = vi.fn();
    renderModal({
      onClose,
      appointment: { ...baseAppointment, ownerId: "20", petId: "10" },
    });

    const petButton = screen.getByRole("button", { name: /^ポチ$/ });
    expect(petButton).toHaveClass("min-h-11", "min-w-11");
    fireEvent.click(petButton);

    expect(paths.owners.detail.pet.getHref("20", "10")).toBe("/owners/20?pet=10");
    expect(navigateMock).toHaveBeenCalledTimes(1);
    expect(navigateMock).toHaveBeenCalledWith("/owners/20?pet=10", {
      state: { from: "/?date=2026-05-29" },
    });
    expect(onClose).toHaveBeenCalled();
  });

  it("owners の閲覧権限が無いペット名は文字のままで遷移しない", () => {
    renderModal({
      canViewOwners: false,
      appointment: { ...baseAppointment, ownerId: "20", petId: "10" },
    });

    const name = screen.getByText("ポチ");
    expect(name.tagName).toBe("SPAN");
    fireEvent.click(name);
    expect(navigateMock).not.toHaveBeenCalled();
  });

  it("ownerId が無いペット名はテキストのままで遷移しない", () => {
    renderModal({
      appointment: { ...baseAppointment, ownerId: "", petId: "10", petName: "ポチ<b>強調</b>" },
    });

    const name = screen.getByText("ポチ<b>強調</b>");
    expect(name.tagName).toBe("SPAN");
    expect(name).not.toContainHTML("<b>");
    expect(screen.queryByRole("button", { name: "ポチ<b>強調</b>" })).not.toBeInTheDocument();
    fireEvent.click(name);
    expect(navigateMock).not.toHaveBeenCalled();
  });

  it("vaccinations canView が false なら予防接種を出さない", () => {
    renderModal({
      appointment: { ...baseAppointment, petId: "10" },
      canViewVaccinations: false,
    });

    expect(screen.queryByRole("button", { name: "予防接種" })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^ポチ$/ })).toBeInTheDocument();
  });

  it("petId が無い予約では canView でも予防接種を出さない", () => {
    renderModal({ canViewVaccinations: true });

    expect(screen.queryByRole("button", { name: "予防接種" })).not.toBeInTheDocument();
  });

  it("canView の予防接種は /vaccinations?petId= へ遷移し作成ルートを開かない", () => {
    const onClose = vi.fn();
    renderModal({
      onClose,
      canViewVaccinations: true,
      appointment: { ...baseAppointment, petId: "10" },
    });

    fireEvent.click(screen.getByRole("button", { name: "予防接種" }));

    expect(navigateMock).toHaveBeenCalledTimes(1);
    expect(navigateMock).toHaveBeenCalledWith("/vaccinations?petId=10", {
      state: { from: "/?date=2026-05-29" },
    });
    const target = String(navigateMock.mock.calls[0]?.[0]);
    expect(target).toBe("/vaccinations?petId=10");
    expect(target).not.toContain("/medical-records/new");
    expect(target).not.toContain("/vaccinations/new");
    expect(onClose).toHaveBeenCalled();
  });
});
