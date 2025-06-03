package utils

import (
	"reflect"
	"runtime"
)

// GetFunctionName returns the name of a function from its reference.
//
// This function uses reflection to get the function's pointer and
// retrieves its name using the runtime package. It can be used to
// dynamically obtain the name of a function at runtime.
//
// Parameters:
//   - fn (any): A reference to the function whose name you want to retrieve.
//
// Returns:
//   - string: The name of the function, typically in the form of "packageName.funcName".
//
// Example usage:
//
//	func example() {}
//	name := utils.GetFunctionName(example) // Returns "main.example"
func GetFunctionName(fn any) string {
	// Get the pointer to the function using reflection
	pc := reflect.ValueOf(fn).Pointer()

	// Use the pointer to get the function object
	funcObj := runtime.FuncForPC(pc)

	// Return the name of the function
	return funcObj.Name()
}

// DefaultValue checks if a given value is non-empty and returns it.
// If the value is empty, it returns a fallback (default) value.
//
// Parameters:
//   - value (string): The value to check for non-emptiness.
//   - other (string): The fallback value to return if the input `value` is empty.
//
// Returns:
//   - string: The `value` if it is non-empty, otherwise the `other` value.
//
// Example usage:
//
//	result := utils.DefaultValue("", "default")  // Returns "default"
//	result := utils.DefaultValue("custom", "default") // Returns "custom"
func DefaultValue(value string, other string) string {
	// Return `value` if it is not empty, otherwise return `other`
	if value != "" {
		return value
	}
	return other
}
