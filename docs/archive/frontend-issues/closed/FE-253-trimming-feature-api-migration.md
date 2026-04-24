# FE-253: カルテ側トリミング機能 — appointments ベース API 対応

**Status**: Open
**Priority**: High
**Affects**: frontend/src/features/trimming/, frontend/src/types/trimming.ts
**Date Created**: 2026-04-16
**Related**: TASK-002, BE-119（前提）, FE-254

## Summary

BE-119 の変更により、カルテ側トリミング管理 API のレスポンス構造が変わる。
`TrimmingRecord` 型（`date DATE`, `status: reserved/in_progress/completed`）から
`Appointment` ベースの flat DTO（`start_time TIMESTAMPTZ`, `status: reservation_status`）に移行する。
`src/types/trimming.ts` / `transforms.ts` / フォームフック / ルートコンポーネントを更新する。

## 現状のコード

```typescript
// src/types/trimming.ts:1-3
import type { TrimmingRecord as BackendTrimmingRecord } from "@/types/generated/models";
export type BackendTrimming = BackendTrimmingRecord;
// ※ BE-118 で TrimmingRecord が models.ts から削除される

// src/features/trimming/api/transforms.ts（抜粋）
export function transformTrimming(data: BackendTrimming): TrimmingUI {
  const statusMap: Record<string, TrimmingUI["status"]> = {
    completed: "完了",
    reserved: "予約",    // ← 廃止
    in_progress: "進行中", // ← 廃止
  };
  return {
    id: String(data.id ?? 0),
    date: data.date && !String(data.date).startsWith("0001")
      ? String(data.date).split("T")[0]   // ← date フィールドは廃止
      : "",
    // ...
    status: statusMap[data.status] ?? "予約",
  };
}

// src/features/trimming/api/create-trimming.ts
type CreateTrimmingRequest = Required<Pick<TrimmingWritable, "pet_id" | "staff_id" | "course_id">> & ...
// ※ "date" フィールドが存在。変更後は "start_time" / "end_time" / "reservation_type_id" に
```

## 必要な変更

### 1. `src/types/trimming.ts` — BackendTrimming 型を新 API DTO に合わせる

BE-118 で `TrimmingRecord` が `models.ts` から削除されるため、
BE-119 の API レスポンス（flat DTO）に合わせた型定義に変更する。

```typescript
// src/types/trimming.ts

/**
 * /v1/trimmings の API レスポンス型（BE-119 の flat DTO に対応）
 * TrimmingRecord が廃止されたため手書き定義
 * NOTE: field names mirror BE-119 trimmingResponse struct
 */
export interface BackendTrimming {
  id: number;
  clinic_id: number;
  reservation_type_id: number;
  start_time: string;  // ISO 8601 ("2026-05-01T09:00:00Z")
  end_time: string;    // ISO 8601
  pet_id?: number;
  staff_id?: number;   // doctor_id をマップ
  status: string;      // reservation_status 値（"pending" | "confirmed" | "in_consultation" | "completed" | "cancelled"）
  source: string;      // "manual" | "line"
  // trimming detail フィールド（appointment_trimming_details から flat化）
  course_id?: number;
  style_request: string;
  bw?: number;
  bw_unit: string;
  bt?: number;
  used_shampoo: string;
  used_ribbon: string;
  remarks: string;
  style_image: string;
  completed_image: string;
  created_at: string;
  updated_at: string;
  // relations
  pet?: {
    id: number;
    name: string;
    pet_number?: string;
    weight?: number;
    owner?: { id: number; name: string };
    animal_species?: { name: string };
  };
  staff?: { id: number; name: string };
  course?: { id: number; name: string; price: number };
  options: Array<{ id: number; name: string }>;
}

/**
 * トリミング作成リクエスト（POST /v1/trimmings）
 */
export interface CreateTrimmingRequest {
  reservation_type_id: number; // category='trimming' の予約区分 ID（必須）
  start_time: string;          // ISO 8601（必須）
  end_time: string;            // ISO 8601（必須）
  pet_id?: number;
  staff_id?: number;
  status?: string;             // デフォルト: "pending"
  course_id?: number;
  style_request?: string;
  bw?: number;
  bw_unit?: string;            // "Kg" | "g"
  bt?: number;
  used_shampoo?: string;
  used_ribbon?: string;
  remarks?: string;
  style_image?: string;
  completed_image?: string;
  option_ids?: number[];
}

/**
 * トリミング更新リクエスト（PATCH /v1/trimmings/:id）
 * 全フィールド optional
 */
export interface UpdateTrimmingRequest {
  start_time?: string;
  end_time?: string;
  pet_id?: number;
  staff_id?: number;
  status?: string;
  course_id?: number;
  style_request?: string;
  bw?: number | null;
  bw_unit?: string;
  bt?: number | null;
  used_shampoo?: string;
  used_ribbon?: string;
  remarks?: string;
  style_image?: string;
  completed_image?: string;
  option_ids?: number[];  // null = 変更なし、空配列 = 全削除
}

export interface TrimmingListResponse {
  data: BackendTrimming[];
  total: number;
  page: number;
  limit: number;
}

// TrimmingFormData は UI 専用の型（手書きOK — バックエンドとの対応は transforms で管理）
export interface TrimmingFormData {
  reservationTypeId: string;  // 追加（必須）
  startTime: string;          // 変更前: 存在しなかった（date のみ）
  endTime: string;            // 追加
  styleRequest: string;
  staffId: string;
  courseId: string;
  optionIds: string[];
  bw: string;
  bwUnit: "Kg" | "g";
  bt: string;
  usedShampoo: string;
  usedRibbon: string;
  remarks: string;
  styleImage: File | null;
  completedImage: File | null;
  status: string;
}
```

### 2. `src/features/trimming/api/transforms.ts` — transformTrimming 更新

```typescript
import type { TrimmingUI } from "@/types";
import type { BackendTrimming } from "@/types/trimming";

// reservation_status → 日本語表示マッピング
const TRIMMING_STATUS_LABEL: Record<string, string> = {
  pending:          "予約待ち",
  confirmed:        "予約確定",
  in_consultation:  "施術中",    // トリミング文脈で "in_consultation" = 施術中
  completed:        "完了",
  cancelled:        "キャンセル",
};

export function transformTrimming(data: BackendTrimming): TrimmingUI {
  return {
    id: String(data.id ?? 0),
    // start_time の日付部分を取得（"YYYY-MM-DDTHH:MM:SSZ" → "YYYY-MM-DD"）
    date: data.start_time
      ? data.start_time.split("T")[0]
      : "",
    startTime: data.start_time ?? "",
    endTime: data.end_time ?? "",
    petId:     data.pet?.id != null ? String(data.pet.id) : undefined,
    ownerId:   data.pet?.owner?.id != null ? String(data.pet.owner.id) : undefined,
    petNumber: data.pet?.pet_number ?? "",
    petName:   data.pet?.name ?? "",
    ownerName: data.pet?.owner?.name ?? "",
    species:   data.pet?.animal_species?.name ?? "",
    weight:    data.pet?.weight != null ? String(data.pet.weight) : "",
    styleRequest: data.style_request ?? "",
    staff:     data.staff?.name ?? "",
    status:    TRIMMING_STATUS_LABEL[data.status] ?? "予約待ち",
    source:    data.source ?? "manual",
    // form fields
    staffId:   data.staff?.id != null ? String(data.staff.id) : "",
    courseId:  data.course?.id != null ? String(data.course.id) : "",
    optionIds: data.options?.map((o) => String(o.id)) ?? [],
    bw:        data.bw != null ? String(data.bw) : "",
    bwUnit:    (data.bw_unit as "Kg" | "g") || "Kg",
    bt:        data.bt != null ? String(data.bt) : "",
    usedShampoo:    data.used_shampoo ?? "",
    usedRibbon:     data.used_ribbon ?? "",
    remarks:        data.remarks ?? "",
    styleImage:     data.style_image || undefined,
    completedImage: data.completed_image || undefined,
  };
}
```

**注意**: `TrimmingUI` 型（`@/types` の `index.ts` または `trimming.ts`）に
`startTime`, `endTime`, `source` フィールドを追加すること。

### 3. `src/features/trimming/hooks/use-trimming-form.ts` — date → start_time/end_time 対応

変更箇所:
- `date: new Date().toISOString()` → `start_time: ""`, `end_time: ""` に置き換え
- `CreateTrimmingRequest` の組み立てで `date` を削除し `start_time` / `end_time` / `reservation_type_id` を追加
- フォームの `TrimmingFormData` 型を新型に合わせる

```typescript
// フォーム送信時の CreateTrimmingRequest 組み立て
const req: CreateTrimmingRequest = {
  reservation_type_id: Number(formData.reservationTypeId), // 追加
  start_time: formData.startTime,  // 変更前: date
  end_time: formData.endTime,      // 変更前: 存在なし
  pet_id: petId,
  staff_id: formData.staffId ? Number(formData.staffId) : undefined,
  course_id: formData.courseId ? Number(formData.courseId) : undefined,
  option_ids: formData.optionIds.map(Number),
  bw: formData.bw ? Number(formData.bw) : undefined,
  bw_unit: formData.bwUnit,
  bt: formData.bt ? Number(formData.bt) : undefined,
  used_shampoo: formData.usedShampoo,
  used_ribbon: formData.usedRibbon,
  remarks: formData.remarks,
  style_image: formData.styleImage ? URL.createObjectURL(formData.styleImage) : undefined,
  completed_image: formData.completedImage ? URL.createObjectURL(formData.completedImage) : undefined,
};
```

### 4. `src/features/trimming/routes/TrimmingForm.tsx` — フィールド変更対応

- `date` フィールド → 日時選択 UI（`start_time` + `end_time` の時刻ピッカーに変更）
- `status` の選択肢を `pending/confirmed/in_consultation/completed/cancelled` に変更し、日本語ラベルを付与
- `reservation_type_id` の選択フィールドを追加（`category='trimming'` の予約区分一覧から選択）
  - 予約区分一覧: 既存の `useGetReservationTypes()` を活用し、`category='trimming'` でフィルタ
  - または専用エンドポイントを使用（現時点では既存の GET /v1/reservation-types で全件取得してフロントでフィルタ）

### 5. `src/features/trimming/routes/TrimmingList.tsx` — 日付表示変更対応

- `date` カラムの表示: `format(new Date(trimming.startTime), "yyyy/MM/dd")` に変更
- ステータスの色分けマップを `reservation_status` 値に対応する色に変更

## UI 操作フロー（変更後）

1. スタッフがトリミングリストを開く
2. 「新規作成」ボタンをクリック
3. フォームで以下を入力:
   - **予約区分** (category='trimming' のものを選択)
   - **施術日時** (start_time + end_time — 日時ピッカー)
   - **ペット** (既存と同様)
   - **担当スタッフ**
   - **コース** (trimming_courses から選択)
   - **オプション** (trimming_options から複数選択)
   - **スタイルリクエスト** / 体重・体温等の臨床情報
4. 保存 → `POST /v1/trimmings` → appointment + trimming_detail が作成される
5. 一覧に戻り、新規トリミング予約が表示される

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useActionState` または `useTransition` でフォーム pending 管理
- [ ] `BackendTrimming` は `TrimmingRecord`（廃止）に依存しない
- [ ] status ラベルはデザイントークン (`C`, `STYLE`) を使用

## 依存関係

- BE-118 完了（`TrimmingRecord` が `models.ts` から削除されている）
- BE-119 完了（API エンドポイントが appointments ベースで動作している）

## 完了条件

- [ ] `src/types/trimming.ts` が `TrimmingRecord` を参照していない（型エラーなし）
- [ ] `transformTrimming` が `status` を `reservation_status` 値にマップしている
- [ ] `transformTrimming` が `start_time` を日付として正しく変換している
- [ ] `CreateTrimmingRequest` に `reservation_type_id`, `start_time`, `end_time` が含まれている
- [ ] フォームで日時選択が可能（`date` フィールドの代替として）
- [ ] `docker compose exec frontend pnpm lint` が通る
- [ ] `docker compose exec frontend pnpm build` が通る（型エラー 0件）
- [ ] トリミング一覧・作成・更新・削除が正常に動作する
