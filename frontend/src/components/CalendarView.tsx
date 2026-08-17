import { useEffect, useState } from 'react'
import { api, type DailyTask, type GoalCalendar, type Workspace } from '../api'
import {
  startOfPeriod,
  endOfPeriod,
  shiftPeriod,
  formatISODate,
  formatPeriodLabel,
  formatShortDateJa,
  eachDate,
} from '../date'
import { toUserMessage } from '../errors'
import WeekPathView from './WeekPathView'
import './CalendarView.css'

type Props = {
  workspaceId: string
  workspaces?: Workspace[]
  refreshKey?: number
}

function CalendarView({ workspaceId, workspaces, refreshKey }: Props) {
  const [anchor, setAnchor] = useState(() => startOfPeriod('week', new Date()))
  const [selectedDate, setSelectedDate] = useState(() => formatISODate(new Date()))
  const [dayTasks, setDayTasks] = useState<DailyTask[]>([])
  const [calendar, setCalendar] = useState<GoalCalendar[]>([])
  const [error, setError] = useState<string | null>(null)
  const [expandedMemoTaskId, setExpandedMemoTaskId] = useState<string | null>(null)

  const isToday = selectedDate === formatISODate(new Date())

  useEffect(() => {
    api
      .listDailyTasks(selectedDate)
      .then(setDayTasks)
      .catch((err) => setError(toUserMessage(err)))
  }, [selectedDate, refreshKey])

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
  const modeByGoalId = new Map(calendar.map((gc) => [gc.goal.id, gc.goal.mode]))
  const workspaceNameById = new Map((workspaces ?? []).map((w) => [w.id, w.name]))
  // Incomplete tasks first so you can see what's left without hunting through
  // done ones; stable within each group (no re-sort beyond done/not-done).
  const sortedDayTasks = [...dayTasks].sort((a, b) => Number(a.done) - Number(b.done))
  const doneCount = dayTasks.filter((t) => t.done).length

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

  async function handleSaveMemo(task: DailyTask, memo: string) {
    const trimmed = memo.trim()
    const nextMemo = trimmed === '' ? undefined : trimmed
    if (nextMemo === task.memo) return
    setDayTasks((prev) => prev.map((t) => (t.id === task.id ? { ...t, memo: nextMemo } : t)))
    try {
      await api.updateTaskMemo(task.id, nextMemo ?? null)
    } catch (err) {
      setDayTasks((prev) => prev.map((t) => (t.id === task.id ? { ...t, memo: task.memo } : t)))
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

      <WeekPathView
        calendar={calendar}
        days={days}
        workspaceNameById={workspaceId === '' ? workspaceNameById : undefined}
        selectedDateKey={selectedDate}
        onSelectDate={setSelectedDate}
      />

      <div className="today-task-list">
        <div className="today-task-list-header">
          <h3>{isToday ? '今日のタスク' : `${formatShortDateJa(selectedDate)}のタスク`}</h3>
          <div className="today-task-list-header-right">
            {dayTasks.length > 0 && (
              <span className="today-task-count">
                {doneCount}/{dayTasks.length}
              </span>
            )}
            {!isToday && (
              <button type="button" className="today-task-reset" onClick={() => setSelectedDate(formatISODate(new Date()))}>
                今日に戻る
              </button>
            )}
          </div>
        </div>
        <ul className="task-list">
          {sortedDayTasks.map((task) => {
            const memoOpen = expandedMemoTaskId === task.id
            return (
              <li key={task.id}>
                <div className="task-item-row">
                  <label className="task-item">
                    <input type="checkbox" checked={task.done} onChange={() => handleToggleDone(task)} />
                    {modeByGoalId.get(task.goal_id) === 'strict' && (
                      <span className="task-item-strict" title="必達">
                        必達
                      </span>
                    )}
                    <span className={task.done ? 'task-item-done' : ''}>{task.content}</span>
                  </label>
                  <button
                    type="button"
                    className={`task-item-memo-toggle${task.memo ? ' has-memo' : ''}`}
                    onClick={() => setExpandedMemoTaskId(memoOpen ? null : task.id)}
                    aria-label={`${task.content}のメモを${memoOpen ? '閉じる' : '開く'}`}
                  >
                    {task.memo ? '📝' : '+メモ'}
                  </button>
                </div>
                {memoOpen && (
                  <textarea
                    className="task-item-memo-input"
                    placeholder="メモ（体調・気づきなど）"
                    defaultValue={task.memo ?? ''}
                    onBlur={(e) => handleSaveMemo(task, e.target.value)}
                  />
                )}
              </li>
            )
          })}
          {dayTasks.length === 0 && <li>{isToday ? '今日のタスクはありません。' : 'この日のタスクはありません。'}</li>}
        </ul>
      </div>
    </section>
  )
}

export default CalendarView
