package sync

import (
	"math/rand/v2"
	"regexp"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorClass string

const (
	ErrorClassTransient ErrorClass = "transient"
	ErrorClassPermanent ErrorClass = "permanent"

	baseRetryDelay = 2 * time.Second
	maxRetryDelay  = 300 * time.Second
	maxBackoffExp  = 8
)

var sqlStateRegex = regexp.MustCompile(`SQLSTATE ([0-9A-Z]{5})`)

var permanentValidationSubstrings = []string{
	"unsupported sync entity type",
	"unsupported entity type for upsert",
	"unsupported customer sync action",
	"customer sync requires id, organization_id, and customer_code",
	"already belongs to another cloud customer",
	"invalid customer sync payload",
	"invalid checkout",
	"checkout order ID is required",
	"checkout must be fulfilled locally",
	"checkout has no order lines",
	"checkout line does not belong to order",
	"checkout line ID belongs to another order",
	"empty sync payload",
	"invalid transaction payload",
}

// ClassifySyncError categorizes a push/sync error into permanent (quarantine immediately)
// or transient (retry infinitely with exponential backoff and jitter).
//
// Permanent errors:
//   - PostgreSQL SQLSTATE 22xxx (data exception) and 23xxx (integrity constraint violation, including 23503 FK)
//   - Explicit cloud domain/payload validation errors
//   - gRPC InvalidArgument, NotFound, AlreadyExists, FailedPrecondition, OutOfRange, Unimplemented
//
// Transient errors:
//   - Network / dial / stream errors, Unavailable, DeadlineExceeded, ResourceExhausted, Aborted
//   - PostgreSQL SQLSTATE 40xxx (serialization/deadlock), 53xxx (resources), 55P03 (lock), 57xxx (crash/shutdown)
//   - SHA-256 transport checksum mismatches
func ClassifySyncError(err error) ErrorClass {
	if err == nil {
		return ErrorClassTransient
	}

	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument,
			codes.NotFound,
			codes.AlreadyExists,
			codes.FailedPrecondition,
			codes.OutOfRange,
			codes.Unimplemented:
			return ErrorClassPermanent
		case codes.Unavailable,
			codes.DeadlineExceeded,
			codes.ResourceExhausted,
			codes.Aborted,
			codes.Canceled:
			return ErrorClassTransient
		}
	}

	msg := err.Error()
	if code := ExtractSQLState(msg); code != "" {
		if strings.HasPrefix(code, "22") || strings.HasPrefix(code, "23") {
			return ErrorClassPermanent
		}
		return ErrorClassTransient
	}

	lower := strings.ToLower(msg)
	for _, needle := range permanentValidationSubstrings {
		if strings.Contains(lower, strings.ToLower(needle)) {
			return ErrorClassPermanent
		}
	}

	return ErrorClassTransient
}

// ExtractSQLState extracts a 5-character PostgreSQL SQLSTATE code from an error message if present.
func ExtractSQLState(errMsg string) string {
	matches := sqlStateRegex.FindStringSubmatch(errMsg)
	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}

// ComputeBackoffDelay returns a jittered exponential backoff duration in [0.5 * cap, 1.0 * cap],
// where cap = min(300s, 2s * 2^min(retryCount, 8)).
func ComputeBackoffDelay(retryCount int) time.Duration {
	if retryCount < 0 {
		retryCount = 0
	}
	exp := retryCount
	if exp > maxBackoffExp {
		exp = maxBackoffExp
	}
	ceiling := baseRetryDelay * time.Duration(1<<exp)
	if ceiling > maxRetryDelay {
		ceiling = maxRetryDelay
	}
	// Equal/full jitter in [0.5, 1.0] of ceiling
	factor := 0.5 + rand.Float64()*0.5
	return time.Duration(float64(ceiling) * factor)
}
