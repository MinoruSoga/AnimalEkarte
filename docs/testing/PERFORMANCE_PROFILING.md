# パフォーマンス測定・プロファイリングガイド (Performance & Profiling)

> **目的**: パフォーマンスプロファイリングの手順を定義する。
> **読者**: 実装者。
> **タイミング**: パフォーマンス調査時。

> **Animal Ekarte**: 大規模データ下での高速な操作性の維持
> **最新更新**: 2026-07-10

---

## 1. 概要

本ガイドは、システムのレスポンスタイムの測定方法、ボトルネックの特定手順、およびそれらを解消するためのプロファイリング手法を定義します。

---

## 2. パフォーマンス目標 (KPIs)

臨床現場でのストレスを最小限にするため、以下の基準を設定しています。

- **画面初期表示**: 1.5 秒以内。
- **検索（インクリメンタル）**: 200ms 以内（Debounce 後）。
- **保存アクション**: 1.0 秒以内（非同期完了まで）。
- **集計（月次レポート）**: 3.0 秒以内。

---

## 3. プロファイリング手法

### 3.1 フロントエンド (React)
React DevTools の **Profiler** タブを使用して、不要な再描画（Re-render）を特定します。
- **重点監視**: `medical-records`, `accounting`, `reception` の各フォーム。
- **対策**: `memo()`, `useCallback`, `useMemo` によるコンポーネントの保護。

### 3.2 バックエンド (Go)
CLI ツール `backend/scripts/profile.go` は CI ランナーのホスト自身（起動したての別プロセス）をプロファイリングしており、稼働中の dockerized backend の実測になっていなかったため 2026-07-10 に削除済み（commit `3f692a73`）。バックエンドは HTTP の `/debug/pprof` エンドポイントも公開していない。現時点で稼働中の backend プロセスに対する専用プロファイリング手段は存在しない（`net/http/pprof` を `GIN_MODE=debug` 時のみ有効化する案は未実装）。

### 3.3 データベース (PostgreSQL)
`EXPLAIN ANALYZE` を使用して、集計クエリのインデックス効力を検証します。
- **重点監視**: `billings`, `lstep_delivery_trigger_log` などの成長率の高いテーブル。

### 3.4 N+1 クエリ回帰テスト (Performance Regression Tests)
ループ内での個別クエリ実行や設定フェッチによる N+1 問題の発生を未然に防ぐため、テストコードによる自動回帰テストを導入しています。
- **実装場所**: `backend/internal/service/perf_n1_regression_test.go`
- **検証アプローチ**: 依存サービスのメソッド呼び出し回数をモック（Spy）でカウントし、ループの回数（N回）ではなく 1 回のみのフェッチ（ループ外へのホイスト）で処理できているかをテストコードでアサート（Assert）します。
  - 例: `SyncHealthPreventionTagsForClinic` 内で `GetHealthPreventionThresholds` を呼び出す回数が、飼い主数 `N` によらず `1` 回に抑えられているかを検証します。
- **実行方法**:
  ```bash
  docker compose exec backend go test -v ./internal/service/ -run "TestPERF|TestH1"
  ```

---

## 4. 負荷テスト (Load Testing)

`k6` を使用し、突発的な負荷増加（`load-tests/k6-spike-test.js` は 5 秒で 100 ユーザーへ急増するスパイクを想定）に対する安定稼働を検証します。定常負荷（50 ユーザー、p95 500ms 以下）は `load-tests/k6-api-endpoints.js` で別途検証します。
```bash
# 負荷テストの実行
k6 run load-tests/k6-spike-test.js
```

---
