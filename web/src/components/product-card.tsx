import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { Product } from '@/services/product-service'

interface ProductCardProps {
  product: Product
}

export function ProductCard({ product }: ProductCardProps) {
  return (
    <Card className="flex flex-col h-full">
      <CardHeader>
        <CardTitle className="text-lg">{product.name}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-2 flex-1">
        <p className="text-muted-foreground text-sm flex-1">
          {product.description}
        </p>
        <span className="inline-block mt-2 text-sm font-semibold bg-primary text-primary-foreground px-3 py-1 rounded-full w-fit">
          ${product.price.toFixed(2)}
        </span>
      </CardContent>
    </Card>
  )
}
