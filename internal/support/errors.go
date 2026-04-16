package support

import "fmt"

// RuyiPageError 是全部公开错误的根错误类型。
type RuyiPageError struct {
	Message string
	Cause   error
}

func NewRuyiPageError(message string, cause error) *RuyiPageError {
	return &RuyiPageError{
		Message: message,
		Cause:   cause,
	}
}

func (e *RuyiPageError) Error() string {
	if e == nil {
		return ""
	}
	return composeErrorMessage(e.Message, e.Cause)
}

func (e *RuyiPageError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type ElementNotFoundError struct {
	RuyiPageError
}

func NewElementNotFoundError(message string, cause error) *ElementNotFoundError {
	return &ElementNotFoundError{RuyiPageError: newBaseError(message, cause)}
}

func (e *ElementNotFoundError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &e.RuyiPageError
}

type ElementLostError struct {
	RuyiPageError
}

func NewElementLostError(message string, cause error) *ElementLostError {
	return &ElementLostError{RuyiPageError: newBaseError(message, cause)}
}

func (e *ElementLostError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &e.RuyiPageError
}

type ContextLostError struct {
	RuyiPageError
}

func NewContextLostError(message string, cause error) *ContextLostError {
	return &ContextLostError{RuyiPageError: newBaseError(message, cause)}
}

func (e *ContextLostError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &e.RuyiPageError
}

type BiDiError struct {
	RuyiPageError
	Code        string
	BiDiMessage string
	Stacktrace  string
}

func NewBiDiError(code, message, stacktrace string, cause error) *BiDiError {
	return &BiDiError{
		RuyiPageError: newBaseError(formatBiDiMessage(code, message), cause),
		Code:          code,
		BiDiMessage:   message,
		Stacktrace:    stacktrace,
	}
}

func (e *BiDiError) Error() string {
	if e == nil {
		return ""
	}
	message := e.Message
	if message == "" {
		message = formatBiDiMessage(e.Code, e.BiDiMessage)
	}
	return composeErrorMessage(message, e.Cause)
}

func (e *BiDiError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &e.RuyiPageError
}

type PageDisconnectedError struct {
	RuyiPageError
}

func NewPageDisconnectedError(message string, cause error) *PageDisconnectedError {
	return &PageDisconnectedError{RuyiPageError: newBaseError(message, cause)}
}

func (e *PageDisconnectedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &e.RuyiPageError
}

type JavaScriptError struct {
	RuyiPageError
	ExceptionDetails any
}

func NewJavaScriptError(message string, exceptionDetails any, cause error) *JavaScriptError {
	return &JavaScriptError{
		RuyiPageError:    newBaseError(message, cause),
		ExceptionDetails: exceptionDetails,
	}
}

func (e *JavaScriptError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &e.RuyiPageError
}

type BrowserConnectError struct {
	RuyiPageError
}

func NewBrowserConnectError(message string, cause error) *BrowserConnectError {
	return &BrowserConnectError{RuyiPageError: newBaseError(message, cause)}
}

func (e *BrowserConnectError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &e.RuyiPageError
}

type BrowserLaunchError struct {
	RuyiPageError
}

func NewBrowserLaunchError(message string, cause error) *BrowserLaunchError {
	return &BrowserLaunchError{RuyiPageError: newBaseError(message, cause)}
}

func (e *BrowserLaunchError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &e.RuyiPageError
}

type AlertExistsError struct {
	RuyiPageError
}

func NewAlertExistsError(message string, cause error) *AlertExistsError {
	return &AlertExistsError{RuyiPageError: newBaseError(message, cause)}
}

func (e *AlertExistsError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &e.RuyiPageError
}

type WaitTimeoutError struct {
	RuyiPageError
}

func NewWaitTimeoutError(message string, cause error) *WaitTimeoutError {
	return &WaitTimeoutError{RuyiPageError: newBaseError(message, cause)}
}

func (e *WaitTimeoutError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &e.RuyiPageError
}

type NoRectError struct {
	RuyiPageError
}

func NewNoRectError(message string, cause error) *NoRectError {
	return &NoRectError{RuyiPageError: newBaseError(message, cause)}
}

func (e *NoRectError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &e.RuyiPageError
}

type CanNotClickError struct {
	RuyiPageError
}

func NewCanNotClickError(message string, cause error) *CanNotClickError {
	return &CanNotClickError{RuyiPageError: newBaseError(message, cause)}
}

func (e *CanNotClickError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &e.RuyiPageError
}

type LocatorError struct {
	RuyiPageError
}

func NewLocatorError(message string, cause error) *LocatorError {
	return &LocatorError{RuyiPageError: newBaseError(message, cause)}
}

func (e *LocatorError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &e.RuyiPageError
}

func newBaseError(message string, cause error) RuyiPageError {
	return RuyiPageError{
		Message: message,
		Cause:   cause,
	}
}

func composeErrorMessage(message string, cause error) string {
	switch {
	case message != "" && cause != nil:
		return fmt.Sprintf("%s: %v", message, cause)
	case message != "":
		return message
	case cause != nil:
		return cause.Error()
	default:
		return ""
	}
}

func formatBiDiMessage(code, message string) string {
	switch {
	case code != "" && message != "":
		return fmt.Sprintf("%s: %s", code, message)
	case code != "":
		return code
	default:
		return message
	}
}
