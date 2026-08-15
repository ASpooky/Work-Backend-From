import { useEffect, useState } from 'react'
import { api, type DailyTask, type GoalCalendar } from '../api'
import { startOfPeriod, endOfPeriod, shiftPeriod, formatISODate, formatPeriodLabel, eachDate } from '../date'
import { toUserMessage } from '../errors'
import WeekPathView from './WeekPathView'
import GoalProgressList from './GoalProgressList'
import './CalendarView.css'

type Props = {
  workspaceId: string
  refreshKey?: number
}

function CalendarView({ workspaceId, refreshKey }: Props) {
  const [anchor, setAnchor] = useState(() => startOfPeriod('week', new Date()))
  const [dayTasks, setDayTasks] = useState<DailyTask[]>([])
  const [calendar, setCalendar] = useState<GoalCalendar[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api
      .listDailyTasks(formatISODate(new Date()))
      .then(setDayTasks)
      .catch((err) => setError(toUserMessage(err)))
  }, [refreshKey])

  useEffect(() => {
    const from = startOfPeriod('week', anchor)
    const to = endOfPeriod('week', anchor)
    api
      .getCalendar(workspaceId, formatISODate(from), formatISODate(to))
      .then(setCalendar)
      .catch((err) => setError(toUserMessage(err)))
  }, [anchor, workspaceId, refreshKey])

  const from = startOfPeriod('week', anchor)
  const to = endOfPeriod('week', anchor)
  const days = eachDate(from, to)

  async function handleToggleDone(task: DailyTask) {
    const nextDone = !task.done
    setDayTasks((prev) => prev.map((t) => (t.id === task.id ? { ...t, done: nextDone } : t)))
    try {
      await api.updateDailyTaskDone(task.id, nextDone)
      api
        .getCalendar(workspaceId, formatISODate(from), formatISODate(to))
        .then(setCalendar)
        .catch((err) => setError(toUserMessage(err)))
    } catch (err) {
      setDayTasks((prev) => prev.map((t) => (t.id === task.id ? { ...t, done: !nextDone } : t)))
      setError(toUserMessage(err))
    }
  }

  return (
    <section className="calendar">
      <div className="calendar-controls">
        <div className="period-nav">
          <button
            type="button"
            aria-label="前の週"
            onClick={() => setAnchor((prev) => shiftPeriod('week', prev, -1))}
          >
            &lt;
          </button>
          <span className="period-label">{formatPeriodLabel('week', anchor)}</span>
          <button
            type="button"
            aria-label="次の週"
            onClick={() => setAnchor((prev) => shiftPeriod('week', prev, 1))}
          >
            &gt;
          </button>
        </div>
      </div>

      {error && <p className="error">{error}</p>}

      <WeekPathView calendar={calendar} days={days} />

      <div className="today-task-list">
        <h3>今日のタスク</h3>
        <ul className="task-list">
          {dayTasks.map((task) => (
            <li key={task.id}>
              <label className="task-item">
                <input type="checkbox" checked={task.done} onChange={() => handleToggleDone(task)} />
                <span className={task.done ? 'task-item-done' : ''}>{task.content}</span>
              </label>
            </li>
          ))}
          {dayTasks.length === 0 && <li>今日のタスクはありません</li>}
        </ul>
      </div>

      <GoalProgressList calendar={calendar} />
    </section>
  )
}

export default CalendarView
