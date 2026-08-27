import { useState, useCallback, useMemo } from 'react';
import type { CustomerInfo, LiffProfile, PetSelection, OwnerPet } from '../types/models';
import { ProgressDots } from '../components/ProgressDots';
import { PrimaryButton } from '../components/PrimaryButton';
import { BackButton } from '../components/BackButton';
import { EXPLICIT_PRIMARY_CTA_LABEL } from '../lib/advance-policy';

interface CustomerInfoPageProps {
  profile: LiffProfile | null;
  initialInfo: CustomerInfo;
  onNext: (info: CustomerInfo) => void;
  onBack: () => void;
}

/** 既存ペットから PetSelection を生成 */
function ownerPetToSelection(pet: OwnerPet): PetSelection {
  return {
    name: pet.name,
    type: pet.animal_species?.name ?? '',
    isNew: false,
  };
}

/** additional_fields の pets 配列から PetSelection[] を復元 */
function restorePetsFromProfile(profile: LiffProfile | null, initial: PetSelection[]): PetSelection[] {
  if (initial.length > 0) return initial;

  const af = profile?.additional_fields;
  if (af?.pets && af.pets.length > 0) {
    return af.pets.map(p => ({ name: p.name, type: p.type, isNew: p.is_new }));
  }
  return [];
}

export function CustomerInfoPage({
  profile,
  initialInfo,
  onNext,
  onBack,
}: CustomerInfoPageProps) {
  const owner = profile?.owner;
  const ownerPets: OwnerPet[] = useMemo(() => owner?.pets ?? [], [owner?.pets]);
  const f = profile?.additional_fields;

  // ---- 基本情報 ----
  // BUG-387: 初期値優先順 — initialInfo（戻り操作）→ additional_fields（リピーター前回入力）→ owner（紐付け済みオーナー）
  const [name, setName] = useState(() => initialInfo.name || f?.name || owner?.owner_name || '');
  const [phone, setPhone] = useState(() => initialInfo.phone || f?.phone || owner?.phone || '');
  const [ownerName, setOwnerName] = useState(
    () => initialInfo.ownerName || f?.owner_name || owner?.owner_name || '',
  );

  // ---- ペット選択 ----
  // 既存ペットの選択状態（ID → boolean）
  const [selectedPetIds, setSelectedPetIds] = useState<Set<number>>(() => {
    // 初期値: initialInfo.pets にマッチする既存ペットを選択状態にする
    const set = new Set<number>();
    const restored = restorePetsFromProfile(profile, initialInfo.pets);
    for (const pet of restored) {
      if (!pet.isNew) {
        const match = ownerPets.find(op => op.name === pet.name);
        if (match) set.add(match.id);
      }
    }
    // 初回訪問で紐付け済み & ペットが1頭のみ → 自動選択
    if (set.size === 0 && initialInfo.pets.length === 0 && ownerPets.length === 1) {
      set.add(ownerPets[0].id);
    }
    return set;
  });

  // 新規ペット一覧（自由入力）
  const [newPets, setNewPets] = useState<PetSelection[]>(() => {
    const restored = restorePetsFromProfile(profile, initialInfo.pets);
    return restored.filter(p => p.isNew === true);
  });

  // 新規ペット入力フォーム
  const [newPetName, setNewPetName] = useState('');
  const [newPetType, setNewPetType] = useState('');
  const [showNewPetForm, setShowNewPetForm] = useState(false);

  const [errors, setErrors] = useState<Record<string, string>>({});

  // ---- ハンドラ ----
  const togglePet = useCallback((petId: number) => {
    setSelectedPetIds(prev => {
      const next = new Set(prev);
      if (next.has(petId)) {
        next.delete(petId);
      } else {
        next.add(petId);
      }
      return next;
    });
  }, []);

  const handleAddNewPet = useCallback(() => {
    const trimmedName = newPetName.trim();
    if (!trimmedName) return;
    setNewPets(prev => [...prev, { name: trimmedName, type: newPetType.trim(), isNew: true }]);
    setNewPetName('');
    setNewPetType('');
    setShowNewPetForm(false);
  }, [newPetName, newPetType]);

  const handleRemoveNewPet = useCallback((index: number) => {
    setNewPets(prev => prev.filter((_, i) => i !== index));
  }, []);

  const validate = useCallback((): boolean => {
    const newErrors: Record<string, string> = {};
    if (!name.trim()) newErrors.name = 'お名前を入力してください';
    if (!phone.trim()) {
      newErrors.phone = '電話番号を入力してください';
    } else if (!/^0\d{1,4}-?\d{1,4}-?\d{4}$/.test(phone.trim())) {
      newErrors.phone = '電話番号の形式が正しくありません（例：090-1234-5678）';
    }
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }, [name, phone]);

  const handleNext = useCallback(() => {
    if (!validate()) return;

    // 選択済み既存ペット + 新規ペットを統合
    const selectedOwnerPets: PetSelection[] = ownerPets
      .filter(op => selectedPetIds.has(op.id))
      .map(ownerPetToSelection);

    const allPets = [...selectedOwnerPets, ...newPets];

    onNext({
      name: name.trim(),
      phone: phone.trim(),
      ownerName: ownerName.trim(),
      pets: allPets,
    });
  }, [validate, ownerPets, selectedPetIds, newPets, name, phone, ownerName, onNext]);

  const hasOwnerPets = ownerPets.length > 0;

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={1} total={8} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-noah-teal-dark mb-4">お客様情報</h2>
        </div>

        <div className="flex-1 px-4 space-y-4">
          {/* お名前 */}
          <div>
            <label htmlFor="customer-name" className="block text-sm font-medium text-noah-text-sub mb-1">
              お名前 <span className="text-red-500">*</span>
            </label>
            <input
              id="customer-name"
              type="text"
              value={name}
              onChange={e => { setName(e.target.value); setErrors(prev => ({ ...prev, name: '' })); }}
              className="w-full border border-gray-300 rounded-xl px-3 py-2 text-noah-text focus:outline-none focus:ring-2 focus:ring-noah-teal focus:border-transparent"
              placeholder="山田 花子"
            />
            {errors.name ? (
              <p className="mt-1 text-sm text-red-500" role="alert">{errors.name}</p>
            ) : null}
          </div>

          {/* 電話番号 */}
          <div>
            <label htmlFor="customer-phone" className="block text-sm font-medium text-noah-text-sub mb-1">
              電話番号 <span className="text-red-500">*</span>
            </label>
            <input
              id="customer-phone"
              type="tel"
              value={phone}
              onChange={e => { setPhone(e.target.value); setErrors(prev => ({ ...prev, phone: '' })); }}
              className="w-full border border-gray-300 rounded-xl px-3 py-2 text-noah-text focus:outline-none focus:ring-2 focus:ring-noah-teal focus:border-transparent"
              placeholder="090-1234-5678"
            />
            {errors.phone ? (
              <p className="mt-1 text-sm text-red-500" role="alert">{errors.phone}</p>
            ) : null}
          </div>

          {/* 飼い主名 */}
          <div>
            <label htmlFor="owner-name" className="block text-sm font-medium text-noah-text-sub mb-1">
              飼い主名
            </label>
            <input
              id="owner-name"
              type="text"
              value={ownerName}
              onChange={e => setOwnerName(e.target.value)}
              className="w-full border border-gray-300 rounded-xl px-3 py-2 text-noah-text focus:outline-none focus:ring-2 focus:ring-noah-teal focus:border-transparent"
              placeholder="山田 太郎"
            />
          </div>

          {/* ペット選択セクション */}
          <div>
            <p className="block text-sm font-medium text-noah-text-sub mb-2">ペット</p>

            {/* 既存ペット（オーナー紐付け時のみ表示） */}
            {hasOwnerPets ? (
              <div className="space-y-2 mb-3">
                <p className="text-xs text-noah-text-sub">登録済みのペットを選択してください</p>
                {ownerPets.map(pet => (
                  <label
                    key={pet.id}
                    className="flex items-center gap-3 p-3 bg-white rounded-xl border border-gray-200 cursor-pointer hover:border-noah-teal transition-colors"
                  >
                    <input
                      type="checkbox"
                      checked={selectedPetIds.has(pet.id)}
                      onChange={() => togglePet(pet.id)}
                      className="w-5 h-5 rounded border-gray-300 text-noah-teal focus:ring-noah-teal"
                    />
                    <div className="flex-1">
                      <span className="text-sm font-medium text-noah-text">{pet.name}</span>
                      {pet.animal_species?.name ? (
                        <span className="ml-2 text-xs text-noah-text-sub">
                          ({pet.animal_species.name})
                        </span>
                      ) : null}
                    </div>
                  </label>
                ))}
              </div>
            ) : null}

            {/* 新規追加済みペット一覧 */}
            {newPets.length > 0 ? (
              <div className="space-y-2 mb-3">
                {hasOwnerPets ? (
                  <p className="text-xs text-noah-text-sub">新しく追加したペット</p>
                ) : null}
                {newPets.map((pet, index) => (
                  <div
                    key={`new-${index}`}
                    className="flex items-center gap-3 p-3 bg-white rounded-xl border border-dashed border-noah-teal"
                  >
                    <div className="flex-1">
                      <span className="text-sm font-medium text-noah-text">{pet.name}</span>
                      {pet.type ? (
                        <span className="ml-2 text-xs text-noah-text-sub">({pet.type})</span>
                      ) : null}
                    </div>
                    <button
                      type="button"
                      onClick={() => handleRemoveNewPet(index)}
                      className="text-red-400 hover:text-red-600 text-sm px-2"
                      aria-label={`${pet.name}を削除`}
                    >
                      削除
                    </button>
                  </div>
                ))}
              </div>
            ) : null}

            {/* 新規ペット追加フォーム */}
            {showNewPetForm ? (
              <div className="p-3 bg-white rounded-xl border border-gray-200 space-y-3">
                <div>
                  <label htmlFor="new-pet-name" className="block text-xs font-medium text-noah-text-sub mb-1">
                    ペット名
                  </label>
                  <input
                    id="new-pet-name"
                    type="text"
                    value={newPetName}
                    onChange={e => setNewPetName(e.target.value)}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm text-noah-text focus:outline-none focus:ring-2 focus:ring-noah-teal focus:border-transparent"
                    placeholder="ポチ"
                    autoFocus
                  />
                </div>
                <div>
                  <label htmlFor="new-pet-type" className="block text-xs font-medium text-noah-text-sub mb-1">
                    種類
                  </label>
                  <input
                    id="new-pet-type"
                    type="text"
                    value={newPetType}
                    onChange={e => setNewPetType(e.target.value)}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm text-noah-text focus:outline-none focus:ring-2 focus:ring-noah-teal focus:border-transparent"
                    placeholder="トイプードル"
                  />
                </div>
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={handleAddNewPet}
                    disabled={!newPetName.trim()}
                    className="flex-1 py-2 rounded-lg text-sm font-medium text-white bg-noah-teal disabled:opacity-40"
                  >
                    追加
                  </button>
                  <button
                    type="button"
                    onClick={() => { setShowNewPetForm(false); setNewPetName(''); setNewPetType(''); }}
                    className="flex-1 py-2 rounded-lg text-sm font-medium text-noah-text-sub bg-gray-100"
                  >
                    キャンセル
                  </button>
                </div>
              </div>
            ) : (
              <button
                type="button"
                onClick={() => setShowNewPetForm(true)}
                className="w-full py-2.5 rounded-xl border-2 border-dashed border-gray-300 text-sm text-noah-text-sub hover:border-noah-teal hover:text-noah-teal transition-colors"
              >
                + 新しいペットを追加
              </button>
            )}
          </div>
        </div>

        <div className="px-4 py-6">
          <PrimaryButton onClick={handleNext}>{EXPLICIT_PRIMARY_CTA_LABEL}</PrimaryButton>
        </div>
      </div>
    </div>
  );
}
