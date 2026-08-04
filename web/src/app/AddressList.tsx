import { useState, type FormEvent, type JSX } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  ListAddressesUseCase,
  CreateAddressUseCase,
  UpdateAddressUseCase,
  DeleteAddressUseCase
} from '@/modules/addresses/application/address-use-cases'
import { HttpAddressRepository } from '@/modules/addresses/infrastructure/http-address-repository'
import type { UserAddress, AddressInput } from '@/modules/addresses/domain/address'
import styles from './AddressList.module.css'

const repo = new HttpAddressRepository()
const listAddresses = new ListAddressesUseCase(repo)
const createAddress = new CreateAddressUseCase(repo)
const updateAddress = new UpdateAddressUseCase(repo)
const deleteAddress = new DeleteAddressUseCase(repo)

export function AddressList(): JSX.Element {
  const queryClient = useQueryClient()
  const [editingId, setEditingId] = useState<string | null>(null)
  const [isAdding, setIsAdding] = useState(false)
  
  const [recipientName, setRecipientName] = useState('')
  const [phoneNumber, setPhoneNumber] = useState('')
  const [addressLine, setAddressLine] = useState('')
  const [isDefault, setIsDefault] = useState(false)

  const { data: addresses, isLoading } = useQuery({
    queryKey: ['user-addresses'],
    queryFn: () => listAddresses.execute(),
  })

  const createMutation = useMutation({
    mutationFn: (input: AddressInput) => createAddress.execute(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user-addresses'] })
      resetForm()
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input: AddressInput }) => updateAddress.execute(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user-addresses'] })
      resetForm()
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteAddress.execute(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user-addresses'] })
    },
  })

  function resetForm() {
    setIsAdding(false)
    setEditingId(null)
    setRecipientName('')
    setPhoneNumber('')
    setAddressLine('')
    setIsDefault(false)
  }

  function startEdit(addr: UserAddress) {
    setEditingId(addr.id)
    setIsAdding(false)
    setRecipientName(addr.recipientName)
    setPhoneNumber(addr.phoneNumber)
    setAddressLine(addr.addressLine)
    setIsDefault(addr.isDefault)
  }

  function startAdd() {
    resetForm()
    setIsAdding(true)
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    const input = { recipientName, phoneNumber, addressLine, isDefault }
    if (editingId) {
      await updateMutation.mutateAsync({ id: editingId, input })
    } else {
      await createMutation.mutateAsync(input)
    }
  }

  if (isLoading) return <p>Đang tải địa chỉ...</p>

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h2>Sổ địa chỉ</h2>
        {!isAdding && !editingId && (
          <button type="button" className={styles.addBtn} onClick={startAdd}>
            + Thêm địa chỉ
          </button>
        )}
      </div>

      {(isAdding || editingId) && (
        <form className={styles.form} onSubmit={(e) => void onSubmit(e)}>
          <label>
            Người nhận
            <input value={recipientName} onChange={(e) => setRecipientName(e.target.value)} required />
          </label>
          <label>
            Số điện thoại
            <input value={phoneNumber} onChange={(e) => setPhoneNumber(e.target.value)} required />
          </label>
          <label>
            Địa chỉ
            <textarea value={addressLine} onChange={(e) => setAddressLine(e.target.value)} required rows={2} />
          </label>
          <label className={styles.checkbox}>
            <input type="checkbox" checked={isDefault} onChange={(e) => setIsDefault(e.target.checked)} />
            Đặt làm địa chỉ mặc định
          </label>
          <div className={styles.actions}>
            <button type="button" className={styles.cancelBtn} onClick={resetForm}>
              Huỷ
            </button>
            <button type="submit" className={styles.saveBtn} disabled={createMutation.isPending || updateMutation.isPending}>
              Lưu
            </button>
          </div>
        </form>
      )}

      {!isAdding && !editingId && (
        <ul className={styles.list}>
          {addresses?.length === 0 && <p className={styles.empty}>Chưa có địa chỉ nào.</p>}
          {addresses?.map((addr) => (
            <li key={addr.id} className={styles.item}>
              <div className={styles.info}>
                <div className={styles.nameRow}>
                  <strong>{addr.recipientName}</strong>
                  {addr.isDefault && <span className={styles.badge}>Mặc định</span>}
                </div>
                <p>{addr.phoneNumber}</p>
                <p>{addr.addressLine}</p>
              </div>
              <div className={styles.itemActions}>
                <button type="button" onClick={() => startEdit(addr)}>Sửa</button>
                <button type="button" className={styles.deleteBtn} onClick={() => deleteMutation.mutate(addr.id)}>Xoá</button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
