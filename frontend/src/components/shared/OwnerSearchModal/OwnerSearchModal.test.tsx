import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { OwnerSearchModal } from "./OwnerSearchModal";

function renderModal(overrides?: { onSelect?: (owner: unknown) => void }) {
  return render(
    <OwnerSearchModal
      open
      onOpenChange={() => {}}
      onSelect={overrides?.onSelect ?? (() => {})}
    />,
  );
}

async function search(term: string) {
  const user = userEvent.setup();
  await user.type(screen.getByPlaceholderText("飼主名 / 飼主No / 電話番号"), term);
  await user.click(screen.getByRole("button", { name: "検索" }));
}

describe("OwnerSearchModal", () => {
  it("検索前は「検索してください」を表示する", () => {
    renderModal();
    expect(screen.getByText("検索してください")).toBeInTheDocument();
  });

  it("該当する飼主が0件のとき空状態メッセージを表示する", async () => {
    server.use(http.get("*/v1/owners", () => HttpResponse.json({ data: [] })));
    renderModal();

    await search("存在しない飼主ZZZ");

    expect(await screen.findByText("該当する飼主が見つかりません")).toBeInTheDocument();
  });

  it("検索結果がある場合は一覧テーブルを表示する", async () => {
    server.use(
      http.get("*/v1/owners", () =>
        HttpResponse.json({
          data: [
            {
              id: 42,
              owner_name: "山田 太郎",
              phone: "090-1111-2222",
              address1: "東京都渋谷区",
              address2: "",
              discount_rate: 0,
              membership_type: "",
            },
          ],
        }),
      ),
    );
    renderModal();

    await search("山田");

    expect(await screen.findByText("山田 太郎")).toBeInTheDocument();
    expect(screen.getByText("090-1111-2222")).toBeInTheDocument();
    expect(screen.queryByText("該当する飼主が見つかりません")).not.toBeInTheDocument();
  });

  it("検索結果に飼主名が表示される", async () => {
    server.use(
      http.get("*/v1/owners", () =>
        HttpResponse.json({
          data: [
            {
              id: 7,
              owner_name: "鈴木 花子",
              phone: "080-3333-4444",
              address1: "大阪府大阪市",
              address2: "",
              discount_rate: 0,
              membership_type: "",
            },
          ],
        }),
      ),
    );
    renderModal();

    await search("鈴木");

    expect(await screen.findByText("鈴木 花子")).toBeInTheDocument();
  });

  it("APIエラー時は検索結果0件として扱われ空状態を表示する", async () => {
    server.use(http.get("*/v1/owners", () => new HttpResponse(null, { status: 500 })));
    renderModal();

    await search("エラー再現");

    expect(await screen.findByText("該当する飼主が見つかりません")).toBeInTheDocument();
  });
});
