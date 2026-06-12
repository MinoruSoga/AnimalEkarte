# パフォーマンス測定・プロファイリングガイド (Performance & Profiling)

> **Animal Ekarte**: 大規模データ下での高速な操作性の維持
> **最新更新**: 2026-06-12

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
`pprof` を使用して CPU およびメモリ消費の激しい処理を特定します。
```bash
# CPU プロファイルの取得
curl -s http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof -http=:8081 cpu.prof
```

### 3.3 データベース (PostgreSQL)
`EXPLAIN ANALYZE` を使用して、集計クエリのインデックス効力を検証します。
- **重点監視**: `billings`, `lstep_delivery_trigger_log` などの成長率の高いテーブル。

---

## 4. 負荷テスト (Load Testing)

`k6` を使用し、想定される同時接続数（最大50名）での安定稼働を検証します。
```bash
# 負荷テストの実行
k6 run load-tests/k6-spike-test.js
```

---
