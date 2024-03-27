package executor

import (
	"fmt"
	"os"
	"errors"
	"reflect"
	"strings"
	"autochrome/executor/chrome"
	"autochrome/executor/symbols"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/traefik/yaegi/stdlib/syscall"
	"github.com/traefik/yaegi/stdlib/unrestricted"
)

type Executor struct {
	Interp  *interp.Interpreter
}

func NewExecutor() (*Executor, error) {
	exec := &Executor{}

	i := interp.New(interp.Options{
			Env: os.Environ(),
	})
	err := i.Use(stdlib.Symbols)
	if err != nil {
		return nil, err
	}

	err = i.Use(symbols.Symbols)
	if err != nil {
		return nil, err
	}

	err = i.Use(syscall.Symbols)
	if err != nil {
		return nil, err
	}

	if err = os.Setenv("YAEGI_SYSCALL", "1"); err != nil {
		return nil, err
	}

	err = i.Use(unrestricted.Symbols)
	if err != nil {
		return nil, err
	}

	if err = os.Setenv("YAEGI_UNRESTRICTED", "1"); err != nil {
		return nil, err
	}

	exec.Interp = i

	return exec, nil
}

func (d *Executor) safeEval(code string) (res reflect.Value, err error) {
	if strings.TrimSpace(code) == "" {
		return reflect.Value{}, nil
	}

	defer func() {
		e := recover()
		if e == nil {
			return
		}
		switch v := e.(type) {
		case error:
			err = v
		default:
			err = fmt.Errorf("%v", v)
		}
	}()

	res, err = d.Interp.Eval(code)
	if err != nil {
		return res, err
	}
	return res, err
}

func (d *Executor) ChromeNew() (*chrome.Chrome, error) {
	code := `
	import "fmt"
	import "os"
	import "os/signal"
	import "syscall"
	import "context"
	import "time"
	import "github.com/chromedp/chromedp"
	import "github.com/chromedp/chromedp/kb"
	import "autochrome/executor/chrome"

	func unused() {
		fmt.Println(os.Getenv("PATH"))
		fmt.Println(kb.Enter)
		time.Sleep(1*time.Second)
	}

	var VarChrome = chrome.New()
	var VarFunc   = func () (func(ctx context.Context) error) {
		return func(ctx context.Context) error {
			return nil
		}
	}

	var SignalFunc = func () {
		var sigChan chan os.Signal = make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT)
		go func() {
			for sig := range sigChan {
				VarChrome.NavigateAndWaitReady()
			}
		}()
	}
	`
	_, err := d.safeEval(code)
	if err != nil {
		return nil, err
	}

	_, serr := d.safeEval(fmt.Sprintf(`SignalFunc()`))
	if serr != nil {
		return nil, serr
	}

	value, verr := d.safeEval(fmt.Sprintf(`VarChrome`))
	if verr != nil {
		return nil, verr
	}
	varChrome, ok := value.Interface().(*chrome.Chrome)
	if !ok {
		return nil, errors.New("Func 'chrome.New' return type is not '*chrome.Chrome'!")
	}
	return varChrome, nil
}

func (d *Executor) ChromeDelete() error {
	_, err := d.safeEval(`chrome.Delete(VarChrome)`)
	if err != nil {
		return err
	}
	return nil
}

func (d *Executor) ChromeSetSize(width, height int) error {
	_, err := d.safeEval(fmt.Sprintf(`chrome.SetSize(%d, %d)`, width, height))
	if err != nil {
		return err
	}
	return nil
}

func (d *Executor) ChromeSetUrl(url string) error {
	_, err := d.safeEval(fmt.Sprintf(`VarChrome.SetUrl("%s")`, url))
	if err != nil {
		return err
	}
	return nil
}

func (d *Executor) ChromeGetHtml() (string, error) {
	value, err := d.safeEval(fmt.Sprintf(`VarChrome.GetHtml()`))
	if err != nil {
		return "", err
	}
	str, ok := value.Interface().(string)
	if !ok {
		return "", errors.New("Func 'GetHtml' return type is not 'string'!")
	}
	return str, nil
}

func (d *Executor) ChromeNewTab() error {
	_, err := d.safeEval(fmt.Sprintf(`VarChrome.NewTab()`))
	if err != nil {
		return err
	}
	return nil
}

func (d *Executor) ChromeNavigateAndWaitReady() error {
	value, err := d.safeEval(fmt.Sprintf(`VarChrome.NavigateAndWaitReady()`))
	if err != nil {
		return err
	}
	str, ok := value.Interface().(string)
	if !ok {
		return errors.New("Func 'NavigateAndWaitReady' return type is not 'string'!")
	}
	if len(str) > 0 {
		return errors.New(str)
	}
	return nil
}

func (d *Executor) ChromeRunTasks(code string) error {
	_, aerr := d.safeEval(fmt.Sprintf(`
		VarFunc = func () (func(ctx context.Context) error) {
			return func(ctx context.Context) error {
				%s
			}
		}
	`, code))

	if aerr != nil {
		return aerr
	}

	value, berr := d.safeEval(fmt.Sprintf(`VarChrome.RunTasks(VarFunc())`))

	if berr != nil {
		return berr
	}

	str, ok := value.Interface().(string)
	if !ok {
		return errors.New("Func 'RunTasks' return type is not 'string'!")
	}

	if len(str) > 0 {
		return errors.New(str)
	}

	return nil
}


