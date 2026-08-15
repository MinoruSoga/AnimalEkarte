import { describe, it, expect } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { LineReservationSettingsForm } from "./LineReservationSettingsForm";
import type { ReservationSetting } from "../api/types";

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
    })
  );
  return () => capturedBody;
}

// SD-3 決裁 A（q&a.html）: LINE credential（チャネルシークレット/アクセストークン）は
// 平文 UI に置かない。この画面はそもそも credential を扱わないため、
// 対応する input・formData 読み取り・payload キーのいずれも存在してはならない。
describe("LineReservationSettingsForm — LINE credential 非取扱い (SD-3 決裁A)", () => {
  it("定休曜日checkboxのfocusable hit areaを44px以上に保つ", () => {
    render(<LineReservationSettingsForm setting={baseSetting} clinicId={CLINIC_ID} />);

    const weekdayCheckboxes = Array.from(
      document.querySelectorAll<HTMLElement>('[role="checkbox"][id^="closed-weekday-"]'),
    );

    expect(weekdayCheckboxes).toHaveLength(7);
    weekdayCheckboxes.forEach((checkbox) => {
      expect(checkbox).toHaveClass("size-11");
    });
  });

  it("チャネルシークレット・アクセストークンの input は画面に存在しない", () => {
    render(<LineReservationSettingsForm setting={baseSetting} clinicId={CLINIC_ID} />);

    expect(screen.queryByPlaceholderText(/チャネルシークレット/)).not.toBeInTheDocument();
    expect(screen.queryByPlaceholderText(/アクセストークン/)).not.toBeInTheDocument();
    expect(document.querySelector('input[name="line_channel_secret"]')).not.toBeInTheDocument();
    expect(document.querySelector('input[name="line_access_token"]')).not.toBeInTheDocument();
  });

  it("保存すると PUT body に line_channel_secret / line_access_token キー自体が含まれない", async () => {
    const getBody = setupPutHandler();
    render(<LineReservationSettingsForm setting={baseSetting} clinicId={CLINIC_ID} />);

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
    render(<LineReservationSettingsForm setting={baseSetting} clinicId={CLINIC_ID} />);

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

// BUG-028: 最短予約受付（日数）を 0 含む新値で保存後、form action 完了後も UI が新値のまま
describe("LineReservationSettingsForm — booking_window_min_days UI sync (BUG-028)", () => {
  it("最短予約受付を0に変更して保存すると、保存後も入力欄が0のまま残る", async () => {
    const initial: ReservationSetting = { ...baseSetting, booking_window_min_days: 2 };
    const getBody = setupPutHandler({ booking_window_min_days: 0 });
    render(<LineReservationSettingsForm setting={initial} clinicId={CLINIC_ID} />);

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
    render(<LineReservationSettingsForm setting={initial} clinicId={CLINIC_ID} />);

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
