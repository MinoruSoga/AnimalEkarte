import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
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

  describe('R-F4: バリデーション（必須項目）', () => {
    it('お名前・電話番号が未入力のとき「次へ」で止まり、両方のエラーメッセージを表示する', async () => {
      const user = userEvent.setup();
      const onNext = vi.fn();

      render(
        <CustomerInfoPage
          profile={null}
          initialInfo={emptyInfo}
          onNext={onNext}
          onBack={vi.fn()}
        />
      );

      await user.click(screen.getByRole('button', { name: '次へ' }));

      expect(screen.getByText('お名前を入力してください')).toBeInTheDocument();
      expect(screen.getByText('電話番号を入力してください')).toBeInTheDocument();
      expect(onNext).not.toHaveBeenCalled();
    });

    it('BUG-028: 空のお名前で次へするとエラー文言に加え name 入力に invalid 枠線/aria-invalid が付く', async () => {
      const user = userEvent.setup();
      const onNext = vi.fn();

      render(
        <CustomerInfoPage
          profile={null}
          initialInfo={emptyInfo}
          onNext={onNext}
          onBack={vi.fn()}
        />
      );

      await user.click(screen.getByRole('button', { name: '次へ' }));

      const nameInput = screen.getByLabelText(/お名前/);
      expect(screen.getByText('お名前を入力してください')).toBeInTheDocument();
      expect(nameInput).toHaveAttribute('aria-invalid', 'true');
      expect(nameInput.className).toMatch(/border-red-500/);
      expect(onNext).not.toHaveBeenCalled();
    });

    it('必須項目を入力すると次へ進み、onNext に入力値が渡る', async () => {
      const user = userEvent.setup();
      const onNext = vi.fn();

      render(
        <CustomerInfoPage
          profile={null}
          initialInfo={emptyInfo}
          onNext={onNext}
          onBack={vi.fn()}
        />
      );

      await user.type(screen.getByLabelText(/お名前/), '鈴木一郎');
      await user.type(screen.getByLabelText(/電話番号/), '070-9999-8888');
      await user.click(screen.getByRole('button', { name: '次へ' }));

      expect(onNext).toHaveBeenCalledWith({
        name: '鈴木一郎',
        phone: '070-9999-8888',
        ownerName: '',
        pets: [],
      });
    });

    it('エラー表示後に再入力するとそのフィールドのエラーが消える', async () => {
      const user = userEvent.setup();

      render(
        <CustomerInfoPage
          profile={null}
          initialInfo={emptyInfo}
          onNext={vi.fn()}
          onBack={vi.fn()}
        />
      );

      await user.click(screen.getByRole('button', { name: '次へ' }));
      expect(screen.getByText('お名前を入力してください')).toBeInTheDocument();

      await user.type(screen.getByLabelText(/お名前/), '鈴木');
      expect(screen.queryByText('お名前を入力してください')).not.toBeInTheDocument();
    });

    it('BUG-028: お名前を修正すると invalid 枠線と aria-invalid が外れる', async () => {
      const user = userEvent.setup();

      render(
        <CustomerInfoPage
          profile={null}
          initialInfo={emptyInfo}
          onNext={vi.fn()}
          onBack={vi.fn()}
        />
      );

      await user.click(screen.getByRole('button', { name: '次へ' }));
      const nameInput = screen.getByLabelText(/お名前/);
      expect(nameInput).toHaveAttribute('aria-invalid', 'true');
      expect(nameInput.className).toMatch(/border-red-500/);

      await user.type(nameInput, '鈴木');
      expect(nameInput).not.toHaveAttribute('aria-invalid', 'true');
      expect(nameInput.className).not.toMatch(/border-red-500/);
    });
  });

  describe('R-F4: 新規ペット追加/削除', () => {
    it('「+ 新しいペットを追加」→ペット名入力→追加すると一覧に表示され、onNext の pets に反映される', async () => {
      const user = userEvent.setup();
      const onNext = vi.fn();

      render(
        <CustomerInfoPage
          profile={null}
          initialInfo={emptyInfo}
          onNext={onNext}
          onBack={vi.fn()}
        />
      );

      await user.click(screen.getByRole('button', { name: '+ 新しいペットを追加' }));
      await user.type(screen.getByLabelText('ペット名'), 'ポチ');
      await user.type(screen.getByLabelText('種類'), '柴犬');
      await user.click(screen.getByRole('button', { name: '追加' }));

      expect(screen.getByText('ポチ')).toBeInTheDocument();
      expect(screen.getByText('(柴犬)')).toBeInTheDocument();

      await user.type(screen.getByLabelText(/お名前/), '山田');
      await user.type(screen.getByLabelText(/電話番号/), '0312345678');
      await user.click(screen.getByRole('button', { name: '次へ' }));

      expect(onNext).toHaveBeenCalledWith(
        expect.objectContaining({
          pets: [{ name: 'ポチ', type: '柴犬', isNew: true }],
        }),
      );
    });

    it('追加済みの新規ペットを削除できる', async () => {
      const user = userEvent.setup();

      render(
        <CustomerInfoPage
          profile={null}
          initialInfo={{ ...emptyInfo, pets: [{ name: 'ポチ', type: '柴犬', isNew: true }] }}
          onNext={vi.fn()}
          onBack={vi.fn()}
        />
      );

      expect(screen.getByText('ポチ')).toBeInTheDocument();
      await user.click(screen.getByRole('button', { name: 'ポチを削除' }));

      expect(screen.queryByText('ポチ')).not.toBeInTheDocument();
    });

    it('ペット名が空欄のときは「追加」ボタンが無効化される', async () => {
      const user = userEvent.setup();

      render(
        <CustomerInfoPage
          profile={null}
          initialInfo={emptyInfo}
          onNext={vi.fn()}
          onBack={vi.fn()}
        />
      );

      await user.click(screen.getByRole('button', { name: '+ 新しいペットを追加' }));
      expect(screen.getByRole('button', { name: '追加' })).toBeDisabled();
    });
  });
});
