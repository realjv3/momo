let ctx: AudioContext | null = null

// unlockAudio resolves true once audio can play; browsers require a user gesture after each reload.
export async function unlockAudio(): Promise<boolean> {
  ctx ??= new AudioContext()
  if (ctx.state === 'suspended') await ctx.resume()

  return ctx.state === 'running'
}

// pop plays a soft low "plop"; returns false when audio is still locked.
export function pop(): boolean {
  if (!ctx || ctx.state !== 'running') return false

  const t = ctx.currentTime
  const osc = ctx.createOscillator()
  const gain = ctx.createGain()

  osc.type = 'sine'
  osc.frequency.setValueAtTime(520, t)
  osc.frequency.exponentialRampToValueAtTime(130, t + 0.1)

  gain.gain.setValueAtTime(0.0001, t)
  gain.gain.exponentialRampToValueAtTime(0.9, t + 0.012)
  gain.gain.exponentialRampToValueAtTime(0.0001, t + 0.22)

  osc.connect(gain).connect(ctx.destination)
  osc.start(t)
  osc.stop(t + 0.25)

  return true
}
