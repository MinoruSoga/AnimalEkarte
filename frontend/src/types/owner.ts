import { Pet } from "./index";

export interface Owner {
  id: string;
  name: string;
  name_kana: string;
  phone: string;
  email: string;
  address: string;
  notes: string;
  created_at: string;
  updated_at: string;
  pets?: Pet[];
}

export interface CreateOwnerRequest {
  name: string;
  name_kana?: string;
  phone?: string;
  email?: string;
  address?: string;
  notes?: string;
}

export interface UpdateOwnerRequest {
  name?: string;
  name_kana?: string;
  phone?: string;
  email?: string;
  address?: string;
  notes?: string;
}
