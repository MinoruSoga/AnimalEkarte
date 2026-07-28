-- BUG-466: inventory_items.quantity を非負に制約し、負数在庫の永続化を DB 層でも拒否する。
ALTER TABLE inventory_items
  ADD CONSTRAINT chk_inventory_items_quantity_non_negative CHECK (quantity >= 0);
