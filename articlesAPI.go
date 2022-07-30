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
	Id      int
	Title   string `json:"Title"`
	Desc    string `json:"Desc"`
	Content string `json:"Content"`
}

type Articles []Article

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
		ah.createArticle(w, r)
	case "PUT", "PATCH":
		ah.modifyArticle(w, r)
	case "DELETE":
		ah.deleteArticle(w, r)
	default:
		respondError(w, http.StatusMethodNotAllowed, "Invalid Method.")
	}
}

func newArticleHandler() *ArticlesHandler {
	return &ArticlesHandler{
		articles: Articles{
			Article{Id: 1, Title: "Title1", Desc: "Desc1", Content: "Content1"},
			Article{Id: 2, Title: "Title2", Desc: "Desc2", Content: "Content2"},
			Article{Id: 3, Title: "Title3", Desc: "Desc3", Content: "Content3"},
		},
	}
}

func (ah *ArticlesHandler) getAllArticles(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "getAllArticles is called.")
}

func (ah *ArticlesHandler) getArticle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "getArticle is called.")
	defer ah.Unlock()
	ah.Lock()
	id, err := getTitle(r)
	if err != nil {
		respondError(w, http.StatusNotFound, "Title of the article not found.")
		return
	}
	respondJSON(w, http.StatusOK, ah.articles[id])
	fmt.Println(ah.articles[id])
}
func (ah *ArticlesHandler) createArticle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "createArticle is called.")
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	content := r.Header.Get("application/json")
	if content != "application/json" {
		respondError(w, http.StatusUnsupportedMediaType, "Content is not in 'application/json' format.")
		return
	}
	var article Article
	err = json.Unmarshal(body, &article)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	ah.articles = append(ah.articles, article)
	respondJSON(w, http.StatusCreated, article)
}
func (ah *ArticlesHandler) modifyArticle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "modifyArticle is called.")
}
func (ah *ArticlesHandler) deleteArticle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "deleteArticle is called.")
}

func getTitle(r *http.Request) (int, error) {
	parts := strings.Split(r.URL.String(), "/")
	fmt.Println(parts)
	if len(parts) < 2 {
		return 0, errors.New("Not found.")
	}
	id, err := strconv.Atoi(parts[2])
	fmt.Println(id, "------", parts[2])
	if err != nil {
		return 0, errors.New("Not found.")
	}
	return id, nil
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

//Invoke-RestMethod ...
