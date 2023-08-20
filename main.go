package main
 
import (
    "encoding/json"
    "fmt"
    "github.com/lxn/walk"
    . "github.com/lxn/walk/declarative"
    "net/http"
    "os"
    "os/exec"
    "time"
)
 
const appIconResID = 2
 
type News struct {
    Subject string
    Url     string
}
 
func main() {
    var notificationText string
    var news []News
    var threadNumbers []float64
    tray, _ := New()
    _ = tray.Show(appIconResID, "News")
    tray.AppendMenu("Show news", func() {
        createWindow(news)
    })
    tray.AppendMenu("Close", func() {
        os.Exit(0)
    })
    tray.OnClick(func() {
        createWindow(news)
    })
    tray.SetBalloonClick(func() {
        createWindow(news)
    })
    go func() {
        for {
            notificationText = ""
            news = make([]News, 0)
            response, _ := http.Get("https://2ch.hk/news/index.json")
            var data map[string]interface{}
            _ = json.NewDecoder(response.Body).Decode(&data)
            threads := data["threads"].([]interface{})
            for _, thread := range threads {
                posts := thread.(map[string]interface{})["posts"].([]interface{})
                closed := posts[0].(map[string]interface{})["closed"].(float64)
                num := posts[0].(map[string]interface{})["num"].(float64)
                if closed != 1 && !include(num, threadNumbers) {
                    subject := posts[0].(map[string]interface{})["subject"].(string)
                    url := fmt.Sprintf("https://2ch.hk/news/res/%.0f.html", num)
                    notificationText += fmt.Sprintf("%s\n", subject)
                    news = append(news, News{subject, url})
                }
            }
            threadNumbers = make([]float64, 0)
            for _, thread := range threads {
                posts := thread.(map[string]interface{})["posts"].([]interface{})
                closed := posts[0].(map[string]interface{})["closed"].(float64)
                if closed != 1 {
                    num := posts[0].(map[string]interface{})["num"].(float64)
                    threadNumbers = append(threadNumbers, num)
                }
            }
            if len(notificationText) != 0 {
_ = tray.ShowMessage("News", notificationText, false)
}
            time.Sleep(300 * time.Second)
        }
    }()
    _ = tray.Run()
}
 
func include(lhs float64, list []float64) bool {
    for _, rhs := range list {
        if lhs == rhs {
            return true
        }
    }
    return false
}
 
func createWindow(news []News) {
    var newsWidgets []Widget
    for _, newsItem := range news {
        newsWidgets = append(newsWidgets, LinkLabel{
            Text:      fmt.Sprintf(`<a href="%s">%s</a>`, newsItem.Url, newsItem.Subject),
            Alignment: AlignHNearVCenter,
            OnLinkActivated: func(link *walk.LinkLabelLink) {
                _ = exec.Command("rundll32", "url.dll,FileProtocolHandler", link.URL()).Start()
            },
        })
    }
    _, _ = MainWindow{
        Children: newsWidgets,
        Layout:   VBox{},
        Icon:     appIconResID,
        Size:     Size{Width: 400, Height: 200},
        Title:    "News",
    }.Run()
}