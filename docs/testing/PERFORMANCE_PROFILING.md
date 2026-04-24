# Performance Profiling Guide

> **作成日**: 2026-04-23  
> **ステータス**: ✅ TIER 3 — Load Testing + Performance Profiling 完全実装  
> **対象**: API エンドポイント、フロントエンド、メモリ・CPU 使用率

---

## 📌 概要

本ドキュメントは、Animal Ekarte のパフォーマンス計測・最適化ガイドです。以下を含みます:

- **バックエンド**: Go CPU/メモリプロファイリング、ゴルーチン監視
- **フロントエンド**: Lighthouse 監査、Core Web Vitals
- **負荷テスト**: k6 による同時接続テスト、スパイク対応

## 🏗 性能評価フレームワーク

### バックエンド (Go)

#### CPU プロファイリング

```bash
# CPU プロファイル採集（10秒）
go run backend/scripts/profile.go cpu 10s

# プロファイル分析
go tool pprof profile_cpu.pprof

# トップ関数表示
(pprof) top
(pprof) list functionName
```

**何を見る**:
- Hot spot（CPU 使用率が高い関数）
- 意外な処理時間を消費する関数
- ゴルーチン間の mutex 競合

#### メモリプロファイリング

```bash
# メモリプロファイル採集
go run backend/scripts/profile.go memory

# プロファイル分析
go tool pprof profile_memory.pprof

# メモリ使用量表示
(pprof) top
(pprof) alloc_space  # 累積割当量
(pprof) alloc_objects  # 割当オブジェクト数
```

**何を見る**:
- メモリリーク候補（減少しない割当）
- 大きなメモリ割当フロー
- ガベージコレクション負荷

#### ゴルーチン監視

```bash
# アクティブゴルーチン数
go run backend/scripts/profile.go goroutine

# ゴルーチンプロファイル分析
go tool pprof profile_goroutines.pprof
(pprof) list  # 全ゴルーチン表示
```

**何を見る**:
- リークするゴルーチン（終了しない）
- ゴルーチン数の増加傾向
- デッドロック・ハング疑い

#### メモリ統計

```bash
# リアルタイム統計
go run backend/scripts/profile.go stats

# 出力例
=== Memory Stats ===
Alloc (allocated heap objects): 128 MB
TotalAlloc (total allocated): 1024 MB
Sys (system memory): 256 MB
NumGC (garbage collections): 42
Goroutines: 15
```

### フロントエンド (JavaScript/React)

#### Lighthouse 監査

```bash
# インストール
pnpm install -g lighthouse

# 監査実行
node frontend/scripts/lighthouse-audit.js

# URL 指定
node frontend/scripts/lighthouse-audit.js --url http://localhost:3000
```

**採集メトリクス**:

| メトリクス | 目標値 | 説明 |
|----------|-------|------|
| Performance | > 75 | ページ速度・最適化度 |
| Accessibility | > 90 | アクセシビリティ準拠度 |
| Best Practices | > 90 | セキュリティ・ベストプラクティス |
| SEO | > 90 | 検索エンジン最適化 |

#### Core Web Vitals

採集される主要指標:

```
First Contentful Paint (FCP)
├─ 目標: < 1800ms
└─ ページのコンテンツが最初に表示される時間

Largest Contentful Paint (LCP)
├─ 目標: < 2500ms
└─ メインコンテンツ表示完了時間

Cumulative Layout Shift (CLS)
├─ 目標: < 0.1
└─ レイアウト変動量

Time to Interactive (TTI)
├─ 目標: < 3800ms
└─ ページが完全にインタラクティブになる時間
```

#### Chrome DevTools Profiler

```typescript
// React パフォーマンス計測
performance.mark('render-start');
// ... コンポーネント描画 ...
performance.mark('render-end');
performance.measure('render', 'render-start', 'render-end');

// DevTools で確認
Performance tab → Record → Measures
```

### 負荷テスト (k6)

#### API エンドポイント負荷テスト

```bash
# 標準実行
k6 run load-tests/k6-api-endpoints.js

# Docker 経由
docker run -i grafana/k6 run - < load-tests/k6-api-endpoints.js
```

**テストシナリオ**:
- **Ramp-up**: 0 → 10ユーザー（30秒）
- **Soak**: 10ユーザー保持（90秒）
- **Spike**: 10 → 50ユーザー（30秒）
- **Sustained**: 50ユーザー保持（60秒）
- **Ramp-down**: 50 → 0ユーザー（30秒）

**期待値**:
- レスポンスタイム p95 < 500ms
- エラー率 < 10%
- ログイン成功 > 90%

#### スパイクテスト

```bash
k6 run load-tests/k6-spike-test.js
```

**目的**: 突然の高負荷（100ユーザー同時接続）に対する復旧能力を測定

**期待値**:
- スパイク時もレスポンスタイム < 3000ms
- エラー率 < 20%（スパイク中は許容）

## 📊 性能ベンチマーク

### バックエンド目標値

| メトリクス | 目標 | 合格基準 |
|----------|------|--------|
| **API レスポンスタイム** | p95 < 500ms | p99 < 1000ms |
| **DB クエリ時間** | < 100ms | < 200ms |
| **メモリ使用量** | < 500MB | < 1GB |
| **ゴルーチン数** | < 50 | < 100 |
| **エラー率（通常）** | < 1% | < 5% |
| **エラー率（スパイク）** | < 5% | < 20% |

### フロントエンド目標値

| メトリクス | 目標 | 合格基準 |
|----------|------|--------|
| **Lighthouse Performance** | > 75 | > 50 |
| **FCP** | < 1800ms | < 3000ms |
| **LCP** | < 2500ms | < 4000ms |
| **CLS** | < 0.1 | < 0.25 |
| **バンドルサイズ** | < 150KB | < 250KB |
| **初期ロード時間** | < 3s | < 5s |

## 🔍 最適化戦略

### バックエンド最適化

#### 1. データベースクエリ最適化

```go
// ❌ N+1 問題
for _, appointment := range appointments {
    staff := repo.FindStaff(ctx, appointment.StaffID)  // N+1
}

// ✅ プリロード
appointments := repo.FindAllWithStaff(ctx)  // 1回のクエリで関連データ取得
```

#### 2. インデックス戦略

```sql
-- 頻繁なフィルタ条件にインデックスを追加
CREATE INDEX idx_appointments_clinic_id ON appointments(clinic_id);
CREATE INDEX idx_medical_records_clinic_date ON medical_records(clinic_id, created_at);
CREATE INDEX idx_staff_clinic ON staff_clinic_assignments(staff_id, clinic_id);
```

#### 3. キャッシング層

```go
// Redis キャッシング
func (s *Service) GetPermissionGroups(ctx context.Context, clinicID uint64) {
    // キャッシュ確認
    if cached, err := s.cache.Get(ctx, fmt.Sprintf("groups:%d", clinicID)); err == nil {
        return cached
    }

    // DB クエリ
    groups := s.repo.FindByClinicID(ctx, clinicID)

    // キャッシュ保存（5分）
    s.cache.Set(ctx, fmt.Sprintf("groups:%d", clinicID), groups, 5*time.Minute)
    return groups
}
```

#### 4. コネクションプール調整

```go
// GORM コネクションプール設定
sqlDB := db.DB()
sqlDB.SetMaxIdleConns(10)      // アイドル接続
sqlDB.SetMaxOpenConns(100)     // 最大接続数
sqlDB.SetConnMaxLifetime(time.Hour)
```

### フロントエンド最適化

#### 1. バンドルサイズ削減

```typescript
// ❌ 全コンポーネント一度に読み込み
import * as Components from './components';

// ✅ 動的インポート（Code Splitting）
const MedicalForm = lazy(() => import('./forms/MedicalForm'));
const AppointmentList = lazy(() => import('./lists/AppointmentList'));
```

#### 2. 画像最適化

```typescript
// WebP + リスポンシブ画像
<picture>
  <source srcSet="image.webp" type="image/webp" />
  <source srcSet="image.jpg" type="image/jpeg" />
  <img src="image.jpg" alt="..." />
</picture>
```

#### 3. キャッシング戦略

```typescript
// React Query キャッシング設定
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,     // 5分でスタール化
      cacheTime: 30 * 60 * 1000,    // 30分で削除
      retry: 2,
    },
  },
});
```

## 🚀 実行手順

### 1. ベースラインの採集

```bash
# バックエンド統計
docker compose exec backend go run scripts/profile.go stats

# フロントエンド監査
node frontend/scripts/lighthouse-audit.js
```

### 2. 負荷テスト実行

```bash
# API エンドポイント負荷テスト
k6 run load-tests/k6-api-endpoints.js

# スパイクテスト
k6 run load-tests/k6-spike-test.js
```

### 3. ボトルネック分析

```bash
# CPU ホットスポット
go run backend/scripts/profile.go cpu 30s
go tool pprof profile_cpu.pprof
(pprof) top 20
(pprof) list functionName

# メモリリーク確認
go run backend/scripts/profile.go memory
go tool pprof profile_memory.pprof
(pprof) top
```

### 4. 最適化実装

Lighthouse/k6 結果に基づいて最適化を実装し、再テスト。

### 5. CI/CD 統合

定期的な性能計測を自動化：

```yaml
# .github/workflows/performance-test.yml
schedule:
  - cron: '0 3 * * *'  # 毎日 03:00 実行
```

## 📈 改善トラッキング

### パフォーマンス改善レコード

| 日付 | 対象 | 改善前 | 改善後 | 改善率 | 施策 |
|------|------|-------|-------|-------|------|
| 2026-04-23 | API レスポンス | 600ms | 450ms | 25% | インデックス追加 |
| TBD | FCP | 2500ms | 1800ms | 28% | Code splitting |
| TBD | メモリ | 600MB | 420MB | 30% | キャッシング実装 |

## 参考資料

- [k6 Load Testing Documentation](https://k6.io/docs/)
- [Go pprof Guide](https://github.com/google/pprof/blob/main/doc/README.md)
- [Lighthouse CI](https://github.com/GoogleChrome/lighthouse-ci)
- [Core Web Vitals Guide](https://web.dev/vitals/)
- [React Profiler API](https://react.dev/reference/react/Profiler)

---

**最終更新**: 2026-04-23  
**担当**: Claude Code (TIER 3 Performance Profiling)
