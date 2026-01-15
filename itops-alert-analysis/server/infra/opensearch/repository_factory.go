package opensearch

import (
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-analysis/core"
	"github.com/opensearch-project/opensearch-go/v2"
)

type RepositoryFactory struct {
	client *opensearch.Client

	rawEventStore            core.RawEventRepository
	faultPointStore          core.FaultPointRepository
	faultPointRelationStore  core.FaultPointRelationRepository
	problemStore             core.ProblemRepository
	faultCausalStore         core.FaultCausalRepository
	faultCausalRelationStore core.FaultCausalRelationRepository
}

func NewRepositoryFactory(client *opensearch.Client) *RepositoryFactory {
	return &RepositoryFactory{client: client}
}

func (r *RepositoryFactory) RawEvents() core.RawEventRepository {
	if r.rawEventStore == nil {
		r.rawEventStore = NewRawEventStore(r.client)
	}
	return r.rawEventStore
}
func (r *RepositoryFactory) FaultPoints() core.FaultPointRepository {
	if r.faultPointStore == nil {
		r.faultPointStore = NewFaultPointStore(r.client)
	}
	return r.faultPointStore
}

func (r *RepositoryFactory) FaultPointRelations() core.FaultPointRelationRepository {
	if r.faultPointRelationStore == nil {
		r.faultPointRelationStore = NewFaultPointRelationStore(r.client)
	}
	return r.faultPointRelationStore
}

func (r *RepositoryFactory) Problems() core.ProblemRepository {
	if r.problemStore == nil {
		r.problemStore = NewProblemStore(r.client)
	}
	return r.problemStore
}

func (r *RepositoryFactory) FaultCausals() core.FaultCausalRepository {
	if r.faultCausalStore == nil {
		r.faultCausalStore = NewFaultCausalStore(r.client)
	}
	return r.faultCausalStore
}

func (r *RepositoryFactory) FaultCausalRelations() core.FaultCausalRelationRepository {
	if r.faultCausalRelationStore == nil {
		r.faultCausalRelationStore = NewFaultCausalRelationStore(r.client)
	}
	return r.faultCausalRelationStore
}
