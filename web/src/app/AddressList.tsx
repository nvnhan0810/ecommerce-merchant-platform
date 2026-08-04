import { useState, type JSX } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  ListAddressesUseCase,
  DeleteAddressUseCase,
} from '@/modules/addresses/application/address-use-cases'
import { HttpAddressRepository } from '@/modules/addresses/infrastructure/http-address-repository'
import { formatFullAddress, type UserAddress } from '@/modules/addresses/domain/address'
import { AddressForm } from './AddressForm'
import styles from './AddressList.module.css'

const addressRepo = new HttpAddressRepository()
const listAddresses = new ListAddressesUseCase(addressRepo)
const deleteAddress = new DeleteAddressUseCase(addressRepo)

export function AddressList(): JSX.Element {
  const queryClient = useQueryClient()
  const [editing, setEditing] = useState<UserAddress | null>(null)
  const [isAdding, setIsAdding] = useState(false)

  const { data: addresses, isLoading } = useQuery({
    queryKey: ['user-addresses'],
    queryFn: () => listAddresses.execute(),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteAddress.execute(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user-addresses'] })
    },
  })

  function resetForm() {
    setIsAdding(false)
    setEditing(null)
  }

  if (isLoading) return <p>Đang tải địa chỉ...</p>

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1>Địa chỉ giao hàng</h1>
        {!isAdding && !editing && (
          <button type="button" className={styles.addBtn} onClick={() => setIsAdding(true)}>
            + Thêm địa chỉ
          </button>
        )}
      </div>

      {(isAdding || editing) && (
        <AddressForm
          key={editing?.id ?? 'new'}
          initial={editing}
          onCancel={resetForm}
          onSaved={resetForm}
        />
      )}

      {!isAdding && !editing && (
        <ul className={styles.list}>
          {addresses?.length === 0 && <p className={styles.empty}>Chưa có địa chỉ nào.</p>}
          {addresses?.map((addr) => (
            <li key={addr.id} className={styles.item}>
              <div className={styles.info}>
                <div className={styles.nameRow}>
                  <strong>{addr.wardName || addr.provinceName || 'Địa chỉ'}</strong>
                  {addr.isDefault && <span className={styles.badge}>Mặc định</span>}
                </div>
                <p>{formatFullAddress(addr)}</p>
                {addr.latitude != null && addr.longitude != null && (
                  <p className={styles.coordsMeta}>
                    {addr.latitude}, {addr.longitude}
                  </p>
                )}
              </div>
              <div className={styles.itemActions}>
                <button type="button" onClick={() => setEditing(addr)}>
                  Sửa
                </button>
                <button
                  type="button"
                  className={styles.deleteBtn}
                  onClick={() => deleteMutation.mutate(addr.id)}
                >
                  Xoá
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
