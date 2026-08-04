import { createContext, useContext, useMemo, useSyncExternalStore, type JSX, type ReactNode } from 'react'
import { Cart, CartItem } from '../domain/cart'

const KEY = 'ecomerce.user.cart'

type StoredItem = {
  productId: string
  merchantId: string
  name: string
  unitPriceCents: number
  currency: string
  quantity: number
  imageUrl?: string
}

function parseCart(raw: string | null): Cart {
  if (!raw) return new Cart([])
  try {
    const data = JSON.parse(raw) as StoredItem[]
    return new Cart(
      data.map(
        (item) =>
          new CartItem(
            item.productId,
            item.merchantId,
            item.name,
            item.unitPriceCents,
            item.currency || 'VND',
            item.quantity,
            item.imageUrl ?? '',
          ),
      ),
    )
  } catch {
    return new Cart([])
  }
}

let cachedRaw: string | null | undefined = undefined
let cachedCart: Cart = new Cart([])

function readCart(): Cart {
  const raw = localStorage.getItem(KEY)
  if (raw === cachedRaw) {
    return cachedCart
  }
  cachedRaw = raw
  cachedCart = parseCart(raw)
  return cachedCart
}

function writeCart(cart: Cart): void {
  const data: StoredItem[] = cart.items.map((item) => ({
    productId: item.productId,
    merchantId: item.merchantId,
    name: item.name,
    unitPriceCents: item.unitPriceCents,
    currency: item.currency,
    quantity: item.quantity,
    imageUrl: item.imageUrl,
  }))
  const raw = JSON.stringify(data)
  localStorage.setItem(KEY, raw)
  cachedRaw = raw
  cachedCart = cart
}

const listeners = new Set<() => void>()
function subscribe(cb: () => void): () => void {
  listeners.add(cb)
  return () => listeners.delete(cb)
}
function emit(): void {
  listeners.forEach((cb) => cb())
}
function getSnapshot(): Cart {
  return readCart()
}
function getServerSnapshot(): Cart {
  return cachedCart
}

type CartContextValue = {
  cart: Cart
  addItem: (item: CartItem) => void
  setQuantity: (productId: string, quantity: number) => void
  removeItem: (productId: string) => void
  clear: () => void
}

const CartContext = createContext<CartContextValue | null>(null)

export function CartProvider({ children }: { children: ReactNode }): JSX.Element {
  const cart = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot)

  const value = useMemo<CartContextValue>(
    () => ({
      cart,
      addItem(item) {
        writeCart(readCart().add(item))
        emit()
      },
      setQuantity(productId, quantity) {
        writeCart(readCart().setQuantity(productId, quantity))
        emit()
      },
      removeItem(productId) {
        writeCart(readCart().remove(productId))
        emit()
      },
      clear() {
        writeCart(readCart().clear())
        emit()
      },
    }),
    [cart],
  )

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>
}

export function useCart(): CartContextValue {
  const ctx = useContext(CartContext)
  if (!ctx) {
    throw new Error('useCart must be used within CartProvider')
  }
  return ctx
}
