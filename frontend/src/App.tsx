import { useEffect, useState } from 'react'
import { Routes, Route } from 'react-router-dom'
import { api, type Goal, type Workspace } from './api'
import Sidebar from './components/Sidebar'
import CalendarPage from './pages/CalendarPage'
import RegisterPage from './pages/RegisterPage'
import SettingsPage from './pages/SettingsPage'
import './App.css'

function App() {
  const [workspace, setWorkspace] = useState<Workspace | null>(null)
  const [goals, setGoals] = useState<Goal[]>([])
  const [error, setError] = useState<string | null>(null)
  const [refreshKey, setRefreshKey] = useState(0)

  useEffect(() => {
    api
      .listWorkspaces()
      .then((workspaces) => setWorkspace(workspaces[0] ?? null))
      .catch((err) => setError(String(err)))
  }, [])

  useEffect(() => {
    if (!workspace) return
    api.listGoals(workspace.id).then(setGoals).catch((err) => setError(String(err)))
  }, [workspace, refreshKey])

  function bumpRefresh() {
    setRefreshKey((k) => k + 1)
  }

  return (
    <div className="app-shell">
      {workspace && <Sidebar workspaceName={workspace.name} />}

      <main className="app">
        <h1>Goal Tracker</h1>
        {error && <p className="error">{error}</p>}
        {!workspace && <p>Loading workspace...</p>}

        {workspace && (
          <Routes>
            <Route path="/" element={<CalendarPage workspace={workspace} refreshKey={refreshKey} />} />
            <Route
              path="/register"
              element={<RegisterPage workspace={workspace} goals={goals} onCreated={bumpRefresh} />}
            />
            <Route path="/settings" element={<SettingsPage />} />
          </Routes>
        )}
      </main>
    </div>
  )
}

export default App
