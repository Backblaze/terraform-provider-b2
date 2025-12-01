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
	Buckets      []AllowedBucketsSchema `json:"buckets"`
	BucketId     string                 `json:"bucketId"`   // deprecated
	BucketName   string                 `json:"bucketName"` // deprecated
	Capabilities []string               `json:"capabilities"`
	NamePrefix   string                 `json:"namePrefix"`
}

type AllowedBucketsSchema struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type AccountInfoSchema struct {
	AccountId        string          `json:"accountId"`
	AccountAuthToken string          `json:"accountAuthToken"`
	ApiUrl           string          `json:"apiUrl"`
	Allowed          []AllowedSchema `json:"allowed"`
	DownloadUrl      string          `json:"downloadUrl"`
	S3ApiUrl         string          `json:"s3ApiUrl"`
}

type ApplicationKeySchema struct {
	ApplicationKeyId string   `json:"applicationKeyId"`
	ApplicationKey   string   `json:"applicationKey"`
	BucketIds        []string `json:"bucketIds"`
	BucketId         string   `json:"bucketId"` // deprecated
	Capabilities     []string `json:"capabilities"`
	KeyName          string   `json:"keyName"`
	NamePrefix       string   `json:"namePrefix"`
	Options          []string `json:"options"`
}

type CorsRuleSchema struct {
	CorsRuleName      string   `json:"corsRuleName"`
	AllowedOrigins    []string `json:"allowedOrigins"`
	AllowedOperations []string `json:"allowedOperations"`
	MaxAgeSeconds     int      `json:"maxAgeSeconds"`
	AllowedHeaders    []string `json:"allowedHeaders"`
	ExposeHeaders     []string `json:"exposeHeaders"`
}

type RetentionPeriodSchema struct {
	Duration int    `json:"duration"`
	Unit     string `json:"unit"`
}

type DefaultRetentionSchema struct {
	Mode   string                 `json:"mode"`
	Period *RetentionPeriodSchema `json:"period"`
}

type FileLockConfigurationSchema struct {
	IsFileLockEnabled bool                    `json:"isFileLockEnabled"`
	DefaultRetention  *DefaultRetentionSchema `json:"defaultRetention"`
}

type ServerSideEncryptionSchema struct {
	Mode      string `json:"mode"`
	Algorithm string `json:"algorithm"`
}

type LifecycleRuleSchema struct {
	FileNamePrefix                                  string `json:"fileNamePrefix"`
	DaysFromHidingToDeleting                        int    `json:"daysFromHidingToDeleting"`
	DaysFromUploadingToHiding                       int    `json:"daysFromUploadingToHiding"`
	DaysFromStartingToCancelingUnfinishedLargeFiles int    `json:"daysFromStartingToCancelingUnfinishedLargeFiles"`
}

type BucketSchema struct {
	AccountId                   string                       `json:"accountId"`
	BucketId                    string                       `json:"bucketId"`
	BucketInfo                  map[string]string            `json:"bucketInfo"`
	BucketName                  string                       `json:"bucketName"`
	BucketType                  string                       `json:"bucketType"`
	CorsRules                   []CorsRuleSchema             `json:"corsRules"`
	DefaultServerSideEncryption *ServerSideEncryptionSchema  `json:"defaultServerSideEncryption"`
	FileLockConfiguration       *FileLockConfigurationSchema `json:"fileLockConfiguration"`
	LifecycleRules              []LifecycleRuleSchema        `json:"lifecycleRules"`
	Options                     []string                     `json:"options"`
	Revision                    int                          `json:"revision"`
}

type FileVersionSchema struct {
	Action               string                      `json:"action"`
	BucketId             string                      `json:"bucketId"`
	ContentMd5           string                      `json:"contentMd5"`
	ContentSha1          string                      `json:"contentSha1"`
	ContentType          string                      `json:"contentType"`
	FileId               string                      `json:"fileId"`
	FileInfo             map[string]string           `json:"fileInfo"`
	FileName             string                      `json:"fileName"`
	ServerSideEncryption *ServerSideEncryptionSchema `json:"serverSideEncryption"`
	Size                 int                         `json:"size"`
	UploadTimestamp      int                         `json:"uploadTimestamp"`
}

type BucketFileSchema struct {
	Sha1         string              `json:"_sha1"`
	BucketId     string              `json:"bucketId"`
	FileName     string              `json:"fileName"`
	ShowVersions bool                `json:"showVersions"`
	FileVersions []FileVersionSchema `json:"fileVersions"`
}

type BucketFilesSchema struct {
	Sha1         string              `json:"_sha1"`
	BucketId     string              `json:"bucketId"`
	FolderName   string              `json:"folderName"`
	Recursive    bool                `json:"recursive"`
	ShowVersions bool                `json:"showVersions"`
	FileVersions []FileVersionSchema `json:"fileVersions"`
}

type BucketFileSignedUrlSchema struct {
	BucketId  string `json:"bucketId"`
	Duration  int    `json:"duration"`
	FileName  string `json:"fileName"`
	SignedUrl string `json:"signedUrl"`
}

type CustomHeaderSchema struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type TargetConfigurationSchema struct {
	CustomHeaders           []CustomHeaderSchema `json:"customHeaders"`
	HmacSha256SigningSecret string               `json:"hmacSha256SigningSecret"`
	TargetType              string               `json:"targetType"`
	Url                     string               `json:"url"`
}

type NotificationRuleSchema struct {
	EventTypes          []string                   `json:"eventTypes"`
	IsEnabled           bool                       `json:"isEnabled"`
	IsSuspended         bool                       `json:"isSuspended"`
	Name                string                     `json:"name"`
	ObjectNamePrefix    string                     `json:"objectNamePrefix"`
	SuspensionReason    string                     `json:"suspensionReason"`
	TargetConfiguration *TargetConfigurationSchema `json:"targetConfiguration"`
}

type BucketNotificationRulesSchema struct {
	BucketId          string                   `json:"bucketId"`
	NotificationRules []NotificationRuleSchema `json:"notificationRules"`
}

type ResourceFileEncryptionKeySchema struct {
	KeyId     string `json:"keyId"`
	SecretB64 string `json:"secretB64"`
}

type ResourceFileEncryptionSchema struct {
	Algorithm string                            `json:"algorithm"`
	Key       []ResourceFileEncryptionKeySchema `json:"key"`
	Mode      string                            `json:"mode"`
}

type BucketFileVersionSchema struct {
	Action               string                        `json:"action"`
	BucketId             string                        `json:"bucketId"`
	ContentMd5           string                        `json:"contentMd5"`
	ContentSha1          string                        `json:"contentSha1"`
	ContentType          string                        `json:"contentType"`
	FileId               string                        `json:"fileId"`
	FileInfo             map[string]string             `json:"fileInfo"`
	FileName             string                        `json:"fileName"`
	ServerSideEncryption *ResourceFileEncryptionSchema `json:"serverSideEncryption"`
	Size                 int                           `json:"size"`
	Source               string                        `json:"source"`
	UploadTimestamp      int                           `json:"uploadTimestamp"`
}
