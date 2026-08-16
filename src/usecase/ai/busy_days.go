package ai

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

const busyDaysLookaheadDays = 28
const busyDaysMinTaskCount = 2

// WorkspaceTaskRangeReader is satisfied by DailyTaskRepository, used to see
// what's already scheduled across a workspace's other goals — separate from
// TaskRangeReader (summarize_goal.go), which is scoped to one goal's own
// tasks for its own stats.
type WorkspaceTaskRangeReader interface {
	FindByWorkspaceIDAndDateRange(workspaceID string, from, to time.Time) ([]*entity.DailyTask, error)
}

// summarizeBusyDays turns a workspace's upcoming tasks into a short block
// for an AI prompt, so it can spread new/rescheduled tasks onto quieter days
// instead of piling everything onto ones that already have several from
// other goals. excludeGoalID lets a goal's own tasks not count as
// "competition" against itself when revising that same goal. Returns "" if
// there's nothing worth mentioning (single-goal workspaces, or workspaces
// with light scheduling), so the prompt doesn't grow a section for no
// reason.
func summarizeBusyDays(tasks []*entity.DailyTask, excludeGoalID string) string {
	counts := make(map[string]int)
	for _, t := range tasks {
		if t.GoalID == excludeGoalID {
			continue
		}
		counts[t.Date.Format(planDateLayout)]++
	}

	dates := make([]string, 0, len(counts))
	for d, n := range counts {
		if n >= busyDaysMinTaskCount {
			dates = append(dates, d)
		}
	}
	if len(dates) == 0 {
		return ""
	}
	sort.Strings(dates)

	var b strings.Builder
	b.WriteString("直近4週間で、他の目標のタスクが既に複数入っている日(参考情報。無理にすべて避ける必要はない):\n")
	for _, d := range dates {
		fmt.Fprintf(&b, "- %s: %d件\n", d, counts[d])
	}
	return b.String()
}
