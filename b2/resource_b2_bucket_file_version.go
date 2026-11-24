//####################################################################
//
// File: b2/resource_b2_bucket_file_version.go
//
// Copyright 2021 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// serverAddedFileInfoKeys are the file info keys that B2 adds on upload, so they are in the state
// without ever being in the config.
var serverAddedFileInfoKeys = []string{"sse_c_key_id", "large_file_sha1"}

func resourceB2BucketFileVersion() *schema.Resource {
	return &schema.Resource{
		Description: "B2 bucket file version resource.",

		CreateContext: resourceB2BucketFileVersionCreate,
		ReadContext:   resourceB2BucketFileVersionRead,
		DeleteContext: resourceB2BucketFileVersionDelete,
		CustomizeDiff: resourceB2BucketFileVersionDiff,
		Schema: map[string]*schema.Schema{
			"bucket_id": {
				Description:  "The ID of the bucket.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"file_name": {
				Description:  "The name of the B2 file.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"source": {
				Description: "Path to the local file. During planning, content changes require replacement when a usable " +
					"stored checksum and a readable regular source are available. Otherwise content comparison is skipped.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"content_type": {
				Description: "Content type. If not set, it will be set based on the file extension.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return new == "" // The API sets default value
				},
			},
			"file_info": {
				Description: "The custom information that is uploaded with the file. B2 converts keys to lower " +
					"case, so they are stored and returned in lower case. B2 also adds 'sse_c_key_id' to files " +
					"uploaded in SSE-C mode, and 'large_file_sha1' to large files.",
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
				ValidateDiagFunc: validateMapKeys(serverAddedFileInfoKeys...),
				DiffSuppressFunc: suppressMapKeyCaseDiff("file_info", serverAddedFileInfoKeys...),
			},
			"server_side_encryption": {
				Description: "Server-side encryption settings.",
				Type:        schema.TypeList,
				Elem:        getResourceFileEncryptionElem(),
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				MaxItems:    1,
			},
			"action": {
				Description: "One of 'start', 'upload', 'hide', 'folder', or other values added in the future.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"content_md5": {
				Description: "MD5 sum of the content.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"content_sha1": {
				Description: "SHA1 hash of the content.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"file_id": {
				Description: "The unique identifier for this version of this file.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"size": {
				Description: "The file size.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"upload_timestamp": {
				Description: "This is a UTC time when this file was uploaded.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

// resourceB2BucketFileVersionDiff compares the source with B2's full-file SHA-1.
// The checksum is used for B2 compatibility, not for authentication.
func resourceB2BucketFileVersionDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if d.Id() == "" || !d.NewValueKnown("source") || d.HasChange("source") {
		return nil
	}

	oldChecksum, _ := d.GetChange("content_sha1")
	checksum := oldChecksum.(string)
	if checksum == "" || checksum == "none" {
		oldInfo, _ := d.GetChange("file_info")
		if value, ok := oldInfo.(map[string]interface{})["large_file_sha1"].(string); ok {
			checksum = value
		}
	}
	decoded, err := hex.DecodeString(checksum)
	if err != nil || len(decoded) != sha1.Size {
		return nil
	}

	path := d.Get("source").(string)
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return ctx.Err()
	}
	file, err := openFileVersionSource(path)
	if err != nil {
		return ctx.Err()
	}
	defer func() { _ = file.Close() }()
	info, err = file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return ctx.Err()
	}
	stopClose := context.AfterFunc(ctx, func() { _ = file.Close() })
	defer stopClose()
	hash := sha1.New()
	if _, err := io.Copy(hash, &fileVersionContextReader{ctx: ctx, reader: file}); err != nil {
		return ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), checksum) {
		return nil
	}
	if err := d.SetNewComputed("file_id"); err != nil {
		return err
	}
	return d.ForceNew("file_id")
}

type fileVersionContextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *fileVersionContextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

func resourceB2BucketFileVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := BucketFileVersionInput{
		BucketId:             d.Get("bucket_id").(string),
		FileName:             d.Get("file_name").(string),
		Source:               d.Get("source").(string),
		ContentType:          d.Get("content_type").(string),
		FileInfo:             d.Get("file_info").(map[string]interface{}),
		ServerSideEncryption: d.Get("server_side_encryption").([]interface{}),
	}
	var output BucketFileVersionOutput
	err := client.Apply(ctx, OpResourceCreate, &input, &output)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(output.FileId)
	retainSSECKey(d, &output)

	err = client.Populate(ctx, OpResourceCreate, &output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceB2BucketFileVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := BucketFileVersionInput{
		FileId: d.Id(),
	}

	var output BucketFileVersionOutput
	err := client.Apply(ctx, OpResourceRead, &input, &output)
	if err != nil {
		return diag.FromErr(err)
	}
	if output.FileId == "" && !d.IsNewResource() {
		// A missing version returns only bridge metadata, without a file ID.
		d.SetId("")
		return nil
	}

	// These fields are not returned by the API but are needed for the resource
	output.BucketId = d.Get("bucket_id").(string)
	output.Source = d.Get("source").(string)
	retainSSECKey(d, &output)

	err = client.Populate(ctx, OpResourceRead, &output, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// B2 returns encryption metadata but never the supplied customer key. Preserve
// the configured key so unchanged plans converge and rotations remain visible.
func retainSSECKey(d *schema.ResourceData, output *BucketFileVersionOutput) {
	if output.ServerSideEncryption == nil || output.ServerSideEncryption.Mode != "SSE-C" {
		return
	}
	for _, item := range d.Get("server_side_encryption.0.key").([]interface{}) {
		key := item.(map[string]interface{})
		output.ServerSideEncryption.Key = append(output.ServerSideEncryption.Key, ResourceFileEncryptionKey{
			KeyId: key["key_id"].(string), SecretB64: key["secret_b64"].(string),
		})
	}
}

func resourceB2BucketFileVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	input := BucketFileVersionInput{
		FileId:   d.Id(),
		FileName: d.Get("file_name").(string),
	}

	err := client.Apply(ctx, OpResourceDelete, &input, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
