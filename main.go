package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"

	"cloud.google.com/go/bigtable"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	//"github.com/dgrijalva/jwt-go"
	"cloud.google.com/go/storage"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"

	//elastic "gopkg.in/olivere/elastic.v7"
	elastic "github.com/olivere/elastic/v7"
)

const (
	INDEX    = "around"
	TYPE     = "post"
	DISTANCE = "200km"
	// Needs to update
	PROJECT_ID  = "around-422204"
	BT_INSTANCE = "around-post"
	// Needs to update this URL if you deploy it to cloud.
	ES_URL      = "http://haptoai.com:9200"
	BUCKET_NAME = "post-images-750156"
)

// Needs to update this bucket based on your gcs bucket name.

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
type Post struct {
	// `json:"user"` is for the json parsing of this User field. Otherwise, by default it's 'User'.
	User     string   `json:"user"`
	Message  string   `json:"message"`
	Location Location `json:"location"`
	Url      string   `json:"url"`
}

var mySigningKey = []byte("secret")

func main() {

	// Create a client
	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
		//log.Fatalf("Error creating Elasticsearch client: %v", err)
		return
	}
	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(INDEX).Do(context.Background())
	if err != nil {
		panic(err)
	}
	if !exists {
		// Create a new index.
		mapping := `{
			"mappings":{
				"properties":{
					"location":{
						"type":"geo_point"
					}
				}
			}
		}`
		_, err := client.CreateIndex(INDEX).Body(mapping).Do(context.Background())
		if err != nil {
			// Handle error
			panic(err)
		}
	}

	//fmt.Println("started service successfully")
	//http.HandleFunc("/post", handlerPost)
	//http.HandleFunc("/search", handlerSearch)
	//log.Fatal(http.ListenAndServe(":8080", nil))

	// Here we are instantiating the gorilla/mux router
	r := mux.NewRouter()

	var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return mySigningKey, nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})
	r.Handle("/post", jwtMiddleware.Handler(http.HandlerFunc(handlerPost))).Methods("POST")
	r.Handle("/search", jwtMiddleware.Handler(http.HandlerFunc(handlerSearch))).Methods("GET")
	r.Handle("/login", http.HandlerFunc(loginHandler)).Methods("POST")
	r.Handle("/signup", http.HandlerFunc(signupHandler)).Methods("POST")
	http.Handle("/", r)
	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func handlerSearch(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("Received one request for search")
	//lat := r.URL.Query().Get("lat")
	//lon := r.URL.Query().Get("lon")
	//fmt.Fprintf(w, "Search received: %s %s", lat, lon)

	fmt.Println("Received one request for search")
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lon, _ := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	// range is optional
	ran := DISTANCE
	if val := r.URL.Query().Get("range"); val != "" {
		ran = val + "km"
	}

	fmt.Printf("Search received: %f %f %s\n", lat, lon, ran)
	// Create a client
	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
		return
	}

	// Define geo distance query as specified in
	// https://www.elastic.co/guide/en/elasticsearch/reference/5.2/query-dsl-geo-distance-query.html
	q := elastic.NewGeoDistanceQuery("location")
	q = q.Distance(ran).Lat(lat).Lon(lon)
	// Some delay may range from seconds to minutes. So if you don't get enough results. Try it later.
	searchResult, err := client.Search().
		Index(INDEX).
		Query(q).
		Pretty(true).
		Do(context.Background())
	if err != nil {
		// Handle error
		panic(err)
	}
	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch.
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
	// TotalHits is another convenience function that works even when something goes wrong.
	fmt.Printf("Found a total of %d post\n", searchResult.TotalHits())
	// Each is a convenience function that iterates over hits in a search result.
	// It makes sure you don't need to check for nil values in the response.
	// However, it ignores errors in serialization.
	var typ Post
	var ps []Post
	for _, item := range searchResult.Each(reflect.TypeOf(typ)) { // instance of
		p := item.(Post) // p = (Post) item
		fmt.Printf("Post by %s: %s at lat %v and lon %v\n", p.User, p.Message, p.Location.Lat,
			p.Location.Lon)
		// TODO(student homework): Perform filtering based on keywords such as web spam etc.
		ps = append(ps, p)
	}
	js, err := json.Marshal(ps)
	if err != nil {
		panic(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(js)
}

//	fmt.Println("range is ", ran)
//	// Return a fake post
//	p := &Post{
//		User:    "1111",
//		Message: "一生必去的100个地方",
//		Location: Location{
//			Lat: lat,
//			Lon: lon,
//		},
//	}
//	js, err := json.Marshal(p)
//	if err != nil {
//		panic(err)
//		return
//	}
//	w.Header().Set("Content-Type", "application/json")
//	w.Write(js)

//}

//func main() {
//	fmt.Println("Hello, world")
//}

func handlerPost(w http.ResponseWriter, r *http.Request) {

	// Other codes
	//w.Header().Set("Content-Type", "application/json")
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

	// other codes
	user := r.Context().Value("user")
	claims := user.(*jwt.Token).Claims
	username := claims.(jwt.MapClaims)["username"]

	// 32 << 20 is the maxMemory param for ParseMultipartForm, equals to 32MB (1MB = 1024 * 1024 bytes = 2^20 bytes)
	// After you call ParseMultipartForm, the file will be saved in the server memory with maxMemory size.
	// If the file size is larger than maxMemory, the rest of the data will be saved in a system temporary file.
	r.ParseMultipartForm(32 << 20)
	// Parse from form data.
	fmt.Printf("Received one post request %s\n", r.FormValue("message"))
	lat, _ := strconv.ParseFloat(r.FormValue("lat"), 64)
	lon, _ := strconv.ParseFloat(r.FormValue("lon"), 64)
	p := &Post{

		User: username.(string),
		//User:    "1111",
		Message: r.FormValue("message"),
		Location: Location{
			Lat: lat,
			Lon: lon,
		},
	}
	id := uuid.New()
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image is not available", http.StatusInternalServerError)
		fmt.Printf("Image is not available %v.\n", err)
		return
	}
	defer file.Close()
	ctx := context.Background()
	// replace it with your real bucket name.
	_, attrs, err := saveToGCS(ctx, file, BUCKET_NAME, id)
	if err != nil {
		http.Error(w, "GCS is not setup", http.StatusInternalServerError)
		fmt.Printf("GCS is not setup %v\n", err)
		return
	}
	// Update the media link after saving to GCS.
	p.Url = attrs.MediaLink
	// Save to ES.
	saveToES(p, id)
	// Save to BigTable.
	//saveToBigTable(p, id)

	// Parse from body of request to get a json object.
	fmt.Println("Received one post request")
	decoder := json.NewDecoder(r.Body)
	//var p Post
	if err := decoder.Decode(&p); err != nil {
		panic(err)
		return
	}
	//id := uuid.New()
	// Save to ES.
	//saveToES(&p, id)

	fmt.Printf("Post is saved to Index: %s\n", p.Message)
	//ctx := context.Background()
	// you must update project name here
	bt_client, err := bigtable.NewClient(ctx, PROJECT_ID, BT_INSTANCE)
	if err != nil {
		panic(err)
		return
	}
	tbl := bt_client.Open("post")
	mut := bigtable.NewMutation()
	t := bigtable.Now()
	mut.Set("post", "user", t, []byte(p.User))
	mut.Set("post", "message", t, []byte(p.Message))
	mut.Set("location", "lat", t, []byte(strconv.FormatFloat(p.Location.Lat, 'f', 1, 64)))
	mut.Set("location", "lon", t, []byte(strconv.FormatFloat(p.Location.Lon, 'f', 1, 64)))
	err = tbl.Apply(ctx, id, mut)
	if err != nil {
		panic(err)
		return
	}
	fmt.Printf("Post is saved to BigTable: %s\n", p.Message)
	// TODO (student questions) save Post into BT as well
}

func saveToGCS(ctx context.Context, r io.Reader, bucketName, name string) (*storage.ObjectHandle,
	*storage.ObjectAttrs, error) {
	// Student questions

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer client.Close()
	bucket := client.Bucket(bucketName)
	// Next check if the bucket exists
	if _, err = bucket.Attrs(ctx); err != nil {
		return nil, nil, err
	}
	obj := bucket.Object(name)
	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, r); err != nil {
		return nil, nil, err
	}
	if err := w.Close(); err != nil {
		return nil, nil, err
	}
	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return nil, nil, err
	}
	attrs, err := obj.Attrs(ctx)
	fmt.Printf("Post is saved to GCS: %s\n", attrs.MediaLink)
	return obj, attrs, err

}

// Save a post to ElasticSearch
func saveToES(p *Post, id string) {
	// Create a client
	es_client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
		return
	}
	// Save it to index
	_, err = es_client.Index().
		Index(INDEX).
		//		Type(TYPE).
		Id(id).
		BodyJson(p).
		Refresh("true").
		Do(context.Background())
	if err != nil {
		panic(err)
		return
	}
	fmt.Printf("Post is saved to Index: %s\n", p.Message)
}

//	fmt.Fprintf(w, "Post received: %s\n", p.Message)
//}
