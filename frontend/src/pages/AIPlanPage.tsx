import { useEffect, useState, type SubmitEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type Conversation, type ConversationMessage, type Plan, type PlannedGoal, type Workspace } from '../api'
import { toUserMessage } from '../errors'
import { describeRecurrence } from '../date'
import './AIPlanPage.css'

type Props = {
  workspace: Workspace
  isAllWorkspaces: boolean
  onCreated: () => void
}

function AIPlanPage({ workspace, isAllWorkspaces, onCreated }: Props) {
  const navigate = useNavigate()

  const [aiEnabled, setAiEnabled] = useState<boolean | null>(null)
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [activeConversationId, setActiveConversationId] = useState<string | null>(null)
  const [messages, setMessages] = useState<ConversationMessage[]>([])
  const [input, setInput] = useState('')
  const [sending, setSending] = useState(false)
  const [planning, setPlanning] = useState(false)
  const [committing, setCommitting] = useState(false)
  const [plan, setPlan] = useState<Plan | null>(null)
  const [editedGoals, setEditedGoals] = useState<PlannedGoal[]>([])
  // Removed task indexes per goal, same shape/order as plan.goals.
  const [removedTaskIndexes, setRemovedTaskIndexes] = useState<number[][]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api
      .aiStatus()
      .then((s) => setAiEnabled(s.enabled))
      .catch(() => setAiEnabled(false))
  }, [])

  useEffect(() => {
    setActiveConversationId(null)
    setMessages([])
    setPlan(null)

    if (aiEnabled !== true) return
    api
      .listConversations(workspace.id)
      .then(setConversations)
      .catch((err) => setError(toUserMessage(err)))
  }, [aiEnabled, workspace.id])

  useEffect(() => {
    if (plan) {
      setEditedGoals(plan.goals.map((g) => g.goal))
      setRemovedTaskIndexes(plan.goals.map(() => []))
    }
  }, [plan])

  function updateEditedGoal(goalIndex: number, patch: Partial<PlannedGoal>) {
    setEditedGoals((prev) => prev.map((g, i) => (i === goalIndex ? { ...g, ...patch } : g)))
  }

  function toggleTaskRemoved(goalIndex: number, taskIndex: number) {
    setRemovedTaskIndexes((prev) =>
      prev.map((indexes, i) =>
        i !== goalIndex ? indexes : indexes.includes(taskIndex) ? indexes.filter((x) => x !== taskIndex) : [...indexes, taskIndex],
      ),
    )
  }

  function handleNewChat() {
    setActiveConversationId(null)
    setMessages([])
    setPlan(null)
    setError(null)
  }

  async function handleSelectConversation(id: string) {
    setActiveConversationId(id)
    setPlan(null)
    setError(null)
    try {
      setMessages(await api.getConversation(id))
    } catch (err) {
      setError(toUserMessage(err))
    }
  }

  async function handleSend(e: SubmitEvent<HTMLFormElement>) {
    e.preventDefault()
    if (!input.trim() || sending) return

    const content = input
    setMessages((prev) => [
      ...prev,
      {
        id: `pending-${Date.now()}`,
        conversation_id: activeConversationId ?? '',
        role: 'user',
        content,
        created_at: new Date().toISOString(),
      },
    ])
    setInput('')
    setSending(true)
    try {
      const { conversation_id, reply } = await api.sendChatMessage(workspace.id, activeConversationId, content)
      setActiveConversationId(conversation_id)
      setMessages((prev) => [
        ...prev,
        {
          id: `reply-${Date.now()}`,
          conversation_id,
          role: 'model',
          content: reply,
          created_at: new Date().toISOString(),
        },
      ])
      setConversations(await api.listConversations(workspace.id))
    } catch (err) {
      setError(toUserMessage(err))
    } finally {
      setSending(false)
    }
  }

  async function handleGeneratePlan() {
    if (!activeConversationId) return
    setPlanning(true)
    setError(null)
    try {
      setPlan(await api.aiPlan(workspace.id, activeConversationId))
    } catch (err) {
      setError(toUserMessage(err))
    } finally {
      setPlanning(false)
    }
  }

  async function handleConfirmPlan() {
    if (!plan || editedGoals.length === 0) return
    setCommitting(true)
    setError(null)
    try {
      for (let gi = 0; gi < plan.goals.length; gi++) {
        const editedGoal = editedGoals[gi]
        const goal = await api.createGoal({
          workspace_id: workspace.id,
          title: editedGoal.title,
          detail: editedGoal.detail,
          achievement_condition: editedGoal.achievement_condition,
          end_date: editedGoal.end_date,
          mode: editedGoal.mode,
        })

        const removed = removedTaskIndexes[gi] ?? []
        const tasks = plan.goals[gi].tasks.filter((_, ti) => !removed.includes(ti))
        for (const task of tasks) {
          await api.createRecurringDailyTasks({
            goal_id: goal.id,
            content: task.content,
            start_date: task.start_date,
            end_date: task.end_date,
            rule:
              task.rule_type === 'interval'
                ? { type: 'interval', interval_days: task.interval_days ?? 1 }
                : { type: 'weekly', weekdays: task.weekdays ?? [] },
          })
        }
      }

      onCreated()
      navigate('/')
    } catch (err) {
      setError(toUserMessage(err))
      setCommitting(false)
    }
  }

  if (isAllWorkspaces) {
    return (
      <section>
        <h2>AIと目標を作る</h2>
        <p className="hint">「すべて」表示では利用できません。サイドバーで特定のワークスペースを選んでください。</p>
      </section>
    )
  }

  if (aiEnabled === false) {
    return (
      <section>
        <h2>AIと目標を作る</h2>
        <p className="hint">
          AI機能はサーバー側の設定が済んでいないため、現在利用できません。
        </p>
      </section>
    )
  }

  return (
    <section className="ai-plan-page">
      <h2>AIと目標を作る</h2>

      {error && <p className="error">{error}</p>}

      <div className="ai-plan-layout">
        <aside className="ai-conversation-list">
          <button type="button" className="ai-new-chat" onClick={handleNewChat}>
            + 新しい会話
          </button>
          <ul>
            {conversations.map((c) => (
              <li key={c.id}>
                <button
                  type="button"
                  className={c.id === activeConversationId ? 'active' : ''}
                  onClick={() => handleSelectConversation(c.id)}
                >
                  {c.title}
                </button>
              </li>
            ))}
            {conversations.length === 0 && <li className="ai-conversation-empty">まだ会話がありません。</li>}
          </ul>
        </aside>

        <div className="ai-chat-pane">
          {!plan && (
            <>
              <div className="ai-chat-log">
                {messages.length === 0 && (
                  <p className="hint">達成したいことを話しかけてみてください。強度や期限を一緒に詰めていきます。</p>
                )}
                {messages.map((m) => (
                  <div key={m.id} className={`ai-chat-bubble ${m.role}`}>
                    {m.content}
                  </div>
                ))}
                {sending && <div className="ai-chat-bubble model pending">…</div>}
              </div>

              <form className="ai-chat-form" onSubmit={handleSend}>
                <input
                  placeholder="例: 3ヶ月後のフルマラソンに向けて走り込みたい"
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  disabled={aiEnabled === null}
                />
                <button type="submit" disabled={sending || aiEnabled === null}>
                  送信
                </button>
              </form>

              {messages.length > 0 && (
                <button type="button" onClick={handleGeneratePlan} disabled={planning} className="ai-plan-generate">
                  {planning ? 'プランを作成中…' : 'この内容でタスクに落とし込む'}
                </button>
              )}
            </>
          )}

          {plan && editedGoals.length > 0 && (
            <div className="ai-plan-review">
              <h3>
                {plan.goals.length > 1
                  ? `提案されたプラン(${plan.goals.length}件のフェーズ・確認・編集してから登録してください)`
                  : '提案されたプラン(確認・編集してから登録してください)'}
              </h3>

              {plan.goals.map((goalPlan, gi) => {
                const editedGoal = editedGoals[gi]
                const removedForGoal = removedTaskIndexes[gi] ?? []
                return (
                  <div key={gi} className="ai-plan-phase">
                    {plan.goals.length > 1 && <div className="ai-plan-phase-label">フェーズ {gi + 1}</div>}

                    <div className="ai-plan-goal-fields">
                      <label>
                        タイトル
                        <input
                          value={editedGoal.title}
                          onChange={(e) => updateEditedGoal(gi, { title: e.target.value })}
                        />
                      </label>
                      <label>
                        詳細
                        <input
                          value={editedGoal.detail}
                          onChange={(e) => updateEditedGoal(gi, { detail: e.target.value })}
                        />
                      </label>
                      <label>
                        達成条件
                        <input
                          value={editedGoal.achievement_condition}
                          onChange={(e) => updateEditedGoal(gi, { achievement_condition: e.target.value })}
                        />
                      </label>
                      <label>
                        期限
                        <input
                          type="date"
                          value={editedGoal.end_date}
                          onChange={(e) => updateEditedGoal(gi, { end_date: e.target.value })}
                        />
                      </label>
                      <label>
                        モード
                        <select
                          value={editedGoal.mode}
                          onChange={(e) => updateEditedGoal(gi, { mode: e.target.value as 'strict' | 'want' })}
                        >
                          <option value="strict">必達</option>
                          <option value="want">努力目標</option>
                        </select>
                      </label>
                    </div>

                    <h3>日々のタスク</h3>
                    <ul className="ai-plan-task-list">
                      {goalPlan.tasks.map((task, ti) => {
                        const removed = removedForGoal.includes(ti)
                        return (
                          <li key={ti} className={removed ? 'removed' : ''}>
                            <div>
                              <div className="ai-plan-task-content">{task.content}</div>
                              <div className="ai-plan-task-sub">
                                {describeRecurrence(task)} ・ {task.start_date} 〜 {task.end_date}
                              </div>
                            </div>
                            <button type="button" onClick={() => toggleTaskRemoved(gi, ti)}>
                              {removed ? '戻す' : '削除'}
                            </button>
                          </li>
                        )
                      })}
                    </ul>
                  </div>
                )
              })}

              <div className="ai-plan-actions">
                <button type="button" onClick={() => setPlan(null)}>
                  会話に戻る
                </button>
                <button type="button" onClick={handleConfirmPlan} disabled={committing}>
                  {committing
                    ? '登録中…'
                    : plan.goals.length > 1
                      ? `この内容で${plan.goals.length}件登録する`
                      : 'この内容で登録する'}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </section>
  )
}

export default AIPlanPage
