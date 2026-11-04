import { AnimatePresence, motion } from 'framer-motion'
import { Radar } from 'lucide-react'

import { fmtCompact, fmtDateTime, fmtPct, fmtPrice } from '../format'
import { SESSIONS } from '../sessions'
import type { Signal } from '../types'

const HEADERS = ['Date / time (ET)', 'Symbol', 'Change %', 'Price', 'Free float', 'Volume']

interface Props {
  signals: Signal[]
  freshIds: Set<number>
}

export function SignalTable({ signals, freshIds }: Props) {
  if (signals.length === 0) {
    return (
      <div className="grid place-items-center rounded-2xl border border-dashed border-white/10 py-24 text-center">
        <Radar className="mb-3 size-8 animate-pulse text-emerald-400/70" />
        <p className="text-sm text-zinc-300">Scanning for momentum</p>
        <p className="mt-1 text-xs text-zinc-500">New rows appear at the top the moment a symbol meets your criteria.</p>
      </div>
    )
  }

  return (
    <div className="overflow-hidden rounded-2xl bg-zinc-900/60 ring-1 ring-white/10 backdrop-blur">
      <table className="w-full border-collapse text-sm">
        <thead>
          <tr className="border-b border-white/5 text-left text-[11px] uppercase tracking-wider text-zinc-500">
            {HEADERS.map((h, i) => (
              <th key={h} className={`px-5 py-3 font-medium ${i >= 2 ? 'text-right' : ''}`}>
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          <AnimatePresence initial={false}>
            {signals.map((s) => {
              const chip = SESSIONS[s.session]

              return (
                <motion.tr
                  key={s.id}
                  layout="position"
                  initial={{ opacity: 0, backgroundColor: 'rgba(52, 211, 153, 0.35)' }}
                  animate={{ opacity: 1, backgroundColor: freshIds.has(s.id) ? 'rgba(52, 211, 153, 0.12)' : 'rgba(52, 211, 153, 0)' }}
                  transition={{ duration: 0.5 }}
                  className="border-b border-white/5 last:border-0 hover:bg-white/[0.03]"
                >
                  <td className="px-5 py-3 font-mono text-xs tabular-nums text-zinc-400">{fmtDateTime(s.time)}</td>
                  <td className="px-5 py-3">
                    <span className="font-mono text-[15px] font-semibold tracking-wide">{s.symbol}</span>
                    <span className={`ml-2 rounded px-1.5 py-0.5 text-[10px] font-semibold ${chip.chip}`}>{chip.short}</span>
                  </td>
                  <td className={`px-5 py-3 text-right font-mono font-semibold tabular-nums ${s.changePct >= 0 ? 'text-emerald-400' : 'text-rose-400'}`}>
                    {fmtPct(s.changePct)}
                  </td>
                  <td className="px-5 py-3 text-right font-mono tabular-nums">{fmtPrice(s.price)}</td>
                  <td className="px-5 py-3 text-right font-mono tabular-nums text-zinc-300">{fmtCompact(s.float)}</td>
                  <td className="px-5 py-3 text-right font-mono tabular-nums text-zinc-300">{fmtCompact(s.volume)}</td>
                </motion.tr>
              )
            })}
          </AnimatePresence>
        </tbody>
      </table>
    </div>
  )
}
