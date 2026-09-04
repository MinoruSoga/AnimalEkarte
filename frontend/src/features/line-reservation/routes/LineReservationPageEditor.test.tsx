import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
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
  closed_weekdays: [] as unknown as string,
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

vi.mock("../api/get-line-reservation-setting", () => ({
  useGetLineReservationSetting: () => ({ data: baseSetting, isLoading: false }),
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

function renderEditor() {
  return render(<LineReservationPageEditor />, {
    wrapper: createTestWrapper({ router: true }),
  });
}

describe("LineReservationPageEditor — FE-RC-211 mutation permission re-check", () => {
  beforeEach(() => {
    permission.current = { canView: true, canCreate: true, canEdit: true, canDelete: false };
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.error).mockClear();
  });

  it("canEdit=false では PUT せず toast.error する", async () => {
    permission.current = { canView: true, canCreate: true, canEdit: false, canDelete: false };
    const getBody = setupPutHandler();
    renderEditor();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "変更を保存" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
    });
    expect(getBody()).toBeNull();
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("canEdit=true では PUT して成功 toast する", async () => {
    const getBody = setupPutHandler();
    renderEditor();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "変更を保存" }));

    await waitFor(() => {
      expect(getBody()).not.toBeNull();
    });
    expect(toast.success).toHaveBeenCalledWith("ページ内容を保存しました");
    expect(toast.error).not.toHaveBeenCalled();
  });
});
