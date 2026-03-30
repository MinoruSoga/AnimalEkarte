        <form action={formAction} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="shift-type">シフト種別</Label>
            <Select value={form.shiftType} onValueChange={handleShiftTypeChange}>
              <SelectTrigger id="shift-type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>{SHIFT_TYPE_OPTIONS}</SelectContent>
            </Select>
          </div>

          <div className="space-y-1.5">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="start-time">開始時刻</Label>
                <Input
                  id="start-time"
                  name="startTime"
                  type="time"
                  value={form.startTime}
                  onChange={handleInputChange}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="end-time">終了時刻</Label>
                <Input
                  id="end-time"
                  name="endTime"
                  type="time"
                  value={form.endTime}
                  onChange={handleInputChange}
                />
              </div>
            </div>
            {timeError ? (
              <FormFieldError message={timeError} />
            ) : null}
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="note">メモ</Label>
            <Input
              id="note"
              name="note"
              placeholder="メモ（任意）"
              value={form.note}
              onChange={handleInputChange}
            />
          </div>

          <DialogFooter className="gap-2">
            {isEdit ? (
              <Button
                type="button"
                variant="destructive"
                size="sm"
                onClick={handleDelete}
                disabled={isPending || isDeletePending}
              >
                {isDeletePending ? "削除中..." : "削除"}
              </Button>
            ) : null}
            <Button type="button" variant="outline" onClick={onClose} disabled={isPending || isDeletePending}>
              キャンセル
            </Button>
            <SubmitButton size="sm">
              保存
            </SubmitButton>
          </DialogFooter>
        </form>
