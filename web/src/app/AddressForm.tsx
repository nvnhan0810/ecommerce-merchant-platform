import { useEffect, useState, type FormEvent, type JSX } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  CreateAddressUseCase,
  UpdateAddressUseCase,
  ListCountriesUseCase,
  ListProvincesUseCase,
  ListWardsUseCase,
} from '@/modules/addresses/application/address-use-cases'
import {
  HttpAddressRepository,
  HttpGeoRepository,
} from '@/modules/addresses/infrastructure/http-address-repository'
import type { AddressInput, UserAddress } from '@/modules/addresses/domain/address'
import { SearchableSelect } from '@/shared/SearchableSelect'
import styles from './AddressList.module.css'

const addressRepo = new HttpAddressRepository()
const geoRepo = new HttpGeoRepository()
const createAddress = new CreateAddressUseCase(addressRepo)
const updateAddress = new UpdateAddressUseCase(addressRepo)
const listCountries = new ListCountriesUseCase(geoRepo)
const listProvinces = new ListProvincesUseCase(geoRepo)
const listWards = new ListWardsUseCase(geoRepo)

const DEFAULT_COUNTRY = 'VN'

type Props = {
  initial?: UserAddress | null
  onCancel: () => void
  onSaved: (address: UserAddress) => void
}

export function AddressForm({ initial, onCancel, onSaved }: Props): JSX.Element {
  const queryClient = useQueryClient()
  const [addressLine, setAddressLine] = useState(initial?.addressLine ?? '')
  const [countryCode, setCountryCode] = useState(initial?.countryCode || DEFAULT_COUNTRY)
  const [provinceCode, setProvinceCode] = useState(initial?.provinceCode || '')
  const [wardCode, setWardCode] = useState(initial?.wardCode || '')
  const [latitude, setLatitude] = useState(initial?.latitude != null ? String(initial.latitude) : '')
  const [longitude, setLongitude] = useState(initial?.longitude != null ? String(initial.longitude) : '')
  const [isDefault, setIsDefault] = useState(initial?.isDefault ?? false)

  const { data: countries } = useQuery({
    queryKey: ['geo-countries'],
    queryFn: () => listCountries.execute(),
  })

  const { data: provinces } = useQuery({
    queryKey: ['geo-provinces', countryCode],
    queryFn: () => listProvinces.execute(countryCode || DEFAULT_COUNTRY),
    enabled: Boolean(countryCode),
  })

  const { data: wards } = useQuery({
    queryKey: ['geo-wards', provinceCode],
    queryFn: () => listWards.execute(provinceCode),
    enabled: Boolean(provinceCode),
  })

  useEffect(() => {
    if (!countries?.length) return
    const def = countries.find((c) => c.isDefault) ?? countries.find((c) => c.code === DEFAULT_COUNTRY)
    if (def && !countryCode) setCountryCode(def.code)
  }, [countries, countryCode])

  const createMutation = useMutation({
    mutationFn: (input: AddressInput) => createAddress.execute(input),
    onSuccess: async (addr) => {
      await queryClient.invalidateQueries({ queryKey: ['user-addresses'] })
      onSaved(addr)
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input: AddressInput }) => updateAddress.execute(id, input),
    onSuccess: async (addr) => {
      await queryClient.invalidateQueries({ queryKey: ['user-addresses'] })
      onSaved(addr)
    },
  })

  function onCountryChange(code: string) {
    setCountryCode(code || DEFAULT_COUNTRY)
    setProvinceCode('')
    setWardCode('')
    setLatitude('')
    setLongitude('')
  }

  function onProvinceChange(code: string) {
    setProvinceCode(code)
    setWardCode('')
    setLatitude('')
    setLongitude('')
  }

  function onWardChange(code: string) {
    setWardCode(code)
    const ward = wards?.find((w) => w.code === code)
    if (ward?.latitude != null) setLatitude(String(ward.latitude))
    if (ward?.longitude != null) setLongitude(String(ward.longitude))
  }

  function parseCoord(raw: string): number | null {
    const t = raw.trim()
    if (!t) return null
    const n = Number(t)
    return Number.isFinite(n) ? n : null
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    e.stopPropagation()
    const input: AddressInput = {
      addressLine,
      countryCode: countryCode || DEFAULT_COUNTRY,
      provinceCode,
      wardCode,
      latitude: parseCoord(latitude),
      longitude: parseCoord(longitude),
      isDefault,
    }
    if (initial?.id) {
      await updateMutation.mutateAsync({ id: initial.id, input })
    } else {
      await createMutation.mutateAsync(input)
    }
  }

  const formError =
    (createMutation.error as Error | null)?.message ||
    (updateMutation.error as Error | null)?.message ||
    ''

  return (
    <form className={styles.form} onSubmit={(e) => void onSubmit(e)}>
      <label>
        Quốc gia
        <SearchableSelect
          aria-label="Quốc gia"
          options={(
            countries ?? [{ code: DEFAULT_COUNTRY, name: 'Việt Nam', nameEn: 'Vietnam', isDefault: true }]
          ).map((c) => ({ value: c.code, label: c.name }))}
          value={countryCode}
          onChange={onCountryChange}
          placeholder="Tìm quốc gia…"
          required
          isClearable={false}
        />
      </label>
      <label>
        Tỉnh / Thành phố
        <SearchableSelect
          aria-label="Tỉnh / Thành phố"
          options={(provinces ?? []).map((p) => ({ value: p.code, label: p.name }))}
          value={provinceCode}
          onChange={onProvinceChange}
          placeholder="Tìm tỉnh/thành…"
          required
        />
      </label>
      <label>
        Phường / Xã
        <SearchableSelect
          aria-label="Phường / Xã"
          options={(wards ?? []).map((w) => ({ value: w.code, label: w.name }))}
          value={wardCode}
          onChange={onWardChange}
          placeholder={provinceCode ? 'Tìm phường/xã…' : 'Chọn tỉnh/thành trước'}
          required
          isDisabled={!provinceCode}
        />
      </label>
      <label>
        Địa chỉ chi tiết
        <textarea
          value={addressLine}
          onChange={(e) => setAddressLine(e.target.value)}
          required
          rows={2}
          placeholder="Số nhà, tên đường…"
        />
      </label>
      <div className={styles.coords}>
        <label>
          Vĩ độ (lat)
          <input
            value={latitude}
            onChange={(e) => setLatitude(e.target.value)}
            inputMode="decimal"
            placeholder="Tự điền khi chọn phường/xã"
          />
        </label>
        <label>
          Kinh độ (long)
          <input
            value={longitude}
            onChange={(e) => setLongitude(e.target.value)}
            inputMode="decimal"
            placeholder="Tự điền khi chọn phường/xã"
          />
        </label>
      </div>
      <label className={styles.checkbox}>
        <input type="checkbox" checked={isDefault} onChange={(e) => setIsDefault(e.target.checked)} />
        Đặt làm địa chỉ mặc định
      </label>
      {formError && <p className={styles.error}>{formError}</p>}
      <div className={styles.actions}>
        <button type="button" className={styles.cancelBtn} onClick={onCancel}>
          Huỷ
        </button>
        <button
          type="submit"
          className={styles.saveBtn}
          disabled={createMutation.isPending || updateMutation.isPending}
        >
          Lưu
        </button>
      </div>
    </form>
  )
}
