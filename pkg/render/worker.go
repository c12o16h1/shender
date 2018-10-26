package render

import (
	"log"
	"time"
	"context"

	"github.com/chromedp/chromedp/runner"
	"github.com/chromedp/chromedp"

	"github.com/c12o16h1/shender/pkg/models"
)

type Worker struct {
	models.Renderer
	models.Closer
	worker  *chromedp.CDP
	context *context.Context
}

func NewWorker() (*Worker, error) {
	ctxt, cancel := context.WithCancel(context.Background())
	c, err := chromedp.New(ctxt, chromedp.WithLog(log.Printf), chromedp.WithRunnerOptions(
		runner.Flag("headless", false),
		runner.Flag("disable-gpu", false),
	))
	if err != nil {
		return nil, err
	}

	w := Worker{
		worker:  c,
		context: &ctxt,
	}
	w.Close = cancel
	return &w, nil
}

func (w *Worker) Render(url string) (string, error) {
	var body string
	w.worker.Run(w.context, renderTasks(url, &body))
	return body, nil
}

func renderTasks(url string, body *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Sleep(2 * time.Second),
		chromedp.InnerHTML("html", body),
	}
}
