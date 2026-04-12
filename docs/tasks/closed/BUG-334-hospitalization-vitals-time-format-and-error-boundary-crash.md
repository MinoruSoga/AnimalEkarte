# BUG-334: 入院管理バイタル保存で400エラー後にErrorBoundaryクラッシュ

**Status**: OPEN  
**Priority**: High  
**Discovery**: 機能テスト Section 7 入院・ホテル管理 (2026-04-12)

## 概要

`/hospitalization/:id` の日次記録タブでバイタルを追加すると、バックエンドが 400 を返し、エラートースト表示ではなく React Router の ErrorBoundary（「エラーが発生しました」ページ）が表示される。根本原因は2つある：
1. **時刻フォーマット不一致**: フロントエンドが `"HH:MM"` を送信するがバックエンドは `"HH:MM:SS"` を要求
2. **try-catch 欠如**: `mutateAsync` を `startVitalTransition` 内で呼び出しているが try-catch がなく、400 エラーが React ErrorBoundary まで伝播する

## 再現手順

1. 管理者アカウントでログイン
2. `/hospitalization` で入院中のペットを開く
3. 日次記録タブを開き、バイタル追加フォームを開く
4. 時刻に `"18:03"` 等を入力し保存
5. **結果**: 「エラーが発生しました」ページにクラッシュする（ErrorBoundary 発動）
6. **期待**: エラートーストが表示され、フォームはそのまま残る

## 現状コード

### Sub-bug 1: 時刻フォーマット不一致

#### `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyVitalsSection.tsx:66-80`
```ts
const handleSubmit = useCallback(() => {
    if (!form.time) return;

    const payload: CreateVitalRecordRequest = {
        time: form.time,  // ← "18:03"（HH:MM）をそのまま送信
    };
    // ...
    onAddVital(payload);
}, [form, onAddVital]);
```

#### `backend/internal/handler/daily_record_handler.go:125-129`
```go
parsedVitalTime, err := time.Parse("15:04:05", req.Time)
if err != nil {
    // "18:03" を "15:04:05" パターンで解析 → 必ず失敗 → 400 返却
    RespondError(c, apperrors.WrapInvalidInput("invalid time format, expected HH:MM:SS"))
    return
}
```

### Sub-bug 2: try-catch 欠如 → ErrorBoundary クラッシュ

#### `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx:83-108`
```ts
const handleAddVital = useCallback(
    (payload: CreateVitalRecordRequest) => {
        startVitalTransition(async () => {
            await createVital.mutateAsync(payload);  // ← try-catch なし
            // 400 エラー → unhandled rejection → React ErrorBoundary へ伝播
        });
    },
    [createVital]
);

const handleAddCareLog = useCallback(
    (payload: CreateCareLogRequest) => {
        startCareLogTransition(async () => {
            await createCareLog.mutateAsync(payload);  // ← 同じ問題
        });
    },
    [createCareLog]
);

const handleAddStaffNote = useCallback(
    (payload: CreateStaffNoteRequest) => {
        startStaffNoteTransition(async () => {
            await createStaffNote.mutateAsync(payload);  // ← 同じ問題
        });
    },
    [createStaffNote]
);
```

### 比較: 正しい実装（プロジェクト内参照実装）

```ts
// frontend/src/features/owners/hooks/use-owner-form.ts:226-228 — try-catch + handleApiError
} catch (error) {
    handleApiError(error, "保存");
    return { ...prevState, success: false, timestamp: Date.now() };
}
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyVitalsSection.tsx:70` | `time: form.time` — HH:MM を送信 | 要修正 |
| `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx:85-86` | `handleAddVital` — try-catch なし | 要修正 |
| `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx:94-95` | `handleAddCareLog` — try-catch なし | 要修正 |
| `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx:103-104` | `handleAddStaffNote` — try-catch なし | 要修正 |

## 修正方針

### 1. 時刻フォーマット修正 — `DailyVitalsSection.tsx:70`

```ts
const payload: CreateVitalRecordRequest = {
    time: form.time.length === 5 ? `${form.time}:00` : form.time,  // "HH:MM" → "HH:MM:SS"
};
```

### 2. try-catch 追加 — `DailyRecordsTab.tsx:83-108`

```ts
import { handleApiError } from "@/lib/handle-api-error";

const handleAddVital = useCallback(
    (payload: CreateVitalRecordRequest) => {
        startVitalTransition(async () => {
            try {
                await createVital.mutateAsync(payload);
            } catch (error) {
                handleApiError(error, "バイタル追加");
            }
        });
    },
    [createVital]
);

const handleAddCareLog = useCallback(
    (payload: CreateCareLogRequest) => {
        startCareLogTransition(async () => {
            try {
                await createCareLog.mutateAsync(payload);
            } catch (error) {
                handleApiError(error, "ケアログ追加");
            }
        });
    },
    [createCareLog]
);

const handleAddStaffNote = useCallback(
    (payload: CreateStaffNoteRequest) => {
        startStaffNoteTransition(async () => {
            try {
                await createStaffNote.mutateAsync(payload);
            } catch (error) {
                handleApiError(error, "スタッフメモ追加");
            }
        });
    },
    [createStaffNote]
);
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/error-handling.md` — handleApiError 必須

> **すべての `catch` ブロックで `handleApiError` を呼び出す。**
> ```typescript
> // ✅ MANDATE: すべての catch ブロックで handleApiError を使用
> try {
>   await api.updateOwner(id, data);
> } catch (error) {
>   handleApiError(error, "オーナーの更新");
> }
> ```

`mutateAsync` を `startTransition` 内で呼び出す場合も同様に try-catch が必要。

### プロジェクト内参照実装

- `frontend/src/features/owners/hooks/use-owner-form.ts:226-228` — mutateAsync の try-catch + handleApiError パターン

## 優先度

**High** — バイタル保存操作でアプリ全体がクラッシュし、ユーザーは手動でブラウザ更新しなければ操作を継続できない。入院管理の中核機能が使用不能になる。

## 関連チケット

なし

## 関連ファイル

- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyVitalsSection.tsx:66-80` — 時刻フォーマット修正対象
- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx:83-108` — try-catch 追加対象
- `backend/internal/handler/daily_record_handler.go:125-129` — バックエンド時刻パース（正常・変更不要）
- `frontend/src/lib/handle-api-error.ts` — エラーハンドリング共通処理
