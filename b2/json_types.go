//####################################################################
//
// File: b2/json_types.go
//
// Copyright 2025 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

type AllowedSchema struct {
	BucketId     string   `json:"bucket_id"`
	BucketName   string   `json:"bucket_name"`
	Capabilities []string `json:"capabilities"`
	NamePrefix   string   `json:"name_prefix"`
}

type AccountInfoSchema struct {
	AccountId        string          `json:"account_id"`
	AccountAuthToken string          `json:"account_auth_token"`
	ApiUrl           string          `json:"api_url"`
	Allowed          []AllowedSchema `json:"allowed"`
	DownloadUrl      string          `json:"download_url"`
	S3ApiUrl         string          `json:"s3_api_url"`
}

type ApplicationKeySchema struct {
	ApplicationKeyId string   `json:"application_key_id"`
	ApplicationKey   string   `json:"application_key"`
	BucketId         string   `json:"bucket_id"`
	Capabilities     []string `json:"capabilities"`
	KeyName          string   `json:"key_name"`
	NamePrefix       string   `json:"name_prefix"`
	Options          []string `json:"options"`
}

type CorsRuleSchema struct {
	CorsRuleName      string   `json:"cors_rule_name"`
	AllowedOrigins    []string `json:"allowed_origins"`
	AllowedOperations []string `json:"allowed_operations"`
	MaxAgeSeconds     int      `json:"max_age_seconds"`
	AllowedHeaders    []string `json:"allowed_headers"`
	ExposeHeaders     []string `json:"expose_headers"`
}

type RetentionPeriodSchema struct {
	Duration int    `json:"duration"`
	Unit     string `json:"unit"`
}

type DefaultRetentionSchema struct {
	Mode   string                  `json:"mode"`
	Period []RetentionPeriodSchema `json:"period"`
}

type FileLockConfigurationSchema struct {
	IsFileLockEnabled bool                     `json:"is_file_lock_enabled"`
	DefaultRetention  []DefaultRetentionSchema `json:"default_retention"`
}

type ServerSideEncryptionSchema struct {
	Mode      string `json:"mode"`
	Algorithm string `json:"algorithm"`
}

type LifecycleRuleSchema struct {
	FileNamePrefix                                  string `json:"file_name_prefix"`
	DaysFromHidingToDeleting                        int    `json:"days_from_hiding_to_deleting"`
	DaysFromUploadingToHiding                       int    `json:"days_from_uploading_to_hiding"`
	DaysFromStartingToCancelingUnfinishedLargeFiles int    `json:"days_from_starting_to_canceling_unfinished_large_files"`
}

type BucketSchema struct {
	AccountId                   string                        `json:"account_id"`
	BucketId                    string                        `json:"bucket_id"`
	BucketInfo                  map[string]string             `json:"bucket_info"`
	BucketName                  string                        `json:"bucket_name"`
	BucketType                  string                        `json:"bucket_type"`
	CorsRules                   []CorsRuleSchema              `json:"cors_rules"`
	DefaultServerSideEncryption []ServerSideEncryptionSchema  `json:"default_server_side_encryption"`
	FileLockConfiguration       []FileLockConfigurationSchema `json:"file_lock_configuration"`
	LifecycleRules              []LifecycleRuleSchema         `json:"lifecycle_rules"`
	Options                     []string                      `json:"options"`
	Revision                    int                           `json:"revision"`
}

type FileVersionSchema struct {
	Action               string                       `json:"action"`
	BucketId             string                       `json:"bucket_id"`
	ContentMd5           string                       `json:"content_md5"`
	ContentSha1          string                       `json:"content_sha1"`
	ContentType          string                       `json:"content_type"`
	FileId               string                       `json:"file_id"`
	FileInfo             map[string]string            `json:"file_info"`
	FileName             string                       `json:"file_name"`
	ServerSideEncryption []ServerSideEncryptionSchema `json:"server_side_encryption"`
	Size                 int                          `json:"size"`
	UploadTimestamp      int                          `json:"upload_timestamp"`
}

type BucketFileSchema struct {
	Sha1         string              `json:"_sha1"`
	BucketId     string              `json:"bucket_id"`
	FileName     string              `json:"file_name"`
	ShowVersions bool                `json:"show_versions"`
	FileVersions []FileVersionSchema `json:"file_versions"`
}

type BucketFilesSchema struct {
	Sha1         string              `json:"_sha1"`
	BucketId     string              `json:"bucket_id"`
	FolderName   string              `json:"folder_name"`
	Recursive    bool                `json:"recursive"`
	ShowVersions bool                `json:"show_versions"`
	FileVersions []FileVersionSchema `json:"file_versions"`
}

type BucketFileSignedUrlSchema struct {
	BucketId  string `json:"bucket_id"`
	Duration  int    `json:"duration"`
	FileName  string `json:"file_name"`
	SignedUrl string `json:"signed_url"`
}

type CustomHeaderSchema struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type TargetConfigurationSchema struct {
	CustomHeaders           []CustomHeaderSchema `json:"custom_headers"`
	HmacSha256SigningSecret string               `json:"hmac_sha256_signing_secret"`
	TargetType              string               `json:"target_type"`
	Url                     string               `json:"url"`
}

type NotificationRuleSchema struct {
	EventTypes          []string                    `json:"event_types"`
	IsEnabled           bool                        `json:"is_enabled"`
	IsSuspended         bool                        `json:"is_suspended"`
	Name                string                      `json:"name"`
	ObjectNamePrefix    string                      `json:"object_name_prefix"`
	SuspensionReason    string                      `json:"suspension_reason"`
	TargetConfiguration []TargetConfigurationSchema `json:"target_configuration"`
}

type BucketNotificationRulesSchema struct {
	BucketId          string                   `json:"bucket_id"`
	NotificationRules []NotificationRuleSchema `json:"notification_rules"`
}

type ResourceFileEncryptionKeySchema struct {
	KeyId     string `json:"key_id"`
	SecretB64 string `json:"secret_b64"`
}

type ResourceFileEncryptionSchema struct {
	Algorithm string                            `json:"algorithm"`
	Key       []ResourceFileEncryptionKeySchema `json:"key"`
	Mode      string                            `json:"mode"`
}

type BucketFileVersionSchema struct {
	Action               string                         `json:"action"`
	BucketId             string                         `json:"bucket_id"`
	ContentMd5           string                         `json:"content_md5"`
	ContentSha1          string                         `json:"content_sha1"`
	ContentType          string                         `json:"content_type"`
	FileId               string                         `json:"file_id"`
	FileInfo             map[string]string              `json:"file_info"`
	FileName             string                         `json:"file_name"`
	ServerSideEncryption []ResourceFileEncryptionSchema `json:"server_side_encryption"`
	Size                 int                            `json:"size"`
	Source               string                         `json:"source"`
	UploadTimestamp      int                            `json:"upload_timestamp"`
}
