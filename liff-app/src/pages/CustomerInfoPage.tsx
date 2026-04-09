import { useState, useCallback } from 'react';
import type { CustomerInfo, LiffProfile } from '../types/models';
import { ProgressDots } from '../components/ProgressDots';
import { PrimaryButton } from '../components/PrimaryButton';
import { BackButton } from '../components/BackButton';

interface CustomerInfoPageProps {
  profile: LiffProfile | null;
  initialInfo: CustomerInfo;
  onNext: (info: CustomerInfo) => void;
  onBack: () => void;
}

export function CustomerInfoPage({
  profile,
  initialInfo,
  onNext,
  onBack,
}: CustomerInfoPageProps) {
  const [info, setInfo] = useState<CustomerInfo>(() => {
    if (profile?.additional_fields) {
      const f = profile.additional_fields;
      return {
        name: initialInfo.name || f.name || '',
        phone: initialInfo.phone || f.phone || '',
        ownerName: initialInfo.ownerName || f.owner_name || '',
        petName: initialInfo.petName || f.pet_name || '',
        petType: initialInfo.petType || f.pet_type || '',
      };
    }
    return initialInfo;
  });

  const [errors, setErrors] = useState<Partial<Record<keyof CustomerInfo, string>>>({});

  const handleChange = useCallback((field: keyof CustomerInfo, value: string) => {
    setInfo(prev => ({ ...prev, [field]: value }));
    setErrors(prev => ({ ...prev, [field]: '' }));
  }, []);

  const validate = useCallback((): boolean => {
    const newErrors: Partial<Record<keyof CustomerInfo, string>> = {};
    if (!info.name.trim()) newErrors.name = 'お名前を入力してください';
    if (!info.phone.trim()) newErrors.phone = '電話番号を入力してください';
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }, [info]);

  const handleNext = useCallback(() => {
    if (validate()) {
      onNext(info);
    }
  }, [validate, onNext, info]);

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={1} total={8} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-gray-800 mb-4">お客様情報</h2>
        </div>

        <div className="flex-1 px-4 space-y-4">
          <div>
            <label htmlFor="customer-name" className="block text-sm font-medium text-gray-700 mb-1">
              お名前 <span className="text-red-500">*</span>
            </label>
            <input
              id="customer-name"
              type="text"
              value={info.name}
              onChange={e => handleChange('name', e.target.value)}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-gray-800 focus:outline-none focus:ring-2 focus:ring-line-green focus:border-transparent"
              placeholder="山田 花子"
            />
            {errors.name ? (
              <p className="mt-1 text-sm text-red-500" role="alert">{errors.name}</p>
            ) : null}
          </div>

          <div>
            <label htmlFor="customer-phone" className="block text-sm font-medium text-gray-700 mb-1">
              電話番号 <span className="text-red-500">*</span>
            </label>
            <input
              id="customer-phone"
              type="tel"
              value={info.phone}
              onChange={e => handleChange('phone', e.target.value)}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-gray-800 focus:outline-none focus:ring-2 focus:ring-line-green focus:border-transparent"
              placeholder="090-1234-5678"
            />
            {errors.phone ? (
              <p className="mt-1 text-sm text-red-500" role="alert">{errors.phone}</p>
            ) : null}
          </div>

          <div>
            <label htmlFor="owner-name" className="block text-sm font-medium text-gray-700 mb-1">
              飼い主名
            </label>
            <input
              id="owner-name"
              type="text"
              value={info.ownerName}
              onChange={e => handleChange('ownerName', e.target.value)}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-gray-800 focus:outline-none focus:ring-2 focus:ring-line-green focus:border-transparent"
              placeholder="山田 太郎"
            />
          </div>

          <div>
            <label htmlFor="pet-name" className="block text-sm font-medium text-gray-700 mb-1">
              ペット名
            </label>
            <input
              id="pet-name"
              type="text"
              value={info.petName}
              onChange={e => handleChange('petName', e.target.value)}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-gray-800 focus:outline-none focus:ring-2 focus:ring-line-green focus:border-transparent"
              placeholder="ポチ"
            />
          </div>

          <div>
            <label htmlFor="pet-type" className="block text-sm font-medium text-gray-700 mb-1">
              ペット種類
            </label>
            <input
              id="pet-type"
              type="text"
              value={info.petType}
              onChange={e => handleChange('petType', e.target.value)}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-gray-800 focus:outline-none focus:ring-2 focus:ring-line-green focus:border-transparent"
              placeholder="トイプードル"
            />
          </div>
        </div>

        <div className="px-4 py-6">
          <PrimaryButton onClick={handleNext}>次へ</PrimaryButton>
        </div>
      </div>
    </div>
  );
}
