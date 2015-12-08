package main

import (
	"bytes"
	"github.com/wm/release-blogger/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/wm/release-blogger/confluence"
	"github.com/wm/release-blogger/server"
	"html/template"
	"log"
	"os"
)

const blogPost = `
	<h4>
		<p>
			<img alt="@{{.Sender.Login}}" class="avatar" height="20" src="{{.Sender.AvatarURL}}&amp;s=40" width="20"></img>
			<a href="{{.Sender.URL}}">{{.Sender.Login}}</a> created release <a href="{{.Release.HTMLURL}}">{{.Release.TagName}}</a>
			of <a href="{{.Repo.HTMLURL}}">{{.Repo.Name}}</a>.
		</p>
	</h4>
	<p>{{.Release.Body}}</p>
`

func releaseToContent(c *cli.Context, event *server.ReleaseEvent, client *confluence.Client) (*confluence.Content, error) {
	blogTitle := *event.Repo.Name + " "
	blogTitle += *event.Release.TagName + " - " + *event.Release.Name

	bodyBuff := bytes.NewBufferString("")
	t, err := template.New("blogPost").Parse(blogPost)
	err = t.Execute(bodyBuff, event)
	if err != nil {
		log.Fatal(err)
	}

	space := &confluence.Space{c.String("space")}
	storage := &confluence.Storage{bodyBuff.String(), "storage"}
	body := &confluence.Body{*storage}

	contentIn := &confluence.Content{"blogpost", blogTitle, *space, *body}
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
