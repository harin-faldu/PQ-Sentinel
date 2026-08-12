package scanner

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/smithy-go"
)

const maxConcurrentRegions = 5 // bounded worker pool

// LiveKMSFinding is an internal intermediate shape before conversion to
// models.CryptoAsset — makes the alias-based drift matcher's job easier
// since it needs both KeyId and Alias, not just a flattened asset.
type LiveKMSFinding struct {
	Region          string
	KeyID           string
	KeySpec         string
	KeyUsage        string
	Enabled         bool
	RotationEnabled bool
	RotationUnknown bool   // true if GetKeyRotationStatus was denied
	Alias           string
	AccessDenied    bool
	AccessDeniedMsg string
}

// ScanLiveAWSKMS discovers real, deployed KMS keys across every enabled
// region in the account. Never returns an error for per-key or per-region
// access issues — those are captured per-finding so a partial IAM policy
// still yields partial (not zero) results.
func ScanLiveAWSKMS(ctx context.Context) ([]LiveKMSFinding, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS credentials: %w", err)
	}

	regions := DiscoverEnabledRegions(ctx, cfg)

	var (
		mu       sync.Mutex
		findings []LiveKMSFinding
		wg       sync.WaitGroup
		sem      = make(chan struct{}, maxConcurrentRegions)
	)

	for _, region := range regions {
		wg.Add(1)
		sem <- struct{}{} // acquire a worker slot
		go func(region string) {
			defer wg.Done()
			defer func() { <-sem }() // release the slot

			regionalFindings := scanRegion(ctx, cfg, region)

			mu.Lock()
			findings = append(findings, regionalFindings...)
			mu.Unlock()
		}(region)
	}
	wg.Wait()

	return findings, nil
}

func scanRegion(ctx context.Context, cfg aws.Config, region string) []LiveKMSFinding {
	regionalCfg := cfg.Copy()
	regionalCfg.Region = region
	client := kms.NewFromConfig(regionalCfg)

	aliasByKeyID := fetchAliasMap(ctx, client)

	var findings []LiveKMSFinding
	paginator := kms.NewListKeysPaginator(client, &kms.ListKeysInput{})
	for paginator.HasMorePages() {
		page, err := withRetry(ctx, func() (*kms.ListKeysOutput, error) {
			return paginator.NextPage(ctx)
		})
		if err != nil {
			// Whole-region failure (e.g. region disabled, no ListKeys permission at all).
			findings = append(findings, LiveKMSFinding{
				Region:          region,
				AccessDenied:    true,
				AccessDeniedMsg: err.Error(),
			})
			return findings
		}
		for _, keyEntry := range page.Keys {
			if keyEntry.KeyId == nil {
				continue
			}
			findings = append(findings, describeOneKey(ctx, client, region, *keyEntry.KeyId, aliasByKeyID))
		}
	}
	return findings
}

func describeOneKey(ctx context.Context, client *kms.Client, region, keyID string, aliasByKeyID map[string]string) LiveKMSFinding {
	f := LiveKMSFinding{Region: region, KeyID: keyID, Alias: aliasByKeyID[keyID]}

	descOut, err := withRetry(ctx, func() (*kms.DescribeKeyOutput, error) {
		return client.DescribeKey(ctx, &kms.DescribeKeyInput{KeyId: aws.String(keyID)})
	})
	if err != nil {
		if isAccessDenied(err) {
			f.AccessDenied = true
			f.AccessDeniedMsg = "DescribeKey: access denied"
			return f
		}
		f.AccessDenied = true
		f.AccessDeniedMsg = fmt.Sprintf("DescribeKey failed: %v", err)
		return f
	}
	if descOut.KeyMetadata != nil {
		f.KeySpec = string(descOut.KeyMetadata.KeySpec)
		f.KeyUsage = string(descOut.KeyMetadata.KeyUsage)
		f.Enabled = descOut.KeyMetadata.Enabled
	}

	// Rotation status is a SEPARATE call and can be denied independently
	rotOut, err := withRetry(ctx, func() (*kms.GetKeyRotationStatusOutput, error) {
		return client.GetKeyRotationStatus(ctx, &kms.GetKeyRotationStatusInput{KeyId: aws.String(keyID)})
	})
	if err != nil {
		f.RotationUnknown = true
	} else {
		f.RotationEnabled = rotOut.KeyRotationEnabled
	}

	return f
}

func fetchAliasMap(ctx context.Context, client *kms.Client) map[string]string {
	result := make(map[string]string)
	paginator := kms.NewListAliasesPaginator(client, &kms.ListAliasesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return result // best-effort
		}
		for _, a := range page.Aliases {
			if a.TargetKeyId != nil && a.AliasName != nil {
				result[*a.TargetKeyId] = *a.AliasName
			}
		}
	}
	return result
}

// isAccessDenied checks AWS's typed error interface rather than string-matching
func isAccessDenied(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		return code == "AccessDeniedException" || code == "AccessDenied" || code == "NotAuthorizedException"
	}
	return false
}

// withRetry applies simple exponential backoff for ThrottlingException,
// leaving all other errors (including AccessDenied) to fail immediately.
func withRetry[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	var lastErr error
	backoff := 250 * time.Millisecond
	for attempt := 0; attempt < 4; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "ThrottlingException" {
			lastErr = err
			select {
			case <-time.After(backoff):
				backoff *= 2
				continue
			case <-ctx.Done():
				var zero T
				return zero, ctx.Err()
			}
		}
		var zero T
		return zero, err // non-throttling error — fail fast
	}
	var zero T
	return zero, lastErr
}
