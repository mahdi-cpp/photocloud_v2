package collection

//
//// FetchJSONData makes an HTTP GET request to the given URL
//// and unmarshals the JSON response body into the provided 'target' interface.
//// 'target' should be a pointer to the struct or slice you want to unmarshal into.
//func FetchJSONData(url string, target interface{}) error {
//
//	client := &http.Client{Timeout: 10 * time.Second} // Add a timeout for robustness
//
//	resp, err := client.Get(url)
//	if err != nil {
//		return fmt.Errorf("error making HTTP request to %s: %w", url, err)
//	}
//	defer resp.Body.Close() // Ensure the response body is closed
//
//	if resp.StatusCode != http.StatusOK {
//		return fmt.Errorf("received non-OK HTTP status for %s: %d %s", url, resp.StatusCode, resp.Status)
//	}
//
//	body, err := io.ReadAll(resp.Body)
//	if err != nil {
//		return fmt.Errorf("error reading response body from %s: %w", url, err)
//	}
//
//	err = json.Unmarshal(body, target)
//	if err != nil {
//		// Include the raw body in the error for debugging unmarshalling
//		return fmt.Errorf("error unmarshalling JSON from %s into target: %w\nRaw JSON: %s", url, err, string(body))
//	}
//
//	return nil
//}
//
//// FetchJSONDataMapByID makes an HTTP GET request to the given URL
//// and unmarshals the JSON response (expected to be an array) into a map.
//// The map's keys are integer IDs extracted from each item, and values are pointers to the items.
//// The generic type T must implement the Identifiable interface.
//func FetchJSONDataMapByID[T CollectionItem](url string) (map[int]*T, error) {
//	var items []T // First, unmarshal the data into a slice of type T
//
//	// Use the existing FetchJSONData function to fetch and unmarshal into the slice
//	err := FetchJSONData(url, &items)
//	if err != nil {
//		return nil, fmt.Errorf("failed to fetch data for map conversion: %w", err)
//	}
//
//	resultMap := make(map[int]*T)
//	for i := range items {
//		item := items[i]      // Get a copy of the item from the slice
//		id := item.GetID()    // Use the GetID method from the Identifiable interface
//		resultMap[id] = &item // Store a pointer to the item in the map using its ID as the key
//	}
//
//	return resultMap, nil
//}
//
//func Read[T CollectionItem]() (*T, error) {
//
//	data := new(T)
//	file, err := os.ReadFile(control.filePath)
//	if err != nil {
//		if os.IsNotExist(err) {
//			return data, nil
//		}
//		return nil, err
//	}
//
//	if len(file) == 0 {
//		return data, nil
//	}
//
//	if err := json.Unmarshal(file, data); err != nil {
//		return nil, err
//	}
//
//	return data, nil
//}
