import CalendarView from '../components/CalendarView'
import type { Workspace } from '../api'

type Props = {
  workspace: Workspace
  workspaces: Workspace[]
  refreshKey: number
}

function CalendarPage({ workspace, workspaces, refreshKey }: Props) {
  return (
    <section>
      <h2>カレンダー</h2>
      <CalendarView workspaceId={workspace.id} workspaces={workspaces} refreshKey={refreshKey} />
    </section>
  )
}

export default CalendarPage
