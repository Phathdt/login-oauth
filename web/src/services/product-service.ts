import { apiClient } from '@/lib/axios-client'

export interface Product {
  id: string
  name: string
  description: string
  price: number
}

export async function getProducts(): Promise<Product[]> {
  const { data } = await apiClient.get<Product[]>('/api/products')
  return data
}
