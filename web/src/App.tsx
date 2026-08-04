import type { JSX } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import { AppShell } from './app/AppShell'
import { CartPage } from './app/CartPage'
import { CheckoutPage } from './app/CheckoutPage'
import { HomePage } from './app/HomePage'
import { LoginPage } from './app/LoginPage'
import { OrderDetailPage } from './app/OrderDetailPage'
import { OrdersPage } from './app/OrdersPage'
import { ProductDetailPage } from './app/ProductDetailPage'
import { ProfilePage } from './app/ProfilePage'

export default function App(): JSX.Element {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<HomePage />} />
        <Route path="products/:id" element={<ProductDetailPage />} />
        <Route path="cart" element={<CartPage />} />
        <Route path="checkout" element={<CheckoutPage />} />
        <Route path="orders" element={<OrdersPage />} />
        <Route path="orders/:id" element={<OrderDetailPage />} />
        <Route path="profile" element={<ProfilePage />} />
        <Route path="login" element={<LoginPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}
