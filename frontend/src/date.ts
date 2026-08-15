export type Scale = 'day' | 'week' | 'month' | 'year'

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

export function eachDate(from: Date, to: Date): Date[] {
  const days: Date[] = []
  const cursor = new Date(from)
  while (cursor <= to) {
    days.push(new Date(cursor))
    cursor.setDate(cursor.getDate() + 1)
  }
  return days
}
