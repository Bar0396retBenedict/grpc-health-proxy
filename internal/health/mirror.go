package health

import (
	"context"
	"sync"
)

// MirrorResult holds the results from a mirrored health check pair.
type MirrorResult struct {
	Primary  StatusResult
	Mirror   StatusResult
	Agreement bool
}

// WithMirror runs fn and mirror concurrently and returns a MirrorResult
// containing both outcomes. Agreement is true when both return the same Status
// and neither has an error. The caller decides which result to act on.
func WithMirror(fn HealthCheckFn, mirror HealthCheckFn) func(ctx context.Context, service string) MirrorResult {
	return func(ctx context.Context, service string) MirrorResult {
		var (
			wg            sync.WaitGroup
			primaryResult StatusResult
			mirrorResult  StatusResult
		)

		wg.Add(2)
		go func() {
			defer wg.Done()
			primaryResult = fn(ctx, service)
		}()
		go func() {
			defer wg.Done()
			mirrorResult = mirror(ctx, service)
		}()
		wg.Wait()

		agreement := primaryResult.Err == nil &&
			mirrorResult.Err == nil &&
			primaryResult.Status == mirrorResult.Status

		return MirrorResult{
			Primary:   primaryResult,
			Mirror:    mirrorResult,
			Agreement: agreement,
		}
	}
}
