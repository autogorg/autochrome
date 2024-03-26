package symbols

import "reflect"

//go:generate go run github.com/traefik/yaegi/internal/cmd/extract github.com/chromedp/chromedp
//go:generate go run github.com/traefik/yaegi/internal/cmd/extract github.com/chromedp/chromedp/kb
//go:generate go run github.com/traefik/yaegi/internal/cmd/extract aicoder/dslexec/chrome

// Symbols variable stores the map of stdlib symbols per package.
var Symbols = map[string]map[string]reflect.Value{}

// MapTypes variable contains a map of functions which have an interface{} as parameter but
// do something special if the parameter implements a given interface.
var MapTypes = map[reflect.Value][]reflect.Type{}
