import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useExaminationForm } from '../use-examination-form';
import { useSearchParams } from 'react-router';
import { useGetPet } from '@/hooks/use-pet';
import { usePetSelection } from '@/hooks/use-pet-selection';
import { useDeleteExamination } from '../../api/delete-examination';

// Mock dependencies
const mockNavigate = vi.fn();

vi.mock('react-router', () => ({
  useSearchParams: vi.fn(),
  useNavigate: vi.fn(() => mockNavigate),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn(), info: vi.fn() },
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

vi.mock('../../api/get-examination', () => ({
  useGetExamination: vi.fn(() => ({ data: null })),
}));

vi.mock('../../api/create-examination', () => ({
  useCreateExamination: vi.fn(() => ({ mutateAsync: vi.fn().mockResolvedValue({}) })),
}));

vi.mock('../../api/update-examination', () => ({
  useUpdateExamination: vi.fn(() => ({ mutateAsync: vi.fn().mockResolvedValue({}) })),
}));

vi.mock('../../api/delete-examination', () => ({
  useDeleteExamination: vi.fn(() => ({ mutate: vi.fn(), isPending: false })),
}));

// ─────────────────────────────────────────────────────────────

describe('useExaminationForm — 新規作成モード', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGetPet).mockReturnValue({ data: null, isLoading: false, isError: false } as ReturnType<typeof useGetPet>);
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);
    vi.mocked(useSearchParams).mockReturnValue([new URLSearchParams(), vi.fn()]);
  });

  it('isEdit は false（id なし）', () => {
    const { result } = renderHook(() => useExaminationForm());
    expect(result.current.isEdit).toBe(false);
  });

  it('初期 isSaving は false', () => {
    const { result } = renderHook(() => useExaminationForm());
    expect(result.current.isSaving).toBe(false);
  });

  it('初期 isDeleting は false', () => {
    const { result } = renderHook(() => useExaminationForm());
    expect(result.current.isDeleting).toBe(false);
  });

  it('status の初期値は "依頼中"', () => {
    const { result } = renderHook(() => useExaminationForm());
    expect(result.current.formData.status).toBe('依頼中');
  });

  it('doctorId なしの場合、formData.doctorId は undefined', () => {
    vi.mocked(useSearchParams).mockReturnValue([new URLSearchParams(''), vi.fn()]);
    const { result } = renderHook(() => useExaminationForm());
    expect(result.current.formData.doctorId).toBeUndefined();
  });

  it('doctorId あり → formData.doctorId に反映される', () => {
    vi.mocked(useSearchParams).mockReturnValue([new URLSearchParams('doctorId=789'), vi.fn()]);
    const { result } = renderHook(() => useExaminationForm());
    expect(result.current.formData.doctorId).toBe('789');
  });

  it('複数クエリパラメータがある場合、doctorId を正確に抽出', () => {
    vi.mocked(useSearchParams).mockReturnValue([
      new URLSearchParams('petId=456&doctorId=789&medicalRecordId=101'),
      vi.fn(),
    ]);
    const { result } = renderHook(() => useExaminationForm());
    expect(result.current.formData.doctorId).toBe('789');
  });

  it('petId がない & ローディングでもない場合、ペット選択ページへリダイレクト', async () => {
    vi.mocked(useSearchParams).mockReturnValue([new URLSearchParams(''), vi.fn()]);
    vi.mocked(useGetPet).mockReturnValue({ data: null, isLoading: false, isError: false } as ReturnType<typeof useGetPet>);

    await act(async () => {
      renderHook(() => useExaminationForm());
      await Promise.resolve();
    });

    expect(mockNavigate).toHaveBeenCalledWith(
      expect.stringContaining('select-pet')
    );
  });
});

// ─────────────────────────────────────────────────────────────

describe('useExaminationForm — petFromQuery あり', () => {
  const mockPet = {
    id: '42',
    name: 'ポチ',
    ownerName: '田中太郎',
    ownerId: '5',
    species: '犬',
    breed: '',
    birthday: '',
    gender: '男',
    weight: null,
    imageUrl: null,
    status: '生存' as const,
    microchipNumber: null,
    insuranceNumber: null,
    insuranceExpiry: null,
    memo: null,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useSearchParams).mockReturnValue([new URLSearchParams('petId=42'), vi.fn()]);
  });

  it('petFromQuery が非null のとき setSelectedPets が呼ばれる（line 140）', async () => {
    const mockSetSelectedPets = vi.fn();
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [],
      setSelectedPets: mockSetSelectedPets,
    } as ReturnType<typeof usePetSelection>);

    vi.mocked(useGetPet).mockReturnValue({
      data: mockPet,
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);

    await act(async () => {
      renderHook(() => useExaminationForm());
      await Promise.resolve();
    });

    expect(mockSetSelectedPets).toHaveBeenCalledWith([mockPet]);
  });

  it('petFromQuery から ownerName / petName を formData に反映する', async () => {
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [mockPet],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);

    vi.mocked(useGetPet).mockReturnValue({
      data: mockPet,
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);

    const { result } = renderHook(() => useExaminationForm());

    expect(result.current.formData.ownerName).toBe('田中太郎');
    expect(result.current.formData.petName).toBe('ポチ');
  });
});

// ─────────────────────────────────────────────────────────────

describe('useExaminationForm — 編集モード（id あり）', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useSearchParams).mockReturnValue([new URLSearchParams(), vi.fn()]);
    vi.mocked(useGetPet).mockReturnValue({ data: null, isLoading: false, isError: false } as ReturnType<typeof useGetPet>);
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);
  });

  it('id を渡すと isEdit = true になる', () => {
    const { result } = renderHook(() => useExaminationForm('exam-001'));
    expect(result.current.isEdit).toBe(true);
  });

  it('isEdit = true のとき useEffect でリダイレクトしない', async () => {
    await act(async () => {
      renderHook(() => useExaminationForm('exam-001'));
      await Promise.resolve();
    });
    expect(mockNavigate).not.toHaveBeenCalled();
  });

  it('handleDelete は isEdit = false のとき何もしない', () => {
    const mockMutate = vi.fn();
    vi.mocked(useDeleteExamination).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
    } as ReturnType<typeof useDeleteExamination>);

    const { result } = renderHook(() => useExaminationForm()); // no id → isEdit = false
    act(() => { result.current.handleDelete(); });
    expect(mockMutate).not.toHaveBeenCalled();
  });

  it('handleDelete は isEdit = true のとき deleteMutation.mutate を呼ぶ（line 151-155）', async () => {
    const mockMutate = vi.fn();
    vi.mocked(useDeleteExamination).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
    } as ReturnType<typeof useDeleteExamination>);

    const { result } = renderHook(() => useExaminationForm('exam-001'));

    await act(async () => {
      result.current.handleDelete();
      await Promise.resolve();
    });

    expect(mockMutate).toHaveBeenCalledWith('exam-001', expect.objectContaining({ onSuccess: expect.any(Function) }));
  });

  it('handleDelete の onSuccess コールバックが toast.success を呼ぶ', async () => {
    const { toast } = await import('sonner');
    let capturedOnSuccess: (() => void) | undefined;
    const mockMutate = vi.fn((_id: string, opts: { onSuccess: () => void }) => {
      capturedOnSuccess = opts.onSuccess;
    });
    vi.mocked(useDeleteExamination).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
    } as ReturnType<typeof useDeleteExamination>);

    const { result } = renderHook(() => useExaminationForm('exam-001'));

    await act(async () => {
      result.current.handleDelete();
      await Promise.resolve();
    });

    act(() => { capturedOnSuccess?.(); });
    expect(toast.success).toHaveBeenCalledWith('検査記録を削除しました');
  });
});

// ─────────────────────────────────────────────────────────────

describe('useExaminationForm — formAction（useActionState コールバック）', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useSearchParams).mockReturnValue([new URLSearchParams(), vi.fn()]);
    vi.mocked(useGetPet).mockReturnValue({ data: null, isLoading: false, isError: false } as ReturnType<typeof useGetPet>);
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);
  });

  it('testTypeId と doctorId がない場合、バリデーションエラーを返す（line 92-97）', async () => {
    const { toast } = await import('sonner');
    const { result } = renderHook(() => useExaminationForm());

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(toast.error).toHaveBeenCalledWith('未入力の項目があります');
    expect(result.current.formState.success).toBe(false);
    expect(result.current.formState.fieldErrors).toMatchObject({
      testTypeId: expect.any(String),
      doctorId: expect.any(String),
    });
  });

  it('doctorId のみない場合、doctorId バリデーションエラーを返す', async () => {
    const { toast } = await import('sonner');
    const { result } = renderHook(() => useExaminationForm());

    act(() => { result.current.setFormData({ testTypeId: '5' }); });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(toast.error).toHaveBeenCalledWith('未入力の項目があります');
    expect(result.current.formState.fieldErrors?.doctorId).toBeDefined();
    expect(result.current.formState.fieldErrors?.testTypeId).toBeUndefined();
  });

  it('バリデーション通過 & 新規 & selectedPets なし → success: false（line 115）', async () => {
    // selectedPets = [] なので pet がない → early return
    const { result } = renderHook(() => useExaminationForm());

    act(() => {
      result.current.setFormData({ testTypeId: '5', doctorId: '3' });
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(result.current.formState.success).toBe(false);
  });

  it('バリデーション通過 & 新規 & selectedPets あり → createMutation.mutateAsync 呼ぶ（line 125）', async () => {
    const { useCreateExamination } = await import('../../api/create-examination');
    const mockMutateAsync = vi.fn().mockResolvedValue({});
    vi.mocked(useCreateExamination).mockReturnValue({ mutateAsync: mockMutateAsync } as ReturnType<typeof useCreateExamination>);

    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [{ id: '42', name: 'ポチ', ownerName: '田中', ownerId: '5', species: '犬', breed: '', birthday: '', gender: '男', weight: null, imageUrl: null, status: '生存', microchipNumber: null, insuranceNumber: null, insuranceExpiry: null, memo: null }],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);

    const { result } = renderHook(() => useExaminationForm());

    act(() => {
      result.current.setFormData({ testTypeId: '5', doctorId: '3' });
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(mockMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        pet_id: 42,
        exam_type_id: 5,
        doctor_id: 3,
      })
    );
    expect(result.current.formState.success).toBe(true);
  });

  it('mutateAsync が失敗した場合、toast.error を呼ぶ（line 129-130）', async () => {
    const { toast } = await import('sonner');
    const { useCreateExamination } = await import('../../api/create-examination');
    vi.mocked(useCreateExamination).mockReturnValue({
      mutateAsync: vi.fn().mockRejectedValue(new Error('API error')),
    } as ReturnType<typeof useCreateExamination>);

    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [{ id: '42', name: 'ポチ', ownerName: '田中', ownerId: '5', species: '犬', breed: '', birthday: '', gender: '男', weight: null, imageUrl: null, status: '生存', microchipNumber: null, insuranceNumber: null, insuranceExpiry: null, memo: null }],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);

    const { result } = renderHook(() => useExaminationForm());
    act(() => { result.current.setFormData({ testTypeId: '5', doctorId: '3' }); });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(toast.error).toHaveBeenCalledWith('保存に失敗しました');
    expect(result.current.formState.success).toBe(false);
  });

  it('編集モード & バリデーション通過 → updateMutation.mutateAsync 呼ぶ（line 112）', async () => {
    const { useUpdateExamination } = await import('../../api/update-examination');
    const mockMutateAsync = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({ mutateAsync: mockMutateAsync } as ReturnType<typeof useUpdateExamination>);

    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [{ id: '42', name: 'ポチ', ownerName: '田中', ownerId: '5', species: '犬', breed: '', birthday: '', gender: '男', weight: null, imageUrl: null, status: '生存', microchipNumber: null, insuranceNumber: null, insuranceExpiry: null, memo: null }],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);

    const { result } = renderHook(() => useExaminationForm('exam-001'));

    act(() => {
      result.current.setFormData({ testTypeId: '5', doctorId: '3', status: '完了', resultSummary: '正常' });
    });

    await act(async () => {
      await result.current.formAction(new FormData());
    });

    expect(mockMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        id: 'exam-001',
        req: expect.objectContaining({ status: 'completed', result_summary: '正常' }),
      })
    );
    expect(result.current.formState.success).toBe(true);
  });
});

// ─────────────────────────────────────────────────────────────

describe('useExaminationForm — setFormData（ローカルオーバーライド）', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useSearchParams).mockReturnValue([new URLSearchParams(), vi.fn()]);
    vi.mocked(useGetPet).mockReturnValue({ data: null, isLoading: false, isError: false } as ReturnType<typeof useGetPet>);
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);
  });

  it('setFormData でフィールドを更新できる', () => {
    const { result } = renderHook(() => useExaminationForm());
    act(() => { result.current.setFormData({ resultSummary: '異常なし' }); });
    expect(result.current.formData.resultSummary).toBe('異常なし');
  });

  it('setFormData を複数回呼んだ場合マージされる', () => {
    const { result } = renderHook(() => useExaminationForm());
    act(() => { result.current.setFormData({ resultSummary: '異常なし' }); });
    act(() => { result.current.setFormData({ machine: 'MRI' }); });
    expect(result.current.formData.resultSummary).toBe('異常なし');
    expect(result.current.formData.machine).toBe('MRI');
  });

  it('setFormData で doctorId を上書きできる', () => {
    vi.mocked(useSearchParams).mockReturnValue([new URLSearchParams('doctorId=1'), vi.fn()]);
    const { result } = renderHook(() => useExaminationForm());
    act(() => { result.current.setFormData({ doctorId: '99' }); });
    expect(result.current.formData.doctorId).toBe('99');
  });
});
