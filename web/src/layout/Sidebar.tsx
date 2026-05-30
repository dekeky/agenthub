import { NavLink } from 'react-router-dom'
import { Grid3X3, UploadCloud, Boxes } from 'lucide-react'
import { ThemeToggle } from '../theme/ThemeToggle'
import { useConnection } from '../hooks/useAgents'

export function Sidebar() {
  const { data: isConnected } = useConnection()

  return (
    <aside className="sidebar">
      <div className="sidebar-header">
        <div className="logo">
          <div className="logo-icon">
            <Boxes size={18} />
          </div>
          <span className="logo-text">AgentHub</span>
        </div>
        <ThemeToggle />
      </div>

      <nav className="sidebar-nav">
        <NavLink
          to="/"
          end
          className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
        >
          <Grid3X3 size={18} />
          <span>浏览仓库</span>
        </NavLink>
        <NavLink
          to="/upload"
          className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
        >
          <UploadCloud size={18} />
          <span>上传 Agent</span>
        </NavLink>
      </nav>

      <div className="sidebar-footer">
        <div className={`connection-indicator ${isConnected ? 'online' : 'error'}`}>
          <span className="indicator-dot" />
          <span className="indicator-text">{isConnected ? '已连接' : '未连接'}</span>
        </div>
      </div>
    </aside>
  )
}
