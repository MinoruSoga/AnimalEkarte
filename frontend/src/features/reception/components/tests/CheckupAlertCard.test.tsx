/**
 * FEAT-372 Phase B: CheckupAlertCard — ダッシュボード定期検査件数カード
 *
 * カバー観点:
 * 1. API 成功時に overdue / upcoming の件数が正しく表示される
 * 2. 0 件時は "0件" を表示する
 * 3. ローディング中は "--" を表示する
 * 4. API エラー時は "取得失敗" を表示する
 * 5. クリックで onClick コールバックが呼ばれる
 */
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CheckupAlertCard } from "../CheckupAlertCard";
import type { CheckupAlertCategory } from "@/features/checkups";

const makeCategory = (count: number): CheckupAlertCategory => ({
  count,
  items: [],
});

describe("CheckupAlertCard — overdue variant", () => {
  it("件数を正しく表示する", () => {
    render(
      <CheckupAlertCard
        variant="overdue"
        data={makeCategory(5)}
        isLoading={false}
        isError={false}
        onClick={vi.fn()}
      />,
    );
    expect(screen.getByTestId("checkup-alert-overdue-count").textContent).toBe("5件");
  });

  it("0 件のとき '0件' を表示する", () => {
    render(
      <CheckupAlertCard
        variant="overdue"
        data={makeCategory(0)}
        isLoading={false}
        isError={false}
        onClick={vi.fn()}
      />,
    );
    expect(screen.getByTestId("checkup-alert-overdue-count").textContent).toBe("0件");
  });

  it("ローディング中は '--' を表示する", () => {
    render(
      <CheckupAlertCard
        variant="overdue"
        data={undefined}
        isLoading={true}
        isError={false}
        onClick={vi.fn()}
      />,
    );
    expect(screen.getByTestId("checkup-alert-overdue-count").textContent).toBe("--");
  });

  it("API エラー時は '取得失敗' を表示する", () => {
    render(
      <CheckupAlertCard
        variant="overdue"
        data={undefined}
        isLoading={false}
        isError={true}
        onClick={vi.fn()}
      />,
    );
    expect(screen.getByTestId("checkup-alert-overdue-count").textContent).toBe("取得失敗");
  });

  it("クリックで onClick が呼ばれる", async () => {
    const onClick = vi.fn();
    render(
      <CheckupAlertCard
        variant="overdue"
        data={makeCategory(3)}
        isLoading={false}
        isError={false}
        onClick={onClick}
      />,
    );
    await userEvent.click(screen.getByTestId("checkup-alert-overdue"));
    expect(onClick).toHaveBeenCalledOnce();
  });
});

describe("CheckupAlertCard — upcoming variant", () => {
  it("件数を正しく表示する", () => {
    render(
      <CheckupAlertCard
        variant="upcoming"
        data={makeCategory(12)}
        isLoading={false}
        isError={false}
        onClick={vi.fn()}
      />,
    );
    expect(screen.getByTestId("checkup-alert-upcoming-count").textContent).toBe("12件");
  });

  it("0 件のとき '0件' を表示する", () => {
    render(
      <CheckupAlertCard
        variant="upcoming"
        data={makeCategory(0)}
        isLoading={false}
        isError={false}
        onClick={vi.fn()}
      />,
    );
    expect(screen.getByTestId("checkup-alert-upcoming-count").textContent).toBe("0件");
  });

  it("クリックで onClick が呼ばれる", async () => {
    const onClick = vi.fn();
    render(
      <CheckupAlertCard
        variant="upcoming"
        data={makeCategory(7)}
        isLoading={false}
        isError={false}
        onClick={onClick}
      />,
    );
    await userEvent.click(screen.getByTestId("checkup-alert-upcoming"));
    expect(onClick).toHaveBeenCalledOnce();
  });
});
