/*
 *  Copyright (c) 2024, WSO2 LLC. (http://www.wso2.org) All Rights Reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package eventhub

import (
	"fmt"

	loggers "github.com/wso2/product-apim-tooling/apim-apk-agent/internal/loggers"
	"github.com/wso2/product-apim-tooling/apim-apk-agent/pkg/eventhub/types"
	eventhubTypes "github.com/wso2/product-apim-tooling/apim-apk-agent/pkg/eventhub/types"
	"github.com/wso2/product-apim-tooling/apim-apk-agent/pkg/managementserver"
	"github.com/wso2/product-apim-tooling/apim-apk-agent/pkg/transformer"
	"github.com/wso2/product-apim-tooling/apim-apk-agent/pkg/utils"
)

// SubscriptionList for struct list of applications
type SubscriptionList struct {
	List []Subscription `json:"list"`
}

// Application for struct application
type Application struct {
	UUID         string            `json:"uuid"`
	ID           int32             `json:"id" json:"applicationId"`
	Name         string            `json:"name" json:"applicationName"`
	SubName      string            `json:"subName" json:"subscriber"`
	Policy       string            `json:"policy" json:"applicationPolicy"`
	TokenType    string            `json:"tokenType"`
	Attributes   map[string]string `json:"attributes"`
	TenantID     int32             `json:"tenanId,omitempty"`
	TenantDomain string            `json:"tenanDomain,omitempty"`
	TimeStamp    int64             `json:"timeStamp,omitempty"`
}

// ApplicationList for struct list of application
type ApplicationList struct {
	List []Application `json:"list"`
}

// ApplicationKeyMapping for struct applicationKeyMapping
type ApplicationKeyMapping struct {
	ApplicationID   int32  `json:"applicationId"`
	ApplicationUUID string `json:"applicationUUID"`
	ConsumerKey     string `json:"consumerKey"`
	KeyType         string `json:"keyType"`
	KeyManager      string `json:"keyManager"`
	TenantID        int32  `json:"tenanId,omitempty"`
	TenantDomain    string `json:"tenanDomain,omitempty"`
	TimeStamp       int64  `json:"timeStamp,omitempty"`
}

// ApplicationKeyMappingList for struct list of applicationKeyMapping
type ApplicationKeyMappingList struct {
	List []ApplicationKeyMapping `json:"list"`
}

// Subscription for struct subscription
type Subscription struct {
	SubscriptionID    int32  `json:"subscriptionId"`
	SubscriptionUUID  string `json:"subscriptionUUID"`
	PolicyID          string `json:"policyId"`
	APIID             int32  `json:"apiId"`
	APIUUID           string `json:"apiUUID"`
	AppID             int32  `json:"appId" json:"applicationId"`
	ApplicationUUID   string `json:"applicationUUID"`
	SubscriptionState string `json:"subscriptionState"`
	TenantID          int32  `json:"tenanId,omitempty"`
	TenantDomain      string `json:"tenanDomain,omitempty"`
	TimeStamp         int64  `json:"timeStamp,omitempty"`
}

// KeyManager for struct keyManager
type KeyManager struct {
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Issuer      string `json:"issuer"`
	Certificate string `json:"certificate"`
}

// MarshalKeyManagers is used to update the key managers during the startup where
// multiple key managers are pulled at once. And then it returns the KeyManagerMap.
func MarshalKeyManagers(keyManagersList *[]eventhubTypes.KeyManager) []eventhubTypes.ResolvedKeyManager {
	loggers.LoggerEventhub.Infof("Starting to marshal %d key managers", len(*keyManagersList))
	resourceMap := make([]eventhubTypes.ResolvedKeyManager, 0)
	for i, keyManager := range *keyManagersList {
		loggers.LoggerEventhub.Debugf("Marshalling key manager %d: %s (UUID: %s)", i+1, keyManager.Name, keyManager.UUID)
		keyManagerSub := MarshalKeyManager(&keyManager)
		resourceMap = append(resourceMap, keyManagerSub)
	}
	loggers.LoggerEventhub.Infof("Successfully marshalled %d key managers", len(resourceMap))
	return resourceMap
}

// MarshalMultipleApplications is used to update the applicationList during the startup where
func MarshalMultipleApplications(appList *types.ApplicationList) {
	loggers.LoggerEventhub.Infof("Starting to marshal %d applications", len(appList.List))
	applicationMap := make(map[string]managementserver.Application)
	for i, application := range appList.List {
		loggers.LoggerEventhub.Debugf("Marshalling application %d: %s (UUID: %s, Organization: %s)", i+1, application.Name, application.UUID, application.Organization)
		applicationSub := MarshalApplication(&application)
		applicationMap[applicationSub.UUID] = applicationSub
	}
	loggers.LoggerEventhub.Infof("Successfully marshalled %d applications. Adding all to management server", len(applicationMap))
	managementserver.AddAllApplications(applicationMap)
}

// MarshalMultipleApplicationKeyMappings is used to update the application key mappings during the startup where
// multiple key mappings are pulled at once. And then it returns the ApplicationKeyMappingList.
func MarshalMultipleApplicationKeyMappings(keymappingList *types.ApplicationKeyMappingList) {
	loggers.LoggerEventhub.Infof("Starting to marshal %d application key mappings", len(keymappingList.List))
	resourceMap := make(map[string]managementserver.ApplicationKeyMapping)
	for i, keyMapping := range keymappingList.List {
		applicationKeyMappingReference := GetApplicationKeyMappingReference(&keyMapping)
		loggers.LoggerEventhub.Debugf("Marshalling key mapping %d: %s (Consumer Key: %s, Key Manager: %s)", i+1, applicationKeyMappingReference, keyMapping.ConsumerKey, keyMapping.KeyManager)
		keyMappingSub := marshalKeyMapping(&keyMapping)
		resourceMap[applicationKeyMappingReference] = keyMappingSub
	}
	loggers.LoggerEventhub.Infof("Successfully marshalled %d application key mappings. Adding all to management server", len(resourceMap))
	managementserver.AddAllApplicationKeyMappings(resourceMap)
}

// MarshalMultipleSubscriptions is used to update the subscriptions during the startup where
// multiple subscriptions are pulled at once. And then it returns the SubscriptionList.
func MarshalMultipleSubscriptions(subscriptionsList *types.SubscriptionList) {
	loggers.LoggerEventhub.Infof("Starting to marshal %d subscriptions", len(subscriptionsList.List))
	subscriptionMap := make(map[string]managementserver.Subscription)
	applicationMappingMap := make(map[string]managementserver.ApplicationMapping)
	for i, subscription := range subscriptionsList.List {
		loggers.LoggerEventhub.Debugf("Marshalling subscription %d: UUID: %s, API: %s v%s, Organization: %s", i+1, subscription.SubscriptionUUID, subscription.APIName, subscription.APIVersion, subscription.ApplicationOrganization)
		subscriptionSub := MarshalSubscription(&subscription)
		subscriptionMap[subscriptionSub.UUID] = subscriptionSub
		appMappingID := utils.GetUniqueIDOfApplicationMapping(subscription.ApplicationUUID, subscription.SubscriptionUUID)
		loggers.LoggerEventhub.Debugf("Creating application mapping for subscription: AppMappingID: %s", appMappingID)
		applicationMappingMap[subscriptionSub.UUID] = managementserver.ApplicationMapping{
			UUID:            appMappingID,
			ApplicationRef:  subscription.ApplicationUUID,
			SubscriptionRef: subscription.SubscriptionUUID,
			Organization:    subscriptionSub.Organization,
		}
	}
	loggers.LoggerEventhub.Infof("Successfully marshalled %d subscriptions and %d application mappings. Adding to management server", len(subscriptionMap), len(applicationMappingMap))
	managementserver.AddAllApplicationMappings(applicationMappingMap)
	managementserver.AddAllSubscriptions(subscriptionMap)

}

// MarshalSubscription is used to map to internal Subscription struct
func MarshalSubscription(subscriptionInternal *types.Subscription) managementserver.Subscription {
	// Compute RateLimit value the same way as in notification events
	// PolicyID should be combined with tenant domain and hashed
	policyIdentifier := subscriptionInternal.PolicyID
	var rateLimitValue string
	if policyIdentifier != "" {
		rateLimitValue = transformer.GetSha1Value(fmt.Sprintf("%s-%s", policyIdentifier, subscriptionInternal.ApplicationOrganization))
		loggers.LoggerEventhub.Debugf("Computed RateLimit value for subscription %s: '%s' (from policy identifier: '%s', organization: '%s')",
			subscriptionInternal.SubscriptionUUID, rateLimitValue, policyIdentifier, subscriptionInternal.ApplicationOrganization)
	} else {
		loggers.LoggerEventhub.Warnf("PolicyID is empty for subscription %s, RateLimit will be empty",
			subscriptionInternal.SubscriptionUUID)
		rateLimitValue = ""
	}

	sub := managementserver.Subscription{
		SubStatus:     subscriptionInternal.SubscriptionState,
		UUID:          subscriptionInternal.SubscriptionUUID,
		Organization:  subscriptionInternal.ApplicationOrganization,
		SubscribedAPI: &managementserver.SubscribedAPI{Name: subscriptionInternal.APIName, Version: subscriptionInternal.APIVersion},
		RateLimit:     rateLimitValue,
		PolicyName:    rateLimitValue, // Use the hash value as policy name - this matches the RateLimitPolicy CR name
		TimeStamp:     subscriptionInternal.TimeStamp,
	}
	loggers.LoggerEventhub.Infof("Marshalled subscription %s: RateLimit='%s', PolicyName='%s'",
		subscriptionInternal.SubscriptionUUID, rateLimitValue, rateLimitValue)
	return sub
}

// MarshalApplication is used to map to internal Application struct
func MarshalApplication(appInternal *types.Application) managementserver.Application {
	loggers.LoggerEventhub.Debugf("Mapping application: Name: %s, UUID: %s, Owner: %s, Organization: %s, Attributes count: %d",
		appInternal.Name, appInternal.UUID, appInternal.SubName, appInternal.Organization, len(appInternal.Attributes))
	app := managementserver.Application{
		UUID:         appInternal.UUID,
		Name:         appInternal.Name,
		Owner:        appInternal.SubName,
		Organization: appInternal.Organization,
		Attributes:   appInternal.Attributes,
		TimeStamp:    appInternal.TimeStamp,
	}
	loggers.LoggerEventhub.Debugf("Application mapping completed for: %s", appInternal.UUID)
	return app
}

func marshalKeyMapping(keyMappingInternal *types.ApplicationKeyMapping) managementserver.ApplicationKeyMapping {
	loggers.LoggerEventhub.Debugf("Mapping application key: ApplicationUUID: %s, ConsumerKey: %s, KeyType: %s, KeyManager: %s",
		keyMappingInternal.ApplicationUUID, keyMappingInternal.ConsumerKey, keyMappingInternal.KeyType, keyMappingInternal.KeyManager)
	return managementserver.ApplicationKeyMapping{
		ApplicationUUID:       keyMappingInternal.ApplicationUUID,
		ApplicationIdentifier: keyMappingInternal.ConsumerKey,
		KeyType:               keyMappingInternal.KeyType,
		SecurityScheme:        "OAuth2",
		EnvID:                 "Default",
		Timestamp:             keyMappingInternal.TimeStamp,
	}
}
func marshalKeyManagrConfig(configuration map[string]interface{}) eventhubTypes.KeyManagerConfig {
	loggers.LoggerEventhub.Debugf("Starting to marshal key manager configuration with %d properties", len(configuration))
	marshalledConfiguration := eventhubTypes.KeyManagerConfig{}
	if configuration["token_format_string"] != nil {
		marshalledConfiguration.TokenFormatString = configuration["token_format_string"].(string)
		loggers.LoggerEventhub.Debugf("Set TokenFormatString: %s", marshalledConfiguration.TokenFormatString)
	}
	if configuration["issuer"] != nil {
		marshalledConfiguration.Issuer = configuration["issuer"].(string)
		loggers.LoggerEventhub.Debugf("Set Issuer: %s", marshalledConfiguration.Issuer)
	}
	if configuration["ServerURL"] != nil {
		marshalledConfiguration.ServerURL = configuration["ServerURL"].(string)
		loggers.LoggerEventhub.Debugf("Set ServerURL: %s", marshalledConfiguration.ServerURL)
	}
	if configuration["validation_enable"] != nil {
		marshalledConfiguration.ValidationEnable = configuration["validation_enable"].(bool)
		loggers.LoggerEventhub.Debugf("Set ValidationEnable: %v", marshalledConfiguration.ValidationEnable)
	}
	if configuration["claim_mappings"] != nil {
		marshalledConfiguration.ClaimMappings = marshalClaimMappings(configuration["claim_mappings"].([]interface{}))
		loggers.LoggerEventhub.Debugf("Marshalled %d claim mappings", len(marshalledConfiguration.ClaimMappings))
	}
	if configuration["grant_types"] != nil {
		marshalledConfiguration.GrantTypes = marshalGrantTypes(configuration["grant_types"].([]interface{}))
		loggers.LoggerEventhub.Debugf("Marshalled %d grant types", len(marshalledConfiguration.GrantTypes))
	}
	if configuration["OAuthConfigurations.EncryptPersistedTokens"] != nil {
		marshalledConfiguration.EncryptPersistedTokens = configuration["OAuthConfigurations.EncryptPersistedTokens"].(bool)
		loggers.LoggerEventhub.Debugf("Set EncryptPersistedTokens: %v", marshalledConfiguration.EncryptPersistedTokens)
	}
	if configuration["enable_oauth_app_creation"] != nil {
		marshalledConfiguration.EnableOauthAppCreation = configuration["enable_oauth_app_creation"].(bool)
		loggers.LoggerEventhub.Debugf("Set EnableOauthAppCreation: %v", marshalledConfiguration.EnableOauthAppCreation)
	}
	if configuration["VALIDITY_PERIOD"] != nil {
		marshalledConfiguration.ValidityPeriod = configuration["VALIDITY_PERIOD"].(string)
		loggers.LoggerEventhub.Debugf("Set ValidityPeriod: %s", marshalledConfiguration.ValidityPeriod)
	}
	if configuration["enable_token_generation"] != nil {
		marshalledConfiguration.EnableTokenGeneration = configuration["enable_token_generation"].(bool)
		loggers.LoggerEventhub.Debugf("Set EnableTokenGeneration: %v", marshalledConfiguration.EnableTokenGeneration)
	}
	if configuration["enable_map_oauth_consumer_apps"] != nil {
		marshalledConfiguration.EnableMapOauthConsumerApps = configuration["enable_map_oauth_consumer_apps"].(bool)
		loggers.LoggerEventhub.Debugf("Set EnableMapOauthConsumerApps: %v", marshalledConfiguration.EnableMapOauthConsumerApps)
	}
	if configuration["enable_token_hash"] != nil {
		marshalledConfiguration.EnableTokenHash = configuration["enable_token_hash"].(bool)
		loggers.LoggerEventhub.Debugf("Set EnableTokenHash: %v", marshalledConfiguration.EnableTokenHash)
	}
	if configuration["self_validate_jwt"] != nil {
		marshalledConfiguration.SelfValidateJwt = configuration["self_validate_jwt"].(bool)
		loggers.LoggerEventhub.Debugf("Set SelfValidateJwt: %v", marshalledConfiguration.SelfValidateJwt)
	}
	if configuration["revoke_endpoint"] != nil {
		marshalledConfiguration.RevokeEndpoint = configuration["revoke_endpoint"].(string)
		loggers.LoggerEventhub.Debugf("Set RevokeEndpoint: %s", marshalledConfiguration.RevokeEndpoint)
	}
	if configuration["enable_token_encryption"] != nil {
		marshalledConfiguration.EnableTokenEncryption = configuration["enable_token_encryption"].(bool)
		loggers.LoggerEventhub.Debugf("Set EnableTokenEncryption: %v", marshalledConfiguration.EnableTokenEncryption)
	}
	if configuration["RevokeURL"] != nil {
		marshalledConfiguration.RevokeURL = configuration["RevokeURL"].(string)
		loggers.LoggerEventhub.Debugf("Set RevokeURL: %s", marshalledConfiguration.RevokeURL)
	}
	if configuration["token_endpoint"] != nil {
		marshalledConfiguration.TokenURL = configuration["token_endpoint"].(string)
		loggers.LoggerEventhub.Debugf("Set TokenURL: %s", marshalledConfiguration.TokenURL)
	}
	if configuration["certificate_type"] != nil {
		marshalledConfiguration.CertificateType = configuration["certificate_type"].(string)
		loggers.LoggerEventhub.Debugf("Set CertificateType: %s", marshalledConfiguration.CertificateType)
	}
	if configuration["certificate_value"] != nil {
		marshalledConfiguration.CertificateValue = configuration["certificate_value"].(string)
		loggers.LoggerEventhub.Debugf("Set CertificateValue (length: %d)", len(marshalledConfiguration.CertificateValue))
	}
	if configuration["consumer_key_claim"] != nil {
		marshalledConfiguration.ConsumerKeyClaim = configuration["consumer_key_claim"].(string)
		loggers.LoggerEventhub.Debugf("Set ConsumerKeyClaim: %s", marshalledConfiguration.ConsumerKeyClaim)
	}
	if configuration["scopes_claim"] != nil {
		marshalledConfiguration.ScopesClaim = configuration["scopes_claim"].(string)
		loggers.LoggerEventhub.Debugf("Set ScopesClaim: %s", marshalledConfiguration.ScopesClaim)
	}
	loggers.LoggerEventhub.Debugf("Key manager configuration marshalling completed")
	return marshalledConfiguration
}
func marshalGrantTypes(grantTypes []interface{}) []string {
	loggers.LoggerEventhub.Debugf("Starting to marshal %d grant types", len(grantTypes))
	resolvedGrantTypes := make([]string, 0)
	for i, grantType := range grantTypes {
		if resolvedGrantType, ok := grantType.(string); ok {
			resolvedGrantTypes = append(resolvedGrantTypes, resolvedGrantType)
			loggers.LoggerEventhub.Debugf("Grant type %d: %s", i+1, resolvedGrantType)
		} else {
			loggers.LoggerEventhub.Warnf("Failed to cast grant type at index %d to string", i)
		}
	}
	loggers.LoggerEventhub.Debugf("Successfully marshalled %d grant types", len(resolvedGrantTypes))
	return resolvedGrantTypes

}
func marshalClaimMappings(claimMappings []interface{}) []eventhubTypes.Claim {
	loggers.LoggerEventhub.Debugf("Starting to marshal %d claim mappings", len(claimMappings))
	resolvedClaimMappings := make([]eventhubTypes.Claim, 0)
	for i, claim := range claimMappings {
		if resolvedClaim, ok := claim.(eventhubTypes.Claim); ok {
			resolvedClaimMappings = append(resolvedClaimMappings, resolvedClaim)
			loggers.LoggerEventhub.Debugf("Claim mapping %d marshalled successfully", i+1)
		} else {
			loggers.LoggerEventhub.Warnf("Failed to cast claim mapping at index %d to Claim type", i)
		}
	}
	loggers.LoggerEventhub.Debugf("Successfully marshalled %d claim mappings", len(resolvedClaimMappings))
	return resolvedClaimMappings
}

// MarshalKeyManager is used to map Internal key manager
func MarshalKeyManager(keyManagerInternal *types.KeyManager) eventhubTypes.ResolvedKeyManager {
	loggers.LoggerEventhub.Debugf("Mapping key manager: Name: %s, UUID: %s, Type: %s, Organization: %s, Enabled: %v",
		keyManagerInternal.Name, keyManagerInternal.UUID, keyManagerInternal.Type, keyManagerInternal.Organization, keyManagerInternal.Enabled)
	resolvedKeyManager := eventhubTypes.ResolvedKeyManager{
		UUID:             keyManagerInternal.UUID,
		Name:             keyManagerInternal.Name,
		Enabled:          keyManagerInternal.Enabled,
		Type:             keyManagerInternal.Type,
		Organization:     keyManagerInternal.Organization,
		TokenType:        keyManagerInternal.TokenType,
		KeyManagerConfig: marshalKeyManagrConfig(keyManagerInternal.Configuration),
	}
	loggers.LoggerEventhub.Debugf("Key manager mapping completed for: %s", keyManagerInternal.UUID)
	return resolvedKeyManager
}

// GetApplicationKeyMappingReference returns unique reference for each key Mapping event.
// It is the combination of consumerKey:keyManager
func GetApplicationKeyMappingReference(keyMapping *types.ApplicationKeyMapping) string {
	reference := keyMapping.ConsumerKey + ":" + keyMapping.KeyManager
	loggers.LoggerEventhub.Debugf("Generated application key mapping reference: %s", reference)
	return reference
}

// CheckIfAPIMetadataIsAlreadyAvailable returns true only if the API Metadata for the given API UUID
// is already available
// func CheckIfAPIMetadataIsAlreadyAvailable(apiUUID, label string) bool {
// 	if _, labelAvailable := APIListMap[label]; labelAvailable {
// 		if _, apiAvailale := APIListMap[label][apiUUID]; apiAvailale {
// 			return true
// 		}
// 	}
// 	return false
// }
