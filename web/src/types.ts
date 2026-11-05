export type SessionName = 'pre' | 'regular' | 'post' | 'closed'

export interface Signal {
  id: number
  time: string
  symbol: string
  session: SessionName
  changePct: number
  price: number
  float: number
  volume: number
}

export interface SignalPage {
  signals: Signal[]
  total: number
}

export interface Status {
  session: SessionName
  feed: string
  tracked: number
}

export interface Filters {
  minPrice: number
  maxPrice: number
  minChangePct: number
  maxChangePct: number
  minChangeFromOpenPct: number
  minSessionVolume: number
  minBarVolume: number
  priorBars: number
  minPriorBarVolume: number
  minFloat: number
  maxFloat: number
  moveWindowMinutes: number
  minMovePct: number
  volWindowMinutes: number
  volRiseFactor: number
  emaPeriod: number
  cooldownMinutes: number
  preMarket: boolean
  regular: boolean
  postMarket: boolean
}
