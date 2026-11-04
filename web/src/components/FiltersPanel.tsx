import { AnimatePresence, motion } from 'framer-motion'
import { RotateCcw, X } from 'lucide-react'
import { useEffect, useState } from 'react'

import { SESSIONS } from '../sessions'
import type { Filters } from '../types'

type NumKey = { [K in keyof Filters]: Filters[K] extends number ? K : never }[keyof Filters]

interface Field {
  key: NumKey
  label: string
  hint?: string
  scale?: number
  step?: number
}

const GROUPS: { title: string; fields: Field[] }[] = [
  {
    title: 'Price and change',
    fields: [
      { key: 'minPrice', label: 'Min price ($)', step: 0.01 },
      { key: 'maxPrice', label: 'Max price ($)', hint: '0 = no cap', step: 0.01 },
      { key: 'minChangePct', label: 'Min change %', hint: 'vs prior close', step: 0.5 },
      { key: 'maxChangePct', label: 'Max change %', hint: '0 = no cap', step: 0.5 },
      { key: 'minChangeFromOpenPct', label: 'Min change from open %', hint: '0 = off', step: 0.5 },
    ],
  },
  {
    title: 'Volume and float',
    fields: [
      { key: 'minSessionVolume', label: 'Min session volume', step: 1000 },
      { key: 'minBarVolume', label: 'Min 1-min volume', step: 500 },
      { key: 'priorBars', label: 'Prior bars checked', step: 1 },
      { key: 'minPriorBarVolume', label: 'Min volume per prior bar', hint: 'missing minute = 0', step: 500 },
      { key: 'minFloat', label: 'Min float (M)', scale: 1e6, step: 0.5 },
      { key: 'maxFloat', label: 'Max float (M)', hint: '0 = no cap', scale: 1e6, step: 0.5 },
    ],
  },
  {
    title: 'Momentum',
    fields: [
      { key: 'emaPeriod', label: 'EMA period', hint: 'price must be above', step: 1 },
      { key: 'moveWindowMinutes', label: 'Move window (min)', step: 1 },
      { key: 'minMovePct', label: 'Min rise off window low %', step: 0.5 },
      { key: 'volWindowMinutes', label: 'Volume baseline (min)', step: 1 },
      { key: 'volRiseFactor', label: 'Volume vs baseline (x)', step: 0.1 },
      { key: 'cooldownMinutes', label: 'Per-symbol cooldown (min)', step: 1 },
    ],
  },
]

const SESSION_TOGGLES = [
  { key: 'preMarket', session: 'pre' },
  { key: 'regular', session: 'regular' },
  { key: 'postMarket', session: 'post' },
] as const

interface Props {
  open: boolean
  onClose: () => void
}

export function FiltersPanel({ open, onClose }: Props) {
  const [draft, setDraft] = useState<Filters | null>(null)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) return

    fetch('/api/filters')
      .then((r) => r.json() as Promise<Filters>)
      .then(setDraft)
      .catch(() => setError('Could not load filters'))
  }, [open])

  const reset = async () => {
    setSaving(true)
    setError('')

    try {
      const res = await fetch('/api/filters', { method: 'DELETE' })
      if (!res.ok) throw new Error(await res.text())

      setDraft((await res.json()) as Filters)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Reset failed')
    } finally {
      setSaving(false)
    }
  }

  const save = async () => {
    if (!draft) return

    setSaving(true)
    setError('')

    try {
      const res = await fetch('/api/filters', { method: 'PUT', body: JSON.stringify(draft) })
      if (!res.ok) throw new Error(await res.text())

      onClose()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Save failed')
    } finally {
      setSaving(false)
    }
  }

  return (
    <AnimatePresence>
      {open && (
        <>
          <motion.div
            className="fixed inset-0 z-30 bg-black/50 backdrop-blur-sm"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
          />
          <motion.aside
            className="fixed inset-y-0 right-0 z-40 flex w-full max-w-md flex-col border-l border-white/10 bg-zinc-950"
            initial={{ x: '100%' }}
            animate={{ x: 0 }}
            exit={{ x: '100%' }}
            transition={{ type: 'spring', damping: 32, stiffness: 320 }}
          >
            <div className="flex items-center justify-between border-b border-white/5 px-6 py-4">
              <h2 className="text-base font-semibold">Scan criteria</h2>
              <button onClick={onClose} className="rounded-lg p-1.5 text-zinc-400 hover:bg-white/5 hover:text-zinc-100">
                <X className="size-5" />
              </button>
            </div>

            <div className="flex-1 space-y-8 overflow-y-auto px-6 py-6">
              {!draft && <p className="text-sm text-zinc-500">Loading...</p>}

              {draft && (
                <>
                  <section>
                    <h3 className="mb-3 text-xs font-semibold uppercase tracking-wider text-zinc-500">Sessions</h3>
                    <div className="flex gap-2">
                      {SESSION_TOGGLES.map((s) => (
                        <button
                          key={s.key}
                          onClick={() => setDraft({ ...draft, [s.key]: !draft[s.key] })}
                          className={`flex-1 rounded-lg px-3 py-2 text-xs font-medium ring-1 transition ${draft[s.key] ? 'bg-emerald-400/15 text-emerald-300 ring-emerald-400/30' : 'bg-white/5 text-zinc-500 ring-white/10'}`}
                        >
                          {SESSIONS[s.session].label}
                        </button>
                      ))}
                    </div>
                  </section>

                  {GROUPS.map((g) => (
                    <section key={g.title}>
                      <h3 className="mb-3 text-xs font-semibold uppercase tracking-wider text-zinc-500">{g.title}</h3>
                      <div className="grid grid-cols-2 gap-3">
                        {g.fields.map((f) => (
                          <label key={f.key} className="block">
                            <span className="mb-1 block text-xs text-zinc-400">{f.label}</span>
                            <input
                              type="number"
                              step={f.step}
                              value={draft[f.key] / (f.scale ?? 1)}
                              onChange={(e) => setDraft({ ...draft, [f.key]: Number(e.target.value) * (f.scale ?? 1) })}
                              className="w-full rounded-lg bg-white/5 px-3 py-2 font-mono text-sm tabular-nums outline-none ring-1 ring-white/10 transition focus:ring-emerald-400/60"
                            />
                            {f.hint && <span className="mt-1 block text-[10px] text-zinc-600">{f.hint}</span>}
                          </label>
                        ))}
                      </div>
                    </section>
                  ))}
                </>
              )}
            </div>

            <div className="border-t border-white/5 px-6 py-4">
              {error && <p className="mb-2 text-xs text-rose-400">{error}</p>}
              <div className="flex gap-3">
                <button
                  onClick={reset}
                  disabled={saving}
                  className="flex items-center gap-2 rounded-xl bg-white/5 px-4 py-2.5 text-sm font-medium text-zinc-300 ring-1 ring-white/10 transition hover:bg-white/10 disabled:opacity-50"
                >
                  <RotateCcw className="size-4" />
                  Reset
                </button>
                <button
                  onClick={save}
                  disabled={!draft || saving}
                  className="flex-1 rounded-xl bg-emerald-400 py-2.5 text-sm font-semibold text-zinc-950 transition hover:bg-emerald-300 disabled:opacity-50"
                >
                  {saving ? 'Working...' : 'Apply filters'}
                </button>
              </div>
            </div>
          </motion.aside>
        </>
      )}
    </AnimatePresence>
  )
}
