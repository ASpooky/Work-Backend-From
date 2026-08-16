import { useState, type SubmitEvent } from 'react'
import type { Workspace } from '../api'
import './WorkspaceSwitcher.css'

type Props = {
  workspaces: Workspace[]
  activeWorkspaceId: string | null
  onSwitch: (id: string) => void
  onCreate: (name: string) => void
  onRename: (id: string, name: string) => void
  onDelete: (id: string) => void
}

function WorkspaceSwitcher({ workspaces, activeWorkspaceId, onSwitch, onCreate, onRename, onDelete }: Props) {
  const [open, setOpen] = useState(false)
  const [renamingId, setRenamingId] = useState<string | null>(null)
  const [renameValue, setRenameValue] = useState('')
  const [creating, setCreating] = useState(false)
  const [createValue, setCreateValue] = useState('')

  const active = workspaces.find((w) => w.id === activeWorkspaceId)

  function startRename(w: Workspace) {
    setRenamingId(w.id)
    setRenameValue(w.name)
  }

  function commitRename() {
    if (renamingId && renameValue.trim()) {
      onRename(renamingId, renameValue.trim())
    }
    setRenamingId(null)
  }

  function submitRename(e: SubmitEvent<HTMLFormElement>) {
    e.preventDefault()
    commitRename()
  }

  function submitCreate(e: SubmitEvent<HTMLFormElement>) {
    e.preventDefault()
    if (createValue.trim()) {
      onCreate(createValue.trim())
    }
    setCreateValue('')
    setCreating(false)
  }

  function handleDelete(w: Workspace) {
    if (workspaces.length <= 1) return
    if (window.confirm(`「${w.name}」を削除しますか？関連する目標・タスク・AI会話もすべて削除されます。`)) {
      onDelete(w.id)
    }
  }

  return (
    <div className="workspace-switcher">
      <button type="button" className="workspace-switcher-trigger" onClick={() => setOpen((o) => !o)}>
        <span>{active?.name ?? '...'}</span>
        <span className="workspace-switcher-caret">▾</span>
      </button>

      {open && (
        <div className="workspace-switcher-menu">
          <ul>
            {workspaces.map((w) => (
              <li key={w.id}>
                {renamingId === w.id ? (
                  <form onSubmit={submitRename}>
                    <input
                      autoFocus
                      value={renameValue}
                      onChange={(e) => setRenameValue(e.target.value)}
                      onBlur={commitRename}
                    />
                  </form>
                ) : (
                  <div className="workspace-switcher-row">
                    <button
                      type="button"
                      className={`workspace-switcher-item${w.id === activeWorkspaceId ? ' active' : ''}`}
                      onClick={() => {
                        onSwitch(w.id)
                        setOpen(false)
                      }}
                    >
                      {w.name}
                    </button>
                    <button
                      type="button"
                      className="workspace-switcher-icon"
                      onClick={() => startRename(w)}
                      aria-label={`${w.name}の名前を変更`}
                    >
                      ✎
                    </button>
                    {workspaces.length > 1 && (
                      <button
                        type="button"
                        className="workspace-switcher-icon"
                        onClick={() => handleDelete(w)}
                        aria-label={`${w.name}を削除`}
                      >
                        🗑
                      </button>
                    )}
                  </div>
                )}
              </li>
            ))}
          </ul>

          {creating ? (
            <form onSubmit={submitCreate} className="workspace-switcher-create-form">
              <input
                autoFocus
                placeholder="新しいワークスペース名"
                value={createValue}
                onChange={(e) => setCreateValue(e.target.value)}
                onBlur={() => {
                  if (!createValue.trim()) setCreating(false)
                }}
              />
            </form>
          ) : (
            <button type="button" className="workspace-switcher-new" onClick={() => setCreating(true)}>
              + 新しいワークスペース
            </button>
          )}
        </div>
      )}
    </div>
  )
}

export default WorkspaceSwitcher
