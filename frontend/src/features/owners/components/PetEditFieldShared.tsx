import type { Dispatch, SetStateAction } from "react";

import { SelectItem } from "@/components/ui/select";
import { C, STYLE } from "@/lib/design-tokens";

import {
  ACQUISITION_TYPE_VALUES,
  DANGER_LEVEL_VALUES,
  PET_GENDER_VALUES,
  type PetFormData,
} from "../types";

export const LABEL_CLS = `text-sm ${C.text60}`;
export const INPUT_CLS = STYLE.formInput;

export const BREED_SUGGESTIONS: Record<string, string[]> = {
  "犬": [
    "柴犬",
    "トイプードル",
    "チワワ",
    "ダックスフンド",
    "フレンチブルドッグ",
    "ゴールデンレトリバー",
    "ラブラドールレトリバー",
    "ポメラニアン",
    "ビーグル",
    "シバイヌ(赤)",
    "シバイヌ(黒)",
    "ミニチュアシュナウザー",
    "マルチーズ",
    "ヨークシャーテリア",
    "シーズー",
    "ボーダーコリー",
    "コーギー",
    "ハスキー",
    "サモエド",
    "柴ミックス",
    "雑種",
  ],
  "猫": [
    "アメリカンショートヘア",
    "スコティッシュフォールド",
    "ロシアンブルー",
    "メインクーン",
    "ペルシャ",
    "ノルウェージャンフォレストキャット",
    "ラグドール",
    "ベンガル",
    "マンチカン",
    "ヒマラヤン",
    "アビシニアン",
    "バーマン",
    "ブリティッシュショートヘア",
    "日本猫",
    "雑種",
  ],
  "鳥": ["セキセイインコ", "オカメインコ", "コザクラインコ", "文鳥", "カナリア", "その他"],
  "ウサギ": ["ネザーランドドワーフ", "ホーランドロップ", "ミニレッキス", "その他"],
  "ハムスター": ["ゴールデンハムスター", "ジャンガリアン", "キャンベル", "その他"],
  "フェレット": ["フェレット"],
};

export const GENDER_SELECT_ITEMS = PET_GENDER_VALUES.map((g) => (
  <SelectItem key={g} value={g}>
    {g}
  </SelectItem>
));

export const ACQUISITION_SELECT_ITEMS = ACQUISITION_TYPE_VALUES.map((t) => (
  <SelectItem key={t} value={t}>
    {t}
  </SelectItem>
));

export const DANGER_SELECT_ITEMS = DANGER_LEVEL_VALUES.map((d) => (
  <SelectItem key={d} value={d}>
    {d}
  </SelectItem>
));

export interface PetFieldSectionProps {
  formData: PetFormData;
  setFormData: Dispatch<SetStateAction<PetFormData>>;
  fieldErrors: Record<string, string>;
  clearFieldError: (field: string) => void;
}
