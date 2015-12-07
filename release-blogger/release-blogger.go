package main

import (
	"github.com/codegangsta/cli"
	"github.com/wm/release-blogger/confluence"
	"github.com/wm/release-blogger/server"
	"log"
	"os"
)

func releaseToContent(c *cli.Context, event *server.ReleaseEvent, client *confluence.Client) (*confluence.Content, error) {
	space := &confluence.Space{c.String("space")}
	title := event.Release.Name
	storage := &confluence.Storage{*event.Release.Body, "storage"}
	body := &confluence.Body{*storage}
	contentIn := &confluence.Content{"blogpost", *title, *space, *body}

	contentOut, _, err := client.Content.Create(contentIn)
	return contentOut, err
}

func main() {

	app := cli.NewApp()
	app.Name = "release-blogger"
	app.Usage = "A web service that listents to web hook calls from Github that are releases and pushes them to a confluence blog\n\n"
	app.Usage += "EXAMPLE:\n"
	app.Usage += "    release-blogger \\ \n"
	app.Usage += "      --port 8080 \\ \n"
	app.Usage += "      --secret my-github-hmac-secret \\ \n"
	app.Usage += "      --space BLOG \\ \n"
	app.Usage += "      --url https://mycompany.atlassian.net \\ \n"
	app.Usage += "      --username my-confluence-username \\ \n"
	app.Usage += "      --password my-confluence-password"
	app.Version = "1.0"
	app.Author = "Will Mernagh"
	app.Email = "wmernagh@gmail.com"

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "port, p",
			Value: 8080,
			Usage: "port on which to listen for github webhooks",
		},
		cli.StringFlag{
			Name:  "secret, s",
			Value: "",
			Usage: "Github secret for HMAC verification.",
		},
		cli.StringFlag{
			Name:  "space, S",
			Value: "",
			Usage: "The space to which you are posting the blog entry",
		},
		cli.StringFlag{
			Name:  "url, u",
			Value: "",
			Usage: "Web url hosting your instance of Confluence",
		},
		cli.StringFlag{
			Name:  "username, U",
			Value: "",
			Usage: "Your confluence username",
		},
		cli.StringFlag{
			Name:  "password, P",
			Value: "",
			Usage: "Your confluence password",
		},
	}

	app.Action = func(c *cli.Context) {
		server := server.NewServer()
		server.Port = c.Int("port")
		server.Secret = c.String("secret")
		server.GoListenAndServe()

		username := c.String("username")
		password := c.String("password")
		confluenceUrl := c.String("url")
		client := confluence.NewClient(username, password, confluenceUrl)

		for event := range server.Events {
			content, err := releaseToContent(c, &event, client)
			if err != nil {
				log.Printf("Create Blog Error: %v\nFor Event: %v\n", err, event)
			} else {
				log.Printf("Event published as blog entry %v", content.Title)
			}
		}
	}

	app.Run(os.Args)
}
