import { NavLink } from 'react-router-dom'
import type { Workspace } from '../api'
import WorkspaceSwitcher from './WorkspaceSwitcher'
import './Sidebar.css'

type Props = {
  workspaces: Workspace[]
  activeWorkspaceId: string | null
  onSwitchWorkspace: (id: string) => void
  onCreateWorkspace: (name: string) => void
  onRenameWorkspace: (id: string, name: string) => void
  onDeleteWorkspace: (id: string) => void
}

function Sidebar({
  workspaces,
  activeWorkspaceId,
  onSwitchWorkspace,
  onCreateWorkspace,
  onRenameWorkspace,
  onDeleteWorkspace,
}: Props) {
  return (
    <aside className="sidebar">
      <div className="sidebar-brand">目標トラッカー</div>

      <div className="sidebar-workspace">
        <WorkspaceSwitcher
          workspaces={workspaces}
          activeWorkspaceId={activeWorkspaceId}
          onSwitch={onSwitchWorkspace}
          onCreate={onCreateWorkspace}
          onRename={onRenameWorkspace}
          onDelete={onDeleteWorkspace}
        />
      </div>

      <nav className="sidebar-nav">
        <NavLink to="/" end className={({ isActive }) => `sidebar-nav-item${isActive ? ' active' : ''}`}>
          カレンダー
        </NavLink>
        <NavLink to="/goals" className={({ isActive }) => `sidebar-nav-item${isActive ? ' active' : ''}`}>
          目標
        </NavLink>
        <NavLink to="/register" className={({ isActive }) => `sidebar-nav-item${isActive ? ' active' : ''}`}>
          登録
        </NavLink>
        <NavLink to="/ai-plan" className={({ isActive }) => `sidebar-nav-item${isActive ? ' active' : ''}`}>
          AIと計画
        </NavLink>
        <NavLink to="/settings" className={({ isActive }) => `sidebar-nav-item${isActive ? ' active' : ''}`}>
          設定
        </NavLink>
      </nav>
    </aside>
  )
}

export default Sidebar
