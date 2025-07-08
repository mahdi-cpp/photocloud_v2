package indexer

//
//type TextIndexer struct {
//	mu    sync.RWMutex
//	index map[string]map[int]bool // word → assetID → exists
//}
//
//func NewTextIndexer() *TextIndexer {
//	return &TextIndexer{
//		index: make(map[string]map[int]bool),
//	}
//}
//
//func (idx *TextIndexer) AddAsset(assetID int, filename string) {
//	words := tokenize(filename)
//
//	idx.mu.Lock()
//	defer idx.mu.Unlock()
//
//	for _, word := range words {
//		if idx.index[word] == nil {
//			idx.index[word] = make(map[int]bool)
//		}
//		idx.index[word][assetID] = true
//	}
//}
//
//func (idx *TextIndexer) Search(query string) []int {
//	words := tokenize(query)
//	results := make(map[int]int) // assetID → match count
//
//	idx.mu.RLock()
//	defer idx.mu.RUnlock()
//
//	for _, word := range words {
//		if assets, exists := idx.index[word]; exists {
//			for assetID := range assets {
//				results[assetID]++
//			}
//		}
//	}
//
//	// Order by relevance
//	var sortedIDs []int
//	for assetID, count := range results {
//		if count == len(words) { // All words matched
//			sortedIDs = append([]int{assetID}, sortedIDs...)
//		} else {
//			sortedIDs = append(sortedIDs, assetID)
//		}
//	}
//
//	return sortedIDs
//}
//
//func tokenize(text string) []string {
//	words := strings.Fields(strings.ToLower(text))
//	var tokens []string
//	for _, word := range words {
//		if len(word) > 2 { // Skip short words
//			tokens = append(tokens, word)
//		}
//	}
//	return tokens
//}
