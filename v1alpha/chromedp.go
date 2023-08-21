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

func (x *Chromedp) Bind(edges []runtime.Edge) error {
	for _, e := range edges {
		switch res := e.Result.(type) {
		case *Source:
			x.Source = res
		default:
			return fmt.Errorf("ChromedpRunner: unexpected type in input: %T", e.Result)
		}
	}
	if x.Source == nil {
		return fmt.Errorf("ChromedpRunner: expeced *Source input")
	}

	return nil
}

func (r *Chromedp) Run(ctx context.Context) (any, error) {
	// Start chromedp context.
	chromedpCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	u, err := r.Source.URL()
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
