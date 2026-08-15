import { useState, type SubmitEvent } from 'react'
import { api, type Goal, type Workspace } from '../api'
import { toUserMessage } from '../errors'

function todayISO(): string {
  const now = new Date()
  const yyyy = now.getFullYear()
  const mm = String(now.getMonth() + 1).padStart(2, '0')
  const dd = String(now.getDate()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd}`
}

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

  const today = todayISO()

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
    try {
      await api.createDailyTask({
        goal_id: taskGoalId,
        date: today,
        content: taskContent,
      })
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
        <h2>今日のタスクを追加 ({today})</h2>
        <form onSubmit={handleCreateTask}>
          <label>
            目標
            <select value={taskGoalId} onChange={(e) => setTaskGoalId(e.target.value)} required>
              <option value="">目標を選択</option>
              {goals.map((goal) => (
                <option key={goal.id} value={goal.id}>
                  {goal.title}
                </option>
              ))}
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
