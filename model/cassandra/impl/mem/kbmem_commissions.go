package cassandra_mem

import (
	"polydawn.net/repeatr/model/catalog"
	"polydawn.net/repeatr/model/formula"
)

func (kb *Base) SelectCommissionsByInputCatalog(catIDs ...catalog.ID) []*formula.Commission {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()
	markedSet := make([]*formula.Commission, 0)
	for _, plan := range kb.commissions {
		for iname, _ := range plan.Inputs { // INDEXABLE
			for _, catID := range catIDs {
				if iname == string(catID) {
					markedSet = append(markedSet, plan)
				}
			}
		}
	}
	return markedSet
}

func (kb *Base) ObserveCommissions(ch chan<- formula.CommissionID) {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()
	kb.commissionObservers = append(kb.commissionObservers, ch)
}

func (kb *Base) PublishCommission(cmsh *formula.Commission) {
	kb.mutex.Lock()
	kb.commissions[cmsh.ID] = cmsh
	var observers []chan<- formula.CommissionID
	copy(kb.commissionObservers, observers)
	kb.mutex.Unlock()
	for _, obvs := range observers {
		obvs <- cmsh.ID
	}
}
