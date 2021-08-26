package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var (
	ErrGeneric = "Something Went Wrong"
	ErrNotFoud = "Article Not Found"
)

type Article struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	Content string `json:"content"`
}

type ArticleStore struct {
	Articles map[int]Article
}

func newArticleStore() *ArticleStore {
	return &ArticleStore{
		Articles: make(map[int]Article),
	}
}

func (as *ArticleStore) addArticleToStore(a Article) {
	as.Articles[a.ID] = a
}

func (as *ArticleStore) getArticles(w http.ResponseWriter, r *http.Request) {
	articles := struct {
		Articles []Article `json:"articles"`
	}{Articles: make([]Article, 0, len(as.Articles))}
	for _, a := range as.Articles {
		articles.Articles = append(articles.Articles, a)
	}
	json.NewEncoder(w).Encode(articles)
}

func (as *ArticleStore) getArticleByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, ErrGeneric, http.StatusInternalServerError)
		return
	}
	article, ok := as.Articles[id]
	if !ok {
		http.Error(w, ErrNotFoud, http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(article)
}

func (as *ArticleStore) createArticle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, ErrGeneric, http.StatusInternalServerError)
		return
	}
	var article Article
	if err := json.Unmarshal(b, &article); err != nil {
		http.Error(w, ErrGeneric, http.StatusInternalServerError)
		return
	}
	as.addArticleToStore(article)
}

func (as *ArticleStore) deleteArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, ErrGeneric, http.StatusInternalServerError)
		return
	}
	delete(as.Articles, id)
}

func (as *ArticleStore) updateArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, ErrGeneric, http.StatusInternalServerError)
		return
	}
	article, ok := as.Articles[id]
	if !ok {
		http.Error(w, ErrNotFoud, http.StatusNotFound)
		return
	}
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, ErrGeneric, http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(b, &article); err != nil {
		http.Error(w, ErrGeneric, http.StatusInternalServerError)
		return
	}
	as.addArticleToStore(article)
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Articles API Homepage")
}

func health(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s\n", r.Method, r.URL.Path)
	response := "I'm healthy for now..."
	log.Println(response)
	fmt.Fprintln(w, response)
}

func main() {
	port := flag.String("port", "5000", "port for the http server")
	flag.Parse()

	as := newArticleStore()
	articles := []Article{
		Article{ID: 1, Title: "Article Title 1", Desc: "Article Description 1", Content: "Article Content 1"},
		Article{ID: 2, Title: "Article Title 2", Desc: "Article Description 2", Content: "Article Content 2"},
		Article{ID: 3, Title: "Article Title 3", Desc: "Article Description 3", Content: "Article Content 3"},
	}
	for _, a := range articles {
		as.addArticleToStore(a)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", home)
	r.HandleFunc("/healthz", health)
	r.HandleFunc("/articles", as.getArticles)
	r.HandleFunc("/article/{id:[0-9]+}", as.getArticleByID).Methods("GET")
	r.HandleFunc("/article/{id:[0-9]+}", as.deleteArticle).Methods("DELETE")
	r.HandleFunc("/article/{id:[0-9]+}", as.updateArticle).Methods("PUT")
	r.HandleFunc("/article", as.createArticle).Methods("POST")

	log.Printf("Starting Server on port %s", *port)
	if err := http.ListenAndServe(":"+*port, r); err != nil {
		log.Fatalf("ListenAndServe err = %s", err)
	}
}
