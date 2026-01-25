// Stub inventory hook for medical-records integration
// TODO: Implement real inventory management

interface ConsumeItem {
  inventoryId: string;
  quantity: number;
}

export function useInventory() {
  const consumeStock = (_items: ConsumeItem[]) => {
    // TODO: Implement actual stock consumption API call
  };

  return {
    consumeStock,
  };
}
