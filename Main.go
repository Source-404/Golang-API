package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

type Post struct{
	ID        	primitive.ObjectID 	`json:"_id,omitempty" bson:"_id,omitempty"`
	Caption 	string             	`json:"caption,omitempty" bson:"caption,omitempty"`
	URL 		string             	`json:"url,omitempty" bson:"url,omitempty"`
	CreatedDate time.Time   		`json:"createdDate" bson:"createdDate,omitempty"`	
}


type User struct{
	ID        primitive.ObjectID 	`json:"_id,omitempty" bson:"_id,omitempty"`
	Name 		string				`json:"name,omitempty" bson:"name,omitempty"`
	Email 		string             	`json:"email,omitempty" bson:"email,omitempty"`
	Password 	string             	`json:"password,omitempty" bson:"password,omitempty"`	
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

func createUser(w http.ResponseWriter, r *http.Request){
	w.Header().Add("content-type","application/json")
	if r.Method !="POST"{
		http.Error(w, "method is not supported", http.StatusNotFound)
		return
	}
	var user User
	json.NewDecoder(r.Body).Decode(&user)
	user.Password,_ = HashPassword(user.Password)
	collection := client.Database("thegodb").Collection("users")
	ctx,_ :=context.WithTimeout(context.Background(),10*time.Second)
	result,_ :=collection.InsertOne(ctx,user)
	json.NewEncoder(w).Encode(result)
}

func getUserById(w http.ResponseWriter, r *http.Request){
	w.Header().Add("content-type","application/json")
	if r.Method !="GET"{
		http.Error(w, "method is not supported", http.StatusNotFound)
		return
	}
	parts := strings.Split(r.URL.String(),"/")
	if len(parts) !=3{
		w.WriteHeader(http.StatusNotFound)
		return 
	}
	id,_ :=primitive.ObjectIDFromHex(parts[2])
	var user User
	collection := client.Database("thegodb").Collection("users")
	ctx,_ :=context.WithTimeout(context.Background(),10*time.Second)
	
	err := collection.FindOne(ctx, User{ID: id}).Decode(&user)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(w).Encode(user)
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

func createPost(w http.ResponseWriter, r *http.Request){
	w.Header().Add("content-type","application/json")
	if r.Method !="POST"{
		http.Error(w, "method is not supported", http.StatusNotFound)
		return
	}
	var post Post
	json.NewDecoder(r.Body).Decode(&post)
	post.CreatedDate = time.Now()
	collection := client.Database("thegodb").Collection("posts")
	ctx,_ :=context.WithTimeout(context.Background(),10*time.Second)
	result,_ :=collection.InsertOne(ctx,post)
	json.NewEncoder(w).Encode(result)
}

func getPostById(w http.ResponseWriter, r *http.Request){
	w.Header().Add("content-type","application/json")
	if r.Method !="GET"{
		http.Error(w, "method is not supported", http.StatusNotFound)
		return
	}
	parts := strings.Split(r.URL.String(),"/")
	if len(parts) !=3{
		w.WriteHeader(http.StatusNotFound)
		return 
	}
	id,_ :=primitive.ObjectIDFromHex(parts[2])
	var post Post
	collection := client.Database("thegodb").Collection("posts")
	ctx,_ :=context.WithTimeout(context.Background(),10*time.Second)
	
	err := collection.FindOne(ctx, Post{ID: id}).Decode(&post)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(w).Encode(post)
}

func getAllPosts(w http.ResponseWriter, r *http.Request){
	w.Header().Add("content-type","application/json")
	if r.Method !="GET"{
		http.Error(w, "method is not supported", http.StatusNotFound)
		return
	}
	var posts []Post
	collection := client.Database("thegodb").Collection("posts")
	ctx,_ :=context.WithTimeout(context.Background(),10*time.Second)
	cursor, err :=collection.Find(ctx,bson.M{})
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	
	defer cursor.Close(ctx)
	for cursor.Next(ctx){
		var post Post
		cursor.Decode(&post)
		posts = append(posts, post)
	}
	if err := cursor.Err(); err !=nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(w).Encode(posts)
	
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

var client *mongo.Client

func homeHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprint(w,"<h1>welcome on board</h1>")
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}


func main() {
	clientLocal, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err!=nil{
		panic(err)
	}
	client = clientLocal
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err!=nil{
		panic(err)
	}

	fmt.Println("Server is up on port: 3000")

	http.HandleFunc("/",homeHandler)
	http.HandleFunc("/users",createUser)
	http.HandleFunc("/users/",getUserById)

	http.HandleFunc("/posts",createPost)
	http.HandleFunc("/posts/",getPostById)

	http.HandleFunc("/posts/users/",getAllPosts)

	http.ListenAndServe(":3000",nil)
}