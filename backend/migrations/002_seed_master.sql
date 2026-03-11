-- Master Items Seeder（14カテゴリ・92件）
-- 動物病院向けマスタデータ一括投入
-- master_category ENUM値準拠
-- 重複実行時も安全なON CONFLICT DO NOTHING付き

-- ====================================================================================
-- 1. 検査マスタ（10件）
-- ====================================================================================
INSERT INTO master_items (code, name, category, price, status, description)
VALUES
  ('EXAM-001', '血液検査（CBC）', 'examination', 3500.00, 'active', '赤血球、白血球、血小板数測定'),
  ('EXAM-002', '生化学検査', 'examination', 4500.00, 'active', '肝機能、腎機能、電解質等'),
  ('EXAM-003', '尿検査', 'examination', 1500.00, 'active', 'pH、蛋白、糖、ウロビリノーゲン'),
  ('EXAM-004', '便検査', 'examination', 1500.00, 'active', '寄生虫卵、原虫検査'),
  ('EXAM-005', 'X線検査（胸部）', 'examination', 3000.00, 'active', '胸部単純X線撮影'),
  ('EXAM-006', 'X線検査（腹部）', 'examination', 3000.00, 'active', '腹部単純X線撮影'),
  ('EXAM-007', 'エコー検査（腹部）', 'examination', 4000.00, 'active', '腹部超音波検査'),
  ('EXAM-008', '心電図検査', 'examination', 2000.00, 'active', 'ECG測定'),
  ('EXAM-009', 'フィラリア抗原検査', 'examination', 2000.00, 'active', 'DirofilariaImmitis抗原検査'),
  ('EXAM-010', 'ノミマダニ検査', 'examination', 1500.00, 'active', '寄生虫駆除チェック')
ON CONFLICT DO NOTHING;

-- ====================================================================================
-- 2. 予防接種マスタ（6件）
-- ====================================================================================
INSERT INTO master_items (code, name, category, price, status, description)
VALUES
  ('VAC-001', '犬5種混合ワクチン', 'vaccine', 5000.00, 'active', 'ジステンパー、犬伝染性肝炎等'),
  ('VAC-002', '犬7種混合ワクチン', 'vaccine', 6500.00, 'active', '5種+レプトスピラ2種'),
  ('VAC-003', '狂犬病ワクチン', 'vaccine', 3500.00, 'active', '法定予防接種（年1回）'),
  ('VAC-004', '猫3種混合ワクチン', 'vaccine', 4500.00, 'active', 'ウイルス性鼻気管炎、カリシウイルス、パンレウコペニア'),
  ('VAC-005', '猫5種混合ワクチン', 'vaccine', 6000.00, 'active', '3種+白血病+クラミジア'),
  ('VAC-006', '猫白血病（FeLV）ワクチン', 'vaccine', 3500.00, 'active', '別途接種オプション')
ON CONFLICT DO NOTHING;

-- ====================================================================================
-- 3. 薬剤マスタ（8件）
-- ====================================================================================
INSERT INTO master_items (code, name, category, price, status, description)
VALUES
  ('MED-001', 'アモキシシリン注射液', 'medicine', 1200.00, 'active', 'ペニシリン系抗菌薬'),
  ('MED-002', 'メトロニダゾール散', 'medicine', 1500.00, 'active', '原虫・嫌気性菌用'),
  ('MED-003', 'プレドニゾロン錠', 'medicine', 800.00, 'active', 'ステロイド系抗炎症薬'),
  ('MED-004', 'フロントラインプラス（猫用）', 'medicine', 2500.00, 'active', 'ノミ・ダニ駆除'),
  ('MED-005', 'ネクスガード（小型犬用）', 'medicine', 3000.00, 'active', 'ノミ・ダニ経口駆虫薬'),
  ('MED-006', 'レボリューション（猫用）', 'medicine', 2200.00, 'active', 'ノミ・寄生虫駆除'),
  ('MED-007', 'パモキサン配合液', 'medicine', 1500.00, 'active', '混合寄生虫駆虫薬'),
  ('MED-008', 'ラニチジン錠', 'medicine', 900.00, 'active', 'H2ブロッカー（消化性潰瘍治療）')
ON CONFLICT DO NOTHING;

-- ====================================================================================
-- 4. 診察マスタ（4件）
-- ====================================================================================
INSERT INTO master_items (code, name, category, price, status, description)
VALUES
  ('CONS-001', '初診料', 'consultation', 2000.00, 'active', '初回診察料'),
  ('CONS-002', '再診料', 'consultation', 1500.00, 'active', '再来患者診察料'),
  ('CONS-003', '時間外診察', 'consultation', 5000.00, 'active', '夜間・休日対応'),
  ('CONS-004', '往診料', 'consultation', 8000.00, 'active', '出張診察（交通費別途）')
ON CONFLICT DO NOTHING;

-- ====================================================================================
-- 5. 診療内容マスタ（5件）
-- ====================================================================================
INSERT INTO master_items (code, name, category, price, status, description)
VALUES
  ('SRVTYPE-001', '一般診察', 'serviceType', 2000.00, 'active', '病気・けがの診察'),
  ('SRVTYPE-002', '予防接種', 'serviceType', 3500.00, 'active', 'ワクチン接種'),
  ('SRVTYPE-003', '手術', 'serviceType', 25000.00, 'active', '避妊・去勢・腫瘍摘出等'),
  ('SRVTYPE-004', 'トリミング', 'serviceType', 5000.00, 'active', 'シャンプー、カット'),
  ('SRVTYPE-005', '入院ホテル', 'serviceType', 3500.00, 'active', '1泊ケージ利用料')
ON CONFLICT DO NOTHING;

-- ====================================================================================
-- 6. 処置マスタ（8件）
-- ====================================================================================
INSERT INTO master_items (code, name, category, price, status, description)
VALUES
  ('PROC-001', '皮下注射', 'procedure', 1200.00, 'active', 'SC投与'),
  ('PROC-002', '静脈注射', 'procedure', 1500.00, 'active', 'IV投与'),
  ('PROC-003', '点滴（輸液）', 'procedure', 4000.00, 'active', 'IV液体療法（1時間単位）'),
  ('PROC-004', '傷口処置', 'procedure', 2000.00, 'active', '洗浄、消毒、包帯交換'),
  ('PROC-005', '採血', 'procedure', 1000.00, 'active', '血液採取'),
  ('PROC-006', '導尿', 'procedure', 3000.00, 'active', 'カテーテル留置'),
  ('PROC-007', '酸素吸入', 'procedure', 2000.00, 'active', '酸素療法（30分）'),
  ('PROC-008', '浣腸', 'procedure', 1500.00, 'active', '排便促進処置')
ON CONFLICT DO NOTHING;

-- ====================================================================================
-- 7. 入院マスタ（3件）
-- ====================================================================================
INSERT INTO master_items (code, name, category, price, status, description)
VALUES
  ('HOSP-001', '入院費（犬）', 'hospitalization', 5000.00, 'active', '1泊'),
  ('HOSP-002', '入院費（猫）', 'hospitalization', 4000.00, 'active', '1泊'),
  ('HOSP-003', 'ホテル費', 'hospitalization', 3500.00, 'active', '健康な動物の一時預かり（1泊）')
ON CONFLICT DO NOTHING;

-- ====================================================================================
-- 8. スタッフマスタ（4件）
-- ====================================================================================
INSERT INTO master_items (code, name, category, price, status, description)
VALUES
  ('STAFF-001', '山田獣医太郎', 'staff', NULL, 'active', '獣医師（主任）'),
  ('STAFF-002', '鈴木獣医花子', 'staff', NULL, 'active', '獣医師'),
  ('STAFF-003', '佐藤看護師次郎', 'staff', NULL, 'active', '動物看護師'),
  ('STAFF-004', '田中トリマー太郎', 'staff', NULL, 'active', 'トリマー')
ON CONFLICT DO NOTHING;

-- ====================================================================================
-- 9. 保険マスタ（3件）
-- ====================================================================================
INSERT INTO master_items (code, name, category, price, status, description)
VALUES
  ('INS-001', 'アニコム損保', 'insurance', NULL, 'active', 'どうぶつ健保ふぁみりぃ'),
  ('INS-002', 'アイペット損保', 'insurance', NULL, 'active', 'アイペット損保'),
  ('INS-003', 'PS保険', 'insurance', NULL, 'active', 'ペットメディカルサポート')
ON CONFLICT DO NOTHING;

-- ====================================================================================
-- 10. ケージマスタ（6件）
-- ====================================================================================
INSERT INTO master_items (code, name, category, price, status, description)
VALUES
  ('CAGE-001', 'ケージ犬用S', 'cage', NULL, 'active', '小型犬用（60×45×45cm）'),
  ('CAGE-002', 'ケージ犬用M', 'cage', NULL, 'active', '中型犬用（90×60×60cm）'),
  ('CAGE-003', 'ケージ犬用L', 'cage', NULL, 'active', '大型犬用（120×75×75cm）'),
  ('CAGE-004', 'ケージ猫用A-1', 'cage', NULL, 'active', '猫用（60×45×45cm）'),
  ('CAGE-005', 'ケージ猫用A-2', 'cage', NULL, 'active', '猫用（60×45×45cm）'),
  ('CAGE-006', 'ホテル個室', 'cage', NULL, 'active', 'ホテル用個室ケージ')
ON CONFLICT DO NOTHING;

-- ====================================================================================
-- 11. トリミングコース（5件）
-- ====================================================================================
INSERT INTO master_items (code, name, category, price, status, description)
VALUES
  ('TRIM-001', 'シャンプー（小型犬）', 'trimming_course', 3500.00, 'active', '体重5kg未満'),
  ('TRIM-002', 'シャンプー（中型犬）', 'trimming_course', 4500.00, 'active', '体重5～15kg'),
  ('TRIM-003', 'カット（小型犬）', 'trimming_course', 5500.00, 'active', '体重5kg未満'),
  ('TRIM-004', 'カット（中型犬）', 'trimming_course', 7000.00, 'active', '体重5～15kg'),
  ('TRIM-005', 'シャンプー（猫）', 'trimming_course', 4000.00, 'active', '猫用シャンプーコース')
ON CONFLICT DO NOTHING;

-- ====================================================================================
-- 12. トリミングオプション（5件）
-- ====================================================================================
INSERT INTO master_items (code, name, category, price, status, description)
VALUES
  ('TRIMOPT-001', '歯磨き', 'trimming_option', 1000.00, 'active', '歯石除去・クリーニング'),
  ('TRIMOPT-002', '耳掃除', 'trimming_option', 800.00, 'active', '耳道清掃'),
  ('TRIMOPT-003', '爪切り', 'trimming_option', 500.00, 'active', '爪カット'),
  ('TRIMOPT-004', '肛門腺絞り', 'trimming_option', 800.00, 'active', '肛門腺液排出'),
  ('TRIMOPT-005', 'ブルーベリーパック', 'trimming_option', 2000.00, 'active', '目周り美容パック')
ON CONFLICT DO NOTHING;

-- ====================================================================================
-- 13. 診断カテゴリ（10件）
-- ====================================================================================
INSERT INTO master_items (code, name, category, price, status, description)
VALUES
  ('DIAGCAT-001', '皮膚科', 'diagnosis_category', NULL, 'active', '皮膚疾患'),
  ('DIAGCAT-002', '消化器', 'diagnosis_category', NULL, 'active', '消化器疾患'),
  ('DIAGCAT-003', '呼吸器', 'diagnosis_category', NULL, 'active', '呼吸器疾患'),
  ('DIAGCAT-004', '循環器', 'diagnosis_category', NULL, 'active', '心臓・血管疾患'),
  ('DIAGCAT-005', '整形外科', 'diagnosis_category', NULL, 'active', '骨・関節・筋肉疾患'),
  ('DIAGCAT-006', '神経', 'diagnosis_category', NULL, 'active', '神経疾患'),
  ('DIAGCAT-007', '眼科', 'diagnosis_category', NULL, 'active', '眼疾患'),
  ('DIAGCAT-008', '腫瘍', 'diagnosis_category', NULL, 'active', '悪性腫瘍・良性腫瘍'),
  ('DIAGCAT-009', '感染症', 'diagnosis_category', NULL, 'active', '感染性疾患'),
  ('DIAGCAT-010', '内分泌', 'diagnosis_category', NULL, 'active', 'ホルモン疾患')
ON CONFLICT DO NOTHING;

-- ====================================================================================
-- 14. 診断名（15件）
-- ====================================================================================
INSERT INTO master_items (code, name, category, price, status, description)
VALUES
  ('DIAG-001', 'アレルギー性皮膚炎', 'diagnosis_name', NULL, 'active', '皮膚科'),
  ('DIAG-002', '外耳炎', 'diagnosis_name', NULL, 'active', '皮膚科'),
  ('DIAG-003', '急性胃腸炎', 'diagnosis_name', NULL, 'active', '消化器'),
  ('DIAG-004', '膵炎', 'diagnosis_name', NULL, 'active', '消化器'),
  ('DIAG-005', '気管支炎', 'diagnosis_name', NULL, 'active', '呼吸器'),
  ('DIAG-006', '肺炎', 'diagnosis_name', NULL, 'active', '呼吸器'),
  ('DIAG-007', '僧帽弁閉鎖不全', 'diagnosis_name', NULL, 'active', '循環器'),
  ('DIAG-008', '膝蓋骨脱臼', 'diagnosis_name', NULL, 'active', '整形外科'),
  ('DIAG-009', 'てんかん', 'diagnosis_name', NULL, 'active', '神経'),
  ('DIAG-010', 'ぶどう膜炎', 'diagnosis_name', NULL, 'active', '眼科'),
  ('DIAG-011', '乳腺腫瘍', 'diagnosis_name', NULL, 'active', '腫瘍'),
  ('DIAG-012', '腺癌', 'diagnosis_name', NULL, 'active', '腫瘍'),
  ('DIAG-013', '猫ウイルス性鼻気管炎', 'diagnosis_name', NULL, 'active', '感染症'),
  ('DIAG-014', '甲状腺機能低下症', 'diagnosis_name', NULL, 'active', '内分泌'),
  ('DIAG-015', '糖尿病', 'diagnosis_name', NULL, 'active', '内分泌')
ON CONFLICT DO NOTHING;

-- ====================================================================================
-- 投入検証用クエリ（コメント）
-- ====================================================================================
-- SELECT category, COUNT(*) FROM master_items GROUP BY category ORDER BY category;
-- 期待値: 14カテゴリ・92件
