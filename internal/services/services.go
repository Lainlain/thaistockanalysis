package services

import (
	"fmt"
	"html/template"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"thaistockanalysis/internal/models"

	"github.com/gomarkdown/markdown"
)

// Cache for parsed markdown files
var (
	markdownCache = make(map[string]models.StockData)
	cacheMutex    sync.RWMutex
	cacheExpiry   = make(map[string]time.Time)
)

// Template cache for performance
var (
	templateCache = make(map[string]*template.Template)
	templateMutex sync.RWMutex
)

// MardownService handles markdown file parsing and caching
type MarkdownService struct{}

// NewMarkdownService creates a new markdown service
func NewMarkdownService() *MarkdownService {
	return &MarkdownService{}
}

// GetCachedStockData retrieves stock data from cache or parses if not cached
func (ms *MarkdownService) GetCachedStockData(filePath string) (models.StockData, error) {
	cacheMutex.RLock()
	if data, exists := markdownCache[filePath]; exists {
		if expiry, hasExpiry := cacheExpiry[filePath]; hasExpiry {
			if time.Now().Before(expiry) {
				cacheMutex.RUnlock()
				return data, nil
			}
		}
	}
	cacheMutex.RUnlock()

	// Parse and cache
	data, err := ms.ParseMarkdownArticle(filePath)
	if err != nil {
		return data, err
	}

	cacheMutex.Lock()
	markdownCache[filePath] = data
	cacheExpiry[filePath] = time.Now().Add(5 * time.Minute) // Cache for 5 minutes
	cacheMutex.Unlock()

	return data, nil
}

// ParseMarkdownArticle parses a markdown file into structured stock data
func (ms *MarkdownService) ParseMarkdownArticle(filePath string) (models.StockData, error) {
	data := models.StockData{
		CurrentDate:  time.Now().Format("2 January 2006"),
		KeyTakeaways: []string{},
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return data, err
	}

	lines := strings.Split(string(content), "\n")
	currentSection := ""
	currentSubsection := ""
	analysisContent := ""
	summaryContent := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Main sections
		if strings.HasPrefix(line, "## Morning Session") {
			currentSection = "morning"
			currentSubsection = ""
			continue
		} else if strings.HasPrefix(line, "## Afternoon Session") {
			currentSection = "afternoon"
			currentSubsection = ""
			continue
		}

		// Subsections
		if strings.HasPrefix(line, "### Open Set") {
			currentSubsection = "open"
			analysisContent = ""
			continue
		} else if strings.HasPrefix(line, "### Open Analysis") {
			currentSubsection = "open_analysis"
			analysisContent = ""
			continue
		} else if strings.HasPrefix(line, "### Close Set") {
			currentSubsection = "close"
			summaryContent = ""
			continue
		} else if strings.HasPrefix(line, "### Close Summary") {
			currentSubsection = "close_summary"
			summaryContent = ""
			continue
		} else if strings.HasPrefix(line, "## Key Takeaways") {
			currentSection = "takeaways"
			currentSubsection = ""
			continue
		}

		// Parse content based on section and subsection
		switch currentSection {
		case "morning":
			ms.parseMorningSession(line, currentSubsection, &data, &analysisContent, &summaryContent)
		case "afternoon":
			ms.parseAfternoonSession(line, currentSubsection, &data, &analysisContent, &summaryContent)
		case "takeaways":
			if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
				takeaway := strings.TrimSpace(line[1:])
				if takeaway != "" {
					data.KeyTakeaways = append(data.KeyTakeaways, takeaway)
				}
			}
		}
	}

	return data, nil
}

// parseMorningSession handles parsing of morning session data
func (ms *MarkdownService) parseMorningSession(line, subsection string, data *models.StockData, analysisContent, summaryContent *string) {
	switch subsection {
	case "open":
		if strings.HasPrefix(line, "* Open Index:") {
			data.MorningOpenIndex, data.MorningOpenChange = ms.parseIndexLine(line)
		} else if strings.HasPrefix(line, "* Highlights:") {
			data.MorningOpenHighlights = ms.parseHighlights(line)
		}
	case "open_analysis":
		if strings.HasPrefix(line, "<p>") || *analysisContent != "" {
			if *analysisContent != "" {
				*analysisContent += "\n"
			}
			*analysisContent += line
			if strings.HasSuffix(line, "</p>") || (!strings.HasPrefix(line, "<") && line != "") {
				data.MorningOpenAnalysis = template.HTML(markdown.ToHTML([]byte(*analysisContent), nil, nil))
			}
		}
	case "close":
		if strings.HasPrefix(line, "* Close Index:") {
			data.MorningCloseIndex, data.MorningCloseChange = ms.parseIndexLine(line)
		} else if strings.HasPrefix(line, "* Highlights:") {
			data.MorningCloseHighlights = ms.parseHighlights(line)
		}
	case "close_summary":
		if strings.HasPrefix(line, "<p>") || *summaryContent != "" {
			if *summaryContent != "" {
				*summaryContent += "\n"
			}
			*summaryContent += line
			if strings.HasSuffix(line, "</p>") || (!strings.HasPrefix(line, "<") && line != "") {
				data.MorningCloseSummary = template.HTML(markdown.ToHTML([]byte(*summaryContent), nil, nil))
			}
		}
	}
}

// parseAfternoonSession handles parsing of afternoon session data
func (ms *MarkdownService) parseAfternoonSession(line, subsection string, data *models.StockData, analysisContent, summaryContent *string) {
	switch subsection {
	case "open":
		if strings.HasPrefix(line, "* Open Index:") {
			data.AfternoonOpenIndex, data.AfternoonOpenChange = ms.parseIndexLine(line)
		} else if strings.HasPrefix(line, "* Highlights:") {
			data.AfternoonOpenHighlights = ms.parseHighlights(line)
		}
	case "open_analysis":
		if strings.HasPrefix(line, "<p>") || *analysisContent != "" {
			if *analysisContent != "" {
				*analysisContent += "\n"
			}
			*analysisContent += line
			if strings.HasSuffix(line, "</p>") || (!strings.HasPrefix(line, "<") && line != "") {
				data.AfternoonOpenAnalysis = template.HTML(markdown.ToHTML([]byte(*analysisContent), nil, nil))
			}
		}
	case "close":
		if strings.HasPrefix(line, "* Close Index:") {
			data.AfternoonCloseIndex, data.AfternoonCloseChange = ms.parseIndexLine(line)
		} else if strings.HasPrefix(line, "* Highlights:") {
			data.AfternoonCloseHighlights = ms.parseHighlights(line)
		}
	case "close_summary":
		if strings.HasPrefix(line, "<p>") || *summaryContent != "" {
			if *summaryContent != "" {
				*summaryContent += "\n"
			}
			*summaryContent += line
			if strings.HasSuffix(line, "</p>") || (!strings.HasPrefix(line, "<") && line != "") {
				data.AfternoonCloseSummary = template.HTML(markdown.ToHTML([]byte(*summaryContent), nil, nil))
			}
		}
	}
}

// parseIndexLine extracts index value and change from a line
func (ms *MarkdownService) parseIndexLine(line string) (float64, float64) {
	// Parse "* Open Index: 1270.96 (4.85)" or "* Close Index: 1275.40 (9.29)"
	re := regexp.MustCompile(`(\d+\.?\d*)\s*\(([+-]?\d+\.?\d*)\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		index, _ := strconv.ParseFloat(matches[1], 64)
		change, _ := strconv.ParseFloat(matches[2], 64)
		return index, change
	}
	return 0, 0
}

// parseHighlights extracts highlights from a line
func (ms *MarkdownService) parseHighlights(line string) string {
	// Remove "* Highlights: " prefix
	if strings.HasPrefix(line, "* Highlights: ") {
		return strings.TrimSpace(line[14:])
	}
	return line
}

// ClearCache clears the markdown cache for a specific file
func (ms *MarkdownService) ClearCache(filePath string) {
	cacheMutex.Lock()
	delete(markdownCache, filePath)
	delete(cacheExpiry, filePath)
	cacheMutex.Unlock()
}

// TemplateService handles template caching and rendering
type TemplateService struct{}

// NewTemplateService creates a new template service
func NewTemplateService() *TemplateService {
	return &TemplateService{}
}

// GetTemplate retrieves a cached template or loads and caches it
func (ts *TemplateService) GetTemplate(name string, files ...string) (*template.Template, error) {
	templateMutex.RLock()
	if tmpl, exists := templateCache[name]; exists {
		templateMutex.RUnlock()
		return tmpl, nil
	}
	templateMutex.RUnlock()

	// Create template with custom functions
	funcMap := template.FuncMap{
		"printf":         fmt.Sprintf,
		"html":           func(s string) template.HTML { return template.HTML(s) },
		"add":            func(a, b int) int { return a + b },
		"markdownToHTML": func(s string) template.HTML { return template.HTML(markdown.ToHTML([]byte(s), nil, nil)) },
	}

	// Parse templates
	tmpl := template.New("base.gohtml").Funcs(funcMap)
	tmpl, err := tmpl.ParseFiles(files...)
	if err != nil {
		return nil, err
	}

	// Cache the template
	templateMutex.Lock()
	templateCache[name] = tmpl
	templateMutex.Unlock()

	return tmpl, nil
}

// ClearTemplateCache clears all cached templates
func (ts *TemplateService) ClearTemplateCache() {
	templateMutex.Lock()
	defer templateMutex.Unlock()
	templateCache = make(map[string]*template.Template)
}
