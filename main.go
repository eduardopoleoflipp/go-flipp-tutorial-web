/*
  INTRO:

  The following shows you how to build a simple web service in go. Before we head in some important considerations:

  - This tutorial is not meant to be production ready but to introduce new comers to some of the most relevant concepts of the Go programming language
  - It assumes a minimum level of programming. This is not a "Learn how to program guide in go".
  - It caters to Flipp. This last will be more evident in later sections of the tutorial.

  I recommend that you first have a read through the whole tutorial to have an idea of what we'll be building and then try to recreate it by yourself, peeking when needed. That way you can build muscle memory with the language.
*/

/*
  THE APP:

  We'll be building a simple web app that:
  - Returns a list of posts
  - Returns a particular post given a path and an ID and update some post's visibility metrics.
  - Creates a post given some information.
  - For simplicity the posts will be stored in a local json file.
*/

/*
  Packages are a way to logically group functions. The "main" package is default entry point in a go program.
*/
package main

/*
  We can import packages from the standard library. IDE support for Go is usually very robust, that and the fact that the language is statically type means that you can hover the package to read their description. You can also check the online documentation by right cmd+click into it.
*/
import (
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "os"
  "time"
)

/*

  STRUCTS

  Since we're creating a blog we'll need a way to easily manage the post data as it comes in or out of the app, we'll be using structs for this purpose. Structs are equivalent to value objects in other programming languages. They provide simple getters and setters for its properties.

  One important thing about Go properties naming convention and package visibility rules. Struct properties need to be capitalized if you're planning to access them outside the package the struct is defined. Otherwise GO will assume that they are private and be only accessible within the package the struct is defined. This is not relevant for our toy example since our program only have one package but it's important information to know.

  Notice the `json:property` annotation beside every property, these annotations are called struct tags. In this case they provide the corresponding json mapping used when serializing / deserializing the post information.
*/

type Post struct {
  Title      string `json:"Title"`
  Content    string `json:"Content"`
  CreatedAt  string `json:"CreatedAt"`
  Author     string `json:"Author"`
  ViewCount  int    `json:"ViewCount"`
  LastViewed string `json:"LastViewed"`
}

/*
  STRUCT BEHAVIOUR

  It's possible to add behaviour to structs by using pointer receivers functions. A pointer receiver function can be generally defined as:

  func (objectName *ObjectType) functionName() returnValue ReturnValueType {
    objectName.prop1
    objectName.prop2
  }

  Adding (post *Post) to the function signature indicates that all instances of Post will be able to receive the increaseViewCount() method.

  We'll talk more about the usage of pointers in GO later in this tutorial, for now take a look at the corresponding post functions below.
*/
func (post *Post) increaseViewCount() {
  post.ViewCount += 1
}

func (post *Post) setLastViewed() {
  post.LastViewed = time.Now().Format("2006-01-02")
}

func (post *Post) setCreatedAt() {
  post.CreatedAt = time.Now().Format("2006-01-02")
}

/*
  GLOBAL PACKAGE VARIABLES

  This is a global variable and will be available in all functions within this package.
*/
var (
  filePath string = "posts.json"
)

/*
  MAIN FUNCTION

  The "main" function is the program's entry point. The program execution will always start here.
*/
func main() {
  /*
    The simplest way to setup a web server is by using the http.HandleFunc which takes in a path and a handler function for that particular request. In our case we'll have three different routes one for every feature we'll be supporting:
    - List Posts
    - Create a Post

    It's worth noting that, unlike ruby, functions in go are first class citizens, meaning that you can pass them as arguments to other functions. That's why we're able to provide handler functions.
  */
  http.HandleFunc("/index", index)
  http.HandleFunc("/create", create)

  // The fmt package offers methods to print info to stdout
  fmt.Println("Server running on http://localhost:3000")
  // Finally we're ready to listen for request and sever responses
  http.ListenAndServe(":3000", nil)
}

/*
  INDEX HANDLER

  The index function that will be handling the index response.
*/
func index(w http.ResponseWriter, r *http.Request) {
  /*
    You can define variables ahead of time this way. In most cases you need to provide the type as part of the definition.

    In this case we're declaring a post slice which is a dynamic type of list which types can grow or shrink as needed. This is not to be confused with arrays which should have fixed size that must be declared at creation time. Our example requires a slice because the number of posts is variable.
  */
  var posts []Post
  /*
    GO POINTERS

    Similar to C in go you can access the reference of a piece of data by using the & operator. One of the most common use cases to do this is when you want to mutate the variable that is being passed into a function. If you do not do this, Go will pass a copy of the value instead, and any modifications will only affect the copy, not the original variable. For a more in depth explanation on the topic read https://www.digitalocean.com/community/conceptual-articles/understanding-pointers-in-go.

    In our particular example we want to load all the posts into the posts variable passed in to have them available within the scope of the index function.
  */
  loadPost(&posts, w)

  for i := 0; i < len(posts); i++ {
    // We can declare variables using the short variable declaration operator := . In this case go will inference the variable type based on the value assigned so it's not required to explicitly define the type at declaration time.
    post := &posts[i]
    /*
      We're using the receiver functions declared above to modify the ViewCount and LastView properties.
    */
    post.increaseViewCount()
    post.setLastViewed()
  }

  // Saves the post to the file.
  savePosts(posts)

  // We set the response headers to json so that the browser knows what kind of data we're returning
  w.Header().Set("Content-Type", "application/json")
  // Finally we marshall back the posts to json into the response
  json.NewEncoder(w).Encode(posts)
}

/*
  CREATE HANDLER

  Creates a Post with the given information.
*/
func create(w http.ResponseWriter, r *http.Request) {
  // We make sure that you can only access this function through a post request.
  if r.Method != http.MethodPost {
    http.Error(w, "Please submit a post request", http.StatusMethodNotAllowed)
    return
  }

  /*
    Uses the io package to read the local file where the posts are being saved. Notice that Go supports multiple return values and parallel assignment.
    In this case we're reading the request Body which contains the post params and assign it to the body variable.
  */

  body, err := io.ReadAll(r.Body)
  // This is the common pattern for error handling in Go. Normally methods will return an error object and the caller checks if the error is nil.
  if err != nil {
    http.Error(w, "Error reading request body", http.StatusBadRequest)
    return
  }
  /*
    The defer keyword schedules a function call (in this case, r.Body.Close()) to run after the surrounding function exits, regardless of whether it exits normally or due to an error.
    It's important to close the request body to free resources. We need to do this because we implicitly opened it in the body, err := io.ReadAll(r.Body).
  */
  defer r.Body.Close()

  var newPost Post
  // We then set deserialize the json into a Post struct to be able to access the pointer receiver functions.
  json.Unmarshal(body, &newPost)

  newPost.setCreatedAt()
  newPost.setLastViewed()
  newPost.ViewCount = 0

  /*
    For simplicity sake we're just going to load all the post in memory and the append the new post at the end before saving.
  */
  var posts []Post
  loadPost(&posts, w)
  posts = append(posts, newPost)
  savePosts(posts)

  fmt.Fprintf(w, "Post successfully created")
}

func savePosts(posts []Post) error {
  /*
    Serializes the posts back to a json object
    prefix: "" means that no prefix should be added at the beginning of the line
    indent: "  " means that each level should have a 2 spaces indentation
  */
  data, err := json.MarshalIndent(posts, "", "  ")
  if err != nil {
    return err
  }

  // Writes the post back into the local file
  return os.WriteFile(filePath, data, 0644)
}

/*
  Notice the "posts *[]Post" in the function signature. This is used to indicate that the function expects a reference to the posts slice. See the GO POINTERS comment from above.
*/
func loadPost(posts *[]Post, w http.ResponseWriter) {
  data, err := os.ReadFile(filePath)

  if err != nil {
    http.Error(w, "Error reading request body", http.StatusInternalServerError)
  }

  // This is how we 'transform' the unstructured json into a list of posts structs. The process is commonly referred as unmarshalling or deserialization.
  json.Unmarshal(data, &posts)

  /*
    Contrary to C, you can still use the "."" (dot) operator to access the data from the pointer reference, as oppose to "->". In this case we just need to do post.Title.
  */
  for _, post := range *posts {
    title := post.Title
    fmt.Printf("Loading Post '%s' in memory\n", title)
  }
}
