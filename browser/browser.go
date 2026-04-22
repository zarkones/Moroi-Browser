package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

var (
	ActionTimeout   = 365 * 24 * time.Hour // per action, TODO: Improve to be dynamic based on user settings
	MaxSnapshotRefs = 800                  // safety for huge pages
)

type Browser struct {
	ctx    context.Context
	cancel context.CancelFunc
	refs   map[string]string // ref -> CSS selector (updated on every Snapshot)
}

// SnapshotResult is what the JS returns
type SnapshotResult struct {
	Snapshot string            `json:"snapshot"`
	Refs     map[string]string `json:"refs"`
}

// New creates a fresh browser instance (recommended for each agent session).
func New(headless bool) (*Browser, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
		// chromedp.Flag("disable-gpu", true),
		// chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	ctx, cancelTimeout := context.WithTimeout(ctx, ActionTimeout)

	b := &Browser{
		ctx:    ctx,
		cancel: func() { cancelTimeout(); cancelCtx(); cancelAlloc() },
		refs:   make(map[string]string),
	}
	return b, nil
}

func (b *Browser) Close() {
	if b.cancel != nil {
		b.cancel()
	}
}

func (b *Browser) recheckCtx() (err error) {
	if b.ctx.Err() == nil {
		return
	}
	b, err = New(false)
	return err
}

// Snapshot returns the exact AI-readable hierarchical format you showed + updates the internal ref map.
func (b *Browser) Snapshot() (string, error) {
	b.recheckCtx()
	ctx, cancel := context.WithTimeout(b.ctx, ActionTimeout)
	_ = cancel

	var jsonStr string
	js := snapshotScript() // defined below
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &jsonStr)); err != nil {
		return "", fmt.Errorf("snapshot failed: %w", err)
	}

	var res SnapshotResult
	if err := json.Unmarshal([]byte(jsonStr), &res); err != nil {
		return "", fmt.Errorf("failed to parse snapshot JSON: %w", err)
	}

	// keep only first N refs for safety
	if len(res.Refs) > MaxSnapshotRefs {
		for k := range res.Refs {
			if len(res.Refs) <= MaxSnapshotRefs {
				break
			}
			delete(res.Refs, k)
		}
	}

	b.refs = res.Refs
	return res.Snapshot, nil
}

// ClickByRef clicks using the stable ref from the last snapshot (most reliable way for agents).
func (b *Browser) ClickByRef(ref string) error {
	b.recheckCtx()
	sel, ok := b.refs[ref]
	if !ok {
		return fmt.Errorf("ref %s not found in last snapshot (call Snapshot first)", ref)
	}

	const actionTimeoutThisCall = 10 * time.Second // or 30s; not 365h
	ctx, cancel := context.WithTimeout(b.ctx, actionTimeoutThisCall)
	defer cancel()
	if err := chromedp.Run(ctx, chromedp.Click(sel, chromedp.ByQuery)); err != nil {
		return err
	}
	return nil
}

// TypeByRef types into an input/textarea using the ref.
func (b *Browser) TypeByRef(ref, text string) error {
	b.recheckCtx()
	sel, ok := b.refs[ref]
	if !ok {
		return fmt.Errorf("ref %s not found", ref)
	}
	ctx, cancel := context.WithTimeout(b.ctx, ActionTimeout)
	_ = cancel
	return chromedp.Run(ctx,
		chromedp.Focus(sel, chromedp.ByQuery),
		chromedp.SendKeys(sel, text, chromedp.ByQuery),
	)
}

// ScrollIntoViewByRef scrolls an element into view.
func (b *Browser) ScrollIntoViewByRef(ref string) error {
	b.recheckCtx()
	sel, ok := b.refs[ref]
	if !ok {
		return fmt.Errorf("ref %s not found", ref)
	}
	ctx, cancel := context.WithTimeout(b.ctx, ActionTimeout)
	_ = cancel
	return chromedp.Run(ctx, chromedp.ScrollIntoView(sel, chromedp.ByQuery))
}

func (b *Browser) GetCurrentURL() (string, error) {
	b.recheckCtx()
	ctx, cancel := context.WithTimeout(b.ctx, ActionTimeout)
	_ = cancel
	var url string
	err := chromedp.Run(ctx, chromedp.Location(&url))
	return url, err
}

// Keep your original helpers (now on the struct)
func (b *Browser) Navigate(url string) error {
	b.recheckCtx()
	ctx, cancel := context.WithTimeout(b.ctx, ActionTimeout)
	_ = cancel
	return chromedp.Run(ctx, chromedp.Navigate(url))
}

func (b *Browser) GetHTML() (string, error) {
	b.recheckCtx()
	ctx, cancel := context.WithTimeout(b.ctx, ActionTimeout)
	_ = cancel
	var html string
	err := chromedp.Run(ctx, chromedp.OuterHTML("html", &html))
	return html, err
}

func (b *Browser) GetText() (string, error) {
	b.recheckCtx()
	ctx, cancel := context.WithTimeout(b.ctx, ActionTimeout)
	_ = cancel
	var text string
	err := chromedp.Run(ctx, chromedp.Text("html", &text))
	return text, err
}

func (b *Browser) NavigateBack() error {
	b.recheckCtx()
	ctx, cancel := context.WithTimeout(b.ctx, ActionTimeout)
	_ = cancel
	return chromedp.Run(ctx, chromedp.NavigateBack())
}

func (b *Browser) NavigateForward() error {
	b.recheckCtx()
	ctx, cancel := context.WithTimeout(b.ctx, ActionTimeout)
	_ = cancel
	return chromedp.Run(ctx, chromedp.NavigateForward())
}

func (b *Browser) Refresh() error {
	b.recheckCtx()
	ctx, cancel := context.WithTimeout(b.ctx, ActionTimeout)
	_ = cancel
	return chromedp.Run(ctx, chromedp.Reload())
}

func (b *Browser) TakeScreenshot(path string) error {
	b.recheckCtx()
	ctx, cancel := context.WithTimeout(b.ctx, ActionTimeout)
	_ = cancel
	var buf []byte
	if err := chromedp.Run(ctx, chromedp.CaptureScreenshot(&buf)); err != nil {
		return err
	}
	return os.WriteFile(path, buf, 0644)
}

// The JS that generates your exact snapshot format (with refs)
func snapshotScript() string {
	return `
(function() {
  let refCounter = 1;
  const refMap = {};

  function getSelector(el) {
    if (!el) return '';
    if (el.id) return '#' + el.id;
    let sel = el.tagName.toLowerCase();
    if (el.className && typeof el.className === 'string') {
      sel += '.' + el.className.trim().replace(/\s+/g, '.');
    }
    return sel;
  }

  function buildTree(node, depth = 0) {
    if (!node || node.nodeType !== 1 || depth > 12) return '';
    let out = '';
    const prefix = '  '.repeat(depth) + '- ';
    let role = '';
    let text = (node.textContent || '').trim().replace(/\n/g, ' ').slice(0, 120);
    let url = node.href || node.getAttribute('href') || '';
    let placeholder = node.placeholder || '';
    const tag = node.tagName.toLowerCase();
    const ariaLabel = node.getAttribute('aria-label') || '';

    // Map common elements exactly like your GitHub example
    if (tag === 'a') role = 'link';
    else if (tag === 'button') role = 'button';
    else if (/^h[1-6]$/.test(tag)) role = 'heading';
    else if (tag === 'input' || tag === 'textarea') role = 'textbox';
    else if (tag === 'img') role = 'img';
    else if (tag === 'nav') role = 'navigation';
    else if (tag === 'main') role = 'main';
    else if (tag === 'form') role = 'form';
    else if (tag === 'ul' || tag === 'ol') role = 'list';
    else if (tag === 'li') role = 'listitem';
    else if (node.getAttribute('role')) role = node.getAttribute('role');

    if (role) {
      const displayName = ariaLabel || text || 'unnamed';
      const ref = 'e' + refCounter++;
      refMap[ref] = getSelector(node);

      let extra = '';
      if (/^h[1-6]$/.test(tag)) extra = ' [level=' + tag[1] + ']';
      if (role === 'heading' && ariaLabel) extra += ' "' + ariaLabel + '"';

      out += prefix + role + ' "' + displayName + '" [ref=' + ref + ']' + extra + ':\n';
      if (url) out += '  '.repeat(depth + 1) + '- /url: ' + url + '\n';
      if (placeholder) out += '  '.repeat(depth + 1) + '- /placeholder: ' + placeholder + '\n';
    } else {
      // structural containers (region, banner, tablist, etc.)
      if (node.getAttribute('role') === 'region' || tag === 'section' || tag === 'header' || tag === 'footer') {
        out += prefix + 'region:\n';
      }
    }

    // recurse children
    for (let i = 0; i < node.children.length; i++) {
      out += buildTree(node.children[i], depth + 1);
    }
    return out;
  }

  const tree = buildTree(document.documentElement, 0);
  return JSON.stringify({
    snapshot: 'Snapshot:\n- document:\n' + tree,
    refs: refMap
  });
})()
`
}
