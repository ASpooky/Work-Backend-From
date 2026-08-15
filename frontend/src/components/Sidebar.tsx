import { NavLink } from 'react-router-dom'
import './Sidebar.css'

type Props = {
  workspaceName: string
}

function Sidebar({ workspaceName }: Props) {
  return (
    <aside className="sidebar">
      <div className="sidebar-brand">Goal Tracker</div>
      <div className="sidebar-workspace">{workspaceName}</div>

      <nav className="sidebar-nav">
        <NavLink to="/" end className={({ isActive }) => `sidebar-nav-item${isActive ? ' active' : ''}`}>
          カレンダー
        </NavLink>
        <NavLink to="/register" className={({ isActive }) => `sidebar-nav-item${isActive ? ' active' : ''}`}>
          登録
        </NavLink>
        <NavLink to="/settings" className={({ isActive }) => `sidebar-nav-item${isActive ? ' active' : ''}`}>
          設定
        </NavLink>
      </nav>
    </aside>
  )
}

export default Sidebar
