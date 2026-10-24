# 全555票の後方互換前提監査と文書是正（2026-10-24）

<a id="audit-scope"></a>
## 受領範囲と現在の適用方針

親側の最終監査結果として、555/555票の本文と重複除外後3,428コメントの読了報告を受領した。本docs作業による555票の再取得・再監査ではない。分類は後方互換義務の有無であり、製品欠陥数・全FR/ACの完了判定ではない。添付全文、活動履歴、全外部参照、全source、実行受入の全面監査は含まない。

本人方針「既存のコードに互換性を持たせる必要はない。確定仕様に準じて作り直す」が適用される。以下22票の是正方針はroot採用済み。旧API/DTO/wire/alias/profile保持だけを目的とする中間adapter・二重routeを必須にせず、必要caller/DI/reader/writerを同じ完結単位で置換する。[チケット規約](../../ops/plane-ticket-conventions.md#specification-first-replacement)に従う。選定APIの新しいshape・HTTPをこの記録から推測しない。

安全・現在認可・CAS・C06・原子性・履歴保全と正式移行HOLDは維持する。528/535のCreate既定意味、および対象22票に含まれない502のnumber精度/JCS契約は変更しない。Plane追補は別担当であり、本docs作業では外部投稿・状態/担当/本文/relation変更を行っていない。

基準mainは `ae41e6035728a86dbb708266ad35f9df06bd0aba`（PR #850採用後）。本稿と関連追補は文書上の適用範囲の是正であり、製品置換・API削除・利用開始済みではない。

受領資料（ともにLibrary version 0、ローカル実在・全文可読を確認）:

| 資料 | Library ID | SHA-256 |
|---|---|---|
| audit-summary-final.md | `libfile_ef9d708c57788191ad2854a33a478e6f` | `5f6199bdb7064f37c97d23a6a71ee4792d29b50a400b0245b93c16baff9d823a` |
| correction-proposals-final.jsonl | `libfile_60e6799e7ce0819193f2fac2e1ba7017` | `5c125a3c99d23485a5cd61fa65970fbe0812c3d42d42f1fbb0db1c2353548ba1` |

## 分類別票番号

### clear (433)
EMR-1, EMR-2, EMR-3, EMR-4, EMR-5, EMR-6, EMR-7, EMR-8, EMR-9, EMR-10, EMR-11, EMR-12, EMR-13, EMR-15, EMR-16, EMR-17, EMR-18, EMR-19, EMR-20, EMR-21, EMR-22, EMR-23, EMR-24, EMR-25, EMR-26, EMR-27, EMR-28, EMR-29, EMR-30, EMR-31, EMR-32, EMR-33, EMR-34, EMR-35, EMR-36, EMR-37, EMR-39, EMR-40, EMR-41, EMR-42, EMR-43, EMR-44, EMR-45, EMR-46, EMR-47, EMR-49, EMR-50, EMR-51, EMR-52, EMR-53, EMR-54, EMR-55, EMR-56, EMR-57, EMR-58, EMR-59, EMR-60, EMR-61, EMR-63, EMR-64, EMR-65, EMR-67, EMR-69, EMR-70, EMR-71, EMR-72, EMR-74, EMR-75, EMR-76, EMR-77, EMR-78, EMR-79, EMR-80, EMR-81, EMR-82, EMR-83, EMR-84, EMR-86, EMR-87, EMR-88, EMR-89, EMR-90, EMR-91, EMR-92, EMR-93, EMR-94, EMR-95, EMR-96, EMR-98, EMR-99, EMR-100, EMR-101, EMR-102, EMR-103, EMR-104, EMR-105, EMR-106, EMR-109, EMR-110, EMR-111, EMR-112, EMR-114, EMR-116, EMR-117, EMR-120, EMR-121, EMR-122, EMR-123, EMR-124, EMR-125, EMR-126, EMR-127, EMR-128, EMR-130, EMR-131, EMR-132, EMR-133, EMR-139, EMR-140, EMR-141, EMR-144, EMR-145, EMR-149, EMR-150, EMR-152, EMR-153, EMR-154, EMR-155, EMR-157, EMR-159, EMR-160, EMR-162, EMR-163, EMR-164, EMR-165, EMR-167, EMR-169, EMR-170, EMR-171, EMR-172, EMR-173, EMR-175, EMR-176, EMR-177, EMR-178, EMR-180, EMR-181, EMR-182, EMR-184, EMR-186, EMR-187, EMR-188, EMR-189, EMR-190, EMR-191, EMR-192, EMR-193, EMR-194, EMR-195, EMR-196, EMR-197, EMR-198, EMR-199, EMR-200, EMR-201, EMR-202, EMR-203, EMR-204, EMR-205, EMR-206, EMR-207, EMR-208, EMR-210, EMR-211, EMR-212, EMR-213, EMR-214, EMR-215, EMR-216, EMR-217, EMR-218, EMR-219, EMR-220, EMR-221, EMR-222, EMR-223, EMR-224, EMR-226, EMR-227, EMR-228, EMR-229, EMR-230, EMR-231, EMR-232, EMR-233, EMR-234, EMR-235, EMR-236, EMR-237, EMR-238, EMR-241, EMR-242, EMR-243, EMR-244, EMR-245, EMR-246, EMR-247, EMR-248, EMR-249, EMR-252, EMR-253, EMR-254, EMR-255, EMR-256, EMR-257, EMR-259, EMR-260, EMR-261, EMR-262, EMR-263, EMR-264, EMR-265, EMR-266, EMR-267, EMR-269, EMR-270, EMR-271, EMR-273, EMR-275, EMR-276, EMR-278, EMR-279, EMR-281, EMR-282, EMR-283, EMR-284, EMR-286, EMR-287, EMR-288, EMR-289, EMR-290, EMR-291, EMR-292, EMR-293, EMR-294, EMR-297, EMR-298, EMR-299, EMR-300, EMR-301, EMR-302, EMR-303, EMR-304, EMR-305, EMR-306, EMR-307, EMR-309, EMR-310, EMR-311, EMR-312, EMR-313, EMR-314, EMR-315, EMR-316, EMR-318, EMR-319, EMR-322, EMR-324, EMR-325, EMR-326, EMR-328, EMR-329, EMR-330, EMR-332, EMR-333, EMR-334, EMR-335, EMR-336, EMR-337, EMR-338, EMR-339, EMR-341, EMR-344, EMR-346, EMR-348, EMR-349, EMR-350, EMR-351, EMR-352, EMR-353, EMR-354, EMR-355, EMR-356, EMR-357, EMR-358, EMR-359, EMR-360, EMR-361, EMR-362, EMR-363, EMR-365, EMR-366, EMR-367, EMR-368, EMR-369, EMR-372, EMR-377, EMR-378, EMR-379, EMR-380, EMR-381, EMR-382, EMR-383, EMR-384, EMR-386, EMR-387, EMR-388, EMR-389, EMR-390, EMR-391, EMR-392, EMR-393, EMR-394, EMR-396, EMR-397, EMR-398, EMR-399, EMR-400, EMR-402, EMR-403, EMR-404, EMR-407, EMR-408, EMR-409, EMR-411, EMR-415, EMR-416, EMR-417, EMR-418, EMR-419, EMR-420, EMR-421, EMR-422, EMR-423, EMR-424, EMR-425, EMR-426, EMR-427, EMR-428, EMR-431, EMR-432, EMR-433, EMR-435, EMR-436, EMR-437, EMR-438, EMR-439, EMR-440, EMR-441, EMR-442, EMR-445, EMR-464, EMR-465, EMR-467, EMR-468, EMR-470, EMR-471, EMR-472, EMR-473, EMR-474, EMR-475, EMR-476, EMR-477, EMR-478, EMR-479, EMR-480, EMR-481, EMR-482, EMR-484, EMR-485, EMR-486, EMR-487, EMR-488, EMR-489, EMR-490, EMR-491, EMR-493, EMR-494, EMR-495, EMR-497, EMR-498, EMR-499, EMR-500, EMR-501, EMR-502, EMR-503, EMR-504, EMR-507, EMR-508, EMR-509, EMR-510, EMR-513, EMR-514, EMR-515, EMR-516, EMR-517, EMR-518, EMR-521, EMR-522, EMR-523, EMR-524, EMR-525, EMR-526, EMR-529, EMR-530, EMR-531, EMR-532, EMR-533, EMR-534, EMR-537, EMR-538, EMR-539, EMR-540, EMR-541, EMR-542, EMR-544, EMR-545, EMR-546, EMR-547, EMR-548, EMR-553, EMR-554, EMR-555

### historical-only (95)
EMR-14, EMR-38, EMR-48, EMR-62, EMR-66, EMR-68, EMR-73, EMR-85, EMR-97, EMR-107, EMR-108, EMR-113, EMR-115, EMR-118, EMR-119, EMR-129, EMR-134, EMR-135, EMR-136, EMR-137, EMR-138, EMR-142, EMR-143, EMR-146, EMR-147, EMR-148, EMR-151, EMR-156, EMR-158, EMR-161, EMR-166, EMR-168, EMR-174, EMR-179, EMR-183, EMR-185, EMR-209, EMR-225, EMR-239, EMR-250, EMR-258, EMR-272, EMR-274, EMR-277, EMR-280, EMR-285, EMR-296, EMR-308, EMR-317, EMR-320, EMR-321, EMR-323, EMR-327, EMR-331, EMR-340, EMR-345, EMR-347, EMR-370, EMR-371, EMR-373, EMR-374, EMR-375, EMR-376, EMR-385, EMR-395, EMR-401, EMR-405, EMR-406, EMR-412, EMR-429, EMR-430, EMR-434, EMR-443, EMR-444, EMR-446, EMR-447, EMR-448, EMR-449, EMR-450, EMR-451, EMR-452, EMR-453, EMR-454, EMR-455, EMR-456, EMR-457, EMR-458, EMR-459, EMR-460, EMR-461, EMR-462, EMR-469, EMR-483, EMR-512, EMR-543

### compatibility-obligation (22)
EMR-240, EMR-251, EMR-268, EMR-295, EMR-342, EMR-343, EMR-364, EMR-410, EMR-463, EMR-466, EMR-492, EMR-496, EMR-505, EMR-506, EMR-511, EMR-519, EMR-520, EMR-527, EMR-528, EMR-535, EMR-536, EMR-549

### needs-context (5)
EMR-413, EMR-414, EMR-550, EMR-551, EMR-552

## 状態別件数

- Backlog: 12
- Done: 238
- Ready: 69
- Cancelled: 15
- Duplicate: 37
- Blocked: 8
- UAT: 72
- Phase2: 1
- Infra: 10
- Review: 81
- Needs Human: 11
- Research: 1

<a id="accepted-corrections"></a>
## 22票の受領根拠と採用是正方針

<a id="correction-240"></a>
### EMR-240 FE 未実装 API: LSTEP オプトアウト（lstep-opt-out 重複 endpoint）の整理判断
https://app.plane.so/baritechllc/browse/EMR-240/

根拠: 10/2整理判断が外部利用確認まで旧POST継続稼働を必須化し、10/4受入も旧POST204を要求。現在inventoryにも同趣旨が残る。旧コード不要という本人判断に沿い、将来計画での互換維持条件を外す追補が必要。

採用是正方針: 【再構築時の互換方針】旧POST lstep-opt-outの互換稼働・204応答は将来の必須受入条件にしない。確定仕様で必要な外部／運用consumerを棚卸しし、同じ完結単位で支持契約へ置換／廃止する。存在不明のconsumerを理由に旧APIを恒久維持しない。製品UIは両停止フラグの整合を担うdelivery-exclusion契約に従い、lstep専用PATCHへ機械的に置換しない。EMR-38の送信gateとis_sync_enabledはOFFのまま、opt-out記録と独立停止理由を保全し、同期無効時の外部タグ呼出ゼロを維持する。これは票の計画是正であり今回のAPI削除・送信操作指示ではない。

<a id="correction-251"></a>
### EMR-251 飼主画面にLSTEP配信の停止と再開を足す
https://app.plane.so/baritechllc/browse/EMR-251/

根拠: 最新着手条件が明示的に外部利用不明な互換POSTを残すと規定。UI修正の必須条件にAPI削除を入れないというscope管理自体は問題ないが、旧POST維持の必須化は分離して訂正する。

採用是正方針: 【スコープと互換義務の分離】本UI票単独の完了条件に、対象外API削除を追加しない。ただし旧互換POSTの維持も要件としない。旧APIはEMR-240のconsumer棚卸しと同じ完結単位で置換／廃止する計画へ接続する。UIは実値取得とdelivery_excluded/lstep_opt_outの整合、理由上限100、医院・権限分離を満たす。送信gate OFF・同期無効時の外部呼出ゼロ、転院／全ペット死亡等の独立停止を維持し、停止理由が残るとき配信可能と誤表示しない。

<a id="correction-268"></a>
### EMR-268 検査連携再設計：機器型番入力→実機疎通→連携仕様確定フロー
https://app.plane.so/baritechllc/browse/EMR-268/

根拠: 最新caller計画でRawMessageモデルに対し『RawMessage→既存string DTO』『公開APIを勝手に変更せず』を要求する。確定仕様に沿う型再設計より旧公開string形状保持を前提にしたadapter義務が残る。配列検証や生データ非保存とは別。

採用是正方針: 既存string DTO保持を旧コード互換のための必須条件にはしない。確定仕様がstringを要求するならその根拠で採用し、そうでなければ単一JSON契約へPut/Get/Boardと必要callerを揃える。RawMessageモデル、配列内容検証、null/空/不正の区別、医院分離/TTLは維持。

<a id="correction-295"></a>
### EMR-295 チャット履歴一覧を会話ルーム形式UIに変更（時間ギャップ疑似セッション化）
https://app.plane.so/baritechllc/browse/EMR-295/

根拠: 本文のrooms移行に加え、10月5日コメントd0f169cb-b0d8-44f2-b7cd-567c8d7a646cで旧exchanges消費hook/型/query-keyは削除するがBE /chat/exchanges自体は互換のため残置と明示。後続PR587完了・UATに撤回なし。

採用是正方針: 旧client互換を理由に/chat/exchangesを残す義務は撤回し、確定したrooms APIに従う。別の確定仕様が必要とするendpointまで削除必須とはしない。履歴データ・tenant認可・room境界は保全する。

<a id="correction-342"></a>
### EMR-342 【基盤】取込契約に薬剤マスタの表を足す設計(旧分類から生成)
https://app.plane.so/baritechllc/browse/EMR-342/

根拠: 参照正本DESIGN先頭の10/13改訂でlegacy21 formalの経路/version/pin保持、10/14改訂で表別commit/既存prefix resumeは変更しないと必須化。現在仕様へproducer/consumerを置換可能とする本人方針に反する維持義務部分が明示された。

採用是正方針: 本人確定方針により旧コード・旧API/DTO・旧producer/consumer・legacy profileとの後方互換は完成条件にしない。確定仕様に合わせ、必要なcaller/CLI/reader/writer/contractを同じ完結単位で置換できる。旧21表/22候補の既定選択・旧pin・表別commit/prefix resume・非退行回帰の恒久維持を要求しない。旧成果物・受領履歴は証拠として保全するが、現在の19 CSVに適合済みとは扱わない。payments/payment_splitsや退役列を復活させず、旧会計根拠は版付きevidence-onlyの採択境界で別保全する。新profileのsource全ファイルhash、同一artifact/run/医院/target、独立検証・producer/consumer receipt、fresh/band保護・原子性・未知commit時の読取照合・実操作承認は維持する。旧経路置換の許可は正式activation/HOLD解除/未検証bundle取込許可ではない。

<a id="correction-343"></a>
### EMR-343 取込契約に薬剤マスタ(medicines)の表を足す（取込側の実装。EMR-342 の設計）
https://app.plane.so/baritechllc/browse/EMR-343/

根拠: 本文・最新コメントのlegacy非退行に加え、参照DESIGNとrunbookはlegacy21既定/pin/表別commit/prefix resume維持を明示。単なる未検証artifactのHOLDを越えた旧実装維持義務を確認。

採用是正方針: 本人確定方針により旧コード・旧API/DTO・旧producer/consumer・legacy profileとの後方互換は完成条件にしない。確定仕様に合わせ、必要なcaller/CLI/reader/writer/contractを同じ完結単位で置換できる。旧21表/22候補の既定選択・旧pin・表別commit/prefix resume・非退行回帰の恒久維持を要求しない。旧成果物・受領履歴は証拠として保全するが、現在の19 CSVに適合済みとは扱わない。payments/payment_splitsや退役列を復活させず、旧会計根拠は版付きevidence-onlyの採択境界で別保全する。新profileのsource全ファイルhash、同一artifact/run/医院/target、独立検証・producer/consumer receipt、fresh/band保護・原子性・未知commit時の読取照合・実操作承認は維持する。旧経路置換の許可は正式activation/HOLD解除/未検証bundle取込許可ではない。

<a id="correction-364"></a>
### EMR-364 契約提案｜billing_itemsの履歴tax_rateをCSV取込へ追加する（MIG-44境界・固定pin/互換性合意）
https://app.plane.so/baritechllc/browse/EMR-364/

根拠: 選定正本§3.1は旧profile意味・既定・pinを変更しない、§3.5(5)はlegacy表別commit/complete-prefix resumeを変更しない、§7は旧profile回帰、§8はlegacy prefix互換を要求。新tax契約の安全条件とは分離可能な旧実装維持義務。

採用是正方針: 本人確定方針により旧コード・旧API/DTO・旧producer/consumer・legacy profileとの後方互換は完成条件にしない。確定仕様に合わせ、必要なcaller/CLI/reader/writer/contractを同じ完結単位で置換できる。旧21表/22候補の既定選択・旧pin・表別commit/prefix resume・非退行回帰の恒久維持を要求しない。旧成果物・受領履歴は証拠として保全するが、現在の19 CSVに適合済みとは扱わない。payments/payment_splitsや退役列を復活させず、旧会計根拠は版付きevidence-onlyの採択境界で別保全する。新profileのsource全ファイルhash、同一artifact/run/医院/target、独立検証・producer/consumer receipt、fresh/band保護・原子性・未知commit時の読取照合・実操作承認は維持する。旧経路置換の許可は正式activation/HOLD解除/未検証bundle取込許可ではない。

<a id="correction-410"></a>
### EMR-410 [リサーチ] サポートチャットのSarashina移行とハイブリッド検索の要件を整理する
https://app.plane.so/baritechllc/browse/EMR-410/

根拠: 本文R5はhybrid=false移行期間に旧context requestを受理しFE更新後停止する二重request対応を規定。期限付きでも旧コード/旧client受付の明示技術条件で、新しい共通規約の後方互換非要件と整理が必要。対してSources nullable追従は採用型のconsumer接続とNULL意味保全、検索品質/権限/外部送信gateは業務・安全契約。

採用是正方針: 旧context受理を必須条件から外し、確定ID/hash APIと必要FE callerを同じ成功単位で置換。必要な一時切替手段なら期間/owner/停止条件を限定して選択肢化。履歴NULL/[]意味・検索品質・認可・送信gateは維持。

<a id="correction-463"></a>
### EMR-463 【実装整備】飼主一覧の認可内医院初期値と選択保持を接続する
https://app.plane.so/baritechllc/browse/EMR-463/

根拠: 旧clientへの追加field互換を維持する

採用是正方針: /me専用DTO選定と現在grantを維持し、旧client/builder signature維持を必須条件から外す

<a id="correction-466"></a>
### EMR-466 【実装整備・DB統合】DB全体の単一PRを統合しchecksum・ERD・互換性を再検証する
https://app.plane.so/baritechllc/browse/EMR-466/

根拠: 現行reader互換／consumer互換を統合再検証

採用是正方針: 新仕様consumer適合と履歴保全に置換。現行reader形状維持を条件にしない

<a id="correction-492"></a>
### EMR-492 【実装整備・追加consumer】4業務種別・標準時間・名称固定版と個別再開を予約へ接続する
https://app.plane.so/baritechllc/browse/EMR-492/

根拠: コメント c94f3d49-d866-4e17-a049-236d0e21c0ae（PR805追補）の「旧clear_group_idは互換adapter内で一度だけpresence付きnullへ変換」が旧flag受付を必須化。後続コメントに撤回なし。

採用是正方針: 旧flag互換adapter必須を撤回し、確定group_idの省略=保持/null=解除/値=設定、presence・CAS・現在scope等は維持する。

<a id="correction-496"></a>
### EMR-496 【実装整備・追加consumer】構造化初診問診の複数回答・時期精度・出所付き改訂を接続する
https://app.plane.so/baritechllc/browse/EMR-496/

根拠: 固定ae41e603のemr-496-497-selected-wire-contract-20261021.md §3: category deprecated alias受理、両field互換期間、利用ログ/consumer棚卸しなしalias除去禁止。blob198463645b9dbd14b8f7edfbba04ebef99fa7b98。

採用是正方針: 旧alias受付/二重応答の維持義務のみ撤回。既採択canonical typeとpresence、厳密整数安全、snapshot/履歴・CAS/hash保持。数値をstringへ新規変更する決定ではない。

<a id="correction-505"></a>
### EMR-505 【実装整備・追加consumer】算定固定版・active claim・理由付き取消再請求を接続する
https://app.plane.so/baritechllc/browse/EMR-505/

根拠: 旧数値応答が必要な既存routeは明示互換adapterで分けてsilent変更しない。

採用是正方針: 旧route/数値応答を維持するためだけのadapter義務を撤回。選定exact decimal契約へAPI/FE/callerを同時移行し精度/NULL/履歴を保持。

<a id="correction-506"></a>
### EMR-506 【実装整備・追加consumer】受領identity・複数請求配賦・実返金・記録訂正を接続する
https://app.plane.so/baritechllc/browse/EMR-506/

根拠: コメント 0def0d64-b336-4267-b2cb-d4a403c8c5a4 の「初回201/再送200・Billing wireは維持」「移行が必要ならAPI ownerがadapterを明記」「限定認可/互換維持は選定済み」。

採用是正方針: 旧Complete応答/wireとの互換維持とadapter必須だけを撤回。確定API契約に統一してcaller同期。現在認可、元target anchor、生前精算条件、金銭/監査/原子receiptは維持する。

<a id="correction-511"></a>
### EMR-511 【実装整備・追加consumer】確定受付/過去来院根拠と訂正を集計readerへ接続する
https://app.plane.so/baritechllc/browse/EMR-511/

根拠: 未指定/legacy_record_days_v1は旧意味・型・parser・既定値を変えない。既存互換alias、legacy日付計算不変。

採用是正方針: 旧client互換だけのopt-in二重mode/alias/parser維持義務を撤回。旧指標に独立した確定業務要件がある場合だけ別metricとして残し、actual unknown非補完/履歴/認可/coverageを維持。

<a id="correction-519"></a>
### EMR-519 LIFF予約IDを正規10進文字列で返し全予約導線を移行する
https://app.plane.so/baritechllc/browse/EMR-519/

根拠: 最新PR776 merge記録も第1段階のみとし、旧number-only client解消確認後のproducer文字列化を残す。現行rollout契約でsafe legacy number受理、client互換先行配備、旧producer不在後に互換除去を義務化しており歴史記録だけではない。

採用是正方針: 旧producer/旧number-only clientとの互換を着手・完了条件から外し、確定した正規10進string契約へBEと全必要consumerを揃える。number受理adapterとclient-first段階移行は必須としない。unsafe ID拒否、作成結果不明時の盲目再送禁止、現在認可、問診activation gateは維持する。

<a id="correction-520"></a>
### EMR-520 LIFF予約の実ペットidentityと認可済み表示名を対で返す
https://app.plane.so/baritechllc/browse/EMR-520/

根拠: 最新merge記録でも旧pet_name-only応答を両nullへ正規化する移行実装を採用。PR777現行契約は旧producer number受理・client先行切替・限定legacy shape受理を明示する。実pet認可・名前identity一致とは分離すべき互換要件。

採用是正方針: 旧pet_name-only/両省略shapeの救済と予約number互換を義務化せず、確定したpaired pet_id/pet_name契約とstring IDへ必要consumerを揃える。認可済み同一snapshot、未特定/失権の両null、名前からidentityを推測しない安全要件は維持する。

<a id="correction-527"></a>
### EMR-527 ① 共通model生成型同期と既存caller型互換の整合
https://app.plane.so/baritechllc/browse/EMR-527/

根拠: 題名だけでなく具体consumer表で非永続legacy-wire DTOを新設しnumber/string/null/省略・旧計算・保存意味をそのまま維持する計画。新APIを優先し旧保存型復活は禁ずるが、①完了のため会計/見積/廃止gate/pet旧fieldへの互換adapterを必須にしている。 全13コメントの最新選定00b1d844は808→812の順で現main HTTP互換DTOを先行し後続でexactへ置換する。新wireを旧numberへ戻す/union併存を禁止しており、恒久維持ではなく中間互換工程の必須化だけが論点。

採用是正方針: 新仕様と必要model/callerを直接整合できる場合は旧HTTP互換DTOの先行実装を必須にしない。既採用808を元に新wireへ置換する成果を後退させず、旧保存fieldや旧numberを復活させない。生成物の正規生成・単一DTO所有・null/省略/0の確定意味は維持する。

<a id="correction-528"></a>
### EMR-528 ② 入院プランマスタの任意項目クリアと明示0価格を保存する
https://app.plane.so/baritechllc/browse/EMR-528/

根拠: 本文段階はclearだったが全11コメントのうち10/22 03:28回答6dcdcb7a-2b82-4151-bc6c-c38a5a4eb2da第2項がPOST body_size/billing_unit空文字を「既存互換どおり省略扱いを維持」と明示。後続撤回なし。ただし同回答はCreate意味の技術選定とも明記し、per_day既定・PATCH三値とは別の互換理由の是正として扱う。

採用是正方針: POST空文字許容を旧コード互換だけで必須にしない。確定仕様として空文字＝省略を残す場合は互換理由ではなくその入力契約を正本と明示し、新consumerもその契約へ統一する。互換不要から空文字400へ無断変更しない。POST省略per_day、PATCH省略保持/null clear/空文字拒否、明示0は独立した確定仕様として維持する。

<a id="correction-535"></a>
### EMR-535 入院プランの体格・料金単位を未設定へ戻すUIを接続する
https://app.plane.so/baritechllc/browse/EMR-535/

根拠: 採用根拠2-Aに「POST空文字は既存互換の省略扱い」、AC4に「POST空文字互換…変更しない」と明示。旧wire受理を維持義務としている。 全2コメント。最新も本文AC維持しPOST省略/PATCH nullを区別。本文の既存互換という実装理由は撤回なし。Create既定意味まで撤回せず、旧wire保持理由だけを除く。

採用是正方針: POST空文字受理を旧互換だけで維持必須としない。確定仕様上のPOST未指定/空/null/enum意味を新契約として揃え、PATCH三値・未設定UIとは分けて更新する。

<a id="correction-536"></a>
### EMR-536 ② 薬剤税率のexact往復とPATCH含量clear契約を補完する
https://app.plane.so/baritechllc/browse/EMR-536/

根拠: 保存exact文字列を保持する方針と同時に「既存calculator callerが必要とするnumber互換性は局所adapterで保持」と明示。旧number consumerへの互換adapter保持が義務化。 全9コメント。最新mergeはexact保存字句と編集intent分離、未操作PATCH省略。0% UIなしは正本判断。本文の既存calculator用number互換adapter必須は撤回確認なし。

採用是正方針: 旧calculator number互換adapterを必須にせず、確定exact契約へ当該consumerを同時移行する。表示用変換は保存・計算正本へ逆流させず、既存値の精度保全と未決の新税則禁止を維持する。

<a id="correction-549"></a>
### EMR-549 【基盤】main全ドメインbackendビルド破損 — 共通identityストレージ未接続の仕様判断と修復
https://app.plane.so/baritechllc/browse/EMR-549/

根拠: 共通identity新storage consumer未接続を修復する票。死亡状態源・lock・検索応答の契約確定は必要だが旧field復元要求はない。 全29コメント全文読了。8823f814の互換不要決裁と0afd842aのshift/status同時切替是正が旧要求を撤回。ただし後続cb91b065は旧資格情報キー非空のみ400へ復元し、存在400ならFE同時削除の再指示を要求。旧client受理の残存義務を限定して是正候補とする。

採用是正方針: cb91b065の旧資格情報キーを非空のみ400へ戻す条件を撤回。確定契約に従いBEとFE送信callerを同じ単位で置換可能と明記。liff_idの機能上の必要性と認証保存先の適合は互換と別に維持。旧route廃止/固定エラーの判断は安全な未提供境界として扱い、互換目的では再承認を要さない。

<a id="pending-source-reports"></a>
## 原報告不足5票の判定保留
- EMR-413: ['staging報告17の匿名化した症状、期待/実際、画面/操作、日時/版、頻度。原文がないため互換義務の有無は判断保留。']
- EMR-414: ['staging報告18の匿名化した症状、期待/実際、画面/操作、日時/版、頻度。原文がないため互換義務の有無は判断保留。']
- EMR-550: ['stagingアプリ内バグ報告#51の詳細本文・添付・対応記録。票本文・全0コメントのみ確認。']
- EMR-551: ['stagingアプリ内バグ報告#52の詳細本文・添付・対応記録。票本文・全0コメントのみ確認。']
- EMR-552: ['stagingアプリ内バグ報告#53の詳細本文・添付・対応記録。票本文・全0コメントのみ確認。']

<a id="verification-boundary"></a>
## 検証と限界

555票の分類別一覧・状態件数は受領報告の監査時点の記録。clear433・historical-only95・compatibility-obligation22・needs-context5で計555票。原報告不足5票を解消済みにしない。331の古い再取得待ちは、受領監査で最新PR #850コメントとの照合により解消済みと報告されている。

原監査の内部coverage/JSON群は今回の入力2資料に含まれず、公開文書の参照先にしない。根拠は本稿の票URL・具体コメント/節と下記repo参照に限定する。本作業は関連文書の静的整合確認のみ。tests/build/codegen/DB/Docker/browserはNOT_RUN。未実装・未接続・契約未選定と実行未実施を同一視せず、runtime PASS・全票是正済みとはしない。

<a id="repository-corrections"></a>
## repo文書の対応と残る境界

以下の適用追補は指定22票の旧互換義務に限る。旧記録は基準mainのblobに残し、文書内の当時の観測・履歴は消さない。追補が優先するのは互換維持を必須にする部分だけで、無関係な採択・FR/ACを再選定しない。

| 対象 | repoの参照・対応 |
|---|---|
| 240 / 251 | [FORM-FIELD-INVENTORY.md](../../ops/testing/scenarios/FORM-FIELD-INVENTORY.md#compatibility-scope-20261024)の限定追補。 |
| 342 / 343 / 364 | [DESIGN.md](emr-342-medicines-contract-design/DESIGN.md#compatibility-scope-20261024)の限定追補。 |
| 342 / 343 / 364 | [CLINIC_CSV_IMPORT.md](../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md#compatibility-scope-20261024)の限定追補。 |
| 364 | [emr-364-375-376-csv-column-contract-20261021.md](emr-364-375-376-csv-column-contract-20261021.md#compatibility-scope-20261024)の限定追補。 |
| 496 | [emr-496-497-selected-wire-contract-20261021.md](emr-496-497-selected-wire-contract-20261021.md#compatibility-scope-20261024)の限定追補。 |
| 511 | [emr-511-actual-visit-reader-contract-20261021.md](emr-511-actual-visit-reader-contract-20261021.md#compatibility-scope-20261024)の限定追補。 |
| 519 / 520 | [emr-520-reservation-pet-reader-contract-20261021.md](emr-520-reservation-pet-reader-contract-20261021.md#compatibility-scope-20261024)の限定追補。 |
| 527 | [emr-445-stage1-contract-start-plan-20261021.md](emr-445-stage1-contract-start-plan-20261021.md#compatibility-scope-20261024)の限定追補。 |
| 268 / 527 | [README.md](../model-alignment-20261020/README.md#compatibility-scope-20261024)の限定追補。 |
| 492 | [emr-490-492-selected-wire-contract-20261021.md](emr-490-492-selected-wire-contract-20261021.md#compatibility-scope-20261024)の限定追補。 |
| 506 | [emr-540-308-462-506-selected-handoff-20261022.md](emr-540-308-462-506-selected-handoff-20261022.md#compatibility-scope-20261024)の限定追補。 |
| 506 | [emr-506-fund-wire-design-candidates-20261022.md](emr-506-fund-wire-design-candidates-20261022.md#compatibility-scope-20261024)の限定追補。 |
| 505 / 506 | [technical-spec.md](../../design/domains/accounting/technical-spec.md#compatibility-scope-20261024)の限定追補。 |
| 506 | [owner-deposit-initial-receipt.md](../../design/domains/accounting/owner-deposit-initial-receipt.md#compatibility-scope-20261024)の限定追補。 |
| 463 | [technical-spec.md](../../design/domains/owners-pets/technical-spec.md#compatibility-scope-20261024)の限定追補。 |
| 466 | [unified-db-20261019.md](../db-contracts/unified-db-20261019.md#compatibility-scope-20261024)の限定追補。 |
| 528 / 535 | [audit-next-master-clears-20261023.md](audit-next-master-clears-20261023.md#compatibility-scope-20261024)の限定追補。 |
| 295 / 410 / 536 / 549 | 受領根拠は上記票本文/コメント。関連repoのチャット一覧・型影響・薬剤仕様も照合したが、今回の検索範囲では当該旧endpoint残置、旧context受付、calculator number adapter、資格情報key復元を必須にする独立のrepo本文は特定しなかった。新wire/routeを推測して追加せず、本稿に採用是正方針を残す。全repo/全外部参照の不存在証明ではない。 |

未閉鎖なのは、各ownerでの確定契約への必要consumer置換・未選定wireの閉鎖・供給と実行受入、原報告不足5票の再判定。本PRはこれらの実装・Plane追補の完了を宣言しない。
