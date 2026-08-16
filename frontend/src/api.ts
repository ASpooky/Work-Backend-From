const API_BASE = 'http://localhost:8080'

export type Workspace = {
  id: string
  user_id: string
  name: string
  created_at: string
}

export type Goal = {
  id: string
  workspace_id: string
  title: string
  detail: string
  achievement_condition: string
  end_date: string
  mode: 'strict' | 'want'
  status: 'active' | 'achieved' | 'abandoned'
  created_at: string
}

export type DailyTask = {
  id: string
  goal_id: string
  date: string
  content: string
  done: boolean
  created_at: string
}

export type DayStatus = 'no_task' | 'not_done' | 'done'

export type DayEntry = {
  status: DayStatus
  content: string
}

export type GoalCalendar = {
  goal: Goal
  days: Record<string, DayEntry>
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!res.ok) {
    throw new Error(`${init?.method ?? 'GET'} ${path} failed: ${res.status} ${await res.text()}`)
  }
  return res.json() as Promise<T>
}

export const api = {
  listWorkspaces: () => request<Workspace[]>('/workspaces'),
  createWorkspace: (body: { user_id: string; name: string }) =>
    request<Workspace>('/workspaces', { method: 'POST', body: JSON.stringify(body) }),

  listGoals: (workspaceId: string) =>
    request<Goal[]>(`/goals?workspace_id=${encodeURIComponent(workspaceId)}`),
  createGoal: (body: {
    workspace_id: string
    title: string
    detail: string
    achievement_condition: string
    end_date: string
    mode: 'strict' | 'want'
  }) => request<Goal>('/goals', { method: 'POST', body: JSON.stringify(body) }),

  listDailyTasks: (date: string) =>
    request<DailyTask[]>(`/daily-tasks?date=${encodeURIComponent(date)}`),
  createDailyTask: (body: { goal_id: string; date: string; content: string }) =>
    request<DailyTask>('/daily-tasks', { method: 'POST', body: JSON.stringify(body) }),
  createRecurringDailyTasks: (body: {
    goal_id: string
    content: string
    start_date: string
    end_date: string
    rule: { type: 'interval'; interval_days: number } | { type: 'weekly'; weekdays: number[] }
  }) => request<DailyTask[]>('/daily-tasks/recurring', { method: 'POST', body: JSON.stringify(body) }),
  updateDailyTaskDone: (id: string, done: boolean) =>
    request<{ id: string; done: boolean }>(`/daily-tasks/${encodeURIComponent(id)}`, {
      method: 'PATCH',
      body: JSON.stringify({ done }),
    }),

  getCalendar: (workspaceId: string, from: string, to: string) =>
    request<GoalCalendar[]>(
      `/calendar?workspace_id=${encodeURIComponent(workspaceId)}&from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,
    ),
}
