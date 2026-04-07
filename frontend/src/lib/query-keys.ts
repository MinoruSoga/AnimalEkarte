/**
 * React Query キーファクトリー
 *
 * TanStack Query の階層一致によるキャッシュ無効化を確実にするため、
 * queryKey を一元管理する。
 *
 * 使用方法:
 *   queryKey: queryKeys.accountings.detail(id)
 *   invalidateQueries({ queryKey: queryKeys.accountings.all() })
 *
 * 命名規則:
 *   all()      → 当該エンティティの全キャッシュを無効化したい場合
 *   detail(id) → 詳細キャッシュ (id 指定)
 */

export const queryKeys = {
  owners: {
    all: () => ["owners"] as const,
    detail: (id: string) => ["owners", id] as const,
  },
  pets: {
    all: () => ["pets"] as const,
    byOwner: (ownerId: string) => ["pets", { ownerId }] as const,
    detail: (id: string) => ["pet", id] as const,
  },
  reservations: {
    all: () => ["reservations"] as const,
    detail: (id: string) => ["reservation", id] as const,
  },
  accountings: {
    /** リスト全体の無効化に使用 */
    all: () => ["accountings"] as const,
    /** 詳細キャッシュ。旧 "accounting-detail" に相当 */
    detail: (id: string) => ["accountings", id] as const,
  },
  medicalRecords: {
    all: () => ["medical-records"] as const,
    detail: (id: string) => ["medical-record", id] as const,
  },
  vaccinations: {
    all: () => ["vaccinations"] as const,
    detail: (id: string) => ["vaccination", id] as const,
  },
  examinations: {
    all: () => ["examinations"] as const,
    detail: (id: string) => ["examination", id] as const,
  },
  hospitalization: {
    all: () => ["hospitalizations"] as const,
    detail: (id: string) => ["hospitalization", id] as const,
    raw: (id: string) => ["hospitalization", "raw", id] as const,
  },
  inventory: {
    all: () => ["inventoryItems"] as const,
    detail: (id: string) => ["inventoryItem", id] as const,
  },
  trimming: {
    all: () => ["trimmings"] as const,
    detail: (id: string) => ["trimming", id] as const,
  },
  estimates: {
    all: () => ["estimates"] as const,
    detail: (id: string) => ["estimates", id] as const,
  },
  checkups: {
    all: () => ["checkups"] as const,
  },
  reception: {
    all: () => ["reception"] as const,
    byDate: (date: string) => ["reception", date] as const,
  },
  shifts: {
    all: () => ["shifts"] as const,
  },
  masters: {
    /** 汎用マスタカテゴリキー。"masterItems" の代わりにこれを使う */
    category: (name: string) => ["masters", name] as const,
    staffs: () => ["masters", "staffs"] as const,
    animalSpecies: () => ["masters", "animal-species"] as const,
    cages: () => ["masters", "cages"] as const,
    medicines: () => ["masters", "medicines"] as const,
  },
} as const;
