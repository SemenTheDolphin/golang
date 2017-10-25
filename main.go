package main

import (
	"fmt"
	"time"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"os"
	"net/url"
	"strings"
	"math"
)

type Resp struct {
	Response Answer `json:"response"`
}
type Answer struct {
	Count int `json:"count"`
	Items []Item `json:"items"`
}
type Item struct {
	Comments CommentsCount `json:"comments"`
	Id int    `json:"id"`
	FromId int    `json:"from_id"`
	Text string `json:"text"`
}
type CommentsCount struct {
	Count int `json:"count"`
}
var accessToken = os.Getenv("VK_TOKEN")
var cURL = "https://api.vk.com/method/wall.getComments?v=5.68&count=100&access_token=" + accessToken
var gURL = "https://api.vk.com/method/wall.get?v=5.68&access_token=" + accessToken

//Printing and saving new last posts and comments
func printNewPostsAndComments(groupUrl string, wall *Resp) Resp {
	fmt.Println("\n...")
	newWall, err := getJson(groupUrl)
	if err != nil {
		fmt.Printf("Error: %v!\nPress Enter to exit...", err)
		fmt.Scanln()
		os.Exit(0)
	}

	var j int
	for j = 0; j < (newWall.Response.Count - wall.Response.Count); j++ {
		printPostAndComments(newWall.Response.Items[j], 0, false)
	}
	for i := j; i < len(wall.Response.Items)-j; i++ {
		if newWall.Response.Items[i].Comments.Count > wall.Response.Items[i-j].Comments.Count {
			printPostAndComments(newWall.Response.Items[i], wall.Response.Items[i-j].Comments.Count, true)
		}
	}
	return newWall
}

//Printing comments for i's post (if offset>0 - printing only new comments)
func printComments(post Item, count int, offset int) {
	for j := 0; j < int(math.Ceil(float64(count-offset)/float64(100))); j++ {
		comm, err := getJson(fmt.Sprintf(cURL + "&offset=%d&owner_id=%d&post_id=%d",
			j*100+offset, post.FromId, post.Id))
		if err != nil {
			fmt.Printf("Error: %v!\nPress Enter to exit...", err)
			fmt.Scanln()
			os.Exit(0)
		}

		for _, comment := range comm.Response.Items {
			fmt.Printf("        [%d]: %s\n", comment.Id, comment.Text)
		}
		time.Sleep(time.Millisecond*500)
	}
}

//Printing i's post and new/old comments for it
func printPostAndComments(post Item, offset int, isOld bool)  {
	if isOld {
		fmt.Printf("\n\n[%d (Старый пост)]: %s\n",
			post.Id, post.Text)
	} else {
		fmt.Printf("\n\n[%d (Новый пост)]: %s\n",
			post.Id, post.Text)
	}
	fmt.Println("Комментарии: ")
		printComments(post, post.Comments.Count, offset)
}

//Getting Resp struct from api request for comments or posts
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
	}
	return
}

//Getting group url for count posts
func getGroupUrl(count int) (groupUrl string, err error) {
	if len(os.Args) == 1 {
		fmt.Print("Error: No group adress!\nPress Enter to exit...")
		fmt.Scanln()
		os.Exit(0)
	}

	if len(os.Args) > 2 {
		fmt.Print("Error: Too many arguments!\nPress Enter to exit...")
		fmt.Scanln()
		os.Exit(0)
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
	groupUrl = fmt.Sprintf(gURL + "&domain=%s&count=%d", domain, count)

	return
}

//Initialization and printing first posts
func initialize(groupUrl string) Resp {
	wall, err := getJson(groupUrl)
	if err != nil {
		fmt.Printf("Error: %v!\nPress Enter to exit...", err)
		fmt.Scanln()
		os.Exit(0)
	}
	for _, post := range wall.Response.Items {
		printPostAndComments(post,0, false)
	}
	return wall
}

func main() {
	groupUrl, err := getGroupUrl(10)
	if err != nil {
		fmt.Printf("Error: %v!\nPress Enter to exit...", err)
		fmt.Scanln()
		os.Exit(0)
	}

	wall := initialize(groupUrl)

	for {
		time.Sleep(time.Second * 15)
		printNewPostsAndComments(groupUrl, &wall)
	}
}
