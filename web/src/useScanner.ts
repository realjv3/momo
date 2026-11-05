import { useCallback, useEffect, useRef, useState } from 'react'

import type { Signal, SignalPage, Status } from './types'

const fetchPage = (page: number, size: number) =>
  fetch(`/api/signals?limit=${size}&offset=${page * size}`).then((r) => r.json() as Promise<SignalPage>)

export function useScanner(onLive: (s: Signal) => void) {
  const [signals, setSignals] = useState<Signal[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0)
  const [pageSize, setPageSizeState] = useState(25)
  const [pendingNew, setPendingNew] = useState(0)
  const [status, setStatus] = useState<Status | null>(null)
  const [online, setOnline] = useState(false)
  const [freshIds, setFreshIds] = useState<Set<number>>(new Set())

  const liveRef = useRef(onLive)
  const pageRef = useRef(page)
  const sizeRef = useRef(pageSize)

  const applyPage = useCallback((data: SignalPage, forPage: number, size: number) => {
    setSignals(data.signals)
    setTotal(data.total)

    if (forPage === 0) setPendingNew(0)
    else if (data.signals.length === 0) setPage(Math.max(0, Math.ceil(data.total / size) - 1))
  }, [])

  const reloadRef = useRef(() => {})

  useEffect(() => {
    liveRef.current = onLive
    pageRef.current = page
    sizeRef.current = pageSize
    reloadRef.current = () => {
      const [p, size] = [pageRef.current, sizeRef.current]

      fetchPage(p, size)
        .then((data) => applyPage(data, p, size))
        .catch(() => {})
    }
  })

  useEffect(() => {
    let cancelled = false

    fetchPage(page, pageSize)
      .then((data) => {
        if (!cancelled) applyPage(data, page, pageSize)
      })
      .catch(() => {})

    return () => {
      cancelled = true
    }
  }, [page, pageSize, applyPage])

  useEffect(() => {
    let ws: WebSocket | null = null
    let retry: ReturnType<typeof setTimeout>
    let closed = false

    const connect = () => {
      const proto = location.protocol === 'https:' ? 'wss' : 'ws'
      ws = new WebSocket(`${proto}://${location.host}/api/stream`)
      ws.onopen = () => {
        setOnline(true)
        reloadRef.current()
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

          setTotal((t) => t + 1)

          if (pageRef.current === 0) {
            setSignals((prev) => (prev.some((p) => p.id === s.id) ? prev : [s, ...prev].slice(0, sizeRef.current)))
          } else {
            setPendingNew((n) => n + 1)
          }

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

  const setPageSize = (n: number) => {
    setPageSizeState(n)
    setPage(0)
  }

  return { signals, total, page, setPage, pageSize, setPageSize, pendingNew, status, online, freshIds }
}
