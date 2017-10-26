package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Resp struct {
	Response Answer      `json:"response"`
	Error    ErrorAnswer `json:"error"`
}
type ErrorAnswer struct {
	ErrorMsg string `json:"error_msg"`
}
type Answer struct {
	Count int    `json:"count"`
	Items []Item `json:"items"`
}
type Item struct {
	Comments CommentsCount `json:"comments"`
	ID       int           `json:"id"`
	FromId   int           `json:"from_id"`
	Text     string        `json:"text"`
}
type CommentsCount struct {
	Count int `json:"count"`
}

var (
	accessToken = os.Getenv("VK_TOKEN")
	vkApiURL    = "https://api.vk.com/method/"
	cURL        = vkApiURL + "wall.getComments?v=5.68&count=100&access_token=" + accessToken
	gURL        = vkApiURL + "wall.get?v=5.68&access_token=" + accessToken
)

// Printing and saving new last posts and comments.
func printNewPostsAndComments(groupURL string, wall *Resp) {
	fmt.Println("\n...")
	newWall, err := getJson(groupURL)
	if err != nil {
		log.Fatal(err)
	}
	j := 0
	for i, post := range newWall.Response.Items {
		if j < (newWall.Response.Count - wall.Response.Count) {
			printPostAndComments(post, 0, "Новый пост")
			j++
		} else if post.Comments.Count > wall.Response.Items[i-j].Comments.Count {
			printPostAndComments(post, wall.Response.Items[i-j].Comments.Count, "Старый пост")
		}
	}
	*wall = newWall
}

// Printing comments for post (if offset > 0 - printing only new comments).
func printComments(post Item, offset int) {
	for j := 0; j <= (post.Comments.Count-offset)/100; j++ {
		comm, err := getJson(fmt.Sprintf(cURL+"&offset=%d&owner_id=%d&post_id=%d",
			j*100+offset, post.FromId, post.ID))
		if err != nil {
			log.Fatal(err)
		}

		for _, comment := range comm.Response.Items {
			fmt.Printf("        [%d]: %s\n", comment.ID, comment.Text)
		}
		time.Sleep(time.Millisecond * 500)
	}
}

// Printing new/old post and only new comments for it.
func printPostAndComments(post Item, offset int, postState string) {
	fmt.Printf("\n\n[%d (%s)]: %s\n", post.ID, postState, post.Text)
	fmt.Println("Комментарии: ")
	printComments(post, offset)
}

// Getting Resp struct from api request for comments or posts.
func getJson(url string) (resp Resp, err error) {
	r, err := http.Get(url)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	if len(body) > 0 {
		err = json.Unmarshal(body, &resp)
		if err != nil {
			return
		}
		if resp.Error.ErrorMsg != "" {
			log.Fatal(resp.Error.ErrorMsg)
		}
	}
	return
}

// Getting group url for count posts.
func getGroupURL(count int) (groupURL string, err error) {
	if len(os.Args) == 1 {
		log.Fatal("Error: No group adress!")
	}

	if len(os.Args) > 2 {
		log.Fatal("Error: Too many arguments!")
	}

	r, err := http.Get(os.Args[1])
	if err != nil {
		return
	}

	_, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	u, err := url.Parse(os.Args[1])
	if err != nil {
		return
	}

	domain := strings.Replace(u.Path, "/", "", 1)
	groupURL = fmt.Sprintf(gURL+"&domain=%s&count=%d", domain, count)
	return
}

// Initialization and printing first posts.
func initialize(groupURL string) (wall Resp) {
	wall, err := getJson(groupURL)
	if err != nil {
		log.Fatal(err)
	}

	for _, post := range wall.Response.Items {
		printPostAndComments(post, 0, "Новый пост")
	}
	return
}

func main() {
	groupURL, err := getGroupURL(10)
	if err != nil {
		log.Fatal(err)
	}

	wall := initialize(groupURL)

	for {
		time.Sleep(time.Second * 15)
		printNewPostsAndComments(groupURL, &wall)
	}
}
