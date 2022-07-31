package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Article struct {
	Id      int    `json:"id"`
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	Content string `json:"content"`
}

type Articles map[int]Article

type ArticlesHandler struct {
	sync.Mutex
	articles Articles
}

func main() {
	http.HandleFunc("/", homePage)
	articleHandler := newArticleHandler()
	http.Handle("/articles", articleHandler)
	http.Handle("/articles/", articleHandler)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "HomePage Endpoint Hit")
}

func (ah *ArticlesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		ah.getArticle(w, r)
	case "POST":
		ah.createOrUpdateArticle(w, r, len(ah.articles)+1)
	case "PUT", "PATCH":
		index, err := getIndex(r)
		if _, ok := ah.articles[index]; err != nil || ok == false {
			respondError(w, http.StatusNotFound, "Title of the article not found.")
			return
		}
		ah.createOrUpdateArticle(w, r, index)
	case "DELETE":
		ah.deleteArticle(w, r)
	default:
		respondError(w, http.StatusMethodNotAllowed, "Invalid Method.")
	}
}

func newArticleHandler() *ArticlesHandler {
	return &ArticlesHandler{
		articles: Articles{
			1: Article{Id: 1, Title: "Title1", Desc: "Desc1", Content: "Content1"},
			2: Article{Id: 2, Title: "Title2", Desc: "Desc2", Content: "Content2"},
			3: Article{Id: 3, Title: "Title3", Desc: "Desc3", Content: "Content3"},
		},
	}
}

func (ah *ArticlesHandler) getAllArticles(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "getAllArticles is called.")
	json.NewEncoder(w).Encode(ah.articles)
}

func (ah *ArticlesHandler) getArticle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "getArticle is called.")
	defer ah.Unlock()
	ah.Lock()
	index, err := getIndex(r)
	if err != nil {
		respondError(w, http.StatusNotFound, "Title of the article not found.")
		return
	}
	if index == -1 {
		ah.getAllArticles(w, r)
		return
	}
	respondJSON(w, http.StatusOK, ah.articles[index])
	fmt.Println(ah.articles[index])
}

func (ah *ArticlesHandler) createOrUpdateArticle(w http.ResponseWriter, r *http.Request, index int) {
	fmt.Fprintf(w, "createArticle is called.")
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	contentType := r.Header.Values("Content-Type")[0]
	if contentType != "application/json" {
		respondError(w, http.StatusUnsupportedMediaType, "Content is not in 'application/json' format.")
		return
	}
	var article Article
	err = json.Unmarshal(body, &article)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer ah.Unlock()
	ah.Lock()
	ah.articles[index] = article
	respondJSON(w, http.StatusCreated, ah.articles)
}

func (ah *ArticlesHandler) deleteArticle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "deleteArticle is called.")
	index, err := getIndex(r)
	defer ah.Unlock()
	ah.Lock()
	if _, ok := ah.articles[index]; err != nil || ok == false {
		respondError(w, http.StatusNotFound, "Title of the article not found.")
		return
	}
	delete(ah.articles, index)
	respondJSON(w, http.StatusCreated, ah.articles)
}

func getIndex(r *http.Request) (int, error) {
	parts := strings.Split(r.URL.String(), "/")
	fmt.Println(parts)
	if len(parts) < 2 {
		return 0, errors.New("Not found.")
	}
	index, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, errors.New("Not found.")
	}
	return index, nil
}

func respondError(w http.ResponseWriter, code int, errorMessage string) {
	respondJSON(w, code, map[string]string{"error": errorMessage})
}

func respondJSON(w http.ResponseWriter, code int, data interface{}) {
	response, _ := json.Marshal(data)
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

//Invoke-RestMethod -Method 'Post' http://localhost:8081/articles -Body (@{id=4; title="Title4"; desc="Desc4"; content="Content4"} | ConvertTo-Json) -Headers @{ "Content-Type" = "application/json"}
