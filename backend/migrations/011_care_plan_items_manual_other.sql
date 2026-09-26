-- EMR-179: ケアプラン持ち物（type=item）明細の手入力「その他」対応。
-- 従来 chk_care_plan_item_ref は type=item に hospitalization_plan_id を必須とした。
-- 手入力行（category='other' かつ trim 後非空の other_reason を持つ行）では
-- マスタ参照 NULL を許容する。billing_items の手入力「その他」契約
-- （理由必須・trim・500文字以内は service 層で担保）と同型。
--
-- 注: other_reason の長さ上限（500 rune）はアプリ層検証であり CHECK には含めない
-- （CHECK は NULL 参照の開放条件のみに限定し、billing_items と同じ責務分担を保つ）。

ALTER TABLE care_plan_items
    ADD COLUMN IF NOT EXISTS other_reason text NOT NULL DEFAULT '';

ALTER TABLE care_plan_items
    DROP CONSTRAINT IF EXISTS chk_care_plan_item_ref;

ALTER TABLE care_plan_items
    ADD CONSTRAINT chk_care_plan_item_ref CHECK (
        (type = 'medicine'    AND medicine_id IS NOT NULL) OR
        (type = 'treatment'   AND procedure_id IS NOT NULL) OR
        (type = 'item'        AND hospitalization_plan_id IS NOT NULL) OR
        (type = 'item'        AND hospitalization_plan_id IS NULL
            AND category = 'other' AND btrim(other_reason) <> '') OR
        (type IN ('food', 'instruction'))
    );
