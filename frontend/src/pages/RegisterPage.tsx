import { useEffect, useState, type SubmitEvent } from 'react'
import { api, type Goal, type Workspace } from '../api'
import { toUserMessage } from '../errors'
import { daysUntil, WEEKDAY_LABELS_JA } from '../date'

function todayISO(): string {
  const now = new Date()
  const yyyy = now.getFullYear()
  const mm = String(now.getMonth() + 1).padStart(2, '0')
  const dd = String(now.getDate()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd}`
}

type RecurrenceMode = 'none' | 'interval' | 'weekly'

type Props = {
  workspace: Workspace
  goals: Goal[]
  onCreated: () => void
}

function RegisterPage({ workspace, goals, onCreated }: Props) {
  const [error, setError] = useState<string | null>(null)

  const [goalTitle, setGoalTitle] = useState('')
  const [goalDetail, setGoalDetail] = useState('')
  const [goalCondition, setGoalCondition] = useState('')
  const [goalEndDate, setGoalEndDate] = useState('')
  const [goalMode, setGoalMode] = useState<'strict' | 'want'>('strict')

  const [taskGoalId, setTaskGoalId] = useState(() => localStorage.getItem('goal-tracker:last-task-goal') ?? '')
  const [taskContent, setTaskContent] = useState('')
  const [recurrenceMode, setRecurrenceMode] = useState<RecurrenceMode>('none')
  const [intervalDays, setIntervalDays] = useState(1)
  const [weekdays, setWeekdays] = useState<number[]>([])
  const [startDate, setStartDate] = useState(todayISO())
  const [endDate, setEndDate] = useState('')

  const today = todayISO()

  useEffect(() => {
    if (!taskGoalId || endDate) return
    const goal = goals.find((g) => g.id === taskGoalId)
    if (goal) setEndDate(goal.end_date.slice(0, 10))
  }, [taskGoalId, goals, endDate])

  function toggleWeekday(day: number) {
    setWeekdays((prev) => (prev.includes(day) ? prev.filter((d) => d !== day) : [...prev, day].sort()))
  }

  async function handleCreateGoal(e: SubmitEvent<HTMLFormElement>) {
    e.preventDefault()
    try {
      await api.createGoal({
        workspace_id: workspace.id,
        title: goalTitle,
        detail: goalDetail,
        achievement_condition: goalCondition,
        end_date: goalEndDate,
        mode: goalMode,
      })
      setGoalTitle('')
      setGoalDetail('')
      setGoalCondition('')
      setGoalEndDate('')
      onCreated()
    } catch (err) {
      setError(toUserMessage(err))
    }
  }

  async function handleCreateTask(e: SubmitEvent<HTMLFormElement>) {
    e.preventDefault()
    if (!taskGoalId) return
    if (recurrenceMode === 'weekly' && weekdays.length === 0) {
      setError('曜日を1つ以上選択してください')
      return
    }

    try {
      if (recurrenceMode === 'none') {
        await api.createDailyTask({ goal_id: taskGoalId, date: today, content: taskContent })
      } else {
        await api.createRecurringDailyTasks({
          goal_id: taskGoalId,
          content: taskContent,
          start_date: startDate,
          end_date: endDate,
          rule:
            recurrenceMode === 'interval'
              ? { type: 'interval', interval_days: intervalDays }
              : { type: 'weekly', weekdays },
        })
      }
      localStorage.setItem('goal-tracker:last-task-goal', taskGoalId)
      setTaskContent('')
      onCreated()
    } catch (err) {
      setError(toUserMessage(err))
    }
  }

  return (
    <>
      {error && <p className="error">{error}</p>}

      <section>
        <h2>タスクを追加</h2>
        <form onSubmit={handleCreateTask}>
          <label>
            目標
            <select value={taskGoalId} onChange={(e) => setTaskGoalId(e.target.value)} required>
              <option value="">目標を選択</option>
              {goals.map((goal) => {
                const left = daysUntil(goal.end_date)
                return (
                  <option key={goal.id} value={goal.id}>
                    {goal.title} ・ {left >= 0 ? `あと${left}日` : '期限切れ'}
                  </option>
                )
              })}
            </select>
          </label>
          <label>
            内容
            <input
              placeholder="今日は何をする？"
              value={taskContent}
              onChange={(e) => setTaskContent(e.target.value)}
              required
            />
          </label>

          <div className="recurrence-picker">
            <div className="recurrence-mode-options">
              <label className="settings-radio">
                <input
                  type="radio"
                  name="recurrence"
                  checked={recurrenceMode === 'none'}
                  onChange={() => setRecurrenceMode('none')}
                />
                今日だけ
              </label>
              <label className="settings-radio">
                <input
                  type="radio"
                  name="recurrence"
                  checked={recurrenceMode === 'interval'}
                  onChange={() => setRecurrenceMode('interval')}
                />
                間隔で繰り返す
              </label>
              <label className="settings-radio">
                <input
                  type="radio"
                  name="recurrence"
                  checked={recurrenceMode === 'weekly'}
                  onChange={() => setRecurrenceMode('weekly')}
                />
                曜日を指定
              </label>
            </div>

            {recurrenceMode === 'interval' && (
              <label className="recurrence-detail">
                間隔
                <div className="interval-input">
                  <input
                    type="number"
                    min={1}
                    value={intervalDays}
                    onChange={(e) => setIntervalDays(Math.max(1, Number(e.target.value) || 1))}
                  />
                  <span>
                    日おき（{intervalDays === 1 ? '毎日' : intervalDays === 2 ? '隔日' : `${intervalDays}日ごと`}）
                  </span>
                </div>
              </label>
            )}

            {recurrenceMode === 'weekly' && (
              <div className="recurrence-detail">
                曜日
                <div className="weekday-picker">
                  {WEEKDAY_LABELS_JA.map((label, i) => (
                    <button
                      key={label}
                      type="button"
                      className={weekdays.includes(i) ? 'active' : ''}
                      onClick={() => toggleWeekday(i)}
                    >
                      {label}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {recurrenceMode !== 'none' && (
              <div className="recurrence-dates">
                <label>
                  開始日
                  <input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} required />
                </label>
                <label>
                  終了日
                  <input type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} required />
                </label>
              </div>
            )}
          </div>

          <button type="submit">タスクを追加</button>
        </form>
        {goals.length === 0 && <p className="hint">先に下で目標を作成してください。</p>}
      </section>

      <section>
        <h2>新しい目標</h2>
        <form onSubmit={handleCreateGoal}>
          <label>
            タイトル
            <input
              placeholder="例: 5km走る"
              value={goalTitle}
              onChange={(e) => setGoalTitle(e.target.value)}
              required
            />
          </label>
          <label>
            詳細
            <input
              placeholder="例: 毎日のランニング習慣"
              value={goalDetail}
              onChange={(e) => setGoalDetail(e.target.value)}
              required
            />
          </label>
          <label>
            達成条件
            <input
              placeholder="例: 5km以上走る"
              value={goalCondition}
              onChange={(e) => setGoalCondition(e.target.value)}
              required
            />
          </label>
          <label>
            期限
            <input type="date" value={goalEndDate} onChange={(e) => setGoalEndDate(e.target.value)} required />
          </label>
          <label>
            モード
            <select value={goalMode} onChange={(e) => setGoalMode(e.target.value as 'strict' | 'want')}>
              <option value="strict">必達</option>
              <option value="want">努力目標</option>
            </select>
          </label>
          <button type="submit">目標を作成</button>
        </form>
      </section>
    </>
  )
}

export default RegisterPage
