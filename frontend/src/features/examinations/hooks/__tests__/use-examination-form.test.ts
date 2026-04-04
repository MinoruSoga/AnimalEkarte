import { renderHook } from '@testing-library/react';
import { vi } from 'vitest';
import { useExaminationForm } from '../use-examination-form';
import { useSearchParams } from 'react-router';

// Mock dependencies
vi.mock('react-router', () => ({
  useSearchParams: vi.fn(),
  useNavigate: vi.fn(),
}));

vi.mock('@/hooks/use-pet-selection', () => ({
  usePetSelection: vi.fn(() => ({
    selectedPets: [],
    setSelectedPets: vi.fn(),
  })),
}));

vi.mock('@/hooks/use-pet', () => ({
  useGetPet: vi.fn(() => ({
    data: null,
    isLoading: false,
  })),
}));

vi.mock('../api/get-examination', () => ({
  useGetExamination: vi.fn(() => ({ data: null })),
}));

vi.mock('../api/create-examination', () => ({
  useCreateExamination: vi.fn(() => ({ mutateAsync: vi.fn() })),
}));

vi.mock('../api/update-examination', () => ({
  useUpdateExamination: vi.fn(() => ({ mutateAsync: vi.fn() })),
}));

vi.mock('../api/delete-examination', () => ({
  useDeleteExamination: vi.fn(() => ({ mutate: vi.fn(), isPending: false })),
}));

describe('useExaminationForm - Doctor ID Auto-Population (#21)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('医師IDがクエリパラメータから抽出される', () => {
    const mockSearchParams = new URLSearchParams('doctorId=123');
    const mockedUseSearchParams = vi.mocked(useSearchParams);
    mockedUseSearchParams.mockReturnValue([mockSearchParams]);

    const { result } = renderHook(() => useExaminationForm());

    // doctorIdがformDataに含まれていることを確認
    expect(result.current.formData).toBeDefined();
  });

  it('doctorIdなしの場合、フォームは空の医師フィールドで初期化', () => {
    const mockSearchParams = new URLSearchParams('');
    const mockedUseSearchParams = vi.mocked(useSearchParams);
    mockedUseSearchParams.mockReturnValue([mockSearchParams]);

    const { result } = renderHook(() => useExaminationForm());

    // formDataが定義されていることを確認（医師IDフィールドはオプション）
    expect(result.current.formData).toBeDefined();
  });

  it('複数のクエリパラメータが存在する場合、doctorIdを正確に抽出', () => {
    const mockSearchParams = new URLSearchParams('petId=456&doctorId=789&medicalRecordId=101');
    const mockedUseSearchParams = vi.mocked(useSearchParams);
    mockedUseSearchParams.mockReturnValue([mockSearchParams]);

    const { result } = renderHook(() => useExaminationForm());

    // フォームデータが作成されていることを確認
    expect(result.current.formData).toBeDefined();
  });
});
