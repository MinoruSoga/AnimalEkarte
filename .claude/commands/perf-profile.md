---
description: パフォーマンスプロファイリング実行（Go pprof、Lighthouse）
---

# /perf-profile

Backend (pprof) と Frontend (Lighthouse) のパフォーマンス分析を実行します。

## 実行内容

### Backend Go Profiling
```bash
docker compose exec backend go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

### Frontend Lighthouse
```bash
docker compose exec frontend npm run lighthouse
```

### 分析対象
- API レスポンスタイム（p95）
- メモリ使用量
- CPU 使用率
- Core Web Vitals
- バンドルサイズ

## 出力形式

```
## Backend パフォーマンス
- API p95: XX ms
- Memory: XX MB
- Goroutine: XX

## Frontend パフォーマンス
- LCP: XX s
- INP: XX ms
- CLS: XX
- Bundle Size: XX KB
```

## 使用エージェント

`performance-optimizer` (Sonnet) を自動起動
