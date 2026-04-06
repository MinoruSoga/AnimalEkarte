# BUG-151: API レスポンス構造がエンドポイントごとに不統一

## 概要
一覧取得 API のレスポンス構造が2パターン混在している:
- **パターンA**: `{ data: [...], total: N, page: N, per_page: N }` — ページネーション対応
- **パターンB**: `[...]` — 配列直接返却

フロントエンドで各エンドポイントごとにレスポンスの扱いを分岐する必要があり、コードの複雑性が増す。

## 影響範囲

### パターンA（ページネーション対応）— 11 エンドポイント
| エンドポイント | total |
|--------------|-------|
| owners | ✅ |
| reservations | ✅ |
| medical-records | ✅ |
| examinations | ✅ |
| trimmings | ✅ |
| vaccinations | ✅ |
| accountings | ✅ |
| hospitalizations | ✅ |
| inventory | ✅ |
| estimates | ✅ |
| masters/medicines | ✅ |

### パターンB（配列直接）— 15 エンドポイント
| エンドポイント |
|--------------|
| shifts |
| masters/staffs |
| masters/animal-species |
| masters/permission-groups |
| masters/occupations |
| masters/insurances |
| masters/merchandise-items |
| masters/service-types |
| masters/trimming-courses |
| masters/hospitalization-plans |
| masters/cages |
| masters/diagnosis-categories |
| masters/diagnosis-names |
| masters/checkup-types |
| masters/inquiry-templates |

## 期待する動作
全一覧 API で統一された構造:
```json
{
  "data": [...],
  "total": 100,
  "page": 1,
  "per_page": 20
}
```

マスタ系で件数が少なくページネーション不要な場合でも、構造は統一すべき。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/api.md`
> "Return consistent error response format"

エラーだけでなく、正常レスポンスも一貫した構造であるべき。

## 優先度
**Low** — 機能的には動作する。コードの一貫性・保守性の問題。

## 関連ファイル
- `backend/internal/handler/*.go` — 各ハンドラの List メソッド
