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
package policy

import "context"

// AllowAll is a dummy policy implementation that unconditionally permits access.
// It satisfies the Interface and can be used in development environments where
// authorization is not enforced yet.
type AllowAll struct{}

// NewAllowAll returns a new AllowAll policy implementation.
func NewAllowAll() *AllowAll { return &AllowAll{} }

// HasAccessToApplicationIDs always returns nil (success) indicating the user
// has access to all provided application IDs.
func (a *AllowAll) HasAccessToApplicationIDs(ctx context.Context, user string, appIDs []string) error {
	return nil
}
