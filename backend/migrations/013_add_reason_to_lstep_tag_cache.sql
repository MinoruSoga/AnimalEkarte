-- 013: lstep_tag_cache に reason カラムを追加
-- タグ付与理由（例: "最終健診: 2025-11-01"）を保存する。NULL は理由なし。
ALTER TABLE lstep_tag_cache ADD COLUMN IF NOT EXISTS reason TEXT;
