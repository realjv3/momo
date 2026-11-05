import { ChevronLeft, ChevronRight } from 'lucide-react'

const PAGE_SIZES = [10, 25, 50, 100]

interface Props {
  page: number
  pageSize: number
  total: number
  onPage: (page: number) => void
  onPageSize: (size: number) => void
}

export function Pager({ page, pageSize, total, onPage, onPageSize }: Props) {
  const pages = Math.max(1, Math.ceil(total / pageSize))
  const from = total === 0 ? 0 : page * pageSize + 1
  const to = Math.min(total, (page + 1) * pageSize)

  const btn =
    'grid size-8 place-items-center rounded-lg bg-white/5 text-zinc-300 ring-1 ring-white/10 transition hover:bg-white/10 disabled:opacity-30 disabled:hover:bg-white/5'

  return (
    <div className="sticky bottom-0 z-10 mt-4 flex flex-wrap items-center justify-between gap-3 rounded-t-2xl border-t border-white/5 bg-zinc-950/85 px-2 py-3 text-xs text-zinc-400 backdrop-blur-xl">
      <span className="tabular-nums">
        {from}-{to} of {total.toLocaleString()}
      </span>

      <div className="flex items-center gap-3">
        <label className="flex items-center gap-2">
          Rows
          <select
            value={pageSize}
            onChange={(e) => onPageSize(Number(e.target.value))}
            className="rounded-lg bg-white/5 px-2 py-1.5 text-zinc-200 outline-none ring-1 ring-white/10 focus:ring-emerald-400/60"
          >
            {PAGE_SIZES.map((n) => (
              <option key={n} value={n} className="bg-zinc-900">
                {n}
              </option>
            ))}
          </select>
        </label>

        <div className="flex items-center gap-2">
          <button className={btn} disabled={page === 0} onClick={() => onPage(page - 1)} aria-label="Previous page">
            <ChevronLeft className="size-4" />
          </button>
          <span className="min-w-20 text-center tabular-nums">
            Page {page + 1} / {pages}
          </span>
          <button className={btn} disabled={page >= pages - 1} onClick={() => onPage(page + 1)} aria-label="Next page">
            <ChevronRight className="size-4" />
          </button>
        </div>
      </div>
    </div>
  )
}
