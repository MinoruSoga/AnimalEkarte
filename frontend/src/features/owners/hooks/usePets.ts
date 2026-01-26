import { useState, useEffect, useMemo } from "react";
import { getOwners } from "../api";
import { Pet } from "@/types";

export function usePets(searchTerm: string) {
  const [data, setData] = useState<Pet[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      setIsLoading(true);
      try {
        const owners = await getOwners();
        const flatPets: Pet[] = [];

        owners.forEach((owner) => {
          if (!owner.pets || owner.pets.length === 0) {
            // Owner without pets
            flatPets.push({
              id: `owner-${owner.id}`,
              ownerId: owner.id,
              ownerName: owner.name,
              name: "-",
              species: "-",
              petNumber: "-",
              status: undefined, // Or handle as needed
            } as Pet);
          } else {
            // Owner with pets
            owner.pets.forEach((pet) => {
              flatPets.push({
                ...pet,
                ownerId: owner.id,
                ownerName: owner.name,
                // Ensure properties exist if backend returns null
                name: pet.name || "-", 
                species: pet.species || "-",
              } as unknown as Pet); // Cast if backend Pet type differs slightly from frontend Pet type
            });
          }
        });

        setData(flatPets);
      } catch (err) {
        setError(err as Error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, []);

  const filteredPets = useMemo(() => {
    if (!searchTerm) return data;
    const lowerTerm = searchTerm.toLowerCase();
    return data.filter((pet) => {
      return (
        pet.ownerName.toLowerCase().includes(lowerTerm) ||
        pet.ownerId.includes(lowerTerm) ||
        pet.name.toLowerCase().includes(lowerTerm) ||
        (pet.species && pet.species.toLowerCase().includes(lowerTerm))
      );
    });
  }, [data, searchTerm]);

  return { data: filteredPets, isLoading, error };
}
