package health

// Pipeline composes a HealthcheckFn with a sequence of middleware functions,
// applying them in declaration order (outermost first).
//
// Example:
//
//	fn := health.Pipeline(
//		baseChecker,
//		health.WithTimeout(cfg, ...),
//		health.WithRetry(cfg, ...),
//	)
type Middleware func(HealthcheckFn) HealthcheckFn

// Pipeline wraps inner with each middleware in order, so that middlewares[0]
// is the outermost wrapper.
func Pipeline(inner HealthcheckFn, middlewares ...Middleware) HealthcheckFn {
	// Apply in reverse so that middlewares[0] ends up outermost.
	result := inner
	for i := len(middlewares) - 1; i >= 0; i-- {
		result = middlewares[i](result)
	}
	return result
}
