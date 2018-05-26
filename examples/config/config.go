package main

import (
	"errors"
	"io/ioutil"
	"time"

	"olympos.io/encoding/edn"
)

type Config struct {
	Db    DbConf
	Myapp MyappConf
	Log   Loglevel
	Env   string
}

func (c *Config) UnmarshalEDN(bs []byte) error {
	// Env is a Keyword when we read it in, but we don't really want to unwrap it
	// when we pass it around. So we use a dummy struct and unwrap it afterwards.
	var input struct {
		Db    *DbConf
		Myapp *MyappConf
		Log   *Loglevel
		Env   edn.Keyword
	}
	input.Db = &c.Db
	input.Myapp = &c.Myapp
	input.Log = &c.Log
	err := edn.Unmarshal(bs, &input)
	c.Env = string(input.Env)
	return err
}

type Loglevel edn.Keyword

const (
	LogDebug = Loglevel("debug")
	LogInfo  = Loglevel("info")
	LogWarn  = Loglevel("warn")
)

// This step could also be done as a post validation step instead.
func (l *Loglevel) UnmarshalEDN(bs []byte) error {
	var kw edn.Keyword
	err := edn.Unmarshal(bs, &kw)
	if err != nil {
		return err
	}
	ll := Loglevel(kw)
	switch ll {
	case LogDebug, LogInfo, LogWarn:
		*l = ll
		return nil
	default:
		return errors.New("Unknown log level: " + string(kw))
	}
}

type DbConf struct {
	User     string
	Password string `edn:"pwd"`
	Host     string
	Db       string
	Port     int
}

type MyappConf struct {
	Port         int
	Features     FeatureSet
	FooSetup     FooConfig `edn:"foo"`
	ForeverDate  time.Time `edn:"forever-date"`
	ProcessCount int       `edn:"process-pool"`
}

type FeatureSet map[Feature]bool

// This one supports `:all` as a shortcut to list all features, so manual
// handling for that here.
func (fs *FeatureSet) UnmarshalEDN(bs []byte) error {
	var fm map[Feature]bool
	if string(bs) == ":all" {
		fm = make(map[Feature]bool)
		for _, feature := range ListFeatures() {
			fm[feature] = true
		}
		*fs = FeatureSet(fm)
		return nil
	}

	err := edn.Unmarshal(bs, &fm)
	*fs = FeatureSet(fm)
	return err
}

type Feature edn.Keyword

const (
	AdminPanel        = Feature("admin-panel")
	SwearFilter       = Feature("swear-filter")
	Ads               = Feature("ads")
	KeyboardShortcuts = Feature("keyboard-shortcuts")
)

func ListFeatures() []Feature {
	return []Feature{AdminPanel, SwearFilter, Ads, KeyboardShortcuts}
}

func (feature *Feature) UnmarshalEDN(bs []byte) error {
	var kw edn.Keyword
	err := edn.Unmarshal(bs, &kw)
	if err != nil {
		return err
	}
	f := Feature(kw)
	switch f {
	case AdminPanel, SwearFilter, Ads, KeyboardShortcuts:
		*feature = f
		return nil
	default:
		return errors.New("Unknown feature: " + string(kw))
	}
}

type FooConfig struct {
	Hostname         string
	ApiKeys          []string      `edn:"api-keys"`
	RecheckFrequency time.Duration `edn:"recheck-frequency"`
}

func init() {
	// Add the #duration tag to edn -- time.ParseDuration conveniently satisfies
	// the interface, which is great.
	edn.AddTagFn("duration", time.ParseDuration)
}

func ReadConf(fname string) (*Config, error) {
	bs, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	var c Config
	err = edn.Unmarshal(bs, &c)
	return &c, err
}
