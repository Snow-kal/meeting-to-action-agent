package agents

import "github.com/Snow-kal/meeting-to-action-agent/internal/domain"

type OwnerAgent struct{}

func NewOwnerAgent() *OwnerAgent {
	return &OwnerAgent{}
}

func (a *OwnerAgent) Resolve(tasks []domain.Task, decisions []domain.Decision) []domain.Task {
	decisionOwners := make(map[string]string, len(decisions))
	for _, decision := range decisions {
		if decision.OwnerHint != "" {
			decisionOwners[decision.ID] = decision.OwnerHint
		}
	}

	resolved := make([]domain.Task, 0, len(tasks))
	for _, task := range tasks {
		current := task
		if current.Owner == "" {
			if owner := extractOwner(current.SourceText); owner != "" {
				current.Owner = owner
			} else if owner := extractOwner(current.Description); owner != "" {
				current.Owner = owner
			} else if current.SourceDecisionID != "" {
				current.Owner = decisionOwners[current.SourceDecisionID]
			}
		}
		resolved = append(resolved, current)
	}

	return resolved
}
