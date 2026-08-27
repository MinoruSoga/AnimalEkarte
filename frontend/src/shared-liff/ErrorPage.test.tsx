import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi, afterEach } from 'vitest';

import {
  DEFAULT_ERROR_PAGE_THEME,
  ErrorPage,
  SHARED_ERROR_PAGE_TITLE,
} from './ErrorPage';

describe('ErrorPage（BUG-027: LIFF / LINE-reserve 共通エラー chrome）', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('canonical title と DEFAULT theme の shell / heading / body / icon を出す', () => {
    const { container } = render(<ErrorPage message="詳細メッセージ" />);

    const root = container.firstElementChild as HTMLElement;
    expect(root.className).toContain(DEFAULT_ERROR_PAGE_THEME.bg);
    expect(screen.getByText('⚠️')).toBeInTheDocument();
    const heading = screen.getByRole('heading', { level: 1 });
    expect(heading).toHaveTextContent(SHARED_ERROR_PAGE_TITLE);
    expect(heading.className).toContain(DEFAULT_ERROR_PAGE_THEME.heading);
    const body = screen.getByRole('alert');
    expect(body).toHaveTextContent('詳細メッセージ');
    for (const token of DEFAULT_ERROR_PAGE_THEME.body.split(/\s+/)) {
      expect(body.className).toContain(token);
    }
  });

  it('デフォルトは再読み込みアクションを出し、クリックで reload する', async () => {
    const reload = vi.fn();
    vi.stubGlobal('location', { ...window.location, reload });
    const user = userEvent.setup();

    render(<ErrorPage message="x" />);

    const button = screen.getByRole('button', { name: '再読み込み' });
    expect(button.className).toContain(DEFAULT_ERROR_PAGE_THEME.button);
    await user.click(button);
    expect(reload).toHaveBeenCalledTimes(1);
  });

  it('onAction 指定時はカスタムラベルでハンドラを呼ぶ', async () => {
    const onAction = vi.fn();
    const user = userEvent.setup();

    render(<ErrorPage message="x" onAction={onAction} actionLabel="再試行" />);

    await user.click(screen.getByRole('button', { name: '再試行' }));
    expect(onAction).toHaveBeenCalledTimes(1);
  });

  it('showAction=false のときアクションを出さない（恒久的な設定ミス用）', () => {
    render(<ErrorPage message="クリニックIDが見つかりません" showAction={false} />);

    expect(screen.getByRole('alert')).toHaveTextContent('クリニックIDが見つかりません');
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });
});
