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

type PasswordPolicyConditions struct {
	App                   *AppAndInstancePolicyRuleCondition             `json:"app,omitempty"`
	Apps                  *AppInstancePolicyRuleCondition                `json:"apps,omitempty"`
	AuthContext           *PolicyRuleAuthContextCondition                `json:"authContext,omitempty"`
	AuthProvider          *PasswordPolicyAuthenticationProviderCondition `json:"authProvider,omitempty"`
	BeforeScheduledAction *BeforeScheduledActionPolicyRuleCondition      `json:"beforeScheduledAction,omitempty"`
	Clients               *ClientPolicyCondition                         `json:"clients,omitempty"`
	Context               *ContextPolicyRuleCondition                    `json:"context,omitempty"`
	Device                *DevicePolicyRuleCondition                     `json:"device,omitempty"`
	GrantTypes            *GrantTypePolicyRuleCondition                  `json:"grantTypes,omitempty"`
	Groups                *GroupPolicyRuleCondition                      `json:"groups,omitempty"`
	IdentityProvider      *IdentityProviderPolicyRuleCondition           `json:"identityProvider,omitempty"`
	MdmEnrollment         *MDMEnrollmentPolicyRuleCondition              `json:"mdmEnrollment,omitempty"`
	Network               *PolicyNetworkCondition                        `json:"network,omitempty"`
	People                *PolicyPeopleCondition                         `json:"people,omitempty"`
	Platform              *PlatformPolicyRuleCondition                   `json:"platform,omitempty"`
	Risk                  *RiskPolicyRuleCondition                       `json:"risk,omitempty"`
	RiskScore             *RiskScorePolicyRuleCondition                  `json:"riskScore,omitempty"`
	Scopes                *OAuth2ScopesMediationPolicyRuleCondition      `json:"scopes,omitempty"`
	UserIdentifier        *UserIdentifierPolicyRuleCondition             `json:"userIdentifier,omitempty"`
	UserStatus            *UserStatusPolicyRuleCondition                 `json:"userStatus,omitempty"`
	Users                 *UserPolicyRuleCondition                       `json:"users,omitempty"`
}

func NewPasswordPolicyConditions() *PasswordPolicyConditions {
	return &PasswordPolicyConditions{}
}

func (a *PasswordPolicyConditions) IsPolicyInstance() bool {
	return true
}
