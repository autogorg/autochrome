package chrome

import (
	"fmt"
	"os"
	"context"
	"time"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

type Chrome struct {
	Width   int
	Height  int
	Url     string
	Html    string
	Context context.Context
	Cancel  context.CancelFunc
}


func unused() {
	fmt.Println(os.Getenv("PATH"))
	fmt.Println(kb.Enter)
	time.Sleep(1*time.Second)
}

func New() *Chrome {
	return &Chrome{Width:800, Height:600}
}

func Delete(c *Chrome) {
	if c.Cancel != nil {
		c.Cancel()
	}
}

func (c *Chrome) SetSize(width, height int) {
	c.Width  = width
	c.Height = height
}


func (c *Chrome) SetUrl(url string) {
	c.Url = url
}

func (c *Chrome) GetHtml() string {
	return c.Html
}

func (c *Chrome) NewTab() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.IgnoreCertErrors,
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-web-security", true),
		chromedp.WindowSize(c.Width, c.Height),
	)
	c.Context, c.Cancel = chromedp.NewExecAllocator(context.Background(), opts...)
	c.Context, c.Cancel = chromedp.NewContext(c.Context)
}

func (c *Chrome) NavigateAndWaitReady() string {
	err := chromedp.Run(c.Context,
		chromedp.Navigate(c.Url),
		// Wait document ready
		chromedp.Evaluate(`document.readyState === "complete"`, nil),
		// Read outerHTML
		chromedp.OuterHTML("html", &c.Html),
	)

	if err != nil {
		return fmt.Sprintf("%s", err)
	}

	return ""
}

func (c *Chrome) RunTasks(fun func (ctx context.Context) error) string {
	if fun == nil {
		return "Task fun is nil!"
	}
	err := fun(c.Context)

	if err != nil {
		return fmt.Sprintf("%s", err)
	}

	return ""
}

