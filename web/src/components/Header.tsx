import { SlidersHorizontal, Volume2, VolumeX, Zap } from 'lucide-react'
import { useEffect, useState } from 'react'

import { fmtClock } from '../format'
import { SESSIONS } from '../sessions'
import type { Status } from '../types'

interface Props {
  status: Status | null
  online: boolean
  soundOn: boolean
  onToggleSound: () => void
  onTestSound: () => void
  onOpenFilters: () => void
}

export function Header({ status, online, soundOn, onToggleSound, onTestSound, onOpenFilters }: Props) {
  const [now, setNow] = useState(() => new Date())

  useEffect(() => {
    const t = setInterval(() => setNow(new Date()), 1000)

    return () => clearInterval(t)
  }, [])

  const session = SESSIONS[status?.session ?? 'closed']
  const feedOk = online && status?.feed === 'connected'

  return (
    <header className="sticky top-0 z-20 border-b border-white/5 bg-zinc-950/70 backdrop-blur-xl">
      <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-6 py-4">
        <div className="flex items-center gap-3">
          <div className="grid size-9 place-items-center rounded-xl bg-emerald-400/15 ring-1 ring-emerald-400/30">
            <Zap className="size-5 text-emerald-300" fill="currentColor" />
          </div>
          <div>
            <h1 className="text-lg font-semibold tracking-tight">momo</h1>
            <p className="text-xs text-zinc-500">NASDAQ small-cap momentum</p>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <span className="hidden font-mono text-sm tabular-nums text-zinc-400 sm:block">{fmtClock(now)}</span>

          <span className={`rounded-full px-3 py-1 text-xs font-medium ring-1 ${session.badge}`}>{session.label}</span>

          <span
            title={status?.feed ?? 'no status'}
            className="flex items-center gap-2 rounded-full bg-white/5 px-3 py-1 text-xs text-zinc-400 ring-1 ring-white/10"
          >
            <span className={`size-2 rounded-full ${feedOk ? 'bg-emerald-400 shadow-[0_0_8px] shadow-emerald-400' : 'bg-rose-500'}`} />
            {feedOk ? `Live · ${status?.tracked.toLocaleString()} symbols` : 'Feed offline'}
          </span>

          <div className="flex overflow-hidden rounded-full ring-1 ring-white/10">
            <button
              onClick={onToggleSound}
              className={`flex items-center gap-2 px-3 py-1.5 text-xs font-medium transition ${soundOn ? 'bg-emerald-400/15 text-emerald-300' : 'bg-white/5 text-zinc-400 hover:text-zinc-200'}`}
            >
              {soundOn ? <Volume2 className="size-4" /> : <VolumeX className="size-4" />}
              {soundOn ? 'Sound on' : 'Sound off'}
            </button>
            <button onClick={onTestSound} className="border-l border-white/10 bg-white/5 px-3 text-xs text-zinc-400 transition hover:text-zinc-100">
              Test
            </button>
          </div>

          <button
            onClick={onOpenFilters}
            className="flex items-center gap-2 rounded-full bg-white/5 px-3 py-1.5 text-xs font-medium text-zinc-300 ring-1 ring-white/10 transition hover:bg-white/10"
          >
            <SlidersHorizontal className="size-4" />
            Filters
          </button>
        </div>
      </div>
    </header>
  )
}
