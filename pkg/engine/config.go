package engine

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"reflect"
)

type ExchangeStrategyMount struct {
	// Mounts contains the ExchangeSession name to mount
	Mounts []string `json:"mounts"`

	// Strategy is the strategy we loaded from config
	Strategy SingleExchangeStrategy `json:"strategy"`
}

func (m *ExchangeStrategyMount) Map() (map[string]interface{}, error) {
	strategyID := m.Strategy.Id()

	var params map[string]interface{}

	out, err := json.Marshal(m.Strategy)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(out, &params); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"on":       m.Mounts,
		strategyID: params,
	}, nil
}

type Config struct {

	Sessions map[string]*ExchangeSession `json:"sessions,omitempty" yaml:"sessions,omitempty"`

	ExchangeStrategies      []ExchangeStrategyMount `json:"-" yaml:"-"`
	CrossExchangeStrategies []CrossExchangeStrategy `json:"-" yaml:"-"`
}

type Stash map[string]interface{}

func loadStash(config []byte) (Stash, error) {
	stash := make(Stash)
	if err := yaml.Unmarshal(config, stash); err != nil {
		return nil, err
	}

	return stash, nil
}

func Load(configFile string, loadStrategies bool) (*Config, error) {
	var config Config

	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	stash, err := loadStash(content)
	if err != nil {
		return nil, err
	}

	if loadStrategies {
		if err := loadExchangeStrategies(&config, stash); err != nil {
			return nil, err
		}

		if err := loadCrossExchangeStrategies(&config, stash); err != nil {
			return nil, err
		}
	}

	return &config, nil
}

func loadCrossExchangeStrategies(config *Config, stash Stash) (err error) {
	exchangeStrategiesConf, ok := stash["crossExchangeStrategies"]
	if !ok {
		return nil
	}

	if len(LoadedCrossExchangeStrategies) == 0 {
		return errors.New("no cross exchange strategy is registered")
	}

	configList, ok := exchangeStrategiesConf.([]interface{})
	if !ok {
		return errors.New("expecting list in crossExchangeStrategies")
	}

	for _, entry := range configList {
		configStash, ok := entry.(Stash)
		if !ok {
			return fmt.Errorf("strategy config should be a map, given: %T %+v", entry, entry)
		}

		for id, conf := range configStash {
			// look up the real struct type
			if st, ok := LoadedCrossExchangeStrategies[id]; ok {
				val, err := reUnmarshal(conf, st)
				if err != nil {
					return err
				}

				config.CrossExchangeStrategies = append(config.CrossExchangeStrategies, val.(CrossExchangeStrategy))
			}
		}
	}

	return nil
}

func loadExchangeStrategies(config *Config, stash Stash) (err error) {

	exchangeStrategiesConf, ok := stash["exchangeStrategies"]
	if !ok {
		exchangeStrategiesConf, ok = stash["strategies"]
		if !ok {
			return nil
		}
	}

	if len(LoadedExchangeStrategies) == 0 {
		return errors.New("no exchange strategy is registered")
	}

	configList, ok := exchangeStrategiesConf.([]interface{})
	if !ok {
		return errors.New("expecting list in exchangeStrategies")
	}

	for _, entry := range configList {
		configStash, ok := entry.(Stash)
		if !ok {
			return fmt.Errorf("strategy config should be a map, given: %T %+v", entry, entry)
		}

		var mounts []string
		if val, ok := configStash["on"]; ok {
			switch tv := val.(type) {

			case []string:
				mounts = append(mounts, tv...)

			case string:
				mounts = append(mounts, tv)

			case []interface{}:
				for _, f := range tv {
					s, ok := f.(string)
					if !ok {
						return fmt.Errorf("%+v (%T) is not a string", f, f)
					}

					mounts = append(mounts, s)
				}

			default:
				return fmt.Errorf("unexpected mount type: %T value: %+v", val, val)
			}
		}

		for id, conf := range configStash {

			// look up the real struct type
			if _, ok := LoadedExchangeStrategies[id]; ok {
				st, err := NewStrategyFromMap(id, conf)
				if err != nil {
					return err
				}

				config.ExchangeStrategies = append(config.ExchangeStrategies, ExchangeStrategyMount{
					Mounts:   mounts,
					Strategy: st,
				})
			}
		}
	}

	return nil
}

func NewStrategyFromMap(id string, conf interface{}) (SingleExchangeStrategy, error) {
	if st, ok := LoadedExchangeStrategies[id]; ok {
		val, err := reUnmarshal(conf, st)
		if err != nil {
			return nil, err
		}
		return val.(SingleExchangeStrategy), nil
	}

	return nil, fmt.Errorf("strategy %s not found", id)
}

func reUnmarshal(conf interface{}, tpe interface{}) (interface{}, error) {
	// get the type "*Strategy"
	rt := reflect.TypeOf(tpe)

	// allocate new object from the given type
	val := reflect.New(rt)

	// now we have &(*Strategy) -> **Strategy
	valRef := val.Interface()

	plain, err := json.Marshal(conf)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(plain, valRef); err != nil {
		return nil, errors.Wrapf(err, "json parsing error, given payload: %s", plain)
	}

	return val.Elem().Interface(), nil
}