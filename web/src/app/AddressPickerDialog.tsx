import { useEffect, useRef, useState, type JSX } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  DeleteAddressUseCase,
  ListAddressesUseCase,
} from '@/modules/addresses/application/address-use-cases'
import { HttpAddressRepository } from '@/modules/addresses/infrastructure/http-address-repository'
import { formatFullAddress, type UserAddress } from '@/modules/addresses/domain/address'
import { AddressForm } from './AddressForm'
import styles from './AddressPickerDialog.module.css'

const listAddresses = new ListAddressesUseCase(new HttpAddressRepository())
const deleteAddress = new DeleteAddressUseCase(new HttpAddressRepository())

type Mode = 'list' | 'create' | 'edit'

type Props = {
  open: boolean
  selectedId: string
  onClose: () => void
  onSelect: (addressId: string) => void
}

export function AddressPickerDialog({ open, selectedId, onClose, onSelect }: Props): JSX.Element | null {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const queryClient = useQueryClient()
  const [mode, setMode] = useState<Mode>('list')
  const [editing, setEditing] = useState<UserAddress | null>(null)
  const [draftId, setDraftId] = useState(selectedId)

  const { data: addresses, isLoading } = useQuery({
    queryKey: ['user-addresses'],
    queryFn: () => listAddresses.execute(),
    enabled: open,
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteAddress.execute(id),
    onSuccess: async (_, id) => {
      await queryClient.invalidateQueries({ queryKey: ['user-addresses'] })
      if (draftId === id) setDraftId('')
    },
  })

  useEffect(() => {
    const el = dialogRef.current
    if (!el) return
    if (open && !el.open) {
      setMode('list')
      setEditing(null)
      setDraftId(selectedId)
      el.showModal()
    } else if (!open && el.open) {
      el.close()
    }
  }, [open, selectedId])

  useEffect(() => {
    const el = dialogRef.current
    if (!el) return
    function onCloseEvent() {
      onClose()
    }
    el.addEventListener('close', onCloseEvent)
    return () => el.removeEventListener('close', onCloseEvent)
  }, [onClose])

  if (!open) return null

  function startCreate() {
    setEditing(null)
    setMode('create')
  }

  function startEdit(addr: UserAddress) {
    setEditing(addr)
    setMode('edit')
  }

  function backToList() {
    setMode('list')
    setEditing(null)
  }

  function handleSaved(addr: UserAddress) {
    setDraftId(addr.id)
    setMode('list')
    setEditing(null)
  }

  function confirmSelect() {
    if (!draftId) return
    onSelect(draftId)
    onClose()
  }

  const title = mode === 'create' ? 'Thêm địa chỉ' : mode === 'edit' ? 'Sửa địa chỉ' : 'Chọn địa chỉ giao hàng'

  return (
    <dialog
      ref={dialogRef}
      className={styles.dialog}
      onCancel={(e) => {
        e.preventDefault()
        if (mode !== 'list') backToList()
        else onClose()
      }}
    >
      <div className={styles.header}>
        <h2>{title}</h2>
        <button type="button" className={styles.closeBtn} onClick={onClose} aria-label="Đóng">
          ×
        </button>
      </div>

      <div className={styles.body}>
        {mode === 'list' && (
          <>
            <div className={styles.toolbar}>
              <button type="button" className={styles.primaryBtn} onClick={startCreate}>
                + Thêm địa chỉ mới
              </button>
            </div>
            {isLoading && <p className={styles.hint}>Đang tải…</p>}
            {!isLoading && addresses?.length === 0 && (
              <p className={styles.hint}>Chưa có địa chỉ nào. Hãy thêm địa chỉ mới.</p>
            )}
            <ul className={styles.list}>
              {addresses?.map((addr) => (
                <li key={addr.id}>
                  <label className={`${styles.option} ${draftId === addr.id ? styles.optionSelected : ''}`}>
                    <input
                      type="radio"
                      name="checkout-address"
                      checked={draftId === addr.id}
                      onChange={() => setDraftId(addr.id)}
                    />
                    <div className={styles.optionBody}>
                      <div className={styles.optionTitle}>
                        <strong>
                          {[addr.wardName, addr.provinceName].filter(Boolean).join(', ') || 'Địa chỉ'}
                        </strong>
                        {addr.isDefault && <span className={styles.badge}>Mặc định</span>}
                      </div>
                      <p>{formatFullAddress(addr)}</p>
                    </div>
                  </label>
                  <div className={styles.optionActions}>
                    <button type="button" onClick={() => startEdit(addr)}>
                      Sửa
                    </button>
                    <button
                      type="button"
                      className={styles.deleteBtn}
                      onClick={() => {
                        if (window.confirm('Xoá địa chỉ này?')) deleteMutation.mutate(addr.id)
                      }}
                    >
                      Xoá
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          </>
        )}

        {(mode === 'create' || mode === 'edit') && (
          <AddressForm
            key={editing?.id ?? 'new'}
            initial={editing}
            onCancel={backToList}
            onSaved={handleSaved}
          />
        )}
      </div>

      {mode === 'list' && (
        <div className={styles.footer}>
          <button type="button" className={styles.ghostBtn} onClick={onClose}>
            Đóng
          </button>
          <button type="button" className={styles.primaryBtn} disabled={!draftId} onClick={confirmSelect}>
            Dùng địa chỉ này
          </button>
        </div>
      )}
    </dialog>
  )
}
