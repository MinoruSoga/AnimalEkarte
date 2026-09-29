import { describe, it, expect } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router";

import { useMedicalRecordsUrlState } from "./use-medical-records-url-state";

function Harness() {
  const { handleSortToggle, handlePageChange } = useMedicalRecordsUrlState("reset-key");
  return (
    <>
      <button type="button" onClick={() => handleSortToggle("owner_name")}>
        ソート
      </button>
      <button type="button" onClick={() => handleSortToggle("date")}>
        日付ソート
      </button>
      <button type="button" onClick={() => handlePageChange(1)}>
        ページ1
      </button>
      <button type="button" onClick={() => handlePageChange(2)}>
        ページ2
      </button>
    </>
  );
}

function renderHarness() {
  const router = createMemoryRouter([{ path: "/medical-records", element: <Harness /> }], {
    initialEntries: ["/medical-records"],
  });
  render(<RouterProvider router={router} />);
  return router;
}

const params = (router: ReturnType<typeof renderHarness>) =>
  new URLSearchParams(router.state.location.search);

// RouterProvider は router state を React.startTransition 内でコミットする一方、
// history/URL は先に更新される。実機で遷移コミット前に2回目のクリックが来る
// (medical-records-pagination-sort.spec.ts の desc 停滞)のは、1つの act に
// 2クリックを入れた場合と同じ形状になる。
describe("useMedicalRecordsUrlState", () => {
  it("遷移コミット前の連続クリックでも同一列ソートが asc へトグルする", async () => {
    const router = renderHarness();
    const sortButton = screen.getByRole("button", { name: "ソート" });

    await act(async () => {
      fireEvent.click(sortButton);
      fireEvent.click(sortButton);
    });

    const next = params(router);
    expect(next.get("sort")).toBe("owner_name");
    expect(next.get("order")).toBe("asc");
  });

  it("逐次クリックでは desc → asc → 解除の順にトグルする", async () => {
    const router = renderHarness();
    const sortButton = screen.getByRole("button", { name: "ソート" });

    await act(async () => {
      fireEvent.click(sortButton);
    });
    expect(params(router).get("sort")).toBe("owner_name");
    expect(params(router).get("order")).toBe("desc");

    await act(async () => {
      fireEvent.click(sortButton);
    });
    expect(params(router).get("order")).toBe("asc");

    await act(async () => {
      fireEvent.click(sortButton);
    });
    expect(params(router).get("sort")).toBeNull();
    expect(params(router).get("order")).toBeNull();
  });

  it("page 指定中のソートは page を外して desc で開始する", async () => {
    const router = renderHarness();
    const sortButton = screen.getByRole("button", { name: "ソート" });
    const pageButton = screen.getByRole("button", { name: "ページ2" });

    await act(async () => {
      fireEvent.click(pageButton);
    });
    expect(params(router).get("page")).toBe("2");

    await act(async () => {
      fireEvent.click(sortButton);
    });
    const next = params(router);
    expect(next.get("page")).toBeNull();
    expect(next.get("sort")).toBe("owner_name");
    expect(next.get("order")).toBe("desc");
  });

  it("handlePageChange(1) は既存の page パラメータを削除する", async () => {
    const router = renderHarness();
    const page2Button = screen.getByRole("button", { name: "ページ2" });
    const page1Button = screen.getByRole("button", { name: "ページ1" });

    await act(async () => {
      fireEvent.click(page2Button);
    });
    expect(params(router).get("page")).toBe("2");

    await act(async () => {
      fireEvent.click(page1Button);
    });
    expect(params(router).get("page")).toBeNull();
  });

  it("既存 sort 指定中に別列を選ぶとその列へ差し替えて desc に戻る", async () => {
    const router = renderHarness();
    const ownerButton = screen.getByRole("button", { name: "ソート" });
    const dateButton = screen.getByRole("button", { name: "日付ソート" });

    await act(async () => {
      fireEvent.click(ownerButton);
    });
    await act(async () => {
      fireEvent.click(ownerButton);
    });
    expect(params(router).get("order")).toBe("asc");

    await act(async () => {
      fireEvent.click(dateButton);
    });
    const next = params(router);
    expect(next.get("sort")).toBe("date");
    expect(next.get("order")).toBe("desc");
  });
});
