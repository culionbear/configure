package configure

import (
	"errors"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"maps"
	"sync"
)

type _Unit struct {
	option *UnitOption
	unit   Unit
}

type _Driver struct {
	driver     Driver
	hotUpgrade bool
	units      map[string]*_Unit
}

type Configure struct {
	defaultDriver    *_Driver
	drivers          map[string]*_Driver
	units            mapset.Set[string]
	onceSetUnits     sync.Once
	firstIgnoreError bool
}

type Option func(*Configure) error

func WithOtherDrivers(drivers ...Driver) Option {
	return func(conf *Configure) error {
		for _, driver := range drivers {
			if driver == nil {
				return errors.New("driver is nil")
			}
			if driver.Name() == "" {
				return errors.New("driver name is empty")
			}
			if _, ok := conf.drivers[driver.Name()]; ok {
				return fmt.Errorf("driver %s already exists", driver.Name())
			}
			conf.drivers[driver.Name()] = &_Driver{
				driver: driver,
			}
		}
		return nil
	}
}

func FirstIgnoreError() Option {
	return func(conf *Configure) error {
		conf.firstIgnoreError = true
		return nil
	}
}

func New(driver Driver, opts ...Option) (*Configure, error) {
	if driver == nil {
		return nil, errors.New("driver is nil")
	}
	if driver.Name() == "" {
		return nil, errors.New("driver name is empty")
	}
	conf := &Configure{
		defaultDriver: &_Driver{driver: driver},
		drivers:       make(map[string]*_Driver),
		units:         mapset.NewThreadUnsafeSet[string](),
	}
	conf.drivers[driver.Name()] = conf.defaultDriver
	for _, opt := range opts {
		if err := opt(conf); err != nil {
			return nil, err
		}
	}
	return conf, nil
}

func (conf *Configure) Listen(units ...Unit) (err error) {
	conf.onceSetUnits.Do(func() {
		for _, value := range units {
			err = conf.addUnit(value)
			if err != nil {
				return
			}
		}
		var list []*topologicalSortNode
		list, err = conf.topologicalSort()
		if err != nil {
			return
		}
		for _, node := range list {
			err = conf.initUnit(node)
			if err != nil {
				if conf.firstIgnoreError {
					err = nil
					continue
				}
				return
			}
		}
		successListenDrivers := make([]Driver, 0)
		for _, driver := range conf.drivers {
			if !driver.hotUpgrade {
				continue
			}
			ctx := &Context{
				callback: conf.callback,
				driver:   driver.driver,
				schemes:  make(map[string]KV),
			}
			for _, unit := range driver.units {
				if !unit.option.allowHotUpgrade {
					continue
				}
				ctx.schemes[unit.option.key] = maps.Clone(unit.option.schemes)
			}
			err = driver.driver.Listen(ctx)
			if err != nil {
				for _, val := range successListenDrivers {
					val.Close()
				}
				return
			}
			successListenDrivers = append(successListenDrivers, driver.driver)
		}
	})
	return
}

func (conf *Configure) Close() {
	for _, driver := range conf.drivers {
		driver.driver.Close()
	}
}

type topologicalSortNode struct {
	unit   *_Unit
	driver Driver
}

func (conf *Configure) topologicalSort() ([]*topologicalSortNode, error) {
	list, keys, point := make([]*topologicalSortNode, conf.units.Cardinality(), conf.units.Cardinality()), mapset.NewThreadUnsafeSet[string](), 0
	for {
		if point == len(list) {
			break
		}
		current := point
		for _, driver := range conf.drivers {
			for _, unit := range driver.units {
				if keys.Contains(unit.option.key) {
					continue
				}
				preUnits := unit.unit.PreUnits()
				if preUnits == nil {
					list[point] = &topologicalSortNode{
						driver: driver.driver,
						unit:   unit,
					}
					point++
					keys.Add(unit.option.key)
					continue
				}
				sum := len(preUnits)
				for _, preUnit := range preUnits {
					if !conf.units.Contains(preUnit) {
						return nil, fmt.Errorf("unit %s not found", preUnit)
					}
					if keys.Contains(preUnit) {
						sum--
					}
				}
				if sum == 0 {
					list[point] = &topologicalSortNode{
						driver: driver.driver,
						unit:   unit,
					}
					point++
					keys.Add(unit.option.key)
				}
			}
		}
		if current == point {
			return nil, errors.New("unit loop")
		}
	}
	return list, nil
}

func (conf *Configure) initUnit(node *topologicalSortNode) error {
	driver, unit := node.driver, node.unit
	buf, err := driver.Get(unit.option.key, maps.Clone(unit.option.schemes))
	if err != nil {
		return err
	}
	if unit.option.before != nil && unit.option.before.firstExecute {
		err = unit.option.before.fn()
		if err != nil {
			return err
		}
	}
	if unit.option.locker != nil {
		unit.option.locker.Lock()
	}
	err = unit.option.parser.Unmarshal(unit.unit.Slot(), buf)
	if unit.option.locker != nil {
		unit.option.locker.Unlock()
	}
	if err != nil {
		return err
	}
	if unit.option.after != nil && unit.option.after.firstExecute {
		err = unit.option.after.fn()
		if err != nil {
			return err
		}
	}
	return nil
}

func (conf *Configure) callback(driver Driver, key string, buf []byte) {
	_driver := conf.drivers[driver.Name()]
	if _driver == nil {
		return
	}
	_unit := _driver.units[key]
	if _unit == nil {
		return
	}
	err := conf.execCallback(_unit.option.before, _unit.option.errorFunc)
	if err != nil {
		return
	}
	if _unit.option.locker != nil {
		_unit.option.locker.Lock()
	}
	err = _unit.option.parser.Unmarshal(_unit.unit.Slot(), buf)
	if _unit.option.locker != nil {
		_unit.option.locker.Unlock()
	}
	if err != nil {
		if _unit.option.errorFunc != nil {
			_unit.option.errorFunc(err)
		}
		return
	}
	_ = conf.execCallback(_unit.option.after, _unit.option.errorFunc)
}

func (conf *Configure) execCallback(fn *callback, errFunc func(error)) error {
	if fn == nil {
		return nil
	}
	err := fn.fn()
	if err != nil && errFunc != nil {
		errFunc(err)
	}
	return err
}

func (conf *Configure) addUnit(value Unit) error {
	if value == nil {
		return errors.New("unit is nil")
	}
	option := value.Option()
	if option == nil {
		return errors.New("unit option is nil")
	}
	if option.key == "" {
		return errors.New("unit key is empty")
	}
	if option.parser == nil {
		return errors.New("unit parser is nil")
	}
	if conf.units.Contains(option.key) {
		return fmt.Errorf("unit %s already exists", option.key)
	}
	info := &_Unit{
		option: option,
		unit:   value,
	}
	if option.driverName == "" {
		err := conf.joinUnitInDriver(conf.defaultDriver, info)
		if err != nil {
			return err
		}
	} else if driver, ok := conf.drivers[option.driverName]; !ok {
		return fmt.Errorf("driver %s not found", option.driverName)
	} else {
		err := conf.joinUnitInDriver(driver, info)
		if err != nil {
			return err
		}
	}
	conf.units.Add(option.key)
	return nil
}

func (conf *Configure) joinUnitInDriver(driver *_Driver, unit *_Unit) error {
	if driver.units == nil {
		driver.units = map[string]*_Unit{
			unit.option.key: unit,
		}
		if unit.option.allowHotUpgrade {
			driver.hotUpgrade = true
		}
		return nil
	}
	driver.units[unit.option.key] = unit
	if unit.option.allowHotUpgrade {
		driver.hotUpgrade = true
	}
	return nil
}
