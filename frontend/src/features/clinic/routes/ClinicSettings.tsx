// React/Framework
import { useEffect } from "react";
import { useForm } from "react-hook-form";

// External
import { Building2, Save } from "lucide-react";

// Internal
import { PageLayout } from "@/components/shared/PageLayout";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";

// Shared Hooks
import { useClinicInfo } from "@/hooks/use-clinic-info";

// Types
import type { ClinicInfo } from "@/types";

export const ClinicSettings = () => {
  const { clinicInfo, updateClinicInfo, loading } = useClinicInfo();
  
  const { register, handleSubmit, reset, formState: { isDirty } } = useForm<ClinicInfo>({
    defaultValues: clinicInfo
  });

  useEffect(() => {
    if (!loading) {
      reset(clinicInfo);
    }
  }, [clinicInfo, loading, reset]);

  const onSubmit = (data: ClinicInfo) => {
    updateClinicInfo(data);
    reset(data); // Reset dirty state
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <PageLayout title="病院情報設定" icon={<Building2 className="h-6 w-6" />} maxWidth="max-w-4xl">
      <form onSubmit={handleSubmit(onSubmit)}>
        <Card>
          <CardHeader>
            <CardTitle>基本情報</CardTitle>
            <CardDescription>
              領収書や明細書、処方箋などに印字される病院情報です。
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="name">病院名 <span className="text-red-500">*</span></Label>
                <Input id="name" {...register("name", { required: true })} placeholder="例: わんにゃん動物病院" />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="branchName">支店名</Label>
                <Input id="branchName" {...register("branchName")} placeholder="例: 八王子院" />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="postalCode">郵便番号</Label>
                <Input id="postalCode" {...register("postalCode")} placeholder="例: 100-0001" />
              </div>
              <div className="grid gap-2 md:col-span-2">
                <Label htmlFor="address">住所</Label>
                <Input id="address" {...register("address")} placeholder="例: 東京都千代田区千代田1-1" />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="phoneNumber">電話番号</Label>
                <Input id="phoneNumber" {...register("phoneNumber")} placeholder="例: 03-1234-5678" />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="faxNumber">FAX番号</Label>
                <Input id="faxNumber" {...register("faxNumber")} placeholder="例: 03-1234-5679" />
              </div>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="registrationNumber">登録番号</Label>
              <Input id="registrationNumber" {...register("registrationNumber")} placeholder="例: 東京都獣医師会 第12345号" />
              <p className="text-sm text-gray-500">領収書などに記載される登録番号です。</p>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="directorName">院長名</Label>
              <Input id="directorName" {...register("directorName")} placeholder="例: 山田 太郎" />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="email">メールアドレス</Label>
                <Input id="email" type="email" {...register("email")} placeholder="info@example.com" />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="website">WebサイトURL</Label>
                <Input id="website" {...register("website")} placeholder="https://example.com" />
              </div>
            </div>
          </CardContent>
          <CardFooter className="bg-gray-50 border-t p-4 flex justify-end">
            <Button type="submit" disabled={!isDirty}>
              <Save className="mr-2 h-4 w-4" />
              設定を保存
            </Button>
          </CardFooter>
        </Card>
      </form>
    </PageLayout>
  );
}
