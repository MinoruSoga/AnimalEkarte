import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { PatientSelectionTable } from "./PatientSelectionTable";
import type { Pet } from "@/types";

// BUG-452: 検索は backend の search / species / owner_id 述語とページネーションへ
// 委譲する。ここでは実 HTTP 境界（MSW）でクエリパラメータと描画行を突き合わせる。
// フックを差し替えると「パラメータは正しいが応答をクライアント側で削っている」
// 退行を素通りさせるため、応答行が1行も減らずに描画されることまで検証する。

interface BackendPetRow {
  id: number;
  owner_id: number;
  owner: {
    id: number;
    name: string;
    name_kana: string;
    phone: string;
    clinic_id: number;
  };
  name: string;
  pet_name_kana: string;
  animal_species: { name: string };
  animal_species_id: number;
  gender: string;
  status: string;
}

function backendPet(id: number, overrides: Partial<BackendPetRow> = {}): BackendPetRow {
  return {
    id,
    owner_id: 100,
    owner: {
      id: 100,
      name: `タナカ${id}`,
      name_kana: "タナカ",
      phone: "090-0000-0000",
      clinic_id: 1,
    },
    name: `ミケ${id}`,
    pet_name_kana: "ミケ",
    animal_species: { name: "犬" },
    animal_species_id: 1,
    gender: "male",
    status: "alive",
    ...overrides,
  };
}

const SPECIES_MASTER = [
  { id: 1, name: "犬", is_active: true },
  { id: 2, name: "猫", is_active: true },
];

let petRequests: URL[] = [];

interface PetListBody {
  data: BackendPetRow[];
  total?: number;
  page?: number;
  limit?: number;
}

function mockPetList(respond: (url: URL) => PetListBody) {
  server.use(
    http.get("/api/v1/pets", ({ request }) => {
      const url = new URL(request.url);
      petRequests.push(url);
      return HttpResponse.json(respond(url));
    }),
  );
}

function mockSpecies(body: unknown = SPECIES_MASTER, status = 200) {
  server.use(
    http.get("/api/v1/masters/animal-species", () =>
      status === 200 ? HttpResponse.json(body) : new HttpResponse(null, { status }),
    ),
  );
}

const SEARCH_LABEL = "検索（ペット名・飼主名・よみ・電話）";

function renderTable(props: Partial<React.ComponentProps<typeof PatientSelectionTable>> = {}) {
  return render(<PatientSelectionTable onSelect={vi.fn()} selectedPets={[]} {...props} />, {
    wrapper: createTestWrapper(),
  });
}

beforeEach(() => {
  petRequests = [];
  mockSpecies();
});

afterEach(() => {
  server.resetHandlers();
  vi.useRealTimers();
});

// ─────────────────────────────────────────────────────────────
// BUG-452 中核: 「取得済み先頭20件に対するクライアント側絞り込み」の消滅
// ─────────────────────────────────────────────────────────────
describe("PatientSelectionTable — サーバサイド検索への委譲 (BUG-452)", () => {
  it("検索語を search 述語として API へ渡し、応答行を1行も削らずに描画する", async () => {
    // 検索語「やまだ」に対し、旧実装の normalizedIncludes なら全滅する行を返す。
    // クライアント側フィルタへ退行するとこのテストは 0 件描画で必ず落ちる。
    const rows = Array.from({ length: 20 }, (_, i) => backendPet(i + 1));
    mockPetList(() => ({ data: rows, total: 1204, page: 1, limit: 20 }));

    const user = userEvent.setup();
    renderTable();

    await user.type(screen.getByLabelText(SEARCH_LABEL), "やまだ");

    await waitFor(() => expect(petRequests.length).toBeGreaterThan(0));
    const url = petRequests[petRequests.length - 1];
    expect(url.searchParams.get("search")).toBe("やまだ");
    expect(url.searchParams.get("page")).toBe("1");
    expect(url.searchParams.get("limit")).toBe("20");

    expect(await screen.findByText("ミケ20")).toBeInTheDocument();
    for (const row of rows) {
      expect(screen.getByText(row.name)).toBeInTheDocument();
    }
  });

  it("飼主No を owner_id 述語として渡す", async () => {
    mockPetList(() => ({
      data: [backendPet(1)],
      total: 1,
      page: 1,
      limit: 20,
    }));

    const user = userEvent.setup();
    renderTable();

    await user.type(screen.getByLabelText("飼主No"), "30042");

    await waitFor(() => expect(petRequests.length).toBeGreaterThan(0));
    expect(petRequests[petRequests.length - 1].searchParams.get("owner_id")).toBe("30042");
  });

  it("種別はフリー入力ではなく animal_species_id の数値としてセレクトから渡る", async () => {
    mockPetList(() => ({
      data: [backendPet(1)],
      total: 1,
      page: 1,
      limit: 20,
    }));

    const user = userEvent.setup();
    renderTable();

    await user.click(await screen.findByLabelText("種別"));
    await waitFor(() => screen.getByRole("option", { name: "猫" }));
    await user.click(screen.getByRole("option", { name: "猫" }));

    await waitFor(() => expect(petRequests.length).toBeGreaterThan(0));
    expect(petRequests[petRequests.length - 1].searchParams.get("species")).toBe("2");
  });

  it("直接記録selectorでは死亡sentinelを取得する", async () => {
    mockPetList(() => ({ data: [], total: 0, page: 1, limit: 20 }));

    const user = userEvent.setup();
    renderTable({ includeDeceased: true });
    await user.type(screen.getByLabelText(SEARCH_LABEL), "テスト");

    await waitFor(() => expect(petRequests.length).toBeGreaterThan(0));
    expect(petRequests[petRequests.length - 1].searchParams.get("include_deceased")).toBe("true");
  });
});

// ─────────────────────────────────────────────────────────────
// 総件数とページャ: 「20件」を全件であるかのように断定表示しない
// ─────────────────────────────────────────────────────────────
describe("PatientSelectionTable — 総件数とページャ", () => {
  it("総件数が描画行数を超えるとき「N件中 1-20件」を提示しページャを出す", async () => {
    const rows = Array.from({ length: 20 }, (_, i) => backendPet(i + 1));
    mockPetList(() => ({ data: rows, total: 1204, page: 1, limit: 20 }));

    const user = userEvent.setup();
    renderTable();
    await user.type(screen.getByLabelText(SEARCH_LABEL), "や");

    expect(await screen.findByText("ミケ1")).toBeInTheDocument();
    expect(screen.getAllByText("1,204件中 1-20件").length).toBeGreaterThan(0);
    expect(screen.queryByText("20件")).not.toBeInTheDocument();
    expect(screen.getByTestId("pagination-next")).toBeInTheDocument();
  });

  it("ページ送りで page 述語が進み、次ページの行が描画される", async () => {
    mockPetList((url) => {
      const page = Number(url.searchParams.get("page") ?? "1");
      return {
        data: [backendPet(page === 1 ? 1 : 21)],
        total: 40,
        page,
        limit: 20,
      };
    });

    const user = userEvent.setup();
    renderTable();
    await user.type(screen.getByLabelText(SEARCH_LABEL), "や");
    expect(await screen.findByText("ミケ1")).toBeInTheDocument();

    await user.click(screen.getByTestId("pagination-next"));

    expect(await screen.findByText("ミケ21")).toBeInTheDocument();
    await waitFor(() =>
      expect(petRequests.some((u) => u.searchParams.get("page") === "2")).toBe(true),
    );
  });

  it("デバウンス待ちの間は行を残しつつ、古い条件の結果から選択させない", async () => {
    mockPetList(() => ({
      data: [backendPet(1)],
      total: 40,
      page: 1,
      limit: 20,
    }));

    const user = userEvent.setup();
    renderTable();
    await user.type(screen.getByLabelText(SEARCH_LABEL), "や");
    expect(await screen.findByRole("button", { name: "選択: ミケ1 (ID 1)" })).toBeEnabled();

    // 追加入力でデバウンス待ちに入る。行は消さない（消すと preservePreviousData が
    // 無意味になり「前回取得: N件中 …」が空の表の横に出る）が、選択は塞ぐ。
    await user.type(screen.getByLabelText(SEARCH_LABEL), "まだ");

    const staleButton = await screen.findByRole("button", {
      name: "読み込み中・選択不可: ミケ1 (ID 1)",
    });
    expect(staleButton).toBeDisabled();
    expect(screen.getByText("ミケ1")).toBeInTheDocument();
  });

  it("応答の page / limit が不正でも破綻した範囲を表示しない", async () => {
    // 外部境界の値をそのまま使うと「1,204件中 -19-0件」や totalPages=Infinity になる。
    mockPetList(() => ({
      data: [backendPet(1)],
      total: 1204,
      page: 0,
      limit: 0,
    }));

    const user = userEvent.setup();
    renderTable();
    await user.type(screen.getByLabelText(SEARCH_LABEL), "や");

    expect(await screen.findByText("ミケ1")).toBeInTheDocument();
    expect(screen.queryByText(/-19/)).not.toBeInTheDocument();
    expect(screen.queryByText(/Infinity/)).not.toBeInTheDocument();
    expect(screen.getAllByText("1,204件中 1-20件").length).toBeGreaterThan(0);
  });

  it("応答の total が欠落しているときは件数を主張しない", async () => {
    mockPetList(() => ({ data: [backendPet(1)] }));

    const user = userEvent.setup();
    renderTable();
    await user.type(screen.getByLabelText(SEARCH_LABEL), "や");

    expect(await screen.findByText("ミケ1")).toBeInTheDocument();
    expect(screen.queryByText("0件")).not.toBeInTheDocument();
    expect(screen.queryByText(/件中/)).not.toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// 複数選択のページ跨ぎ保持
// ─────────────────────────────────────────────────────────────
describe("PatientSelectionTable — ページを跨いだ複数選択", () => {
  const selectedOnPage1 = {
    id: "1",
    ownerId: "100",
    name: "ミケ1",
    ownerName: "タナカ1",
    species: "犬",
    status: "生存",
  } as unknown as Pet;
  const selectedOnPage2 = {
    id: "21",
    ownerId: "100",
    name: "ミケ21",
    ownerName: "タナカ21",
    species: "犬",
    status: "生存",
  } as unknown as Pet;

  it("ページ移動後も選択が保持され、現ページ外の選択件数が提示される", async () => {
    mockPetList((url) => {
      const page = Number(url.searchParams.get("page") ?? "1");
      return {
        data: [backendPet(page === 1 ? 1 : 21)],
        total: 40,
        page,
        limit: 20,
      };
    });

    const user = userEvent.setup();
    renderTable({ selectedPets: [selectedOnPage1, selectedOnPage2] });
    await user.type(screen.getByLabelText(SEARCH_LABEL), "や");

    // ページ1: ミケ1 が選択中、ミケ21 は別ページの選択として提示される
    expect(await screen.findByRole("button", { name: "選択中: ミケ1 (ID 1)" })).toBeInTheDocument();
    expect(screen.getByText("選択中: 2件（別ページの1件を含む）")).toBeInTheDocument();

    await user.click(screen.getByTestId("pagination-next"));

    // ページ2: 選択は失われず、ミケ21 が選択中として描画される
    expect(
      await screen.findByRole("button", { name: "選択中: ミケ21 (ID 21)" }),
    ).toBeInTheDocument();
    expect(screen.getByText("選択中: 2件（別ページの1件を含む）")).toBeInTheDocument();
  });

  it("未検索で一覧が空のうちは選択を「別ページ」と断定しない", async () => {
    mockPetList(() => ({ data: [], total: 0, page: 1, limit: 20 }));

    renderTable({ selectedPets: [selectedOnPage1, selectedOnPage2] });

    // 一覧が空なのに「別ページの2件」と言うと、存在しないページを示唆する嘘になる。
    expect(await screen.findByText("選択中: 2件")).toBeInTheDocument();
    expect(screen.queryByText(/別ページ/)).not.toBeInTheDocument();
  });

  it("選択が全て現ページにあるときは別ページの注記を出さない", async () => {
    mockPetList(() => ({
      data: [backendPet(1)],
      total: 1,
      page: 1,
      limit: 20,
    }));

    const user = userEvent.setup();
    renderTable({ selectedPets: [selectedOnPage1] });
    await user.type(screen.getByLabelText(SEARCH_LABEL), "や");

    expect(await screen.findByRole("button", { name: "選択中: ミケ1 (ID 1)" })).toBeInTheDocument();
    expect(screen.getByText("選択中: 1件")).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// 検索欄の簡素化: backend の述語と1対1の3コントロール
// ─────────────────────────────────────────────────────────────
describe("PatientSelectionTable — 検索コントロール", () => {
  it("検索・飼主No・種別の3コントロールのみを持つ", async () => {
    renderTable();

    expect(screen.getByLabelText(SEARCH_LABEL)).toBeInTheDocument();
    expect(screen.getByLabelText("飼主No")).toBeInTheDocument();
    // 種別はフリー入力ではなくセレクト（combobox）であること
    expect(await screen.findByLabelText("種別")).toHaveAttribute("role", "combobox");

    expect(screen.getAllByRole("textbox")).toHaveLength(2);
    for (const removed of ["飼主名", "飼主名よみ", "電話番号", "ペット名", "ペット名よみ"]) {
      expect(screen.queryByLabelText(removed)).not.toBeInTheDocument();
    }
  });

  it("検索ボタンを持たず、開いた直後は一覧が空である", async () => {
    mockPetList(() => ({
      data: [backendPet(1)],
      total: 1,
      page: 1,
      limit: 20,
    }));

    renderTable();

    expect(screen.queryByRole("button", { name: "検索" })).not.toBeInTheDocument();
    // クリアは常設する（押した瞬間に消えるとフォーカスが失われるため）
    expect(screen.getByRole("button", { name: "クリア" })).toBeInTheDocument();
    expect(screen.queryByText("ミケ1")).not.toBeInTheDocument();
    await waitFor(() => expect(screen.getByText("検索条件を入力してください")).toBeInTheDocument());
    expect(petRequests).toHaveLength(0);
  });
});

// ─────────────────────────────────────────────────────────────
// デバウンス自動検索の配線（フック単体の green は配線の証明にならない）
// ─────────────────────────────────────────────────────────────
describe("PatientSelectionTable — デバウンス自動検索", () => {
  it("初期マウントでは問い合わせず、入力停止から300ms後に1回だけ問い合わせる", async () => {
    mockPetList(() => ({
      data: [backendPet(1)],
      total: 1,
      page: 1,
      limit: 20,
    }));

    vi.useFakeTimers({ shouldAdvanceTime: true });
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    renderTable();

    expect(petRequests).toHaveLength(0);

    await user.type(screen.getByLabelText(SEARCH_LABEL), "やまだ");
    expect(petRequests).toHaveLength(0);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(300);
    });
    await waitFor(() => expect(petRequests).toHaveLength(1));
    expect(petRequests[0].searchParams.get("search")).toBe("やまだ");
  });
});

// ─────────────────────────────────────────────────────────────
// BUG-2026-07-27-01 回帰: 取得失敗を「該当0件」と誤表示しない
// ─────────────────────────────────────────────────────────────
describe("PatientSelectionTable — 取得失敗をゼロ件と誤表示しない", () => {
  it("取得失敗時はエラー文言を出し「見つかりませんでした」も件数も表示しない", async () => {
    server.use(http.get("/api/v1/pets", () => new HttpResponse(null, { status: 500 })));

    const user = userEvent.setup();
    renderTable();
    await user.type(screen.getByLabelText(SEARCH_LABEL), "や");

    expect(await screen.findByText("患者一覧の取得に失敗しました")).toBeInTheDocument();
    expect(screen.queryByText("該当する患者が見つかりませんでした")).not.toBeInTheDocument();
    expect(screen.queryByText("0件")).not.toBeInTheDocument();
  });

  it("取得成功かつ本当に0件なら空状態を表示する", async () => {
    mockPetList(() => ({ data: [], total: 0, page: 1, limit: 20 }));

    const user = userEvent.setup();
    renderTable();
    await user.type(screen.getByLabelText(SEARCH_LABEL), "や");

    expect(await screen.findByText("該当する患者が見つかりませんでした")).toBeInTheDocument();
    expect(screen.queryByText("患者一覧の取得に失敗しました")).not.toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// TASK-SPECIES-01: 動物種の取得状態を区別する
// ─────────────────────────────────────────────────────────────
describe("PatientSelectionTable — 動物種の取得状態", () => {
  it("動物種の取得に失敗したら文言で明示し、種別セレクトを操作させない", async () => {
    mockSpecies(null, 500);
    mockPetList(() => ({ data: [], total: 0, page: 1, limit: 20 }));

    renderTable();

    const message = await screen.findByText("動物種の取得に失敗しました。");
    const trigger = screen.getByLabelText("種別");
    expect(trigger).toBeDisabled();
    // なぜ選べないのかがトリガーから辿れること
    expect(trigger).toHaveAttribute("aria-describedby", message.id);
  });

  it("動物種マスタが0件ならその旨を提示する", async () => {
    mockSpecies([]);
    mockPetList(() => ({ data: [], total: 0, page: 1, limit: 20 }));

    renderTable();

    expect(await screen.findByText("動物種マスタが登録されていません。")).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// 既存挙動の維持: 行全体では選択せず、死亡ペットは選択不可
// ─────────────────────────────────────────────────────────────
describe("PatientSelectionTable — 選択操作", () => {
  it("行クリックでは選択せず、44px の選択ボタンだけが患者を選択する", async () => {
    mockPetList(() => ({
      data: [backendPet(1)],
      total: 1,
      page: 1,
      limit: 20,
    }));

    const user = userEvent.setup();
    const onSelect = vi.fn();
    renderTable({ onSelect });
    await user.type(screen.getByLabelText(SEARCH_LABEL), "や");

    expect(await screen.findByText("タナカ1")).toBeInTheDocument();
    await user.click(screen.getByText("タナカ1"));
    expect(onSelect).not.toHaveBeenCalled();

    const selectButton = screen.getByRole("button", {
      name: "選択: ミケ1 (ID 1)",
    });
    expect(selectButton).toHaveClass("min-w-11");
    await user.click(selectButton);
    expect(onSelect).toHaveBeenCalledTimes(1);
    expect(onSelect.mock.calls[0][0]).toMatchObject({ id: "1", name: "ミケ1" });
  });

  it("死亡ペットは選択不可のまま描画する", async () => {
    mockPetList(() => ({
      data: [backendPet(1, { status: "deceased" })],
      total: 1,
      page: 1,
      limit: 20,
    }));

    const user = userEvent.setup();
    renderTable();
    await user.type(screen.getByLabelText(SEARCH_LABEL), "や");

    expect(
      await screen.findByRole("button", {
        name: "死亡・選択不可: ミケ1 (ID 1)",
      }),
    ).toBeDisabled();
  });

  it("未知statusは生存扱いせず状態不明として選択不可にする", async () => {
    mockPetList(() => ({
      data: [backendPet(1, { status: "unexpected" })],
      total: 1,
      page: 1,
      limit: 20,
    }));

    const user = userEvent.setup();
    const onSelect = vi.fn();
    renderTable({ onSelect });
    await user.type(screen.getByLabelText(SEARCH_LABEL), "や");

    const button = await screen.findByRole("button", {
      name: "状態不明・選択不可: ミケ1 (ID 1)",
    });
    expect(button).toBeDisabled();
    await user.click(button);
    expect(onSelect).not.toHaveBeenCalled();
  });
});
