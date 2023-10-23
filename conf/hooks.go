package conf

import (
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
	"github.com/snowzach/golib/log"
	"github.com/spf13/cast"
)

// DefaultDecodeHooks are the default decoding hooks used to unmarshal into a struct.
// This includes hooks for parsing string to time.Duration, time.Time(RFC3339 format),
// net.IP and net.IPNet. You can use this function to grab the defaults plus add your
// own with the extras option.
func DefaultDecodeHooks(extras ...mapstructure.DecodeHookFunc) []mapstructure.DecodeHookFunc {
	return append([]mapstructure.DecodeHookFunc{
		mapstructure.RecursiveStructToMapHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToTimeHookFunc(time.RFC3339),
		mapstructure.StringToIPHookFunc(),
		mapstructure.StringToIPNetHookFunc(),
		DecimalHookFunc(),
		SLogLevelHookFunc(),
	}, extras...)
}

// DecimalHookFunc returns a DecodeHookFunc that converts
// strings or numbers to decimal.Decimal
func DecimalHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, v interface{}) (interface{}, error) {
		if t != reflect.TypeOf(decimal.Decimal{}) {
			return v, nil
		}
		switch value := v.(type) {
		case float64:
			return decimal.NewFromFloat(value), nil
		case float32:
			return decimal.NewFromFloat32(value), nil
		case int64:
			return decimal.NewFromInt(value), nil
		case int32:
			return decimal.NewFromInt32(value), nil
		case uint32, uint16, uint8, int16, int8, int:
			return decimal.NewFromInt(cast.ToInt64(value)), nil
		case uint64:
			return decimal.NewFromString(strconv.FormatUint(value, 10))
		case string:
			return decimal.NewFromString(value)
		case decimal.Decimal:
			return value, nil
		default:
			return v, fmt.Errorf("unparsable number type: %T", v)
		}
	}
}

// SLogLevelHookFunc decodes string config into a slog.Level.
func SLogLevelHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, v interface{}) (interface{}, error) {

		if t != reflect.TypeOf(slog.LevelInfo) {
			return v, nil
		}
		levelString, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("could not parse log level of type %T", v)
		}

		return log.ParseLogLevel(levelString)

	}
}
