package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var units = map[string]int64{
	"d": int64(time.Hour) * 24,
	"w": int64(time.Hour) * 24 * 7,
}

type LongDuration string

var ErrInvalidLongDuration = fmt.Errorf("invalid long duration amount")

func (ld LongDuration) ErrInvalid(detail string) error {
	return fmt.Errorf("%w %q: %s", ErrInvalidLongDuration, ld, detail)
}

func (ld LongDuration) AddTo(t time.Time) (time.Time, error) {
	s := string(ld)
	if s == "" {
		return t, nil
	}
	unit := s[len(s)-1:]
	multiplier, ok := units[unit]
	if !ok {
		d, err := time.ParseDuration(s)
		if err != nil {
			return t, err
		}
		return t.Add(d), nil
	}
	parts := strings.Split(s[:len(s)-1], ".")
	length := len(parts)
	if length == 0 {
		return t, nil
	}
	if length > 2 {
		return t, ld.ErrInvalid("too many decimal parts")
	}
	ones, neg, onesErr := ld.parseInt(parts[0])
	if onesErr != nil {
		return t, ld.ErrInvalid(onesErr.Error())
	}
	d := ones * multiplier
	if length == 2 {
		decimal, decimalErr := strconv.ParseFloat("0."+parts[1], 64)
		if decimalErr != nil {
			return t, ld.ErrInvalid(decimalErr.Error())
		}
		d += int64(decimal * float64(multiplier))
	}
	if neg {
		d *= -1
	}
	return t.Add(time.Duration(d)), nil
}

func (ld LongDuration) parseInt(s string) (amount int64, neg bool, err error) {
	if len(s) == 0 {
		return 0, false, nil
	}
	if s[:1] == "-" {
		neg = true
		s = s[1:]
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false, ld.ErrInvalid(err.Error())
	}
	return v, neg, nil
}
