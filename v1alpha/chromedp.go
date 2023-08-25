package v1alpha

import (
	"context"
	"fmt"
	"sync"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/wenooij/nuggit/runtime"
)

// Chromedp runs a chromedp executor which fetches the outer HTML of an HTML document.
type Chromedp struct {
	Source *Source `json:"source,omitempty"`
}

func (x *Chromedp) Bind(e runtime.Edge) error {
	switch e.SrcField {
	case "source":
		x.Source = e.Result.(*Source)
	case "":
		*x = *e.Result.(*Chromedp)
	default:
		return fmt.Errorf("ChromedpRunner: unexpected field in input: %v", e.SrcField)
	}
	return nil
}

func (x *Chromedp) Run(ctx context.Context) (any, error) {
	if x.Source == nil {
		return nil, fmt.Errorf("ChromedpRunner: expeced *Source input")
	}
	// Start chromedp context.
	chromedpCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	u, err := x.Source.URL()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	var data chromedpData
	if err := chromedp.Run(chromedpCtx,
		chromedp.Navigate(u.String()),
		chromedp.WaitReady("body", chromedp.ByQuery),
		fetchTask(func(v chromedpData) { data = v; wg.Done() }),
	); err != nil {
		return nil, err
	}
	wg.Wait()

	return data.OuterHTML, nil
}

type chromedpData struct {
	*cdp.Node
	OuterHTML string
}

func fetchTask(cb func(data chromedpData)) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			doc, err := dom.GetDocument().WithDepth(-1).Do(ctx)
			if err != nil {
				return err
			}
			outerHTML, err := dom.GetOuterHTML().WithNodeID(doc.NodeID).Do(ctx)
			if err != nil {
				return err
			}
			var data chromedpData
			data.Node = doc
			data.OuterHTML = outerHTML
			cb(data)
			return nil
		}),
	}
}
