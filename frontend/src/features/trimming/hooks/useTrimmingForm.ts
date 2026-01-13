import { useState, useEffect } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router-dom";
import { toast } from "sonner";
import { MOCK_PETS } from "../../../lib/constants";
import { usePetSelection } from "../../pets/hooks/usePetSelection";

export interface TrimmingFormData {
  styleRequest: string;
  memo: string;
  eggs: string;
  parts: {
    nail: boolean;
    analGland: boolean;
    eye: boolean;
    ear: boolean;
    skin: boolean;
    oral: boolean;
  };
  styleImage: File | null;
  bw: string;
  bwUnit: "Kg" | "g";
  bt: string;
  usedShampoo: string;
  usedRibbon: string;
  treatment: string;
  medicine: string;
  charge: string;
  finalCheck: string;
  remarks: string;
  completedImage: File | null;

  // New fields for Master Selection
  courseId: string;
  optionIds: string[];
}

export function useTrimmingForm(id?: string) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const mode = id ? "edit" : "new";

  const petSelection = usePetSelection();
  const { setSelectedPets } = petSelection;

  const [formData, setFormData] = useState<TrimmingFormData>({
    styleRequest: "",
    memo: "",
    eggs: "",
    parts: {
      nail: false,
      analGland: false,
      eye: false,
      ear: false,
      skin: false,
      oral: false,
    },
    styleImage: null,
    bw: "",
    bwUnit: "Kg",
    bt: "",
    usedShampoo: "",
    usedRibbon: "",
    treatment: "",
    medicine: "",
    charge: "",
    finalCheck: "",
    remarks: "",
    completedImage: null,
    
    courseId: "",
    optionIds: [],
  });

  const [styleImagePreview, setStyleImagePreview] = useState<string | null>(null);
  const [completedImagePreview, setCompletedImagePreview] = useState<string | null>(null);

  useEffect(() => {
    if (mode === "edit") {
        // Mock Data Loading
        // Find pet based on ID - assuming ID might map to pet for this mock
        // For ID "1", let's use Iris. For "2", Max.
        let targetPetName = "";
        if (id === "1") targetPetName = "Iris";
        if (id === "2") targetPetName = "Max";

        if (targetPetName) {
            const pet = MOCK_PETS.find(p => p.name.includes(targetPetName));
            if (pet) {
                setSelectedPets([pet]);
            }
        }

        if (id === "1") {
            setFormData(prev => ({
                ...prev,
                styleRequest: "サマーカット希望。短めにカットしてください。",
                memo: "毛玉が多いので丁寧にブラッシングをお願いします。",
                eggs: "なし",
                parts: {
                  nail: true,
                  analGland: true,
                  eye: true,
                  ear: true,
                  skin: false,
                  oral: false,
                },
                bw: "10.5",
                bwUnit: "Kg",
                bt: "38.5",
                usedShampoo: "低刺激シャンプー",
                usedRibbon: "ピンクリボン",
                treatment: "保湿トリートメント",
                medicine: "なし",
                charge: "8,000円",
                finalCheck: "完了",
                remarks: "次回は1ヶ月後を予定",
            }));
        }
    } else {
        if (petId) {
            const foundPet = MOCK_PETS.find(p => p.id === petId);
            if (foundPet) {
                setSelectedPets([foundPet]);
            } else {
                navigate("/trimming/select-pet");
            }
        }
    }
  }, [id, mode, petId, setSelectedPets, navigate]);

  const handleStyleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setFormData({ ...formData, styleImage: file });
      const reader = new FileReader();
      reader.onloadend = () => {
        setStyleImagePreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleCompletedImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setFormData({ ...formData, completedImage: file });
      const reader = new FileReader();
      reader.onloadend = () => {
        setCompletedImagePreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const removeStyleImage = () => {
    setFormData({ ...formData, styleImage: null });
    setStyleImagePreview(null);
  };

  const removeCompletedImage = () => {
    setFormData({ ...formData, completedImage: null });
    setCompletedImagePreview(null);
  };

  const handleSave = () => {
    toast.success(mode === "edit" ? "トリミング情報を更新しました" : "トリミング情報を登録しました");
    setTimeout(() => {
        if (location.state?.from) {
            navigate(location.state.from);
        } else {
            navigate("/trimming");
        }
    }, 500);
  };

  return {
    mode,
    formData,
    setFormData,
    styleImagePreview,
    completedImagePreview,
    petSelection,
    handleStyleImageChange,
    handleCompletedImageChange,
    removeStyleImage,
    removeCompletedImage,
    handleSave,
  };
}
