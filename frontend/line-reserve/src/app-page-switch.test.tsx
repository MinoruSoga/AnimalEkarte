import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ReservationPageSwitch, type ReservationPageSwitchProps } from "./app-page-switch";
import type { LiffSettings, ReservationFlow } from "./types/models";
import type { RevisitIntakeDraftStore } from "./hooks/use-revisit-intake-drafts";

const settings = {
  liff_id: "2000000000-abcdefgh",
  status: "running",
} as unknown as LiffSettings;

const flow = {
  courseCategory: "normal",
  customerInfo: { name: "山田", phone: "090", ownerName: "", pets: [] },
  courseName: "再診",
  courseId: null,
  staffId: null,
  staffName: "",
  date: "2026-08-01",
  startTime: "1000",
  endTime: "1030",
  requestText: "",
  trimmingCourseId: null,
  trimmingCourseName: "",
  trimmingOptionIds: [],
} as unknown as ReservationFlow;

const draftStore = {
  get: vi.fn(),
  set: vi.fn(),
  clear: vi.fn(),
  hasDraft: vi.fn(() => false),
  findPending: vi.fn(() => null),
} as unknown as RevisitIntakeDraftStore;

function makeProps(
  completedReservationId: string,
  handleOpenRevisitIntake: ReservationPageSwitchProps["handleOpenRevisitIntake"],
): ReservationPageSwitchProps {
  return {
    page: "step8",
    settings,
    profile: null,
    clinicId: "1",
    idToken: "token",
    flow,
    isTrimming: false,
    completedReservationId,
    createResultUnknown: false,
    completedNotes: "memo",
    revisitIntake: null,
    revisitIntakeDraftStore: draftStore,
    goTo: vi.fn(),
    setCustomerInfo: vi.fn(),
    setCourse: vi.fn(),
    setStaff: vi.fn(),
    setDate: vi.fn(),
    setTime: vi.fn(),
    setRequestText: vi.fn(),
    setTrimmingCourse: vi.fn(),
    setTrimmingOptions: vi.fn(),
    handleNewReservation: vi.fn(),
    handleConfirm: vi.fn(),
    handleResultUnknown: vi.fn(),
    handleSlotTaken: vi.fn(),
    handleCompleteToMyReservations: vi.fn(),
    handleCompleteNewReservation: vi.fn(),
    handleOpenRevisitIntake,
  };
}

describe("ReservationPageSwitch step8 → 再診問診", () => {
  it.each(["1", "9007199254740991", "9007199254740992", "9007199254740993", "9223372036854775807"])(
    "EMR-519: completedReservationId %s を Number 変換せず appointmentId へ渡す",
    async (id) => {
      const user = userEvent.setup();
      const onOpen = vi.fn();
      render(<ReservationPageSwitch {...makeProps(id, onOpen)} />);

      await user.click(screen.getByRole("button", { name: "事前問診に回答する（任意）" }));
      expect(onOpen).toHaveBeenCalledWith(
        expect.objectContaining({ appointmentId: id, returnTo: "step8" }),
      );
    },
  );
});
