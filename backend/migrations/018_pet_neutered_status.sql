-- 避妊・去勢の施術状況フラグ。既存の neutered_date（手術日）は日付しか保持できず
-- 「施術済みだが日付不明」「未施術」を区別できなかったため、独立したステータスを追加する。
-- 値は性別中立（オス=去勢/メス=避妊を共通の done で表現し、ラベル差は表示層が担う）。
-- unknown=不明（既定）/ not_done=未施術 / done=施術済み。
-- Do not edit 001_init.sql. User must run: make migrate

CREATE TYPE pet_neutered_status AS ENUM ('unknown', 'not_done', 'done');

ALTER TABLE pets
    ADD COLUMN neutered_status pet_neutered_status NOT NULL DEFAULT 'unknown';

-- 既存データ: 手術日が記録済みのペットは施術済みとみなして backfill する。
UPDATE pets
   SET neutered_status = 'done'
 WHERE neutered_date IS NOT NULL;
