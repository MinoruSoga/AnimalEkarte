import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { toast } from "sonner";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { LineReservationSettingsForm } from "./LineReservationSettingsForm";
import type { ReservationSetting } from "../api/types";

const permission = vi.hoisted(() => ({
  current: { canView: true, canCreate: true, canEdit: true, canDelete: false },
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => permission.current,
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

const CLINIC_ID = "clinic-test-1";

const baseSetting: ReservationSetting = {
  id: 1,
  clinic_id: 1,
  status: "running",
  header_text: "",
  reservation_notice: "",
  cancel_notice: "",
  privacy_policy: "",
  // 実行時は JSON.parse 済みのオブジェクト/配列として届く（line-reservation-settings-form-model.ts の asJsonb 参照）
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

function setupPutHandler(responseOverrides: Partial<ReservationSetting> = {}) {
  let capturedBody: Record<string, unknown> | null = null;
  server.use(
    http.put(`/api/v1/clinics/${CLINIC_ID}/line-reservation-settings`, async ({ request }) => {
      capturedBody = (await request.json()) as Record<string, unknown>;
      return HttpResponse.json({ ...baseSetting, ...capturedBody, ...responseOverrides });
    }),
  );
  return () => capturedBody;
}

function renderForm(setting: ReservationSetting = baseSetting) {
  return render(<LineReservationSettingsForm setting={setting} clinicId={CLINIC_ID} />, {
    wrapper: createTestWrapper({ router: true }),
  });
}

beforeEach(() => {
  permission.current = { canView: true, canCreate: true, canEdit: true, canDelete: false };
  vi.mocked(toast.success).mockClear();
  vi.mocked(toast.error).mockClear();
});

function setupClinicHolidayWriteSpies() {
  const posts: unknown[] = [];
  const deletes: string[] = [];
  server.use(
    http.post("/api/v1/clinic-holidays", async ({ request }) => {
      posts.push(await request.json());
      return HttpResponse.json(
        { id: 99, clinic_id: 1, date: "2026-08-11", reason: "" },
        { status: 201 },
      );
    }),
    http.delete("/api/v1/clinic-holidays/:date", ({ params }) => {
      deletes.push(String(params.date));
      return new HttpResponse(null, { status: 204 });
    }),
  );
  return { posts, deletes };
}

// SD-3 決裁 A（q&a.html）: LINE credential（チャネルシークレット/アクセストークン）は
// 平文 UI に置かない。この画面はそもそも credential を扱わないため、
// 対応する input・formData 読み取り・payload キーのいずれも存在してはならない。
describe("LineReservationSettingsForm — LINE credential 非取扱い (SD-3 決裁A)", () => {
  it("定休曜日checkboxのfocusable hit areaを44px以上に保つ", () => {
    renderForm();

    const weekdayCheckboxes = Array.from(
      document.querySelectorAll<HTMLElement>('[role="checkbox"][id^="closed-weekday-"]'),
    );

    expect(weekdayCheckboxes).toHaveLength(7);
    weekdayCheckboxes.forEach((checkbox) => {
      expect(checkbox).toHaveClass("size-11");
    });
  });

  it("チャネルシークレット・アクセストークンの input は画面に存在しない", () => {
    renderForm();

    expect(screen.queryByPlaceholderText(/チャネルシークレット/)).not.toBeInTheDocument();
    expect(screen.queryByPlaceholderText(/アクセストークン/)).not.toBeInTheDocument();
    expect(document.querySelector('input[name="line_channel_secret"]')).not.toBeInTheDocument();
    expect(document.querySelector('input[name="line_access_token"]')).not.toBeInTheDocument();
  });

  it("保存すると PUT body に line_channel_secret / line_access_token キー自体が含まれない", async () => {
    const getBody = setupPutHandler();
    renderForm();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "設定を保存" }));

    await waitFor(() => {
      expect(getBody()).not.toBeNull();
    });
    const body = getBody() as Record<string, unknown>;
    expect(Object.prototype.hasOwnProperty.call(body, "line_channel_secret")).toBe(false);
    expect(Object.prototype.hasOwnProperty.call(body, "line_access_token")).toBe(false);
  });

  it("チャネルID・LIFF ID は引き続き通常の input として編集・送信できる（credential ではないため対象外）", async () => {
    const getBody = setupPutHandler();
    renderForm();

    expect(screen.getByPlaceholderText("LINE チャネルID")).toHaveValue("existing-channel-id");
    expect(screen.getByPlaceholderText("LIFF ID")).toHaveValue("existing-liff-id");

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "設定を保存" }));

    await waitFor(() => {
      expect(getBody()).not.toBeNull();
    });
    expect(getBody()?.line_channel_id).toBe("existing-channel-id");
    expect(getBody()?.liff_id).toBe("existing-liff-id");
  });
});

describe("LineReservationSettingsForm — 個別定休日の二重入力削除", () => {
  it("特定定休日の日付入力と追加UIを出さず、シフト管理カレンダーへの案内だけを表示する", () => {
    const setting: ReservationSetting = {
      ...baseSetting,
      closed_dates: ["2026-08-11"] as unknown as string,
    };
    renderForm(setting);

    expect(screen.queryByLabelText("特定定休日 1")).not.toBeInTheDocument();
    expect(screen.queryByDisplayValue("2026-08-11")).not.toBeInTheDocument();
    expect(screen.queryByText("特定定休日は設定されていません")).not.toBeInTheDocument();

    const closedDatesRow = screen.getByText("特定定休日").parentElement;
    expect(closedDatesRow).not.toBeNull();
    expect(
      within(closedDatesRow as HTMLElement).queryByRole("button", { name: "+ 追加" }),
    ).not.toBeInTheDocument();

    const shiftLink = screen.getByRole("link", { name: "シフト管理" });
    expect(shiftLink).toHaveAttribute("href", "/shifts");
    expect(screen.getByRole("switch", { name: "祝日休診" })).toBeInTheDocument();
    expect(screen.getByLabelText("月")).toBeInTheDocument();
  });

  it("保存時は closed_weekdays と national_holiday_closed を含み、clinic-holidays へ POST しない", async () => {
    const getBody = setupPutHandler();
    const holidayWrites = setupClinicHolidayWriteSpies();
    renderForm();

    const user = userEvent.setup();
    await user.click(screen.getByLabelText("月"));
    await user.click(screen.getByRole("switch", { name: "祝日休診" }));
    await user.click(screen.getByRole("button", { name: "設定を保存" }));

    await waitFor(() => {
      expect(getBody()).not.toBeNull();
    });
    const body = getBody() as Record<string, unknown>;
    expect(body.closed_weekdays).toEqual(["1"]);
    expect(body.national_holiday_closed).toBe(true);
    expect(holidayWrites.posts).toEqual([]);
    expect(holidayWrites.deletes).toEqual([]);
  });

  it("既存 closed_dates は PUT で round-trip し、UI から新しい日付を追加できない", async () => {
    const existingDates = ["2026-08-11", "2026-12-30"];
    const setting: ReservationSetting = {
      ...baseSetting,
      closed_dates: existingDates as unknown as string,
    };
    const getBody = setupPutHandler();
    const holidayWrites = setupClinicHolidayWriteSpies();
    renderForm(setting);

    expect(screen.queryByLabelText("特定定休日 1")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("特定定休日 2")).not.toBeInTheDocument();
    expect(screen.queryByDisplayValue("2026-08-11")).not.toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "設定を保存" }));

    await waitFor(() => {
      expect(getBody()).not.toBeNull();
    });
    const body = getBody() as Record<string, unknown>;
    expect(body.closed_dates).toEqual(existingDates);
    expect(holidayWrites.posts).toEqual([]);
    expect(holidayWrites.deletes).toEqual([]);
  });
});

// BUG-028: 最短予約受付（日数）を 0 含む新値で保存後、form action 完了後も UI が新値のまま
describe("LineReservationSettingsForm — booking_window_min_days UI sync (BUG-028)", () => {
  it("最短予約受付を0に変更して保存すると、保存後も入力欄が0のまま残る", async () => {
    const initial: ReservationSetting = { ...baseSetting, booking_window_min_days: 2 };
    const getBody = setupPutHandler({ booking_window_min_days: 0 });
    renderForm(initial);

    const input = screen.getByRole("spinbutton", { name: "最短予約受付（日数）" });
    expect(input).toHaveValue(2);

    const user = userEvent.setup();
    await user.clear(input);
    await user.type(input, "0");
    expect(input).toHaveValue(0);

    await user.click(screen.getByRole("button", { name: "設定を保存" }));

    await waitFor(() => {
      expect(getBody()).not.toBeNull();
    });
    expect(getBody()?.booking_window_min_days).toBe(0);

    // form action 完了後も controlled 状態が維持され、古い defaultValue(2) に戻らない
    await waitFor(() => {
      expect(screen.getByRole("spinbutton", { name: "最短予約受付（日数）" })).toHaveValue(0);
    });
  });

  it("最短予約受付を非0の新値で保存しても入力欄が新値のまま残る", async () => {
    const initial: ReservationSetting = { ...baseSetting, booking_window_min_days: 2 };
    const getBody = setupPutHandler({ booking_window_min_days: 5 });
    renderForm(initial);

    const input = screen.getByRole("spinbutton", { name: "最短予約受付（日数）" });
    const user = userEvent.setup();
    await user.clear(input);
    await user.type(input, "5");
    await user.click(screen.getByRole("button", { name: "設定を保存" }));

    await waitFor(() => {
      expect(getBody()?.booking_window_min_days).toBe(5);
    });
    await waitFor(() => {
      expect(screen.getByRole("spinbutton", { name: "最短予約受付（日数）" })).toHaveValue(5);
    });
  });
});

describe("LineReservationSettingsForm — FE-RC-210 mutation permission re-check", () => {
  it("canEdit=false では PUT せず toast.error する", async () => {
    permission.current = { canView: true, canCreate: true, canEdit: false, canDelete: false };
    const getBody = setupPutHandler();
    renderForm();

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "設定を保存" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("この操作を行う権限がありません");
    });
    expect(getBody()).toBeNull();
    expect(toast.success).not.toHaveBeenCalled();
  });
});
