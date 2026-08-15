import { useState, type SubmitEvent } from 'react'
import { api, type Goal, type Workspace } from '../api'

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

  const [taskGoalId, setTaskGoalId] = useState('')
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
      setError(String(err))
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
      setTaskContent('')
      onCreated()
    } catch (err) {
      setError(String(err))
    }
  }

  return (
    <>
      {error && <p className="error">{error}</p>}

      <section>
        <h2>新しい目標</h2>
        <form onSubmit={handleCreateGoal}>
          <input
            placeholder="Title"
            value={goalTitle}
            onChange={(e) => setGoalTitle(e.target.value)}
            required
          />
          <input
            placeholder="Detail"
            value={goalDetail}
            onChange={(e) => setGoalDetail(e.target.value)}
            required
          />
          <input
            placeholder="Achievement condition"
            value={goalCondition}
            onChange={(e) => setGoalCondition(e.target.value)}
            required
          />
          <input
            type="date"
            value={goalEndDate}
            onChange={(e) => setGoalEndDate(e.target.value)}
            required
          />
          <select value={goalMode} onChange={(e) => setGoalMode(e.target.value as 'strict' | 'want')}>
            <option value="strict">strict</option>
            <option value="want">want</option>
          </select>
          <button type="submit">Create Goal</button>
        </form>
      </section>

      <section>
        <h2>今日のタスクを追加 ({today})</h2>
        <form onSubmit={handleCreateTask}>
          <select value={taskGoalId} onChange={(e) => setTaskGoalId(e.target.value)} required>
            <option value="">Select a goal</option>
            {goals.map((goal) => (
              <option key={goal.id} value={goal.id}>
                {goal.title}
              </option>
            ))}
          </select>
          <input
            placeholder="What will you do today?"
            value={taskContent}
            onChange={(e) => setTaskContent(e.target.value)}
            required
          />
          <button type="submit">Add Task</button>
        </form>
      </section>
    </>
  )
}

export default RegisterPage
