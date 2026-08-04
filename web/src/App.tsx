import type { JSX } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import { AppShell } from './app/AppShell'
import { CartPage } from './app/CartPage'
import { CheckoutPage } from './app/CheckoutPage'
import { HomePage } from './app/HomePage'
import { LoginPage } from './app/LoginPage'
import { MerchantPage } from './app/MerchantPage'
import { OrderDetailPage } from './app/OrderDetailPage'
import { OrdersPage } from './app/OrdersPage'
import { PaymentResultPage } from './app/PaymentResultPage'
import { ProductDetailPage } from './app/ProductDetailPage'
import { ProfilePage } from './app/ProfilePage'
import { AddressesPage } from './app/AddressesPage'
import { UserGuidePage } from './app/UserGuidePage'

export default function App(): JSX.Element {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<HomePage />} />
        <Route path="products/:id" element={<ProductDetailPage />} />
        <Route path="merchants/:id" element={<MerchantPage />} />
        <Route path="cart" element={<CartPage />} />
        <Route path="checkout" element={<CheckoutPage />} />
        <Route path="orders" element={<OrdersPage />} />
        <Route path="orders/payment/result" element={<PaymentResultPage />} />
        <Route path="orders/:id" element={<OrderDetailPage />} />
        <Route path="profile" element={<ProfilePage />} />
        <Route path="addresses" element={<AddressesPage />} />
        <Route path="huong-dan" element={<UserGuidePage />} />
        <Route path="login" element={<LoginPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}
