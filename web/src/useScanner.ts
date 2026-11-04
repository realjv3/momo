import { useEffect, useRef, useState } from 'react'

import type { Signal, Status } from './types'

const MAX_ROWS = 500

const byTimeDesc = (a: Signal, b: Signal) => b.time.localeCompare(a.time) || b.id - a.id

export function useScanner(onLive: (s: Signal) => void) {
  const [signals, setSignals] = useState<Signal[]>([])
  const [status, setStatus] = useState<Status | null>(null)
  const [online, setOnline] = useState(false)
  const [freshIds, setFreshIds] = useState<Set<number>>(new Set())
  const liveRef = useRef(onLive)

  useEffect(() => {
    liveRef.current = onLive
  })

  useEffect(() => {
    let ws: WebSocket | null = null
    let retry: ReturnType<typeof setTimeout>
    let closed = false

    // silent merge by id: used on every (re)connect so gap rows appear without the pop
    const loadSignals = () =>
      fetch('/api/signals')
        .then((r) => r.json() as Promise<Signal[]>)
        .then((loaded) =>
          setSignals((prev) => {
            const seen = new Set(prev.map((s) => s.id))

            return [...prev, ...loaded.filter((s) => !seen.has(s.id))].sort(byTimeDesc).slice(0, MAX_ROWS)
          }),
        )
        .catch(() => {})

    const connect = () => {
      const proto = location.protocol === 'https:' ? 'wss' : 'ws'
      ws = new WebSocket(`${proto}://${location.host}/api/stream`)
      ws.onopen = () => {
        setOnline(true)
        void loadSignals()
      }
      ws.onclose = () => {
        setOnline(false)
        if (!closed) retry = setTimeout(connect, 2000)
      }
      ws.onmessage = (ev) => {
        const msg = JSON.parse(ev.data) as { type: string; data: unknown }

        if (msg.type === 'status') {
          setStatus(msg.data as Status)

          return
        }

        if (msg.type === 'signal') {
          const s = msg.data as Signal

          setSignals((prev) => (prev.some((p) => p.id === s.id) ? prev : [s, ...prev].slice(0, MAX_ROWS)))
          setFreshIds((prev) => new Set(prev).add(s.id))
          setTimeout(
            () =>
              setFreshIds((prev) => {
                const next = new Set(prev)
                next.delete(s.id)

                return next
              }),
            2500,
          )
          liveRef.current(s)
        }
      }
    }

    connect()

    return () => {
      closed = true
      clearTimeout(retry)
      ws?.close()
    }
  }, [])

  return { signals, status, online, freshIds }
}
