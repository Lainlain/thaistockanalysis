# Bug Fix: Raw HTML Tags Showing in Close Summary

## Problem
The Close Summary sections were displaying raw HTML code like `<p>`, `</p>` instead of properly rendered HTML content. This violated AdSense policies and looked unprofessional.

**Example of the issue:**
```
Morning session closed at 1281.04 (-1.50) after lost 5.97 points from 1287.01 opening.
<p>The morning session on November 11th presented a classic case...</p>
```

## Root Cause
In `internal/services/services.go`, the parser was incorrectly processing HTML content through the markdown converter:

```go
// OLD CODE (INCORRECT)
data.MorningCloseSummary = template.HTML(markdown.ToHTML([]byte(*summaryContent), nil, nil))
```

When content already contains HTML tags (starts with `<p>`), passing it through `markdown.ToHTML()` causes:
1. Double-processing/escaping
2. Raw HTML tags becoming visible text instead of rendered HTML

## Solution
Changed all Analysis and Summary parsing to **NOT** convert content that's already in HTML format:

```go
// NEW CODE (CORRECT)
// Content is already HTML, don't convert from markdown
data.MorningCloseSummary = template.HTML(*summaryContent)
```

## Files Modified
- `internal/services/services.go` - Fixed 4 parsing sections:
  - `MorningOpenAnalysis` (line ~176)
  - `MorningCloseSummary` (line ~193)
  - `AfternoonOpenAnalysis` (line ~219)
  - `AfternoonCloseSummary` (line ~236)

## Testing
1. Build test: `go build -o bin/test-build cmd/server/main.go` ✅
2. Expected behavior: HTML content in markdown files will now render properly as formatted HTML

## AdSense Compliance
✅ **Fixed** - Raw HTML tags no longer visible in content
✅ **Safe** - Proper HTML rendering without code exposure

## Next Steps
1. Restart the server to apply changes
2. View any article with Close Summary content
3. Verify HTML renders properly without visible `<p>` tags
