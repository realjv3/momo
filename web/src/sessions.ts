import type { SessionName } from './types'

interface SessionStyle {
  label: string
  short: string
  badge: string
  chip: string
}

export const SESSIONS: Record<SessionName, SessionStyle> = {
  pre: { label: 'Pre-market', short: 'PRE', badge: 'bg-amber-400/15 text-amber-300 ring-amber-400/30', chip: 'text-amber-300 bg-amber-400/10' },
  regular: { label: 'Regular hours', short: 'RTH', badge: 'bg-emerald-400/15 text-emerald-300 ring-emerald-400/30', chip: 'text-emerald-300 bg-emerald-400/10' },
  post: { label: 'After-hours', short: 'POST', badge: 'bg-violet-400/15 text-violet-300 ring-violet-400/30', chip: 'text-violet-300 bg-violet-400/10' },
  closed: { label: 'Closed', short: '-', badge: 'bg-zinc-500/15 text-zinc-400 ring-zinc-500/30', chip: 'text-zinc-400 bg-zinc-500/10' },
}
