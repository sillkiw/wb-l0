package validation

import (
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var rePhoneStrict = regexp.MustCompile(`^\+[0-9]{11}$`)
var allowedCurrencies = map[string]struct{}{
	"RUB": {}, "KZT": {}, "BYN": {}, "KGS": {},
	"AMD": {}, "TRY": {}, "UZS": {}, "AZN": {}, "GEL": {},
}

func nonEmpty(s string) bool             { return strings.TrimSpace(s) != "" }
func phoneLike(s string) bool            { return rePhoneStrict.MatchString(s) }
func parseableEmail(s string) bool       { _, err := mail.ParseAddress(s); return err == nil }
func pathItem(i int, name string) string { return "items[" + strconv.Itoa(i) + "]." + name }

func isAllowedCurrency(s string) bool {
	s = strings.ToUpper(strings.TrimSpace(s))
	_, ok := allowedCurrencies[s]
	return ok
}

func allowedCurrenciesList() string {
	return "RUB,KZT,BYN,KGS,AMD,TRY,UZS,AZN,GEL"
}

func tsNotFuture(t time.Time) bool {
	return !t.IsZero() && !t.After(time.Now().Add(5*time.Minute))
}
