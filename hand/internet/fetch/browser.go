package fetch

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

type RenderOptions struct {
	Timeout     time.Duration
	Scroll      bool
	ScrollTimes int
	ScrollDelta int
	ScrollWait  time.Duration
}

func defaultRenderOptions() RenderOptions {
	return RenderOptions{
		Timeout:     30 * time.Second,
		Scroll:      true,
		ScrollTimes: 6,
		ScrollDelta: 800,
		ScrollWait:  600 * time.Millisecond,
	}
}

func RenderWithChromedp(ctx context.Context, rawURL string, opt *RenderOptions) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cfg := defaultRenderOptions()
	if opt != nil {
		if opt.Timeout > 0 {
			cfg.Timeout = opt.Timeout
		}
		cfg.Scroll = opt.Scroll
		if opt.ScrollTimes > 0 {
			cfg.ScrollTimes = opt.ScrollTimes
		}
		if opt.ScrollDelta > 0 {
			cfg.ScrollDelta = opt.ScrollDelta
		}
		if opt.ScrollWait > 0 {
			cfg.ScrollWait = opt.ScrollWait
		}
	}

	cctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	allocatorOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"),
	)
	if proxy := proxyForChromedp(rawURL); proxy != "" {
		allocatorOpts = append(allocatorOpts, chromedp.Flag("proxy-server", proxy))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(cctx, allocatorOpts...)
	defer cancelAlloc()

	bctx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	var rendered string
	tasks := []chromedp.Action{
		chromedp.Navigate(rawURL),
		chromedp.WaitReady("html", chromedp.ByQuery),
		chromedp.Sleep(1200 * time.Millisecond),
	}
	if cfg.Scroll {
		for i := 0; i < cfg.ScrollTimes; i++ {
			tasks = append(tasks,
				chromedp.EvaluateAsDevTools(`window.scrollBy(0, `+itoa(cfg.ScrollDelta)+`);`, nil),
				chromedp.Sleep(cfg.ScrollWait),
			)
		}
	}
	tasks = append(tasks, chromedp.OuterHTML("html", &rendered, chromedp.ByQuery))

	if err := chromedp.Run(bctx, tasks...); err != nil {
		return "", err
	}
	return rendered, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [32]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
