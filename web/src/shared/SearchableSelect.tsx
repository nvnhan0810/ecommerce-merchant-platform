import { useId, useMemo, type JSX } from 'react'
import Select, { type SingleValue, type StylesConfig, type FilterOptionOption } from 'react-select'

export type SearchableOption = {
  value: string
  label: string
}

type Props = {
  options: SearchableOption[]
  value: string
  onChange: (value: string) => void
  placeholder?: string
  isDisabled?: boolean
  isClearable?: boolean
  required?: boolean
  inputId?: string
  'aria-label'?: string
}

function stripDiacritics(s: string): string {
  return s
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/đ/g, 'd')
    .replace(/Đ/g, 'D')
    .toLowerCase()
}

function filterOption(option: FilterOptionOption<SearchableOption>, rawInput: string): boolean {
  const q = stripDiacritics(rawInput.trim())
  if (!q) return true
  const hay = `${stripDiacritics(option.label)} ${stripDiacritics(option.value)}`
  return hay.includes(q)
}

const selectStyles: StylesConfig<SearchableOption, false> = {
  control: (base, state) => ({
    ...base,
    minHeight: 38,
    borderColor: state.isFocused ? '#0f766e' : '#cbd5e1',
    borderRadius: 6,
    boxShadow: state.isFocused ? '0 0 0 1px #0f766e' : 'none',
    backgroundColor: state.isDisabled ? '#f1f5f9' : '#fff',
    fontWeight: 400,
    '&:hover': {
      borderColor: state.isFocused ? '#0f766e' : '#94a3b8',
    },
  }),
  valueContainer: (base) => ({
    ...base,
    padding: '2px 8px',
  }),
  placeholder: (base) => ({
    ...base,
    color: '#94a3b8',
  }),
  menu: (base) => ({
    ...base,
    zIndex: 20,
    borderRadius: 8,
    overflow: 'hidden',
    boxShadow: '0 8px 24px rgba(15, 23, 42, 0.12)',
  }),
  option: (base, state) => ({
    ...base,
    backgroundColor: state.isSelected
      ? '#0f766e'
      : state.isFocused
        ? '#f0fdfa'
        : '#fff',
    color: state.isSelected ? '#fff' : '#0f172a',
    cursor: 'pointer',
  }),
  indicatorSeparator: () => ({ display: 'none' }),
  dropdownIndicator: (base) => ({
    ...base,
    color: '#64748b',
    padding: 6,
  }),
  clearIndicator: (base) => ({
    ...base,
    padding: 6,
  }),
}

export function SearchableSelect({
  options,
  value,
  onChange,
  placeholder = 'Tìm và chọn…',
  isDisabled = false,
  isClearable = true,
  required = false,
  inputId,
  'aria-label': ariaLabel,
}: Props): JSX.Element {
  const autoId = useId()
  const selected = useMemo(
    () => options.find((o) => o.value === value) ?? null,
    [options, value],
  )

  return (
    <div style={{ position: 'relative' }}>
      <Select<SearchableOption, false>
        inputId={inputId ?? autoId}
        aria-label={ariaLabel}
        options={options}
        value={selected}
        onChange={(opt: SingleValue<SearchableOption>) => onChange(opt?.value ?? '')}
        placeholder={placeholder}
        isDisabled={isDisabled}
        isClearable={isClearable}
        isSearchable
        filterOption={filterOption}
        noOptionsMessage={() => 'Không có kết quả'}
        loadingMessage={() => 'Đang tải…'}
        styles={selectStyles}
        classNamePrefix="search-select"
      />
      {/* Native required validation for form submit */}
      {required && (
        <input
          tabIndex={-1}
          aria-hidden
          required
          value={value}
          onChange={() => undefined}
          style={{
            opacity: 0,
            height: 0,
            width: 0,
            position: 'absolute',
            pointerEvents: 'none',
          }}
        />
      )}
    </div>
  )
}
