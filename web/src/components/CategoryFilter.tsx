import { ChevronDown } from 'lucide-react'

interface CategoryFilterProps {
  value: string
  onChange: (value: string) => void
  categories: string[]
}

export function CategoryFilter({ value, onChange, categories }: CategoryFilterProps) {
  return (
    <div className="select-wrapper">
      <select
        className="filter-select"
        value={value}
        onChange={e => onChange(e.target.value)}
      >
        <option value="">全部类别</option>
        {categories.map(cat => (
          <option key={cat} value={cat}>
            {cat.charAt(0).toUpperCase() + cat.slice(1)}
          </option>
        ))}
      </select>
      <ChevronDown size={14} />
    </div>
  )
}
