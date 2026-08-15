import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { CPMStageSummary } from './CPMStageSummary';
import type { CPMStageCounts } from '../api/get-cpm-stage-counts';

const counts: CPMStageCounts = {
  cpm_encounter: 5,
  cpm_growing: 12,
  cpm_core: 42,
  cpm_spot: 3,
  cpm_noah: 1,
  cpm_dormant: 20,
};

describe('CPMStageSummary (ISSUE-180)', () => {
  it('renders all six segments with headcount plus an すべて chip', () => {
    render(
      <CPMStageSummary
        counts={counts}
        total={83}
        isLoading={false}
        isError={false}
        onSelect={() => {}}
      />
    );

    expect(screen.getByText('Core')).toBeInTheDocument();
    expect(screen.getByText('42')).toBeInTheDocument();
    expect(screen.getByText('Noah')).toBeInTheDocument();
    expect(screen.getByText('Dormant')).toBeInTheDocument();
    expect(screen.getByText('すべて')).toBeInTheDocument();
    expect(screen.getByText('83')).toBeInTheDocument();
    screen.getAllByRole('button').forEach((button) => {
      expect(button).toHaveClass('min-h-11', 'min-w-11');
    });
  });

  it('calls onSelect with the stage when a chip is clicked', async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(
      <CPMStageSummary
        counts={counts}
        total={83}
        isLoading={false}
        isError={false}
        onSelect={onSelect}
      />
    );

    await user.click(screen.getByRole('button', { name: /Core/ }));
    expect(onSelect).toHaveBeenCalledWith('cpm_core');
  });

  it('toggles back to undefined when the already-selected chip is clicked', async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(
      <CPMStageSummary
        counts={counts}
        total={83}
        isLoading={false}
        isError={false}
        selected="cpm_core"
        onSelect={onSelect}
      />
    );

    await user.click(screen.getByRole('button', { name: /Core/ }));
    expect(onSelect).toHaveBeenCalledWith(undefined);
  });

  it('marks the selected segment with aria-pressed', () => {
    render(
      <CPMStageSummary
        counts={counts}
        total={83}
        isLoading={false}
        isError={false}
        selected="cpm_core"
        onSelect={() => {}}
      />
    );

    expect(screen.getByRole('button', { name: /Core/ })).toHaveAttribute(
      'aria-pressed',
      'true'
    );
    expect(screen.getByRole('button', { name: /すべて/ })).toHaveAttribute(
      'aria-pressed',
      'false'
    );
  });

  it('shows an ellipsis while loading and a dash on error', () => {
    const { rerender } = render(
      <CPMStageSummary
        counts={counts}
        total={83}
        isLoading={true}
        isError={false}
        onSelect={() => {}}
      />
    );
    expect(screen.getAllByText('…').length).toBeGreaterThan(0);

    rerender(
      <CPMStageSummary
        counts={counts}
        total={83}
        isLoading={false}
        isError={true}
        onSelect={() => {}}
      />
    );
    expect(screen.getAllByText('—').length).toBeGreaterThan(0);
  });
});
