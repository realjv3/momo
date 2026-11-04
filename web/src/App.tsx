import { Volume2 } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'

import { FiltersPanel } from './components/FiltersPanel'
import { Header } from './components/Header'
import { SignalTable } from './components/SignalTable'
import { audioReady, pop, unlockAudio } from './sound'
import { useScanner } from './useScanner'

const SOUND_KEY = 'momo.sound'

export default function App() {
  const [soundOn, setSoundOn] = useState(() => localStorage.getItem(SOUND_KEY) === '1')
  const [locked, setLocked] = useState(false)
  const [filtersOpen, setFiltersOpen] = useState(false)

  const onLive = useCallback(() => {
    if (soundOn && !pop()) setLocked(true)
  }, [soundOn])

  const { signals, status, online, freshIds } = useScanner(onLive)

  // browsers block audio until a gesture; unlock on the first interaction
  useEffect(() => {
    if (!soundOn) return

    const unlock = () => {
      unlockAudio()
      setLocked(!audioReady())
    }

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

    if (next) {
      unlockAudio()
      setLocked(!audioReady())
    }
  }

  const testSound = () => {
    unlockAudio()
    // context resume is async; retry once it is running
    setTimeout(() => setLocked(!pop()), 50)
  }

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
            onClick={() => {
              unlockAudio()
              setLocked(false)
            }}
            className="mb-4 flex w-full items-center justify-center gap-2 rounded-xl bg-amber-400/10 py-3 text-sm font-medium text-amber-300 ring-1 ring-amber-400/30"
          >
            <Volume2 className="size-4" />
            Click anywhere to enable audio alerts
          </button>
        )}

        <SignalTable signals={signals} freshIds={freshIds} />
      </main>

      <FiltersPanel open={filtersOpen} onClose={() => setFiltersOpen(false)} />
    </div>
  )
}
