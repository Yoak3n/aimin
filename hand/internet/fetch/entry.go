package fetch

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Run(args []string) int {
	return RunWithWriters(args, os.Stdout, os.Stderr)
}

func RunWithWriters(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("fetch", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("out", "", "output directory (optional, empty=do not save files)")
	timeout := fs.Duration("timeout", 20*time.Second, "request timeout")
	maxPages := fs.Int("max-pages", 1, "max pages to fetch")
	depth := fs.Int("depth", 0, "crawl depth")
	delay := fs.Duration("delay", 0, "delay between pages")
	sameDomainOnly := fs.Bool("same-domain", false, "only crawl same domain as start url")
	include := fs.String("include", "", "include url regexp")
	exclude := fs.String("exclude", "", "exclude url regexp")
	followPagination := fs.Bool("follow-pagination", false, "follow rel=next pagination")
	js := fs.Bool("js", true, "enable js render fallback when content is too short")
	jsTimeout := fs.Duration("js-timeout", 30*time.Second, "js render timeout")
	pdfMaxPages := fs.Int("pdf-max-pages", 0, "max pdf pages to extract (0=all)")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	rawURL := strings.TrimSpace(fs.Arg(0))
	if rawURL == "" {
		fmt.Fprintln(stderr, "[ERROR] url is required")
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	outArg := strings.TrimSpace(*out)
	saveFiles := outArg != ""

	if *maxPages <= 1 && *depth <= 0 && !*followPagination {
		outDir := ""
		if saveFiles {
			outDir = outArg
			if strings.Contains(filepath.Base(outDir), ".") {
				outDir = filepath.Dir(outDir)
			}
		}
		r, err := ProcessOne(ctx, rawURL, &ProcessOptions{
			OutDir:      outDir,
			JSFallback:  *js,
			JSTimeout:   *jsTimeout,
			PDFMaxPages: *pdfMaxPages,
		})
		if err != nil {
			fmt.Fprintln(stderr, "[ERROR]", err)
			return 1
		}
		if saveFiles {
			fmt.Fprintln(stdout, "[OK] fetched:", r.FinalURL)
			fmt.Fprintln(stdout, "[OK] saved :", filepath.Base(r.SavedBase))
			fmt.Fprintln(stdout)
		}
		fmt.Fprintln(stdout, "URL:", r.FinalURL)
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, r.Text)
		return 0
	}

	if !saveFiles {
		fmt.Fprintln(stderr, "[ERROR] crawl mode requires --out to save pages")
		return 1
	}

	outDir := outArg
	if strings.Contains(filepath.Base(outDir), ".") {
		outDir = filepath.Dir(outDir)
	}
	results, err := Crawl(ctx, rawURL, &CrawlOptions{
		OutDir:           outDir,
		MaxPages:         *maxPages,
		Depth:            *depth,
		Delay:            *delay,
		SameDomainOnly:   *sameDomainOnly,
		IncludePattern:   *include,
		ExcludePattern:   *exclude,
		FollowPagination: *followPagination,
		JSFallback:       *js,
		JSTimeout:        *jsTimeout,
		PDFMaxPages:      *pdfMaxPages,
	})
	if err != nil {
		fmt.Fprintln(stderr, "[ERROR]", err)
		return 1
	}
	fmt.Fprintf(stdout, "[OK] fetched %d pages\n", len(results))
	fmt.Fprintln(stdout, "[OK] saved :", outDir)
	return 0
}
