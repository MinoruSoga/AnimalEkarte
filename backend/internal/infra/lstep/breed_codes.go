package lstep

import "strings"

// breedCodeMap は日本語品種名から Lステップタグコード（breed_ プレフィックスなし）へのマッピング。
var breedCodeMap = map[string]string{
	// 犬
	"柴犬":            "shiba_inu",
	"トイプードル":        "toy_poodle",
	"ミニチュアプードル":     "miniature_poodle",
	"スタンダードプードル":    "standard_poodle",
	"チワワ":           "chihuahua",
	"ゴールデンレトリバー":    "golden_retriever",
	"ラブラドールレトリバー":   "labrador_retriever",
	"ダックスフント":       "dachshund",
	"ミニチュアダックスフント":  "miniature_dachshund",
	"ポメラニアン":        "pomeranian",
	"ビーグル":          "beagle",
	"フレンチブルドッグ":     "french_bulldog",
	"マルチーズ":         "maltese",
	"ヨークシャーテリア":     "yorkshire_terrier",
	"ウェルシュコーギー":     "welsh_corgi",
	"ミニチュアシュナウザー":   "miniature_schnauzer",
	"パグ":            "pug",
	"ボーダーコリー":       "border_collie",
	"シェットランドシープドッグ": "shetland_sheepdog",
	"シェルティ":         "shetland_sheepdog",
	"秋田犬":           "akita_inu",
	"北海道犬":          "hokkaido_inu",
	"甲斐犬":           "kai_inu",
	"土佐犬":           "tosa_inu",
	"紀州犬":           "kishu_inu",
	"四国犬":           "shikoku_inu",
	"ジャックラッセルテリア":   "jack_russell_terrier",
	"ミニチュアピンシャー":    "miniature_pinscher",
	"バセットハウンド":      "basset_hound",
	"シーズー":          "shih_tzu",
	"ペキニーズ":         "pekingese",
	"サモエド":          "samoyed",
	"シベリアンハスキー":     "siberian_husky",
	"アラスカンマラミュート":   "alaskan_malamute",
	"ブルドッグ":         "bulldog",
	"ボクサー":          "boxer",
	"ドーベルマン":        "doberman",
	"ロットワイラー":       "rottweiler",
	"グレートデン":        "great_dane",
	"セントバーナード":      "saint_bernard",
	"バーニーズマウンテンドッグ": "bernese_mountain_dog",
	"ウィペット":         "whippet",
	"グレーハウンド":       "greyhound",
	"アフガンハウンド":      "afghan_hound",
	"ダルメシアン":        "dalmatian",
	"コッカースパニエル":     "cocker_spaniel",
	"スプリンガースパニエル":   "springer_spaniel",
	"ゴールデンドゥードル":    "goldendoodle",
	"ラブラドゥードル":      "labradoodle",
	"マルプー":          "maltipoo",
	"ポメプー":          "pomapoo",
	"チワプー":          "chihuapoo",
	// 猫
	"アメリカンショートヘア":  "american_shorthair",
	"スコティッシュフォールド": "scottish_fold",
	"マンチカン":        "munchkin",
	"ペルシャ":         "persian",
	"ロシアンブルー":      "russian_blue",
	"ラグドール":        "ragdoll",
	"アビシニアン":       "abyssinian",
	"ベンガル":         "bengal",
	"メインクーン":       "maine_coon",
	"ノルウェージャンフォレストキャット": "norwegian_forest_cat",
	"ブリティッシュショートヘア":     "british_shorthair",
	"シャム":          "siamese",
	"バーミーズ":        "burmese",
	"ソマリ":          "somali",
	"エジプシャンマウ":     "egyptian_mau",
	"トンキニーズ":       "tonkinese",
	"シンガプーラ":       "singapura",
	"オシキャット":       "ocicat",
	"サバンナ":         "savannah",
	"セルカークレックス":    "selkirk_rex",
	"コーニッシュレックス":   "cornish_rex",
	"デボンレックス":      "devon_rex",
	"スフィンクス":       "sphynx",
	"ターキッシュアンゴラ":   "turkish_angora",
	"ターキッシュバン":     "turkish_van",
	"バーマン":         "birman",
	"ヒマラヤン":        "himalayan",
	"エキゾチックショートヘア": "exotic_shorthair",
	"オリエンタルショートヘア": "oriental_shorthair",
}

// BreedTagName は品種文字列と混合フォールバック名からタグ名（"breed_"+code）を返す。
// 既知品種は完全一致→部分一致の順で解決する。不明品種は fallback を返す。
func BreedTagName(breed, fallback string) string {
	if breed == "" {
		return fallback
	}
	// 完全一致
	if code, ok := breedCodeMap[breed]; ok {
		return "breed_" + code
	}
	// 部分一致（長いキーを優先するため安定性は保証しないが実用上問題なし）
	lower := strings.ToLower(breed)
	for key, code := range breedCodeMap {
		if strings.Contains(lower, strings.ToLower(key)) {
			return "breed_" + code
		}
	}
	return fallback
}
