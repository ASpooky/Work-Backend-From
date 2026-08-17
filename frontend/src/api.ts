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
  priority: number
}

export type DailyTask = {
  id: string
  goal_id: string
  date: string
  content: string
  done: boolean
  created_at: string
  memo?: string
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

export type GoalStats = {
  goal: Goal
  scheduled_count: number
  done_count: number
  achievement_rate: number
  postpone_count: number
  days_remaining: number
}

export type Conversation = {
  id: string
  workspace_id: string
  title: string
  created_at: string
  updated_at: string
}

export type ConversationMessage = {
  id: string
  conversation_id: string
  role: 'user' | 'model'
  content: string
  created_at: string
}

export type PlannedGoal = {
  title: string
  detail: string
  achievement_condition: string
  end_date: string
  mode: 'strict' | 'want'
}

export type PlannedTask = {
  content: string
  start_date: string
  end_date: string
  rule_type: 'interval' | 'weekly'
  interval_days?: number
  weekdays?: number[]
}

export type PlannedGoalPlan = {
  goal: PlannedGoal
  tasks: PlannedTask[]
}

export type Plan = {
  goals: PlannedGoalPlan[]
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!res.ok) {
    throw new Error(`${init?.method ?? 'GET'} ${path} failed: ${res.status} ${await res.text()}`)
  }
  if (res.status === 204) {
    return undefined as T
  }
  return res.json() as Promise<T>
}

export const api = {
  listWorkspaces: () => request<Workspace[]>('/workspaces'),
  createWorkspace: (body: { user_id: string; name: string }) =>
    request<Workspace>('/workspaces', { method: 'POST', body: JSON.stringify(body) }),
  renameWorkspace: (id: string, name: string) =>
    request<{ id: string; name: string }>(`/workspaces/${encodeURIComponent(id)}`, {
      method: 'PATCH',
      body: JSON.stringify({ name }),
    }),
  deleteWorkspace: (id: string) =>
    request<void>(`/workspaces/${encodeURIComponent(id)}`, { method: 'DELETE' }),

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
  getGoal: (id: string) => request<GoalStats>(`/goals/${encodeURIComponent(id)}`),
  updateGoal: (
    id: string,
    body: {
      title: string
      detail: string
      achievement_condition: string
      end_date: string
      mode: 'strict' | 'want'
    },
  ) => request<{ id: string }>(`/goals/${encodeURIComponent(id)}`, { method: 'PATCH', body: JSON.stringify(body) }),
  deleteGoal: (id: string) => request<void>(`/goals/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  reorderGoal: (id: string, priority: number) =>
    request<{ id: string; priority: number }>(`/goals/${encodeURIComponent(id)}/priority`, {
      method: 'PATCH',
      body: JSON.stringify({ priority }),
    }),

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
  updateTaskMemo: (id: string, memo: string | null) =>
    request<{ id: string; memo: string | null }>(`/daily-tasks/${encodeURIComponent(id)}/memo`, {
      method: 'PATCH',
      body: JSON.stringify({ memo }),
    }),

  getCalendar: (workspaceId: string, from: string, to: string) =>
    request<GoalCalendar[]>(
      `/calendar?workspace_id=${encodeURIComponent(workspaceId)}&from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,
    ),

  aiStatus: () => request<{ enabled: boolean }>('/ai/status'),
  listConversations: (workspaceId: string) =>
    request<Conversation[]>(`/ai/conversations?workspace_id=${encodeURIComponent(workspaceId)}`),
  getConversation: (id: string) =>
    request<ConversationMessage[]>(`/ai/conversations/${encodeURIComponent(id)}/messages`),
  sendChatMessage: (workspaceId: string, conversationId: string | null, content: string) =>
    request<{ conversation_id: string; reply: string }>('/ai/chat', {
      method: 'POST',
      body: JSON.stringify({ workspace_id: workspaceId, conversation_id: conversationId ?? '', content }),
    }),
  aiPlan: (workspaceId: string, conversationId: string) =>
    request<Plan>('/ai/plan', {
      method: 'POST',
      body: JSON.stringify({ workspace_id: workspaceId, conversation_id: conversationId }),
    }),
  summarizeGoal: (id: string) =>
    request<{ summary: string }>(`/ai/goals/${encodeURIComponent(id)}/summary`, { method: 'POST' }),

  listGoalConversations: (goalId: string) =>
    request<Conversation[]>(`/ai/goals/${encodeURIComponent(goalId)}/conversations`),
  sendGoalReviewMessage: (goalId: string, workspaceId: string, conversationId: string | null, content: string) =>
    request<{ conversation_id: string; reply: string }>(`/ai/goals/${encodeURIComponent(goalId)}/chat`, {
      method: 'POST',
      body: JSON.stringify({ workspace_id: workspaceId, conversation_id: conversationId ?? '', content }),
    }),
  reviseGoal: (goalId: string, conversationId: string) =>
    request<PlannedGoal>(`/ai/goals/${encodeURIComponent(goalId)}/revise`, {
      method: 'POST',
      body: JSON.stringify({ conversation_id: conversationId }),
    }),
}
