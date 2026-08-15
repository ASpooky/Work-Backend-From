import type { GoalCalendar } from '../api'

type Props = {
  calendar: GoalCalendar[]
}

function weekProgress(gc: GoalCalendar): number | null {
  const entries = Object.values(gc.days)
  const scheduled = entries.filter((d) => d.status !== 'no_task')
  if (scheduled.length === 0) return null

  const done = scheduled.filter((d) => d.status === 'done')
  return Math.round((done.length / scheduled.length) * 100)
}

function GoalProgressList({ calendar }: Props) {
  return (
    <div className="goal-progress-list">
      <h3>今週の進捗</h3>
      <ul>
        {calendar.map((gc) => {
          const percent = weekProgress(gc)
          return (
            <li key={gc.goal.id} className="goal-progress-row">
              <span className="goal-progress-title">{gc.goal.title}</span>
              <div className="goal-progress-bar">
                <div className="goal-progress-bar-fill" style={{ width: `${percent ?? 0}%` }} />
              </div>
              <span className="goal-progress-percent">{percent === null ? '—' : `${percent}%`}</span>
            </li>
          )
        })}
        {calendar.length === 0 && <li className="goal-progress-empty">まだ目標がありません</li>}
      </ul>
    </div>
  )
}

export default GoalProgressList
