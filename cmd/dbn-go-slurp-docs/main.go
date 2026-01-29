// Copyright (c) 2025 Neomantra Corp
//
// dbn-go-slurp-docs reads the Databento documentation sitemap,
// scrapes each page, and generates coding-agent-focused summaries.
//

package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/chromedp/chromedp"
)

///////////////////////////////////////////////////////////////////////////////

const (
	baseURL       = "https://databento.com/docs"
	defaultDelay  = 3 * time.Second
	maxRetries    = 3
	pageTimeout   = 60 * time.Second
	progressFile  = ".slurp-progress.json"
)

// Progress tracks which URLs have been successfully fetched
type Progress struct {
	Completed map[string]bool   `json:"completed"`
	Failed    map[string]string `json:"failed"` // url -> error message
	mu        sync.RWMutex
}

func NewProgress() *Progress {
	return &Progress{
		Completed: make(map[string]bool),
		Failed:    make(map[string]string),
	}
}

func (p *Progress) IsCompleted(url string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Completed[url]
}

func (p *Progress) MarkCompleted(url string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Completed[url] = true
	delete(p.Failed, url)
}

func (p *Progress) MarkFailed(url string, errMsg string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Failed[url] = errMsg
}

func (p *Progress) Stats() (completed, failed int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.Completed), len(p.Failed)
}

func (p *Progress) Save(outputDir string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(outputDir, progressFile)
	return os.WriteFile(path, data, 0644)
}

func (p *Progress) Load(outputDir string) error {
	path := filepath.Join(outputDir, progressFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No progress file yet
		}
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	return json.Unmarshal(data, p)
}

type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []URL    `xml:"url"`
}

type URL struct {
	Loc string `xml:"loc"`
}

type DocPage struct {
	URL        string
	Path       string
	Section    string
	SubSection string
	Title      string
	Content    string
	CodeBlocks []CodeBlock
	Summary    string
}

// CodeBlock represents a code example with language detection
type CodeBlock struct {
	Code     string `json:"code"`
	Language string `json:"language"`
	Context  string `json:"context,omitempty"` // e.g., "API Request", "Response", etc.
}

type ExtractedData struct {
	Content  string      `json:"content"`
	Codes    []CodeBlock `json:"codes"`
	Headings []string    `json:"headings"`
}

///////////////////////////////////////////////////////////////////////////////

func main() {
	var (
		outputDir   = flag.String("output", "docs-corpus", "Output directory for documentation corpus")
		shouldFetch = flag.Bool("fetch", false, "Fetch page content (requires Chrome)")
		genSitemap  = flag.Bool("generate-sitemap", false, "Generate sitemap from Databento docs (requires Chrome)")
		delay       = flag.Duration("delay", defaultDelay, "Base delay between requests")
		verbose     = flag.Bool("v", false, "Verbose logging")
		noResume    = flag.Bool("no-resume", false, "Don't resume from previous run")
	)
	flag.Parse()

	// Allow positional argument for output directory
	if len(flag.Args()) >= 1 {
		*outputDir = flag.Arg(0)
	}

	// Sitemap is always stored in the output directory
	sitemapPath := filepath.Join(*outputDir, "databento-docs-sitemap.xml.gz")

	if !*verbose {
		log.SetFlags(log.Ltime)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Generate sitemap if requested
	if *genSitemap {
		log.Println("Generating sitemap from Databento documentation...")
		if err := generateSitemap(context.Background(), sitemapPath); err != nil {
			log.Fatalf("Error generating sitemap: %v", err)
		}
		log.Printf("Sitemap saved to: %s", sitemapPath)
		if !*shouldFetch {
			return
		}
	}

	// Check if sitemap exists
	if _, err := os.Stat(sitemapPath); os.IsNotExist(err) {
		log.Printf("Sitemap not found at %s", sitemapPath)
		log.Println("Run with -generate-sitemap to create it, or place a sitemap file at the above path")
		os.Exit(1)
	}

	urls, err := readSitemap(sitemapPath)
	if err != nil {
		log.Fatalf("Error reading sitemap from %s: %v", sitemapPath, err)
	}

	log.Printf("Found %d URLs in sitemap", len(urls))

	pages := parsePages(urls)
	log.Printf("Parsed %d pages into sections", len(pages))

	if err := createDirectoryStructure(pages, *outputDir); err != nil {
		log.Fatalf("Error creating directory structure: %v", err)
	}

	if err := generateSectionIndices(pages, *outputDir); err != nil {
		log.Fatalf("Error generating section indices: %v", err)
	}

	if err := generateRootIndex(pages, *outputDir); err != nil {
		log.Fatalf("Error generating root index: %v", err)
	}

	if *shouldFetch {
		progress := NewProgress()

		if !*noResume {
			if err := progress.Load(*outputDir); err != nil {
				log.Printf("Warning: could not load progress: %v", err)
			} else {
				completed, failed := progress.Stats()
				log.Printf("Resuming: %d completed, %d failed from previous run", completed, failed)
			}
		}

		// Set up signal handling for graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			log.Println("\nReceived interrupt signal, shutting down gracefully...")
			cancel()
		}()

		log.Printf("Fetching page content...")
		if err := fetchAndSummarize(ctx, pages, *outputDir, *delay, progress); err != nil {
			if err == context.Canceled {
				log.Println("Fetch canceled by user")
			} else {
				log.Printf("Error fetching content: %v", err)
			}
		}

		// Save progress before exiting
		if err := progress.Save(*outputDir); err != nil {
			log.Printf("Warning: could not save progress: %v", err)
		}

		completed, failed := progress.Stats()
		log.Printf("Final status: %d completed, %d failed", completed, failed)

		cancel()
	} else {
		log.Println("Directory structure created. Run with --fetch to scrape content.")
		log.Printf("Output directory: %s/\n", *outputDir)
	}
}

// generateSitemap crawls the Databento docs to create a sitemap
func generateSitemap(ctx context.Context, outputPath string) error {
	chromePath := findChrome()
	if chromePath == "" {
		return fmt.Errorf("chrome not found - please install Chrome or set CHROME_PATH")
	}

	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.0"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	log.Println("Navigating to Databento documentation...")

	// Navigate to the docs homepage
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(baseURL),
		chromedp.WaitReady("body"),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		return fmt.Errorf("navigation failed: %w", err)
	}

	// Extract all links from the navigation
	var result struct {
		URLs []string `json:"urls"`
	}

	log.Println("Extracting documentation URLs...")

	err = chromedp.Run(browserCtx,
		chromedp.Evaluate(`
			(function() {
				const urls = new Set();
				
				// Add the base URL
				urls.add('https://databento.com/docs');
				
				// Find all links in the navigation/sidebar
				const links = document.querySelectorAll('a[href^="/docs/"], a[href^="https://databento.com/docs/"]');
				links.forEach(link => {
					let href = link.getAttribute('href');
					if (href) {
						// Normalize to full URL
						if (href.startsWith('/')) {
							href = 'https://databento.com' + href;
						}
						// Remove hash and query params
						href = href.split('#')[0].split('?')[0];
						// Only include docs URLs
						if (href.includes('/docs/')) {
							urls.add(href);
						}
					}
				});
				
				return { urls: Array.from(urls).sort() };
			})()
		`, &result),
	)
	if err != nil {
		return fmt.Errorf("failed to extract URLs: %w", err)
	}

	if len(result.URLs) == 0 {
		return fmt.Errorf("no URLs found in documentation")
	}

	log.Printf("Found %d unique URLs", len(result.URLs))

	// Create the sitemap
	urlset := URLSet{
		URLs: make([]URL, 0, len(result.URLs)),
	}
	for _, u := range result.URLs {
		urlset.URLs = append(urlset.URLs, URL{Loc: u})
	}

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(urlset, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sitemap: %w", err)
	}

	// Add XML header
	xmlData = append([]byte(xml.Header), xmlData...)

	// Compress with gzip
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create sitemap file: %w", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	if _, err := gzWriter.Write(xmlData); err != nil {
		return fmt.Errorf("failed to write sitemap: %w", err)
	}

	return nil
}

func readSitemap(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening sitemap: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file

	if strings.HasSuffix(path, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("creating gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	var urlset URLSet
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(&urlset); err != nil {
		return nil, fmt.Errorf("decoding XML: %w", err)
	}

	urls := make([]string, 0, len(urlset.URLs))
	for _, u := range urlset.URLs {
		urls = append(urls, u.Loc)
	}

	return urls, nil
}

func parsePages(urls []string) []DocPage {
	pages := make([]DocPage, 0, len(urls))

	for _, url := range urls {
		path := strings.TrimPrefix(url, baseURL)
		path = strings.Trim(path, "/")

		if path == "" {
			pages = append(pages, DocPage{
				URL:     url,
				Path:    "_root",
				Section: "_root",
				Title:   "Databento Documentation",
			})
			continue
		}

		parts := strings.Split(path, "/")
		section := parts[0]

		subSection := ""
		if len(parts) > 1 {
			subSection = parts[1]
		}

		title := formatPageTitle(path)

		pages = append(pages, DocPage{
			URL:        url,
			Path:       path,
			Section:    section,
			SubSection: subSection,
			Title:      title,
		})
	}

	return pages
}

func createDirectoryStructure(pages []DocPage, outputDir string) error {
	sections := make(map[string]bool)
	for _, p := range pages {
		if p.Section != "_root" {
			sections[p.Section] = true
		}
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	for section := range sections {
		sectionDir := filepath.Join(outputDir, sanitizeFilename(section))
		if err := os.MkdirAll(sectionDir, 0755); err != nil {
			return fmt.Errorf("creating section dir %s: %w", sectionDir, err)
		}
	}

	return nil
}

func generateSectionIndices(pages []DocPage, outputDir string) error {
	sections := make(map[string][]DocPage)
	for _, p := range pages {
		if p.Section != "_root" {
			sections[p.Section] = append(sections[p.Section], p)
		}
	}

	for section, sectionPages := range sections {
		filename := filepath.Join(outputDir, sanitizeFilename(section), "index.md")
		content := generateSectionIndexMarkdown(section, sectionPages)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing section index %s: %w", filename, err)
		}
	}

	return nil
}

func generateRootIndex(pages []DocPage, outputDir string) error {
	sections := make(map[string][]DocPage)
	sectionOrder := []string{}

	for _, p := range pages {
		if p.Section == "_root" {
			continue
		}
		if _, exists := sections[p.Section]; !exists {
			sectionOrder = append(sectionOrder, p.Section)
		}
		sections[p.Section] = append(sections[p.Section], p)
	}

	var b strings.Builder

	b.WriteString("# Databento Documentation Corpus\n\n")
	b.WriteString("This corpus contains coding-agent-focused summaries of the Databento documentation.\n\n")

	totalPages := 0
	for _, pages := range sections {
		totalPages += len(pages)
	}
	b.WriteString(fmt.Sprintf("**Total Sections:** %d\n", len(sections)))
	b.WriteString(fmt.Sprintf("**Total Pages:** %d\n\n", totalPages))

	b.WriteString("## Key Sections for Developers\n\n")
	keySections := []string{
		"schemas-and-data-formats",
		"standards-and-conventions",
		"api-reference-historical",
		"api-reference-live",
		"examples",
		"venues-and-datasets",
	}

	for _, key := range keySections {
		if pages, ok := sections[key]; ok {
			sectionName := formatTitle(key)
			b.WriteString(fmt.Sprintf("- [%s](%s/index.md) - %d pages\n", sectionName, sanitizeFilename(key), len(pages)))
		}
	}

	b.WriteString("\n## All Sections\n\n")
	for _, section := range sectionOrder {
		pages := sections[section]
		sectionName := formatTitle(section)
		safeName := sanitizeFilename(section)
		b.WriteString(fmt.Sprintf("- [%s](%s/index.md) - %d pages\n", sectionName, safeName, len(pages)))
	}

	b.WriteString("\n## Using This Corpus\n\n")
	b.WriteString("Each page summary includes:\n\n")
	b.WriteString("- **Overview**: Brief description of the topic\n")
	b.WriteString("- **Key Concepts**: Important terminology and concepts\n")
	b.WriteString("- **Code Examples**: Relevant code snippets organized by language\n")
	b.WriteString("  - Python, Rust, C++, C examples\n")
	b.WriteString("  - Shell/HTTP examples\n")
	b.WriteString("  - JSON responses\n")
	b.WriteString("- **API Reference**: Function/method signatures\n")
	b.WriteString("- **Data Structures**: Record types and fields\n")
	b.WriteString("- **Common Patterns**: Usage patterns and best practices\n")
	b.WriteString("- **Related Topics**: Links to related documentation\n\n")

	filename := filepath.Join(outputDir, "README.md")
	return os.WriteFile(filename, []byte(b.String()), 0644)
}

func fetchAndSummarize(ctx context.Context, pages []DocPage, outputDir string, baseDelay time.Duration, progress *Progress) error {
	var workItems []DocPage
	for _, p := range pages {
		if p.Section != "_root" && !progress.IsCompleted(p.URL) {
			workItems = append(workItems, p)
		}
	}

	completed, failed := progress.Stats()
	log.Printf("Starting fetch: %d pending, %d already completed, %d previously failed",
		len(workItems), completed, failed)

	if len(workItems) == 0 {
		log.Println("All pages already fetched!")
		return nil
	}

	successCount := 0
	failCount := 0
	currentDelay := baseDelay

	// Progress save ticker - save every 10 seconds
	saveTicker := time.NewTicker(10 * time.Second)
	defer saveTicker.Stop()

	go func() {
		for range saveTicker.C {
			if err := progress.Save(outputDir); err != nil {
				log.Printf("Warning: could not save progress: %v", err)
			}
		}
	}()

	for i, page := range workItems {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		log.Printf("[%d/%d] Processing: %s", i+1, len(workItems), page.URL)

		var success bool
		var lastErr error

		for attempt := 0; attempt <= maxRetries; attempt++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if attempt > 0 {
				waitTime := currentDelay * time.Duration(attempt)
				log.Printf("  Retry %d/%d after %v...", attempt, maxRetries, waitTime)

				// Use a timer that respects context cancellation
				timer := time.NewTimer(waitTime)
				select {
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				case <-timer.C:
				}
			}

			// Create fresh browser context for each attempt
			data, err := fetchWithNewBrowser(ctx, page.URL)
			if err != nil {
				log.Printf("  Error (attempt %d): %v", attempt+1, err)
				lastErr = err
				currentDelay = min(currentDelay*2, 30*time.Second)
				continue
			}

			// Success
			summary := generateCodingSummary(page, data)

			sectionDir := sanitizeFilename(page.Section)
			filename := sanitizeFilename(filepath.Base(page.Path)) + ".md"
			if filename == ".md" || filename == "index.md" {
				filename = "overview.md"
			}
			filepath := filepath.Join(outputDir, sectionDir, filename)

			if err := os.WriteFile(filepath, []byte(summary), 0644); err != nil {
				log.Printf("  Error writing %s: %v", filepath, err)
				lastErr = err
				continue
			}

			log.Printf("  Saved: %s (%d bytes)", filepath, len(summary))
			progress.MarkCompleted(page.URL)
			success = true
			successCount++
			currentDelay = baseDelay
			break
		}

		if !success {
			log.Printf("  FAILED after %d attempts: %v", maxRetries+1, lastErr)
			progress.MarkFailed(page.URL, lastErr.Error())
			failCount++
		}

		if i < len(workItems)-1 {
			// Sleep between requests with context cancellation support
			timer := time.NewTimer(currentDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}

	log.Printf("Completed: %d succeeded, %d failed", successCount, failCount)
	return nil
}

// fetchWithNewBrowser creates a fresh browser instance and extracts all content
func fetchWithNewBrowser(ctx context.Context, pageURL string) (*ExtractedData, error) {
	// Find Chrome executable
	chromePath := findChrome()
	if chromePath == "" {
		return nil, fmt.Errorf("chrome not found - please install Chrome or set CHROME_PATH")
	}

	// Create new browser context with options to make it more stable
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.0"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	// Set a shorter timeout for navigation
	navCtx, navCancel := context.WithTimeout(browserCtx, 30*time.Second)
	defer navCancel()

	// Navigate and wait for load
	err := chromedp.Run(navCtx,
		chromedp.Navigate(pageURL),
		chromedp.WaitReady("body"),
	)
	if err != nil {
		return nil, fmt.Errorf("navigation failed: %w", err)
	}

	// Wait for page to render
	chromedp.Sleep(2 * time.Second).Do(browserCtx)

	var content string
	var headings []string
	allCodes := make(map[string][]CodeBlock)

	// First, get content and headings
	extractCtx, extractCancel := context.WithTimeout(browserCtx, 15*time.Second)
	defer extractCancel()

	var baseData ExtractedData
	err = chromedp.Run(extractCtx,
		chromedp.Evaluate(`
			(function() {
				const selectors = [
					'[data-testid="docs-content"]',
					'article[class*="content"]',
					'main article',
					'[class*="documentation"]',
					'[class*="docs-content"]',
					'article',
					'main'
				];
				
				let contentEl = null;
				for (const sel of selectors) {
					const el = document.querySelector(sel);
					if (el && el.innerText.length > 200) {
						contentEl = el;
						break;
					}
				}
				
				const headings = [];
				document.querySelectorAll('h1, h2, h3').forEach(h => {
					const text = h.innerText.trim();
					if (text && text.length < 100 && !text.includes('DOCS')) {
						headings.push(text);
					}
				});
				
				let text = contentEl ? contentEl.innerText : document.body.innerText;
				
				return {
					content: text,
					headings: headings.slice(0, 15)
				};
			})()
		`, &baseData),
	)
	if err != nil {
		return nil, fmt.Errorf("base extraction failed: %w", err)
	}

	content = baseData.Content
	headings = baseData.Headings

	// Now extract code for each language by clicking through dropdowns
	languages := []struct {
		name    string
		display string
	}{
		{"python", "Python"},
		{"rust", "Rust"},
		{"cpp", "C++"},
		{"c", "C"},
		{"http", "HTTP"},
	}

	for _, lang := range languages {
		// Click dropdown for this language and extract code
		var codes []CodeBlock

		langCtx, langCancel := context.WithTimeout(browserCtx, 10*time.Second)

		// First click the dropdown to open it
		err = chromedp.Run(langCtx,
			chromedp.Evaluate(`
				(function() {
					// Find a Python dropdown button and click it
					const buttons = document.querySelectorAll('button');
					for (const btn of buttons) {
						if (btn.textContent.trim() === 'Python' && btn.className.includes('dropdown-toggle')) {
							btn.click();
							return true;
						}
					}
					return false;
				})()
			`, nil),
		)

		if err == nil {
			// Wait for dropdown to open
			chromedp.Sleep(300 * time.Millisecond).Do(browserCtx)

			// Now click the specific language option
			err = chromedp.Run(langCtx,
				chromedp.Evaluate(fmt.Sprintf(`
					(function() {
						const items = document.querySelectorAll('.dropdown-item');
						for (const item of items) {
							if (item.textContent.trim() === '%s') {
								item.click();
								return true;
							}
						}
						return false;
					})()
				`, lang.display), nil),
			)

			if err == nil {
				// Wait for content to update
				chromedp.Sleep(500 * time.Millisecond).Do(browserCtx)

				// Extract code blocks
				var langData struct {
					Codes []CodeBlock `json:"codes"`
				}

				extractCtx, extractCancel := context.WithTimeout(browserCtx, 10*time.Second)
				err = chromedp.Run(extractCtx,
					chromedp.Evaluate(fmt.Sprintf(`
						(function() {
							const codes = [];
							document.querySelectorAll('pre code').forEach(el => {
								const code = el.textContent.trim();
								if (code.length < 20 || code.length > 5000) return;
								
								let language = '%s';
								
								// Check class for language
								const className = el.className || '';
								if (className.includes('language-python')) language = 'python';
								else if (className.includes('language-rust')) language = 'rust';
								else if (className.includes('language-cpp') || className.includes('language-c++')) language = 'cpp';
								else if (className.includes('language-c') && !className.includes('cpp') && !className.includes('c++')) language = 'c';
								
								// Only include if it matches our target language
								if (language === '%s' || '%s' === 'http') {
									codes.push({
										code: code.substring(0, 3000),
										language: language,
										context: ''
									});
								}
							});
							
							return { codes: codes.slice(0, 8) };
						})()
					`, lang.name, lang.name, lang.name), &langData),
				)
				extractCancel()

				if err == nil {
					codes = langData.Codes
				}
			}
		}

		langCancel()
		allCodes[lang.name] = codes
	}

	// Flatten codes for output
	var flatCodes []CodeBlock
	langOrder := []string{"python", "rust", "cpp", "c", "http"}
	for _, lang := range langOrder {
		if codes, ok := allCodes[lang]; ok {
			flatCodes = append(flatCodes, codes...)
		}
	}

	// If no codes extracted via dropdowns, fallback to extracting visible code
	if len(flatCodes) == 0 {
		fallbackCtx, fallbackCancel := context.WithTimeout(browserCtx, 10*time.Second)
		var fallbackData struct {
			Codes []CodeBlock `json:"codes"`
		}

		err = chromedp.Run(fallbackCtx,
			chromedp.Evaluate(`
				(function() {
					const codes = [];
					document.querySelectorAll('pre code').forEach(el => {
						const code = el.textContent.trim();
						if (code.length < 20 || code.length > 5000) return;
						
						let language = '';
						const className = el.className || '';
						
						if (className.includes('python')) language = 'python';
						else if (className.includes('rust')) language = 'rust';
						else if (className.includes('cpp') || className.includes('c++')) language = 'cpp';
						else if (className.includes(' language-c')) language = 'c';
						else if (className.includes('json')) language = 'json';
						else if (className.includes('bash') || className.includes('shell')) language = 'shell';
						
						// Fallback detection
						if (!language) {
							if (code.includes('def ') || code.includes('import databento')) language = 'python';
							else if (code.includes('fn ') || code.includes('use databento')) language = 'rust';
						}
						
						codes.push({
							code: code.substring(0, 3000),
							language: language || 'text',
							context: ''
						});
					});
					
					return { codes: codes.slice(0, 15) };
				})()
			`, &fallbackData),
		)
		fallbackCancel()

		if err == nil {
			flatCodes = fallbackData.Codes
		}
	}

	content = cleanContent(content)

	if len(content) < 100 {
		return nil, fmt.Errorf("content too short (%d chars)", len(content))
	}

	return &ExtractedData{
		Content:  content,
		Codes:    flatCodes,
		Headings: headings,
	}, nil
}

func cleanContent(text string) string {
	lines := strings.Split(text, "\n")
	var cleanLines []string

	skipPatterns := []string{
		"DOCS", "Home", "Quickstart", "Examples and tutorials",
		"API REFERENCE", "CORE CONCEPTS", "RESOURCES",
		"Sign in", "Sign up", "Search", "Menu",
		"Previous", "Next", "Was this page helpful?",
		"Edit this page", "Scroll to top", "Table of contents",
		"On this page", "Documentation", "Standards and conventions",
		"Schemas and data formats", "Venues and datasets",
		"Architecture", "FAQs", "Portal", "Release notes",
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		skip := false
		for _, pattern := range skipPatterns {
			if line == pattern || strings.HasPrefix(line, pattern+" ") {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		if len(line) < 3 {
			continue
		}
		if len(line) < 20 && strings.ToUpper(line) == line {
			continue
		}

		cleanLines = append(cleanLines, line)
	}

	return strings.Join(cleanLines, "\n")
}

func generateCodingSummary(page DocPage, data *ExtractedData) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s\n\n", page.Title))
	b.WriteString(fmt.Sprintf("**Source:** [%s](%s)\n\n", page.URL, page.URL))

	if page.SubSection != "" {
		b.WriteString(fmt.Sprintf("**Section:** %s > %s\n\n", formatTitle(page.Section), formatTitle(page.SubSection)))
	} else {
		b.WriteString(fmt.Sprintf("**Section:** %s\n\n", formatTitle(page.Section)))
	}

	if len(data.Headings) > 0 {
		b.WriteString("## Contents\n\n")
		for _, h := range data.Headings {
			if h != page.Title && len(h) < 100 {
				b.WriteString(fmt.Sprintf("- %s\n", h))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("## Overview\n\n")
	b.WriteString(truncateText(data.Content, 3000))
	b.WriteString("\n\n")

	// Organize code examples by language
	if len(data.Codes) > 0 {
		codeByLang := make(map[string][]CodeBlock)
		var langOrder []string

		for _, code := range data.Codes {
			lang := code.Language
			if lang == "" {
				lang = "other"
			}
			if _, exists := codeByLang[lang]; !exists {
				langOrder = append(langOrder, lang)
			}
			codeByLang[lang] = append(codeByLang[lang], code)
		}

		// Prioritize languages
		priority := []string{"python", "rust", "cpp", "c", "go", "json", "shell", "http", "raw", "other"}
		var orderedLangs []string
		for _, lang := range priority {
			if _, exists := codeByLang[lang]; exists {
				orderedLangs = append(orderedLangs, lang)
			}
		}
		// Add any remaining languages not in priority list
		for _, lang := range langOrder {
			found := false
			for _, l := range orderedLangs {
				if l == lang {
					found = true
					break
				}
			}
			if !found {
				orderedLangs = append(orderedLangs, lang)
			}
		}

		// Output organized code examples
		for _, lang := range orderedLangs {
			codes := codeByLang[lang]
			langDisplay := formatLanguageName(lang)

			b.WriteString(fmt.Sprintf("## %s Examples\n\n", langDisplay))

			for i, code := range codes {
				if i >= 4 { // Limit to 4 examples per language
					b.WriteString(fmt.Sprintf("*... and %d more %s examples ...*\n\n", len(codes)-4, langDisplay))
					break
				}

				b.WriteString(fmt.Sprintf("### Example %d\n\n", i+1))
				if code.Context != "" {
					b.WriteString(fmt.Sprintf("*%s*\n\n", code.Context))
				}
				b.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n", lang, truncateText(code.Code, 2000)))
			}
		}
	}

	b.WriteString("## Key Concepts\n\n")
	if len(data.Headings) > 0 {
		for _, h := range data.Headings[1:] {
			if h != page.Title && len(h) < 100 {
				b.WriteString(fmt.Sprintf("- %s\n", h))
			}
		}
	}
	b.WriteString("\n")

	b.WriteString("## Developer Notes\n\n")
	b.WriteString("### Common Usage Patterns\n\n")
	b.WriteString("- Review the code examples above for implementation patterns\n")
	b.WriteString("- Check related pages in this section for complete context\n")
	b.WriteString(fmt.Sprintf("- Official documentation: %s\n", page.URL))
	b.WriteString("\n")

	b.WriteString("### Integration Points\n\n")
	b.WriteString("- Databento provides official clients in Python, Rust, C++, and C\n")
	b.WriteString("- DBN format is language-agnostic with open-source libraries\n")
	b.WriteString("- All schemas and record types are standardized across languages\n")
	b.WriteString("\n")

	return b.String()
}

func formatLanguageName(lang string) string {
	names := map[string]string{
		"python": "Python",
		"rust":   "Rust",
		"cpp":    "C++",
		"c":      "C",
		"go":     "Go",
		"json":   "JSON",
		"shell":  "Shell",
		"http":   "HTTP/Raw",
		"raw":    "Raw",
		"other":  "Other",
	}
	if name, ok := names[lang]; ok {
		return name
	}
	return strings.ToUpper(lang[:1]) + lang[1:]
}

func generateSectionIndexMarkdown(section string, pages []DocPage) string {
	var b strings.Builder

	title := formatTitle(section)
	b.WriteString(fmt.Sprintf("# %s\n\n", title))
	b.WriteString(fmt.Sprintf("**Total Pages:** %d\n\n", len(pages)))

	subSections := make(map[string][]DocPage)
	for _, p := range pages {
		sub := p.SubSection
		if sub == "" {
			sub = "(General)"
		}
		subSections[sub] = append(subSections[sub], p)
	}

	for sub, subPages := range subSections {
		if sub != "(General)" {
			b.WriteString(fmt.Sprintf("## %s\n\n", formatTitle(sub)))
		}

		for _, p := range subPages {
			filename := sanitizeFilename(filepath.Base(p.Path)) + ".md"
			if filename == ".md" || filename == "index.md" {
				filename = "overview.md"
			}
			b.WriteString(fmt.Sprintf("- [%s](%s) - [Official Docs](%s)\n", p.Title, filename, p.URL))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ToLower(name)
	return name
}

func formatTitle(name string) string {
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")

	replacements := map[string]string{
		"api": "API", "dbn": "DBN", "mbo": "MBO", "mbp": "MBP",
		"tbbo": "TBBO", "bbo": "BBO", "ohlcv": "OHLCV", "eod": "EOD",
		"oi": "OI", "csv": "CSV", "json": "JSON", "http": "HTTP",
		"url": "URL", "faq": "FAQ", "faqs": "FAQs", "zte": "ZTE", "dte": "DTE",
		"nasdaq": "Nasdaq", "nyse": "NYSE", "cme": "CME", "ice": "ICE",
		"iex": "IEX", "miax": "MIAX", "memx": "MEMX", "opra": "OPRA",
		"eobi": "EOBI", "itch": "ITCH", "glbx": "GLBX", "mdp3": "MDP3",
		"xnas": "XNAS", "xnys": "XNYS", "xpsx": "XPSX", "xbos": "XBOS",
		"xchi": "XCHI", "xase": "XASE", "xcis": "XCIS", "xeur": "XEUR",
		"xeee": "XEEE", "ifus": "IFUS", "ifll": "IFLL", "ifeu": "IFEU",
		"ndex": "NDEX", "equs": "EQUS", "eprl": "EPRL",
	}

	words := strings.Fields(name)
	for i, word := range words {
		lower := strings.ToLower(word)
		if replacement, ok := replacements[lower]; ok {
			words[i] = replacement
		} else if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}

	return strings.Join(words, " ")
}

func formatPageTitle(url string) string {
	trimmed := strings.Trim(url, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 {
		return "Home"
	}
	return formatTitle(parts[len(parts)-1])
}

func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "\n\n... [content truncated]"
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// findChrome returns the path to the Chrome executable
func findChrome() string {
	// Check environment variable first
	if path := os.Getenv("CHROME_PATH"); path != "" {
		if _, err := exec.LookPath(path); err == nil {
			return path
		}
	}

	// Platform-specific paths
	switch runtime.GOOS {
	case "darwin":
		candidates := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chrome.app/Contents/MacOS/Chrome",
		}
		for _, path := range candidates {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	case "linux":
		candidates := []string{
			"google-chrome",
			"google-chrome-stable",
			"chromium",
			"chromium-browser",
		}
		for _, name := range candidates {
			if path, err := exec.LookPath(name); err == nil {
				return path
			}
		}
	}

	// Fall back to letting chromedp find it
	return ""
}
