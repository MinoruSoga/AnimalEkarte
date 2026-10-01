import { afterEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { VaccinationList } from "./VaccinationList";

// EMR-60: useFilterVaccinations / useGetVaccinations はモックせず実物を使い、
// 検索 fetch 中に isLoading → LoadingFallback で PropertyFilter が unmount される
// 回帰を再現する。周辺依存のみ VaccinationList.test.tsx と同じ形でモックする。
// PageLayout の resource → PermissionBadges → usePermission → useAuth が呼ばれるため
// AuthProvider 非依存でレンダーできるよう use-auth もモックする
// （CashRegisterHistoryPage.test.tsx と同パターン）。
vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: {
      clinics: [{ clinicId: "1", clinicName: "テスト動物病院", isMain: true }],
      clinic: { name: "テスト動物病院" },
    },
    currentClinicId: "1",
    hasPermission: () => true,
  }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  })),
}));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: vi.fn(() => ({
    data: undefined,
    isLoading: false,
    isError: false,
  })),
}));

vi.mock("../api/delete-vaccination", () => ({
  useDeleteVaccination: vi.fn(() => ({ mutate: vi.fn() })),
}));

const backendVaccination = (id: number, ownerName: string, petName: string) => ({
  id,
  pet_id: id,
  vaccine_id: 1,
  date: "2026-07-13T00:00:00+09:00",
  next_date: "2027-07-13T00:00:00+09:00",
  vaccine: { name: "混合ワクチン" },
  pet: { name: petName, owner: { name: ownerName } },
});

describe("VaccinationList 検索 (EMR-60)", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("検索語の変更中も一覧と検索欄をunmountしない", async () => {
    let releaseSearch: (() => void) | undefined;
    let searchRequestSeen = false;
    const gate = new Promise<void>((resolve) => {
      releaseSearch = resolve;
    });
    server.use(
      http.get("/api/v1/vaccinations", async ({ request }) => {
        const url = new URL(request.url);
        if (url.searchParams.get("search") === "山田") {
          searchRequestSeen = true;
          await gate;
          return HttpResponse.json({
            data: [backendVaccination(1, "山田太郎", "ポチ")],
            total: 1,
            page: 1,
            limit: HISTORY_FETCH_LIMIT,
          });
        }
        return HttpResponse.json({
          data: [
            backendVaccination(1, "山田太郎", "ポチ"),
            backendVaccination(2, "佐藤花子", "ハナ"),
          ],
          total: 2,
          page: 1,
          limit: HISTORY_FETCH_LIMIT,
        });
      }),
    );

    const user = userEvent.setup();
    render(<VaccinationList />, {
      wrapper: createTestWrapper({ initialEntries: ["/vaccinations"] }),
    });

    await screen.findByText("山田太郎");
    expect(screen.getByText("佐藤花子")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "検索" }));
    const searchInput = screen.getByPlaceholderText("飼主名、ペット名、予防接種名...");
    fireEvent.change(searchInput, { target: { value: "山田" } });
    // EMR-247: 入力では発火しない。Enter / 検索ボタンの確定操作で search が送信される。
    fireEvent.keyDown(searchInput, { key: "Enter" });

    // 検索リクエストが handler に届き、応答を保留している状態を確認する。
    await waitFor(() => expect(searchRequestSeen).toBe(true));

    // placeholderData で前回行を保持する間、検索欄と一覧見出しが unmount されない。
    expect(searchInput).toBeInTheDocument();
    expect(searchInput).toHaveValue("山田");
    expect(screen.getByRole("heading", { name: "予防接種管理" })).toBeInTheDocument();

    releaseSearch?.();

    await waitFor(() => expect(screen.queryByText("佐藤花子")).not.toBeInTheDocument());
    expect(screen.getByText("山田太郎")).toBeInTheDocument();
    expect(searchInput).toBeVisible();
  });
});
