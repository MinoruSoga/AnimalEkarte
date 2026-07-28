import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactElement } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";

import { PetSubOwnersSection } from "./PetSubOwnersSection";

const mocks = vi.hoisted(() => ({
  subOwners: {
    sub_owners: [
      {
        owner_id: 12,
        name: "山田 花子",
        name_kana: "ヤマダ ハナコ",
        relationship: "妻",
      },
    ],
  },
  metadata: { owner_id: 10, version: 3 },
  candidates: [
    { ownerId: 10, name: "山田 太郎", nameKana: "ヤマダ タロウ" },
    { ownerId: 12, name: "山田 花子", nameKana: "ヤマダ ハナコ" },
    { ownerId: 13, name: "鈴木 次郎", nameKana: "スズキ ジロウ" },
  ],
  mutateAsync: vi.fn(),
  subOwnersError: false,
}));

vi.mock("@/lib/axios", () => ({
  axios: {
    get: vi.fn(),
  },
}));

vi.mock("../api/get-pet-sub-owners", async (importOriginal) => ({
  ...(await importOriginal<
    typeof import("../api/get-pet-sub-owners")
  >()),
  useGetPetSubOwners: () => ({
    data: mocks.subOwnersError ? undefined : mocks.subOwners,
    isLoading: false,
    error: mocks.subOwnersError ? new Error("load failed") : null,
  }),
  useGetPetSubOwnerMetadata: () => ({
    data: mocks.metadata,
    isLoading: false,
    error: null,
  }),
}));

vi.mock("../api/replace-pet-sub-owners", () => ({
  useReplacePetSubOwners: () => ({
    mutateAsync: mocks.mutateAsync,
  }),
}));

const mockedGet = vi.mocked(axios.get);

function renderSection(element: ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      {element}
    </QueryClientProvider>,
  );
}

describe("PetSubOwnersSection", () => {
  beforeEach(() => {
    mockedGet.mockReset();
    mockedGet.mockResolvedValue({
      data: {
        data: mocks.candidates.map((candidate) => ({
          id: candidate.ownerId,
          owner_name: candidate.name,
          owner_name_kana: candidate.nameKana,
        })),
        total: mocks.candidates.length,
        page: 1,
        limit: 20,
      },
    });
    mocks.mutateAsync.mockReset();
    mocks.mutateAsync.mockResolvedValue(undefined);
    mocks.subOwnersError = false;
  });

  it("副飼主の追加・続柄編集・行削除ができる", async () => {
    const user = userEvent.setup();
    renderSection(<PetSubOwnersSection petId="7" canEdit />);

    await user.type(
      screen.getByRole("searchbox", { name: "副飼主を検索" }),
      "鈴木",
    );
    await waitFor(() => expect(mockedGet).toHaveBeenCalledTimes(1));

    await user.click(screen.getByRole("combobox", { name: "副飼主を追加" }));
    expect(
      screen.queryByRole("option", { name: "山田 太郎" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("option", { name: "山田 花子" }),
    ).not.toBeInTheDocument();
    await user.click(screen.getByRole("option", { name: "鈴木 次郎" }));
    expect(
      screen.getByRole("searchbox", { name: "副飼主を検索" }),
    ).toHaveValue("");

    const relationship = screen.getByRole("textbox", {
      name: "続柄（鈴木 次郎）",
    });
    expect(relationship).toHaveValue("");
    await user.type(relationship, "祖父");
    expect(relationship).toHaveValue("祖父");

    await user.click(
      screen.getByRole("button", { name: "副飼主 山田 花子を削除" }),
    );
    expect(screen.queryByText("山田 花子")).not.toBeInTheDocument();
    expect(screen.getByText("鈴木 次郎")).toBeInTheDocument();
  });

  it("全行を削除して空配列で保存できる", async () => {
    const user = userEvent.setup();
    renderSection(<PetSubOwnersSection petId="7" canEdit />);

    await user.click(
      screen.getByRole("button", { name: "副飼主 山田 花子を削除" }),
    );
    await user.click(screen.getByRole("button", { name: "副飼主を保存" }));

    await waitFor(() =>
      expect(mocks.mutateAsync).toHaveBeenCalledWith({
        petId: "7",
        request: { version: 3, sub_owners: [] },
      }),
    );
  });

  it("409 では再読み込みが必要と伝え、ローカル編集を保持する", async () => {
    mocks.mutateAsync.mockRejectedValueOnce({
      isAxiosError: true,
      response: { status: 409, data: { error: "version conflict" } },
    });
    const user = userEvent.setup();
    renderSection(<PetSubOwnersSection petId="7" canEdit />);

    const relationship = screen.getByRole("textbox", {
      name: "続柄（山田 花子）",
    });
    await user.clear(relationship);
    await user.type(relationship, "母");
    await user.click(screen.getByRole("button", { name: "副飼主を保存" }));

    expect(
      await screen.findByRole("alert"),
    ).toHaveTextContent(
      "他の端末でペット情報が変更されました。再読み込みしてから、もう一度保存してください。",
    );
    expect(relationship).toHaveValue("母");
  });

  it("400 ではサーバーのエラー内容を表示する", async () => {
    mocks.mutateAsync.mockRejectedValueOnce({
      isAxiosError: true,
      response: {
        status: 400,
        data: { error: "続柄は1〜50文字で入力してください" },
      },
    });
    const user = userEvent.setup();
    renderSection(<PetSubOwnersSection petId="7" canEdit />);

    await user.click(screen.getByRole("button", { name: "副飼主を保存" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "副飼主を保存できませんでした。続柄は1〜50文字で入力してください",
    );
  });

  it("保存成功を支援技術へ通知する", async () => {
    const user = userEvent.setup();
    renderSection(<PetSubOwnersSection petId="7" canEdit />);

    await user.click(screen.getByRole("button", { name: "副飼主を保存" }));

    expect(await screen.findByRole("status")).toHaveTextContent(
      "副飼主を保存しました",
    );
  });

  it("保存中は編集を無効化し、保存後の再編集では成功表示を隠す", async () => {
    let resolveSave: (() => void) | undefined;
    mocks.mutateAsync.mockImplementationOnce(
      () =>
        new Promise<void>((resolve) => {
          resolveSave = resolve;
        }),
    );
    const user = userEvent.setup();
    renderSection(<PetSubOwnersSection petId="7" canEdit />);

    await user.click(screen.getByRole("button", { name: "副飼主を保存" }));

    expect(
      screen.getByRole("textbox", { name: "続柄（山田 花子）" }),
    ).toBeDisabled();
    expect(
      screen.getByRole("button", { name: "副飼主 山田 花子を削除" }),
    ).toBeDisabled();
    expect(
      screen.getByRole("searchbox", { name: "副飼主を検索" }),
    ).toBeDisabled();
    expect(screen.getByRole("combobox", { name: "副飼主を追加" })).toBeDisabled();

    resolveSave?.();
    expect(await screen.findByRole("status")).toHaveTextContent(
      "副飼主を保存しました",
    );

    const relationship = screen.getByRole("textbox", {
      name: "続柄（山田 花子）",
    });
    await user.type(relationship, "（変更）");
    expect(screen.queryByText("副飼主を保存しました")).not.toBeInTheDocument();
  });

  it("副飼主一覧の取得に失敗した場合は空配列で保存しない", async () => {
    mocks.subOwnersError = true;
    const user = userEvent.setup();
    renderSection(<PetSubOwnersSection petId="7" canEdit />);

    expect(screen.getByRole("button", { name: "副飼主を保存" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "副飼主を保存" }));
    expect(mocks.mutateAsync).not.toHaveBeenCalled();
  });

  it("続柄エラーを対象入力へ関連付ける", async () => {
    const user = userEvent.setup();
    renderSection(<PetSubOwnersSection petId="7" canEdit />);

    const relationship = screen.getByRole("textbox", {
      name: "続柄（山田 花子）",
    });
    await user.clear(relationship);
    await user.click(screen.getByRole("button", { name: "副飼主を保存" }));

    const alert = await screen.findByRole("alert");
    expect(relationship).toHaveAttribute("aria-invalid", "true");
    expect(relationship).toHaveAttribute("aria-describedby", alert.id);
  });

  it("行削除後は副飼主検索へフォーカスを戻す", async () => {
    const user = userEvent.setup();
    renderSection(<PetSubOwnersSection petId="7" canEdit />);

    await user.click(
      screen.getByRole("button", { name: "副飼主 山田 花子を削除" }),
    );

    expect(
      screen.getByRole("searchbox", { name: "副飼主を検索" }),
    ).toHaveFocus();
  });

  it("連続入力を300msデバウンスし、入力文字数より少ない候補取得に抑える", async () => {
    const user = userEvent.setup();
    renderSection(<PetSubOwnersSection petId="7" canEdit />);

    const searchInput = screen.getByRole("searchbox", {
      name: "副飼主を検索",
    });
    await user.type(searchInput, "山田");

    expect(mockedGet).not.toHaveBeenCalled();
    await waitFor(() => expect(mockedGet).toHaveBeenCalledTimes(1));
    expect(mockedGet.mock.calls.length).toBeLessThan("山田".length);
    expect(mockedGet).toHaveBeenCalledWith("/v1/owners", {
      params: { search: "山田" },
    });
  });

  it("空文字と空白だけの検索では候補を取得しない", async () => {
    const user = userEvent.setup();
    renderSection(<PetSubOwnersSection petId="7" canEdit />);

    await user.type(
      screen.getByRole("searchbox", { name: "副飼主を検索" }),
      "   ",
    );

    await act(
      () => new Promise((resolve) => setTimeout(resolve, 350)),
    );
    expect(mockedGet).not.toHaveBeenCalled();
    expect(
      screen.getByRole("combobox", { name: "副飼主を追加" }),
    ).toBeDisabled();
  });
});
