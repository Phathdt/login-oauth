import { useQuery } from '@tanstack/react-query'
import { Navbar } from '@/components/navbar'
import { ProductCard } from '@/components/product-card'
import { getProducts } from '@/services/product-service'

function ProductsGridSkeleton() {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {Array.from({ length: 6 }).map((_, i) => (
        <div
          key={i}
          className="h-40 rounded-lg bg-muted animate-pulse"
        />
      ))}
    </div>
  )
}

export default function ProductsPage() {
  const { data: products, isLoading, isError } = useQuery({
    queryKey: ['products'],
    queryFn: getProducts,
  })

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="container mx-auto px-4 py-8">
        <h1 className="text-2xl font-bold mb-6">Products</h1>

        {isLoading && <ProductsGridSkeleton />}

        {isError && (
          <div className="text-center py-12">
            <p className="text-destructive">
              Failed to load products. Please try again.
            </p>
          </div>
        )}

        {products && products.length === 0 && (
          <p className="text-muted-foreground text-center py-12">
            No products available.
          </p>
        )}

        {products && products.length > 0 && (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {products.map((product) => (
              <ProductCard key={product.id} product={product} />
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
