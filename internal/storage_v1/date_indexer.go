package storage_v1

//
//type DateIndexer struct {
//	mu    sync.RWMutex
//	index map[string][]int // YYYY-MM-DD → assetIDs
//}
//
//func NewDateIndexer() *DateIndexer {
//	return &DateIndexer{
//		index: make(map[string][]int),
//	}
//}
//
//func (idx *DateIndexer) AddAsset(assetID int, date time.Time) {
//	dateKey := date.Format("2006-01-02")
//
//	idx.mu.Lock()
//	defer idx.mu.Unlock()
//	idx.index[dateKey] = append(idx.index[dateKey], assetID)
//}
//
//func (idx *DateIndexer) Filter(assetIDs []int, start, end *time.Time) []int {
//	if start == nil && end == nil {
//		return assetIDs
//	}
//
//	idx.mu.RLock()
//	defer idx.mu.RUnlock()
//
//	filtered := make([]int, 0, len(assetIDs))
//
//	for _, id := range assetIDs {
//		// Find date for asset
//		var found bool
//		for dateKey, ids := range idx.index {
//			for _, assetID := range ids {
//				if assetID == id {
//					assetDate, _ := time.Parse("2006-01-02", dateKey)
//
//					if start != nil && assetDate.Before(*start) {
//						continue
//					}
//					if end != nil && assetDate.After(*end) {
//						continue
//					}
//
//					filtered = append(filtered, id)
//					found = true
//					break
//				}
//			}
//			if found {
//				break
//			}
//		}
//	}
//
//	return filtered
//}
