/*
Copyright (C) 2022-2025 Contributors | TIM S.p.A. to CAMARA a Series of LF Projects, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package trafficvolume

import (
	"errors"
)

var ErrPermanentFailure = errors.New("permanent failure")

type ThrottlingError struct {
	msg string
}

func (e *ThrottlingError) Error() string {
	return e.msg
}

func NewThrottlingError(msg string) error {
	return &ThrottlingError{msg: msg}
}

func IsThrottlingError(err error) bool {
	if err == nil {
		return false
	}

	var throttlingErr *ThrottlingError
	return errors.As(err, &throttlingErr)
}
