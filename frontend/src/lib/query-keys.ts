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
 *   all()       → 当該エンティティの全キャッシュを無効化したい場合
 *   list(x)     → フィルタ/パラメータ付き一覧
 *   detail(id)  → 詳細キャッシュ (id 指定)
 *
 * 設計方針 (FE-refactor.md 残件1 / FE-R1):
 *   - clinicId はキーに含めない（clinic 切替は queryClient.clear() + full reload
 *     で担保されるため。SPA 切替をやる場合は別途設計する）。
 *   - このファイルは feature 層より下位（lib/）に位置するため、feature 固有の
 *     filters/params 型はここに import しない。呼び出し側の型で渡し、
 *     ジェネリクスでそのままタプルに埋め込む。
 *   - 既存の inline queryKey タプルをバイト同一で温存することを優先する。
 *     命名の不整合（例: "unbilledItems" の camelCase、"accountings" と
 *     "accounting" が別プレフィックスであること）はキャッシュキーの変更を
 *     避けるためにそのまま保持し、コメントで注記する。
 */

export const queryKeys = {
  // ── accounting (会計) ────────────────────────────────────────────
  accountings: {
    /** リスト全体の無効化に使用 */
    all: () => ["accountings"] as const,
    /** 詳細キャッシュ。旧 "accounting-detail" に相当 */
    detail: (id: string) => ["accountings", id] as const,
    /** フィルタ付き一覧フェッチ用。invalidate は all() で足りる（prefix match） */
    list: <F>(filters: F) => ["accountings", filters] as const,
  },
  /** 単数形 "accounting" を第一要素に持つ別ネームスペース（accountings とは別キー） */
  accounting: {
    unpaidAll: () => ["accounting", "unpaid"] as const,
    unpaidBillings: <P>(groupBy: "owner" | "billing" | "monthly", params: P) =>
      ["accounting", "unpaid", groupBy, params] as const,
    ungroupedItems: (petId: string, date: string) => ["accounting-ungrouped", petId, date] as const,
    merchandiseItems: () => ["accounting", "merchandise-items"] as const,
    dailySummary: (date: string, clinicIds: readonly string[]) =>
      ["accounting", "daily-summary", date, clinicIds] as const,
  },
  accountingRefunds: (billingId: string) => ["accounting-refunds", billingId] as const,
  billingItemDiscountSuggestions: (itemId: string) => ["billing-item-discount-suggestions", itemId] as const,
  // P2-11: clinicId をキーに含める（拠点横断で開いた記録のクリニックごとに残高が異なるため、
  // ownerId のみでは他クリニックのキャッシュ値を誤って再利用してしまう）
  ownerUnpaidBalance: (clinicId: string, ownerId: string | undefined) =>
    ["owner-unpaid-balance", clinicId, ownerId] as const,
  /** 既存 ad-hoc キーの camelCase 表記をそのまま温存（キャッシュキー変更回避） */
  unbilledItems: (petId: string) => ["unbilledItems", petId] as const,

  // ── masters (マスタ設定) ──────────────────────────────────────────
  masters: {
    /** 汎用マスタカテゴリキー。"masterItems" の代わりにこれを使う */
    category: (name: string) => ["masters", name] as const,
    /**
     * 担当者セレクト用の薄い staff 一覧。
     * masters.category("staffs") のフル master Staff shape とは別キー。
     * 異なる transform 結果を同一 key に載せない（cache poison 防止）。
     * invalidate は ["masters","staffs"] prefix で両方を無効化できる。
     */
    staffSelectorList: () => ["masters", "staffs", "selector-list"] as const,
    /**
     * トリミングコース CRUD のフル shape（isActive 等）。
     * useGetMasterItems("trimmingCourse") の MasterItem[] とは別キー（cache poison 防止）。
     * invalidate は ["masters","trimmingCourse"] prefix で両方を無効化できる。
     */
    trimmingCoursesFull: () => ["masters", "trimmingCourse", "full"] as const,
    animalSpecies: {
      /** features/master/api/animal-species.ts の CRUD 用ベースキー */
      all: () => ["masters", "animal-species"] as const,
      /** hooks/use-animal-species.ts の一覧フェッチ用（active/all フィルタ付き） */
      filtered: (includeInactive: boolean) =>
        ["masters", "animal-species", includeInactive ? "all" : "active"] as const,
    },
    /** features/master/api/medicine-dose-params.ts の種別（dog/cat）パラメータ CRUD 用 (#201) */
    medicineDoseParams: (medicineId: string) => ["masters", "medicines", medicineId, "dose-params"] as const,
    /**
     * get-diagnosis-options.ts の診断名一覧。masters.category とは別シェイプ（第3要素あり）。
     * type_id は診断名 API 上は数値だが、呼び出し元の型 (number | string | null) をそのまま許容する。
     */
    diagnosisNames: (typeId?: number | string | null) => ["masters", "diagnosis-names", typeId ?? null] as const,
    /**
     * 既知の未解決バグ: reservation-types CRUD (SERVICE_TYPES_QUERY_KEY =
     * masters.category("reservation-types")) を invalidate しても、このキーは
     * prefix が異なる ("reservationType" camelCase 単数)ため無効化されない。
     * 本移行はこの既存挙動を変更せず温存する（意図的な仕様変更は別途判断）。
     */
    reservationTypesGrouped: () => ["masters", "reservationType", "grouped"] as const,
    /** clinicId は ReservationFormModal 経由の呼び出しで未確定時 null になり得る */
    reservationTypeSubResource: (
      clinicId: string | null,
      reservationTypeId: string,
      resource: "available-slots" | "occupations" | "unavailable-times",
    ) => ["masters", "clinics", clinicId, "reservation-types", reservationTypeId, resource] as const,
  },

  staffs: {
    subResource: (
      staffId: string,
      resource: "capable-reservation-types" | "clinics" | "permission-groups",
    ) => ["masters", "staffs", staffId, resource] as const,
    allPermissionGroupMap: (staffIds: readonly string[]) =>
      ["masters", "staffs", "all-permission-group-map", ...staffIds] as const,
  },

  clinics: {
    /** features/clinic-settings/api/clinics.ts の CRUD 一覧キー */
    all: () => ["clinics"] as const,
    /** hooks 側の "clinics-list" は別プレフィックス（clinics.all() とは無関係） */
    list: (scope?: string) => ["clinics-list", scope ?? "assigned"] as const,
    reservationStaffs: (clinicId: string) => ["clinics", clinicId, "reservation-staffs"] as const,
  },

  // ── shifts (シフト) ───────────────────────────────────────────────
  shifts: {
    all: () => ["shifts"] as const,
    /** date は GetShiftsParams.date (string | undefined) をそのまま許容する */
    list: (date: string | undefined, staffId?: string) => ["shifts", date, staffId] as const,
    onDutyStaffs: (date: string) => ["shifts", "on-duty-staffs", date] as const,
  },
  shiftTemplates: {
    all: () => ["shift-templates"] as const,
  },
  staffsForShift: () => ["staffs-for-shift"] as const,

  clinicHolidays: {
    all: () => ["clinic-holidays"] as const,
    byMonth: (yearMonth: string) => ["clinic-holidays", yearMonth] as const,
  },

  // ── reservations / reception ─────────────────────────────────────
  reservations: {
    all: () => ["reservations"] as const,
    list: <F>(filters: F) => ["reservations", filters] as const,
    detail: (id: string) => ["reservation", id] as const,
    availableTimes: (reservationTypeId: string, date: string, staffId?: string) =>
      ["reservations", "available-times", reservationTypeId, date, staffId ?? ""] as const,
  },
  reception: {
    all: () => ["reception"] as const,
    byDate: (date: string) => ["reception", date] as const,
  },

  // ── examinations / vaccinations ───────────────────────────────────
  examinations: {
    all: () => ["examinations"] as const,
    list: <F>(filters: F) => ["examinations", filters] as const,
    detail: (id: string) => ["examination", id] as const,
    items: (id: string) => ["examination-items", id] as const,
    // Nested under examinations.all() so update/unconfirm/items invalidations refresh print.
    printSnapshot: (id: string, version?: number) =>
      ["examinations", "print-snapshot", id, version ?? "current"] as const,
    typeFields: (examTypeId: string) => ["exam-type-fields", examTypeId] as const,
    /** features/medical-records 側からの参照。["pet",...]ではなく["examinations","pet",petId] */
    byPet: (petId: string) => ["examinations", "pet", petId] as const,
    byRecord: (medicalRecordId: string) => ["examinations", "record", medicalRecordId] as const,
  },
  labDevice: {
    board: () => ["lab-device", "board"] as const,
    unlinked: () => ["lab-device", "unlinked"] as const,
    station: () => ["lab-device", "station"] as const,
  },
  vaccinations: {
    all: () => ["vaccinations"] as const,
    list: <F>(filters: F) => ["vaccinations", filters] as const,
    detail: (id: string) => ["vaccination", id] as const,
    byPet: (petId: string) => ["vaccinations", "pet", petId] as const,
  },

  // ── hospitalization ───────────────────────────────────────────────
  hospitalizations: {
    all: () => ["hospitalizations"] as const,
    list: <F>(filters: F) => ["hospitalizations", filters] as const,
    detail: (id: string) => ["hospitalization", id] as const,
    detailRaw: (id: string) => ["hospitalization", "raw", id] as const,
    carePlanItems: (hospitalizationId: string) =>
      ["hospitalizations", hospitalizationId, "care-plan-items"] as const,
    treatmentPlans: (hospitalizationId: string) =>
      ["hospitalizations", hospitalizationId, "treatment-plans"] as const,
    dailyRecords: {
      all: (hospitalizationId: string) => ["hospitalizations", hospitalizationId, "daily-records"] as const,
      byDate: (hospitalizationId: string, date: string) =>
        ["hospitalizations", hospitalizationId, "daily-records", date] as const,
    },
  },

  // ── trimming ──────────────────────────────────────────────────────
  trimmings: {
    all: () => ["trimmings"] as const,
    list: <F>(filters: F) => ["trimmings", filters] as const,
    detail: (id: string) => ["trimming", id] as const,
    byPet: (petId: string) => ["trimmings", "pet", petId] as const,
  },

  // ── estimates (見積) ──────────────────────────────────────────────
  estimates: {
    all: () => ["estimates"] as const,
    list: <P>(params: P) => ["estimates", params] as const,
    detail: (id: string) => ["estimates", id] as const,
  },

  // ── checkups (健診) ───────────────────────────────────────────────
  /**
   * 既知の第一要素衝突（移行前から存在・本移行が新規に持ち込んだものではない）:
   * checkups.list(filters) は ["checkups", filtersObj] だが、medicalRecords.checkups(id)
   * は ["checkups", medicalRecordIdString] であり、どちらも "checkups" を第一要素に持つ
   * 別エンティティ。invalidateQueries({queryKey:["checkups"]}) の prefix match は両方を
   * 巻き込む点に注意（現状どちらの mutation もそこまで広く invalidate していないため実害なし）。
   */
  checkups: {
    list: <F>(filters: F) => ["checkups", filters] as const,
    typeFields: (checkupTypeId: string) => ["checkup-type-fields", checkupTypeId] as const,
  },

  // ── cash register (レジ締め) ──────────────────────────────────────
  cashRegister: {
    preview: {
      all: () => ["cash-register-preview"] as const,
      byDatePeriod: (date: string, period: string) => ["cash-register-preview", date, period] as const,
      byDatePeriodRefresh: (date: string, period: string, refreshNonce: number) =>
        ["cash-register-preview", date, period, refreshNonce] as const,
    },
    closes: {
      all: () => ["cash-register-closes"] as const,
      list: <P>(params: P) => ["cash-register-closes", params] as const,
    },
  },

  // ── aggregation (集計) ────────────────────────────────────────────
  ownerAggregations: {
    list: <P>(params: P) => ["owner-aggregations", params] as const,
    cpmStageCounts: <P>(stage: string, baseParams: P) => ["owner-aggregations-cpm-count", stage, baseParams] as const,
  },

  // ── inventory (在庫) ──────────────────────────────────────────────
  /** 既存の camelCase 文字列リテラルをそのまま温存（kebab-case への統一はキー変更を伴うため見送り） */
  inventoryItems: {
    all: () => ["inventoryItems"] as const,
    list: <P>(params: P) => ["inventoryItems", params] as const,
    /** 既知の未解決ギャップ: 詳細プレフィックスが list と異なり、update/create は list しか invalidate していない */
    detail: (id: string) => ["inventoryItem", id] as const,
  },

  // ── lstep ─────────────────────────────────────────────────────────
  /** clinicId は getClinicId() (string | null) をそのまま許容する（未選択時は enabled で無効化） */
  lstepSettings: (clinicId: string | null) => ["lstep-settings", clinicId] as const,
  lstepTagConfig: {
    autoManagedPrefixes: () => ["lstep-tag-config", "auto-managed-prefixes"] as const,
    conditionTagMappings: () => ["lstep-tag-config", "condition-tag-mappings"] as const,
    sendPurposeTagPrefixes: () => ["lstep-tag-config", "send-purpose-tag-prefixes"] as const,
  },
  triggerPriorities: (clinicId: string | null) => ["trigger-priorities", clinicId] as const,
  lstepTagCodeMappings: (clinicId: string | null) => ["lstep-tag-code-mappings", clinicId] as const,
  lstepCsvImports: {
    all: () => ["lstep-csv-imports"] as const,
    list: (limit: number) => ["lstep-csv-imports", limit] as const,
  },
  lstepDeliveryTriggerSummary: {
    all: () => ["lstep-delivery-trigger-summary"] as const,
    summary: (from: string, to: string, triggerType?: string | null) =>
      ["lstep-delivery-trigger-summary", from, to, triggerType ?? ""] as const,
  },
  lstepDeliveryTriggerLogs: {
    all: () => ["lstep-delivery-trigger-logs"] as const,
    logs: (
      from: string,
      to: string,
      triggerType: string | null | undefined,
      status: string | null | undefined,
      page: number,
    ) => ["lstep-delivery-trigger-logs", from, to, triggerType ?? "", status ?? "", page] as const,
  },
  lstepTagSummary: () => ["lstep-tag-summary"] as const,
  lstepTagOwners: {
    list: <P>(params: P) => ["lstep-tag-owners", params] as const,
  },
  checkupSyncPreview: <P>(clinicId: string | null, params: P) => ["checkup-sync-preview", clinicId, params] as const,
  lstepVisitConversion: (yearMonth: string, days: number) => ["lstep-visit-conversion", yearMonth, days] as const,
  lstepDeliveryStats: (yearMonth: string) => ["lstep-delivery-stats", yearMonth] as const,

  // ── owners ────────────────────────────────────────────────────────
  owners: {
    all: () => ["owners"] as const,
    detail: (id: string) => ["owners", id] as const,
    subOwnerOptions: (search: string) =>
      ["owners", { scope: "sub-owner-options", search }] as const,
  },
  ownerSharedPets: {
    detail: (ownerId: string) => ["owner-shared-pets", ownerId] as const,
  },
  ownerLineTags: (ownerId: string) => ["owner-line-tags", ownerId] as const,
  lineSendHistory: (ownerId: string) => ["line-send-history", ownerId] as const,

  // ── line-reservation ──────────────────────────────────────────────
  lineCustomers: (clinicId: string) => ["line-customers", clinicId] as const,
  lineReservationSettings: (clinicId: string) => ["reservation-settings", clinicId] as const,

  // ── pets ──────────────────────────────────────────────────────────
  pets: {
    /** ownerId 指定時は {ownerId} オブジェクトを第2要素に持つ（既存シェイプを温存） */
    list: (ownerId?: string, options?: { includeDeceased?: boolean }) => {
      if (options?.includeDeceased) {
        return ["pets", { ...(ownerId ? { ownerId } : {}), includeDeceased: true }] as const;
      }
      return ownerId ? (["pets", { ownerId }] as const) : (["pets"] as const);
    },
    detail: (id: string) => ["pet", id] as const,
  },
  petSubOwners: {
    detail: (petId: string) => ["pet-sub-owners", petId] as const,
    metadata: (petId: string) =>
      ["pet", petId, "sub-owner-metadata"] as const,
  },

  // ── owner-report ──────────────────────────────────────────────────
  ownerReportPets: (ownerId: string) =>
    ["owner-report-pets", ownerId] as const,
  petTrimmingHistory: (petId: string) => ["pet-trimmings", "report", petId] as const,
  petTreatmentHistory: <F, O>(petId: string, filter: F, options: O) =>
    ["pet-treatment-history", petId, filter, options] as const,
  /** owner-report 側の集計。features/medical-records の examinations.byPet とは別シェイプ */
  petExaminationsReport: (petId: string) => ["pet-examinations", "report", petId] as const,
  petFirstVisit: (petId: string) => ["pet-first-visit", petId] as const,
  petCheckupResultsReport: (petId: string) => ["pet-checkup-results", "report", petId] as const,

  // ── medical-records ───────────────────────────────────────────────
  medicalRecords: {
    all: () => ["medical-records"] as const,
    list: <F>(filters: F) => ["medical-records", filters] as const,
    history: (petId: string) => ["medical-records", "history", petId] as const,
    detail: (id: string) => ["medical-record", id] as const,
    /**
     * P2-15: clinicId を末尾に含めるのは任意（拠点横断で開いたレコードの子リソースが
     * X-Clinic-ID 解決前後で別クエリとして自然に再フェッチされるようにするため）。
     * clinicId 省略時は従来通り2要素キー — invalidateQueries の prefix match は
     * clinicId 付きキーも包含するため mutation 側の invalidate は変更不要。
     */
    vitals: (medicalRecordId: string, clinicId?: string) =>
      clinicId ? (["vitals", medicalRecordId, clinicId] as const) : (["vitals", medicalRecordId] as const),
    treatments: (medicalRecordId: string, clinicId?: string) =>
      clinicId ? (["treatments", medicalRecordId, clinicId] as const) : (["treatments", medicalRecordId] as const),
    checkups: (medicalRecordId: string) => ["checkups", medicalRecordId] as const,
    addenda: (medicalRecordId: string) => ["record-addenda", medicalRecordId] as const,
    images: (medicalRecordId: string, clinicId?: string) =>
      clinicId ? (["record-images", medicalRecordId, clinicId] as const) : (["record-images", medicalRecordId] as const),
    clinicalPlan: (medicalRecordId: string, clinicId?: string) =>
      clinicId ? (["clinical-plan", medicalRecordId, clinicId] as const) : (["clinical-plan", medicalRecordId] as const),
    estimate: (medicalRecordId: string) => ["estimate", "record", medicalRecordId] as const,
    /** medicalRecords.detail(id) と "medical-record" プレフィックスを共有（意図的） */
    billingConfirmation: (medicalRecordId: string) =>
      ["medical-record", medicalRecordId, "billing-confirmation"] as const,
  },

  // ── identity-links (同一飼主・ペット連携) ──────────────────────────
  identityLinks: {
    ownerSearch: (q: string) => ["identity-links", "owner-search", q] as const,
    petSearch: (q: string) => ["identity-links", "pet-search", q] as const,
    ownerGroup: (clinicId: number, ownerId: number) =>
      ["identity-links", "owner-group", clinicId, ownerId] as const,
    petGroup: (clinicId: number, petId: number) =>
      ["identity-links", "pet-group", clinicId, petId] as const,
  },

  // ── manual / closing-settings ─────────────────────────────────────
  manualArticles: {
    all: () => ["manual-articles"] as const,
  },
  closingSettings: {
    all: () => ["closing-settings"] as const,
    holidays: () => ["closing-settings", "holidays"] as const,
  },

  // ── test-only ─────────────────────────────────────────────────────
  /** react-query.test.ts の QueryCache onError 回帰テスト専用キー。本番コードから参照しないこと */
  _test: {
    fe516Fail: () => ["fe5-16-fail"] as const,
    fe516Silent: () => ["fe5-16-silent"] as const,
  },
} as const;

/**
 * 認証ユーザー情報 (/me) の query key。
 * auth feature の useGetMe と、権限変更後の invalidateQueries（master 等）で共有する。
 */
export const ME_QUERY_KEY = ["me"] as const;
