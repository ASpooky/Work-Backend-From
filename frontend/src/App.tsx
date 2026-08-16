import { useEffect, useState } from 'react'
import { Routes, Route, useLocation } from 'react-router-dom'
import { api, type Goal, type Workspace } from './api'
import { toUserMessage } from './errors'
import Sidebar from './components/Sidebar'
import AIPlanPage from './pages/AIPlanPage'
import CalendarPage from './pages/CalendarPage'
import RegisterPage from './pages/RegisterPage'
import SettingsPage from './pages/SettingsPage'
import GoalDetailPage from './pages/GoalDetailPage'
import GoalsListPage from './pages/GoalsListPage'
import './App.css'

const ACTIVE_WORKSPACE_KEY = 'goal-tracker:active-workspace'

// A synthetic workspace representing "every workspace at once". Its id is the
// empty string, which is exactly the sentinel the backend already treats as
// "all workspaces" for /goals, /calendar and /daily-tasks.
const ALL_WORKSPACES: Workspace = { id: '', user_id: '', name: 'すべて', created_at: '' }

function App() {
  const location = useLocation()
  const [workspaces, setWorkspaces] = useState<Workspace[]>([])
  const [activeWorkspaceId, setActiveWorkspaceId] = useState<string | null>(() =>
    localStorage.getItem(ACTIVE_WORKSPACE_KEY),
  )
  const [goals, setGoals] = useState<Goal[]>([])
  const [error, setError] = useState<string | null>(null)
  const [refreshKey, setRefreshKey] = useState(0)

  const workspace =
    activeWorkspaceId === '' ? ALL_WORKSPACES : workspaces.find((w) => w.id === activeWorkspaceId) ?? null
  const isAllWorkspaces = workspace?.id === ''

  function reloadWorkspaces() {
    api
      .listWorkspaces()
      .then((list) => {
        setWorkspaces(list)
        setActiveWorkspaceId((prev) => {
          if (prev === '') return ''
          if (prev && list.some((w) => w.id === prev)) return prev
          const next = list[0]?.id ?? null
          if (next) localStorage.setItem(ACTIVE_WORKSPACE_KEY, next)
          return next
        })
      })
      .catch((err) => setError(toUserMessage(err)))
  }

  useEffect(() => {
    reloadWorkspaces()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (!workspace) return
    api.listGoals(workspace.id).then(setGoals).catch((err) => setError(toUserMessage(err)))
  }, [workspace, refreshKey])

  function bumpRefresh() {
    setRefreshKey((k) => k + 1)
  }

  function handleSwitchWorkspace(id: string) {
    setActiveWorkspaceId(id)
    localStorage.setItem(ACTIVE_WORKSPACE_KEY, id)
  }

  async function handleCreateWorkspace(name: string) {
    const userId = workspaces[0]?.user_id
    if (!userId) return
    try {
      const created = await api.createWorkspace({ user_id: userId, name })
      reloadWorkspaces()
      handleSwitchWorkspace(created.id)
    } catch (err) {
      setError(toUserMessage(err))
    }
  }

  async function handleRenameWorkspace(id: string, name: string) {
    try {
      await api.renameWorkspace(id, name)
      reloadWorkspaces()
    } catch (err) {
      setError(toUserMessage(err))
    }
  }

  async function handleDeleteWorkspace(id: string) {
    try {
      await api.deleteWorkspace(id)
      reloadWorkspaces()
    } catch (err) {
      setError(toUserMessage(err))
    }
  }

  return (
    <div className="app-shell">
      {workspace && (
        <Sidebar
          workspaces={workspaces}
          activeWorkspaceId={activeWorkspaceId}
          onSwitchWorkspace={handleSwitchWorkspace}
          onCreateWorkspace={handleCreateWorkspace}
          onRenameWorkspace={handleRenameWorkspace}
          onDeleteWorkspace={handleDeleteWorkspace}
        />
      )}

      <main className={`app${location.pathname === '/ai-plan' ? ' app-wide' : ''}`}>
        <h1>Goal Tracker</h1>
        {error && <p className="error">{error}</p>}
        {!workspace && <p>Loading workspace...</p>}

        {workspace && (
          <Routes>
            <Route
              path="/"
              element={
                <CalendarPage
                  workspace={workspace}
                  workspaces={workspaces}
                  refreshKey={refreshKey}
                />
              }
            />
            <Route
              path="/goals"
              element={
                <GoalsListPage
                  workspace={workspace}
                  workspaces={workspaces}
                  refreshKey={refreshKey}
                  onDeleted={bumpRefresh}
                />
              }
            />
            <Route
              path="/register"
              element={
                <RegisterPage
                  workspace={workspace}
                  goals={goals}
                  isAllWorkspaces={isAllWorkspaces}
                  onCreated={bumpRefresh}
                />
              }
            />
            <Route
              path="/ai-plan"
              element={<AIPlanPage workspace={workspace} isAllWorkspaces={isAllWorkspaces} onCreated={bumpRefresh} />}
            />
            <Route path="/settings" element={<SettingsPage />} />
            <Route path="/goals/:id" element={<GoalDetailPage onUpdated={bumpRefresh} />} />
          </Routes>
        )}
      </main>
    </div>
  )
}

export default App
