import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { CustomerInfoPage } from './CustomerInfoPage';
import type { CustomerInfo, LiffProfile } from '../types/models';

const emptyInfo: CustomerInfo = {
  name: '',
  phone: '',
  ownerName: '',
  pets: [],
};

describe('CustomerInfoPage', () => {
  it('profile.additional_fields から電話番号・飼い主名・新規ペットを復元する', () => {
    const profile: LiffProfile = {
      line_user_id: 'line-user-1',
      display_name: 'LINE表示名',
      additional_fields: {
        name: '山田花子',
        phone: '090-1234-5678',
        owner_name: '山田太郎',
        pets: [{ name: 'ポチ', type: '柴犬', is_new: true }],
      },
    };

    render(
      <CustomerInfoPage
        profile={profile}
        initialInfo={emptyInfo}
        onNext={vi.fn()}
        onBack={vi.fn()}
      />
    );

    expect(screen.getByDisplayValue('山田花子')).toBeInTheDocument();
    expect(screen.getByDisplayValue('090-1234-5678')).toBeInTheDocument();
    expect(screen.getByDisplayValue('山田太郎')).toBeInTheDocument();
    expect(screen.getByText('ポチ')).toBeInTheDocument();
    expect(screen.getByText('(柴犬)')).toBeInTheDocument();
  });

  it('owner 紐付け済みでペットが1頭のみのとき自動選択される', () => {
    const profile: LiffProfile = {
      line_user_id: 'line-user-2',
      display_name: 'LINE表示名',
      additional_fields: {},
      owner: {
        owner_name: '田中太郎',
        phone: '080-1111-2222',
        pets: [{ id: 10, name: 'タマ', animal_species: { name: '猫' } }],
      },
    };

    render(
      <CustomerInfoPage
        profile={profile}
        initialInfo={emptyInfo}
        onNext={vi.fn()}
        onBack={vi.fn()}
      />
    );

    expect(screen.getAllByDisplayValue('田中太郎')).toHaveLength(2);
    expect(screen.getByDisplayValue('080-1111-2222')).toBeInTheDocument();
    expect(screen.getByRole('checkbox')).toBeChecked();
    expect(screen.getByText('タマ')).toBeInTheDocument();
    expect(screen.getByText('(猫)')).toBeInTheDocument();
  });
});
