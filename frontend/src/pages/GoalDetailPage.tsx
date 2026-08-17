import { useEffect, useState, type SubmitEvent } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { api, type ConversationMessage, type GoalRevision, type GoalStats, type PlannedGoal } from '../api'
import { toUserMessage } from '../errors'
import { describeRecurrence } from '../date'
import './AIPlanPage.css'
import './GoalDetailPage.css'

type Props = {
  onUpdated: () => void
}

type Mode = 'view' | 'edit' | 'ai'

function GoalDetailPage({ onUpdated }: Props) {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [stats, setStats] = useState<GoalStats | null>(null)
  const [notFound, setNotFound] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [mode, setMode] = useState<Mode>('view')
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)

  const [title, setTitle] = useState('')
  const [detail, setDetail] = useState('')
  const [condition, setCondition] = useState('')
  const [endDate, setEndDate] = useState('')
  const [goalMode, setGoalMode] = useState<'strict' | 'want'>('strict')

  const [aiEnabled, setAiEnabled] = useState<boolean | null>(null)
  const [summary, setSummary] = useState<string | null>(null)
  const [summaryLoading, setSummaryLoading] = useState(false)

  // AI review chat state (separate from the manual edit form above).
  const [reviewConversationId, setReviewConversationId] = useState<string | null>(null)
  const [reviewMessages, setReviewMessages] = useState<ConversationMessage[]>([])
  const [reviewInput, setReviewInput] = useState('')
  const [reviewSending, setReviewSending] = useState(false)
  const [revising, setRevising] = useState(false)
  const [revision, setRevision] = useState<GoalRevision | null>(null)
  const [editedRevisionGoal, setEditedRevisionGoal] = useState<PlannedGoal | null>(null)
  // Task ids from revision.removed_tasks the user un-flagged (i.e. wants to
  // keep after all, undoing the AI's proposed removal).
  const [keptTaskIds, setKeptTaskIds] = useState<string[]>([])
  // Indexes into revision.new_tasks the user doesn't want to add.
  const [removedNewTaskIndexes, setRemovedNewTaskIndexes] = useState<number[]>([])
  const [confirmingRevision, setConfirmingRevision] = useState(false)

  function load() {
    if (!id) return
    api
      .getGoal(id)
      .then((s) => {
        setStats(s)
        setTitle(s.goal.title)
        setDetail(s.goal.detail)
        setCondition(s.goal.achievement_condition)
        setEndDate(s.goal.end_date.slice(0, 10))
        setGoalMode(s.goal.mode)
      })
      .catch((err) => {
        if (err instanceof Error && /404/.test(err.message)) {
          setNotFound(true)
          return
        }
        setError(toUserMessage(err))
      })
  }

  useEffect(load, [id])

  useEffect(() => {
    api
      .aiStatus()
      .then((s) => setAiEnabled(s.enabled))
      .catch(() => setAiEnabled(false))
  }, [])

  useEffect(() => {
    if (revision) {
      setEditedRevisionGoal(revision.goal)
      setKeptTaskIds([])
      setRemovedNewTaskIndexes([])
    }
  }, [revision])

  async function handleSave(e: SubmitEvent<HTMLFormElement>) {
    e.preventDefault()
    if (!id) return
    setSaving(true)
    setError(null)
    try {
      await api.updateGoal(id, {
        title,
        detail,
        achievement_condition: condition,
        end_date: endDate,
        mode: goalMode,
      })
      setMode('view')
      load()
      onUpdated()
    } catch (err) {
      setError(toUserMessage(err))
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!id || !stats) return
    if (!window.confirm(`「${stats.goal.title}」を削除しますか？関連するタスクやAI会話もすべて削除されます。`)) return

    setDeleting(true)
    setError(null)
    try {
      await api.deleteGoal(id)
      onUpdated()
      navigate('/goals')
    } catch (err) {
      setError(toUserMessage(err))
      setDeleting(false)
    }
  }

  async function handleSummarize() {
    if (!id) return
    setSummaryLoading(true)
    setError(null)
    try {
      const res = await api.summarizeGoal(id)
      setSummary(res.summary)
    } catch (err) {
      setError(toUserMessage(err))
    } finally {
      setSummaryLoading(false)
    }
  }

  async function handleOpenAiReview() {
    setMode('ai')
    setRevision(null)
    setError(null)
    if (!id) return
    try {
      const conversations = await api.listGoalConversations(id)
      if (conversations.length > 0) {
        setReviewConversationId(conversations[0].id)
        setReviewMessages(await api.getConversation(conversations[0].id))
      } else {
        setReviewConversationId(null)
        setReviewMessages([])
      }
    } catch (err) {
      setError(toUserMessage(err))
    }
  }

  function handleNewReviewChat() {
    setReviewConversationId(null)
    setReviewMessages([])
    setRevision(null)
    setError(null)
  }

  async function handleSendReview(e: SubmitEvent<HTMLFormElement>) {
    e.preventDefault()
    if (!id || !stats || !reviewInput.trim() || reviewSending) return

    const content = reviewInput
    setReviewMessages((prev) => [
      ...prev,
      {
        id: `pending-${Date.now()}`,
        conversation_id: reviewConversationId ?? '',
        role: 'user',
        content,
        created_at: new Date().toISOString(),
      },
    ])
    setReviewInput('')
    setReviewSending(true)
    try {
      const { conversation_id, reply } = await api.sendGoalReviewMessage(
        id,
        stats.goal.workspace_id,
        reviewConversationId,
        content,
      )
      setReviewConversationId(conversation_id)
      setReviewMessages((prev) => [
        ...prev,
        {
          id: `reply-${Date.now()}`,
          conversation_id,
          role: 'model',
          content: reply,
          created_at: new Date().toISOString(),
        },
      ])
    } catch (err) {
      setError(toUserMessage(err))
    } finally {
      setReviewSending(false)
    }
  }

  async function handleGenerateRevision() {
    if (!id || !reviewConversationId) return
    setRevising(true)
    setError(null)
    try {
      const got = await api.reviseGoal(id, reviewConversationId)
      // Defensive: normalize in case either array ever comes back null
      // (a bare Go nil slice serializes to JSON null, not []).
      setRevision({ ...got, removed_tasks: got.removed_tasks ?? [], new_tasks: got.new_tasks ?? [] })
    } catch (err) {
      setError(toUserMessage(err))
    } finally {
      setRevising(false)
    }
  }

  async function handleConfirmRevision() {
    if (!id || !revision || !editedRevisionGoal) return
    setConfirmingRevision(true)
    setError(null)
    try {
      await api.updateGoal(id, {
        title: editedRevisionGoal.title,
        detail: editedRevisionGoal.detail,
        achievement_condition: editedRevisionGoal.achievement_condition,
        end_date: editedRevisionGoal.end_date,
        mode: editedRevisionGoal.mode,
      })

      const tasksToRemove = revision.removed_tasks.filter((t) => !keptTaskIds.includes(t.id))
      for (const task of tasksToRemove) {
        await api.deleteDailyTask(task.id)
      }

      const tasksToAdd = revision.new_tasks.filter((_, i) => !removedNewTaskIndexes.includes(i))
      for (const task of tasksToAdd) {
        await api.createRecurringDailyTasks({
          goal_id: id,
          content: task.content,
          start_date: task.start_date,
          end_date: task.end_date,
          rule:
            task.rule_type === 'interval'
              ? { type: 'interval', interval_days: task.interval_days ?? 1 }
              : { type: 'weekly', weekdays: task.weekdays ?? [] },
        })
      }

      setMode('view')
      setRevision(null)
      load()
      onUpdated()
    } catch (err) {
      setError(toUserMessage(err))
    } finally {
      setConfirmingRevision(false)
    }
  }

  if (notFound) {
    return (
      <section>
        <p className="error">目標が見つかりませんでした。</p>
        <button type="button" onClick={() => navigate('/')}>
          カレンダーに戻る
        </button>
      </section>
    )
  }

  if (!stats) {
    return <p className="hint">読み込み中…</p>
  }

  const goal = stats.goal
  const achievementPercent = Math.round(stats.achievement_rate * 100)

  return (
    <section className="goal-detail">
      <button type="button" className="goal-detail-back" onClick={() => navigate(-1)}>
        ← 戻る
      </button>

      {error && <p className="error">{error}</p>}

      {mode === 'edit' && (
        <form className="goal-detail-edit-form" onSubmit={handleSave}>
          <label>
            タイトル
            <input value={title} onChange={(e) => setTitle(e.target.value)} required />
          </label>
          <label>
            詳細
            <input value={detail} onChange={(e) => setDetail(e.target.value)} required />
          </label>
          <label>
            達成条件
            <input value={condition} onChange={(e) => setCondition(e.target.value)} required />
          </label>
          <label>
            期限
            <input type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} required />
          </label>
          <label>
            モード
            <select value={goalMode} onChange={(e) => setGoalMode(e.target.value as 'strict' | 'want')}>
              <option value="strict">必達</option>
              <option value="want">努力目標</option>
            </select>
          </label>
          <div className="goal-detail-edit-actions">
            <button type="submit" disabled={saving}>
              {saving ? '保存中…' : '保存'}
            </button>
            <button type="button" onClick={() => setMode('view')} disabled={saving}>
              キャンセル
            </button>
          </div>
        </form>
      )}

      {mode === 'ai' && (
        <div className="goal-review-panel">
          <div className="goal-review-header">
            <h3>AIと見直す</h3>
            <div className="goal-review-header-actions">
              <button type="button" onClick={handleNewReviewChat}>
                新しい会話
              </button>
              <button type="button" onClick={() => setMode('view')}>
                閉じる
              </button>
            </div>
          </div>

          {!revision && (
            <>
              <div className="goal-review-log">
                {reviewMessages.length === 0 && (
                  <p className="hint">
                    見直したい内容を話しかけてください（例: 「期限を1ヶ月延ばしたい」）。AIが現在の目標の内容と進捗を踏まえて相談に乗ります。
                  </p>
                )}
                {reviewMessages.map((m) => (
                  <div key={m.id} className={`ai-chat-bubble ${m.role}`}>
                    {m.content}
                  </div>
                ))}
                {reviewSending && <div className="ai-chat-bubble model pending">…</div>}
              </div>

              <form className="ai-chat-form" onSubmit={handleSendReview}>
                <input
                  placeholder="例: そろそろ厳しいので期限を延ばしたい"
                  value={reviewInput}
                  onChange={(e) => setReviewInput(e.target.value)}
                />
                <button type="submit" disabled={reviewSending}>
                  送信
                </button>
              </form>

              {reviewMessages.length > 0 && (
                <button
                  type="button"
                  onClick={handleGenerateRevision}
                  disabled={revising || !reviewConversationId}
                  className="ai-plan-generate"
                >
                  {revising ? '見直し案を作成中…' : 'この内容で見直しを反映する'}
                </button>
              )}
            </>
          )}

          {revision && editedRevisionGoal && (
            <div className="ai-plan-review">
              <h3>見直し案（確認・編集してから保存してください）</h3>
              <div className="ai-plan-goal-fields">
                <label>
                  タイトル
                  <input
                    value={editedRevisionGoal.title}
                    onChange={(e) => setEditedRevisionGoal({ ...editedRevisionGoal, title: e.target.value })}
                  />
                </label>
                <label>
                  詳細
                  <input
                    value={editedRevisionGoal.detail}
                    onChange={(e) => setEditedRevisionGoal({ ...editedRevisionGoal, detail: e.target.value })}
                  />
                </label>
                <label>
                  達成条件
                  <input
                    value={editedRevisionGoal.achievement_condition}
                    onChange={(e) =>
                      setEditedRevisionGoal({ ...editedRevisionGoal, achievement_condition: e.target.value })
                    }
                  />
                </label>
                <label>
                  期限
                  <input
                    type="date"
                    value={editedRevisionGoal.end_date}
                    onChange={(e) => setEditedRevisionGoal({ ...editedRevisionGoal, end_date: e.target.value })}
                  />
                </label>
                <label>
                  モード
                  <select
                    value={editedRevisionGoal.mode}
                    onChange={(e) =>
                      setEditedRevisionGoal({ ...editedRevisionGoal, mode: e.target.value as 'strict' | 'want' })
                    }
                  >
                    <option value="strict">必達</option>
                    <option value="want">努力目標</option>
                  </select>
                </label>
              </div>

              {revision.removed_tasks.length > 0 && (
                <>
                  <h3>削除するタスク</h3>
                  <ul className="ai-plan-task-list">
                    {revision.removed_tasks.map((task) => {
                      const kept = keptTaskIds.includes(task.id)
                      return (
                        <li key={task.id} className={kept ? 'removed' : ''}>
                          <div>
                            <div className="ai-plan-task-content">{task.content}</div>
                            <div className="ai-plan-task-sub">{task.date.slice(0, 10)}</div>
                          </div>
                          <button
                            type="button"
                            onClick={() =>
                              setKeptTaskIds((prev) =>
                                kept ? prev.filter((x) => x !== task.id) : [...prev, task.id],
                              )
                            }
                          >
                            {kept ? '削除に戻す' : '残す'}
                          </button>
                        </li>
                      )
                    })}
                  </ul>
                </>
              )}

              {revision.new_tasks.length > 0 && (
                <>
                  <h3>追加するタスク</h3>
                  <ul className="ai-plan-task-list">
                    {revision.new_tasks.map((task, i) => {
                      const removed = removedNewTaskIndexes.includes(i)
                      return (
                        <li key={i} className={removed ? 'removed' : ''}>
                          <div>
                            <div className="ai-plan-task-content">{task.content}</div>
                            <div className="ai-plan-task-sub">
                              {describeRecurrence(task)} ・ {task.start_date} 〜 {task.end_date}
                            </div>
                          </div>
                          <button
                            type="button"
                            onClick={() =>
                              setRemovedNewTaskIndexes((prev) =>
                                removed ? prev.filter((x) => x !== i) : [...prev, i],
                              )
                            }
                          >
                            {removed ? '戻す' : '削除'}
                          </button>
                        </li>
                      )
                    })}
                  </ul>
                </>
              )}

              <div className="ai-plan-actions">
                <button type="button" onClick={() => setRevision(null)}>
                  会話に戻る
                </button>
                <button type="button" onClick={handleConfirmRevision} disabled={confirmingRevision}>
                  {confirmingRevision ? '保存中…' : 'この内容で保存する'}
                </button>
              </div>
            </div>
          )}
        </div>
      )}

      {mode === 'view' && (
        <>
          <div className="goal-detail-header">
            <h2>{goal.title}</h2>
            <div className="goal-detail-header-actions">
              {aiEnabled === true && (
                <button type="button" onClick={handleOpenAiReview}>
                  AIと見直す
                </button>
              )}
              <button type="button" onClick={() => setMode('edit')}>
                手動で編集する
              </button>
              <button type="button" onClick={handleDelete} disabled={deleting} className="goal-detail-delete">
                {deleting ? '削除中…' : '削除'}
              </button>
            </div>
          </div>

          <p className="goal-detail-detail">{goal.detail}</p>

          <dl className="goal-detail-facts">
            <div>
              <dt>達成条件</dt>
              <dd>{goal.achievement_condition}</dd>
            </div>
            <div>
              <dt>期限</dt>
              <dd>
                {goal.end_date.slice(0, 10)}（
                {stats.days_remaining >= 0 ? `あと${stats.days_remaining}日` : '期限切れ'}）
                {stats.postpone_count > 0 && (
                  <span className="goal-detail-facts-note">先延ばし後の現在の期限です</span>
                )}
              </dd>
            </div>
            <div>
              <dt>モード</dt>
              <dd>{goal.mode === 'strict' ? '必達' : '努力目標'}</dd>
            </div>
          </dl>

          <div className="goal-detail-stats">
            <div className="stat-card">
              <span className="stat-value">{achievementPercent}%</span>
              <span className="stat-label">達成率</span>
              <span className="stat-note">予定日を過ぎたタスクのみ集計</span>
            </div>
            <div className="stat-card">
              <span className="stat-value">
                {stats.done_count} / {stats.scheduled_count}
              </span>
              <span className="stat-label">完了 / 予定</span>
            </div>
            <div className="stat-card">
              <span className="stat-value">{stats.postpone_count}</span>
              <span className="stat-label">先延ばし回数</span>
            </div>
          </div>

          {aiEnabled === true && (
            <div className="goal-detail-summary">
              <h3>AIによるサマリー</h3>
              <button type="button" onClick={handleSummarize} disabled={summaryLoading}>
                {summaryLoading ? 'AIが分析中…' : summary ? 'もう一度分析する' : 'AIに分析してもらう'}
              </button>
              {summary && <p className="goal-detail-summary-text">{summary}</p>}
            </div>
          )}
        </>
      )}
    </section>
  )
}

export default GoalDetailPage
