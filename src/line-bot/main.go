package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

var port = func() string {
	p := os.Getenv("PORT")
	if p == "" {
		log.Fatal("$PORT must be set")
	}
	return p
}()

var channelID = func() int64 {
	channelID, err := strconv.ParseInt(os.Getenv("LINE_CHANNEL_ID"), 10, 64)
	if err != nil {
		log.Fatal("config error: LINE_CHANNEL_ID", err)
	}
	return channelID
}()

var channelSecret string = os.Getenv("LINE_CHANNEL_SECRET")
var mid string = os.Getenv("LINE_MID")

var proxyURL *url.URL = func() *url.URL {
	p, err := url.Parse(os.Getenv("FIXIE_URL"))
	if err != nil {
		log.Fatal("config error: FIXIE_URL", err)
	}
	return p
}()

func main() {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "hello"})
	})

	router.POST("/callback", func(c *gin.Context) {
		client := &http.Client{
			Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
		}
		bot, err := linebot.NewClient(channelID, channelSecret, mid, linebot.WithHTTPClient(client))
		if err != nil {
			fmt.Println(err)
			return
		}

		received, err := bot.ParseRequest(c.Request)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				fmt.Println(err)
			}
			return
		}
		for _, result := range received.Results {
			content := result.Content()
			if content != nil && content.IsMessage && content.ContentType == linebot.ContentTypeText {
				text, err := content.TextContent()
				res, err := bot.SendText([]string{content.From}, "OK! "+text.Text)
				if err != nil {
					fmt.Println(res)
				}
			}
		}
	})

	router.Run(":" + port)
}
