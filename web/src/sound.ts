let ctx: AudioContext | null = null

export function unlockAudio(): void {
  ctx ??= new AudioContext()
  if (ctx.state === 'suspended') void ctx.resume()
}

export const audioReady = (): boolean => ctx?.state === 'running'

// pop synthesizes a short descending sine blip; returns false when audio is still locked.
export function pop(): boolean {
  if (!ctx || ctx.state !== 'running') return false

  const t = ctx.currentTime
  const osc = ctx.createOscillator()
  const gain = ctx.createGain()

  osc.type = 'sine'
  osc.frequency.setValueAtTime(1100, t)
  osc.frequency.exponentialRampToValueAtTime(240, t + 0.09)
  gain.gain.setValueAtTime(0.0001, t)
  gain.gain.exponentialRampToValueAtTime(0.6, t + 0.006)
  gain.gain.exponentialRampToValueAtTime(0.0001, t + 0.14)

  osc.connect(gain).connect(ctx.destination)
  osc.start(t)
  osc.stop(t + 0.15)

  return true
}
