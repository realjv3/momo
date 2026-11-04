const etTime = new Intl.DateTimeFormat('en-US', {
  timeZone: 'America/New_York',
  month: 'short',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
  hour12: false,
  timeZoneName: 'short',
})

export function fmtDateTime(iso: string): string {
  const parts = Object.fromEntries(etTime.formatToParts(new Date(iso)).map((p) => [p.type, p.value]))

  return `${parts.month} ${parts.day}  ${parts.hour}:${parts.minute}:${parts.second} ${parts.timeZoneName}`
}

const clock = new Intl.DateTimeFormat('en-US', {
  timeZone: 'America/New_York',
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit',
  hour12: false,
  timeZoneName: 'short',
})

export const fmtClock = (d: Date): string => clock.format(d)

export function fmtCompact(n: number): string {
  if (!n) return '-'
  if (n >= 1e9) return `${(n / 1e9).toFixed(2)}B`
  if (n >= 1e6) return `${(n / 1e6).toFixed(2)}M`
  if (n >= 1e3) return `${(n / 1e3).toFixed(1)}K`

  return n.toFixed(0)
}

export const fmtPrice = (n: number): string => `$${n.toFixed(n < 1 ? 4 : 2)}`

export const fmtPct = (n: number): string => `${n >= 0 ? '+' : ''}${n.toFixed(2)}%`
