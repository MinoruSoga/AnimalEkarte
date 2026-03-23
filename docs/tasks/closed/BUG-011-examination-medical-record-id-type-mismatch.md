# BUG-011: 検査登録でmedical_record_idがstring型で送信されHTTP 400

## 種類
バグ（フロントエンド型変換不備）

## 発見日
2026-03-23

## 再現手順
1. `/examinations/select-pet` でペットを選択
2. `/examinations/new?petId=N` でフォームを入力
3. 保存ボタンをクリック

## 期待動作
`POST /api/v1/examinations` が HTTP 201 で成功する

## 実際の動作
`POST /api/v1/examinations` が HTTP 400:
```
json: cannot unmarshal string into Go struct field
createExaminationRequest.medical_record_id of type uint64
```

## 根本原因
フロントエンドから `medical_record_id` が string型（URLパラメータから取得した値）で送信されている。
Go バックエンドは `uint64` 型を期待しているため型不一致エラー。

## 影響範囲
- 検査登録が完全に機能しない
- 重症度: **高**（検査管理の基本機能が使用不可）

## 実際の根本原因（コード調査後）

当初の推測より広範な問題だった:

1. `medical_record_id: ""` — URLパラメータなし、空文字ハードコード
2. `exam_type_id` — IDでなく名前文字列を送信（Selectのvalueがname）
3. `pet_id` — string(FE)をnumberに変換していない
4. `CreateExaminationRequest`の全IDフィールドが`string`型（BEはuint64）
5. `setFormData`がmergeでなくreplaceだったため複数フィールド更新が壊れていた

## 修正内容（2026-03-23）

- `types/index.ts`: `ExaminationRecord`に`testTypeId?`, `doctorId?`追加
- `features/examinations/api/types.ts`: 全IDフィールドを`number`型に変更
- `features/examinations/api/transforms.ts`: `testTypeId`, `doctorId`をpopulate
- `features/examinations/routes/ExaminationForm.tsx`:
  - URLパラメータ`medicalRecordId`読み込み追加
  - SelectItemのvalueをidに変更（name→id）
  - `onValueChange`でnameとidを両方setFormData
- `features/examinations/hooks/use-examination-form.ts`:
  - `medicalRecordId`をURLパラメータから読み込み
  - request構築で`Number()`変換を適用
  - `setFormData`を`prev => ({ ...prev, ...next })`に修正（merge）

## ⚠️ 残存課題

`medical_record_id: 0`はフォールバック値。検査作成フローが独立している場合、
バックエンドFK制約で失敗する可能性あり。
医療記録コンテキストからの遷移時はURLに`?medicalRecordId=N`を付与すること。

## クローズ日
2026-03-23
