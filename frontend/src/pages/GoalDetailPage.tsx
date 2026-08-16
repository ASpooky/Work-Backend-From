import { useEffect, useState, type SubmitEvent } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { api, type GoalStats } from '../api'
import { toUserMessage } from '../errors'
import './GoalDetailPage.css'

type Props = {
  onUpdated: () => void
}

function GoalDetailPage({ onUpdated }: Props) {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [stats, setStats] = useState<GoalStats | null>(null)
  const [notFound, setNotFound] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [editing, setEditing] = useState(false)
  const [saving, setSaving] = useState(false)

  const [title, setTitle] = useState('')
  const [detail, setDetail] = useState('')
  const [condition, setCondition] = useState('')
  const [endDate, setEndDate] = useState('')
  const [mode, setMode] = useState<'strict' | 'want'>('strict')

  const [aiEnabled, setAiEnabled] = useState<boolean | null>(null)
  const [summary, setSummary] = useState<string | null>(null)
  const [summaryLoading, setSummaryLoading] = useState(false)

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
        setMode(s.goal.mode)
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
        mode,
      })
      setEditing(false)
      load()
      onUpdated()
    } catch (err) {
      setError(toUserMessage(err))
    } finally {
      setSaving(false)
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

      {editing ? (
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
            <select value={mode} onChange={(e) => setMode(e.target.value as 'strict' | 'want')}>
              <option value="strict">必達</option>
              <option value="want">努力目標</option>
            </select>
          </label>
          <div className="goal-detail-edit-actions">
            <button type="submit" disabled={saving}>
              {saving ? '保存中…' : '保存'}
            </button>
            <button type="button" onClick={() => setEditing(false)} disabled={saving}>
              キャンセル
            </button>
          </div>
        </form>
      ) : (
        <>
          <div className="goal-detail-header">
            <h2>{goal.title}</h2>
            <button type="button" onClick={() => setEditing(true)}>
              目標を見直す
            </button>
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
