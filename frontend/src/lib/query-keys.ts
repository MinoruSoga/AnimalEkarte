/**
 * React Query キーファクトリー
 *
 * TanStack Query の階層一致によるキャッシュ無効化を確実にするため、
 * queryKey を一元管理する。
 *
 * 使用方法:
 *   queryKey: queryKeys.accountings.detail(id)
 *   invalidateQueries({ queryKey: queryKeys.accountings.all() })
 *
 * 命名規則:
 *   all()      → 当該エンティティの全キャッシュを無効化したい場合
 *   detail(id) → 詳細キャッシュ (id 指定)
 */

export const queryKeys = {
  accountings: {
    /** リスト全体の無効化に使用 */
    all: () => ["accountings"] as const,
    /** 詳細キャッシュ。旧 "accounting-detail" に相当 */
    detail: (id: string) => ["accountings", id] as const,
  },
  masters: {
    /** 汎用マスタカテゴリキー。"masterItems" の代わりにこれを使う */
    category: (name: string) => ["masters", name] as const,
  },
} as const;

/**
 * 認証ユーザー情報 (/me) の query key。
 * auth feature の useGetMe と、権限変更後の invalidateQueries（master 等）で共有する。
 */
export const ME_QUERY_KEY = ["me"] as const;
