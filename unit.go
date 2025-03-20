package configure

import (
	"maps"
	"sync"
)

type KV map[string]any

type callback struct {
	firstExecute bool
	fn           func() error
}

type UnitOption struct {
	key             string
	parser          Parser
	allowHotUpgrade bool
	errorFunc       func(err error)
	locker          sync.Locker
	after           *callback
	before          *callback
	driverName      string
	schemes         KV
}

func NewOption(key string, parser Parser) *UnitOption {
	return &UnitOption{
		key:     key,
		parser:  parser,
		schemes: make(KV),
	}
}

func (opt *UnitOption) HotUpgrade(allow bool) *UnitOption {
	opt.allowHotUpgrade = allow
	return opt
}

func (opt *UnitOption) WithHotUpgradeErrorFunc(fn func(err error)) *UnitOption {
	opt.errorFunc = fn
	return opt
}

func (opt *UnitOption) WithThreadSafe(locker sync.Locker) *UnitOption {
	opt.locker = locker
	return opt
}

func (opt *UnitOption) WithSlotAfterCallback(fn func() error, useToInit bool) *UnitOption {
	opt.after = &callback{
		fn:           fn,
		firstExecute: useToInit,
	}
	return opt
}

func (opt *UnitOption) WithSlotBeforeCallback(fn func() error, useToInit bool) *UnitOption {
	opt.before = &callback{
		fn:           fn,
		firstExecute: useToInit,
	}
	return opt
}

func (opt *UnitOption) WithDriverName(name string) *UnitOption {
	opt.driverName = name
	return opt
}

func (opt *UnitOption) WithSchemes(schemes KV) *UnitOption {
	if schemes != nil {
		maps.Copy(opt.schemes, schemes)
	}
	return opt
}

func (opt *UnitOption) WithScheme(key string, value any) *UnitOption {
	opt.schemes[key] = value
	return opt
}

type Unit interface {
	Option() *UnitOption
	PreUnits() []string
	Slot() any
}
