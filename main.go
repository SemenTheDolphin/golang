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
	CommentsCount int
	Id int    `json:"id"`
	FromId int    `json:"from_id"`
	Text string `json:"text"`
}

//Printing and saving new last posts and comments
func printNewPostsAndComments(accessToken string, groupUrl string, wall Resp) Resp {
	fmt.Println("...")
	newWall := getJson(groupUrl)
	newWall = setCommentsCount(accessToken, newWall)
	var j int
	for j = 0; j < (newWall.Response.Count - wall.Response.Count); j++ {
		printPostAndComments(accessToken, newWall, j, 0, false)
	}
	for i := j; i < len(wall.Response.Items)-j; i++ {
		if newWall.Response.Items[i].CommentsCount > wall.Response.Items[i-j].CommentsCount {
			printPostAndComments(accessToken, newWall, i, wall.Response.Items[i-j].CommentsCount, true)
		}
	}
	return newWall
}
//Printing comments for i's post (if offset>0 - printing only new comments)
func printComments(accessToken string, wall Resp, i int, count int, offset int) {
	for j := 0; j < int(math.Ceil(float64(count-offset)/float64(100))); j++ {
		commentsUrl := fmt.Sprintf("https://api.vk.com/method/wall.getComments?offset=%d&owner_id=%d&post_id=" +
			"%d&sort=asc&count=10&v=5.68&access_token=%s",
			j*100 + offset, wall.Response.Items[i].FromId, wall.Response.Items[i].Id,
			accessToken)

		comm := getJson(commentsUrl)
		time.Sleep(time.Millisecond*500)
		for k := 0; k < len(comm.Response.Items); k++ {
			fmt.Printf("\n        [%d]: %s",
				comm.Response.Items[k].Id, comm.Response.Items[k].Text)
		}
	}
}
//Printing i's post and new/old comments for it
func printPostAndComments(accessToken string, wall Resp, i int, offset int, isOld bool)  {
	if isOld {
		fmt.Printf("\n\n[%d (Старый пост)]: %s",
			wall.Response.Items[i].Id, wall.Response.Items[i].Text)
	} else {
		fmt.Printf("\n\n[%d (Новый пост)]: %s",
			wall.Response.Items[i].Id, wall.Response.Items[i].Text)
	}

	fmt.Println("\nКомментарии: ")
		printComments(accessToken, wall, i, wall.Response.Items[i].CommentsCount, offset)
}
//Getting Resp struct from api request for comments or posts
func getJson(url string) Resp {
	r, err := http.Get(url)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return Resp{}
	}

	var resp Resp

	if len(body) > 0 {
		err = json.Unmarshal(body, &resp)
		if err != nil {
			fmt.Printf("Error: %v", err)
			return Resp{}
		}
	}
	return resp
}
//Setting comments count for each post
func setCommentsCount(accessToken string, wall Resp) Resp {
	for i := 0; i < len(wall.Response.Items); i++ {
		wall.Response.Items[i].CommentsCount = getCommentsCount(accessToken, wall, i)
	}
	return wall
}
//Getting group url for count posts
func getGroupUrl(accessToken string, count int) string {
	if len(os.Args) == 1 {
		fmt.Println("Error: no group adress!\nPress any key...")
		fmt.Scanln()
		os.Exit(0)
	}

	u, _ := url.Parse(os.Args[1])
	domain := strings.Replace(u.Path, "/", "", 1)
	groupUrl := fmt.Sprintf("https://api.vk.com/method/wall.get?domain=%s&count=%d&v=5.68&access_token=%s",
		domain, count, accessToken)
	return groupUrl
}
//Getting comments count for i's post
func getCommentsCount (accessToken string, wall Resp, i int) int {
	commentsUrl := fmt.Sprintf("https://api.vk.com/method/wall.getComments?owner_id=%d&post_id=%d&sort=asc" +
		"&count=1&v=5.68&access_token=%s",
		wall.Response.Items[i].FromId, wall.Response.Items[i].Id,
		accessToken)
	comm := getJson(commentsUrl)
	time.Sleep(time.Millisecond*500)
	return comm.Response.Count
}
//Initialization and printing first posts
func initialize(accessToken string, groupUrl string) Resp {
	wall := getJson(groupUrl)
	wall = setCommentsCount(accessToken, wall)

	for i := 0; i < len(wall.Response.Items); i++ {
		printPostAndComments(accessToken, wall, i,0, false)
	}
	return wall
}

func main() {
	accessToken := os.Getenv("VK_TOKEN")

	groupUrl := getGroupUrl(accessToken, 10)
	wall := initialize(accessToken, groupUrl)

	for {
		time.Sleep(time.Second * 15)
		wall = printNewPostsAndComments(accessToken, groupUrl, wall)
	}
}
