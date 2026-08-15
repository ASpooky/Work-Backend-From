import type { GoalCalendar } from '../api'
import { formatISODate } from '../date'

const WEEKDAY_LABELS = ['月', '火', '水', '木', '金', '土', '日']
const COLUMN_WIDTH = 100 / 7

type Props = {
  calendar: GoalCalendar[]
  days: Date[]
}

function WeekPathView({ calendar, days }: Props) {
  const dayKeys = days.map(formatISODate)

  return (
    <div className="week-path">
      <div className="week-path-header">
        <span className="week-path-range">{formatJapaneseRange(days)}</span>
        <span className="week-path-count">ゴール {calendar.length}件</span>
      </div>

      <div className="wk week-path-weekday-row">
        {days.map((d, i) => (
          <div key={dayKeys[i]} className={i === 6 ? 'accent' : ''}>
            {WEEKDAY_LABELS[i]} {d.getDate()}
          </div>
        ))}
      </div>

      {calendar.map((gc) => (
        <GoalLane key={gc.goal.id} goalCalendar={gc} dayKeys={dayKeys} />
      ))}

      {calendar.length === 0 && <p className="week-path-empty">まだ目標がありません</p>}
    </div>
  )
}

function GoalLane({ goalCalendar, dayKeys }: { goalCalendar: GoalCalendar; dayKeys: string[] }) {
  const goal = goalCalendar.goal
  const goalEndKey = goal.end_date.slice(0, 10)
  const goalEndIndex = dayKeys.indexOf(goalEndKey)
  const firstContentIndex = dayKeys.findIndex((k) => goalCalendar.days[k]?.content)

  const showLine = firstContentIndex !== -1 || goalEndIndex !== -1
  const lineStart = firstContentIndex === -1 ? 0 : firstContentIndex
  const lineLeft = `calc(${lineStart * COLUMN_WIDTH}% + 6px)`
  const lineRight = goalEndIndex === -1 ? '0%' : `${(7 - goalEndIndex) * COLUMN_WIDTH}%`
  const arrowRight = `calc(${(7 - goalEndIndex) * COLUMN_WIDTH}% - 4px)`

  return (
    <div className="lane">
      {showLine && <div className="line" style={{ left: lineLeft, right: lineRight }} />}
      {goalEndIndex !== -1 && (
        <span className="arw" style={{ right: arrowRight }} aria-hidden="true">
          &rsaquo;
        </span>
      )}
      <div className="wk">
        {dayKeys.map((key, i) => (
          <div className="cell" key={key}>
            {i === goalEndIndex ? (
              <div className="goal">
                <b>{goal.title}</b>
                {goal.achievement_condition}
              </div>
            ) : (
              goalCalendar.days[key]?.content && (
                <div className={`card status-${goalCalendar.days[key].status}`}>
                  <b>{goalCalendar.days[key].content}</b>
                  {goalCalendar.days[key].status === 'done' && <span>済</span>}
                  {goalCalendar.days[key].status === 'not_done' && <span>未</span>}
                </div>
              )
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

function formatJapaneseRange(days: Date[]): string {
  const from = days[0]
  const to = days[days.length - 1]
  if (from.getMonth() === to.getMonth()) {
    return `${from.getMonth() + 1}月${from.getDate()}日 – ${to.getDate()}日`
  }
  return `${from.getMonth() + 1}月${from.getDate()}日 – ${to.getMonth() + 1}月${to.getDate()}日`
}

export default WeekPathView
