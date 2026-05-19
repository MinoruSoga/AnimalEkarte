# FE-254: LIFF トリミング予約UI拡張 — コース・オプション選択ステップ追加

**Status**: Closed（2026-05-19 実装済み確認）
**Priority**: Medium
**Affects**: frontend/line-reserve/src/
**Date Created**: 2026-04-16
**Related**: TASK-002, BE-120（前提）

## Summary

LINE予約（LIFF）フローに、`reservation_types.category = 'trimming'` の区分が選択された場合に
「トリミングコース選択」「トリミングオプション選択」ステップを追加する。
既存の7ステップフローを維持しつつ、トリミング区分選択時のみ 2ステップを挿入する。

## 現状のコード

```typescript
// frontend/line-reserve/src/types/models.ts:40-49
export interface Course {
  id: number;
  name: string;
  short_name: string;
  show_short_name: boolean;
  duration_minutes: number;
  reservation_comment: string;
  reservation_image_url: string;
  sort_order: number;
}
// ※ category フィールドなし → BE-118 後に追加が必要

// frontend/line-reserve/src/types/models.ts:119-144
export interface ReservationFlow {
  customerInfo: CustomerInfo;
  courseId: number | null;
  courseName: string;
  staffId: number;
  staffName: string;
  date: string;
  startTime: string;
  endTime: string;
  requestText: string;
}
// ※ トリミング詳細フィールドなし

export type PageType =
  | 'loading' | 'error' | 'maintenance' | 'top' | 'my-reservations'
  | 'step1' | 'step2' | 'step3' | 'step4' | 'step5' | 'step6' | 'step7' | 'step8';
// ※ 'step2b'（トリミングコース選択）/ 'step2c'（オプション選択）がない

// frontend/line-reserve/src/types/models.ts:102-110
export interface CreateReservationBody {
  course_id: number;
  staff_id: number;
  date: string;
  start_time: string;
  end_time: string;
  customer_fields: CustomerFields;
  request_text: string;
}
// ※ trimming_course_id / trimming_option_ids がない
```

## 必要な変更

### 1. `frontend/line-reserve/src/types/models.ts` — 型定義更新

#### 1-a. `Course` に `category` フィールド追加

```typescript
export interface Course {
  id: number;
  name: string;
  short_name: string;
  show_short_name: boolean;
  duration_minutes: number;
  reservation_comment: string;
  reservation_image_url: string;
  sort_order: number;
  category: 'general' | 'trimming'; // ★ 追加（BE-118 で reservation_types に追加）
}
```

#### 1-b. `TrimmingCourse` と `TrimmingOption` 型を追加

```typescript
// 新規追加
export interface TrimmingCourse {
  id: number;
  clinic_id: number;
  name: string;
  price?: number;
  is_active: boolean;
  description: string;
  target_size?: 'small' | 'medium' | 'large' | 'cat';
  duration?: number;
  sort_order: number;
}

export interface TrimmingOption {
  id: number;
  clinic_id: number;
  name: string;
  price?: number;
  is_active: boolean;
  description: string;
  duration?: number;
  is_combinable: boolean;
  sort_order: number;
}
```

#### 1-c. `ReservationFlow` にトリミングフィールド追加

```typescript
export interface ReservationFlow {
  customerInfo: CustomerInfo;
  courseId: number | null;
  courseName: string;
  courseCategory: 'general' | 'trimming'; // ★ 追加（'general' がデフォルト）
  staffId: number;
  staffName: string;
  date: string;
  startTime: string;
  endTime: string;
  requestText: string;
  // ★ 追加: トリミング詳細（courseCategory='trimming' のときのみ使用）
  trimmingCourseId: number | null;
  trimmingCourseName: string;
  trimmingOptionIds: number[];
  trimmingOptionNames: string[]; // 確認画面表示用
  trimmingStyleRequest: string;
}
```

#### 1-d. `PageType` に新ステップを追加

```typescript
export type PageType =
  | 'loading' | 'error' | 'maintenance' | 'top' | 'my-reservations'
  | 'step1'
  | 'step2'
  | 'step2b'  // ★ 追加: トリミングコース選択（category='trimming' のときのみ）
  | 'step2c'  // ★ 追加: トリミングオプション選択（category='trimming' のときのみ）
  | 'step3' | 'step4' | 'step5' | 'step6' | 'step7' | 'step8';
```

#### 1-e. `CreateReservationBody` にトリミングフィールド追加

```typescript
export interface CreateReservationBody {
  course_id: number;
  staff_id: number;
  date: string;
  start_time: string;
  end_time: string;
  customer_fields: CustomerFields;
  request_text: string;
  // ★ 追加（任意: category='trimming' のときのみ送信）
  trimming_course_id?: number;
  trimming_option_ids?: number[];
  trimming_style_request?: string;
}
```

### 2. `frontend/line-reserve/src/api/liff-api.ts` — 新エンドポイント追加

```typescript
// liffApi オブジェクトに追加
getTrimmingCourses: async (clinicId: string, idToken: string): Promise<TrimmingCourse[]> => {
  const res = await httpClient.get<TrimmingCourse[]>(
    `/api/liff/${clinicId}/trimming-courses`,
    { headers: { Authorization: `Bearer ${idToken}` } }
  );
  return res.data;
},

getTrimmingOptions: async (clinicId: string, idToken: string): Promise<TrimmingOption[]> => {
  const res = await httpClient.get<TrimmingOption[]>(
    `/api/liff/${clinicId}/trimming-options`,
    { headers: { Authorization: `Bearer ${idToken}` } }
  );
  return res.data;
},
```

`createReservation` の呼び出し時に trimming フィールドを含める:

```typescript
createReservation: async (clinicId, idToken, body: CreateReservationBody): Promise<CreateReservationResponse> => {
  // 既存の実装を維持（body 型に trimming フィールドが追加されるだけ）
  const res = await httpClient.post<CreateReservationResponse>(
    `/api/liff/${clinicId}/reservations`,
    body,
    { headers: { Authorization: `Bearer ${idToken}` } }
  );
  return res.data;
},
```

### 3. `frontend/line-reserve/src/hooks/use-reservation-flow.ts` — トリミングフィールド追加

```typescript
const initialFlow: ReservationFlow = {
  customerInfo: { name: '', phone: '', ownerName: '', pets: [] },
  courseId: null,
  courseName: '',
  courseCategory: 'general', // ★ 追加
  staffId: 0,
  staffName: '',
  date: '',
  startTime: '',
  endTime: '',
  requestText: '',
  // ★ 追加
  trimmingCourseId: null,
  trimmingCourseName: '',
  trimmingOptionIds: [],
  trimmingOptionNames: [],
  trimmingStyleRequest: '',
};

// setCourse を拡張（category を受け取る）
const setCourse = useCallback((id: number, name: string, category: 'general' | 'trimming') => {
  setFlow(prev => ({ ...prev, courseId: id, courseName: name, courseCategory: category }));
}, []);

// 新メソッド追加
const setTrimmingCourse = useCallback((id: number, name: string) => {
  setFlow(prev => ({ ...prev, trimmingCourseId: id, trimmingCourseName: name }));
}, []);

const setTrimmingOptions = useCallback((ids: number[], names: string[]) => {
  setFlow(prev => ({ ...prev, trimmingOptionIds: ids, trimmingOptionNames: names }));
}, []);

const setTrimmingStyleRequest = useCallback((text: string) => {
  setFlow(prev => ({ ...prev, trimmingStyleRequest: text }));
}, []);
```

### 4. 新規ページ作成

#### 4-a. `frontend/line-reserve/src/pages/TrimmingCourseSelectPage.tsx` — 新規作成

トリミングコース一覧を表示し、選択させるページ。

```typescript
interface TrimmingCourseSelectPageProps {
  clinicId: string;
  idToken: string;
  onSelect: (courseId: number, courseName: string) => void;
  onBack: () => void;
}

export function TrimmingCourseSelectPage({ clinicId, idToken, onSelect, onBack }: TrimmingCourseSelectPageProps) {
  const [courses, setCourses] = useState<TrimmingCourse[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    liffApi.getTrimmingCourses(clinicId, idToken)
      .then(setCourses)
      .catch(() => setCourses([]))
      .finally(() => setIsLoading(false));
  }, [clinicId, idToken]);

  return (
    <div>
      <BackButton onBack={onBack} />
      <h2>トリミングコースを選択</h2>
      {isLoading ? (
        <p>読み込み中...</p>
      ) : courses.length === 0 ? (
        <p>利用可能なコースがありません</p>
      ) : (
        courses.map(course => (
          <ListItem
            key={course.id}
            title={course.name}
            subtitle={course.price != null ? `¥${course.price.toLocaleString()}` : ''}
            description={course.description}
            onClick={() => onSelect(course.id, course.name)}
          />
        ))
      )}
    </div>
  );
}
```

#### 4-b. `frontend/line-reserve/src/pages/TrimmingOptionSelectPage.tsx` — 新規作成

トリミングオプション一覧を複数選択させるページ。

```typescript
interface TrimmingOptionSelectPageProps {
  clinicId: string;
  idToken: string;
  onNext: (optionIds: number[], optionNames: string[]) => void;
  onBack: () => void;
}

export function TrimmingOptionSelectPage({ clinicId, idToken, onNext, onBack }: TrimmingOptionSelectPageProps) {
  const [options, setOptions] = useState<TrimmingOption[]>([]);
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    liffApi.getTrimmingOptions(clinicId, idToken)
      .then(setOptions)
      .catch(() => setOptions([]))
      .finally(() => setIsLoading(false));
  }, [clinicId, idToken]);

  const toggleOption = useCallback((id: number) => {
    setSelectedIds(prev =>
      prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]
    );
  }, []);

  const handleNext = useCallback(() => {
    const selectedNames = options
      .filter(o => selectedIds.includes(o.id))
      .map(o => o.name);
    onNext(selectedIds, selectedNames);
  }, [selectedIds, options, onNext]);

  return (
    <div>
      <BackButton onBack={onBack} />
      <h2>オプションを選択（任意・複数可）</h2>
      {isLoading ? (
        <p>読み込み中...</p>
      ) : (
        <>
          {options.map(option => (
            <ListItem
              key={option.id}
              title={option.name}
              subtitle={option.price != null ? `¥${option.price.toLocaleString()}` : ''}
              isSelected={selectedIds.includes(option.id)}
              onClick={() => toggleOption(option.id)}
            />
          ))}
          <PrimaryButton onClick={handleNext}>次へ</PrimaryButton>
        </>
      )}
    </div>
  );
}
```

**注意**: `ListItem` コンポーネントに `isSelected` props が必要な場合は、
既存の `ListItem.tsx` を確認・拡張すること。

### 5. `frontend/line-reserve/src/App.tsx` — フロー制御更新

#### 5-a. import 追加

```typescript
import { TrimmingCourseSelectPage } from './pages/TrimmingCourseSelectPage';
import { TrimmingOptionSelectPage } from './pages/TrimmingOptionSelectPage';
```

#### 5-b. `useReservationFlow` の返り値に新メソッドを追加

```typescript
const {
  flow,
  setCustomerInfo, setCourse, setStaff, setDate, setTime, setRequestText, resetFlow,
  setTrimmingCourse, setTrimmingOptions, setTrimmingStyleRequest, // ★ 追加
} = useReservationFlow();
```

#### 5-c. step2 のナビゲーション変更（コース選択後の分岐）

```typescript
// Before:
if (page === 'step2') {
  return (
    <CourseSelectPage
      onSelect={(courseId, courseName) => {
        setCourse(courseId, courseName);
        goTo('step3');  // ← 常に step3 へ
      }}
      ...
    />
  );
}

// After:
if (page === 'step2') {
  return (
    <CourseSelectPage
      onSelect={(courseId, courseName, category) => {
        setCourse(courseId, courseName, category);
        // category='trimming' の場合はトリミングコース選択ステップへ
        goTo(category === 'trimming' ? 'step2b' : 'step3');
      }}
      ...
    />
  );
}
```

**注意**: `CourseSelectPage` が `onSelect` コールバックで `category` を返せるよう修正が必要。
`CourseSelectPage.tsx` の `Course` 型が `category` フィールドを持つことを確認すること。

#### 5-d. 新ページレンダリング追加（step2b / step2c）

```typescript
if (page === 'step2b') {
  return (
    <TrimmingCourseSelectPage
      clinicId={clinicId}
      idToken={idToken}
      onSelect={(trimmingCourseId, trimmingCourseName) => {
        setTrimmingCourse(trimmingCourseId, trimmingCourseName);
        goTo('step2c');
      }}
      onBack={() => goTo('step2')}
    />
  );
}

if (page === 'step2c') {
  return (
    <TrimmingOptionSelectPage
      clinicId={clinicId}
      idToken={idToken}
      onNext={(optionIds, optionNames) => {
        setTrimmingOptions(optionIds, optionNames);
        goTo('step3');
      }}
      onBack={() => goTo('step2b')}
    />
  );
}
```

#### 5-e. ConfirmPage に trimming フィールドを渡す

`flow.trimmingCourseId` / `trimmingOptionIds` / `trimmingStyleRequest` を
`ConfirmPage` に props として渡し、確認画面で選択したコース・オプションを表示する。

`ConfirmPage` 内の `createReservation` 呼び出し時に trimming フィールドを追加:

```typescript
const body: CreateReservationBody = {
  course_id: flow.courseId ?? 0,
  staff_id: flow.staffId,
  date: flow.date,
  start_time: flow.startTime,
  end_time: flow.endTime,
  customer_fields: buildCustomerFields(flow.customerInfo),
  request_text: flow.requestText,
  // ★ 追加: トリミング詳細（general の場合は送信しない）
  ...(flow.courseCategory === 'trimming' && {
    trimming_course_id: flow.trimmingCourseId ?? undefined,
    trimming_option_ids: flow.trimmingOptionIds.length > 0
      ? flow.trimmingOptionIds
      : undefined,
    trimming_style_request: flow.requestText, // requestText = trimming style request として流用
  }),
};
```

## UI 操作フロー（変更後）

**一般予約（category='general'）の場合**: 既存と同じ 7ステップ

**トリミング予約（category='trimming'）の場合**: 9ステップ
1. 顧客情報入力（既存 step1）
2. 予約区分選択 — `category='trimming'` を選択（既存 step2）
3. **[NEW] トリミングコース選択**（step2b）— コース一覧から1つ選択
4. **[NEW] トリミングオプション選択**（step2c）— オプション一覧から0〜N個選択（任意）
5. スタッフ選択（既存 step3）
6. 日付選択（既存 step4）
7. 時刻選択（既存 step5）
8. リクエスト入力（既存 step6）
9. 確認 → 予約確定（既存 step7 / step8）

## 依存関係

- BE-120 完了（`GET /api/liff/:clinicId/trimming-courses`, `/trimming-options` が実装されている）
- BE-118 完了（`Course.category` が API レスポンスに含まれている）

## 完了条件

- [ ] `Course` 型に `category: 'general' | 'trimming'` が追加されている
- [ ] `TrimmingCourse`, `TrimmingOption` 型が追加されている
- [ ] `ReservationFlow` にトリミングフィールドが追加されている
- [ ] `PageType` に `'step2b'` / `'step2c'` が追加されている
- [ ] `CreateReservationBody` に trimming フィールドが追加されている
- [ ] `liff-api.ts` に `getTrimmingCourses`, `getTrimmingOptions` が追加されている
- [ ] `TrimmingCourseSelectPage.tsx` が新規作成されている
- [ ] `TrimmingOptionSelectPage.tsx` が新規作成されている
- [ ] `App.tsx` が `category='trimming'` の場合に step2b → step2c を経由する
- [ ] 一般予約（category='general'）は従来通り step2 → step3 に遷移する
- [ ] `ConfirmPage` が trimming フィールドを含めて予約を確定する
- [ ] TypeScript エラー 0件
