import { Volume2 } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'

import { FiltersPanel } from './components/FiltersPanel'
import { Header } from './components/Header'
import { Pager } from './components/Pager'
import { SignalTable } from './components/SignalTable'
import { pop, unlockAudio } from './sound'
import { useScanner } from './useScanner'

const SOUND_KEY = 'momo.sound'

export default function App() {
  const [soundOn, setSoundOn] = useState(() => localStorage.getItem(SOUND_KEY) === '1')
  const [locked, setLocked] = useState(soundOn)
  const [filtersOpen, setFiltersOpen] = useState(false)

  const onLive = useCallback(() => {
    if (soundOn && !pop()) setLocked(true)
  }, [soundOn])

  const { signals, total, page, setPage, pageSize, setPageSize, pendingNew, status, online, freshIds } = useScanner(onLive)

  // browsers block audio until a gesture; unlock on the first interaction
  useEffect(() => {
    if (!soundOn) return

    const unlock = () => void unlockAudio().then((ok) => setLocked(!ok))

    document.addEventListener('pointerdown', unlock, { once: true })
    document.addEventListener('keydown', unlock, { once: true })

    return () => {
      document.removeEventListener('pointerdown', unlock)
      document.removeEventListener('keydown', unlock)
    }
  }, [soundOn])

  const toggleSound = () => {
    const next = !soundOn

    localStorage.setItem(SOUND_KEY, next ? '1' : '0')
    setSoundOn(next)

    if (next) void unlockAudio().then((ok) => setLocked(!ok))
  }

  const testSound = () =>
    void unlockAudio().then((ok) => {
      setLocked(!ok)
      if (ok) pop()
    })

  return (
    <div className="min-h-screen">
      <Header
        status={status}
        online={online}
        soundOn={soundOn}
        onToggleSound={toggleSound}
        onTestSound={testSound}
        onOpenFilters={() => setFiltersOpen(true)}
      />

      <main className="mx-auto max-w-6xl px-6 py-8">
        {soundOn && locked && (
          <button
            onClick={() => void unlockAudio().then((ok) => setLocked(!ok))}
            className="mb-4 flex w-full items-center justify-center gap-2 rounded-xl bg-amber-400/10 py-3 text-sm font-medium text-amber-300 ring-1 ring-amber-400/30"
          >
            <Volume2 className="size-4" />
            Click anywhere to enable audio alerts
          </button>
        )}

        {page > 0 && pendingNew > 0 && (
          <button
            onClick={() => setPage(0)}
            className="mb-4 w-full rounded-xl bg-emerald-400/10 py-2.5 text-sm font-medium text-emerald-300 ring-1 ring-emerald-400/30"
          >
            {pendingNew} new {pendingNew === 1 ? 'signal' : 'signals'} - back to latest
          </button>
        )}

        <SignalTable signals={signals} freshIds={freshIds} />

        {total > 0 && <Pager page={page} pageSize={pageSize} total={total} onPage={setPage} onPageSize={setPageSize} />}
      </main>

      <FiltersPanel open={filtersOpen} onClose={() => setFiltersOpen(false)} />
    </div>
  )
}
