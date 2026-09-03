import { createTestWrapper } from "@/testing/TestUtils";
import type { BackendAccounting } from "../api/types";

/**
 * FE4-18: OwnerAccountingHistory.test.tsx / OwnerAccountingHistory.pagination.test.tsx
 * が共有する fixture / render ヘルパー。800 行超のため describe 境界で 2 ファイルに
 * 分割した際、逐語移動した（値・ロジックは 1 文字も変えていない）。
 */

export const mockOwnerId = "42";

type LoosePet = Omit<Partial<NonNullable<BackendAccounting["pet"]>>, "animal_species"> & {
  animal_species?: Partial<NonNullable<NonNullable<BackendAccounting["pet"]>["animal_species"]>>;
};
type LoosePayment = Partial<NonNullable<BackendAccounting["payments"]>[number]>;

// BackendAccounting の最小フィクスチャを作る。transformToAccounting が
// 期待するキーだけ埋め、その他の任意フィールドは null/undefined で OK。
// pet/payments は生成型の全フィールドではなく transformToAccounting が実際に読む
// フィールドのみ埋める意図的な最小フィクスチャのため、Partial 型で緩めている
// （値・振る舞いは元の .test.tsx から 1 文字も変えていない — FE4-18）。
type LooseBackendAccountingOverrides = Omit<Partial<BackendAccounting>, "pet" | "payments"> & {
  pet?: LoosePet;
  payments?: LoosePayment[];
};

export function makeBackendAccounting(overrides: LooseBackendAccountingOverrides): BackendAccounting {
  return {
    id: 0,
    owner_id: Number(mockOwnerId),
    pet_id: 0,
    status: "waiting",
    scheduled_date: "2026-04-29T00:00:00Z",
    subtotal: 0,
    tax_total: 0,
    total_amount: 0,
    items: [],
    payments: [],
    ...overrides,
  } as BackendAccounting;
}

export const completedFixture: BackendAccounting = makeBackendAccounting({
  id: 101,
  status: "completed",
  scheduled_date: "2026-04-20T00:00:00Z",
  subtotal: 5000,
  tax_total: 500,
  total_amount: 5500,
  pet: { id: 7, name: "ぽち", animal_species: { id: 1, name: "犬" } },
  payments: [
    {
      id: 201,
      billing_id: 101,
      method: "cash",
      total_amount: 5500,
      billing_amount: 5500,
      received_amount: 6000,
      change_amount: 500,
      insurance_amount: 0,
      discount_amount: 0,
    },
  ],
});

export const completedFixture2: BackendAccounting = makeBackendAccounting({
  id: 103,
  status: "completed",
  scheduled_date: "2026-04-22T00:00:00Z",
  subtotal: 3000,
  tax_total: 300,
  total_amount: 3300,
  pet: { id: 9, name: "もも", animal_species: { id: 1, name: "犬" } },
  payments: [
    {
      id: 203,
      billing_id: 103,
      method: "credit_card",
      total_amount: 3300,
      billing_amount: 3300,
      received_amount: 3300,
      change_amount: 0,
      insurance_amount: 0,
      discount_amount: 0,
    },
  ],
});

export const waitingFixture: BackendAccounting = makeBackendAccounting({
  id: 102,
  status: "waiting",
  scheduled_date: "2026-04-29T00:00:00Z",
  pet: { id: 8, name: "たま", animal_species: { id: 2, name: "猫" } },
});

export const pendingFixture: BackendAccounting = makeBackendAccounting({
  id: 104,
  status: "pending",
  scheduled_date: "2026-04-28T00:00:00Z",
  pet: { id: 10, name: "ちゃこ", animal_species: { id: 2, name: "猫" } },
});

export const cancelledFixture: BackendAccounting = makeBackendAccounting({
  id: 105,
  status: "cancelled",
  scheduled_date: "2026-04-15T00:00:00Z",
  pet: { id: 11, name: "ハチ", animal_species: { id: 1, name: "犬" } },
});

export const createWrapper = (initialEntries?: string[]) =>
  createTestWrapper({ initialEntries: initialEntries ?? ["/"] });

/** PAGE_SIZE=10 を超えるフィクスチャを生成する。日付は 2026-04-01 から連番。 */
export const makePaginationFixtures = (n: number): BackendAccounting[] =>
  Array.from({ length: n }, (_, i) =>
    makeBackendAccounting({
      id: 200 + i,
      status: "completed",
      scheduled_date: `2026-04-${String(i + 1).padStart(2, "0")}T00:00:00Z`,
      total_amount: (i + 1) * 1000,
      payments: [
        {
          id: 300 + i,
          billing_id: 200 + i,
          method: "cash",
          total_amount: (i + 1) * 1000,
          billing_amount: (i + 1) * 1000,
          received_amount: (i + 1) * 1000,
          change_amount: 0,
          insurance_amount: 0,
          discount_amount: 0,
        },
      ],
    }),
  );
