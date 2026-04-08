//go:build wasm && js

package main

import (
	"syscall/js"
)

// AsyncWrapper transforms a Go function into a JS Promise-returning function.
// The 'fn' should take JS args and return (result, error).
func AsyncWrapper(fn func(args []js.Value) (any, error)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		// Create a handler for the Promise
		handler := js.FuncOf(func(this js.Value, pArgs []js.Value) any {
			resolve := pArgs[0]
			reject := pArgs[1]

			// Run the Go code in a goroutine so it doesn't block the UI
			go func() {
				res, err := fn(args)
				if err != nil {
					// Create a JS Error object so it's a proper exception
					errorObj := js.Global().Get("Error").New(err.Error())
					reject.Invoke(errorObj)
				} else {
					resolve.Invoke(res)
				}
			}()

			return nil
		})

		// Return the new Promise to JavaScript
		return js.Global().Get("Promise").New(handler)
	})
}
