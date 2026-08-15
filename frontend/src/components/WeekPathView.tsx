import type { GoalCalendar, DayStatus } from '../api'
import { formatISODate } from '../date'
import './WeekPathView.css'

const WEEKDAY_LABELS = ['月', '火', '水', '木', '金', '土', '日']

type Props = {
  calendar: GoalCalendar[]
  days: Date[]
}

type NodeState = 'done' | 'missed' | 'today-pending' | 'upcoming' | 'no-task' | 'milestone'

function dayNodeState(status: DayStatus, isMilestone: boolean, dayKey: string, todayKey: string): NodeState {
  if (isMilestone) return 'milestone'
  if (status === 'no_task') return 'no-task'
  if (status === 'done') return 'done'
  if (dayKey === todayKey) return 'today-pending'
  if (dayKey < todayKey) return 'missed'
  return 'upcoming'
}

function countWeek(days: Record<string, { status: DayStatus }>, dayKeys: string[]) {
  let scheduled = 0
  let done = 0
  for (const key of dayKeys) {
    const status = days[key]?.status ?? 'no_task'
    if (status === 'no_task') continue
    scheduled++
    if (status === 'done') done++
  }
  return { scheduled, done }
}

function WeekPathView({ calendar, days }: Props) {
  const dayKeys = days.map(formatISODate)
  const todayKey = formatISODate(new Date())

  const total = calendar.reduce(
    (acc, gc) => {
      const { scheduled, done } = countWeek(gc.days, dayKeys)
      acc.scheduled += scheduled
      acc.done += done
      return acc
    },
    { done: 0, scheduled: 0 },
  )

  return (
    <div className="week-path">
      <div className="week-path-header">
        <div>
          <div className="week-path-eyebrow">今週の消化</div>
          <div className="week-path-total">
            <span className="week-path-total-num">{total.done}</span>
            <span className="week-path-total-denom">/ {total.scheduled} タスク</span>
          </div>
        </div>
        <span className="week-path-range">{formatJapaneseRange(days)}</span>
      </div>

      {calendar.length === 0 ? (
        <p className="week-path-empty">まだ目標がありません</p>
      ) : (
        <div className="week-path-table-wrap">
          <table className="week-path-table">
            <caption className="sr-only">
              週間目標カレンダー。目標ごとに1行、曜日ごとの達成状況が並び、達成予定日には目印が付く。
            </caption>
            <thead>
              <tr>
                <th scope="col">
                  <span className="sr-only">目標</span>
                </th>
                {days.map((d, i) => (
                  <th scope="col" key={dayKeys[i]} className={dayKeys[i] === todayKey ? 'accent' : ''}>
                    {WEEKDAY_LABELS[i]} {d.getDate()}
                  </th>
                ))}
                <th scope="col">今週</th>
              </tr>
            </thead>
            <tbody>
              {calendar.map((gc) => (
                <GoalRow key={gc.goal.id} goalCalendar={gc} dayKeys={dayKeys} todayKey={todayKey} />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

function GoalRow({
  goalCalendar,
  dayKeys,
  todayKey,
}: {
  goalCalendar: GoalCalendar
  dayKeys: string[]
  todayKey: string
}) {
  const goal = goalCalendar.goal
  const goalEndKey = goal.end_date.slice(0, 10)
  const { scheduled, done } = countWeek(goalCalendar.days, dayKeys)
  const percent = scheduled === 0 ? null : Math.round((done / scheduled) * 100)

  return (
    <tr>
      <th scope="row" className="week-path-goal-cell">
        <div className="week-path-goal-title">{goal.title}</div>
        <div className="week-path-goal-sub">{goal.achievement_condition}</div>
      </th>
      {dayKeys.map((key, i) => {
        const isMilestone = key === goalEndKey
        const entry = goalCalendar.days[key]
        const status = entry?.status ?? 'no_task'
        const state = dayNodeState(status, isMilestone, key, todayKey)
        const content = isMilestone ? goal.achievement_condition : (entry?.content ?? '')
        return (
          <DayCell
            key={key}
            state={state}
            content={content}
            isFirst={i === 0}
            isLast={i === dayKeys.length - 1}
          />
        )
      })}
      <td className="week-path-week-total">
        <div className="week-path-week-total-num">{percent === null ? '—' : `${percent}%`}</div>
        <div className="week-path-week-total-sub">{scheduled === 0 ? '予定なし' : `${done}/${scheduled}`}</div>
        <div className="week-path-bar">
          <div className="week-path-bar-fill" style={{ width: `${percent ?? 0}%` }} />
        </div>
      </td>
    </tr>
  )
}

function DayCell({
  state,
  content,
  isFirst,
  isLast,
}: {
  state: NodeState
  content: string
  isFirst: boolean
  isLast: boolean
}) {
  const segFilled = state === 'done' || state === 'milestone'
  return (
    <td className="week-path-cell">
      <div
        className={`week-path-seg${segFilled ? ' filled' : ''}`}
        style={{ left: isFirst ? '50%' : 0, right: isLast ? '50%' : 0 }}
      />
      <NodeIcon state={state} />
      {content && state !== 'no-task' && <div className="week-path-caption">{content}</div>}
    </td>
  )
}

function NodeIcon({ state }: { state: NodeState }) {
  switch (state) {
    case 'done':
      return (
        <span className="week-path-node node-done" aria-label="達成済み">
          ✓
        </span>
      )
    case 'missed':
      return (
        <span className="week-path-node node-missed" aria-label="未達成">
          ✕
        </span>
      )
    case 'today-pending':
      return (
        <span className="week-path-node node-today" aria-label="本日・未達成">
          <span className="week-path-node-dot" />
        </span>
      )
    case 'milestone':
      return (
        <span className="week-path-node node-milestone" aria-label="達成予定日">
          ★
        </span>
      )
    case 'upcoming':
      return (
        <span className="week-path-node node-upcoming" aria-label="予定">
          <span className="week-path-node-dot" />
        </span>
      )
    case 'no-task':
    default:
      return <span className="week-path-node-tick" aria-hidden="true" />
  }
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
