package main

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	portKey = "PORT"

	templatesDir = "templates/"
	staticDir    = "static/"
	jsDir        = "js/"

	inputDataFileName = "./products.json"
)

type Item struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Keywords []string `json:"-"`
}

var (
	inputData    []*Item
	keywordsMap  map[string][]int
	keywordsSet  []string
	productsByID map[int]*Item
)

type ItemsResponse struct {
	Count int     `json:"count"`
	Items []*Item `json:"items"`
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()["query"][0]
	log.Printf("query: %s\n", query)

	startTime := time.Now()
	defer func() {
		endTime := time.Since(startTime)
		log.Printf("elapsed time: %d ms\n", endTime.Milliseconds())
	}()

	matchingItems := queryAll(query)
	response := &ItemsResponse{
		Count: len(matchingItems),
		Items: matchingItems,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR]: failed to serialize response %v\n", err.Error())

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loadData() {
	inputFile, err := os.Open(inputDataFileName)
	if err != nil {
		log.Panic("Failed to open data file")
	}

	defer inputFile.Close()

	rawData, err := io.ReadAll(inputFile)
	if err != nil {
		log.Panic("Failed to read data")
	}

	var data []*Item
	json.Unmarshal(rawData, &data)

	inputData = data
	productsByID = make(map[int]*Item, len(inputData))

	log.Printf("Lenght of read data: %d\n", len(inputData))

	keywordsMap = make(map[string][]int)
	for i := range inputData {
		inputData[i].Keywords = strings.Split(inputData[i].Name, " ")
		for j, keyword := range inputData[i].Keywords {
			keyword = strings.ToLower(keyword)
			inputData[i].Keywords[j] = keyword

			if _, ok := keywordsMap[keyword]; !ok {
				keywordsMap[keyword] = make([]int, 0)
			}
			keywordsMap[keyword] = append(keywordsMap[keyword], i)
		}
		productsByID[inputData[i].ID] = inputData[i]
	}

	keywordsSet = make([]string, 0, len(keywordsMap))
	for key := range keywordsMap {
		keywordsSet = append(keywordsSet, key)
	}

	sort.Strings(keywordsSet)

	log.Printf("item[0] = %v\n", inputData[0])
	log.Printf("Lenght of keywordsSet: %d\n", len(keywordsSet))
}

// queryProduct: receives a query string. Then will query from inputData to find matching elements and return them.
func queryProduct(query string) []*Item {
	matchingItems := make([]*Item, 0)

	// use a lower-cased query for simplicity
	lowercaseQuery := strings.ToLower(query)

	index, _ := sort.Find(len(keywordsSet), func(i int) int {
		return strings.Compare(lowercaseQuery, keywordsSet[i])
	})

	if index == len(keywordsSet) {
		return matchingItems
	}

	for i := index; i < len(keywordsSet); i++ {
		keyword := keywordsSet[i]
		if !strings.HasPrefix(keyword, lowercaseQuery) {
			break
		}

		itemsList := keywordsMap[keyword]
		for _, v := range itemsList {
			matchingItems = append(matchingItems, inputData[v])
		}
	}

	return matchingItems
}

func queryAny(query string) []*Item {
	matchingSet := make(map[int]*Item)

	queries := strings.Split(query, " ")
	for _, q := range queries {
		products := queryProduct(q)
		for _, product := range products {
			matchingSet[product.ID] = product
		}
	}

	matching := make([]*Item, len(matchingSet))
	index := 0
	for _, product := range matchingSet {
		matching[index] = product
		index += 1
	}

	return matching
}

func queryAll(query string) []*Item {
	matchingSet := make(map[int]struct{})

	queries := strings.Split(query, " ")
	firsIteration := true
	for _, q := range queries {
		products := queryProduct(q)
		currentMatchingSet := make(map[int]struct{})
		for _, product := range products {
			currentMatchingSet[product.ID] = struct{}{}
		}
		matchingSet = intersectMatchingSets(matchingSet, currentMatchingSet, firsIteration)
		firsIteration = false
	}

	log.Printf("len matchingSet: %d\n", len(matchingSet))

	matching := make([]*Item, len(matchingSet))
	index := 0
	for productID := range matchingSet {
		matching[index] = productsByID[productID]
		index += 1
	}

	return matching
}

func intersectMatchingSets(matchingSet map[int]struct{}, currentMatchingSet map[int]struct{}, firstIteration bool) map[int]struct{} {
	finalMatchingSet := make(map[int]struct{})
	for productID := range currentMatchingSet {
		if _, ok := matchingSet[productID]; ok || firstIteration {
			finalMatchingSet[productID] = struct{}{}
		}
	}
	return finalMatchingSet
}

// find a matching keyword given a query
func match(item Item, query string) bool {
	for _, word := range item.Keywords {
		if strings.HasPrefix(word, query) {
			return true
		}
	}
	return false
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(templatesDir + "index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	t.Execute(w, nil)
}

func setupAndRunServer(port string) {
	// serve static content
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(jsDir))))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("GET /data", getHandler)

	log.Fatal(http.ListenAndServe(port, nil))
}

func main() {
	loadData()

	port := os.Getenv(portKey)
	setupAndRunServer(port)
}
