export type Scale = 'day' | 'week' | 'month' | 'year'

export const WEEKDAY_LABELS_JA = ['日', '月', '火', '水', '木', '金', '土']

export function describeRecurrence(rule: {
  rule_type: 'interval' | 'weekly'
  interval_days?: number
  weekdays?: number[]
}): string {
  if (rule.rule_type === 'weekly') {
    const days = (rule.weekdays ?? []).map((d) => WEEKDAY_LABELS_JA[d]).join('・')
    return `毎週${days || '?'}曜日`
  }
  const n = rule.interval_days ?? 1
  if (n === 1) return '毎日'
  if (n === 2) return '隔日'
  return `${n}日ごと`
}

export function startOfPeriod(scale: Scale, date: Date): Date {
  const d = new Date(date.getFullYear(), date.getMonth(), date.getDate())
  switch (scale) {
    case 'day':
      return d
    case 'week': {
      const offset = (d.getDay() + 6) % 7 // Monday = 0 ... Sunday = 6
      d.setDate(d.getDate() - offset)
      return d
    }
    case 'month':
      return new Date(d.getFullYear(), d.getMonth(), 1)
    case 'year':
      return new Date(d.getFullYear(), 0, 1)
  }
}

export function endOfPeriod(scale: Scale, anchor: Date): Date {
  switch (scale) {
    case 'day':
      return anchor
    case 'week': {
      const d = new Date(anchor)
      d.setDate(d.getDate() + 6)
      return d
    }
    case 'month':
      return new Date(anchor.getFullYear(), anchor.getMonth() + 1, 0)
    case 'year':
      return new Date(anchor.getFullYear(), 11, 31)
  }
}

export function shiftPeriod(scale: Scale, anchor: Date, direction: 1 | -1): Date {
  const d = new Date(anchor)
  switch (scale) {
    case 'day':
      d.setDate(d.getDate() + direction)
      break
    case 'week':
      d.setDate(d.getDate() + direction * 7)
      break
    case 'month':
      d.setMonth(d.getMonth() + direction)
      break
    case 'year':
      d.setFullYear(d.getFullYear() + direction)
      break
  }
  return startOfPeriod(scale, d)
}

export function formatISODate(d: Date): string {
  const yyyy = d.getFullYear()
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd}`
}

export function formatPeriodLabel(scale: Scale, anchor: Date): string {
  switch (scale) {
    case 'day':
      return formatISODate(anchor)
    case 'week':
      return `${formatISODate(anchor)} ~ ${formatISODate(endOfPeriod('week', anchor))}`
    case 'month':
      return `${anchor.getFullYear()}-${String(anchor.getMonth() + 1).padStart(2, '0')}`
    case 'year':
      return `${anchor.getFullYear()}`
  }
}

export function daysUntil(isoDate: string): number {
  const target = new Date(isoDate)
  const targetMidnight = new Date(target.getFullYear(), target.getMonth(), target.getDate())
  const now = new Date()
  const todayMidnight = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  return Math.round((targetMidnight.getTime() - todayMidnight.getTime()) / 86_400_000)
}

export function formatShortDateJa(isoDate: string): string {
  // Appending a local-time suffix avoids the bare "YYYY-MM-DD" UTC-parsing
  // gotcha (new Date("2026-08-17") is midnight UTC, which can display as
  // the previous day in negative-UTC-offset timezones).
  const d = new Date(`${isoDate}T00:00:00`)
  return `${d.getMonth() + 1}/${d.getDate()}(${WEEKDAY_LABELS_JA[d.getDay()]})`
}

export function eachDate(from: Date, to: Date): Date[] {
  const days: Date[] = []
  const cursor = new Date(from)
  while (cursor <= to) {
    days.push(new Date(cursor))
    cursor.setDate(cursor.getDate() + 1)
  }
  return days
}
