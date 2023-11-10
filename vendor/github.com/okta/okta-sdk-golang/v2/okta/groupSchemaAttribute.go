/*
* Copyright 2018 - Present Okta, Inc.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*      http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

// Code generated by okta openapi generator. DO NOT EDIT.

package okta

type GroupSchemaAttribute struct {
	Description       string                           `json:"description,omitempty"`
	Enum              []string                         `json:"enum,omitempty"`
	ExternalName      string                           `json:"externalName,omitempty"`
	ExternalNamespace string                           `json:"externalNamespace,omitempty"`
	Items             *UserSchemaAttributeItems        `json:"items,omitempty"`
	Master            *UserSchemaAttributeMaster       `json:"master,omitempty"`
	MaxLength         int64                            `json:"maxLength,omitempty"`
	MinLength         int64                            `json:"minLength,omitempty"`
	Mutability        string                           `json:"mutability,omitempty"`
	OneOf             []*UserSchemaAttributeEnum       `json:"oneOf,omitempty"`
	Permissions       []*UserSchemaAttributePermission `json:"permissions,omitempty"`
	Required          *bool                            `json:"required,omitempty"`
	Scope             string                           `json:"scope,omitempty"`
	Title             string                           `json:"title,omitempty"`
	Type              string                           `json:"type,omitempty"`
	Union             string                           `json:"union,omitempty"`
	Unique            string                           `json:"unique,omitempty"`
}
