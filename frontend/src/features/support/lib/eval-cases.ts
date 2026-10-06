/**
 * eval-cases.ts — サポートチャットのマニュアル検索評価フィクスチャ
 *
 * 宣言的なケース一覧。UNIT-02（JA 検索強化）/ UNIT-03（VN 対応）は
 * この配列にケースを追加するだけで評価対象を拡張できる（ハーネス変更不要）。
 *
 * ケースの形:
 * - expectedSlugs: 期待する記事。`screens/04-medical-records` のように
 *   カテゴリ修飾付き slug で指定する。top-3 のいずれかに出れば HIT。
 * - expectedEmpty: どの記事にもマッチしないことが期待されるコントロール。
 *
 * 採点結果（recall@3）はメトリクスとして出力されるだけで、
 * ベースラインのミス自体はテスト失敗にしない。
 */
type EvalLang = "ja" | "vi";

export interface EvalCase {
  /** 自然文の質問 */
  readonly question: string;
  /** 期待する記事キー（category/slug 形式）。expectedEmpty / expectedWeak と排他 */
  readonly expectedSlugs?: readonly string[];
  /** true の場合、コンテキストが空であることが期待値（pre-gate Fuse 候補 0 件の真のノーマッチ） */
  readonly expectedEmpty?: true;
  /**
   * true の場合、弱ヒット（pre-gate Fuse 候補は出るが根拠ゲートを通らない）で
   * text 空のタイトルリンク候補が 1〜3 件返ることが期待値
   */
  readonly expectedWeak?: true;
  readonly lang: EvalLang;
}

export const evalCases: readonly EvalCase[] = [
  // ── JA パラフレーズ（敬語・口語・漢字ゆれを分散。報告済みケースを含む）──
  {
    question: "カルテの作成方法は？", // 報告済みケース
    expectedSlugs: ["screens/04-medical-records"],
    lang: "ja",
  },
  {
    question: "カルテを新規に作りたいのですが",
    expectedSlugs: ["screens/04-medical-records"],
    lang: "ja",
  },
  {
    question: "会計処理のやり方を教えてください",
    expectedSlugs: ["screens/05-accounting"],
    lang: "ja",
  },
  {
    question: "レジの締め方が分かりません",
    expectedSlugs: ["screens/13-cash-register"],
    lang: "ja",
  },
  {
    question: "予約を登録するにはどうすればいいですか",
    expectedSlugs: ["screens/06-reservations"],
    lang: "ja",
  },
  {
    question: "新しい飼い主さんを登録したい",
    expectedSlugs: ["screens/03-owners"],
    lang: "ja",
  },
  {
    question: "ワクチン接種の記録はどこで入力しますか",
    expectedSlugs: ["screens/08-vaccinations"],
    lang: "ja",
  },
  {
    question: "入院の手続き方法を教えて",
    expectedSlugs: ["screens/09-hospitalization"],
    lang: "ja",
  },
  {
    question: "在庫の数を確認したいです",
    expectedSlugs: ["screens/12-inventory"],
    lang: "ja",
  },
  {
    question: "日次締めの流れを教えてください",
    expectedSlugs: ["workflows/04-daily-closing"],
    lang: "ja",
  },

  // ── VN（主要オペレーション。UNIT-03 で keywords_vi + VN クエリ処理を導入しアサート対象）──
  {
    question: "Cách tạo hồ sơ bệnh án?", // カルテ作成
    expectedSlugs: ["screens/04-medical-records"],
    lang: "vi",
  },
  {
    question: "Làm sao để đăng ký khách hàng mới?", // 飼主登録
    expectedSlugs: ["screens/03-owners"],
    lang: "vi",
  },
  {
    question: "Cách đặt lịch hẹn cho thú cưng?", // 予約
    expectedSlugs: ["screens/06-reservations"],
    lang: "vi",
  },
  {
    question: "Cách thanh toán viện phí?", // 会計
    expectedSlugs: ["screens/05-accounting"],
    lang: "vi",
  },
  {
    question: "Nhập thông tin tiêm phòng vaccine ở đâu?", // ワクチン
    expectedSlugs: ["screens/08-vaccinations"],
    lang: "vi",
  },
  {
    question: "Quy trình nhập viện cho thú cưng là gì?", // 入院
    expectedSlugs: ["screens/09-hospitalization"],
    lang: "vi",
  },
  {
    question: "Cách kiểm tra hàng tồn kho?", // 在庫
    expectedSlugs: ["screens/12-inventory"],
    lang: "vi",
  },

  // ── ノーマッチコントロール（マニュアルに存在しない話題）──
  // UNIT-05 計測済み: 各問い合わせの pre-gate Fuse 候補が 0 件の真のノーマッチのみ置く
  // （候補が出る問い合わせは弱ヒットコントロール側へ移した）。
  { question: "近所でおいしいラーメン屋を教えて", expectedEmpty: true, lang: "ja" },
  { question: "地震の避難方法", expectedEmpty: true, lang: "ja" },
  { question: "宝くじの当選番号は？", expectedEmpty: true, lang: "ja" },
  { question: "ピアノ教室を探したい", expectedEmpty: true, lang: "ja" },

  // ── 弱ヒットコントロール（Fuse 候補は出るが根拠ゲートを通らない。
  //    text 空の参照リンク候補が返ることが期待値。UNIT-05 計測値:
  //    今日の天気は？=2 件, Hôm nay…=49 件, Giá vàng…=22 件。
  //    VN 質問は keywords_vi への fuzzy ヒットで真のノーマッチを作れない
  //    （24〜60 件が必ず候補化）ため、VN 側は弱ヒットのみ）──
  { question: "今日の天気は？", expectedWeak: true, lang: "ja" },
  { question: "Hôm nay trời có mưa không?", expectedWeak: true, lang: "vi" },
  { question: "Giá vàng hôm nay là bao nhiêu?", expectedWeak: true, lang: "vi" },
];
