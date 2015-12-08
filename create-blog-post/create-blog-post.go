package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/wm/release-blogger/confluence"
	"log"
	"os"
)

func displayMessageData(content confluence.Content) {
	fmt.Println("Published Content:", content.Type, content.Title, content.Body.Storage.Value)
}

func main() {
	app := cli.NewApp()
	app.Name = "create-blog"
	app.Usage = "Creates a new blog entry within the given confluence space.\n"
	app.Usage += "EXAMPLE:\n"
	app.Usage += "   create-blog --server 'https://mycompany.atlassian.net' --username jdoe --password secret --space ENG --title 'My First Post' --body '<p>check this out</p><p>cool huh</p>'\n"
	app.Version = "0.1"
	app.Author = "Will Mernagh"
	app.Email = "wmernagh@gmail.com"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "space, s",
			Value: "ENG",
			Usage: "The space to which you are posting the blog entry",
		},
		cli.StringFlag{
			Name:  "webserver, w",
			Value: "https://mycompany.atlassian.net",
			Usage: "Web server hosting your instance of Confluence",
		},
		cli.StringFlag{
			Name:  "username, u",
			Value: "jdoe",
			Usage: "Your username",
		},
		cli.StringFlag{
			Name:  "password, p",
			Value: "",
			Usage: "Your password",
		},
		cli.StringFlag{
			Name:  "title, t",
			Value: "My first command line blog post",
			Usage: "Specify the title for the post you are creating",
		},
		cli.StringFlag{
			Name:  "body, b",
			Value: "<p>Main content will appear here</p>",
			Usage: "The content of the post",
		},
	}

	app.Action = func(c *cli.Context) {
		username := c.String("username")
		password := c.String("password")
		server := c.String("webserver")
		space := &confluence.Space{c.String("space")}
		title := c.String("title")
		storage := &confluence.Storage{c.String("body"), "storage"}
		body := &confluence.Body{*storage}
		contentIn := &confluence.Content{"blogpost", title, *space, *body}

		client := confluence.NewClient(username, password, server)

		contentOut, _, err := client.Content.Create(contentIn)
		if err != nil {
			log.Fatal("Create Error:", err)
		}
		displayMessageData(*contentOut)
	}

	app.Run(os.Args)
}
