// Stub inventory hook for medical-records integration
// TODO: Implement real inventory management

interface ConsumeItem {
  inventoryId: string;
  quantity: number;
}

export function useInventory() {
  const consumeStock = (items: ConsumeItem[]) => {
    // Mock: Log consumption for now
    console.log("Stock consumption requested:", items);
  };

  return {
    consumeStock,
  };
}
