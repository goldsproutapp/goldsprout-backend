package processing

import (
	"time"

	"github.com/goldsproutapp/goldsprout-backend/models"
	"github.com/goldsproutapp/goldsprout-backend/util"
)

func CreateMergedSnapshotMap(snapshots []models.StockSnapshot) (map[time.Time][]models.StockSnapshot, map[string][]models.StockSnapshot) {

	yearStartMap := map[string][]models.StockSnapshot{}
	latestMap := map[string]models.StockSnapshot{}
	snapshotMap := map[uint]map[time.Time][]models.StockSnapshot{}
	now := time.Now()
	// TODO: this should be configurable for l10n
	yearStart := time.Date(now.Year(), 4, 6, 0, 0, 0, 0, time.UTC)
	if now.Month() < 4 || now.Month() == 4 && now.Day() < 6 {
		yearStart = yearStart.AddDate(-1, 0, 0)
	}
	for _, snapshot := range snapshots {
		key := snapshot.Key()
		if !util.ContainsKey(latestMap, key) ||
			latestMap[key].Date.Compare(snapshot.Date) == -1 {
			latestMap[key] = snapshot
		}
		if snapshot.Date.After(yearStart) {
			if util.ContainsKey(yearStartMap, key) {
				yearStartMap[key] = append(yearStartMap[key], snapshot)
			} else {
				yearStartMap[key] = []models.StockSnapshot{snapshot}
			}
		}
		if !util.ContainsKey(snapshotMap, snapshot.AccountID) {
			snapshotMap[snapshot.AccountID] = map[time.Time][]models.StockSnapshot{}
		}
		if !util.ContainsKey(snapshotMap[snapshot.AccountID], snapshot.Date) {
			snapshotMap[snapshot.AccountID][snapshot.Date] = []models.StockSnapshot{snapshot}
		} else {
			snapshotMap[snapshot.AccountID][snapshot.Date] = append(snapshotMap[snapshot.AccountID][snapshot.Date], snapshot)
		}
	}
	snapshotMapMerged := map[time.Time][]models.StockSnapshot{}
	for accountID, m := range snapshotMap {
	dateLoop:
		for date, snapshots := range m {
			if util.ContainsKey(snapshotMapMerged, date) {
				continue dateLoop
			}
			allSnapshots := snapshots
			for otherAccount, otherMap := range snapshotMap {
				if accountID == otherAccount {
					continue
				}
				closest := time.Unix(0, 0)
				earliest := time.Now()
				var closestSnapshots []models.StockSnapshot
				for t, otherSnapshots := range otherMap {
					if t.Before(earliest) {
						earliest = t
					}
					if date.Sub(t).Abs().Seconds() < date.Sub(closest).Abs().Seconds() {
						closest = t
						closestSnapshots = otherSnapshots
					}
				}
				if !earliest.After(date) {
					allSnapshots = append(allSnapshots, closestSnapshots...)
				}
			}
			snapshotMapMerged[date] = allSnapshots
		}
	}
	return snapshotMapMerged, yearStartMap
}

