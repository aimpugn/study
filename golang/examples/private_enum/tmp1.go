package main

// import "context"

// type PGExtendedDataKey int

// const (
// 	FIRST_ISSUED_BILLING_KEY_IMP_UID PGExtendedDataKey = iota
// 	SOME_MERCHANT_UID
// 	SOME_KEY
// )

// type IssueBillingKeyRequest struct {
// 	channelId    string
// 	customerUid  string
// 	extendedData map[PGExtendedDataKey]interface{}
// }

// type IssueBillingKeyResponse struct{}

// type PayWithBillingKeyRequest struct {
// 	billingKey   string
// 	amount       string
// 	extendedData map[PGExtendedDataKey]interface{}
// }

// type PayWithBillingKeyResponse struct{}

// type PGClient interface {
// 	GetExtendedDataKeyForIssueBillingKey(ctx context.Context) []PGExtendedDataKey
// 	IssueBillingKey(ctx context.Context, req IssueBillingKeyRequest) (IssueBillingKeyResponse, error)
// 	GetExtendedDataKeyForPayWithBillingKey(ctx context.Context) []PGExtendedDataKey
// 	PayWithBillingKey(ctx context.Context, req PayWithBillingKeyRequest) (PayWithBillingKeyResponse, error)
// }
