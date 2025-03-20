package configure

type Context struct {
	driver   Driver
	schemes  map[string]KV // unit key -> scheme
	callback func(driver Driver, key string, buf []byte)
}

func (ctx *Context) HotUpgrade(key string, buf []byte) {
	ctx.callback(ctx.driver, key, buf)
}

func (ctx *Context) Schemes() map[string]KV {
	return ctx.schemes
}

func (ctx *Context) Scheme(key string) KV {
	return ctx.schemes[key]
}

type Driver interface {
	Name() string
	Listen(ctx *Context) error
	Get(key string, schemes KV) ([]byte, error)
	Close()
}
