import CalendarView from '../components/CalendarView'
import type { Workspace } from '../api'

type Props = {
  workspace: Workspace
  refreshKey: number
}

function CalendarPage({ workspace, refreshKey }: Props) {
  return (
    <section>
      <h2>Calendar</h2>
      <CalendarView workspaceId={workspace.id} refreshKey={refreshKey} />
    </section>
  )
}

export default CalendarPage
