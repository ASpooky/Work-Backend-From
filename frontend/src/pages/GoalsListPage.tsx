import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type Goal, type Workspace } from '../api'
import { toUserMessage } from '../errors'
import { daysUntil } from '../date'
import './GoalsListPage.css'

type Props = {
  workspace: Workspace
  workspaces: Workspace[]
  refreshKey: number
  onDeleted: () => void
}

function GoalsListPage({ workspace, workspaces, refreshKey, onDeleted }: Props) {
  const [goals, setGoals] = useState<Goal[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api.listGoals(workspace.id).then(setGoals).catch((err) => setError(toUserMessage(err)))
  }, [workspace.id, refreshKey])

  const workspaceNameById = new Map(workspaces.map((w) => [w.id, w.name]))
  const isAllWorkspaces = workspace.id === ''

  async function handleDelete(goal: Goal) {
    if (!window.confirm(`「${goal.title}」を削除しますか？関連するタスクやAI会話もすべて削除されます。`)) return
    try {
      await api.deleteGoal(goal.id)
      setGoals((prev) => prev.filter((g) => g.id !== goal.id))
      onDeleted()
    } catch (err) {
      setError(toUserMessage(err))
    }
  }

  // Swaps the two goals' priorities to their new positions' indexes. This
  // self-heals ordering over time even for goals that still share the
  // migration default priority of 0 (only ties on created_at until touched).
  async function handleMove(index: number, direction: -1 | 1) {
    const targetIndex = index + direction
    if (targetIndex < 0 || targetIndex >= goals.length) return

    const current = goals[index]
    const target = goals[targetIndex]

    try {
      await Promise.all([api.reorderGoal(current.id, targetIndex), api.reorderGoal(target.id, index)])
      setGoals((prev) => {
        const next = [...prev]
        next[index] = { ...target, priority: index }
        next[targetIndex] = { ...current, priority: targetIndex }
        return next
      })
    } catch (err) {
      setError(toUserMessage(err))
    }
  }

  return (
    <section>
      <h2>目標</h2>

      {error && <p className="error">{error}</p>}

      {goals.length === 0 && <p className="hint">まだ目標がありません。「登録」から作成してください。</p>}

      <ul className="goals-list">
        {goals.map((goal, index) => {
          const left = daysUntil(goal.end_date)
          return (
            <li key={goal.id} className="goals-list-item">
              {!isAllWorkspaces && (
                <div className="goals-list-reorder">
                  <button
                    type="button"
                    onClick={() => handleMove(index, -1)}
                    disabled={index === 0}
                    aria-label={`${goal.title}を上に移動`}
                  >
                    ▲
                  </button>
                  <button
                    type="button"
                    onClick={() => handleMove(index, 1)}
                    disabled={index === goals.length - 1}
                    aria-label={`${goal.title}を下に移動`}
                  >
                    ▼
                  </button>
                </div>
              )}
              <Link to={`/goals/${goal.id}`} className="goals-list-card">
                <div className="goals-list-card-header">
                  <span className="goals-list-card-title">{goal.title}</span>
                  <span className={`goals-list-card-mode ${goal.mode}`}>
                    {goal.mode === 'strict' ? '必達' : '努力目標'}
                  </span>
                </div>
                <p className="goals-list-card-condition">{goal.achievement_condition}</p>
                <div className="goals-list-card-footer">
                  <span>{left >= 0 ? `あと${left}日` : '期限切れ'}</span>
                  {isAllWorkspaces && (
                    <span className="goals-list-card-workspace">{workspaceNameById.get(goal.workspace_id)}</span>
                  )}
                </div>
              </Link>
              <button
                type="button"
                className="goals-list-delete"
                onClick={() => handleDelete(goal)}
                aria-label={`${goal.title}を削除`}
              >
                🗑
              </button>
            </li>
          )
        })}
      </ul>
    </section>
  )
}

export default GoalsListPage
