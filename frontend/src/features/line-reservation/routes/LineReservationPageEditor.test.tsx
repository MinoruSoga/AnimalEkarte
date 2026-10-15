import { beforeEach, describe, expect, it, vi } from "vitest";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { toast } from "sonner";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { LineReservationPageEditor } from "./LineReservationPageEditor";
import type { ReservationSetting } from "../api/types";

const CLINIC_ID = "clinic-test-1";

const permission = vi.hoisted(() => ({
  current: { canView: true, canCreate: true, canEdit: true, canDelete: false },
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => permission.current,
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ currentClinicId: CLINIC_ID }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock("@/components/shared/PageLayout/PageLayout", () => ({
  PageLayout: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

const baseSetting: ReservationSetting = {
  id: 1,
  clinic_id: 1,
  status: "running",
  header_text: "ヘッダー",
  reservation_notice: "",
  cancel_notice: "",
  privacy_policy: "",
  closed_weekdays: [1] as unknown as string,
  closed_dates: [] as unknown as string,
  national_holiday_closed: false,
  business_hours: { start: "0900", end: "1900" } as unknown as string,
  business_hours_by_weekday: {} as unknown as string,
  break_hours: [] as unknown as string,
  booking_window_max_days: 30,
  booking_window_min_days: 0,
  calendar_months: 2,
  phone_number: "",
  notification_email: "",
  request_example: "",
  time_slot_mode: "minimize_gaps",
  time_slot_interval_minutes: 15,
  no_staff_mode: "hide",
  show_no_staff_option: false,
  additional_fields: {} as unknown as string,
  line_channel_id: "existing-channel-id",
  liff_id: "existing-liff-id",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

const pageEditorSetting = vi.hoisted(() => ({ current: null as ReservationSetting | null }));

vi.mock("../api/get-line-reservation-setting", () => ({
  useGetLineReservationSetting: () => ({
    data: pageEditorSetting.current ?? baseSetting,
    isLoading: false,
  }),
}));

function setupPutHandler() {
  let capturedBody: Record<string, unknown> | null = null;
  server.use(
    http.put(`/api/v1/clinics/${CLINIC_ID}/line-reservation-settings`, async ({ request }) => {
      capturedBody = (await request.json()) as Record<string, unknown>;
      return HttpResponse.json({ ...baseSetting, ...capturedBody });
    }),
  );
  return () => capturedBody;
}

function renderEditor(setting: ReservationSetting = baseSetting) {
  pageEditorSetting.current = setting;
  return render(<LineReservationPageEditor />, {
    wrapper: createTestWrapper({ router: true }),
  });
}

describe("LineReservationPageEditor — FE-RC-211 mutation permission re-check", () => {
  beforeEach(() => {
    permission.current = { canView: true, canCreate: true, canEdit: true, canDelete: false };
    pageEditorSetting.current = null;
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.error).mockClear();
  });

  it("canEdit=false では PUT せず toast.error する", async () => {
    permission.current = { canView: true, canCreate: true, canEdit: false, canDelete: false };
    const getBody = setupPutHandler();
    renderEditor();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
    });
    expect(getBody()).toBeNull();
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("canEdit=true では PUT して数値の closed_weekdays を保持する", async () => {
    const getBody = setupPutHandler();
    renderEditor();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(getBody()).not.toBeNull();
    });
    expect(getBody()?.closed_weekdays).toEqual([1]);
    expect(toast.success).toHaveBeenCalledWith("ページ内容を保存しました");
    expect(toast.error).not.toHaveBeenCalled();
  });

  it.each([
    {
      name: "actual legacy string weekdays",
      closedWeekdays: ["0", "6"] as unknown as string,
      expected: [0, 6],
    },
    {
      name: "serialized numeric weekdays",
      closedWeekdays: "[0,6]",
      expected: [0, 6],
    },
    {
      name: "serialized legacy string weekdays",
      closedWeekdays: '["0","6"]',
      expected: [0, 6],
    },
    {
      name: "mixed numeric and legacy string weekdays",
      closedWeekdays: [0, "2", 6] as unknown as string,
      expected: [0, 2, 6],
    },
  ])("round-trips $name as canonical numbers", async ({ closedWeekdays, expected }) => {
    const getBody = setupPutHandler();
    renderEditor({
      ...baseSetting,
      closed_weekdays: closedWeekdays,
      closed_dates: ["2026-12-30"] as unknown as string,
    });

    const user = userEvent.setup();
    await user.clear(screen.getByRole("textbox", { name: "ヘッダーテキスト" }));
    await user.type(screen.getByRole("textbox", { name: "ヘッダーテキスト" }), "更新ヘッダー");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(getBody()).not.toBeNull();
    });
    const body = getBody()!;
    expect(body.header_text).toBe("更新ヘッダー");
    expect(body.closed_weekdays).toEqual(expected);
    expect(body.closed_dates).toEqual(["2026-12-30"]);
    expect(body).not.toHaveProperty("line_access_token");
    expect(body).not.toHaveProperty("line_channel_secret");
    expect(toast.success).toHaveBeenCalledWith("ページ内容を保存しました");
  });

  it("invalid closed_weekdays remain visible and block page content saves", async () => {
    const getBody = setupPutHandler();
    renderEditor({ ...baseSetting, closed_weekdays: "[7]" });

    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent(
      "定休日の設定を読み込めません。内容を確認して修正してください。",
    );
    const saveButton = screen.getByRole("button", { name: "保存" });
    expect(saveButton).toBeDisabled();

    const user = userEvent.setup();
    await user.click(saveButton);
    const form = saveButton.closest("form");
    expect(form).not.toBeNull();
    if (form) await act(async () => fireEvent.submit(form));

    expect(screen.getByRole("alert")).toBe(alert);
    expect(getBody()).toBeNull();
  });
});
