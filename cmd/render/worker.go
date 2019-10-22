package main

import (
	"context"
	"os"
	"time"

	"github.com/c12o16h1/shender/pkg/models"

	"github.com/chromedp/chromedp"
)

// Worker is a wrapper for headless Chrome instance
type Worker struct {
	models.Renderer
	worker  *chromedp.Pool
	context *context.Context
	cancel  *context.CancelFunc
	created time.Time
}

// Spawn new worker instance
func NewWorker(start int, end int) (*Worker, error) {
	ctxt, cancel := context.WithCancel(context.Background())
	c, err := chromedp.NewPool(chromedp.PortRange(start, end))
	if err != nil {
		return nil, err
	}

	w := Worker{
		worker:  c,
		context: &ctxt,
		cancel:  &cancel,
		created: time.Now(),
	}
	return &w, nil
}

// Very basic worker function to get page source
func (w *Worker) Close(sig int, out *string) error {
	w.worker.Shutdown()
	os.Exit(sig)
	return nil
}

// Very basic worker function to get page source
func (w *Worker) Render(url string, html *string) error {
	c, err := w.worker.Allocate(*w.context)
	if err != nil {
		return err
	}
	defer c.Release()
	c.Run(*w.context, renderTasks(url, html))
	return nil
}

// Check that worker is alive
func (w *Worker) Heartbeat(in string, out *string) error {
	*out = models.OK
	return nil
}

func renderTasks(url string, body *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Sleep(2 * time.Second),
		chromedp.InnerHTML("html", body),
	}
}
