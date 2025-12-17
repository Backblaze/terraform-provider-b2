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

// ResourceSchema is an interface implemented by all input and output structs
// to provide their resource name for type inference
type ResourceSchema interface {
	ResourceName() string
}

// Shared schemas

type Allowed struct {
	Buckets      []AllowedBuckets `json:"buckets"`
	BucketId     string           `json:"bucketId"`   // deprecated
	BucketName   string           `json:"bucketName"` // deprecated
	Capabilities []string         `json:"capabilities"`
	NamePrefix   string           `json:"namePrefix"`
}

type AllowedBuckets struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type CorsRule struct {
	CorsRuleName      string   `json:"corsRuleName"`
	AllowedOrigins    []string `json:"allowedOrigins"`
	AllowedOperations []string `json:"allowedOperations"`
	MaxAgeSeconds     int      `json:"maxAgeSeconds"`
	AllowedHeaders    []string `json:"allowedHeaders"`
	ExposeHeaders     []string `json:"exposeHeaders"`
}

type RetentionPeriod struct {
	Duration int    `json:"duration"`
	Unit     string `json:"unit"`
}

type DefaultRetention struct {
	Mode   string           `json:"mode"`
	Period *RetentionPeriod `json:"period"`
}

type FileLockConfiguration struct {
	IsFileLockEnabled bool              `json:"isFileLockEnabled"`
	DefaultRetention  *DefaultRetention `json:"defaultRetention"`
}

type ServerSideEncryption struct {
	Mode      string `json:"mode"`
	Algorithm string `json:"algorithm"`
}

type LifecycleRule struct {
	FileNamePrefix                                  string `json:"fileNamePrefix"`
	DaysFromHidingToDeleting                        int    `json:"daysFromHidingToDeleting"`
	DaysFromUploadingToHiding                       int    `json:"daysFromUploadingToHiding"`
	DaysFromStartingToCancelingUnfinishedLargeFiles int    `json:"daysFromStartingToCancelingUnfinishedLargeFiles"`
}

type FileVersion struct {
	Action               string                `json:"action"`
	BucketId             string                `json:"bucketId"`
	ContentMd5           string                `json:"contentMd5"`
	ContentSha1          string                `json:"contentSha1"`
	ContentType          string                `json:"contentType"`
	FileId               string                `json:"fileId"`
	FileInfo             map[string]string     `json:"fileInfo"`
	FileName             string                `json:"fileName"`
	ServerSideEncryption *ServerSideEncryption `json:"serverSideEncryption"`
	Size                 int                   `json:"size"`
	UploadTimestamp      int                   `json:"uploadTimestamp"`
}

type CustomHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type TargetConfiguration struct {
	CustomHeaders           []CustomHeader `json:"customHeaders"`
	HmacSha256SigningSecret string         `json:"hmacSha256SigningSecret"`
	TargetType              string         `json:"targetType"`
	Url                     string         `json:"url"`
}

type NotificationRule struct {
	EventTypes          []string             `json:"eventTypes"`
	IsEnabled           bool                 `json:"isEnabled"`
	IsSuspended         bool                 `json:"isSuspended"`
	Name                string               `json:"name"`
	ObjectNamePrefix    string               `json:"objectNamePrefix"`
	SuspensionReason    string               `json:"suspensionReason"`
	TargetConfiguration *TargetConfiguration `json:"targetConfiguration"`
}

type ResourceFileEncryptionKey struct {
	KeyId     string `json:"keyId"`
	SecretB64 string `json:"secretB64"`
}

type ResourceFileEncryption struct {
	Algorithm string                      `json:"algorithm"`
	Key       []ResourceFileEncryptionKey `json:"key"`
	Mode      string                      `json:"mode"`
}

// AccountInfo

type AccountInfoOutput struct {
	AccountId               string    `json:"accountId"`
	AccountAuthToken        string    `json:"accountAuthToken"`
	ApiUrl                  string    `json:"apiUrl"`
	Allowed                 []Allowed `json:"allowed"`
	DownloadUrl             string    `json:"downloadUrl"`
	S3ApiUrl                string    `json:"s3ApiUrl"`
	RecommendedPartSize     int       `json:"recommendedPartSize"`
	AbsoluteMinimumPartSize int       `json:"absoluteMinimumPartSize"`
}

func (s *AccountInfoOutput) ResourceName() string {
	return "account_info"
}

type AccountInfoInput struct{}

func (s *AccountInfoInput) ResourceName() string {
	return "account_info"
}

// ApplicationKey

type ApplicationKeyOutput struct {
	ApplicationKeyId       string        `json:"applicationKeyId"`
	ApplicationKey         string        `json:"applicationKey"`
	BucketIds              []interface{} `json:"bucketIds"`
	BucketId               string        `json:"bucketId"` // deprecated
	Capabilities           []interface{} `json:"capabilities"`
	ExpirationTimestamp    int           `json:"expirationTimestamp"`
	KeyName                string        `json:"keyName"`
	NamePrefix             string        `json:"namePrefix"`
	Options                []interface{} `json:"options"`
	ValidDurationInSeconds int           `json:"validDurationInSeconds"`
}

func (s *ApplicationKeyOutput) ResourceName() string {
	return "application_key"
}

type ApplicationKeyInput struct {
	ApplicationKeyId       string        `json:"applicationKeyId,omitempty"`
	KeyName                string        `json:"keyName,omitempty"`
	Capabilities           []interface{} `json:"capabilities,omitempty"`
	NamePrefix             string        `json:"namePrefix,omitempty"`
	ValidDurationInSeconds int           `json:"validDurationInSeconds,omitempty"`
	BucketIds              []interface{} `json:"bucketIds,omitempty"`
	BucketId               string        `json:"bucketId,omitempty"` // deprecated
}

func (s *ApplicationKeyInput) ResourceName() string {
	return "application_key"
}

// Bucket

type BucketOutput struct {
	AccountId                   string                 `json:"accountId"`
	BucketId                    string                 `json:"bucketId"`
	BucketInfo                  map[string]string      `json:"bucketInfo"`
	BucketName                  string                 `json:"bucketName"`
	BucketType                  string                 `json:"bucketType"`
	CorsRules                   []CorsRule             `json:"corsRules"`
	DefaultServerSideEncryption *ServerSideEncryption  `json:"defaultServerSideEncryption"`
	FileLockConfiguration       *FileLockConfiguration `json:"fileLockConfiguration"`
	LifecycleRules              []LifecycleRule        `json:"lifecycleRules"`
	Options                     []string               `json:"options"`
	Revision                    int                    `json:"revision"`
}

func (s *BucketOutput) ResourceName() string {
	return "bucket"
}

type BucketInput struct {
	BucketId                    string                 `json:"bucketId,omitempty"`
	BucketName                  string                 `json:"bucketName,omitempty"`
	AccountId                   string                 `json:"accountId,omitempty"`
	BucketType                  string                 `json:"bucketType,omitempty"`
	BucketInfo                  map[string]interface{} `json:"bucketInfo,omitempty"`
	CorsRules                   []interface{}          `json:"corsRules,omitempty"`
	FileLockConfiguration       []interface{}          `json:"fileLockConfiguration,omitempty"`
	DefaultServerSideEncryption []interface{}          `json:"defaultServerSideEncryption,omitempty"`
	LifecycleRules              []interface{}          `json:"lifecycleRules,omitempty"`
}

func (s *BucketInput) ResourceName() string {
	return "bucket"
}

// BucketFile

type BucketFileInput struct {
	BucketId     string `json:"bucketId"`
	FileName     string `json:"fileName"`
	ShowVersions bool   `json:"showVersions"`
}

func (s *BucketFileInput) ResourceName() string {
	return "bucket_file"
}

type BucketFileOutput struct {
	BucketFileInput
	Sha1         string        `json:"_sha1"`
	FileVersions []FileVersion `json:"fileVersions"`
}

func (s *BucketFileOutput) ResourceName() string {
	return "bucket_file"
}

// BucketFiles

type BucketFilesInput struct {
	BucketId     string `json:"bucketId"`
	FolderName   string `json:"folderName"`
	ShowVersions bool   `json:"showVersions"`
	Recursive    bool   `json:"recursive"`
}

func (s *BucketFilesInput) ResourceName() string {
	return "bucket_files"
}

type BucketFilesOutput struct {
	BucketFilesInput
	Sha1         string        `json:"_sha1"`
	FileVersions []FileVersion `json:"fileVersions"`
}

func (s *BucketFilesOutput) ResourceName() string {
	return "bucket_files"
}

// BucketFileSignedUrl

type BucketFileSignedUrlInput struct {
	BucketId string `json:"bucketId"`
	FileName string `json:"fileName"`
	Duration int    `json:"duration"`
}

func (s *BucketFileSignedUrlInput) ResourceName() string {
	return "bucket_file_signed_url"
}

type BucketFileSignedUrlOutput struct {
	BucketFileSignedUrlInput
	SignedUrl string `json:"signedUrl"`
}

func (s *BucketFileSignedUrlOutput) ResourceName() string {
	return "bucket_file_signed_url"
}

// BucketFileVersion

type BucketFileVersionOutput struct {
	Action               string                  `json:"action"`
	BucketId             string                  `json:"bucketId"`
	ContentMd5           string                  `json:"contentMd5"`
	ContentSha1          string                  `json:"contentSha1"`
	ContentType          string                  `json:"contentType"`
	FileId               string                  `json:"fileId"`
	FileInfo             map[string]string       `json:"fileInfo"`
	FileName             string                  `json:"fileName"`
	ServerSideEncryption *ResourceFileEncryption `json:"serverSideEncryption"`
	Size                 int                     `json:"size"`
	Source               string                  `json:"source"`
	UploadTimestamp      int                     `json:"uploadTimestamp"`
}

func (s *BucketFileVersionOutput) ResourceName() string {
	return "bucket_file_version"
}

type BucketFileVersionInput struct {
	FileId               string                 `json:"fileId,omitempty"`
	BucketId             string                 `json:"bucketId,omitempty"`
	FileName             string                 `json:"fileName,omitempty"`
	ContentType          string                 `json:"contentType,omitempty"`
	FileInfo             map[string]interface{} `json:"fileInfo,omitempty"`
	ServerSideEncryption []interface{}          `json:"serverSideEncryption,omitempty"`
	Source               string                 `json:"source,omitempty"`
}

func (s *BucketFileVersionInput) ResourceName() string {
	return "bucket_file_version"
}

// BucketNotificationRules

type BucketNotificationRulesOutput struct {
	BucketId          string             `json:"bucketId"`
	NotificationRules []NotificationRule `json:"notificationRules"`
}

func (s *BucketNotificationRulesOutput) ResourceName() string {
	return "bucket_notification_rules"
}

type BucketNotificationRulesInput struct {
	BucketId          string        `json:"bucketId"`
	NotificationRules []interface{} `json:"notificationRules,omitempty"`
}

func (s *BucketNotificationRulesInput) ResourceName() string {
	return "bucket_notification_rules"
}
