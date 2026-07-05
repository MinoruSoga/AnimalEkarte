# UI改善タスク: 受付ヘッダー テレメトリ表示エリア

> 出典: UI提案 v3 (2026-07-04) のうち採用決定分のみ。カルテ再配置・会計精算パッド等の他提案は**不採用**であり本タスクに含まない。
> 上位原則: [docs/PRODUCT_PHILOSOPHY.md](docs/PRODUCT_PHILOSOPHY.md) — 実践ゲート通過記録を本文に含む。

## 背景・目的（業務目的で記述）

受付スタッフ・院長が「今日どれだけ混んでいて、誰を一番待たせているか」を、Kanbanカラムの目視カウントなしで把握できるようにする。混雑把握→判断（声かけ・診察応援）までの時間を削る。

## 哲学ゲート通過記録

- ① 要件の責任者: **Soga**。要件は業務目的（混雑把握工程の削減）で記述済み。
- ② 削除する工程: カラム件数の目視カウント / 最長待ち患者を探す視線スキャン。
- ③ 例外だけを見せる: 最長待ちは**患者名つきで1名だけ**表示する。全件を人間に走査させない。
- ④ 改善メトリクス: 「受付から診察開始までの待ち時間」（PHILOSOPHY 4-1 の筆頭メトリクス）。本表示はこのメトリクスの常時計測・可視化そのものであり、以後の受付改善の測定基盤になる。
- ⑤ 自動化なし（表示のみ）— 対象外。

## スコープ

- **IN**: 当日受付ページ（`Reception.tsx`）の `FormHeader` 直下に置く水平テレメトリ・ストリップ 1 本。
- **OUT**: カード側の待ち時間着色、カード直アクションボタン、その他 v3 提案の全て（不採用）。推測オーバー実装禁止。

## 仕様

表示は 3 項目のみ。左から:

```
本日受付 32件   ·   平均待ち 14分   ·   最長待ち 32分 — ミルク
```

- 集計は**フィルタ非適用の全体値**（`filteredColumns` ではなく `columns` から算出）。フィルタで数字が変わると「今日の全体量」の意味が壊れる。
- 「本日受付」= 当日の全ステータス件数合計（cancelled / no_show を除く）。
- 「待ち」対象 = `checked_in`（受付済）ステータスの患者。待ち時間 = 現在時刻 − `checked_in_at`。
- 受付済 0 件時、平均・最長は「—」表示。
- デザインは DESIGN.md 準拠: 数値は太字 + `tabular-nums`、ラベルは muted。最長待ちの強調は文字色（orange系テキスト）のみで、構造色 `#0075de` は使わない。

## 段階導入（薄い縦切り — PHILOSOPHY ④-2）

### Phase 1 — FE のみ（即日可）

- 「本日受付 N件」のみ表示。`use-reception-kanban.ts` の `columns` から `useMemo` で導出。
- BE 変更・新規 API なし。平均待ち/最長待ちはデータが存在しないため**この段階では出さない**（下記データギャップ参照）。

### Phase 2 — BE: `checked_in_at` 追加

- **データギャップ（調査済み）**: `reservations` に受付時刻カラムは存在しない。`UpdatedAt` は `autoUpdateTime` のため予約編集でリセットされ、待ち時間の算出に流用してはならない。
- migration: `reservations.checked_in_at TIMESTAMPTZ NULL` を**新規 migration ファイルで追加**（適用済み migration の編集は checksum mismatch を起こすため禁止）。
- service 層: ステータスが `checked_in` へ遷移した時点で `checked_in_at = now()` を記録。「受付済へ戻して再受付」した場合は上書きする（待ち直しとみなす — 最後の checked_in 遷移時刻を採用）。
- レスポンス DTO / FE transforms に `checked_in_at` を追加し、平均・最長を表示。
- 更新頻度: 既存の一覧取得に相乗りし、表示側で 60 秒間隔の再計算。専用ポーリング追加は禁止。

## 実装ノート

| 対象 | 内容 |
|---|---|
| `frontend/src/features/reception/components/ReceptionTelemetryStrip.tsx` | 新規。表示専用コンポーネント（props で集計値を受ける presentational） |
| `frontend/src/features/reception/routes/Reception.tsx` | `FormHeader` 直下に装着。集計は `columns` からの `useMemo` 派生（追加 fetch 禁止） |
| `frontend/src/features/reception/api/transforms.ts` | Phase 2: `checked_in_at` のマッピング追加 |
| `backend` reservation service / handler | Phase 2: `checked_in` 遷移時の記録 + DTO 追加 |
| `backend/migrations/` | Phase 2: 新規ファイルで `checked_in_at` 追加 |

## 受け入れ基準

- [ ] 「本日受付」合計がフィルタ操作に影響されない（全体値のまま）
- [ ] 受付済 0 件で平均・最長が「—」
- [ ] 予約内容の編集（時間変更等）で待ち時間が変動しない（`UpdatedAt` 非依存の実証）
- [ ] 受付済→戻す→再受付で待ち時間がリセットされる
- [ ] Phase 2 の migration は新規ファイル追加のみ（既存 migration 変更なし）

## 検証（スコープ限定）

```
docker compose exec frontend npx vitest run src/features/reception
docker compose exec backend go test ./internal/service/...   # Phase 2 のみ
```
